package performance

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PerformanceCycleRepository) InsertPerformanceCycles(ctx context.Context, v *entity.PerformanceCycle) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceCycleRepository) UpdatePerformanceCycles(ctx context.Context, v *entity.PerformanceCycle) error {
	return r.Update(ctx, v)
}
func (r *PerformanceCycleRepository) DeletePerformanceCyclesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByUUID(ctx context.Context, id string) (*entity.PerformanceCycle, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceCycle, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceCycleRepository) GetAllPerformanceCycles(ctx context.Context, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceCycleRepository) CountPerformanceCycles(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceCycleRepository) SelectPerformanceCyclesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceCycleRepository) CountPerformanceCyclesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByCycleCode(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.FindByCycleCode(ctx, v, limit, offset)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByCycleType(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.FindByCycleType(ctx, v, limit, offset)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByStatus(ctx context.Context, v entity.PerformanceCycleStatus, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *PerformanceCycleRepository) FindPerformanceCyclesByLaunchedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.FindByLaunchedByPrincipalID(ctx, v, limit, offset)
}

func (r *PerformanceGoalRepository) InsertPerformanceGoals(ctx context.Context, v *entity.PerformanceGoal) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceGoalRepository) UpdatePerformanceGoals(ctx context.Context, v *entity.PerformanceGoal) error {
	return r.Update(ctx, v)
}
func (r *PerformanceGoalRepository) DeletePerformanceGoalsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByUUID(ctx context.Context, id string) (*entity.PerformanceGoal, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceGoal, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceGoalRepository) GetAllPerformanceGoals(ctx context.Context, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceGoalRepository) CountPerformanceGoals(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceGoalRepository) SelectPerformanceGoalsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceGoalRepository) CountPerformanceGoalsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByPerformanceCycleID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByPerformanceCycleID(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByParentGoalID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByParentGoalID(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByGoalCode(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByGoalCode(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByGoalType(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByGoalType(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByUnitCode(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByUnitCode(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByGoalStatus(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByGoalStatus(ctx, v, limit, offset)
}
func (r *PerformanceGoalRepository) FindPerformanceGoalsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}

func (r *PerformanceReviewRepository) InsertPerformanceReviews(ctx context.Context, v *entity.PerformanceReview) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceReviewRepository) UpdatePerformanceReviews(ctx context.Context, v *entity.PerformanceReview) error {
	return r.Update(ctx, v)
}
func (r *PerformanceReviewRepository) DeletePerformanceReviewsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByUUID(ctx context.Context, id string) (*entity.PerformanceReview, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceReview, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceReviewRepository) GetAllPerformanceReviews(ctx context.Context, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceReviewRepository) CountPerformanceReviews(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceReviewRepository) SelectPerformanceReviewsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceReviewRepository) CountPerformanceReviewsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByPerformanceCycleID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByPerformanceCycleID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByReviewerPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByReviewerPrincipalID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByOverallRatingCode(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByOverallRatingCode(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByDevelopmentPlanID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByDevelopmentPlanID(ctx, v, limit, offset)
}

func (r *PerformanceGoalUpdateRepository) InsertPerformanceGoalUpdates(ctx context.Context, v *entity.PerformanceGoalUpdate) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceGoalUpdateRepository) UpdatePerformanceGoalUpdates(ctx context.Context, v *entity.PerformanceGoalUpdate) error {
	return r.Update(ctx, v)
}
func (r *PerformanceGoalUpdateRepository) DeletePerformanceGoalUpdatesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByUUID(ctx context.Context, id string) (*entity.PerformanceGoalUpdate, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceGoalUpdate, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) GetAllPerformanceGoalUpdates(ctx context.Context, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) CountPerformanceGoalUpdates(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceGoalUpdateRepository) SelectPerformanceGoalUpdatesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) CountPerformanceGoalUpdatesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByGoalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByGoalID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByEvidenceDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByRecordedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByRecordedByPrincipalID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesBySupersedesUpdateID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindBySupersedesUpdateID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

func (r *PerformanceReviewGoalScoreRepository) InsertPerformanceReviewGoalScores(ctx context.Context, v *entity.PerformanceReviewGoalScore) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceReviewGoalScoreRepository) UpdatePerformanceReviewGoalScores(ctx context.Context, v *entity.PerformanceReviewGoalScore) error {
	return r.Update(ctx, v)
}
func (r *PerformanceReviewGoalScoreRepository) DeletePerformanceReviewGoalScoresByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByUUID(ctx context.Context, id string) (*entity.PerformanceReviewGoalScore, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceReviewGoalScore, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) GetAllPerformanceReviewGoalScores(ctx context.Context, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) CountPerformanceReviewGoalScores(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceReviewGoalScoreRepository) SelectPerformanceReviewGoalScoresWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) CountPerformanceReviewGoalScoresWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByPerformanceReviewID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByPerformanceReviewID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByGoalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByGoalID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByPerformanceCycleID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByPerformanceCycleID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByRubricDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByRubricDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByEvidenceDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, v, limit, offset)
}
