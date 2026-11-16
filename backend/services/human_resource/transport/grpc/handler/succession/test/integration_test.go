//go:build integration

package succession_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/succession"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/succession"
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

type integrationPolicy struct{}

func (integrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestSuccessionPlanRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" ||
		!strings.HasPrefix(strings.TrimPrefix(parsed.Path, "/"), "hrm_disposable_") ||
		!strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") ||
		os.Getenv("HRM_TEST_DATABASE_DISPOSABLE") != "1" {
		t.Fatal("integration test requires explicit opt-in and a disposable HRM test database")
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
	validators, err := security.SuccessionValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: integrationPolicy{}, Integrity: security.NewIntegrity(validators),
		},
		SuccessionPlan: repos.SuccessionPlan, SuccessionCandidate: repos.SuccessionCandidate,
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
	connection, err := grpc.NewClient("passthrough:///hrm-succession-integration",
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
	client := pb.NewSuccessionServiceClient(connection)
	ids := operation.UUIDGenerator{}
	familyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	jobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	positionID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var planID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, id string }{
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
			{"DELETE FROM hr_succession_plans WHERE id = ?", planID},
			{"DELETE FROM hr_positions WHERE id = ?", positionID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
		} {
			if statement.id == "" {
				continue
			}
			if err := db.WithContext(cleanupCtx).Exec(statement.query, statement.id).Error; err != nil {
				t.Errorf("cleanup HRM succession test data: %v", err)
			}
		}
	})
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, "FAMILY-"+familyID, "Test family", "professional",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, "JOB-"+jobID, "Test job", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, position_status, valid_from) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		positionID, "POSITION-"+positionID, "Test position", jobID, familyID, familyID, "active", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	code := "SUCCESSION-" + positionID
	version := int32(1)
	reviewDate := timestamppb.New(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	request := &pb.CreateSuccessionPlanRequest{Input: &pb.SuccessionPlanInput{
		PositionId: &positionID, PlanCode: &code, PlanVersion: &version, ReviewDate: reviewDate,
	}}
	created, err := client.CreateSuccessionPlan(ctx, request)
	if err != nil || created.GetValue().GetId() == "" || created.Value.GetConfidentialityClass() != "highly_restricted" {
		t.Fatalf("create draft succession plan: %v", err)
	}
	planID = created.Value.Id
	if _, err := client.CreateSuccessionPlan(ctx, request); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate position/code/version: %v", err)
	}
	found, err := client.FindSuccessionPlanByPositionID(ctx, &commonv1.StringFinderRequest{Value: &positionID, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetPage().GetTotal() != 1 {
		t.Fatalf("paged succession plan finder: %v", err)
	}
	listed, err := client.GetAllSuccessionPlans(ctx, &commonv1.PageRequest{Limit: 1, Offset: 1})
	if err != nil || listed.GetPage().GetTotal() != 1 || len(listed.GetItems()) != 0 {
		t.Fatalf("succession plan pagination offset: %v", err)
	}
	risk := "low"
	if _, err := client.UpdateSuccessionPlan(ctx, &pb.UpdateSuccessionPlanRequest{
		Id: planID, Input: &pb.SuccessionPlanInput{SuccessionRisk: &risk},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"succession_risk"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("draft workflow field update: %v", err)
	}
	newReviewDate := timestamppb.New(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	updated, err := client.UpdateSuccessionPlan(ctx, &pb.UpdateSuccessionPlanRequest{
		Id: planID, Input: &pb.SuccessionPlanInput{ReviewDate: newReviewDate},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"review_date"}},
	})
	if err != nil || !updated.GetValue().GetReviewDate().AsTime().Equal(newReviewDate.AsTime()) {
		t.Fatalf("update draft review date: %v", err)
	}
	candidateWorkerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	readiness := "ready_now"
	if _, err := client.CreateSuccessionCandidate(ctx, &pb.CreateSuccessionCandidateRequest{Input: &pb.SuccessionCandidateInput{
		SuccessionPlanId: &planID, CandidateWorkerId: &candidateWorkerID,
		ReadinessCode: &readiness, ReadinessAssessedAt: newReviewDate,
	}}); status.Code(err) != codes.Unavailable {
		t.Fatalf("candidate without production validator: %v", err)
	}
	var candidateCount int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_succession_candidates WHERE succession_plan_id = ?", planID).Scan(&candidateCount).Error; err != nil || candidateCount != 0 {
		t.Fatalf("rejected candidate persisted: count=%d error=%v", candidateCount, err)
	}
	if _, err := client.DeleteSuccessionPlan(ctx, &commonv1.IDRequest{Id: planID}); err != nil {
		t.Fatalf("delete unused draft plan: %v", err)
	}
	if _, err := client.GetSuccessionPlanByID(ctx, &commonv1.IDRequest{Id: planID}); status.Code(err) != codes.NotFound {
		t.Fatalf("deleted plan lookup: %v", err)
	}
	var outboxCount int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", planID).Scan(&outboxCount).Error; err != nil || outboxCount != 3 {
		t.Fatalf("outbox count after create, update and delete: count=%d error=%v", outboxCount, err)
	}
}
