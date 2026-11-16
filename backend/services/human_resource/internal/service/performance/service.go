// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime                    operation.Dependencies
	PerformanceCycle           operation.Repository[entity.PerformanceCycle, string]
	PerformanceGoal            operation.Repository[entity.PerformanceGoal, string]
	PerformanceReview          operation.Repository[entity.PerformanceReview, string]
	PerformanceGoalUpdate      operation.Repository[entity.PerformanceGoalUpdate, string]
	PerformanceReviewGoalScore operation.Repository[entity.PerformanceReviewGoalScore, string]
}

// Service mengoordinasikan operasi domain performance.
type Service struct {
	performanceCycle           *operation.Resource[entity.PerformanceCycle, string, *entity.PerformanceCycle]
	performanceGoal            *operation.Resource[entity.PerformanceGoal, string, *entity.PerformanceGoal]
	performanceReview          *operation.Resource[entity.PerformanceReview, string, *entity.PerformanceReview]
	performanceGoalUpdate      *operation.Resource[entity.PerformanceGoalUpdate, string, *entity.PerformanceGoalUpdate]
	performanceReviewGoalScore *operation.Resource[entity.PerformanceReviewGoalScore, string, *entity.PerformanceReviewGoalScore]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.performanceCycle, err = operation.NewResource[entity.PerformanceCycle, string, *entity.PerformanceCycle](deps.PerformanceCycle, deps.Runtime, specPerformanceCycle())
	if err != nil {
		return nil, err
	}
	s.performanceGoal, err = operation.NewResource[entity.PerformanceGoal, string, *entity.PerformanceGoal](deps.PerformanceGoal, deps.Runtime, specPerformanceGoal())
	if err != nil {
		return nil, err
	}
	s.performanceReview, err = operation.NewResource[entity.PerformanceReview, string, *entity.PerformanceReview](deps.PerformanceReview, deps.Runtime, specPerformanceReview())
	if err != nil {
		return nil, err
	}
	s.performanceGoalUpdate, err = operation.NewResource[entity.PerformanceGoalUpdate, string, *entity.PerformanceGoalUpdate](deps.PerformanceGoalUpdate, deps.Runtime, specPerformanceGoalUpdate())
	if err != nil {
		return nil, err
	}
	s.performanceReviewGoalScore, err = operation.NewResource[entity.PerformanceReviewGoalScore, string, *entity.PerformanceReviewGoalScore](deps.PerformanceReviewGoalScore, deps.Runtime, specPerformanceReviewGoalScore())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreatePerformanceCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePerformanceCycle(ctx context.Context, in CreatePerformanceCycleInput) (*entity.PerformanceCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Create(ctx, func() (*entity.PerformanceCycle, error) { return createPerformanceCycle(in) })
}

// UpdatePerformanceCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePerformanceCycle(ctx context.Context, key string, in UpdatePerformanceCycleInput) (*entity.PerformanceCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Update(ctx, key, func(value *entity.PerformanceCycle) error { return patchPerformanceCycle(value, in) })
}

// DeletePerformanceCycle menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePerformanceCycle(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Delete(ctx, key)
}

// GetPerformanceCycleByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPerformanceCycleByID(ctx context.Context, key string) (*entity.PerformanceCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.Get(ctx, key)
}

// GetAllPerformanceCycles menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPerformanceCycles(ctx context.Context, limit, offset int) (operation.Page[entity.PerformanceCycle], error) {
	if s == nil {
		return operation.Page[entity.PerformanceCycle]{}, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.List(ctx, nil, limit, offset)
}

// FindPerformanceCycleByCycleCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceCycleByCycleCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceCycle], error) {
	if s == nil {
		return operation.Page[entity.PerformanceCycle]{}, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.List(ctx, []domainrepo.Filter{{Column: "cycle_code", Values: []any{value}}}, limit, offset)
}

// FindPerformanceCycleByCycleType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceCycleByCycleType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceCycle], error) {
	if s == nil {
		return operation.Page[entity.PerformanceCycle]{}, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.List(ctx, []domainrepo.Filter{{Column: "cycle_type", Values: []any{value}}}, limit, offset)
}

// FindPerformanceCycleByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceCycleByStatus(ctx context.Context, value entity.PerformanceCycleStatus, limit, offset int) (operation.Page[entity.PerformanceCycle], error) {
	if s == nil {
		return operation.Page[entity.PerformanceCycle]{}, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindPerformanceCycleByLaunchedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceCycleByLaunchedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceCycle], error) {
	if s == nil {
		return operation.Page[entity.PerformanceCycle]{}, hrerrors.ErrUnavailable
	}
	return s.performanceCycle.List(ctx, []domainrepo.Filter{{Column: "launched_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreatePerformanceGoal menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePerformanceGoal(ctx context.Context, in CreatePerformanceGoalInput) (*entity.PerformanceGoal, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Create(ctx, func() (*entity.PerformanceGoal, error) { return createPerformanceGoal(in) })
}

// UpdatePerformanceGoal menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePerformanceGoal(ctx context.Context, key string, in UpdatePerformanceGoalInput) (*entity.PerformanceGoal, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Update(ctx, key, func(value *entity.PerformanceGoal) error { return patchPerformanceGoal(value, in) })
}

// DeletePerformanceGoal menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePerformanceGoal(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Delete(ctx, key)
}

// GetPerformanceGoalByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPerformanceGoalByID(ctx context.Context, key string) (*entity.PerformanceGoal, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.Get(ctx, key)
}

