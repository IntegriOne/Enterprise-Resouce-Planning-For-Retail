//go:build integration

package test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	personnelpb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/offboarding"
	personnelhandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/personnel"
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

type integrationPolicy struct{ independent atomic.Bool }

func (p *integrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	principalID := "00000000-0000-4000-8000-000000000001"
	if p.independent.Load() {
		principalID = "00000000-0000-4000-8000-000000000003"
	}
	return operation.Actor{PrincipalID: principalID, SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type confirmationStub struct{ calls atomic.Int32 }

func (s *confirmationStub) VerifyOffboardingConfirmation(_ context.Context, request service.ConfirmationRequest) error {
	s.calls.Add(1)
	if request.ReferenceID != request.TaskID || request.PlanID == "" ||
		request.OwnerManagement != "human_resource" || request.PrincipalID == "" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

type clearanceStub struct {
	calls       atomic.Int32
	denyIAM     atomic.Bool
	denyFinance atomic.Bool
}

func (s *clearanceStub) VerifyIAMClearance(_ context.Context, request service.IAMClearanceRequest) error {
	if s.denyIAM.Load() {
		return hrerrors.ErrForbidden
	}
	if request.PlanID == "" || request.WorkerID == "" || request.WorkerPrincipalID != request.PlanID || request.TerminationEffectiveDate.IsZero() {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (s *clearanceStub) VerifyFinanceClearance(_ context.Context, request service.FinanceClearanceRequest) error {
	s.calls.Add(1)
	if s.denyFinance.Load() {
		return hrerrors.ErrUnavailable
	}
	if request.FinanceClearanceReferenceID != request.WorkerID || request.FinalPayrollRunID != request.WorkerID ||
		request.PlanID == "" ||
		request.TerminationEffectiveDate.IsZero() ||
		request.FinalPayrollResultID != request.WorkerID || request.EmployingLegalEntityID != request.WorkerID ||
		request.CurrencyCode != "IDR" || request.PaymentReferenceID != request.WorkerID || !bytes.Equal(request.RunOutputDigest, bytes.Repeat([]byte{0x12}, 32)) ||
		!bytes.Equal(request.ResultCalculationDigest, bytes.Repeat([]byte{0x12}, 32)) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
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
	policy := &integrationPolicy{}
	withPayroll := strings.HasPrefix(strings.TrimPrefix(parsed.Path, "/"), "hrm_disposable_") &&
		os.Getenv("HRM_TEST_DATABASE_DISPOSABLE") == "1"
	var clearance service.ClearanceVerifier
	confirmation := &confirmationStub{}
	clearanceRecorder := &clearanceStub{}
	if withPayroll {
		clearance, err = service.NewOwnerClearance(clearanceRecorder, clearanceRecorder)
		if err != nil {
			t.Fatal(err)
		}
	}
	registry, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: policy, Integrity: security.NewIntegrity(validators),
	}, appservice.WithOffboardingVerifiers(appservice.OffboardingVerifiers{
		Confirmation: confirmation, Clearance: clearance,
	}))
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, registry.Offboarding); err != nil {
		t.Fatal(err)
	}
	if err := personnelhandler.Register(server, registry.Personnel); err != nil {
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
	var planID, taskID, completionTaskID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct{ query, value string }{
			{"DELETE FROM hr_offboarding_tasks WHERE id = ?", taskID},
			{"DELETE FROM hr_offboarding_tasks WHERE id = ?", completionTaskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", completionTaskID},
			{"DELETE FROM hr_worker_assignment_transfers WHERE id = ?", completionTaskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", taskID},
			{"DELETE FROM hr_worker_lifecycle_events WHERE worker_id = ?", workerID},
			{"DELETE FROM hr_offboarding_plans WHERE id = ?", planID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
		} {
			if statement.value != "" {
				if err := database.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
					t.Errorf("cleanup HRM offboarding test data: %v", err)
				}
			}
		}
		if withPayroll {
			if planID != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_positions WHERE id = ?", planID).Error; err != nil {
					t.Errorf("cleanup transfer target: %v", err)
				}
			}
			cleanupTerminationPayroll(t, database.WithContext(cleanupCtx), workerID)
			for _, query := range []string{
				"UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?",
				"DELETE FROM hr_worker_assignments WHERE worker_id = ?",
				"DELETE FROM hr_employment_contracts WHERE worker_id = ?",
				"DELETE FROM hr_positions WHERE id = ?",
				"DELETE FROM hr_jobs WHERE id = ?",
				"DELETE FROM hr_job_families WHERE id = ?",
			} {
				if err := database.WithContext(cleanupCtx).Exec(query, workerID).Error; err != nil {
					t.Errorf("cleanup termination relationships: %v", err)
				}
			}
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", workerID).Error; err != nil {
			t.Errorf("cleanup offboarding worker: %v", err)
		}
	})
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?)",
		workerID, "OFFBOARD-"+workerID, workerID, "employee", "active", workerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	number, reason := "EXIT-"+workerID, "resignation"
	now := time.Now().UTC().Truncate(24 * time.Hour)
	notice := timestamppb.New(now)
	last := timestamppb.New(now)
	effective := timestamppb.New(now)
	planInput := &pb.OffboardingPlanInput{WorkerId: &workerID, PlanNumber: &number, ReasonCode: &reason,
		NoticeDate: notice, LastWorkingDate: last, TerminationEffectiveDate: effective}
	created, err := client.CreateOffboardingPlan(ctx, &pb.CreateOffboardingPlanRequest{Input: planInput})
	if err != nil || created.GetValue().GetId() == "" || created.GetValue().GetStatus() != "draft" {
		t.Fatalf("create draft plan: %+v, %v", created, err)
	}
	planID = created.GetValue().GetId()
	policy.independent.Store(true)
	if _, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve plan without tasks = %v, want FailedPrecondition", err)
	}
	policy.independent.Store(false)
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
	completionCode := "RETURN-KEY"
	completionInput := &pb.OffboardingTaskInput{
		OffboardingPlanId: &planID, TaskCode: &completionCode, TaskName: &completionCode, OwnerManagement: &owner,
	}
	completionTask, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: completionInput})
	if err != nil {
		t.Fatalf("create completion task: %v", err)
	}
	completionTaskID = completionTask.GetValue().GetId()
	if _, err := client.CompleteOffboardingTask(ctx, &pb.CompleteOffboardingTaskRequest{
		Id: completionTaskID, OwnerConfirmationReferenceId: completionTaskID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete task before plan start: %v", err)
	}
	if calls := confirmation.calls.Load(); calls != 0 {
		t.Fatalf("confirmation verifier called before plan start: %d", calls)
	}
	if _, err := client.StartOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start before approval = %v", err)
	}
	if _, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval = %v, want PermissionDenied", err)
	}
	policy.independent.Store(true)
	const independentPrincipalID = "00000000-0000-4000-8000-000000000003"
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", independentPrincipalID).Error; err != nil {
		t.Fatalf("bind plan subject to approval actor: %v", err)
	}
	if _, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("subject approval = %v, want PermissionDenied", err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", nil).Error; err != nil {
		t.Fatalf("restore plan subject fixture: %v", err)
	}
	approved, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: planID})
	if err != nil || approved.GetValue().GetStatus() != "approved" ||
		approved.GetValue().GetApprovedAt() == nil || approved.GetValue().GetApprovedByPrincipalId() != independentPrincipalID {
		t.Fatalf("approve draft plan: %+v, %v", approved, err)
	}
	if _, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat approval = %v, want FailedPrecondition", err)
	}
	started, err := client.StartOffboardingPlan(ctx, &common.IDRequest{Id: planID})
	if err != nil || started.GetValue().GetStatus() != "in_progress" ||
		started.GetValue().GetApprovedByPrincipalId() != approved.GetValue().GetApprovedByPrincipalId() {
		t.Fatalf("start plan: %+v, %v", started, err)
	}
	if _, err := client.StartOffboardingPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat start = %v", err)
	}
	if _, err := client.DeleteOffboardingTask(ctx, &common.IDRequest{Id: taskID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete task after start = %v", err)
	}
	completeRequest := &pb.CompleteOffboardingTaskRequest{Id: completionTaskID, OwnerConfirmationReferenceId: taskID}
	if _, err := client.CompleteOffboardingTask(ctx, completeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unverified confirmation = %v", err)
	}
	completeRequest.OwnerConfirmationReferenceId = completionTaskID
	if err := database.WithContext(ctx).Table("hr_offboarding_plans").Where("id = ?", planID).
		Update("approved_at", time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatalf("set future plan approval fixture: %v", err)
	}
	if _, err := client.CompleteOffboardingTask(ctx, completeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("complete task before plan approval: %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_offboarding_plans").Where("id = ?", planID).
		Update("approved_at", approved.GetValue().GetApprovedAt().AsTime()).Error; err != nil {
		t.Fatalf("restore plan approval fixture: %v", err)
	}
	if withPayroll {
		if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
			Updates(map[string]any{"status": "terminated", "termination_date": now.AddDate(0, 0, -1)}).Error; err != nil {
			t.Fatalf("set mismatched termination for task completion: %v", err)
		}
		if _, err := client.CompleteOffboardingTask(ctx, completeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("complete task with mismatched termination date: %v", err)
		}
		if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
			Update("termination_date", now).Error; err != nil {
			t.Fatalf("align worker termination date for task completion: %v", err)
		}
	}
	completed, err := client.CompleteOffboardingTask(ctx, completeRequest)
	if err != nil || completed.GetValue().GetStatus() != "completed" || completed.GetValue().GetCompletedAt() == nil ||
		completed.GetValue().GetOwnerConfirmationReferenceId() != completionTaskID {
		t.Fatalf("complete task: %+v, %v", completed, err)
	}
	var completionEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", completionTaskID, "hr_offboarding_tasks.complete").
		Where("payload->>'actor_principal_id' = ?", approved.GetValue().GetApprovedByPrincipalId()).
		Where("payload->'details'->>'offboarding_plan_id' = ?", planID).
		Where("payload->'details'->>'owner_management' = ?", "human_resource").
		Where("payload->'details'->>'owner_confirmation_reference_id' = ?", completed.GetValue().GetOwnerConfirmationReferenceId()).
		Count(&completionEvents).Error; err != nil || completionEvents != 1 {
		t.Fatalf("verified task completion events=%d error=%v", completionEvents, err)
	}

	if withPayroll {
		if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
			Updates(map[string]any{"status": "active", "termination_date": nil}).Error; err != nil {
			t.Fatalf("restore active worker fixture for subsequent task cases: %v", err)
		}
	}
	waiveRequest := &pb.WaiveOffboardingTaskRequest{Id: taskID, Reason: " "}
	if _, err := client.WaiveOffboardingTask(ctx, waiveRequest); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("blank waiver reason = %v", err)
	}
	waiveRequest.Reason = "Badge already disabled"
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", approved.GetValue().GetApprovedByPrincipalId()).Error; err != nil {
		t.Fatalf("bind subject to waiver actor in fixture: %v", err)
	}
	if _, err := client.WaiveOffboardingTask(ctx, waiveRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self-waiver = %v, want PermissionDenied", err)
	}
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", workerID).
		Update("iam_principal_reference_id", nil).Error; err != nil {
		t.Fatalf("restore subject principal fixture: %v", err)
	}
	waived, err := client.WaiveOffboardingTask(ctx, waiveRequest)
	if err != nil || waived.GetValue().GetStatus() != "waived" ||
		waived.GetValue().GetWaiverReason() != waiveRequest.Reason ||
		waived.GetValue().GetWaivedByPrincipalId() != approved.GetValue().GetApprovedByPrincipalId() {
		t.Fatalf("waive task: %+v, %v", waived, err)
	}
	if _, err := client.WaiveOffboardingTask(ctx, waiveRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat waiver = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", taskID).Count(&eventCount).Error; err != nil || eventCount != 3 {
		t.Fatalf("task outbox count after waiver=%d error=%v", eventCount, err)
	}
	expectedEvents := int64(4)
	closeRequest := &pb.CompleteOffboardingPlanRequest{Id: planID, FinalPayrollRunId: workerID, FinanceClearanceReferenceId: workerID}
	if withPayroll {
		seedTerminationPayroll(t, ctx, database, workerID)
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition || clearanceRecorder.calls.Load() != 0 {
			t.Fatalf("worker not terminated must fail without checking external clearance: %v", err)
		}
		personnel := personnelpb.NewPersonnelServiceClient(connection)
		if _, err := personnel.LinkWorkerPrincipal(ctx, &personnelpb.LinkWorkerPrincipalRequest{
			WorkerId: workerID, PrincipalId: workerID, EvidenceId: planID,
		}); status.Code(err) != codes.Unavailable {
			t.Fatalf("principal link without production verifier: %v", err)
		}
		var linkedWorkers int64
		if err := database.WithContext(ctx).Table("hr_workers").
			Where("id = ? AND iam_principal_reference_id IS NOT NULL", workerID).Count(&linkedWorkers).Error; err != nil || linkedWorkers != 0 {
			t.Fatalf("missing verifier allowed worker link: count=%d, err=%v", linkedWorkers, err)
		}

		termination := &personnelpb.TerminateWorkerRequest{WorkerId: workerID, OffboardingPlanId: planID}
		seedTerminationRelationships(t, ctx, database, workerID, planID, now)
		if err := database.WithContext(ctx).Exec(
			"INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, position_status, valid_from) VALUES (?, ?, 'Associate', ?, ?, ?, 'active', ?)",
			planID, "TARGET-"+planID, workerID, workerID, workerID, now.AddDate(0, 0, -2),
		).Error; err != nil {
			t.Fatalf("seed transfer target: %v", err)
		}
		if err := database.WithContext(ctx).Exec(
			"INSERT INTO hr_worker_assignment_transfers (id, worker_id, source_assignment_id, target_position_id, effective_date, reason_code, idempotency_key, requested_by_principal_id, requested_at) VALUES (?, ?, ?, ?, ?, 'TEAM_CHANGE', ?, ?, ?)",
			completionTaskID, workerID, workerID, planID, now, "TRANSFER-"+completionTaskID, workerID, now,
		).Error; err != nil {
			t.Fatalf("seed pending transfer: %v", err)
		}

		if err := database.WithContext(ctx).Table("hr_worker_assignments").Where("id = ?", planID).
			Updates(map[string]any{"valid_from": now.AddDate(0, 0, 1), "approved_at": now.AddDate(0, 0, 1)}).Error; err != nil {
			t.Fatalf("set future assignment fixture: %v", err)
		}
		if _, err := personnel.TerminateWorker(ctx, termination); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("termination with future approval timestamp: %v", err)
		}
		var openAssignments int64
		if err := database.WithContext(ctx).Table("hr_worker_assignments").
			Where("worker_id = ? AND valid_to IS NULL", workerID).Count(&openAssignments).Error; err != nil || openAssignments != 2 {
			t.Fatalf("failed termination did not roll back assignments: count=%d, err=%v", openAssignments, err)
		}
		if err := database.WithContext(ctx).Table("hr_worker_assignments").Where("id = ?", planID).
			Update("approved_at", now.AddDate(0, 0, -2)).Error; err != nil {
			t.Fatalf("align assignment fixture: %v", err)
		}
		terminated, err := personnel.TerminateWorker(ctx, termination)
		if err != nil || terminated.GetItem().GetStatus() != "terminated" ||
			terminated.GetItem().GetTerminationReasonCode() != approved.GetValue().GetReasonCode() ||
			terminated.GetItem().GetTerminationDate() != now.Format("2006-01-02") {
			t.Fatalf("terminate worker through Personnel: %+v, %v", terminated, err)
		}
		if terminated.GetItem().GetPrimaryAssignmentId() != "" {
			t.Fatal("termination retained primary assignment")
		}
		var matchingTerminationEvents int64
		if err := database.WithContext(ctx).Table("hr_outbox_events").
			Where("aggregate_id = ? AND event_type = ?", planID, "hr_worker_terminations.recorded").
			Where("payload->>'worker_id' = ? AND payload->>'offboarding_plan_id' = ?", workerID, planID).
			Where("payload->>'termination_effective_date' = ?", terminated.GetItem().GetTerminationDate()).
			Count(&matchingTerminationEvents).Error; err != nil || matchingTerminationEvents != 1 {
			t.Fatalf("termination event must identify the worker, plan, and effective date: count=%d, err=%v", matchingTerminationEvents, err)
		}
		var closedAssignments, closedContracts int64
		if err := database.WithContext(ctx).Table("hr_worker_assignments").
			Where("worker_id = ? AND valid_to = ? AND approved_at IS NOT NULL", workerID, now).
			Count(&closedAssignments).Error; err != nil || closedAssignments != 1 {
			t.Fatalf("closing approved assignments: count=%d, err=%v", closedAssignments, err)
		}
		if err := database.WithContext(ctx).Table("hr_employment_contracts").
			Where("worker_id = ? AND effective_to = ? AND signed_content_digest = ?", workerID, now, []byte{1, 2, 3}).
			Count(&closedContracts).Error; err != nil || closedContracts != 1 {
			t.Fatalf("closing signed contract: count=%d, err=%v", closedContracts, err)
		}
		cancelledAssignment, err := personnel.GetWorkerAssignmentByID(ctx, &common.IDRequest{Id: planID})
		if err != nil || cancelledAssignment.GetItem().GetCancelledAt() == nil ||
			cancelledAssignment.GetItem().GetApprovedAt() == nil ||
			cancelledAssignment.GetItem().GetValidFrom() != now.AddDate(0, 0, 1).Format("2006-01-02") ||
			cancelledAssignment.GetItem().GetValidTo() != "" {
			t.Fatalf("cancel future assignment and retain original approval and dates: %+v, %v", cancelledAssignment, err)
		}
		var cancelledContracts int64
		if err := database.WithContext(ctx).Table("hr_employment_contracts").
			Where("id = ? AND contract_status = ? AND effective_from = ? AND effective_to IS NULL AND signed_content_digest = ?",
				planID, "cancelled", now.AddDate(0, 0, 1), []byte{4, 5, 6}).
			Count(&cancelledContracts).Error; err != nil || cancelledContracts != 1 {
			t.Fatalf("cancelling future signed contract: count=%d, err=%v", cancelledContracts, err)
		}
		if _, err := personnel.TerminateWorker(ctx, termination); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("repeat termination: %v", err)
		}
		transferDecision := &personnelpb.DecideWorkerAssignmentTransferRequest{
			TransferId: completionTaskID, Decision: personnelpb.TransferDecision_TRANSFER_DECISION_APPROVE,
		}
		if _, err := personnel.DecideWorkerAssignmentTransfer(ctx, transferDecision); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("approve transfer after termination: %v", err)
		}
		transferDecision.Decision = personnelpb.TransferDecision_TRANSFER_DECISION_REJECT
		rejected, err := personnel.DecideWorkerAssignmentTransfer(ctx, transferDecision)
		if err != nil || rejected.GetItem().GetStatus() != "rejected" || rejected.GetItem().GetAppliedAssignmentId() != "" {
			t.Fatalf("reject pending transfer after termination: %+v, %v", rejected, err)
		}
		expectedEvents++
		setTerminationResultFixtureStatus(t, ctx, database, workerID, "paid")
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("close with paid result in approved payroll run: %v", err)
		}
		setTerminationResultFixtureStatus(t, ctx, database, workerID, "cancelled")
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("close with cancelled payroll result: %v", err)
		}
		setTerminationResultFixtureStatus(t, ctx, database, workerID, "calculated")
		if err := database.WithContext(ctx).Table("hr_offboarding_tasks").Where("id = ?", completionTaskID).
			Update("completed_at", approved.GetValue().GetApprovedAt().AsTime().Add(-time.Second)).Error; err != nil {
			t.Fatalf("set task completion before plan approval: %v", err)
		}
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("close with task completed before plan approval: %v", err)
		}
		pending, err := client.GetOffboardingPlanByID(ctx, &common.IDRequest{Id: planID})
		if err != nil || pending.GetValue().GetStatus() != "in_progress" ||
			pending.GetValue().GetCompletedAt() != nil || pending.GetValue().GetFinalPayrollRunId() != "" {
			t.Fatalf("failed close changed plan: %+v, %v", pending, err)
		}
		if err := database.WithContext(ctx).Table("hr_offboarding_tasks").Where("id = ?", completionTaskID).
			Update("completed_at", completed.GetValue().GetCompletedAt().AsTime()).Error; err != nil {
			t.Fatalf("restore task completion time: %v", err)
		}
		if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_runs SET status = 'posted', finance_journal_reference_id = ? WHERE id = ?", workerID, workerID).Error; err != nil {
			t.Fatal(err)
		}
		if err := database.WithContext(ctx).Exec("UPDATE hr_payroll_results SET result_status = 'paid', payment_reference_id = ? WHERE id = ?", workerID, workerID).Error; err != nil {
			t.Fatal(err)
		}
		unlinked, err := registry.Personnel.VerifyTerminationAccess(ctx, workerID, planID, nil)
		if err != nil || unlinked.MayDisable || unlinked.PrincipalID != "" {
			t.Fatalf("termination without linked account: %+v, %v", unlinked, err)
		}
		if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET iam_principal_reference_id = ? WHERE id = ?", planID, workerID).Error; err != nil {
			t.Fatal(err)
		}
		decision, err := registry.Personnel.VerifyTerminationAccess(ctx, workerID, planID, &now)
		if err != nil || !decision.MayDisable || decision.PrincipalID != planID {
			t.Fatalf("termination account closure decision: %+v, %v", decision, err)
		}
		wrongDate := now.AddDate(0, 0, -1)
		if _, err := registry.Personnel.VerifyTerminationAccess(ctx, workerID, planID, &wrongDate); !errors.Is(err, hrerrors.ErrReferenceConflict) {
			t.Fatalf("mismatched termination date: %v", err)
		}
		if err := database.WithContext(ctx).Exec(
			"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id, iam_principal_reference_id) VALUES (?, ?, ?, 'employee', 'active', ?, ?)",
			completionTaskID, "REHIRED-"+completionTaskID, workerID, workerID, planID,
		).Error; err != nil {
			t.Fatal(err)
		}
		stale, err := registry.Personnel.VerifyTerminationAccess(ctx, workerID, planID, &now)
		if err != nil || stale.MayDisable {
			t.Fatalf("stale termination must preserve a current employment: %+v, %v", stale, err)
		}
		if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", completionTaskID).Delete(nil).Error; err != nil {
			t.Fatal(err)
		}
		closeRequest.FinanceClearanceReferenceId = planID
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("unverified finance clearance = %v", err)
		}
		closeRequest.FinanceClearanceReferenceId = workerID
		financeCalls := clearanceRecorder.calls.Load()
		clearanceRecorder.denyIAM.Store(true)
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.PermissionDenied ||
			clearanceRecorder.calls.Load() != financeCalls {
			t.Fatalf("IAM refusal must stop before Finance: %v", err)
		}
		clearanceRecorder.denyIAM.Store(false)
		clearanceRecorder.denyFinance.Store(true)
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.Unavailable ||
			clearanceRecorder.calls.Load() != financeCalls+1 {
			t.Fatalf("Finance outage must prevent completion: %v", err)
		}
		clearanceRecorder.denyFinance.Store(false)
		closed, err := client.CompleteOffboardingPlan(ctx, closeRequest)
		if err != nil || closed.GetValue().GetStatus() != "completed" || closed.GetValue().GetCompletedAt() == nil ||
			closed.GetValue().GetFinalPayrollRunId() != workerID || closed.GetValue().GetFinanceClearanceReferenceId() != workerID {
			t.Fatalf("close plan: %+v, %v", closed, err)
		}
		stored, err := client.GetOffboardingPlanByID(ctx, &common.IDRequest{Id: planID})
		if err != nil || stored.GetValue().GetStatus() != "completed" || stored.GetValue().GetFinalPayrollRunId() != workerID {
			t.Fatalf("read closed plan: %+v, %v", stored, err)
		}
		if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("repeat close = %v", err)
		}
		expectedEvents++
	} else if _, err := client.CompleteOffboardingPlan(ctx, closeRequest); status.Code(err) != codes.Unavailable {
		t.Fatalf("complete without clearance verifier = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", planID).Count(&eventCount).Error; err != nil || eventCount != expectedEvents {
		t.Fatalf("plan outbox count=%d error=%v", eventCount, err)
	}
	testCancelApprovedPlan(t, ctx, database, client, policy, now)
}

