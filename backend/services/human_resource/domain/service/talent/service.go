// Package service memeriksa konsistensi rubrik dan bukti asesmen talent HRM.
package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateFinalization memeriksa rubrik/skor lengkap tanpa memutuskan promosi atau mandat actor.
func ValidateFinalization(input FinalizationInput) error {
	for _, record := range []entity.Validatable{input.Assessment, input.Cycle, input.Framework} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	a, c, f := input.Assessment, input.Cycle, input.Framework
	if (a.Status != "draft" && a.Status != "submitted") || (c.Status != "open" && c.Status != "calibration") {
		return hrerrors.Invalid("status", "assessment must be mutable within an open review cycle")
	}
	if !strings.EqualFold(a.TalentReviewCycleID, c.ID) || !strings.EqualFold(a.FrameworkID, f.ID) || !strings.EqualFold(c.FrameworkID, f.ID) {
		return hrerrors.Invalid("framework_id", "assessment, cycle and framework must match")
	}
	if entity.CompareDates(a.AssessedAt, c.PeriodStart) < 0 || entity.CompareDates(a.AssessedAt, c.PeriodEnd) > 0 || entity.CompareDates(a.AssessedAt, f.EffectiveFrom) < 0 || (f.EffectiveTo != nil && entity.CompareDates(a.AssessedAt, *f.EffectiveTo) > 0) {
		return hrerrors.Invalid("assessed_at", "assessment must fall within the cycle and framework validity")
	}
	if a.PerformanceReviewID != nil {
		review := input.PerformanceReview
		if err := review.Validate(); err != nil {
			return err
		}
		if !strings.EqualFold(*a.PerformanceReviewID, review.ID) || !strings.EqualFold(review.WorkerID, a.WorkerID) || review.FinalizedAt == nil {
			return hrerrors.Invalid("performance_review_id", "performance evidence must be finalized for the same worker")
		}
	}
	dimensions := map[string]entity.TalentDimension{}
	codes := map[string]bool{}
	weights := []entity.Decimal{}
	for _, dimension := range input.Dimensions {
		if err := dimension.Validate(); err != nil {
			return err
		}
		key := strings.ToLower(dimension.ID)
		_, duplicate := dimensions[key]
		if duplicate || codes[dimension.DimensionCode] || !strings.EqualFold(dimension.FrameworkID, f.ID) {
			return hrerrors.Invalid("dimensions", "framework dimensions must be unique and consistent")
		}
		var anchors map[string]string
		if err := json.Unmarshal(dimension.RatingAnchors, &anchors); err != nil {
			return hrerrors.Invalid("rating_anchors", "must define string evidence for levels one to five")
		}
		for level := 1; level <= 5; level++ {
			if strings.TrimSpace(anchors[fmt.Sprint(level)]) == "" {
				return hrerrors.Invalid("rating_anchors", "all five rating anchors are required")
			}
		}
		dimensions[key] = dimension
		codes[dimension.DimensionCode] = true
		weights = append(weights, dimension.WeightPercent)
	}
	total, err := entity.DecimalSum(3, weights...)
	if err != nil {
		return err
	}
	number, err := total.Rat()
	if err != nil {
		return err
	}
	expected, err := entity.Decimal("100").Rat()
	if err != nil {
		return err
	}
	if number.Cmp(expected) != 0 {
		return hrerrors.Invalid("weight_percent", "dimension weights must sum to one hundred")
	}
	seen := map[string]bool{}
	scoreIDs := map[string]bool{}
	for _, score := range input.Scores {
		if err := score.Validate(); err != nil {
			return err
		}
		dimension := strings.ToLower(score.DimensionID)
		scoreID := strings.ToLower(score.ID)
		_, exists := dimensions[dimension]
		if !exists || seen[dimension] || scoreIDs[scoreID] || !strings.EqualFold(score.AssessmentID, a.ID) || !strings.EqualFold(score.FrameworkID, f.ID) {
			return hrerrors.Invalid("scores", "scores must uniquely match the assessment framework")
		}
		seen[dimension] = true
		scoreIDs[scoreID] = true
	}
	if len(seen) == 0 {
		return hrerrors.Invalid("scores", "at least one dimension score is required")
	}
	for id, dimension := range dimensions {
		if dimension.IsRequired && !seen[id] {
			return hrerrors.Invalid("scores", "required dimension score is missing")
		}
	}
	return nil
}

// ValidateCalibration menjaga skor asli dan keputusan terpisah dalam sesi terbuka.
func ValidateCalibration(session *entity.TalentCalibrationSession, assessment *entity.TalentAssessment, source *entity.TalentAssessmentScore, decision *entity.TalentCalibrationDecision) error {
	for _, record := range []entity.Validatable{session, assessment, source, decision} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if session.Status != "open" || assessment.Status != "finalized" ||
		!strings.EqualFold(session.ID, decision.SessionID) ||
		!strings.EqualFold(session.TalentReviewCycleID, assessment.TalentReviewCycleID) ||
		!strings.EqualFold(decision.TalentReviewCycleID, assessment.TalentReviewCycleID) ||
		!strings.EqualFold(source.AssessmentID, assessment.ID) ||
		!strings.EqualFold(source.FrameworkID, assessment.FrameworkID) ||
		!strings.EqualFold(decision.AssessmentID, assessment.ID) ||
		!strings.EqualFold(decision.DimensionID, source.DimensionID) ||
		decision.OriginalScore != source.Score {
		return hrerrors.Invalid("calibration", "open session and original finalized assessment evidence must match")
	}
	return nil
}
