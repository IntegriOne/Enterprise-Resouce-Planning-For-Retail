// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateWorkerInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerInput struct {
	WorkerNumber                  string
	PersonPartyReferenceID        string
	WorkerCategory                entity.WorkerCategory
	Status                        *entity.WorkerStatus
	HireDate                      *time.Time
	ServiceStartDate              *time.Time
	TerminationDate               *time.Time
	TerminationReasonCode         *string
	PrimaryLegalEntityReferenceID string
	PrimaryAssignmentID           *string
	IAMPrincipalReferenceID       *string
}

// UpdateWorkerInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerInput struct {
	WorkerNumber                  operation.Field[string]
	PersonPartyReferenceID        operation.Field[string]
	WorkerCategory                operation.Field[entity.WorkerCategory]
	Status                        operation.Field[entity.WorkerStatus]
	HireDate                      operation.Field[time.Time]
	ServiceStartDate              operation.Field[time.Time]
	TerminationDate               operation.Field[time.Time]
	TerminationReasonCode         operation.Field[string]
	PrimaryLegalEntityReferenceID operation.Field[string]
	PrimaryAssignmentID           operation.Field[string]
	IAMPrincipalReferenceID       operation.Field[string]
}

// CreateEmploymentContractInput adalah object input aplikasi, bukan entity ORM.
type CreateEmploymentContractInput struct {
	WorkerID                         string
	ContractNumber                   string
	ContractType                     string
	ContractStatus                   *string
	EmployingLegalEntityReferenceID  string
	EffectiveFrom                    time.Time
	EffectiveTo                      *time.Time
	SignedAt                         *time.Time
	SignedDocumentReferenceID        *string
	SignedDocumentVersionReferenceID *string
	SignedContentDigest              []byte
	ProbationEndDate                 *time.Time
	NoticePeriodDays                 *int32
	WorkCountryCode                  string
	GoverningJurisdictionCode        string
	CollectiveAgreementID            *string
}

// UpdateEmploymentContractInput membedakan field absen, null dan nilai eksplisit.
type UpdateEmploymentContractInput struct {
	WorkerID                         operation.Field[string]
	ContractNumber                   operation.Field[string]
	ContractType                     operation.Field[string]
	ContractStatus                   operation.Field[string]
	EmployingLegalEntityReferenceID  operation.Field[string]
	EffectiveFrom                    operation.Field[time.Time]
	EffectiveTo                      operation.Field[time.Time]
	SignedAt                         operation.Field[time.Time]
	SignedDocumentReferenceID        operation.Field[string]
	SignedDocumentVersionReferenceID operation.Field[string]
	SignedContentDigest              operation.Field[[]byte]
	ProbationEndDate                 operation.Field[time.Time]
	NoticePeriodDays                 operation.Field[int32]
	WorkCountryCode                  operation.Field[string]
	GoverningJurisdictionCode        operation.Field[string]
	CollectiveAgreementID            operation.Field[string]
}

// CreateWorkerAssignmentInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerAssignmentInput struct {
	WorkerID                 string
	PositionID               string
	ManagerWorkerID          *string
	AssignmentType           string
	EmploymentStatusSnapshot *string
	FTEPercent               *entity.Decimal
	ValidFrom                time.Time
	ValidTo                  *time.Time
	AssignmentReason         *string
	ApprovedByPrincipalID    *string
	ApprovedAt               *time.Time
}

// UpdateWorkerAssignmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerAssignmentInput struct {
	WorkerID                 operation.Field[string]
	PositionID               operation.Field[string]
	ManagerWorkerID          operation.Field[string]
	AssignmentType           operation.Field[string]
	EmploymentStatusSnapshot operation.Field[string]
	FTEPercent               operation.Field[entity.Decimal]
	ValidFrom                operation.Field[time.Time]
	ValidTo                  operation.Field[time.Time]
	AssignmentReason         operation.Field[string]
	ApprovedByPrincipalID    operation.Field[string]
	ApprovedAt               operation.Field[time.Time]
}

// CreateWorkerLifecycleEventInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerLifecycleEventInput struct {
	WorkerID        string
	EventType       string
	OccurredAt      time.Time
	EffectiveDate   time.Time
	PreviousStatus  *entity.WorkerStatus
	NewStatus       *entity.WorkerStatus
	SourceReference *string
	EventPayload    entity.JSON
	IdempotencyKey  string
}

// UpdateWorkerLifecycleEventInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerLifecycleEventInput struct {
	WorkerID        operation.Field[string]
	EventType       operation.Field[string]
	OccurredAt      operation.Field[time.Time]
	EffectiveDate   operation.Field[time.Time]
	PreviousStatus  operation.Field[entity.WorkerStatus]
	NewStatus       operation.Field[entity.WorkerStatus]
	SourceReference operation.Field[string]
	EventPayload    operation.Field[entity.JSON]
	IdempotencyKey  operation.Field[string]
}

// CreateWorkerDocumentInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerDocumentInput struct {
	WorkerID                   string
	DocumentType               string
	DocumentReferenceID        string
	DocumentVersionReferenceID string
	ContentDigest              []byte
	DocumentNumberDigest       []byte
	IssuingCountryCode         *string
	IssuedOn                   *time.Time
	ExpiresOn                  *time.Time
	VerificationStatus         *string
	VerifiedByPrincipalID      *string
	VerifiedAt                 *time.Time
	RetentionClass             string
	AccessClass                *string
}

// UpdateWorkerDocumentInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerDocumentInput struct {
	WorkerID                   operation.Field[string]
	DocumentType               operation.Field[string]
	DocumentReferenceID        operation.Field[string]
	DocumentVersionReferenceID operation.Field[string]
	ContentDigest              operation.Field[[]byte]
	DocumentNumberDigest       operation.Field[[]byte]
	IssuingCountryCode         operation.Field[string]
	IssuedOn                   operation.Field[time.Time]
	ExpiresOn                  operation.Field[time.Time]
	VerificationStatus         operation.Field[string]
	VerifiedByPrincipalID      operation.Field[string]
	VerifiedAt                 operation.Field[time.Time]
	RetentionClass             operation.Field[string]
	AccessClass                operation.Field[string]
}

// CreateEmergencyContactInput adalah object input aplikasi, bukan entity ORM.
type CreateEmergencyContactInput struct {
	WorkerID                 string
	ContactPartyReferenceID  string
	RelationshipCode         string
	PriorityOrder            int32
	AuthorizationConfirmedAt *time.Time
	IsActive                 *bool
}

// UpdateEmergencyContactInput membedakan field absen, null dan nilai eksplisit.
type UpdateEmergencyContactInput struct {
	WorkerID                 operation.Field[string]
	ContactPartyReferenceID  operation.Field[string]
	RelationshipCode         operation.Field[string]
	PriorityOrder            operation.Field[int32]
	AuthorizationConfirmedAt operation.Field[time.Time]
	IsActive                 operation.Field[bool]
}
