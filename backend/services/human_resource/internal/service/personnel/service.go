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
	PrincipalLinkVerifier PrincipalLinkVerifier
	Runtime               operation.Dependencies
	Worker                domainrepo.RecordRepository[entity.Worker, string]
	EmploymentContract    domainrepo.RecordRepository[entity.EmploymentContract, string]
	WorkerAssignment      domainrepo.RecordRepository[entity.WorkerAssignment, string]
	WorkerLifecycleEvent  domainrepo.RecordRepository[entity.WorkerLifecycleEvent, int64]
	WorkerDocument        domainrepo.RecordRepository[entity.WorkerDocument, string]
	EmergencyContact      domainrepo.RecordRepository[entity.EmergencyContact, string]
	OffboardingPlan       domainrepo.RecordRepository[entity.OffboardingPlan, string]
}

// Service mengoordinasikan operasi domain personnel.
type Service struct {
	transferDependencies Dependencies
	lifecycleEvents      domainrepo.RecordRepository[entity.WorkerLifecycleEvent, int64]
	worker               *operation.Resource[entity.Worker, string, *entity.Worker]
	employmentContract   *operation.Resource[entity.EmploymentContract, string, *entity.EmploymentContract]
	workerAssignment     *operation.Resource[entity.WorkerAssignment, string, *entity.WorkerAssignment]
	workerLifecycleEvent *operation.Resource[entity.WorkerLifecycleEvent, int64, *entity.WorkerLifecycleEvent]
	workerDocument       *operation.Resource[entity.WorkerDocument, string, *entity.WorkerDocument]
	emergencyContact     *operation.Resource[entity.EmergencyContact, string, *entity.EmergencyContact]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	if operation.RequireDependency(deps.Runtime.Clock) != nil {
		deps.Runtime.Clock = operation.SystemClock{}
	}
	if operation.RequireDependency(deps.Runtime.IDs) != nil {
		deps.Runtime.IDs = operation.UUIDGenerator{}
	}
	s := &Service{lifecycleEvents: deps.WorkerLifecycleEvent, transferDependencies: deps}
	var err error
	s.worker, err = operation.NewResource[entity.Worker, string, *entity.Worker](
		deps.Worker,
		deps.Runtime,
		specWorker(),
	)
	if err != nil {
		return nil, err
	}
	s.employmentContract, err = operation.NewResource[entity.EmploymentContract, string, *entity.EmploymentContract](
		deps.EmploymentContract,
		deps.Runtime,
		specEmploymentContract(),
	)
	if err != nil {
		return nil, err
	}
	s.workerAssignment, err = operation.NewResource[entity.WorkerAssignment, string, *entity.WorkerAssignment](
		deps.WorkerAssignment,
		deps.Runtime,
		specWorkerAssignment(),
	)
	if err != nil {
		return nil, err
	}
	s.workerLifecycleEvent, err = operation.NewResource[entity.WorkerLifecycleEvent, int64, *entity.WorkerLifecycleEvent](
		deps.WorkerLifecycleEvent,
		deps.Runtime,
		specWorkerLifecycleEvent(),
	)
	if err != nil {
		return nil, err
	}
	s.workerDocument, err = operation.NewResource[entity.WorkerDocument, string, *entity.WorkerDocument](
		deps.WorkerDocument,
		deps.Runtime,
		specWorkerDocument(),
	)
	if err != nil {
		return nil, err
	}
	s.emergencyContact, err = operation.NewResource[entity.EmergencyContact, string, *entity.EmergencyContact](
		deps.EmergencyContact,
		deps.Runtime,
		specEmergencyContact(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateWorker menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorker(ctx context.Context, in CreateWorkerInput) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.worker.Create(ctx, func() (*entity.Worker, error) { return createWorker(in) })
}

// UpdateWorker menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorker(ctx context.Context, key string, in UpdateWorkerInput) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.worker.Update(ctx, key, func(value *entity.Worker) error { return patchWorker(value, in) })
}

// DeleteWorker menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorker(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.worker.Delete(ctx, key)
}

// GetWorkerByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerByID(ctx context.Context, key string) (*entity.Worker, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.worker.Get(ctx, key)
}

// GetAllWorkers menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkers(ctx context.Context, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, nil, limit, offset)
}

// FindWorkerByWorkerNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByWorkerNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "worker_number", Values: []any{value}}}, limit, offset)
}

// FindWorkerByPersonPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByPersonPartyReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "person_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByStatus(ctx context.Context, value entity.WorkerStatus, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindWorkerByTerminationReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByTerminationReasonCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "termination_reason_code", Values: []any{value}}}, limit, offset)
}

// FindWorkerByPrimaryLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByPrimaryLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "primary_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerByPrimaryAssignmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByPrimaryAssignmentID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "primary_assignment_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerByIAMPrincipalReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByIAMPrincipalReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "iam_principal_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerByCreatedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.Worker], error) {
	if s == nil {
		return operation.Page[entity.Worker]{}, hrerrors.ErrUnavailable
	}
	return s.worker.List(ctx, []domainrepo.Filter{{Column: "created_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateEmploymentContract menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEmploymentContract(ctx context.Context, in CreateEmploymentContractInput) (*entity.EmploymentContract, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employmentContract.Create(ctx, func() (*entity.EmploymentContract, error) { return createEmploymentContract(in) })
}

// UpdateEmploymentContract menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEmploymentContract(ctx context.Context, key string, in UpdateEmploymentContractInput) (*entity.EmploymentContract, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employmentContract.Update(ctx, key, func(value *entity.EmploymentContract) error { return patchEmploymentContract(value, in) })
}

// DeleteEmploymentContract menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEmploymentContract(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.employmentContract.Delete(ctx, key)
}

// GetEmploymentContractByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEmploymentContractByID(ctx context.Context, key string) (*entity.EmploymentContract, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.employmentContract.Get(ctx, key)
}

// GetAllEmploymentContracts menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEmploymentContracts(ctx context.Context, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, nil, limit, offset)
}

// FindEmploymentContractByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByContractNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByContractNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "contract_number", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByContractType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByContractType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "contract_type", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByContractStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByContractStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "contract_status", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByEmployingLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "employing_legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractBySignedDocumentReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractBySignedDocumentReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "signed_document_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractBySignedDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractBySignedDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "signed_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByWorkCountryCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByWorkCountryCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "work_country_code", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByGoverningJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByGoverningJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "governing_jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindEmploymentContractByCollectiveAgreementID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmploymentContractByCollectiveAgreementID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.EmploymentContract], error) {
	if s == nil {
		return operation.Page[entity.EmploymentContract]{}, hrerrors.ErrUnavailable
	}
	return s.employmentContract.List(ctx, []domainrepo.Filter{{Column: "collective_agreement_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkerAssignment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerAssignment(ctx context.Context, in CreateWorkerAssignmentInput) (*entity.WorkerAssignment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.Create(ctx, func() (*entity.WorkerAssignment, error) { return createWorkerAssignment(in) })
}

// UpdateWorkerAssignment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerAssignment(ctx context.Context, key string, in UpdateWorkerAssignmentInput) (*entity.WorkerAssignment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.Update(ctx, key, func(value *entity.WorkerAssignment) error { return patchWorkerAssignment(value, in) })
}

// DeleteWorkerAssignment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerAssignment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerAssignment.Delete(ctx, key)
}

// GetWorkerAssignmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerAssignmentByID(ctx context.Context, key string) (*entity.WorkerAssignment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.Get(ctx, key)
}

// GetAllWorkerAssignments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerAssignments(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, nil, limit, offset)
}

// FindWorkerAssignmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerAssignmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerAssignmentByPositionID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerAssignmentByPositionID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "position_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerAssignmentByManagerWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerAssignmentByManagerWorkerID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "manager_worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerAssignmentByAssignmentType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerAssignmentByAssignmentType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "assignment_type", Values: []any{value}}}, limit, offset)
}

// FindWorkerAssignmentByRequestedByPrincipalID menjalankan pencarian dengan scope server.
func (s *Service) FindWorkerAssignmentByRequestedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "requested_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerAssignmentByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerAssignmentByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerAssignment], error) {
	if s == nil {
		return operation.Page[entity.WorkerAssignment]{}, hrerrors.ErrUnavailable
	}
	return s.workerAssignment.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateWorkerLifecycleEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerLifecycleEvent(ctx context.Context, in CreateWorkerLifecycleEventInput) (*entity.WorkerLifecycleEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.Create(ctx, func() (*entity.WorkerLifecycleEvent, error) { return createWorkerLifecycleEvent(in) })
}

// UpdateWorkerLifecycleEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerLifecycleEvent(ctx context.Context, key int64, in UpdateWorkerLifecycleEventInput) (*entity.WorkerLifecycleEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.Update(ctx, key, func(value *entity.WorkerLifecycleEvent) error { return patchWorkerLifecycleEvent(value, in) })
}

// DeleteWorkerLifecycleEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerLifecycleEvent(ctx context.Context, key int64) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.Delete(ctx, key)
}

// GetWorkerLifecycleEventByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerLifecycleEventByID(ctx context.Context, key int64) (*entity.WorkerLifecycleEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.Get(ctx, key)
}

// GetAllWorkerLifecycleEvents menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerLifecycleEvents(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, nil, limit, offset)
}

// FindWorkerLifecycleEventByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerLifecycleEventByEventType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByEventType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "event_type", Values: []any{value}}}, limit, offset)
}

// FindWorkerLifecycleEventByPreviousStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByPreviousStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "previous_status", Values: []any{value}}}, limit, offset)
}

// FindWorkerLifecycleEventByNewStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByNewStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "new_status", Values: []any{value}}}, limit, offset)
}

// FindWorkerLifecycleEventByActorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByActorPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "actor_principal_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerLifecycleEventByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerLifecycleEventByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerLifecycleEvent], error) {
	if s == nil {
		return operation.Page[entity.WorkerLifecycleEvent]{}, hrerrors.ErrUnavailable
	}
	return s.workerLifecycleEvent.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateWorkerDocument menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateWorkerDocument(ctx context.Context, in CreateWorkerDocumentInput) (*entity.WorkerDocument, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerDocument.Create(ctx, func() (*entity.WorkerDocument, error) { return createWorkerDocument(in) })
}

// UpdateWorkerDocument menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateWorkerDocument(ctx context.Context, key string, in UpdateWorkerDocumentInput) (*entity.WorkerDocument, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerDocument.Update(ctx, key, func(value *entity.WorkerDocument) error { return patchWorkerDocument(value, in) })
}

// DeleteWorkerDocument menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteWorkerDocument(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.workerDocument.Delete(ctx, key)
}

// GetWorkerDocumentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetWorkerDocumentByID(ctx context.Context, key string) (*entity.WorkerDocument, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.workerDocument.Get(ctx, key)
}

// GetAllWorkerDocuments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllWorkerDocuments(ctx context.Context, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, nil, limit, offset)
}

// FindWorkerDocumentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByDocumentType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByDocumentType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "document_type", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByDocumentReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByDocumentReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "document_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByIssuingCountryCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByIssuingCountryCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "issuing_country_code", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByVerificationStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByVerificationStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "verification_status", Values: []any{value}}}, limit, offset)
}

// FindWorkerDocumentByVerifiedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindWorkerDocumentByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.WorkerDocument], error) {
	if s == nil {
		return operation.Page[entity.WorkerDocument]{}, hrerrors.ErrUnavailable
	}
	return s.workerDocument.List(ctx, []domainrepo.Filter{{Column: "verified_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreateEmergencyContact menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateEmergencyContact(ctx context.Context, in CreateEmergencyContactInput) (*entity.EmergencyContact, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.Create(ctx, func() (*entity.EmergencyContact, error) { return createEmergencyContact(in) })
}

// UpdateEmergencyContact menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateEmergencyContact(ctx context.Context, key string, in UpdateEmergencyContactInput) (*entity.EmergencyContact, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.Update(ctx, key, func(value *entity.EmergencyContact) error { return patchEmergencyContact(value, in) })
}

// DeleteEmergencyContact menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteEmergencyContact(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.emergencyContact.Delete(ctx, key)
}

// GetEmergencyContactByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetEmergencyContactByID(ctx context.Context, key string) (*entity.EmergencyContact, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.Get(ctx, key)
}

// GetAllEmergencyContacts menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllEmergencyContacts(ctx context.Context, limit, offset int) (operation.Page[entity.EmergencyContact], error) {
	if s == nil {
		return operation.Page[entity.EmergencyContact]{}, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.List(ctx, nil, limit, offset)
}

// FindEmergencyContactByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmergencyContactByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmergencyContact], error) {
	if s == nil {
		return operation.Page[entity.EmergencyContact]{}, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindEmergencyContactByContactPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmergencyContactByContactPartyReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmergencyContact], error) {
	if s == nil {
		return operation.Page[entity.EmergencyContact]{}, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.List(ctx, []domainrepo.Filter{{Column: "contact_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindEmergencyContactByRelationshipCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindEmergencyContactByRelationshipCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.EmergencyContact], error) {
	if s == nil {
		return operation.Page[entity.EmergencyContact]{}, hrerrors.ErrUnavailable
	}
	return s.emergencyContact.List(ctx, []domainrepo.Filter{{Column: "relationship_code", Values: []any{value}}}, limit, offset)
}