// GetAllPerformanceGoals menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPerformanceGoals(ctx context.Context, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, nil, limit, offset)
}

// FindPerformanceGoalByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByPerformanceCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByPerformanceCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "performance_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByParentGoalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByParentGoalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "parent_goal_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByGoalCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByGoalCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "goal_code", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByGoalType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByGoalType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "goal_type", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByUnitCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByUnitCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "unit_code", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByGoalStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByGoalStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "goal_status", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceGoal], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoal]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoal.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreatePerformanceReview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePerformanceReview(ctx context.Context, in CreatePerformanceReviewInput) (*entity.PerformanceReview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReview.Create(ctx, func() (*entity.PerformanceReview, error) { return createPerformanceReview(in) })
}

// UpdatePerformanceReview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePerformanceReview(ctx context.Context, key string, in UpdatePerformanceReviewInput) (*entity.PerformanceReview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReview.Update(ctx, key, func(value *entity.PerformanceReview) error { return patchPerformanceReview(value, in) })
}

// DeletePerformanceReview menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePerformanceReview(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.performanceReview.Delete(ctx, key)
}

// GetPerformanceReviewByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPerformanceReviewByID(ctx context.Context, key string) (*entity.PerformanceReview, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReview.Get(ctx, key)
}

// GetAllPerformanceReviews menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPerformanceReviews(ctx context.Context, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, nil, limit, offset)
}

// FindPerformanceReviewByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewByPerformanceCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewByPerformanceCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, []domainrepo.Filter{{Column: "performance_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewByReviewerPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewByReviewerPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, []domainrepo.Filter{{Column: "reviewer_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewByOverallRatingCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewByOverallRatingCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, []domainrepo.Filter{{Column: "overall_rating_code", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewByDevelopmentPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceReview], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReview]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReview.List(ctx, []domainrepo.Filter{{Column: "development_plan_id", Values: []any{value}}}, limit, offset)
}

// CreatePerformanceGoalUpdate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePerformanceGoalUpdate(ctx context.Context, in CreatePerformanceGoalUpdateInput) (*entity.PerformanceGoalUpdate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.Create(ctx, func() (*entity.PerformanceGoalUpdate, error) { return createPerformanceGoalUpdate(in) })
}

// UpdatePerformanceGoalUpdate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePerformanceGoalUpdate(ctx context.Context, key string, in UpdatePerformanceGoalUpdateInput) (*entity.PerformanceGoalUpdate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.Update(ctx, key, func(value *entity.PerformanceGoalUpdate) error { return patchPerformanceGoalUpdate(value, in) })
}

// DeletePerformanceGoalUpdate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePerformanceGoalUpdate(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.Delete(ctx, key)
}

// GetPerformanceGoalUpdateByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPerformanceGoalUpdateByID(ctx context.Context, key string) (*entity.PerformanceGoalUpdate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.Get(ctx, key)
}

// GetAllPerformanceGoalUpdates menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPerformanceGoalUpdates(ctx context.Context, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, nil, limit, offset)
}

// FindPerformanceGoalUpdateByGoalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalUpdateByGoalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, []domainrepo.Filter{{Column: "goal_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalUpdateByRecordedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalUpdateByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, []domainrepo.Filter{{Column: "recorded_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalUpdateBySupersedesUpdateID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalUpdateBySupersedesUpdateID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, []domainrepo.Filter{{Column: "supersedes_update_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceGoalUpdateByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceGoalUpdateByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceGoalUpdate], error) {
	if s == nil {
		return operation.Page[entity.PerformanceGoalUpdate]{}, hrerrors.ErrUnavailable
	}
	return s.performanceGoalUpdate.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreatePerformanceReviewGoalScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePerformanceReviewGoalScore(ctx context.Context, in CreatePerformanceReviewGoalScoreInput) (*entity.PerformanceReviewGoalScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.Create(ctx, func() (*entity.PerformanceReviewGoalScore, error) { return createPerformanceReviewGoalScore(in) })
}

// UpdatePerformanceReviewGoalScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePerformanceReviewGoalScore(ctx context.Context, key string, in UpdatePerformanceReviewGoalScoreInput) (*entity.PerformanceReviewGoalScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.Update(ctx, key, func(value *entity.PerformanceReviewGoalScore) error {
		return patchPerformanceReviewGoalScore(value, in)
	})
}

// DeletePerformanceReviewGoalScore menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePerformanceReviewGoalScore(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.Delete(ctx, key)
}

// GetPerformanceReviewGoalScoreByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPerformanceReviewGoalScoreByID(ctx context.Context, key string) (*entity.PerformanceReviewGoalScore, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.Get(ctx, key)
}

// GetAllPerformanceReviewGoalScores menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPerformanceReviewGoalScores(ctx context.Context, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, nil, limit, offset)
}

// FindPerformanceReviewGoalScoreByPerformanceReviewID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByPerformanceReviewID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "performance_review_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewGoalScoreByGoalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByGoalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "goal_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewGoalScoreByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewGoalScoreByPerformanceCycleID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByPerformanceCycleID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "performance_cycle_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "rubric_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PerformanceReviewGoalScore], error) {
	if s == nil {
		return operation.Page[entity.PerformanceReviewGoalScore]{}, hrerrors.ErrUnavailable
	}
	return s.performanceReviewGoalScore.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}
