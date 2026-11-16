//go:build integration

package compensation_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/compensation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type compensationIntegrationPolicy struct{}

func (compensationIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (compensationIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestCompensationBandRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.CompensationValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: compensationIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		CompensationBand: repos.CompensationBand, CompensationRecord: repos.CompensationRecord,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if err := <-served; err != nil {
			t.Errorf("serve HRM compensation gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-compensation-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewCompensationServiceClient(connection)
	code := "STORE-L2-" + strings.ReplaceAll(time.Now().Format("150405.000000000"), ".", "")
	version, level, jurisdiction, currency, frequency := int32(1), "L2", "ID", "IDR", "monthly"
	minimum, midpoint, maximum, from := "5000000.000000", "6000000.000000", "7000000.000000", "2026-01-01"
	input := &pb.CompensationBandInput{BandCode: &code, BandVersion: &version, JobLevelCode: &level,
		JurisdictionCode: &jurisdiction, CurrencyCode: &currency, PayFrequency: &frequency,
		MinimumAmount: &minimum, MidpointAmount: &midpoint, MaximumAmount: &maximum, ValidFrom: &from}
	approvalID := "11111111-1111-4111-8111-111111111111"
	input.ApprovalReferenceId = &approvalID
	if _, err := client.CreateCompensationBand(ctx, &pb.CreateCompensationBandRequest{Input: input}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unverified approval status = %v, want FailedPrecondition", status.Code(err))
	}
	input.ApprovalReferenceId = nil
	created, err := client.CreateCompensationBand(ctx, &pb.CreateCompensationBandRequest{Input: input})
	if err != nil || created.GetItem().GetId() == "" {
		t.Fatalf("create band: %+v, %v", created, err)
	}
	id := created.GetItem().GetId()
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_compensation_bands WHERE id = ?", id).Error; err != nil {
			t.Errorf("cleanup band: %v", err)
		}
		if err := database.WithContext(cleanupCtx).Table("hr_outbox_events").Where("aggregate_id = ?", id).Delete(nil).Error; err != nil {
			t.Errorf("cleanup outbox: %v", err)
		}
	})
	if _, err := client.CreateCompensationBand(ctx, &pb.CreateCompensationBandRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate band status = %v, want AlreadyExists", status.Code(err))
	}
	if got, err := client.GetCompensationBandByID(ctx, &common.IDRequest{Id: id}); err != nil || got.GetItem().GetBandCode() != code {
		t.Fatalf("get band: %+v, %v", got, err)
	}
	if page, err := client.FindCompensationBandByBandCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("find band: %+v, %v", page, err)
	}
	if page, err := client.FindCompensationBandByBandCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1, Offset: 1}); err != nil ||
		page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 0 {
		t.Fatalf("offset band: %+v, %v", page, err)
	}
	maximum = "7500000.000000"
	if got, err := client.UpdateCompensationBand(ctx, &pb.UpdateCompensationBandRequest{Id: id,
		Input: &pb.CompensationBandInput{MaximumAmount: &maximum}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"maximum_amount"}}}); err != nil ||
		got.GetItem().GetMaximumAmount() != maximum {
		t.Fatalf("update band: %+v, %v", got, err)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", id).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("outbox after create and update = %d, %v", events, err)
	}
	if _, err := client.DeleteCompensationBand(ctx, &common.IDRequest{Id: id}); err != nil {
		t.Fatalf("delete band: %v", err)
	}
	if _, err := client.GetCompensationBandByID(ctx, &common.IDRequest{Id: id}); status.Code(err) != codes.NotFound {
		t.Fatalf("deleted band status = %v, want NotFound", status.Code(err))
	}
}
