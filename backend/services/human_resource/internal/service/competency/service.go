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
	WorkerQualification        operation.Repository[entity.WorkerQualification, string]
	Competency                 operation.Repository[entity.Competency, string]
	JobCompetency              operation.Repository[entity.JobCompetency, string]
	WorkerCompetencyAssessment operation.Repository[entity.WorkerCompetencyAssessment, string]
}

// Service mengoordinasikan operasi domain competency.
type Service struct {
	workerQualification        *operation.Resource[entity.WorkerQualification, string, *entity.WorkerQualification]
	competency                 *operation.Resource[entity.Competency, string, *entity.Competency]
	jobCompetency              *operation.Resource[entity.JobCompetency, string, *entity.JobCompetency]
	workerCompetencyAssessment *operation.Resource[entity.WorkerCompetencyAssessment, string, *entity.WorkerCompetencyAssessment]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.workerQualification, err = operation.NewResource[entity.WorkerQualification, string, *entity.WorkerQualification](deps.WorkerQualification, deps.Runtime, specWorkerQualification())
	if err != nil {
		return nil, err
	}
	s.competency, err = operation.NewResource[entity.Competency, string, *entity.Competency](deps.Competency, deps.Runtime, specCompetency())
	if err != nil {
		return nil, err
	}
	s.jobCompetency, err = operation.NewResource[entity.JobCompetency, string, *entity.JobCompetency](deps.JobCompetency, deps.Runtime, specJobCompetency())
	if err != nil {
		return nil, err
	}
	s.workerCompetencyAssessment, err = operation.NewResource[entity.WorkerCompetencyAssessment, string, *entity.WorkerCompetencyAssessment](deps.WorkerCompetencyAssessment, deps.Runtime, specWorkerCompetencyAssessment())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateWorkerQualification menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerQualification(ctx context.Context, in CreateWorkerQualificationInput) (*entity.WorkerQualification, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerQualification.Create(ctx, func() (*entity.WorkerQualification, error) { return createWorkerQualification(in) })
}

// UpdateWorkerQualification menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerQualification(ctx context.Context, key string, in UpdateWorkerQualificationInput) (*entity.WorkerQualification, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerQualification.Update(ctx, key, func(value *entity.WorkerQualification) error { return patchWorkerQualification(value, in) })
}

// DeleteWorkerQualification menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerQualification(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerQualification.Delete(ctx, key)
}

// GetWorkerQualificationByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerQualificationByID(ctx context.Context, key string) (*entity.WorkerQualification, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerQualification.Get(ctx, key)
}

// GetAllWorkerQualifications menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerQualifications(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, nil, limit, offset)
}

// FindWorkerQualificationByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByQualificationType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByQualificationType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "qualification_type", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByQualificationCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByQualificationCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "qualification_code", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByIssuerPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByIssuerPartyReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "issuer_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByCredentialReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "credential_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByVerificationStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByVerificationStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "verification_status", Values: []any{value}}}, limit, offset)
}

// FindWorkerQualificationByVerifiedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerQualificationByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerQualification], error) {
	if s == nil {
		return operation.Page[entity.WorkerQualification]{}, hrerrors.ErrUnavailable
	}
	return s.workerQualification.List(ctx, []domainrepo.Filter{{Column: "verified_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCompetency(ctx context.Context, in CreateCompetencyInput) (*entity.Competency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.competency.Create(ctx, func() (*entity.Competency, error) { return createCompetency(in) })
}

// UpdateCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCompetency(ctx context.Context, key string, in UpdateCompetencyInput) (*entity.Competency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.competency.Update(ctx, key, func(value *entity.Competency) error { return patchCompetency(value, in) })
}

// DeleteCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCompetency(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.competency.Delete(ctx, key)
}

// GetCompetencyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCompetencyByID(ctx context.Context, key string) (*entity.Competency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.competency.Get(ctx, key)
}

