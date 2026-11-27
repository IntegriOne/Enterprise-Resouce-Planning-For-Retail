package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type performanceGoalReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PerformanceGoal, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PerformanceGoal, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type performanceGoalValidator struct {
	workers lockedWorkerReader
	cycles  performanceCycleReader
	goals   performanceGoalReader
	updates performanceReferenceCounter
	scores  performanceReferenceCounter
}

// ValidateMutation menjaga hubungan goal dengan pekerja dan siklus yang masih terbuka.
func (v *performanceGoalValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_performance_goals" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.cycles) != nil || operation.RequireDependency(v.goals) != nil ||
		operation.RequireDependency(v.updates) != nil || operation.RequireDependency(v.scores) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PerformanceGoal
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PerformanceGoal)
		if !ok || previous == nil || (previous.GoalStatus != "draft" && previous.GoalStatus != "submitted") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.GoalStatus != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		if err := v.lockParents(ctx, previous); err != nil {
			return err
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	goal, ok := after.(*entity.PerformanceGoal)
	if !ok || goal == nil || (previous != nil && previous.ID != goal.ID) {
		return hrerrors.ErrInvalidInput
	}
	submitting := previous != nil && previous.GoalStatus == "draft" && goal.GoalStatus == "submitted"
	approving := previous != nil && (previous.GoalStatus == "draft" || previous.GoalStatus == "submitted") && goal.GoalStatus == "approved"
	rejecting := previous != nil && previous.GoalStatus == "submitted" && goal.GoalStatus == "rejected"
	if submitting {
		unchanged := *goal
		unchanged.GoalStatus = previous.GoalStatus
		if goal.WeightPercent == nil || goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil || !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if approving {
		unchanged := *goal
		unchanged.GoalStatus = previous.GoalStatus
		unchanged.ApprovedByPrincipalID = previous.ApprovedByPrincipalID
		unchanged.ApprovedAt = previous.ApprovedAt
		if goal.WeightPercent == nil || goal.ApprovedByPrincipalID == nil || goal.ApprovedAt == nil || !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if rejecting {
		unchanged := *goal
		unchanged.GoalStatus = previous.GoalStatus
		if goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil || !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if (previous != nil && previous.GoalStatus != "draft") ||
			goal.GoalStatus != "draft" || goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil || goal.ActualValue != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.WorkerID != goal.WorkerID || previous.PerformanceCycleID != goal.PerformanceCycleID ||
			!reflect.DeepEqual(previous.ParentGoalID, goal.ParentGoalID)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if err := v.lockParents(ctx, goal); err != nil {
		return err
	}
	if approving && goal.ApprovedByPrincipalID != nil {
		worker, err := v.workers.FindByIDForUpdate(ctx, goal.WorkerID)
		if err != nil {
			return err
		}
		if worker != nil && worker.IAMPrincipalReferenceID != nil &&
			strings.EqualFold(*worker.IAMPrincipalReferenceID, *goal.ApprovedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
	}
	if goal.GoalCode != nil && !submitting && !approving && !rejecting {
		matches, err := v.goals.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "worker_id", Values: []any{goal.WorkerID}},
			{Column: "performance_cycle_id", Values: []any{goal.PerformanceCycleID}},
			{Column: "goal_code", Values: []any{*goal.GoalCode}},
		}, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range matches {
			if existing.ID != goal.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
	}
	if previous != nil && !submitting && !approving && !rejecting && !reflect.DeepEqual(previous, goal) {
		return v.rejectIfReferenced(ctx, goal.ID)
	}
	return nil
}

// lockParents memakai urutan worker, siklus, lalu goal induk pada transaksi pemanggil.
func (v *performanceGoalValidator) lockParents(ctx context.Context, goal *entity.PerformanceGoal) error {
	worker, err := v.workers.FindByIDForUpdate(ctx, goal.WorkerID)
	if err != nil {
		return err
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, goal.PerformanceCycleID)
	if err != nil {
		return err
	}
	if cycle == nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		return hrerrors.ErrReferenceConflict
	}
	if goal.DueDate != nil && (entity.CompareDates(*goal.DueDate, cycle.PeriodStart) < 0 ||
		entity.CompareDates(*goal.DueDate, cycle.PeriodEnd) > 0) {
		return hrerrors.ErrInvalidInput
	}
	if goal.ParentGoalID == nil {
		return nil
	}
	if *goal.ParentGoalID == goal.ID {
		return hrerrors.ErrInvalidInput
	}
	parent, err := v.goals.FindByIDForUpdate(ctx, *goal.ParentGoalID)
	if err != nil {
		return err
	}
	if parent == nil || parent.WorkerID != goal.WorkerID || parent.PerformanceCycleID != goal.PerformanceCycleID ||
		(parent.GoalStatus != "draft" && parent.GoalStatus != "submitted" && parent.GoalStatus != "approved") {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (v *performanceGoalValidator) rejectIfReferenced(ctx context.Context, id string) error {
	for _, reference := range []struct {
		counter performanceReferenceCounter
		column  string
	}{
		{counter: v.goals, column: "parent_goal_id"},
		{counter: v.updates, column: "goal_id"},
		{counter: v.scores, column: "goal_id"},
	} {
		count, err := reference.counter.CountWhere(ctx, []domainrepo.Filter{{Column: reference.column, Values: []any{id}}})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
