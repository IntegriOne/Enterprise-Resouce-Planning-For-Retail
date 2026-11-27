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

// Dependencies mendefinisikan repository dan komponen runtime untuk domain promotion.
type Dependencies struct {
	Runtime                 operation.Dependencies
	PromotionRecommendation domainrepo.RecordRepository[entity.PromotionRecommendation, string]
	PromotionDecision       domainrepo.RecordRepository[entity.PromotionDecision, string]
	Worker                  domainrepo.RecordRepository[entity.Worker, string]
	WorkerAssignment        domainrepo.RecordRepository[entity.WorkerAssignment, string]
	WorkerLifecycleEvent    domainrepo.RecordRepository[entity.WorkerLifecycleEvent, int64]
	Position                domainrepo.RecordRepository[entity.Position, string]
	TalentAssessment        domainrepo.RecordRepository[entity.TalentAssessment, string]
}

// Service mengoordinasikan operasi domain promotion.
type Service struct {
	dependencies            Dependencies
	promotionRecommendation *operation.Resource[entity.PromotionRecommendation, string, *entity.PromotionRecommendation]
	promotionDecision       *operation.Resource[entity.PromotionDecision, string, *entity.PromotionDecision]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{dependencies: deps}
	var err error
	s.promotionRecommendation, err = operation.NewResource[
		entity.PromotionRecommendation,
		string,
		*entity.PromotionRecommendation,
	](
		deps.PromotionRecommendation,
		deps.Runtime,
		specPromotionRecommendation(),
	)
	if err != nil {
		return nil, err
	}
	s.promotionDecision, err = operation.NewResource[entity.PromotionDecision, string, *entity.PromotionDecision](
		deps.PromotionDecision,
		deps.Runtime,
		specPromotionDecision(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreatePromotionRecommendation menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePromotionRecommendation(ctx context.Context, in CreatePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Create(ctx, func() (*entity.PromotionRecommendation, error) { return createPromotionRecommendation(in) })
}

// UpdatePromotionRecommendation menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePromotionRecommendation(ctx context.Context, key string, in UpdatePromotionRecommendationInput) (*entity.PromotionRecommendation, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Update(ctx, key, func(value *entity.PromotionRecommendation) error { return patchPromotionRecommendation(value, in) })
}

// DeletePromotionRecommendation menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePromotionRecommendation(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Delete(ctx, key)
}

// GetPromotionRecommendationByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPromotionRecommendationByID(ctx context.Context, key string) (*entity.PromotionRecommendation, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.Get(ctx, key)
}

// GetAllPromotionRecommendations menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPromotionRecommendations(ctx context.Context, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, nil, limit, offset)
}

// FindPromotionRecommendationByRecommendationNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByRecommendationNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "recommendation_number", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationBySourceAssignmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationBySourceAssignmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "source_assignment_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByTargetPositionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByTargetPositionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "target_position_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByAssessmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByAssessmentID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "assessment_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationBySupersedesRecommendationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationBySupersedesRecommendationID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "supersedes_recommendation_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByRationaleDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "rationale_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByRecommendedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByRecommendedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "recommended_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByApprovalDecisionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByApprovalDecisionID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "approval_decision_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByAppliedAssignmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByAppliedAssignmentID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "applied_assignment_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByAppliedLifecycleEventID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByAppliedLifecycleEventID(ctx context.Context, value *int64, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "applied_lifecycle_event_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionRecommendationByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionRecommendationByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionRecommendation], error) {
	if s == nil {
		return operation.Page[entity.PromotionRecommendation]{}, hrerrors.ErrUnavailable
	}
	return s.promotionRecommendation.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreatePromotionDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePromotionDecision(ctx context.Context, in CreatePromotionDecisionInput) (*entity.PromotionDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.Create(ctx, func() (*entity.PromotionDecision, error) { return createPromotionDecision(in) })
}

// UpdatePromotionDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePromotionDecision(ctx context.Context, key string, in UpdatePromotionDecisionInput) (*entity.PromotionDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.Update(ctx, key, func(value *entity.PromotionDecision) error { return patchPromotionDecision(value, in) })
}

// DeletePromotionDecision menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePromotionDecision(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.promotionDecision.Delete(ctx, key)
}

// GetPromotionDecisionByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPromotionDecisionByID(ctx context.Context, key string) (*entity.PromotionDecision, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.Get(ctx, key)
}

// GetAllPromotionDecisions menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPromotionDecisions(ctx context.Context, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, nil, limit, offset)
}

// FindPromotionDecisionByRecommendationID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionDecisionByRecommendationID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, []domainrepo.Filter{{Column: "recommendation_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionDecisionByStepNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionDecisionByStepNumber(ctx context.Context, value int32, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, []domainrepo.Filter{{Column: "step_number", Values: []any{value}}}, limit, offset)
}

// FindPromotionDecisionByDecidedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionDecisionByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, []domainrepo.Filter{{Column: "decided_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionDecisionByRationaleDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionDecisionByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, []domainrepo.Filter{{Column: "rationale_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPromotionDecisionByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPromotionDecisionByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PromotionDecision], error) {
	if s == nil {
		return operation.Page[entity.PromotionDecision]{}, hrerrors.ErrUnavailable
	}
	return s.promotionDecision.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}
