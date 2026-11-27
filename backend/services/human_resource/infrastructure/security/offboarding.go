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

type offboardingPayrollReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PayrollRun, error)
}

type offboardingPlanValidator struct {
	payroll offboardingPayrollReader
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
		"hr_offboarding_plans": &offboardingPlanValidator{workers: repos.Worker, plans: repos.OffboardingPlan, tasks: repos.OffboardingTask, payroll: repos.PayrollRun},
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
		if !ok || old == nil || (old.Status != "draft" && old.Status != "approved" && old.Status != "in_progress") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if old == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	plan := old
	approving := false
	starting := false
	completing := false
	cancelling := false
	if after != nil {
		var ok bool
		plan, ok = after.(*entity.OffboardingPlan)
		if !ok || plan == nil || plan.ExitInterviewDocumentVersionReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		approving = old != nil && old.Status == "draft" && plan.Status == "approved"
		starting = old != nil && old.Status == "approved" && plan.Status == "in_progress"
		completing = old != nil && old.Status == "in_progress" && plan.Status == "completed"
		cancelling = old != nil && old.Status == "approved" && plan.Status == "cancelled"
		if cancelling {
			expected := *old
			expected.Status = "cancelled"
			expected.CancelledAt, expected.CancelledByPrincipalID, expected.CancellationReason = plan.CancelledAt, plan.CancelledByPrincipalID, plan.CancellationReason
			if old.ApprovedAt == nil || old.ApprovedByPrincipalID == nil || plan.CancelledAt == nil ||
				plan.CancelledByPrincipalID == nil || plan.CancellationReason == nil || strings.TrimSpace(*plan.CancellationReason) == "" ||
				plan.CancelledAt.Before(*old.ApprovedAt) || !reflect.DeepEqual(&expected, plan) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if plan.CancelledAt != nil || plan.CancelledByPrincipalID != nil || plan.CancellationReason != nil {
			return hrerrors.ErrWorkflowRequired
		} else if completing {
			expected := *old
			expected.Status = "completed"
			expected.CompletedAt = plan.CompletedAt
			expected.FinalPayrollRunID = plan.FinalPayrollRunID
			expected.FinanceClearanceReferenceID = plan.FinanceClearanceReferenceID
			if old.CompletedAt != nil || old.FinalPayrollRunID != nil || old.FinanceClearanceReferenceID != nil ||
				plan.CompletedAt == nil || plan.FinalPayrollRunID == nil || plan.FinanceClearanceReferenceID == nil ||
				old.ApprovedAt == nil || plan.CompletedAt.Before(*old.ApprovedAt) ||
				entity.CompareDates(plan.CompletedAt.UTC(), plan.TerminationEffectiveDate) < 0 ||
				!reflect.DeepEqual(&expected, plan) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if plan.CompletedAt != nil || plan.FinalPayrollRunID != nil || plan.FinanceClearanceReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		} else if approving {
			expected := *old
			expected.Status = "approved"
			expected.ApprovedAt = plan.ApprovedAt
			expected.ApprovedByPrincipalID = plan.ApprovedByPrincipalID
			if old.ApprovedAt != nil || old.ApprovedByPrincipalID != nil ||
				plan.ApprovedAt == nil || plan.ApprovedByPrincipalID == nil ||
				strings.EqualFold(old.CreatedByPrincipalID, *plan.ApprovedByPrincipalID) ||
				plan.ApprovedAt.Before(old.CreatedAt) ||
				entity.CompareDates(plan.ApprovedAt.UTC(), plan.TerminationEffectiveDate) > 0 ||
				!reflect.DeepEqual(&expected, plan) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if starting {
			expected := *old
			expected.Status = "in_progress"
			if old.ApprovedAt == nil || old.ApprovedByPrincipalID == nil ||
				strings.EqualFold(old.CreatedByPrincipalID, *old.ApprovedByPrincipalID) ||
				old.ApprovedAt.Before(old.CreatedAt) ||
				!reflect.DeepEqual(&expected, plan) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if (old != nil && old.Status != "draft") || plan.Status != "draft" || plan.ApprovedAt != nil || plan.ApprovedByPrincipalID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if old != nil && (old.ID != plan.ID || old.WorkerID != plan.WorkerID ||
			old.CreatedByPrincipalID != plan.CreatedByPrincipalID || !old.CreatedAt.Equal(plan.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil && old.Status != "draft" {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return offboardingReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusNoticePeriod &&
		(!completing || worker.Status != entity.WorkerStatusTerminated) ||
		worker.HireDate != nil && entity.CompareDates(plan.NoticeDate, *worker.HireDate) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	if cancelling {
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, *plan.CancelledByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		return nil
	}
	if after != nil && !completing && worker.ServiceStartDate != nil &&
		entity.CompareDates(plan.LastWorkingDate, *worker.ServiceStartDate) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	if approving && worker.IAMPrincipalReferenceID != nil &&
		strings.EqualFold(*worker.IAMPrincipalReferenceID, *plan.ApprovedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	if completing {
		if worker.Status != entity.WorkerStatusTerminated || worker.TerminationDate == nil ||
			entity.CompareDates(*worker.TerminationDate, plan.TerminationEffectiveDate) != 0 ||
			(worker.TerminationReasonCode != nil && *worker.TerminationReasonCode != plan.ReasonCode) {
			return hrerrors.ErrReferenceConflict
		}
		if operation.RequireDependency(v.payroll) != nil {
			return hrerrors.ErrUnavailable
		}
		run, err := v.payroll.FindByIDForUpdate(ctx, *plan.FinalPayrollRunID)
		if err != nil {
			return offboardingReferenceError(err)
		}
		if run == nil || !strings.EqualFold(run.EmployingLegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) {
			return hrerrors.ErrReferenceConflict
		}
		mandatory := []domainrepo.Filter{
			{Column: "offboarding_plan_id", Values: []any{plan.ID}},
			{Column: "is_mandatory", Values: []any{true}},
		}
		total, err := v.tasks.CountWhere(ctx, mandatory)
		if err != nil {
			return err
		}
		tasks, err := operation.CollectComplete(ctx, total, func(limit, offset int) (operation.Page[entity.OffboardingTask], error) {
			items, err := v.tasks.SelectWhere(ctx, mandatory, limit, offset)
			return operation.Page[entity.OffboardingTask]{Items: items}, err
		})
		if err != nil {
			return err
		}
		for _, task := range tasks {
			switch task.Status {
			case "completed":
				if task.CompletedAt == nil || task.OwnerConfirmationReferenceID == nil ||
					task.CompletedAt.Before(*plan.ApprovedAt) || task.CompletedAt.After(*plan.CompletedAt) {
					return hrerrors.ErrReferenceConflict
				}
			case "waived":
				if task.WaivedByPrincipalID == nil || task.WaiverReason == nil || strings.TrimSpace(*task.WaiverReason) == "" {
					return hrerrors.ErrReferenceConflict
				}
			default:
				return hrerrors.ErrReferenceConflict
			}
		}
		return nil
	}
	requiresTasks := approving || starting
	if requiresTasks || (old != nil && (after == nil || !reflect.DeepEqual(old, plan))) {
		count, err := v.tasks.CountWhere(ctx, []domainrepo.Filter{{Column: "offboarding_plan_id", Values: []any{plan.ID}}})
		if err != nil {
			return err
		}
		if (requiresTasks && count == 0) || (!requiresTasks && count != 0) {
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
	waiving := false
	completing := false
	if after != nil {
		var ok bool
		task, ok = after.(*entity.OffboardingTask)
		if !ok || task == nil ||
			task.AssignedPrincipalID != nil || task.OwnerManagement != "human_resource" {
			return hrerrors.ErrWorkflowRequired
		}
		completing = old != nil && task.Status == "completed"
		waiving = old != nil && task.Status == "waived"
		if completing {
			expected := *old
			expected.Status = "completed"
			expected.CompletedAt = task.CompletedAt
			expected.OwnerConfirmationReferenceID = task.OwnerConfirmationReferenceID
			if task.CompletedAt == nil || task.OwnerConfirmationReferenceID == nil || !reflect.DeepEqual(&expected, task) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if task.CompletedAt != nil || task.OwnerConfirmationReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		} else if waiving {
			expected := *old
			expected.Status = "waived"
			expected.WaiverReason = task.WaiverReason
			expected.WaivedByPrincipalID = task.WaivedByPrincipalID
			if task.WaiverReason == nil || strings.TrimSpace(*task.WaiverReason) == "" ||
				task.WaivedByPrincipalID == nil || !reflect.DeepEqual(&expected, task) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if task.Status != "pending" || task.WaiverReason != nil || task.WaivedByPrincipalID != nil {
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
	requiredStatus := "draft"
	if waiving || completing {
		requiredStatus = "in_progress"
	}
	if plan == nil || plan.Status != requiredStatus {
		return hrerrors.ErrReferenceConflict
	}
	if completing && (plan.ApprovedAt == nil || task.CompletedAt.Before(*plan.ApprovedAt)) {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, plan.WorkerID)
	if err != nil {
		return offboardingReferenceError(err)
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	if worker.Status != entity.WorkerStatusActive && worker.Status != entity.WorkerStatusNoticePeriod &&
		(!(waiving || completing) || worker.Status != entity.WorkerStatusTerminated ||
			worker.TerminationDate == nil || entity.CompareDates(*worker.TerminationDate, plan.TerminationEffectiveDate) != 0) {
		return hrerrors.ErrReferenceConflict
	}
	if waiving && worker.IAMPrincipalReferenceID != nil &&
		strings.EqualFold(*worker.IAMPrincipalReferenceID, *task.WaivedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	if after == nil || waiving || completing {
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
