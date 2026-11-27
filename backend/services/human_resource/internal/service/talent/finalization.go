package service

import (
	"context"
	"strings"
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
		if err := s.validateAssessmentCompleteness(txCtx, assessment); err != nil {
			return nil, err
		}
		assessment.Status = "finalized"
		assessment.FinalizedAt = &now
		assessment.FinalizedByPrincipalID = &actor.PrincipalID
		return assessment, nil
	})
}

func (s *Service) validateAssessmentCompleteness(ctx context.Context, assessment *entity.TalentAssessment) error {
	cycle, err := s.talentReviewCycle.GetForUpdate(ctx, assessment.TalentReviewCycleID)
	if err != nil {
		return err
	}
	framework, err := s.talentFramework.Get(ctx, assessment.FrameworkID)
	if err != nil {
		return err
	}
	dimensionFilter := []domainrepo.Filter{{Column: "framework_id", Values: []any{framework.ID}}}
	dimensionCount, err := s.dependencies.TalentDimension.CountWhere(ctx, dimensionFilter)
	if err != nil {
		return err
	}
	dimensions, err := operation.CollectComplete(ctx, dimensionCount, func(limit, offset int) (operation.Page[entity.TalentDimension], error) {
		return s.talentDimension.List(ctx, dimensionFilter, limit, offset)
	})
	if err != nil {
		return err
	}
	scoreFilter := []domainrepo.Filter{{Column: "assessment_id", Values: []any{assessment.ID}}}
	scoreCount, err := s.dependencies.TalentAssessmentScore.CountWhere(ctx, scoreFilter)
	if err != nil {
		return err
	}
	scores, err := operation.CollectComplete(ctx, scoreCount, func(limit, offset int) (operation.Page[entity.TalentAssessmentScore], error) {
		return s.talentAssessmentScore.List(ctx, scoreFilter, limit, offset)
	})
	if err != nil {
		return err
	}
	var review *entity.PerformanceReview
	if assessment.PerformanceReviewID != nil {
		if err := operation.RequireDependency(s.dependencies.PerformanceReviewReader); err != nil {
			return err
		}
		review, err = s.dependencies.PerformanceReviewReader.GetPerformanceReviewByID(ctx, *assessment.PerformanceReviewID)
		if err != nil {
			return err
		}
	}
	return talentpolicy.ValidateFinalization(talentpolicy.FinalizationInput{
		Assessment: assessment, Cycle: cycle, Framework: framework,
		Dimensions: dimensions, Scores: scores, PerformanceReview: review,
	})
}

// SubmitTalentAssessment mengajukan asesmen draf milik assessor setelah seluruh skor lengkap.
func (s *Service) SubmitTalentAssessment(ctx context.Context, assessmentID string) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Command(ctx, assessmentID, "submit", func(txCtx context.Context, assessment *entity.TalentAssessment, actor operation.Actor, _ time.Time) (*entity.TalentAssessment, error) {
		if actor.Kind != "human" || !strings.EqualFold(assessment.AssessedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if assessment.Status != "draft" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.validateAssessmentCompleteness(txCtx, assessment); err != nil {
			return nil, err
		}
		assessment.Status = "submitted"
		return assessment, nil
	})
}

// WithdrawTalentAssessment menarik kembali asesmen talent yang belum difinalisasi.
func (s *Service) WithdrawTalentAssessment(ctx context.Context, assessmentID string) (*entity.TalentAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentAssessment.Command(ctx, assessmentID, "withdraw", func(txCtx context.Context, assessment *entity.TalentAssessment, actor operation.Actor, _ time.Time) (*entity.TalentAssessment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if assessment.Status != "draft" && assessment.Status != "submitted" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		cycle, err := s.talentReviewCycle.GetForUpdate(txCtx, assessment.TalentReviewCycleID)
		if err != nil {
			return nil, err
		}
		if cycle.Status != "open" && cycle.Status != "calibration" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		assessment.Status = "withdrawn"
		return assessment, nil
	})
}
