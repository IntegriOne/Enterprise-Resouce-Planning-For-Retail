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

type onboardingWorkerStub struct {
	value *entity.Worker
	err   error
}

func (s *onboardingWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.value, s.err
}

type onboardingPlanStub struct {
	value *entity.OnboardingPlan
	rows  []entity.OnboardingPlan
	err   error
}

func (s *onboardingPlanStub) FindByIDForUpdate(context.Context, string) (*entity.OnboardingPlan, error) {
	return s.value, s.err
}

func (s *onboardingPlanStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OnboardingPlan, error) {
	if len(filters) != 3 || filters[0].Column != "worker_id" || filters[1].Column != "plan_code" ||
		filters[2].Column != "plan_version" || limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, s.err
}

type onboardingTaskStub struct {
	rows    []entity.OnboardingTask
	count   int64
	pending int64
	err     error
}

func TestOnboardingTaskValidatorAllowsOnlyWorkerCompletionOnStartedPlan(t *testing.T) {
	t.Parallel()
	principal := "00000000-0000-4000-8000-000000000001"
	started := time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC)
	completed := started.Add(time.Hour)
	evidence := "00000000-0000-4000-8000-000000000002"
	worker := &onboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &principal}}
	plan := &onboardingPlanStub{value: &entity.OnboardingPlan{Status: "in_progress", StartedAt: &started, StartDate: started}}
	tasks := &onboardingTaskStub{}
	validator := &onboardingTaskValidator{workers: worker, plans: plan, tasks: tasks}
	before := &entity.OnboardingTask{ID: "task", OnboardingPlanID: "plan", TaskCode: "ORIENTATION", TaskName: "Orientation", TaskOwnerType: "worker", AssignedPrincipalID: &principal, IsMandatory: true}
	after := *before
	after.CompletedAt = &completed
	after.CompletedByPrincipalID = &principal
	after.CompletionEvidenceReferenceID = &evidence
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", before, &after); err != nil {
		t.Fatalf("valid worker completion: %v", err)
	}
	forged := after
	forged.TaskName = "Changed during completion"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", before, &forged); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("completion changed task fields: %v", err)
	}
	plan.value.Status = "draft"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", before, &after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("completion on draft plan: %v", err)
	}
}

func (s *onboardingTaskStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OnboardingTask, error) {
	if len(filters) != 2 || filters[0].Column != "onboarding_plan_id" || filters[1].Column != "task_code" ||
		limit != 2 || offset != 0 {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, s.err
}

func (s *onboardingTaskStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) == 3 && filters[0].Column == "onboarding_plan_id" &&
		filters[1].Column == "is_mandatory" && filters[2].Column == "completed_at" {
		return s.pending, s.err
	}
	if len(filters) != 1 || filters[0].Column != "onboarding_plan_id" {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, s.err
}

func TestOnboardingPlanValidatorKeepsDraftAndWorkerReferences(t *testing.T) {
	t.Parallel()
	start := time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)
	worker := &onboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusPreboarding}}
	plans := &onboardingPlanStub{}
	tasks := &onboardingTaskStub{}
	validator := &onboardingPlanValidator{workers: worker, plans: plans, tasks: tasks}
	principal := "00000000-0000-4000-8000-000000000001"
	plan := &entity.OnboardingPlan{ID: "plan-1", WorkerID: "worker-1", PlanCode: "FIRST-WEEK", PlanVersion: 1,
		StartDate: start, Status: "draft", CreatedByPrincipalID: &principal}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", nil, plan); err != nil {
		t.Fatalf("create draft for preboarding worker: %v", err)
	}
	plans.rows = []entity.OnboardingPlan{{ID: "plan-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", nil, plan); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate worker/code/version: %v", err)
	}
	plans.rows = nil
	worker.value.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", nil, plan); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("terminated worker: %v", err)
	}
	worker.value.Status = entity.WorkerStatusActive
	changed := *plan
	changed.PlanCode = "SECOND-WEEK"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); err != nil {
		t.Fatalf("update unused draft: %v", err)
	}
	tasks.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("change plan with tasks: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("cascade delete with tasks: %v", err)
	}
	tasks.count = 0
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, nil); err != nil {
		t.Fatalf("delete unused draft: %v", err)
	}
	changed.CreatedByPrincipalID = nil
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("remove server creator: %v", err)
	}
	changed = *plan
	changed.Status = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("start through CRUD: %v", err)
	}
	changed = *plan
	changed.Status = "in_progress"
	changed.StartedAt = &start
	changed.StartedByPrincipalID = &principal
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("start without tasks: %v", err)
	}
	tasks.count = 1
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); err != nil {
		t.Fatalf("start with tasks: %v", err)
	}
	changed.PlanCode = "FORGED"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", plan, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("start changing plan data: %v", err)
	}
	changed = *plan
	changed.TemplateReferenceID = &principal
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified template reference: %v", err)
	}
}

