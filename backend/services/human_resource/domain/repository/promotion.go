package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// PromotionRecommendationRepository adalah port persistensi hr_promotion_recommendations.
type PromotionRecommendationRepository interface {
	Repository[entity.PromotionRecommendation, string]
	InsertPromotionRecommendations(context.Context, *entity.PromotionRecommendation) error
	UpdatePromotionRecommendations(context.Context, *entity.PromotionRecommendation) error
	FindPromotionRecommendationsByUUID(context.Context, string) (*entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByUUIDForUpdate(context.Context, string) (*entity.PromotionRecommendation, error)
	DeletePromotionRecommendationsByUUID(context.Context, string) error
	GetAllPromotionRecommendations(context.Context, int, int) ([]entity.PromotionRecommendation, error)
	CountPromotionRecommendations(context.Context) (int64, error)
	SelectPromotionRecommendationsWhere(context.Context, []Filter, int, int) ([]entity.PromotionRecommendation, error)
	CountPromotionRecommendationsWhere(context.Context, []Filter) (int64, error)
	FindPromotionRecommendationsByRecommendationNumber(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByWorkerID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsBySourceAssignmentID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByTargetPositionID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByAssessmentID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsBySupersedesRecommendationID(context.Context, *string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByRationaleDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByRecommendedByPrincipalID(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByStatus(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByApprovalDecisionID(context.Context, *string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByAppliedAssignmentID(context.Context, *string, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByAppliedLifecycleEventID(context.Context, *int64, int, int) ([]entity.PromotionRecommendation, error)
	FindPromotionRecommendationsByIdempotencyKey(context.Context, string, int, int) ([]entity.PromotionRecommendation, error)
	FindByRecommendationNumber(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindBySourceAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByTargetPositionID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindBySupersedesRecommendationID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByRecommendedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByApprovalDecisionID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByAppliedAssignmentID(ctx context.Context, value *string, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByAppliedLifecycleEventID(ctx context.Context, value *int64, limit, offset int) ([]entity.PromotionRecommendation, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionRecommendation, error)
}

// PromotionDecisionRepository adalah port persistensi hr_promotion_decisions.
type PromotionDecisionRepository interface {
	Repository[entity.PromotionDecision, string]
	InsertPromotionDecisions(context.Context, *entity.PromotionDecision) error
	UpdatePromotionDecisions(context.Context, *entity.PromotionDecision) error
	FindPromotionDecisionsByUUID(context.Context, string) (*entity.PromotionDecision, error)
	FindPromotionDecisionsByUUIDForUpdate(context.Context, string) (*entity.PromotionDecision, error)
	DeletePromotionDecisionsByUUID(context.Context, string) error
	GetAllPromotionDecisions(context.Context, int, int) ([]entity.PromotionDecision, error)
	CountPromotionDecisions(context.Context) (int64, error)
	SelectPromotionDecisionsWhere(context.Context, []Filter, int, int) ([]entity.PromotionDecision, error)
	CountPromotionDecisionsWhere(context.Context, []Filter) (int64, error)
	FindPromotionDecisionsByRecommendationID(context.Context, string, int, int) ([]entity.PromotionDecision, error)
	FindPromotionDecisionsByStepNumber(context.Context, int32, int, int) ([]entity.PromotionDecision, error)
	FindPromotionDecisionsByDecidedByPrincipalID(context.Context, string, int, int) ([]entity.PromotionDecision, error)
	FindPromotionDecisionsByRationaleDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PromotionDecision, error)
	FindPromotionDecisionsByIdempotencyKey(context.Context, string, int, int) ([]entity.PromotionDecision, error)
	FindByRecommendationID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error)
	FindByStepNumber(ctx context.Context, value int32, limit, offset int) ([]entity.PromotionDecision, error)
	FindByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error)
	FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PromotionDecision, error)
}
