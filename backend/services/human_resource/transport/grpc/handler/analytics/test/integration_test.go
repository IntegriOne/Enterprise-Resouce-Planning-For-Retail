//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/analytics/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/analytics"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/analytics"
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

type integrationPolicy struct{}

func (*integrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (*integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestMetricDefinitionRPCWithPostgreSQL(t *testing.T) {
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Errorf("close HRM test database pool: %v", err)
		}
	})
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(db)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.AnalyticsValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	runtime := operation.Dependencies{
		Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
		Policy: &integrationPolicy{}, Integrity: security.NewIntegrity(validators),
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: runtime, HumanCapitalMetricDefinition: repos.HumanCapitalMetricDefinition,
		HumanCapitalMetricSnapshot: repos.HumanCapitalMetricSnapshot,
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
			t.Errorf("serve HRM test gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-analytics-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close HRM test gRPC connection: %v", err)
		}
	})
	client := pb.NewAnalyticsServiceClient(connection)
	code, version, name, domain := "retention", "1", "Retention rate", "turnover"
	created, err := client.CreateHumanCapitalMetricDefinition(ctx, &pb.CreateHumanCapitalMetricDefinitionRequest{
		Input: &pb.HumanCapitalMetricDefinitionInput{
			MetricCode: &code, MetricVersion: &version, MetricName: &name, MetricDomain: &domain,
		},
	})
	if err != nil || created.GetValue().GetId() == "" {
		t.Fatalf("create metric definition: %v", err)
	}
	id := created.Value.Id
	var snapshotID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if snapshotID != "" {
			if err := db.WithContext(cleanupCtx).Exec("DELETE FROM hr_human_capital_metric_snapshots WHERE id = ?", snapshotID).Error; err != nil {
				t.Errorf("cleanup metric snapshot: %v", err)
			}
		}
		for _, table := range []string{"hr_outbox_events", "hr_human_capital_metric_definitions"} {
			column := "aggregate_id"
			if table == "hr_human_capital_metric_definitions" {
				column = "id"
			}
			if err := db.WithContext(cleanupCtx).Exec("DELETE FROM "+table+" WHERE "+column+" = ?", id).Error; err != nil {
				t.Errorf("cleanup metric definition: %v", err)
			}
		}
	})
	if _, err := client.CreateHumanCapitalMetricDefinition(ctx, &pb.CreateHumanCapitalMetricDefinitionRequest{
		Input: &pb.HumanCapitalMetricDefinitionInput{
			MetricCode: &code, MetricVersion: &version, MetricName: &name, MetricDomain: &domain,
		},
	}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate code/version: %v", err)
	}
	read, err := client.GetHumanCapitalMetricDefinitionByID(ctx, &common.IDRequest{Id: id})
	if err != nil || read.GetValue().GetMetricCode() != code {
		t.Fatalf("read metric definition: %v", err)
	}
	page, err := client.FindHumanCapitalMetricDefinitionByMetricCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1})
	if err != nil || len(page.GetItems()) != 1 || page.GetPage().GetTotal() != 1 {
		t.Fatalf("paged metric finder: %v", err)
	}
	updatedName := "Employee retention rate"
	updated, err := client.UpdateHumanCapitalMetricDefinition(ctx, &pb.UpdateHumanCapitalMetricDefinitionRequest{
		Id: id, Input: &pb.HumanCapitalMetricDefinitionInput{MetricName: &updatedName},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"metric_name"}},
	})
	if err != nil || updated.GetValue().GetMetricName() != updatedName {
		t.Fatalf("update unreferenced definition: %v", err)
	}
	ids := operation.UUIDGenerator{}
	snapshotID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_human_capital_metric_snapshots (id, metric_definition_id, period_start, period_end, methodology_version, source_digest) VALUES (?, ?, ?, ?, ?, decode('ab', 'hex'))",
		snapshotID, id, "2026-01-01", "2026-01-31", "1",
	).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateHumanCapitalMetricDefinition(ctx, &pb.UpdateHumanCapitalMetricDefinitionRequest{
		Id: id, Input: &pb.HumanCapitalMetricDefinitionInput{MetricName: &name},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"metric_name"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("change referenced definition: %v", err)
	}
	if _, err := client.DeleteHumanCapitalMetricDefinition(ctx, &common.IDRequest{Id: id}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete referenced definition: %v", err)
	}
	if err := db.WithContext(ctx).Exec("DELETE FROM hr_human_capital_metric_snapshots WHERE id = ?", snapshotID).Error; err != nil {
		t.Fatal(err)
	}
	snapshotID = ""
	if _, err := client.DeleteHumanCapitalMetricDefinition(ctx, &common.IDRequest{Id: id}); err != nil {
		t.Fatalf("delete unreferenced definition: %v", err)
	}
	var events int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", id).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 3 {
		t.Fatalf("metric definition outbox count = %d, want 3", events)
	}
}
