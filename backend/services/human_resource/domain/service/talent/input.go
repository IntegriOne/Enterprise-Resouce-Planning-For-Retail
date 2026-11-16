package service

import "github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"

// FinalizationInput membawa snapshot lengkap rubrik dan skor untuk pemeriksaan domain.
type FinalizationInput struct {
	Assessment        *entity.TalentAssessment
	Cycle             *entity.TalentReviewCycle
	Framework         *entity.TalentFramework
	Dimensions        []entity.TalentDimension
	Scores            []entity.TalentAssessmentScore
	PerformanceReview *entity.PerformanceReview
}
