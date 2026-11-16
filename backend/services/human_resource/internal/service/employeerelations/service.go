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
	Runtime                operation.Dependencies
	EmployeeRelationsCase  operation.Repository[entity.EmployeeRelationsCase, string]
	EmployeeRelationsEvent operation.Repository[entity.EmployeeRelationsEvent, int64]
	CollectiveAgreement    operation.Repository[entity.CollectiveAgreement, string]
}

// Service mengoordinasikan operasi domain employeerelations.
type Service struct {
	employeeRelationsCase  *operation.Resource[entity.EmployeeRelationsCase, string, *entity.EmployeeRelationsCase]
	employeeRelationsEvent *operation.Resource[entity.EmployeeRelationsEvent, int64, *entity.EmployeeRelationsEvent]
	collectiveAgreement    *operation.Resource[entity.CollectiveAgreement, string, *entity.CollectiveAgreement]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.employeeRelationsCase, err = operation.NewResource[entity.EmployeeRelationsCase, string, *entity.EmployeeRelationsCase](deps.EmployeeRelationsCase, deps.Runtime, specEmployeeRelationsCase())
	if err != nil {
		return nil, err
	}
	s.employeeRelationsEvent, err = operation.NewResource[entity.EmployeeRelationsEvent, int64, *entity.EmployeeRelationsEvent](deps.EmployeeRelationsEvent, deps.Runtime, specEmployeeRelationsEvent())
	if err != nil {
		return nil, err
	}
	s.collectiveAgreement, err = operation.NewResource[entity.CollectiveAgreement, string, *entity.CollectiveAgreement](deps.CollectiveAgreement, deps.Runtime, specCollectiveAgreement())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateEmployeeRelationsCase menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEmployeeRelationsCase(ctx context.Context, in CreateEmployeeRelationsCaseInput) (*entity.EmployeeRelationsCase, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.Create(ctx, func() (*entity.EmployeeRelationsCase, error) { return createEmployeeRelationsCase(in) })
}

// UpdateEmployeeRelationsCase menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEmployeeRelationsCase(ctx context.Context, key string, in UpdateEmployeeRelationsCaseInput) (*entity.EmployeeRelationsCase, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.Update(ctx, key, func(value *entity.EmployeeRelationsCase) error { return patchEmployeeRelationsCase(value, in) })
}

// DeleteEmployeeRelationsCase menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEmployeeRelationsCase(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.Delete(ctx, key)
}

// GetEmployeeRelationsCaseByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEmployeeRelationsCaseByID(ctx context.Context, key string) (*entity.EmployeeRelationsCase, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.Get(ctx, key)
}

// GetAllEmployeeRelationsCases menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEmployeeRelationsCases(ctx context.Context, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, nil, limit, offset)
}

// FindEmployeeRelationsCaseByCaseNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByCaseNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "case_number", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByCaseType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByCaseType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "case_type", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByStatus(ctx context.Context, value entity.CaseStatus, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseBySubjectWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseBySubjectWorkerID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "subject_worker_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByReportingPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByReportingPartyReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "reporting_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "assigned_investigator_principal_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByOutcomeCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByOutcomeCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "outcome_code", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByRestrictedSummaryReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByRestrictedSummaryReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "restricted_summary_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByLegalHoldReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByLegalHoldReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "legal_hold_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsCaseByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsCaseByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsCase], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsCase]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsCase.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateEmployeeRelationsEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEmployeeRelationsEvent(ctx context.Context, in CreateEmployeeRelationsEventInput) (*entity.EmployeeRelationsEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.Create(ctx, func() (*entity.EmployeeRelationsEvent, error) { return createEmployeeRelationsEvent(in) })
}

// UpdateEmployeeRelationsEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEmployeeRelationsEvent(ctx context.Context, key int64, in UpdateEmployeeRelationsEventInput) (*entity.EmployeeRelationsEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.Update(ctx, key, func(value *entity.EmployeeRelationsEvent) error { return patchEmployeeRelationsEvent(value, in) })
}

// DeleteEmployeeRelationsEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEmployeeRelationsEvent(ctx context.Context, key int64) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.Delete(ctx, key)
}

// GetEmployeeRelationsEventByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEmployeeRelationsEventByID(ctx context.Context, key int64) (*entity.EmployeeRelationsEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.Get(ctx, key)
}

// GetAllEmployeeRelationsEvents menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEmployeeRelationsEvents(ctx context.Context, limit, offset int) (operation.Page[entity.EmployeeRelationsEvent], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsEvent]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.List(ctx, nil, limit, offset)
}

// FindEmployeeRelationsEventByCaseID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsEventByCaseID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsEvent], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsEvent]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.List(ctx, []domainrepo.Filter{{Column: "case_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsEventByEventType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsEventByEventType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmployeeRelationsEvent], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsEvent]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.List(ctx, []domainrepo.Filter{{Column: "event_type", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsEventByActorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsEventByActorPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsEvent], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsEvent]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.List(ctx, []domainrepo.Filter{{Column: "actor_principal_id", Values: []any{value}}}, limit, offset)
}

// FindEmployeeRelationsEventByEventReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmployeeRelationsEventByEventReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmployeeRelationsEvent], error) {
	if s == nil {
		return operation.Page[entity.EmployeeRelationsEvent]{}, hrerrors.ErrUnavailable
	}
	return s.employeeRelationsEvent.List(ctx, []domainrepo.Filter{{Column: "event_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateCollectiveAgreement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateCollectiveAgreement(ctx context.Context, in CreateCollectiveAgreementInput) (*entity.CollectiveAgreement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.Create(ctx, func() (*entity.CollectiveAgreement, error) { return createCollectiveAgreement(in) })
}

// UpdateCollectiveAgreement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateCollectiveAgreement(ctx context.Context, key string, in UpdateCollectiveAgreementInput) (*entity.CollectiveAgreement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.Update(ctx, key, func(value *entity.CollectiveAgreement) error { return patchCollectiveAgreement(value, in) })
}

// DeleteCollectiveAgreement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteCollectiveAgreement(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.Delete(ctx, key)
}

// GetCollectiveAgreementByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetCollectiveAgreementByID(ctx context.Context, key string) (*entity.CollectiveAgreement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.Get(ctx, key)
}

// GetAllCollectiveAgreements menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllCollectiveAgreements(ctx context.Context, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, nil, limit, offset)
}

// FindCollectiveAgreementByAgreementCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCollectiveAgreementByAgreementCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, []domainrepo.Filter{{Column: "agreement_code", Values: []any{value}}}, limit, offset)
}

// FindCollectiveAgreementByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCollectiveAgreementByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindCollectiveAgreementByRepresentativePartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCollectiveAgreementByRepresentativePartyReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, []domainrepo.Filter{{Column: "representative_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindCollectiveAgreementByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCollectiveAgreementByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindCollectiveAgreementByDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindCollectiveAgreementByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.CollectiveAgreement], error) {
	if s == nil {
		return operation.Page[entity.CollectiveAgreement]{}, hrerrors.ErrUnavailable
	}
	return s.collectiveAgreement.List(ctx, []domainrepo.Filter{{Column: "document_version_reference_id", Values: []any{value}}}, limit, offset)
}
