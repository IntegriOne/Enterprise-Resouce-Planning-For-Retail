//go:build integration

package talent_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/talent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dimensionIntegrationPolicy struct{}

func (dimensionIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (dimensionIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestTalentDimensionRPCWithPostgreSQL(t *testing.T) {
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
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Errorf("close HRM test database pool: %v", err)
		}
	})
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.TalentDimensionValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	ids := operation.UUIDGenerator{}
	frameworkID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	cycleID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_talent_frameworks (id, framework_code, framework_version, framework_name, rubric_document_version_reference_id, rubric_digest, effective_from, approved_by_principal_id, approved_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		frameworkID, "DIMENSION-"+frameworkID, 1, "Dimension test", "33333333-3333-4333-8333-333333333333", []byte{1}, "2026-01-01", "44444444-4444-4444-8444-444444444444", time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
	).Error; err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: dimensionIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		TalentFramework: repos.TalentFramework, TalentDimension: repos.TalentDimension,
		TalentReviewCycle: repos.TalentReviewCycle, TalentAssessment: repos.TalentAssessment,
		TalentAssessmentScore: repos.TalentAssessmentScore, TalentCalibrationSession: repos.TalentCalibrationSession,
		TalentCalibrationDecision: repos.TalentCalibrationDecision, TalentPool: repos.TalentPool,
		TalentPoolMembership: repos.TalentPoolMembership,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///hrm-talent-dimension-integration", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewTalentServiceClient(conn)
	code, name, definition, weight := "LEADERSHIP", "Leadership", "Observed leadership", "100.000"
	input := &pb.TalentDimensionInput{FrameworkId: &frameworkID, DimensionCode: &code,
		DimensionName: &name, Definition: &definition, WeightPercent: &weight,
		RatingAnchors: &common.JsonDocument{Utf8: []byte(`{"1":"basic","2":"developing","3":"capable","4":"strong","5":"expert"}`)}}
	invalidInput := proto.CloneOf(input)
	invalidInput.RatingAnchors = &common.JsonDocument{Utf8: []byte(`{"1":"basic"}`)}
	if _, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: invalidInput}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("incomplete anchors status = %v, want InvalidArgument", status.Code(err))
	}
	created, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: input})
	if err != nil || created.GetValue().GetId() == "" {
		t.Fatalf("create dimension: %+v, %v", created, err)
	}
	id := created.GetValue().GetId()
	if _, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate code status = %v, want AlreadyExists", status.Code(err))
	}
	updatedName := "Leadership updated"
	updated, err := client.UpdateTalentDimension(ctx, &pb.UpdateTalentDimensionRequest{Id: id,
		Input: &pb.TalentDimensionInput{DimensionName: &updatedName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"dimension_name"}}})
	if err != nil || updated.GetValue().GetDimensionName() != updatedName {
		t.Fatalf("update dimension: %+v, %v", updated, err)
	}
	loaded, err := client.GetTalentDimensionByID(ctx, &common.IDRequest{Id: id})
	if err != nil || loaded.GetValue().GetDimensionName() != updatedName {
		t.Fatalf("get updated dimension: %+v, %v", loaded, err)
	}
	secondCode := "EXECUTION"
	secondInput := proto.CloneOf(input)
	secondInput.DimensionCode = &secondCode
	second, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: secondInput})
	if err != nil || second.GetValue().GetId() == "" {
		t.Fatalf("create second dimension: %+v, %v", second, err)
	}
	for _, offset := range []int32{0, 1} {
		page, err := client.FindTalentDimensionByFrameworkID(ctx, &common.StringFinderRequest{Value: &frameworkID, Limit: 1, Offset: offset})
		if err != nil || len(page.GetItems()) != 1 || page.GetPage().GetTotal() != 2 {
			t.Fatalf("framework finder page %d: %+v, %v", offset, page, err)
		}
	}
	if _, err := client.DeleteTalentDimension(ctx, &common.IDRequest{Id: second.GetValue().GetId()}); err != nil {
		t.Fatalf("delete dimension before cycle: %v", err)
	}
	if _, err := client.GetTalentDimensionByID(ctx, &common.IDRequest{Id: second.GetValue().GetId()}); status.Code(err) != codes.NotFound {
		t.Fatalf("deleted dimension status = %v, want NotFound", status.Code(err))
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_talent_review_cycles (id, cycle_code, framework_id, legal_entity_reference_id, period_start, period_end, status, owner_principal_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		cycleID, "DIMENSION-CYCLE-"+cycleID, frameworkID, "55555555-5555-4555-8555-555555555555", "2026-01-01", "2026-12-31", "draft", "66666666-6666-4666-8666-666666666666",
	).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateTalentDimension(ctx, &pb.UpdateTalentDimensionRequest{Id: id,
		Input: &pb.TalentDimensionInput{DimensionName: &name}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"dimension_name"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used framework update status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteTalentDimension(ctx, &common.IDRequest{Id: id}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used framework delete status = %v, want FailedPrecondition", status.Code(err))
	}
	secondInput.DimensionCode = &secondCode
	if _, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: secondInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used framework create status = %v, want FailedPrecondition", status.Code(err))
	}
}
