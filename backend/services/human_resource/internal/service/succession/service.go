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

type approvalWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type Dependencies struct {
	ApprovalWorkers     approvalWorkerReader
	ApprovalVerifier    ApprovalVerifier
	Runtime             operation.Dependencies
	SuccessionPlan      domainrepo.RecordRepository[entity.SuccessionPlan, string]
	SuccessionCandidate domainrepo.RecordRepository[entity.SuccessionCandidate, string]
}

// Service mengoordinasikan operasi domain succession.
type Service struct {
	approvalWorkers     approvalWorkerReader
	approvalVerifier    ApprovalVerifier
	candidateRecords    domainrepo.RecordRepository[entity.SuccessionCandidate, string]
	successionPlan      *operation.Resource[entity.SuccessionPlan, string, *entity.SuccessionPlan]
	successionCandidate *operation.Resource[entity.SuccessionCandidate, string, *entity.SuccessionCandidate]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{
		approvalVerifier: deps.ApprovalVerifier, approvalWorkers: deps.ApprovalWorkers,
		candidateRecords: deps.SuccessionCandidate,
	}
	var err error
	s.successionPlan, err = operation.NewResource[entity.SuccessionPlan, string, *entity.SuccessionPlan](
		deps.SuccessionPlan,
		deps.Runtime,
		specSuccessionPlan(),
	)
	if err != nil {
		return nil, err
	}
	s.successionCandidate, err = operation.NewResource[entity.SuccessionCandidate, string, *entity.SuccessionCandidate](
		deps.SuccessionCandidate,
		deps.Runtime,
		specSuccessionCandidate(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateSuccessionPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateSuccessionPlan(ctx context.Context, in CreateSuccessionPlanInput) (*entity.SuccessionPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionPlan.Create(ctx, func() (*entity.SuccessionPlan, error) { return createSuccessionPlan(in) })
}

// UpdateSuccessionPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateSuccessionPlan(ctx context.Context, key string, in UpdateSuccessionPlanInput) (*entity.SuccessionPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionPlan.Update(ctx, key, func(value *entity.SuccessionPlan) error { return patchSuccessionPlan(value, in) })
}

// DeleteSuccessionPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteSuccessionPlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.successionPlan.Delete(ctx, key)
}

// GetSuccessionPlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetSuccessionPlanByID(ctx context.Context, key string) (*entity.SuccessionPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionPlan.Get(ctx, key)
}

// GetAllSuccessionPlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllSuccessionPlans(ctx context.Context, limit, offset int) (operation.Page[entity.SuccessionPlan], error) {
	if s == nil {
		return operation.Page[entity.SuccessionPlan]{}, hrerrors.ErrUnavailable
	}
	return s.successionPlan.List(ctx, nil, limit, offset)
}

// FindSuccessionPlanByPositionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionPlanByPositionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.SuccessionPlan], error) {
	if s == nil {
		return operation.Page[entity.SuccessionPlan]{}, hrerrors.ErrUnavailable
	}
	return s.successionPlan.List(ctx, []domainrepo.Filter{{Column: "position_id", Values: []any{value}}}, limit, offset)
}

// FindSuccessionPlanByPlanCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionPlanByPlanCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.SuccessionPlan], error) {
	if s == nil {
		return operation.Page[entity.SuccessionPlan]{}, hrerrors.ErrUnavailable
	}
	return s.successionPlan.List(ctx, []domainrepo.Filter{{Column: "plan_code", Values: []any{value}}}, limit, offset)
}

// FindSuccessionPlanByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionPlanByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.SuccessionPlan], error) {
	if s == nil {
		return operation.Page[entity.SuccessionPlan]{}, hrerrors.ErrUnavailable
	}
	return s.successionPlan.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateSuccessionCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateSuccessionCandidate(ctx context.Context, in CreateSuccessionCandidateInput) (*entity.SuccessionCandidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.Create(ctx, func() (*entity.SuccessionCandidate, error) { return createSuccessionCandidate(in) })
}

// UpdateSuccessionCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateSuccessionCandidate(ctx context.Context, key string, in UpdateSuccessionCandidateInput) (*entity.SuccessionCandidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.Update(ctx, key, func(value *entity.SuccessionCandidate) error { return patchSuccessionCandidate(value, in) })
}

// DeleteSuccessionCandidate menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteSuccessionCandidate(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.successionCandidate.Delete(ctx, key)
}

// GetSuccessionCandidateByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetSuccessionCandidateByID(ctx context.Context, key string) (*entity.SuccessionCandidate, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.Get(ctx, key)
}

// GetAllSuccessionCandidates menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllSuccessionCandidates(ctx context.Context, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, nil, limit, offset)
}

// FindSuccessionCandidateBySuccessionPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionCandidateBySuccessionPlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, []domainrepo.Filter{{Column: "succession_plan_id", Values: []any{value}}}, limit, offset)
}

// FindSuccessionCandidateByCandidateWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionCandidateByCandidateWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, []domainrepo.Filter{{Column: "candidate_worker_id", Values: []any{value}}}, limit, offset)
}

// FindSuccessionCandidateByReadinessCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionCandidateByReadinessCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, []domainrepo.Filter{{Column: "readiness_code", Values: []any{value}}}, limit, offset)
}

// FindSuccessionCandidateByTalentAssessmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionCandidateByTalentAssessmentID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, []domainrepo.Filter{{Column: "talent_assessment_id", Values: []any{value}}}, limit, offset)
}

// FindSuccessionCandidateByDevelopmentPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindSuccessionCandidateByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
	if s == nil {
		return operation.Page[entity.SuccessionCandidate]{}, hrerrors.ErrUnavailable
	}
	return s.successionCandidate.List(ctx, []domainrepo.Filter{{Column: "development_plan_id", Values: []any{value}}}, limit, offset)
}
