// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// PerformanceCycleRepository adalah port persistensi hr_performance_cycles.
type PerformanceCycleRepository interface {
	Repository[entity.PerformanceCycle, string]
	InsertPerformanceCycles(context.Context, *entity.PerformanceCycle) error
	UpdatePerformanceCycles(context.Context, *entity.PerformanceCycle) error
	DeletePerformanceCyclesByUUID(context.Context, string) error
	FindPerformanceCyclesByUUID(context.Context, string) (*entity.PerformanceCycle, error)
	FindPerformanceCyclesByUUIDForUpdate(context.Context, string) (*entity.PerformanceCycle, error)
	GetAllPerformanceCycles(context.Context, int, int) ([]entity.PerformanceCycle, error)
	CountPerformanceCycles(context.Context) (int64, error)
	SelectPerformanceCyclesWhere(context.Context, []Filter, int, int) ([]entity.PerformanceCycle, error)
	CountPerformanceCyclesWhere(context.Context, []Filter) (int64, error)
	FindPerformanceCyclesByCycleCode(context.Context, string, int, int) ([]entity.PerformanceCycle, error)
	FindPerformanceCyclesByCycleType(context.Context, string, int, int) ([]entity.PerformanceCycle, error)
	FindPerformanceCyclesByStatus(context.Context, entity.PerformanceCycleStatus, int, int) ([]entity.PerformanceCycle, error)
	FindPerformanceCyclesByLaunchedByPrincipalID(context.Context, *string, int, int) ([]entity.PerformanceCycle, error)
	FindByCycleCode(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceCycle, error)
	FindByCycleType(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceCycle, error)
	FindByStatus(ctx context.Context, value entity.PerformanceCycleStatus, limit, offset int) ([]entity.PerformanceCycle, error)
	FindByLaunchedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceCycle, error)
}

