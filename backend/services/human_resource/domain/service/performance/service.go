// Package service memeriksa hubungan goal, review dan bobot skor performance.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateReviewGoals memeriksa seluruh goal yang menjadi cakupan review sebelum finalisasi.
func ValidateReviewGoals(review *entity.PerformanceReview, goals []entity.PerformanceGoal, scores []entity.PerformanceReviewGoalScore) error {
	if err := review.Validate(); err != nil {
		return err
	}
	if review.FinalizedAt != nil {
		return hrerrors.Invalid("finalized_at", "finalized review requires a new revision")
	}
	indexed := map[string]entity.PerformanceGoal{}
	for _, goal := range goals {
		if err := goal.Validate(); err != nil {
			return err
		}
		key := strings.ToLower(goal.ID)
		if _, exists := indexed[key]; exists {
			return hrerrors.Invalid("goals", "duplicate goal")
		}
		if !strings.EqualFold(goal.WorkerID, review.WorkerID) || !strings.EqualFold(goal.PerformanceCycleID, review.PerformanceCycleID) {
			return hrerrors.Invalid("goals", "goals must belong to the reviewed worker and cycle")
		}
		indexed[key] = goal
	}
	seen := map[string]bool{}
	ids := map[string]bool{}
	weights := []entity.Decimal{}
	for _, score := range scores {
		if err := score.Validate(); err != nil {
			return err
		}
		key := strings.ToLower(score.GoalID)
		id := strings.ToLower(score.ID)
		goal, exists := indexed[key]
		if !exists || seen[key] || ids[id] || !strings.EqualFold(score.PerformanceReviewID, review.ID) || !strings.EqualFold(score.WorkerID, review.WorkerID) || !strings.EqualFold(score.PerformanceCycleID, review.PerformanceCycleID) {
			return hrerrors.Invalid("scores", "score references must match each reviewed goal exactly once")
		}
		if goal.WeightPercent == nil {
			return hrerrors.Invalid("weight_percent", "scored goal requires an approved weight")
		}
		wanted, err := goal.WeightPercent.Rat()
		if err != nil {
			return err
		}
		actual, err := score.WeightPercentSnapshot.Rat()
		if err != nil {
			return err
		}
		if wanted.Cmp(actual) != 0 {
			return hrerrors.Invalid("weight_percent_snapshot", "weight snapshot must match the goal")
		}
		seen[key] = true
		ids[id] = true
		weights = append(weights, score.WeightPercentSnapshot)
	}
	total, err := entity.DecimalSum(3, weights...)
	if err != nil {
		return err
	}
	number, err := total.Rat()
	if err != nil {
		return err
	}
	hundred, err := entity.Decimal("100").Rat()
	if err != nil {
		return err
	}
	if len(seen) != len(indexed) || number.Cmp(hundred) != 0 {
		return hrerrors.Invalid("scores", "all reviewed goals must be scored with total weight one hundred")
	}
	return nil
}
