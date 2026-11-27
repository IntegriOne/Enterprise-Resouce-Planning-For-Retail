// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// WorkerRepository adalah port persistensi hr_workers.
type WorkerRepository interface {
	Repository[entity.Worker, string]
	InsertWorkers(context.Context, *entity.Worker) error
	UpdateWorkers(context.Context, *entity.Worker) error
	DeleteWorkersByUUID(context.Context, string) error
	FindWorkersByUUID(context.Context, string) (*entity.Worker, error)
	FindWorkersByUUIDForUpdate(context.Context, string) (*entity.Worker, error)
	GetAllWorkers(context.Context, int, int) ([]entity.Worker, error)
	CountWorkers(context.Context) (int64, error)
	SelectWorkersWhere(context.Context, []Filter, int, int) ([]entity.Worker, error)
	CountWorkersWhere(context.Context, []Filter) (int64, error)
	FindWorkersByWorkerNumber(context.Context, string, int, int) ([]entity.Worker, error)
	FindWorkersByPersonPartyReferenceID(context.Context, string, int, int) ([]entity.Worker, error)
	FindWorkersByStatus(context.Context, entity.WorkerStatus, int, int) ([]entity.Worker, error)
	FindWorkersByTerminationReasonCode(context.Context, *string, int, int) ([]entity.Worker, error)
	FindWorkersByPrimaryLegalEntityReferenceID(context.Context, string, int, int) ([]entity.Worker, error)
	FindWorkersByPrimaryAssignmentID(context.Context, *string, int, int) ([]entity.Worker, error)
	FindWorkersByIAMPrincipalReferenceID(context.Context, *string, int, int) ([]entity.Worker, error)
	FindWorkersByCreatedByPrincipalID(context.Context, *string, int, int) ([]entity.Worker, error)
	FindByWorkerNumber(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error)
	FindByPersonPartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error)
	FindByStatus(ctx context.Context, value entity.WorkerStatus, limit, offset int) ([]entity.Worker, error)
	FindByTerminationReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error)
	FindByPrimaryLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error)
	FindByPrimaryAssignmentID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error)
	FindByIAMPrincipalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error)
	FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error)
}

