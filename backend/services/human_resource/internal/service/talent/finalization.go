package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	talentpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ReviewReader menggunakan service pemilik review agar pembacaan tetap melewati policy performance.
type ReviewReader interface {
	GetPerformanceReviewByID(context.Context, string) (*entity.PerformanceReview, error)
}

// FinalizeTalentAssessment memeriksa seluruh dimensi dan skor di bawah lock assessment.
// Finalisasi menyimpan fakta asesmen; tidak menerapkan promosi atau grant IAM.
func (s *Service) FinalizeTalentAssessment(ctx context.Context, assessmentID string) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Command(ctx, assessmentID, "finalize", func(txCtx context.Context, assessment *entity.TalentAssessment, actor operation.Actor, now time.Time) (*entity.TalentAssessment, error) {
		cycle, err := s.talentReviewCycle.GetForUpdate(txCtx, assessment.TalentReviewCycleID)
		if err != nil {
			return nil, err
		}
		framework, err := s.talentFramework.Get(txCtx, assessment.FrameworkID)
		if err != nil {
			return nil, err
		}
		dimensionFilter := []domainrepo.Filter{{Column: "framework_id", Values: []any{framework.ID}}}
		dimensionCount, err := s.dependencies.TalentDimension.CountWhere(txCtx, dimensionFilter)
		if err != nil {
			return nil, err
		}
		dimensions, err := operation.CollectComplete(txCtx, dimensionCount, func(limit, offset int) (operation.Page[entity.TalentDimension], error) {
			return s.talentDimension.List(txCtx, dimensionFilter, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		scoreFilter := []domainrepo.Filter{{Column: "assessment_id", Values: []any{assessmentID}}}
		scoreCount, err := s.dependencies.TalentAssessmentScore.CountWhere(txCtx, scoreFilter)
		if err != nil {
			return nil, err
		}
		scores, err := operation.CollectComplete(txCtx, scoreCount, func(limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
			return s.talentAssessmentScore.List(txCtx, scoreFilter, limit, offset)
		})
		if err != nil {
			return nil, err
		}
		var review *entity.PerformanceReview
		if assessment.PerformanceReviewID != nil {
			if err := operation.RequireDependency(s.dependencies.PerformanceReviewReader); err != nil {
				return nil, err
			}
			review, err = s.dependencies.PerformanceReviewReader.GetPerformanceReviewByID(txCtx, *assessment.PerformanceReviewID)
			if err != nil {
				return nil, err
			}
		}
		if err := talentpolicy.ValidateFinalization(talentpolicy.FinalizationInput{
			Assessment: assessment, Cycle: cycle, Framework: framework,
			Dimensions: dimensions, Scores: scores, PerformanceReview: review,
		}); err != nil {
			return nil, err
		}
		assessment.Status = "finalized"
		assessment.FinalizedAt = &now
		assessment.FinalizedByPrincipalID = &actor.PrincipalID
		return assessment, nil
	})
}
