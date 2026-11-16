// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateEmployeeRelationsCaseInput adalah object input aplikasi, bukan entity ORM.
type CreateEmployeeRelationsCaseInput struct {
	CaseNumber                      string
	CaseType                        string
	Status                          *entity.CaseStatus
	SubjectWorkerID                 *string
	ReportingPartyReferenceID       *string
	AssignedInvestigatorPrincipalID *string
	DueAt                           *time.Time
	ResolvedAt                      *time.Time
	OutcomeCode                     *string
	RestrictedSummaryReferenceID    string
	ConfidentialityClass            *string
	LegalHoldReferenceID            *string
	IdempotencyKey                  string
}

// UpdateEmployeeRelationsCaseInput membedakan field absen, null dan nilai eksplisit.
type UpdateEmployeeRelationsCaseInput struct {
	CaseNumber                      operation.Field[string]
	CaseType                        operation.Field[string]
	Status                          operation.Field[entity.CaseStatus]
	SubjectWorkerID                 operation.Field[string]
	ReportingPartyReferenceID       operation.Field[string]
	AssignedInvestigatorPrincipalID operation.Field[string]
	DueAt                           operation.Field[time.Time]
	ResolvedAt                      operation.Field[time.Time]
	OutcomeCode                     operation.Field[string]
	RestrictedSummaryReferenceID    operation.Field[string]
	ConfidentialityClass            operation.Field[string]
	LegalHoldReferenceID            operation.Field[string]
	IdempotencyKey                  operation.Field[string]
}

// CreateEmployeeRelationsEventInput adalah object input aplikasi, bukan entity ORM.
type CreateEmployeeRelationsEventInput struct {
	CaseID           string
	EventType        string
	EventDigest      []byte
	EventReferenceID *string
}

// UpdateEmployeeRelationsEventInput membedakan field absen, null dan nilai eksplisit.
type UpdateEmployeeRelationsEventInput struct {
	CaseID           operation.Field[string]
	EventType        operation.Field[string]
	EventDigest      operation.Field[[]byte]
	EventReferenceID operation.Field[string]
}

// CreateCollectiveAgreementInput adalah object input aplikasi, bukan entity ORM.
type CreateCollectiveAgreementInput struct {
	AgreementCode                  string
	AgreementVersion               int32
	LegalEntityReferenceID         string
	RepresentativePartyReferenceID string
	JurisdictionCode               string
	EffectiveFrom                  time.Time
	EffectiveTo                    *time.Time
	SignedAt                       time.Time
	DocumentVersionReferenceID     string
	ContentDigest                  []byte
}

// UpdateCollectiveAgreementInput membedakan field absen, null dan nilai eksplisit.
type UpdateCollectiveAgreementInput struct {
	AgreementCode                  operation.Field[string]
	AgreementVersion               operation.Field[int32]
	LegalEntityReferenceID         operation.Field[string]
	RepresentativePartyReferenceID operation.Field[string]
	JurisdictionCode               operation.Field[string]
	EffectiveFrom                  operation.Field[time.Time]
	EffectiveTo                    operation.Field[time.Time]
	SignedAt                       operation.Field[time.Time]
	DocumentVersionReferenceID     operation.Field[string]
	ContentDigest                  operation.Field[[]byte]
}
