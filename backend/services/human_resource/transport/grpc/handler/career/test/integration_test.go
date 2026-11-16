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

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/career"
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

func TestDevelopmentPlanAndActionRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.CareerValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	runtime := operation.Dependencies{
		Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
		Policy: &integrationPolicy{}, Integrity: security.NewIntegrity(validators),
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: runtime, DevelopmentPlan: repos.DevelopmentPlan, DevelopmentAction: repos.DevelopmentAction,
		CareerPath: repos.CareerPath, CareerPathTransition: repos.CareerPathTransition,
		WorkerCareerAspiration: repos.WorkerCareerAspiration,
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
	connection, err := grpc.NewClient("passthrough:///hrm-career-integration",
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
	client := pb.NewCareerServiceClient(connection)
	ids := operation.UUIDGenerator{}
	familyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	jobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	mentorID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var planID, actionID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, value string }{
			{"DELETE FROM hr_development_actions WHERE id = ?", actionID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", actionID},
			{"DELETE FROM hr_development_plans WHERE id = ?", planID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
			{"DELETE FROM hr_workers WHERE id = ?", mentorID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
		} {
			if statement.value == "" {
				continue
			}
			if err := db.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
				t.Errorf("cleanup HRM career test data: %v", err)
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
	for _, id := range []string{workerID, mentorID} {
		if err := db.WithContext(ctx).Exec(
			"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
			id, "WORKER-"+id, familyID, "employee", "active", familyID,
		).Error; err != nil {
			t.Fatal(err)
		}
	}
	start := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	target := timestamppb.New(time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	code, version := "growth-"+workerID, int32(1)
	created, err := client.CreateDevelopmentPlan(ctx, &pb.CreateDevelopmentPlanRequest{Input: &pb.DevelopmentPlanInput{
		WorkerId: &workerID, PlanCode: &code, PlanVersion: &version, TargetJobId: &jobID,
		MentorWorkerId: &mentorID, StartDate: start, TargetDate: target,
	}})
	if err != nil || created.GetValue().GetId() == "" {
		t.Fatalf("create draft development plan: %v", err)
	}
	planID = created.Value.Id
	if _, err := client.CreateDevelopmentPlan(ctx, &pb.CreateDevelopmentPlanRequest{Input: &pb.DevelopmentPlanInput{
		WorkerId: &workerID, PlanCode: &code, PlanVersion: &version, StartDate: start, TargetDate: target,
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate worker/code/version: %v", err)
	}
	found, err := client.FindDevelopmentPlanByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetPage().GetTotal() != 1 {
		t.Fatalf("paged development plan finder: %v", err)
	}
	objective, actionType, number := "Prepare for job", "coaching", int32(1)
	action, err := client.CreateDevelopmentAction(ctx, &pb.CreateDevelopmentActionRequest{Input: &pb.DevelopmentActionInput{
		DevelopmentPlanId: &planID, ActionNumber: &number, ActionType: &actionType,
		Objective: &objective, DueDate: target,
	}})
	if err != nil || action.GetValue().GetId() == "" {
		t.Fatalf("create development action: %v", err)
	}
	actionID = action.Value.Id
	externalEvidence := familyID
	if _, err := client.UpdateDevelopmentAction(ctx, &pb.UpdateDevelopmentActionRequest{
		Id: actionID, Input: &pb.DevelopmentActionInput{EvidenceDocumentVersionReferenceId: &externalEvidence},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"evidence_document_version_reference_id"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("change action evidence through CRUD: %v", err)
	}
	if _, err := client.DeleteDevelopmentPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete plan with action: %v", err)
	}
	updatedObjective := "Prepare for promotion"
	updated, err := client.UpdateDevelopmentAction(ctx, &pb.UpdateDevelopmentActionRequest{
		Id: actionID, Input: &pb.DevelopmentActionInput{Objective: &updatedObjective},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"objective"}},
	})
	if err != nil || updated.GetValue().GetObjective() != updatedObjective {
		t.Fatalf("update draft development action: %v", err)
	}
	if _, err := client.DeleteDevelopmentAction(ctx, &common.IDRequest{Id: actionID}); err != nil {
		t.Fatalf("delete draft development action: %v", err)
	}
	updatedCode := "growth-revised-" + workerID
	updatedPlan, err := client.UpdateDevelopmentPlan(ctx, &pb.UpdateDevelopmentPlanRequest{
		Id: planID, Input: &pb.DevelopmentPlanInput{PlanCode: &updatedCode},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_code"}},
	})
	if err != nil || updatedPlan.GetValue().GetPlanCode() != updatedCode {
		t.Fatalf("update draft development plan: %v", err)
	}
	if _, err := client.DeleteDevelopmentPlan(ctx, &common.IDRequest{Id: planID}); err != nil {
		t.Fatalf("delete empty draft development plan: %v", err)
	}
	for _, aggregate := range []struct{ id string; want int64 }{{planID, 3}, {actionID, 3}} {
		var count int64
		if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", aggregate.id).Scan(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != aggregate.want {
			t.Fatalf("outbox events for %s = %d, want %d", aggregate.id, count, aggregate.want)
		}
	}
}