// PerformanceGoalRepository adalah port persistensi hr_performance_goals.
type PerformanceGoalRepository interface {
	Repository[entity.PerformanceGoal, string]
	InsertPerformanceGoals(context.Context, *entity.PerformanceGoal) error
	UpdatePerformanceGoals(context.Context, *entity.PerformanceGoal) error
	DeletePerformanceGoalsByUUID(context.Context, string) error
	FindPerformanceGoalsByUUID(context.Context, string) (*entity.PerformanceGoal, error)
	FindPerformanceGoalsByUUIDForUpdate(context.Context, string) (*entity.PerformanceGoal, error)
	GetAllPerformanceGoals(context.Context, int, int) ([]entity.PerformanceGoal, error)
	CountPerformanceGoals(context.Context) (int64, error)
	SelectPerformanceGoalsWhere(context.Context, []Filter, int, int) ([]entity.PerformanceGoal, error)
	CountPerformanceGoalsWhere(context.Context, []Filter) (int64, error)
	FindPerformanceGoalsByWorkerID(context.Context, string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByPerformanceCycleID(context.Context, string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByParentGoalID(context.Context, *string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByGoalCode(context.Context, *string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByGoalType(context.Context, string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByUnitCode(context.Context, *string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByGoalStatus(context.Context, string, int, int) ([]entity.PerformanceGoal, error)
	FindPerformanceGoalsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.PerformanceGoal, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByParentGoalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByGoalCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByGoalType(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByUnitCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByGoalStatus(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error)
}

// PerformanceReviewRepository adalah port persistensi hr_performance_reviews.
type PerformanceReviewRepository interface {
	Repository[entity.PerformanceReview, string]
	InsertPerformanceReviews(context.Context, *entity.PerformanceReview) error
	UpdatePerformanceReviews(context.Context, *entity.PerformanceReview) error
	DeletePerformanceReviewsByUUID(context.Context, string) error
	FindPerformanceReviewsByUUID(context.Context, string) (*entity.PerformanceReview, error)
	FindPerformanceReviewsByUUIDForUpdate(context.Context, string) (*entity.PerformanceReview, error)
	GetAllPerformanceReviews(context.Context, int, int) ([]entity.PerformanceReview, error)
	CountPerformanceReviews(context.Context) (int64, error)
	SelectPerformanceReviewsWhere(context.Context, []Filter, int, int) ([]entity.PerformanceReview, error)
	CountPerformanceReviewsWhere(context.Context, []Filter) (int64, error)
	FindPerformanceReviewsByWorkerID(context.Context, string, int, int) ([]entity.PerformanceReview, error)
	FindPerformanceReviewsByPerformanceCycleID(context.Context, string, int, int) ([]entity.PerformanceReview, error)
	FindPerformanceReviewsByReviewerPrincipalID(context.Context, string, int, int) ([]entity.PerformanceReview, error)
	FindPerformanceReviewsByOverallRatingCode(context.Context, *string, int, int) ([]entity.PerformanceReview, error)
	FindPerformanceReviewsByDevelopmentPlanID(context.Context, *string, int, int) ([]entity.PerformanceReview, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error)
	FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error)
	FindByReviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error)
	FindByOverallRatingCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceReview, error)
	FindByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceReview, error)
}

// PerformanceGoalUpdateRepository adalah port persistensi hr_performance_goal_updates.
type PerformanceGoalUpdateRepository interface {
	Repository[entity.PerformanceGoalUpdate, string]
	InsertPerformanceGoalUpdates(context.Context, *entity.PerformanceGoalUpdate) error
	UpdatePerformanceGoalUpdates(context.Context, *entity.PerformanceGoalUpdate) error
	DeletePerformanceGoalUpdatesByUUID(context.Context, string) error
	FindPerformanceGoalUpdatesByUUID(context.Context, string) (*entity.PerformanceGoalUpdate, error)
	FindPerformanceGoalUpdatesByUUIDForUpdate(context.Context, string) (*entity.PerformanceGoalUpdate, error)
	GetAllPerformanceGoalUpdates(context.Context, int, int) ([]entity.PerformanceGoalUpdate, error)
	CountPerformanceGoalUpdates(context.Context) (int64, error)
	SelectPerformanceGoalUpdatesWhere(context.Context, []Filter, int, int) ([]entity.PerformanceGoalUpdate, error)
	CountPerformanceGoalUpdatesWhere(context.Context, []Filter) (int64, error)
	FindPerformanceGoalUpdatesByGoalID(context.Context, string, int, int) ([]entity.PerformanceGoalUpdate, error)
	FindPerformanceGoalUpdatesByEvidenceDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PerformanceGoalUpdate, error)
	FindPerformanceGoalUpdatesByRecordedByPrincipalID(context.Context, string, int, int) ([]entity.PerformanceGoalUpdate, error)
	FindPerformanceGoalUpdatesBySupersedesUpdateID(context.Context, *string, int, int) ([]entity.PerformanceGoalUpdate, error)
	FindPerformanceGoalUpdatesByIdempotencyKey(context.Context, string, int, int) ([]entity.PerformanceGoalUpdate, error)
	FindByGoalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error)
	FindByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error)
	FindBySupersedesUpdateID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoalUpdate, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error)
}

// PerformanceReviewGoalScoreRepository adalah port persistensi hr_performance_review_goal_scores.
type PerformanceReviewGoalScoreRepository interface {
	Repository[entity.PerformanceReviewGoalScore, string]
	InsertPerformanceReviewGoalScores(context.Context, *entity.PerformanceReviewGoalScore) error
	UpdatePerformanceReviewGoalScores(context.Context, *entity.PerformanceReviewGoalScore) error
	DeletePerformanceReviewGoalScoresByUUID(context.Context, string) error
	FindPerformanceReviewGoalScoresByUUID(context.Context, string) (*entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByUUIDForUpdate(context.Context, string) (*entity.PerformanceReviewGoalScore, error)
	GetAllPerformanceReviewGoalScores(context.Context, int, int) ([]entity.PerformanceReviewGoalScore, error)
	CountPerformanceReviewGoalScores(context.Context) (int64, error)
	SelectPerformanceReviewGoalScoresWhere(context.Context, []Filter, int, int) ([]entity.PerformanceReviewGoalScore, error)
	CountPerformanceReviewGoalScoresWhere(context.Context, []Filter) (int64, error)
	FindPerformanceReviewGoalScoresByPerformanceReviewID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByGoalID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByWorkerID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByPerformanceCycleID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByRubricDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindPerformanceReviewGoalScoresByEvidenceDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.PerformanceReviewGoalScore, error)
	FindByPerformanceReviewID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
	FindByGoalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
	FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
	FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error)
}