func seedTerminationRelationships(t *testing.T, ctx context.Context, database *gorm.DB, workerID, secondaryID string, lastDay time.Time) {
	t.Helper()
	start := lastDay.AddDate(0, 0, -2)
	approverID := "00000000-0000-4000-8000-000000000003"
	for _, fixture := range []struct {
		query string
		args  []any
	}{
		{"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, 'Store', 'operations')", []any{workerID, "FAMILY-" + workerID}},
		{"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, 'Associate', ?)", []any{workerID, workerID, "JOB-" + workerID, start}},
		{"INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id, legal_entity_reference_id, valid_from) VALUES (?, ?, 'Associate', ?, ?, ?, ?)", []any{workerID, "POSITION-" + workerID, workerID, workerID, workerID, start}},
		{"INSERT INTO hr_worker_assignments (id, worker_id, position_id, assignment_type, fte_percent, valid_from, requested_by_principal_id, requested_at, approved_by_principal_id, approved_at) VALUES (?, ?, ?, 'primary', 75, ?, ?, ?, ?, ?), (?, ?, ?, 'secondary', 25, ?, ?, ?, ?, ?)", []any{workerID, workerID, workerID, start, workerID, start, approverID, start, secondaryID, workerID, workerID, start, workerID, start, approverID, start}},
		{"UPDATE hr_workers SET primary_assignment_id = ? WHERE id = ?", []any{workerID, workerID}},
		{"INSERT INTO hr_employment_contracts (id, worker_id, contract_number, contract_type, contract_status, employing_legal_entity_reference_id, effective_from, signed_at, signed_document_reference_id, signed_document_version_reference_id, signed_content_digest, work_country_code, governing_jurisdiction_code) VALUES (?, ?, ?, 'permanent', 'signed', ?, ?, ?, ?, ?, ?, 'ID', 'ID')", []any{workerID, workerID, "CONTRACT-" + workerID, workerID, start, start, workerID, workerID, []byte{1, 2, 3}}},
		{"INSERT INTO hr_employment_contracts (id, worker_id, contract_number, contract_type, contract_status, employing_legal_entity_reference_id, effective_from, signed_at, signed_document_reference_id, signed_document_version_reference_id, signed_content_digest, work_country_code, governing_jurisdiction_code) VALUES (?, ?, ?, 'permanent', 'signed', ?, ?, ?, ?, ?, ?, 'ID', 'ID')", []any{secondaryID, workerID, "CONTRACT-" + secondaryID, workerID, lastDay.AddDate(0, 0, 1), start, secondaryID, secondaryID, []byte{4, 5, 6}}},
	} {
		if err := database.WithContext(ctx).Exec(fixture.query, fixture.args...).Error; err != nil {
			t.Fatalf("seed termination relationships: %v", err)
		}
	}
}

