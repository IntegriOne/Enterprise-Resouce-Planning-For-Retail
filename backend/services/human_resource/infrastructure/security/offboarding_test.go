package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type offboardingWorkerStub struct{ value *entity.Worker }

func (s *offboardingWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	if s.value == nil {
		return nil, hrerrors.ErrNotFound
	}
	return s.value, nil
}

type offboardingPlanStub struct {
	value *entity.OffboardingPlan
	rows  []entity.OffboardingPlan
}

func (s *offboardingPlanStub) FindByIDForUpdate(context.Context, string) (*entity.OffboardingPlan, error) {
	if s.value == nil {
		return nil, hrerrors.ErrNotFound
	}
	return s.value, nil
}
func (s *offboardingPlanStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OffboardingPlan, error) {
	if len(filters) != 2 || filters[0].Column != "worker_id" || filters[1].Column != "status" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type offboardingTaskStub struct {
	rows  []entity.OffboardingTask
	count int64
}

func (s *offboardingTaskStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OffboardingTask, error) {
	if len(filters) == 2 && filters[0].Column == "offboarding_plan_id" && filters[1].Column == "is_mandatory" &&
		limit == domainrepo.MaxLimit && offset == 0 {
		return s.rows, nil
	}
	if len(filters) != 2 || filters[0].Column != "offboarding_plan_id" || filters[1].Column != "task_code" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}
func (s *offboardingTaskStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) < 1 || len(filters) > 2 || filters[0].Column != "offboarding_plan_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestOffboardingDraftPlanValidator(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	worker := &offboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusActive}}
	plans := &offboardingPlanStub{}
	tasks := &offboardingTaskStub{}
	v := &offboardingPlanValidator{workers: worker, plans: plans, tasks: tasks}
	plan := &entity.OffboardingPlan{ID: "plan-1", WorkerID: "worker-1", PlanNumber: "EXIT-1", ReasonCode: "RESIGNATION",
		NoticeDate: start, LastWorkingDate: start.AddDate(0, 0, 7), TerminationEffectiveDate: start.AddDate(0, 0, 8), Status: "draft"}
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, plan); err != nil {
		t.Fatalf("valid draft: %v", err)
	}
	serviceStart := plan.LastWorkingDate.AddDate(0, 0, 1)
	worker.value.ServiceStartDate = &serviceStart
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("last working date before service begins: %v", err)
	}
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", plan, nil); err != nil {
		t.Fatalf("delete outdated draft: %v", err)
	}
	worker.value.ServiceStartDate = &plan.LastWorkingDate
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, plan); err != nil {
		t.Fatalf("last working date on service start: %v", err)
	}
	worker.value.ServiceStartDate = nil
	plans.rows = []entity.OffboardingPlan{{ID: "plan-2"}}
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, plan); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("second open plan: %v", err)
	}
	plans.rows = nil
	changed := *plan
	changed.FinalPayrollRunID = &plan.WorkerID
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified payroll: %v", err)
	}
	changed = *plan
	changed.WorkerID = "worker-2"
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("worker changed: %v", err)
	}
	worker.value.Status = entity.WorkerStatusTerminated
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("terminated worker: %v", err)
	}
	worker.value.Status = entity.WorkerStatusNoticePeriod
	tasks.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced plan: %v", err)
	}
}

func TestOffboardingApprovalRequiresTasks(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	approver := "approver"
	draft := &entity.OffboardingPlan{
		ID: "plan-1", WorkerID: "worker-1", CreatedByPrincipalID: "creator",
		Status: "draft", NoticeDate: now, TerminationEffectiveDate: now,
	}
	approved := *draft
	approved.Status = "approved"
	approved.ApprovedAt = &now
	approved.ApprovedByPrincipalID = &approver
	tasks := &offboardingTaskStub{}
	validator := &offboardingPlanValidator{
		workers: &offboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusActive}},
		plans:   &offboardingPlanStub{},
		tasks:   tasks,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", draft, &approved); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("approve without tasks: %v", err)
	}
	tasks.count = 1
	draft.CreatedAt = now
	approved.CreatedAt = now
	beforeCreation := now.Add(-time.Second)
	approved.ApprovedAt = &beforeCreation
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", draft, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("approve before creation: %v", err)
	}
	started := approved
	started.Status = "in_progress"
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", &approved, &started); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("start plan with approval before creation: %v", err)
	}
	approved.ApprovedAt = &now
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", draft, &approved); err != nil {
		t.Fatalf("approve with tasks: %v", err)
	}
}

