// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateWorkerQualificationInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerQualificationInput struct {
	WorkerID               string
	QualificationType      string
	QualificationCode      *string
	Title                  string
	IssuerPartyReferenceID *string
	CredentialReferenceID  *string
	ObtainedOn             *time.Time
	ExpiresOn              *time.Time
	VerificationStatus     *string
	VerifiedByPrincipalID  *string
	VerifiedAt             *time.Time
}

// UpdateWorkerQualificationInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerQualificationInput struct {
	WorkerID               operation.Field[string]
	QualificationType      operation.Field[string]
	QualificationCode      operation.Field[string]
	Title                  operation.Field[string]
	IssuerPartyReferenceID operation.Field[string]
	CredentialReferenceID  operation.Field[string]
	ObtainedOn             operation.Field[time.Time]
	ExpiresOn              operation.Field[time.Time]
	VerificationStatus     operation.Field[string]
	VerifiedByPrincipalID  operation.Field[string]
	VerifiedAt             operation.Field[time.Time]
}

// CreateCompetencyInput adalah object input aplikasi, bukan entity ORM.
type CreateCompetencyInput struct {
	CompetencyCode    string
	CompetencyVersion int32
	CompetencyName    string
	CategoryCode      string
	ProficiencyScale  entity.JSON
	IsActive          *bool
}

// UpdateCompetencyInput membedakan field absen, null dan nilai eksplisit.
type UpdateCompetencyInput struct {
	CompetencyCode    operation.Field[string]
	CompetencyVersion operation.Field[int32]
	CompetencyName    operation.Field[string]
	CategoryCode      operation.Field[string]
	ProficiencyScale  operation.Field[entity.JSON]
	IsActive          operation.Field[bool]
}

// CreateJobCompetencyInput adalah object input aplikasi, bukan entity ORM.
type CreateJobCompetencyInput struct {
	JobID         string
	CompetencyID  string
	RequiredLevel int16
	IsMandatory   *bool
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

// UpdateJobCompetencyInput membedakan field absen, null dan nilai eksplisit.
type UpdateJobCompetencyInput struct {
	JobID         operation.Field[string]
	CompetencyID  operation.Field[string]
	RequiredLevel operation.Field[int16]
	IsMandatory   operation.Field[bool]
	EffectiveFrom operation.Field[time.Time]
	EffectiveTo   operation.Field[time.Time]
}

// CreateWorkerCompetencyAssessmentInput adalah object input aplikasi, bukan entity ORM.
type CreateWorkerCompetencyAssessmentInput struct {
	WorkerID                           string
	CompetencyID                       string
	AssessedLevel                      int16
	AssessmentMethod                   string
	AssessedAt                         time.Time
	ExpiresOn                          *time.Time
	EvidenceDocumentVersionReferenceID *string
	EvidenceDigest                     []byte
	IdempotencyKey                     string
}

// UpdateWorkerCompetencyAssessmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateWorkerCompetencyAssessmentInput struct {
	WorkerID                           operation.Field[string]
	CompetencyID                       operation.Field[string]
	AssessedLevel                      operation.Field[int16]
	AssessmentMethod                   operation.Field[string]
	AssessedAt                         operation.Field[time.Time]
	ExpiresOn                          operation.Field[time.Time]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	EvidenceDigest                     operation.Field[[]byte]
	IdempotencyKey                     operation.Field[string]
}
