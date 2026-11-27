package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	performancepolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/performance"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// SubmitPerformanceGoal mengajukan goal draft agar siap ditinjau dan disetujui.
func (s *Service) SubmitPerformanceGoal(ctx context.Context, goalID string) (*entity.PerformanceGoal, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Command(ctx, goalID, "submit", func(_ context.Context, goal *entity.PerformanceGoal, actor operation.Actor, _ time.Time) (*entity.PerformanceGoal, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if goal.GoalStatus != "draft" || goal.WeightPercent == nil || goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		goal.GoalStatus = "submitted"
		return goal, nil
	})
}

// ApprovePerformanceGoal menyetujui goal sambil menegakkan pemisahan wewenang terhadap pemilik goal.
func (s *Service) ApprovePerformanceGoal(ctx context.Context, goalID string) (*entity.PerformanceGoal, error) {
	if s == nil || operation.RequireDependency(s.dependencies.Worker) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Command(ctx, goalID, "approve", func(txCtx context.Context, goal *entity.PerformanceGoal, actor operation.Actor, now time.Time) (*entity.PerformanceGoal, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if (goal.GoalStatus != "draft" && goal.GoalStatus != "submitted") ||
			goal.WeightPercent == nil || goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, goal.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		goal.GoalStatus = "approved"
		goal.ApprovedByPrincipalID = &actor.PrincipalID
		goal.ApprovedAt = &now
		return goal, nil
	})
}

// RejectPerformanceGoal menolak goal yang diajukan oleh pekerja.
func (s *Service) RejectPerformanceGoal(ctx context.Context, goalID string) (*entity.PerformanceGoal, error) {
	if s == nil || operation.RequireDependency(s.dependencies.Worker) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Command(ctx, goalID, "reject", func(txCtx context.Context, goal *entity.PerformanceGoal, actor operation.Actor, _ time.Time) (*entity.PerformanceGoal, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if goal.GoalStatus != "submitted" || goal.ApprovedByPrincipalID != nil || goal.ApprovedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, goal.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		goal.GoalStatus = "rejected"
		return goal, nil
	})
}

// SubmitPerformanceReview mengirimkan self-review draf sebelum proses finalisasi peninjau.
func (s *Service) SubmitPerformanceReview(ctx context.Context, reviewID string) (*entity.PerformanceReview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReview.Command(ctx, reviewID, "submit", func(_ context.Context, review *entity.PerformanceReview, actor operation.Actor, now time.Time) (*entity.PerformanceReview, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if !strings.EqualFold(review.ReviewerPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if review.SubmittedAt != nil || review.FinalizedAt != nil ||
			review.ReviewerRole != "self" || review.VisibilityClass != "restricted" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		review.SubmittedAt = &now
		return review, nil
	})
}

// FinalizePerformanceReview membekukan review setelah seluruh goal disetujui memiliki skor dengan total bobot 100%.
func (s *Service) FinalizePerformanceReview(ctx context.Context, reviewID string) (*entity.PerformanceReview, error) {
	if s == nil || operation.RequireDependency(s.dependencies.Worker) != nil ||
		operation.RequireDependency(s.dependencies.PerformanceGoal) != nil ||
		operation.RequireDependency(s.dependencies.PerformanceReviewGoalScore) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReview.Command(ctx, reviewID, "finalize", func(txCtx context.Context, review *entity.PerformanceReview, actor operation.Actor, now time.Time) (*entity.PerformanceReview, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if review.FinalizedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		worker, err := s.dependencies.Worker.FindByIDForUpdate(txCtx, review.WorkerID)
		if err != nil {
			return nil, referenceConflictOnMissing(err)
		}
		if worker == nil || worker.Status != entity.WorkerStatusActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		approvedGoals, err := collectAllPages(func(limit, offset int) ([]entity.PerformanceGoal, error) {
			return s.dependencies.PerformanceGoal.SelectWhere(txCtx, []domainrepo.Filter{
				{Column: "worker_id", Values: []any{review.WorkerID}},
				{Column: "performance_cycle_id", Values: []any{review.PerformanceCycleID}},
				{Column: "goal_status", Values: []any{"approved"}},
			}, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		scores, err := collectAllPages(func(limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
			return s.dependencies.PerformanceReviewGoalScore.SelectWhere(txCtx, []domainrepo.Filter{
				{Column: "performance_review_id", Values: []any{review.ID}},
			}, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		if err := performancepolicy.ValidateReviewGoals(review, approvedGoals, scores); err != nil {
			return nil, err
		}
		if review.SubmittedAt == nil {
			review.SubmittedAt = &now
		} else if now.Before(*review.SubmittedAt) {
			return nil, hrerrors.ErrInvalidInput
		}
		review.FinalizedAt = &now
		return review, nil
	})
}

// ClosePerformanceCycle menutup siklus kinerja yang aktif apabila seluruh goal dan review telah tuntas.
func (s *Service) ClosePerformanceCycle(ctx context.Context, cycleID string) (*entity.PerformanceCycle, error) {
	if s == nil || operation.RequireDependency(s.dependencies.PerformanceGoal) != nil ||
		operation.RequireDependency(s.dependencies.PerformanceReview) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Command(ctx, cycleID, "close", func(txCtx context.Context, cycle *entity.PerformanceCycle, actor operation.Actor, _ time.Time) (*entity.PerformanceCycle, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if cycle.Status != entity.PerformanceCycleStatusOpen {
			return nil, hrerrors.ErrWorkflowRequired
		}
		pendingGoals, err := s.dependencies.PerformanceGoal.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "performance_cycle_id", Values: []any{cycle.ID}},
			{Column: "goal_status", Values: []any{"draft", "submitted"}},
		})
		if err != nil {
			return nil, err
		}
		if pendingGoals != 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		reviews, err := collectAllPages(func(limit, offset int) ([]entity.PerformanceReview, error) {
			return s.dependencies.PerformanceReview.SelectWhere(txCtx, []domainrepo.Filter{
				{Column: "performance_cycle_id", Values: []any{cycle.ID}},
			}, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		for _, review := range reviews {
			if review.FinalizedAt == nil {
				return nil, hrerrors.ErrWorkflowRequired
			}
		}
		cycle.Status = entity.PerformanceCycleStatusFinalized
		return cycle, nil
	})
}

func collectAllPages[T any](fetch func(limit, offset int) ([]T, error)) ([]T, error) {
	items := make([]T, 0)
	for offset := 0; ; {
		page, err := fetch(domainrepo.MaxLimit, offset)
		if err != nil {
			return nil, err
		}
		items = append(items, page...)
		if len(page) < domainrepo.MaxLimit {
			return items, nil
		}
		offset += len(page)
	}
}

func referenceConflictOnMissing(err error) error {
	if errors.Is(err, hrerrors.ErrNotFound) {
		return hrerrors.ErrReferenceConflict
	}
	return err
}