func TestOffboardingDraftTaskValidator(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	plan := &entity.OffboardingPlan{ID: "plan-1", WorkerID: "worker-1", Status: "draft",
		NoticeDate: start, LastWorkingDate: start.AddDate(0, 0, 7), TerminationEffectiveDate: start.AddDate(0, 0, 8)}
	worker := &offboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusActive}}
	plans := &offboardingPlanStub{value: plan}
	tasks := &offboardingTaskStub{}
	v := &offboardingTaskValidator{workers: worker, plans: plans, tasks: tasks}
	task := &entity.OffboardingTask{ID: "task-1", OffboardingPlanID: plan.ID, TaskCode: "RETURN-BADGE",
		TaskName: "Return badge", OwnerManagement: "human_resource", Status: "pending"}
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, task); err != nil {
		t.Fatalf("valid draft task: %v", err)
	}
	other := *task
	other.OwnerManagement = "finance_accounting"
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, &other); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified external owner: %v", err)
	}
	other = *task
	other.AssignedPrincipalID = &plan.WorkerID
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, &other); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified assignee: %v", err)
	}
	other = *task
	other.Status = "completed"
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, &other); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("completion through CRUD: %v", err)
	}
	other = *task
	late := plan.TerminationEffectiveDate.AddDate(0, 0, 1)
	other.DueAt = &late
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, &other); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("late due date: %v", err)
	}
	tasks.rows = []entity.OffboardingTask{{ID: "task-2"}}
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, task); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate task code: %v", err)
	}
	plans.value.Status = "approved"
	if err := v.ValidateMutation(t.Context(), "hr_offboarding_tasks", nil, task); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("task on approved plan: %v", err)
	}
}

func TestOffboardingTaskResolutionAfterTermination(t *testing.T) {
	t.Parallel()
	termination := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	approved := termination.AddDate(0, 0, -1)
	worker := &entity.Worker{Status: entity.WorkerStatusTerminated, TerminationDate: &termination}
	plan := &entity.OffboardingPlan{ID: "plan", Status: "in_progress", ApprovedAt: &approved, TerminationEffectiveDate: termination}
	validator := &offboardingTaskValidator{
		workers: &offboardingWorkerStub{value: worker},
		plans:   &offboardingPlanStub{value: plan}, tasks: &offboardingTaskStub{},
	}
	before := &entity.OffboardingTask{ID: "task", OffboardingPlanID: plan.ID, OwnerManagement: "human_resource", Status: "pending"}
	after := *before
	after.Status = "completed"
	after.CompletedAt = &termination
	reference := "evidence"
	after.OwnerConfirmationReferenceID = &reference
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_tasks", before, &after); err != nil {
		t.Fatalf("complete after matching termination: %v", err)
	}
	worker.TerminationDate = &approved
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_tasks", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("complete after mismatched termination = %v, want ErrReferenceConflict", err)
	}
}

type offboardingPayrollStub struct{ run *entity.PayrollRun }

func (s *offboardingPayrollStub) FindByIDForUpdate(context.Context, string) (*entity.PayrollRun, error) {
	return s.run, nil
}

func TestOffboardingCompletionRequiresResolvedMandatoryTasks(t *testing.T) {
	t.Parallel()
	effective := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	approved := effective.AddDate(0, 0, -1)
	reference := "reference"
	before := &entity.OffboardingPlan{
		ID: "plan", WorkerID: "worker", Status: "in_progress", ApprovedAt: &approved, ReasonCode: "RESIGNATION",
		TerminationEffectiveDate: effective,
	}
	after := *before
	after.Status = "completed"
	after.CompletedAt = &effective
	after.FinalPayrollRunID = &reference
	after.FinanceClearanceReferenceID = &reference
	reason := "Documented exception"
	tasks := &offboardingTaskStub{count: 2, rows: []entity.OffboardingTask{
		{Status: "pending"}, {Status: "waived", WaiverReason: &reason, WaivedByPrincipalID: &reference},
	}}
	payroll := &offboardingPayrollStub{run: &entity.PayrollRun{EmployingLegalEntityReferenceID: "employer"}}
	validator := &offboardingPlanValidator{
		payroll: payroll,
		workers: &offboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusTerminated, TerminationDate: &effective, PrimaryLegalEntityReferenceID: "employer"}},
		plans:   &offboardingPlanStub{}, tasks: tasks,
	}
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("unresolved mandatory task: %v", err)
	}
	early := approved.Add(-time.Second)
	tasks.rows[0] = entity.OffboardingTask{Status: "completed", CompletedAt: &early, OwnerConfirmationReferenceID: &reference}
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("task completed before plan approval: %v", err)
	}
	tasks.rows[0].CompletedAt = &effective
	payroll.run.EmployingLegalEntityReferenceID = "other employer"
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("final payroll belongs to another employer: %v", err)
	}
	payroll.run.EmployingLegalEntityReferenceID = "employer"
	wrongReason := "OTHER"
	validator.workers.(*offboardingWorkerStub).value.TerminationReasonCode = &wrongReason
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("worker termination reason differs from approved plan: %v", err)
	}
	validator.workers.(*offboardingWorkerStub).value.TerminationReasonCode = &before.ReasonCode
	if err := validator.ValidateMutation(t.Context(), "hr_offboarding_plans", before, &after); err != nil {
		t.Fatalf("complete with all mandatory tasks resolved: %v", err)
	}
}