func testCancelApprovedPlan(t *testing.T, ctx context.Context, database *gorm.DB, client pb.OffboardingServiceClient, policy *integrationPolicy, now time.Time) {
	t.Helper()
	workerID, err := (operation.UUIDGenerator{}).NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id) VALUES (?, ?, ?, 'employee', 'active', ?)",
		workerID, "CANCEL-"+workerID, workerID, workerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	var oldID, newID, taskID, replacementTaskID string
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, stmt := range []struct{ query, id string }{
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", taskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", replacementTaskID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", oldID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", newID},
			{"DELETE FROM hr_offboarding_tasks WHERE id = ?", taskID},
			{"DELETE FROM hr_offboarding_tasks WHERE id = ?", replacementTaskID},
			{"DELETE FROM hr_offboarding_plans WHERE id = ?", oldID},
			{"DELETE FROM hr_offboarding_plans WHERE id = ?", newID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.id != "" {
				if err := database.WithContext(cleanup).Exec(stmt.query, stmt.id).Error; err != nil {
					t.Errorf("cleanup cancelled plan fixtures: %v", err)
				}
			}
		}
	})
	policy.independent.Store(false)
	number, reason := "CANCEL-"+workerID, "contract_end"
	date := timestamppb.New(now)
	input := &pb.OffboardingPlanInput{WorkerId: &workerID, PlanNumber: &number, ReasonCode: &reason,
		NoticeDate: date, LastWorkingDate: date, TerminationEffectiveDate: date}
	created, err := client.CreateOffboardingPlan(ctx, &pb.CreateOffboardingPlanRequest{Input: input})
	if err != nil {
		t.Fatal(err)
	}
	oldID = created.GetValue().GetId()
	code, owner := "RETURN-BADGE", "human_resource"
	task, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: &pb.OffboardingTaskInput{
		OffboardingPlanId: &oldID, TaskCode: &code, TaskName: &code, OwnerManagement: &owner,
	}})
	if err != nil {
		t.Fatal(err)
	}
	taskID = task.GetValue().GetId()
	cancelRequest := &pb.CancelOffboardingPlanRequest{Id: oldID, Reason: "Contract extension"}
	if _, err := client.CancelOffboardingPlan(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel draft offboarding plan: %v", err)
	}
	policy.independent.Store(true)
	approved, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: oldID})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := client.CancelOffboardingPlan(ctx, cancelRequest)
	if err != nil || cancelled.GetValue().GetStatus() != "cancelled" ||
		cancelled.GetValue().GetCancellationReason() != cancelRequest.GetReason() ||
		cancelled.GetValue().GetCancelledAt() == nil ||
		cancelled.GetValue().GetApprovedByPrincipalId() != approved.GetValue().GetApprovedByPrincipalId() {
		t.Fatalf("cancel approved plan: %+v, %v", cancelled, err)
	}
	if _, err := client.StartOffboardingPlan(ctx, &common.IDRequest{Id: oldID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("start cancelled plan: %v", err)
	}
	number = "REPLACE-" + workerID
	replacement, err := client.CreateOffboardingPlan(ctx, &pb.CreateOffboardingPlanRequest{Input: input})
	if err != nil || replacement.GetValue().GetStatus() != "draft" {
		t.Fatalf("create replacement plan: %+v, %v", replacement, err)
	}
	newID = replacement.GetValue().GetId()
	replacementTask, err := client.CreateOffboardingTask(ctx, &pb.CreateOffboardingTaskRequest{Input: &pb.OffboardingTaskInput{
		OffboardingPlanId: &newID, TaskCode: &code, TaskName: &code, OwnerManagement: &owner,
	}})
	if err != nil {
		t.Fatal(err)
	}
	replacementTaskID = replacementTask.GetValue().GetId()
	policy.independent.Store(false)
	if _, err := client.ApproveOffboardingPlan(ctx, &common.IDRequest{Id: newID}); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	outcomes := make(chan codes.Code, 2)
	go func() {
		<-start
		_, err := client.StartOffboardingPlan(ctx, &common.IDRequest{Id: newID})
		outcomes <- status.Code(err)
	}()
	go func() {
		<-start
		_, err := client.CancelOffboardingPlan(ctx, &pb.CancelOffboardingPlanRequest{Id: newID, Reason: "Revised decision"})
		outcomes <- status.Code(err)
	}()
	close(start)
	first, second := <-outcomes, <-outcomes
	if !((first == codes.OK && second == codes.FailedPrecondition) ||
		(first == codes.FailedPrecondition && second == codes.OK)) {
		t.Fatalf("concurrent start and cancellation returned %v and %v", first, second)
	}
	var events int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type IN ?", newID, []string{"hr_offboarding_plans.start", "hr_offboarding_plans.cancel"}).
		Count(&events).Error; err != nil || events != 1 {
		t.Fatalf("concurrent start and cancellation outbox count=%d, err=%v", events, err)
	}
}
