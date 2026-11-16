// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreatePolicyInput adalah object input aplikasi, bukan entity ORM.
type CreatePolicyInput struct {
	PolicyCode                 string
	PolicyVersion              int32
	PolicyName                 string
	CategoryCode               string
	LegalEntityReferenceID     string
	EffectiveFrom              time.Time
	EffectiveTo                *time.Time
	DocumentVersionReferenceID string
	ContentDigest              []byte
	ApprovedByPrincipalID      string
	ApprovedAt                 time.Time
}

// UpdatePolicyInput membedakan field absen, null dan nilai eksplisit.
type UpdatePolicyInput struct {
	PolicyCode                 operation.Field[string]
	PolicyVersion              operation.Field[int32]
	PolicyName                 operation.Field[string]
	CategoryCode               operation.Field[string]
	LegalEntityReferenceID     operation.Field[string]
	EffectiveFrom              operation.Field[time.Time]
	EffectiveTo                operation.Field[time.Time]
	DocumentVersionReferenceID operation.Field[string]
	ContentDigest              operation.Field[[]byte]
	ApprovedByPrincipalID      operation.Field[string]
	ApprovedAt                 operation.Field[time.Time]
}

// CreatePolicyAcknowledgementInput adalah object input aplikasi, bukan entity ORM.
type CreatePolicyAcknowledgementInput struct {
	PolicyID                           string
	WorkerID                           string
	DueAt                              *time.Time
	AcknowledgedAt                     *time.Time
	AcknowledgedByPrincipalID          *string
	EvidenceDocumentVersionReferenceID *string
}

// UpdatePolicyAcknowledgementInput membedakan field absen, null dan nilai eksplisit.
type UpdatePolicyAcknowledgementInput struct {
	PolicyID                           operation.Field[string]
	WorkerID                           operation.Field[string]
	DueAt                              operation.Field[time.Time]
	AcknowledgedAt                     operation.Field[time.Time]
	AcknowledgedByPrincipalID          operation.Field[string]
	EvidenceDocumentVersionReferenceID operation.Field[string]
}
