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
	if len(filters) != 2 || filters[0].Column != "offboarding_plan_id" || filters[1].Column != "task_code" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}
func (s *offboardingTaskStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "offboarding_plan_id" {
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
