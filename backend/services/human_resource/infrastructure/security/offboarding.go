package security

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type offboardingWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type offboardingPlanReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.OffboardingPlan, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.OffboardingPlan, error)
}

type offboardingTaskReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.OffboardingTask, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type offboardingPlanValidator struct {
	workers offboardingWorkerReader
	plans   offboardingPlanReader
	tasks   offboardingTaskReader
}

type offboardingTaskValidator struct {
	workers offboardingWorkerReader
	plans   offboardingPlanReader
	tasks   offboardingTaskReader
}

// OffboardingValidators mengaktifkan pengelolaan draf tanpa menyetujui clearance lintas owner.
func OffboardingValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.OffboardingPlan) != nil || operation.RequireDependency(repos.OffboardingTask) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_offboarding_plans": &offboardingPlanValidator{workers: repos.Worker, plans: repos.OffboardingPlan, tasks: repos.OffboardingTask},
		"hr_offboarding_tasks": &offboardingTaskValidator{workers: repos.Worker, plans: repos.OffboardingPlan, tasks: repos.OffboardingTask},
	}, nil
}

func (v *offboardingPlanValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_offboarding_plans" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.tasks) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var old *entity.OffboardingPlan
	if before != nil {
		var ok bool
		old, ok = before.(*entity.OffboardingPlan)
		if !ok || old == nil || old.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if old == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	plan := old
	if after != nil {
		var ok bool
		plan, ok = after.(*entity.OffboardingPlan)
		if !ok || plan == nil || plan.Status != "draft" || plan.ApprovedByPrincipalID != nil ||
			plan.ApprovedAt != nil || plan.CompletedAt != nil || plan.ExitInterviewDocumentVersionReferenceID != nil ||
			plan.FinalPayrollRunID != nil || plan.FinanceClearanceReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if old != nil && (old.ID != plan.ID || old.WorkerID != plan.WorkerID ||
			old.CreatedByPrincipalID != plan.CreatedByPrincipalID || !old.CreatedAt.Equal(plan.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return offboardingReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusNoticePeriod ||
		worker.HireDate != nil && entity.CompareDates(plan.NoticeDate, *worker.HireDate) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	if old != nil && (after == nil || !reflect.DeepEqual(old, plan)) {
		count, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{{Column: "offboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	if after == nil {
		return nil
	}
	active, err := v.plans.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{plan.WorkerID}},
		{Column: "status", Values: []any{"draft", "approved", "in_progress"}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range active {
		if existing.ID != plan.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *offboardingTaskValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_offboarding_tasks" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.plans) != nil || operation.RequireDependency(v.tasks) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var old *entity.OffboardingTask
	if before != nil {
		var ok bool
		old, ok = before.(*entity.OffboardingTask)
		if !ok || old == nil || old.Status != "pending" || old.CompletedAt != nil ||
			old.OwnerConfirmationReferenceID != nil || old.WaiverReason != nil || old.WaivedByPrincipalID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if old == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	task := old
	if after != nil {
		var ok bool
		task, ok = after.(*entity.OffboardingTask)
		if !ok || task == nil || task.Status != "pending" || task.CompletedAt != nil ||
			task.OwnerConfirmationReferenceID != nil || task.WaiverReason != nil || task.WaivedByPrincipalID != nil ||
			task.AssignedPrincipalID != nil || task.OwnerManagement != "human_resource" {
			return hrerrors.ErrWorkflowRequired
		}
		if old != nil && (old.ID != task.ID || old.OffboardingPlanID != task.OffboardingPlanID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	plan, err := v.plans.FindByIDForUpdate(ctx, task.OffboardingPlanID)
	if err != nil {
		return offboardingReferenceError(err)
	}
	if plan == nil || plan.Status != "draft" {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return offboardingReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusNoticePeriod {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if task.DueAt != nil && (offboardingDateBefore(*task.DueAt, plan.NoticeDate) ||
		offboardingDateBefore(plan.TerminationEffectiveDate, *task.DueAt)) {
		return hrerrors.Invalid("due_at", "must fall within the offboarding plan dates")
	}
	duplicates, err := v.tasks.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "offboarding_plan_id", Values: []any{task.OffboardingPlanID}},
		{Column: "task_code", Values: []any{task.TaskCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range duplicates {
		if existing.ID != task.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func offboardingDateBefore(left, right time.Time) bool {
	ly, lm, ld := left.UTC().Date()
	ry, rm, rd := right.UTC().Date()
	return ly < ry || ly == ry && (lm < rm || lm == rm && ld < rd)
}

func offboardingReferenceError(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