func TestOnboardingPlanValidatorRequiresCompletedMandatoryTasks(t *testing.T) {
	t.Parallel()
	now := time.Date(2027, 1, 11, 12, 0, 0, 0, time.UTC)
	principal := "00000000-0000-4000-8000-000000000001"
	worker := &onboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusPreboarding}}
	tasks := &onboardingTaskStub{count: 2, pending: 1}
	validator := &onboardingPlanValidator{workers: worker, plans: &onboardingPlanStub{}, tasks: tasks}
	started := &entity.OnboardingPlan{ID: "plan-1", WorkerID: "worker-1", PlanCode: "FIRST-WEEK", PlanVersion: 1,
		StartDate: now.Add(-24 * time.Hour), Status: "in_progress", StartedAt: &now, StartedByPrincipalID: &principal}
	completed := *started
	completed.Status = "completed"
	finishedAt := now.Add(time.Hour)
	completed.CompletedAt = &finishedAt
	completed.CompletedByPrincipalID = &principal
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", started, &completed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("pending mandatory task: %v", err)
	}
	tasks.pending = 0
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", started, &completed); err != nil {
		t.Fatalf("all mandatory tasks completed: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", started, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete started plan: %v", err)
	}
	forged := completed
	forged.PlanCode = "FORGED"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_plans", started, &forged); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("completion changed plan fields: %v", err)
	}
}

func TestOnboardingTaskValidatorChecksParentWorkerAndCompletion(t *testing.T) {
	t.Parallel()
	start := time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	principal := "00000000-0000-4000-8000-000000000001"
	worker := &onboardingWorkerStub{value: &entity.Worker{Status: entity.WorkerStatusPreboarding,
		IAMPrincipalReferenceID: &principal}}
	plans := &onboardingPlanStub{value: &entity.OnboardingPlan{ID: "plan-1", WorkerID: "worker-1", Status: "draft",
		StartDate: start, TargetCompletionDate: &end}}
	tasks := &onboardingTaskStub{}
	validator := &onboardingTaskValidator{workers: worker, plans: plans, tasks: tasks}
	due := end.Add(12 * time.Hour)
	task := &entity.OnboardingTask{ID: "task-1", OnboardingPlanID: "plan-1", TaskCode: "ORIENTATION",
		TaskName: "Orientation", TaskOwnerType: "worker", AssignedPrincipalID: &principal, DueAt: &due, IsMandatory: true}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, task); err != nil {
		t.Fatalf("create worker task on final plan day: %v", err)
	}
	tasks.rows = []entity.OnboardingTask{{ID: "task-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, task); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate task code: %v", err)
	}
	tasks.rows = nil
	wrongPrincipal := "00000000-0000-4000-8000-000000000009"
	changed := *task
	changed.AssignedPrincipalID = &wrongPrincipal
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("unrelated principal: %v", err)
	}
	changed = *task
	changed.TaskOwnerType = "manager"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("unverified manager assignment: %v", err)
	}
	changed.AssignedPrincipalID = nil
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); err != nil {
		t.Fatalf("unassigned manager planning task: %v", err)
	}
	changed.TaskOwnerType = "other"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unknown task owner: %v", err)
	}
	changed = *task
	late := end.AddDate(0, 0, 1)
	changed.DueAt = &late
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("due after plan completion date: %v", err)
	}
	changed = *task
	changed.CompletedAt = &due
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, &changed); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("complete through CRUD: %v", err)
	}
	changed = *task
	changed.TaskName = "Updated orientation"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", task, &changed); err != nil {
		t.Fatalf("update pending task: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", task, nil); err != nil {
		t.Fatalf("delete pending task: %v", err)
	}
	plans.value.Status = "active"
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", task, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete task in active plan: %v", err)
	}
	plans.value.Status = "draft"
	worker.value.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_onboarding_tasks", nil, task); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("task for terminated worker: %v", err)
	}
}
