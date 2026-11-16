package security

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type onboardingWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type onboardingPlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.OnboardingPlan, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.OnboardingPlan, error)
}

type onboardingTaskReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.OnboardingTask, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type onboardingPlanValidator struct {
	workers onboardingWorkerReader
	plans   onboardingPlanReader
	tasks   onboardingTaskReader
}

type onboardingTaskValidator struct {
	workers onboardingWorkerReader
	plans   onboardingPlanReader
	tasks   onboardingTaskReader
}

// OnboardingValidators mengaktifkan penyusunan draf rencana dan tugas onboarding.
func OnboardingValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.OnboardingPlan) != nil ||
		operation.RequireDependency(repos.OnboardingTask) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_onboarding_plans": &onboardingPlanValidator{
			workers: repos.Worker, plans: repos.OnboardingPlan, tasks: repos.OnboardingTask,
		},
		"hr_onboarding_tasks": &onboardingTaskValidator{
			workers: repos.Worker, plans: repos.OnboardingPlan, tasks: repos.OnboardingTask,
		},
	}, nil
}

func (v *onboardingPlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_onboarding_plans" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.tasks) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.OnboardingPlan
	starting := false
	completing := false
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.OnboardingPlan)
		if !ok || previous == nil || (previous.Status != "draft" && previous.Status != "in_progress") || previous.TemplateReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	plan := previous
	if after != nil {
		var ok bool
		plan, ok = after.(*entity.OnboardingPlan)
		if !ok || plan == nil {
			return hrerrors.ErrInvalidInput
		}
		starting = previous != nil && validOnboardingStart(previous, plan)
		completing = previous != nil && validOnboardingComplete(previous, plan)
		if plan.Status != "draft" && !starting && !completing || plan.TemplateReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != plan.ID || previous.WorkerID != plan.WorkerID ||
			!reflect.DeepEqual(previous.CreatedByPrincipalID, plan.CreatedByPrincipalID)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous != nil && previous.Status == "in_progress" && !completing {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return onboardingReferenceError(err)
	}
	if worker == nil || (worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive) {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil && !starting && !completing && (after == nil || !reflect.DeepEqual(previous, plan)) {
		count, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	if after == nil {
		return nil
	}
	if starting {
		count, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if count == 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	if completing {
		count, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{{Column: "onboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if count == 0 {
			return hrerrors.ErrReferenceConflict
		}
		pending, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{
			{Column: "onboarding_plan_id", Values: []any{plan.ID}},
			{Column: "is_mandatory", Values: []any{true}},
			{Column: "completed_at", Values: []any{nil}},
		})
		if err != nil {
			return err
		}
		if pending > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	matches, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{plan.WorkerID}},
		{Column: "plan_code", Values: []any{plan.PlanCode}},
		{Column: "plan_version", Values: []any{plan.PlanVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func validOnboardingStart(before, after *entity.OnboardingPlan) bool {
	if before == nil || after == nil || before.Status != "draft" || after.Status != "in_progress" ||
		before.StartedAt != nil || before.StartedByPrincipalID != nil || after.StartedAt == nil ||
		after.StartedByPrincipalID == nil || after.CompletedAt != nil || after.CompletedByPrincipalID != nil {
		return false
	}
	unchanged := *after
	unchanged.Status = before.Status
	unchanged.StartedAt = before.StartedAt
	unchanged.StartedByPrincipalID = before.StartedByPrincipalID
	return reflect.DeepEqual(*before, unchanged)
}

func validOnboardingComplete(before, after *entity.OnboardingPlan) bool {
	if before == nil || after == nil || before.Status != "in_progress" || after.Status != "completed" ||
		before.StartedAt == nil || before.StartedByPrincipalID == nil || before.CompletedAt != nil ||
		before.CompletedByPrincipalID != nil || after.CompletedAt == nil || after.CompletedByPrincipalID == nil ||
		after.CompletedAt.Before(*before.StartedAt) {
		return false
	}
	unchanged := *after
	unchanged.Status = before.Status
	unchanged.CompletedAt = before.CompletedAt
	unchanged.CompletedByPrincipalID = before.CompletedByPrincipalID
	return reflect.DeepEqual(*before, unchanged)
}

func (v *onboardingTaskValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_onboarding_tasks" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.plans) != nil ||
		operation.RequireDependency(v.tasks) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.OnboardingTask
	completing := false
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.OnboardingTask)
		if !ok || previous == nil || previous.CompletedAt != nil || previous.CompletionEvidenceReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	task := previous
	if after != nil {
		var ok bool
		task, ok = after.(*entity.OnboardingTask)
		if !ok || task == nil {
			return hrerrors.ErrInvalidInput
		}
		completing = previous != nil && validOnboardingTaskComplete(previous, task)
		if !completing && (task.CompletedAt != nil || task.CompletedByPrincipalID != nil || task.CompletionEvidenceReferenceID != nil) ||
			!validOnboardingTaskOwner(task.TaskOwnerType) {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != task.ID || previous.OnboardingPlanID != task.OnboardingPlanID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, task.OnboardingPlanID)
	if err != nil {
		return onboardingReferenceError(err)
	}
	if plan == nil || plan.TemplateReferenceID != nil ||
		(completing && plan.Status != "in_progress") || (!completing && plan.Status != "draft") {
		return hrerrors.ErrReferenceConflict
	}
	if completing && (plan.StartedAt == nil || task.CompletedAt.Before(*plan.StartedAt)) {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return onboardingReferenceError(err)
	}
	if worker == nil || (worker.Status != entity.WorkerStatusPreboarding && worker.Status != entity.WorkerStatusActive) {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if task.AssignedPrincipalID != nil && (task.TaskOwnerType != "worker" ||
		worker.IAMPrincipalReferenceID == nil || !strings.EqualFold(*task.AssignedPrincipalID, *worker.IAMPrincipalReferenceID)) {
		return hrerrors.ErrReferenceConflict
	}
	if completing && (task.TaskOwnerType != "worker" || task.AssignedPrincipalID == nil ||
		!strings.EqualFold(*task.AssignedPrincipalID, *task.CompletedByPrincipalID)) {
		return hrerrors.ErrReferenceConflict
	}
	if completing {
		return nil
	}
	if task.DueAt != nil && (dateBefore(*task.DueAt, plan.StartDate) ||
		plan.TargetCompletionDate != nil && dateBefore(*plan.TargetCompletionDate, *task.DueAt)) {
		return hrerrors.Invalid("due_at", "must fall within the onboarding plan dates")
	}
	matches, err := v.tasks.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "onboarding_plan_id", Values: []any{task.OnboardingPlanID}},
		{Column: "task_code", Values: []any{task.TaskCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != task.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func validOnboardingTaskComplete(before, after *entity.OnboardingTask) bool {
	if before == nil || after == nil || before.CompletedAt != nil || before.CompletedByPrincipalID != nil ||
		before.CompletionEvidenceReferenceID != nil || after.CompletedAt == nil ||
		after.CompletedByPrincipalID == nil || after.CompletionEvidenceReferenceID == nil {
		return false
	}
	unchanged := *after
	unchanged.CompletedAt = before.CompletedAt
	unchanged.CompletedByPrincipalID = before.CompletedByPrincipalID
	unchanged.CompletionEvidenceReferenceID = before.CompletionEvidenceReferenceID
	return reflect.DeepEqual(*before, unchanged)
}

func validOnboardingTaskOwner(value string) bool {
	switch value {
	case "worker", "manager", "hr", "it", "security", "facilities":
		return true
	default:
		return false
	}
}

func dateBefore(left, right time.Time) bool {
	ly, lm, ld := left.UTC().Date()
	ry, rm, rd := right.UTC().Date()
	return ly < ry || ly == ry && (lm < rm || lm == rm && ld < rd)
}

func onboardingReferenceError(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