// GetAllCompetencies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCompetencies(ctx context.Context, limit, offset int) (operation.Page[entity.Competency], error) {
	if s == nil {
		return operation.Page[entity.Competency]{}, hrerrors.ErrUnavailable
	}
	return s.competency.List(ctx, nil, limit, offset)
}

// FindCompetencyByCompetencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompetencyByCompetencyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Competency], error) {
	if s == nil {
		return operation.Page[entity.Competency]{}, hrerrors.ErrUnavailable
	}
	return s.competency.List(ctx, []domainrepo.Filter{{Column: "competency_code", Values: []any{value}}}, limit, offset)
}

// FindCompetencyByCategoryCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCompetencyByCategoryCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Competency], error) {
	if s == nil {
		return operation.Page[entity.Competency]{}, hrerrors.ErrUnavailable
	}
	return s.competency.List(ctx, []domainrepo.Filter{{Column: "category_code", Values: []any{value}}}, limit, offset)
}

// CreateJobCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateJobCompetency(ctx context.Context, in CreateJobCompetencyInput) (*entity.JobCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.Create(ctx, func() (*entity.JobCompetency, error) { return createJobCompetency(in) })
}

// UpdateJobCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateJobCompetency(ctx context.Context, key string, in UpdateJobCompetencyInput) (*entity.JobCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.Update(ctx, key, func(value *entity.JobCompetency) error { return patchJobCompetency(value, in) })
}

// DeleteJobCompetency menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteJobCompetency(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.jobCompetency.Delete(ctx, key)
}

// GetJobCompetencyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetJobCompetencyByID(ctx context.Context, key string) (*entity.JobCompetency, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.Get(ctx, key)
}

// GetAllJobCompetencies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllJobCompetencies(ctx context.Context, limit, offset int) (operation.Page[entity.JobCompetency], error) {
	if s == nil {
		return operation.Page[entity.JobCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.List(ctx, nil, limit, offset)
}

// FindJobCompetencyByJobID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobCompetencyByJobID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobCompetency], error) {
	if s == nil {
		return operation.Page[entity.JobCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.List(ctx, []domainrepo.Filter{{Column: "job_id", Values: []any{value}}}, limit, offset)
}

// FindJobCompetencyByCompetencyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindJobCompetencyByCompetencyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.JobCompetency], error) {
	if s == nil {
		return operation.Page[entity.JobCompetency]{}, hrerrors.ErrUnavailable
	}
	return s.jobCompetency.List(ctx, []domainrepo.Filter{{Column: "competency_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkerCompetencyAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerCompetencyAssessment(ctx context.Context, in CreateWorkerCompetencyAssessmentInput) (*entity.WorkerCompetencyAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.Create(ctx, func() (*entity.WorkerCompetencyAssessment, error) { return createWorkerCompetencyAssessment(in) })
}

// UpdateWorkerCompetencyAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerCompetencyAssessment(ctx context.Context, key string, in UpdateWorkerCompetencyAssessmentInput) (*entity.WorkerCompetencyAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.Update(ctx, key, func(value *entity.WorkerCompetencyAssessment) error {
		return patchWorkerCompetencyAssessment(value, in)
	})
}

// DeleteWorkerCompetencyAssessment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerCompetencyAssessment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.Delete(ctx, key)
}

// GetWorkerCompetencyAssessmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerCompetencyAssessmentByID(ctx context.Context, key string) (*entity.WorkerCompetencyAssessment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.Get(ctx, key)
}

// GetAllWorkerCompetencyAssessments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerCompetencyAssessments(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, nil, limit, offset)
}

// FindWorkerCompetencyAssessmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCompetencyAssessmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCompetencyAssessmentByCompetencyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCompetencyAssessmentByCompetencyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, []domainrepo.Filter{{Column: "competency_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCompetencyAssessmentByAssessedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCompetencyAssessmentByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, []domainrepo.Filter{{Column: "assessed_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerCompetencyAssessmentByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerCompetencyAssessmentByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerCompetencyAssessment], error) {
	if s == nil {
		return operation.Page[entity.WorkerCompetencyAssessment]{}, hrerrors.ErrUnavailable
	}
	return s.workerCompetencyAssessment.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}
