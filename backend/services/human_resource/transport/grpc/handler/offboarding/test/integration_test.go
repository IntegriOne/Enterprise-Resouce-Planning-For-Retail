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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/offboarding"
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
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestOffboardingDraftRPCWithPostgreSQL(t *testing.T) {
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
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.OffboardingValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: integrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		OffboardingPlan: repos.OffboardingPlan, OffboardingTask: repos.OffboardingTask,
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
	connection, err := grpc.NewClient("passthrough:///hrm-offboarding-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewOffboardingServiceClient(connection)
	ids := operation.UUIDGenerator{}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var planID, taskID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, value string }{
			{"DELETE FROM hr_offboarding_tasks WHERE id = ?", taskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", taskID},
			{"DELETE FROM hr_offboarding_plans WHERE id = ?", planID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if statement.value != "" {
				if err := database.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
					t.Errorf("cleanup HRM offboarding test data: %v", err)
				}
			}
		}
	})
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
		workerID, "OFFBOARD-"+workerID, workerID, "employee", "active", workerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	number, reason := "EXIT-"+workerID, "resignation"
	notice := timestamppb.New(time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC))
	last := timestamppb.New(time.Date(2027, 1, 17, 0, 0, 0, 0, time.UTC))
	effective := timestamppb.New(time.Date(2027, 1, 18, 0, 0, 0, 0, time.UTC))
	planInput := &pb.OffboardingPlanInput{WorkerId: &workerID, PlanNumber: &number, ReasonCode: &reason,
		NoticeDate: notice, LastWorkingDate: last, TerminationEffectiveDate: effective}
	created, err := client.CreateOffboardingPlan(ctx, &pb.CreateOffboardingPlanRequest{Input: planInput})
	if err != nil || created.GetValue().GetId() == "" || created.GetValue().GetStatus() != "draft" {
		t.Fatalf("create draft plan: %+v, %v", created, err)
	}
	planID = created.GetValue().GetId()
	if _, err := client.CreateOffboardingPlan(ctx, &pb.CreateOffboardingPlanRequest{Input: planInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("second open plan: %v", err)
	}
	if _, err := client.GetOffboardingPlanByID(ctx, &common.IDRequest{Id: planID}); err != nil {
		t.Fatalf("read draft plan: %v", err)
	}
	listed, err := client.FindOffboardingPlanByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1})
	if err != nil || len(listed.GetItems()) != 1 || listed.GetPage().GetTotal() != 1 {
		t.Fatalf("paged plan finder: %+v, %v", listed, err)
	}
	if _, err := client.GetAllOffboardingPlans(ctx, &common.PageRequest{Limit: 1}); err != nil {
		t.Fatalf("paged plan list: %v", err)
	}
	revisedReason := "contract_end"
	if _, err := client.UpdateOffboardingPlan(ctx, &pb.UpdateOffboardingPlanRequest{Id: planID,
		Input: &pb.OffboardingPlanInput{ReasonCode: &revisedReason}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"reason_code"}},
	}); err != nil {
		t.Fatalf("update unused draft plan: %v", err)
	}
	payrollRunID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateOffboardingPlan(ctx, &pb.UpdateOffboardingPlanRequest{Id: planID,
		Input: &pb.OffboardingPlanInput{FinalPayrollRunId: &payrollRunID}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"final_payroll_run_id"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unverified final payroll: %v", err)
	}
	taskCode, taskName, owner := "RETURN-BADGE", "Return badge", "human_resource"
	taskInput := &pb.OffboardingTaskInput{OffboardingPlanId: &planID, TaskCode: &taskCode, TaskName: &taskName, OwnerManagement: &owner}
	taskCreated, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: taskInput})
	if err != nil || taskCreated.GetValue().GetId() == "" || taskCreated.GetValue().GetStatus() != "pending" {
		t.Fatalf("create pending task: %+v, %v", taskCreated, err)
	}
	taskID = taskCreated.GetValue().GetId()
	if _, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: taskInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate task code: %v", err)
	}
	otherOwner := "finance_accounting"
	if _, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: &pb.OffboardingTaskInput{
		OffboardingPlanId: &planID, TaskCode: &taskCode, TaskName: &taskName, OwnerManagement: &otherOwner,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("external owner without contract: %v", err)
	}
	if _, err := client.DeleteOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete plan with task: %v", err)
	}
	tasks, err := client.FindOffboardingTaskByOffboardingPlanID(ctx, &common.StringFinderRequest{Value: &planID, Limit: 1})
	if err != nil || len(tasks.GetItems()) != 1 || tasks.GetPage().GetTotal() != 1 {
		t.Fatalf("paged task finder: %+v, %v", tasks, err)
	}
	newName := "Return store badge"
	if _, err := client.UpdateOffboardingTask(ctx, &pb.UpdateOffboardingTaskRequest{Id: taskID,
		Input: &pb.OffboardingTaskInput{TaskName: &newName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"task_name"}},
	}); err != nil {
		t.Fatalf("update pending task: %v", err)
	}
	var eventCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", taskID).Count(&eventCount).Error; err != nil || eventCount != 2 {
		t.Fatalf("task outbox count=%d error=%v", eventCount, err)
	}
	if _, err := client.DeleteOffboardingTask(ctx, &common.IDRequest{Id: taskID}); err != nil {
		t.Fatalf("delete pending task: %v", err)
	}
	if _, err := client.DeleteOffboardingPlan(ctx, &common.IDRequest{Id: planID}); err != nil {
		t.Fatalf("delete draft plan: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", planID).Count(&eventCount).Error; err != nil || eventCount != 3 {
		t.Fatalf("plan outbox count=%d error=%v", eventCount, err)
	}
}
