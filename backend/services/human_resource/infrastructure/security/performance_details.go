package security

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type performanceReviewLockReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PerformanceReview, error)
}

type performanceGoalUpdateReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PerformanceGoalUpdate, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PerformanceGoalUpdate, error)
}

type performanceReviewGoalScoreReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PerformanceReviewGoalScore, error)
}

type performanceGoalUpdateValidator struct {
	goals   performanceGoalReader
	cycles  performanceCycleReader
	workers lockedWorkerReader
	updates performanceGoalUpdateReader
}

type performanceReviewGoalScoreValidator struct {
	reviews performanceReviewLockReader
	goals   performanceGoalReader
	cycles  performanceCycleReader
	scores  performanceReviewGoalScoreReader
}

// ValidateMutation menjaga riwayat pengukuran KPI tetap append-only dan terikat pada goal yang telah disetujui.
func (v *performanceGoalUpdateValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_performance_goal_updates" ||
		operation.RequireDependency(v.goals) != nil ||
		operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.updates) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	// Riwayat KPI bersifat append-only sesuai trigger hr_reject_immutable_change.
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	update, ok := after.(*entity.PerformanceGoalUpdate)
	if !ok || update == nil {
		return hrerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	if !strings.EqualFold(caller.PrincipalID, update.RecordedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	if strings.TrimSpace(update.IdempotencyKey) == "" || len(update.EvidenceDigest) == 0 {
		return hrerrors.ErrInvalidInput
	}
	if update.SupersedesUpdateID != nil && *update.SupersedesUpdateID == update.ID {
		return hrerrors.ErrInvalidInput
	}
	goal, err := v.goals.FindByIDForUpdate(ctx, update.GoalID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if goal == nil || goal.GoalStatus != "approved" {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, goal.WorkerID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, goal.PerformanceCycleID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if cycle == nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		return hrerrors.ErrReferenceConflict
	}
	if update.SupersedesUpdateID != nil {
		superseded, err := v.updates.FindByIDForUpdate(ctx, *update.SupersedesUpdateID)
		if err != nil {
			return reviewReferenceError(err)
		}
		if superseded == nil || superseded.GoalID != update.GoalID {
			return hrerrors.ErrReferenceConflict
		}
		dupSupersedes, err := v.updates.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "supersedes_update_id", Values: []any{*update.SupersedesUpdateID}},
		}, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range dupSupersedes {
			if existing.ID != update.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
	}
	dupIdem, err := v.updates.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{update.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range dupIdem {
		if existing.ID != update.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// ValidateMutation menjaga skor goal konsisten dengan bobot goal disetujui serta membekukan skor setelah review difinalisasi.
func (v *performanceReviewGoalScoreValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_performance_review_goal_scores" ||
		operation.RequireDependency(v.reviews) != nil ||
		operation.RequireDependency(v.goals) != nil ||
		operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.scores) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PerformanceReviewGoalScore
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PerformanceReviewGoalScore)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.lockMutableReviewAndCycle(ctx, previous.PerformanceReviewID, previous.PerformanceCycleID)
	}
	score, ok := after.(*entity.PerformanceReviewGoalScore)
	if !ok || score == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && (previous.ID != score.ID ||
		previous.PerformanceReviewID != score.PerformanceReviewID ||
		previous.GoalID != score.GoalID ||
		previous.WorkerID != score.WorkerID ||
		previous.PerformanceCycleID != score.PerformanceCycleID) {
		return hrerrors.ErrWorkflowRequired
	}
	review, err := v.reviews.FindByIDForUpdate(ctx, score.PerformanceReviewID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if review == nil || review.FinalizedAt != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if review.WorkerID != score.WorkerID || review.PerformanceCycleID != score.PerformanceCycleID {
		return hrerrors.ErrReferenceConflict
	}
	goal, err := v.goals.FindByIDForUpdate(ctx, score.GoalID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if goal == nil || goal.WorkerID != score.WorkerID || goal.PerformanceCycleID != score.PerformanceCycleID ||
		goal.GoalStatus != "approved" || goal.WeightPercent == nil {
		return hrerrors.ErrReferenceConflict
	}
	expectedWeight, err := goal.WeightPercent.Rat()
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	actualWeight, err := score.WeightPercentSnapshot.Rat()
	if err != nil {
		return hrerrors.ErrInvalidInput
	}
	if expectedWeight.Cmp(actualWeight) != 0 {
		return hrerrors.ErrInvalidInput
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, score.PerformanceCycleID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if cycle == nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		return hrerrors.ErrReferenceConflict
	}
	matches, err := v.scores.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "performance_review_id", Values: []any{score.PerformanceReviewID}},
		{Column: "goal_id", Values: []any{score.GoalID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != score.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *performanceReviewGoalScoreValidator) lockMutableReviewAndCycle(ctx context.Context, reviewID, cycleID string) error {
	review, err := v.reviews.FindByIDForUpdate(ctx, reviewID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if review == nil || review.FinalizedAt != nil {
		return hrerrors.ErrWorkflowRequired
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, cycleID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if cycle == nil || cycle.Status != entity.PerformanceCycleStatusOpen {
		return hrerrors.ErrWorkflowRequired
	}
	return nil
}
