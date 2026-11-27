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
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type snapshotFoundationStub struct{}

func (snapshotFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (snapshotFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (snapshotFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (snapshotFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (snapshotFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (snapshotFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

func TestMetricSnapshotRPCWithPostgreSQL(t *testing.T) {
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
	t.Cleanup(func() { _ = pool.Close() })
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
	snapValidators, err := security.AnalyticsSnapshotValidators(repos, snapshotFoundationStub{})
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range snapValidators {
		validators[k] = v
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: &integrationPolicy{}, Integrity: security.NewIntegrity(validators),
		},
		HumanCapitalMetricDefinition: repos.HumanCapitalMetricDefinition,
		HumanCapitalMetricSnapshot:   repos.HumanCapitalMetricSnapshot,
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
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-analytics-snapshot-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewAnalyticsServiceClient(conn)

	ids := operation.UUIDGenerator{}
	unitID, _ := ids.NewID()
	var defID, snapID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_human_capital_metric_snapshots WHERE id = ?", snapID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", snapID},
			{"DELETE FROM hr_human_capital_metric_definitions WHERE id = ?", defID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", defID},
		} {
			if stmt.v != "" {
				_ = db.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error
			}
		}
	})

	defCode, defVer, defName, defDom := "SNAP-HC-"+unitID, "v1", "Headcount Snapshot", "workforce"
	defRes, err := client.CreateHumanCapitalMetricDefinition(ctx, &pb.CreateHumanCapitalMetricDefinitionRequest{Input: &pb.HumanCapitalMetricDefinitionInput{
		MetricCode: &defCode, MetricVersion: &defVer, MetricName: &defName, MetricDomain: &defDom,
	}})
	if err != nil {
		t.Fatalf("create metric definition: %v", err)
	}
	defID = defRes.GetValue().GetId()

	pStart := timestamppb.New(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	pEnd := timestamppb.New(time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC))
	valNum := "42.00000000"
	pop := int64(42)
	snapVer := int32(1)
	meth, conf, qual := "iso30414-v1", "restricted", "provisional"
	snapRes, err := client.CreateHumanCapitalMetricSnapshot(ctx, &pb.CreateHumanCapitalMetricSnapshotRequest{Input: &pb.HumanCapitalMetricSnapshotInput{
		MetricDefinitionId: &defID, PeriodStart: pStart, PeriodEnd: pEnd,
		OrganizationUnitReferenceId: &unitID, ValueNumeric: &valNum, PopulationCount: &pop,
		DataQualityStatus: &qual, SnapshotVersion: &snapVer, MethodologyVersion: &meth,
		ConfidentialityClass: &conf, SourceDigest: []byte("sha256"),
	}})
	if err != nil {
		t.Fatalf("create metric snapshot: %v", err)
	}
	snapID = snapRes.GetValue().GetId()

	if _, err := client.CreateHumanCapitalMetricSnapshot(ctx, &pb.CreateHumanCapitalMetricSnapshotRequest{Input: &pb.HumanCapitalMetricSnapshotInput{
		MetricDefinitionId: &defID, PeriodStart: pStart, PeriodEnd: pEnd,
		OrganizationUnitReferenceId: &unitID, ValueNumeric: &valNum, PopulationCount: &pop,
		DataQualityStatus: &qual, SnapshotVersion: &snapVer, MethodologyVersion: &meth,
		ConfidentialityClass: &conf, SourceDigest: []byte("sha256"),
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate metric snapshot status = %v, want AlreadyExists", status.Code(err))
	}
	if _, err := client.UpdateHumanCapitalMetricSnapshot(ctx, &pb.UpdateHumanCapitalMetricSnapshotRequest{
		Id: snapID, Input: &pb.HumanCapitalMetricSnapshotInput{MethodologyVersion: &meth},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"methodology_version"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update metric snapshot status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteHumanCapitalMetricSnapshot(ctx, &common.IDRequest{Id: snapID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete metric snapshot status = %v, want FailedPrecondition", status.Code(err))
	}
}