// EmploymentContractRepository adalah port persistensi hr_employment_contracts.
type EmploymentContractRepository interface {
	Repository[entity.EmploymentContract, string]
	InsertEmploymentContracts(context.Context, *entity.EmploymentContract) error
	UpdateEmploymentContracts(context.Context, *entity.EmploymentContract) error
	DeleteEmploymentContractsByUUID(context.Context, string) error
	FindEmploymentContractsByUUID(context.Context, string) (*entity.EmploymentContract, error)
	FindEmploymentContractsByUUIDForUpdate(context.Context, string) (*entity.EmploymentContract, error)
	GetAllEmploymentContracts(context.Context, int, int) ([]entity.EmploymentContract, error)
	CountEmploymentContracts(context.Context) (int64, error)
	SelectEmploymentContractsWhere(context.Context, []Filter, int, int) ([]entity.EmploymentContract, error)
	CountEmploymentContractsWhere(context.Context, []Filter) (int64, error)
	FindEmploymentContractsByWorkerID(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByContractNumber(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByContractType(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByContractStatus(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByEmployingLegalEntityReferenceID(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsBySignedDocumentReferenceID(context.Context, *string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsBySignedDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByWorkCountryCode(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByGoverningJurisdictionCode(context.Context, string, int, int) ([]entity.EmploymentContract, error)
	FindEmploymentContractsByCollectiveAgreementID(context.Context, *string, int, int) ([]entity.EmploymentContract, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByContractNumber(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByContractType(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByContractStatus(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindBySignedDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error)
	FindBySignedDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByWorkCountryCode(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByGoverningJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.EmploymentContract, error)
	FindByCollectiveAgreementID(ctx context.Context, value *string, limit, offset int) ([]entity.EmploymentContract, error)
}

// WorkerAssignmentRepository adalah port persistensi hr_worker_assignments.
type WorkerAssignmentRepository interface {
	Repository[entity.WorkerAssignment, string]
	InsertWorkerAssignments(context.Context, *entity.WorkerAssignment) error
	UpdateWorkerAssignments(context.Context, *entity.WorkerAssignment) error
	DeleteWorkerAssignmentsByUUID(context.Context, string) error
	FindWorkerAssignmentsByUUID(context.Context, string) (*entity.WorkerAssignment, error)
	FindWorkerAssignmentsByUUIDForUpdate(context.Context, string) (*entity.WorkerAssignment, error)
	GetAllWorkerAssignments(context.Context, int, int) ([]entity.WorkerAssignment, error)
	CountWorkerAssignments(context.Context) (int64, error)
	SelectWorkerAssignmentsWhere(context.Context, []Filter, int, int) ([]entity.WorkerAssignment, error)
	CountWorkerAssignmentsWhere(context.Context, []Filter) (int64, error)
	FindWorkerAssignmentsByWorkerID(context.Context, string, int, int) ([]entity.WorkerAssignment, error)
	FindWorkerAssignmentsByPositionID(context.Context, string, int, int) ([]entity.WorkerAssignment, error)
	FindWorkerAssignmentsByManagerWorkerID(context.Context, *string, int, int) ([]entity.WorkerAssignment, error)
	FindWorkerAssignmentsByAssignmentType(context.Context, string, int, int) ([]entity.WorkerAssignment, error)
	FindWorkerAssignmentsByRequestedByPrincipalID(context.Context, *string, int, int) ([]entity.WorkerAssignment, error)
	FindWorkerAssignmentsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.WorkerAssignment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error)
	FindByPositionID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error)
	FindByManagerWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error)
	FindByAssignmentType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error)
	FindByRequestedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error)
}

// WorkerLifecycleEventRepository adalah port persistensi hr_worker_lifecycle_events.
type WorkerLifecycleEventRepository interface {
	Repository[entity.WorkerLifecycleEvent, int64]
	InsertWorkerLifecycleEvents(context.Context, *entity.WorkerLifecycleEvent) error
	UpdateWorkerLifecycleEvents(context.Context, *entity.WorkerLifecycleEvent) error
	DeleteWorkerLifecycleEventsByID(context.Context, int64) error
	FindWorkerLifecycleEventsByID(context.Context, int64) (*entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByIDForUpdate(context.Context, int64) (*entity.WorkerLifecycleEvent, error)
	GetAllWorkerLifecycleEvents(context.Context, int, int) ([]entity.WorkerLifecycleEvent, error)
	CountWorkerLifecycleEvents(context.Context) (int64, error)
	SelectWorkerLifecycleEventsWhere(context.Context, []Filter, int, int) ([]entity.WorkerLifecycleEvent, error)
	CountWorkerLifecycleEventsWhere(context.Context, []Filter) (int64, error)
	FindWorkerLifecycleEventsByWorkerID(context.Context, string, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByEventType(context.Context, string, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByPreviousStatus(context.Context, *entity.WorkerStatus, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByNewStatus(context.Context, *entity.WorkerStatus, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByActorPrincipalID(context.Context, *string, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindWorkerLifecycleEventsByIdempotencyKey(context.Context, string, int, int) ([]entity.WorkerLifecycleEvent, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
	FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
	FindByPreviousStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
	FindByNewStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
	FindByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error)
}

// WorkerDocumentRepository adalah port persistensi hr_worker_documents.
type WorkerDocumentRepository interface {
	Repository[entity.WorkerDocument, string]
	InsertWorkerDocuments(context.Context, *entity.WorkerDocument) error
	UpdateWorkerDocuments(context.Context, *entity.WorkerDocument) error
	DeleteWorkerDocumentsByUUID(context.Context, string) error
	FindWorkerDocumentsByUUID(context.Context, string) (*entity.WorkerDocument, error)
	FindWorkerDocumentsByUUIDForUpdate(context.Context, string) (*entity.WorkerDocument, error)
	GetAllWorkerDocuments(context.Context, int, int) ([]entity.WorkerDocument, error)
	CountWorkerDocuments(context.Context) (int64, error)
	SelectWorkerDocumentsWhere(context.Context, []Filter, int, int) ([]entity.WorkerDocument, error)
	CountWorkerDocumentsWhere(context.Context, []Filter) (int64, error)
	FindWorkerDocumentsByWorkerID(context.Context, string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByDocumentType(context.Context, string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByDocumentReferenceID(context.Context, string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByIssuingCountryCode(context.Context, *string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByVerificationStatus(context.Context, string, int, int) ([]entity.WorkerDocument, error)
	FindWorkerDocumentsByVerifiedByPrincipalID(context.Context, *string, int, int) ([]entity.WorkerDocument, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByDocumentType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByDocumentReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByIssuingCountryCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByVerificationStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error)
	FindByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerDocument, error)
}

// EmergencyContactRepository adalah port persistensi hr_emergency_contacts.
type EmergencyContactRepository interface {
	Repository[entity.EmergencyContact, string]
	InsertEmergencyContacts(context.Context, *entity.EmergencyContact) error
	UpdateEmergencyContacts(context.Context, *entity.EmergencyContact) error
	DeleteEmergencyContactsByUUID(context.Context, string) error
	FindEmergencyContactsByUUID(context.Context, string) (*entity.EmergencyContact, error)
	FindEmergencyContactsByUUIDForUpdate(context.Context, string) (*entity.EmergencyContact, error)
	GetAllEmergencyContacts(context.Context, int, int) ([]entity.EmergencyContact, error)
	CountEmergencyContacts(context.Context) (int64, error)
	SelectEmergencyContactsWhere(context.Context, []Filter, int, int) ([]entity.EmergencyContact, error)
	CountEmergencyContactsWhere(context.Context, []Filter) (int64, error)
	FindEmergencyContactsByWorkerID(context.Context, string, int, int) ([]entity.EmergencyContact, error)
	FindEmergencyContactsByContactPartyReferenceID(context.Context, string, int, int) ([]entity.EmergencyContact, error)
	FindEmergencyContactsByRelationshipCode(context.Context, string, int, int) ([]entity.EmergencyContact, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error)
	FindByContactPartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error)
	FindByRelationshipCode(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error)
}

// WorkerAssignmentTransferRepository menyimpan permintaan dan keputusan transfer.
type WorkerAssignmentTransferRepository interface {
	Insert(context.Context, *entity.WorkerAssignmentTransfer) error
	FindByIdempotencyKey(context.Context, string) (*entity.WorkerAssignmentTransfer, error)
	FindByID(context.Context, string) (*entity.WorkerAssignmentTransfer, error)
	ListByWorkerID(context.Context, string, int, int) ([]entity.WorkerAssignmentTransfer, int64, error)
	FindByIDForUpdate(context.Context, string) (*entity.WorkerAssignmentTransfer, error)
	UpdateDecision(context.Context, *entity.WorkerAssignmentTransfer) error
}
