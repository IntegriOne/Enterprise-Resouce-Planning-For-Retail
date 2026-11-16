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

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/onboarding"
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

type integrationEvidenceVerifier struct{}

func (integrationEvidenceVerifier) VerifyOnboardingEvidence(_ context.Context, request service.EvidenceRequest) (service.EvidenceResult, error) {
	return service.EvidenceResult{VersionID: request.VersionID, Verified: true, Immutable: true}, nil
}

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

func TestOnboardingDraftRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
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
	validators, err := security.OnboardingValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: integrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		OnboardingPlan: repos.OnboardingPlan, OnboardingTask: repos.OnboardingTask,
		Evidence: integrationEvidenceVerifier{},
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
	connection, err := grpc.NewClient("passthrough:///hrm-onboarding-integration",
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
	client := pb.NewOnboardingServiceClient(connection)
	ids := operation.UUIDGenerator{}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var planID, taskID, startedPlanID, startedTaskID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, value string }{
			{"DELETE FROM hr_onboarding_tasks WHERE id = ?", startedTaskID},
			{"DELETE FROM hr_onboarding_tasks WHERE id = ?", taskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", startedTaskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", taskID},
			{"DELETE FROM hr_onboarding_plans WHERE id = ?", startedPlanID},
			{"DELETE FROM hr_onboarding_plans WHERE id = ?", planID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", startedPlanID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if statement.value == "" {
				continue
			}
			if err := database.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
				t.Errorf("cleanup HRM onboarding test data: %v", err)
			}
		}
	})
	principal := "00000000-0000-4000-8000-000000000001"
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id, iam_principal_reference_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		workerID, "ONBOARD-"+workerID, workerID, "employee", "preboarding", workerID, principal,
	).Error; err != nil {
		t.Fatal(err)
	}
	code, start, end := "INTRO-"+workerID, "2027-01-10", "2027-01-17"
	planInput := &pb.OnboardingPlanInput{WorkerId: &workerID, PlanCode: &code, StartDate: &start, TargetCompletionDate: &end}
	created, err := client.CreateOnboardingPlan(ctx, &pb.CreateOnboardingPlanRequest{Input: planInput})
	if err != nil || created.GetItem().GetId() == "" || created.GetItem().GetStatus() != "draft" ||
		created.GetItem().GetCreatedByPrincipalId() != principal {
		t.Fatalf("create draft plan: %+v, %v", created, err)
	}
	planID = created.Item.Id
	if _, err := client.CreateOnboardingPlan(ctx, &pb.CreateOnboardingPlanRequest{Input: planInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate plan version: %v", err)
	}
	unknownWorker, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateOnboardingPlan(ctx, &pb.CreateOnboardingPlanRequest{Input: &pb.OnboardingPlanInput{
		WorkerId: &unknownWorker, PlanCode: &code, StartDate: &start,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unknown worker reference: %v", err)
	}
	active := "active"
	if _, err := client.UpdateOnboardingPlan(ctx, &pb.UpdateOnboardingPlanRequest{
		Id: planID, Input: &pb.OnboardingPlanInput{Status: &active},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("start plan through CRUD: %v", err)
	}
	loaded, err := client.GetOnboardingPlanByID(ctx, &common.IDRequest{Id: planID})
	if err != nil || loaded.GetItem().GetPlanCode() != code {
		t.Fatalf("read draft plan: %v", err)
	}
	listed, err := client.FindOnboardingPlanByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1})
	if err != nil || len(listed.GetItems()) != 1 || listed.GetPage().GetTotal() != 1 {
		t.Fatalf("paged plan finder: %v", err)
	}
	allPlans, err := client.GetAllOnboardingPlans(ctx, &common.PageRequest{Limit: 1})
	if err != nil || allPlans.GetPage().GetLimit() != 1 {
		t.Fatalf("paged plan list: %v", err)
	}
	revisedCode := "INTRO-REVISED-" + workerID
	if _, err := client.UpdateOnboardingPlan(ctx, &pb.UpdateOnboardingPlanRequest{
		Id: planID, Input: &pb.OnboardingPlanInput{PlanCode: &revisedCode},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_code"}},
	}); err != nil {
		t.Fatalf("update unused draft plan: %v", err)
	}
	name, owner, taskCode := "Orientation", "worker", "ORIENTATION"
	due := timestamppb.New(time.Date(2027, 1, 17, 12, 0, 0, 0, time.UTC))
	taskInput := &pb.OnboardingTaskInput{OnboardingPlanId: &planID, TaskCode: &taskCode,
		TaskName: &name, TaskOwnerType: &owner, AssignedPrincipalId: &principal, DueAt: due}
	taskCreated, err := client.CreateOnboardingTask(ctx, &pb.CreateOnboardingTaskRequest{Input: taskInput})
	if err != nil || taskCreated.GetItem().GetId() == "" || !taskCreated.GetItem().GetIsMandatory() {
		t.Fatalf("create pending worker task: %+v, %v", taskCreated, err)
	}
	taskID = taskCreated.Item.Id
	if _, err := client.CreateOnboardingTask(ctx, &pb.CreateOnboardingTaskRequest{Input: taskInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate task code: %v", err)
	}
	if _, err := client.CreateOnboardingTask(ctx, &pb.CreateOnboardingTaskRequest{Input: &pb.OnboardingTaskInput{
		OnboardingPlanId: &planID, TaskCode: &taskCode, TaskName: &name, TaskOwnerType: &owner,
		CompletedAt: due,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("client completed task through CRUD: %v", err)
	}
	if _, err := client.DeleteOnboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete plan with task: %v", err)
	}
	if _, err := client.UpdateOnboardingPlan(ctx, &pb.UpdateOnboardingPlanRequest{
		Id: planID, Input: &pb.OnboardingPlanInput{PlanCode: &code},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_code"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("change referenced plan: %v", err)
	}
	loadedTask, err := client.GetOnboardingTaskByID(ctx, &common.IDRequest{Id: taskID})
	if err != nil || loadedTask.GetItem().GetTaskCode() != taskCode {
		t.Fatalf("read pending task: %v", err)
	}
	listedTasks, err := client.FindOnboardingTaskByOnboardingPlanID(ctx, &common.StringFinderRequest{Value: &planID, Limit: 1})
	if err != nil || len(listedTasks.GetItems()) != 1 || listedTasks.GetPage().GetTotal() != 1 {
		t.Fatalf("paged task finder: %v", err)
	}
	allTasks, err := client.GetAllOnboardingTasks(ctx, &common.PageRequest{Limit: 1})
	if err != nil || allTasks.GetPage().GetLimit() != 1 {
		t.Fatalf("paged task list: %v", err)
	}
	newName := "Store orientation"
	if _, err := client.UpdateOnboardingTask(ctx, &pb.UpdateOnboardingTaskRequest{
		Id: taskID, Input: &pb.OnboardingTaskInput{TaskName: &newName},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"task_name"}},
	}); err != nil {
		t.Fatalf("update pending task: %v", err)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", taskID).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("task outbox after accepted/rejected mutations: count=%d, error=%v", events, err)
	}
	if _, err := client.DeleteOnboardingTask(ctx, &common.IDRequest{Id: taskID}); err != nil {
		t.Fatalf("delete pending task: %v", err)
	}
	if _, err := client.DeleteOnboardingPlan(ctx, &common.IDRequest{Id: planID}); err != nil {
		t.Fatalf("delete unused draft plan: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", planID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("plan outbox after create/update/delete: count=%d, error=%v", events, err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	startCode := "START-" + workerID
	startPlan, err := client.CreateOnboardingPlan(ctx, &pb.CreateOnboardingPlanRequest{Input: &pb.OnboardingPlanInput{
		WorkerId: &workerID, PlanCode: &startCode, StartDate: &today,
	}})
	if err != nil {
		t.Fatalf("create plan for start: %v", err)
	}
	startedPlanID = startPlan.GetItem().GetId()
	if _, err := client.StartOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start without tasks: %v", err)
	}
	startTaskCode := "START-ORIENTATION"
	startTask, err := client.CreateOnboardingTask(ctx, &pb.CreateOnboardingTaskRequest{Input: &pb.OnboardingTaskInput{
		OnboardingPlanId: &startedPlanID, TaskCode: &startTaskCode, TaskName: &name, TaskOwnerType: &owner,
		AssignedPrincipalId: &principal,
	}})
	if err != nil {
		t.Fatalf("create task for start: %v", err)
	}
	startedTaskID = startTask.GetItem().GetId()
	started, err := client.StartOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID})
	if err != nil || started.GetItem().GetStatus() != "in_progress" || started.GetItem().GetStartedAt() == nil ||
		started.GetItem().GetStartedByPrincipalId() != principal {
		t.Fatalf("start plan: %+v, %v", started, err)
	}
	if _, err := client.StartOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat start: %v", err)
	}
	if _, err := client.CompleteOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete with pending mandatory task: %v", err)
	}
	// Verifier tiruan membuktikan alur HRM; integrasi DRM tetap perlu diuji sendiri.
	finishedTask, err := client.CompleteOnboardingTask(ctx, &pb.CompleteOnboardingTaskRequest{
		Id: startedTaskID, EvidenceDocumentVersionReferenceId: workerID,
	})
	if err != nil || finishedTask.GetItem().GetCompletedAt() == nil ||
		finishedTask.GetItem().GetCompletedByPrincipalId() != principal ||
		finishedTask.GetItem().GetCompletionEvidenceReferenceId() != workerID {
		t.Fatalf("complete task: %+v, %v", finishedTask, err)
	}
	if _, err := client.CompleteOnboardingTask(ctx, &pb.CompleteOnboardingTaskRequest{
		Id: startedTaskID, EvidenceDocumentVersionReferenceId: workerID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat task completion: %v", err)
	}
	completed, err := client.CompleteOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID})
	if err != nil || completed.GetItem().GetStatus() != "completed" || completed.GetItem().GetCompletedAt() == nil ||
		completed.GetItem().GetCompletedByPrincipalId() != principal {
		t.Fatalf("complete plan after mandatory task: %+v, %v", completed, err)
	}
	if _, err := client.CompleteOnboardingPlan(ctx, &common.IDRequest{Id: startedPlanID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat completion: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", startedPlanID).Count(&events).Error; err != nil || events != 3 {
		t.Fatalf("start and complete plan outbox count=%d, error=%v", events, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", startedTaskID).Count(&events).Error; err != nil || events != 2 {
		t.Fatalf("task create and complete outbox count=%d, error=%v", events, err)
	}
}
