// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateServiceRequestInput adalah object input aplikasi, bukan entity ORM.
type CreateServiceRequestInput struct {
	RequestNumber                               string
	WorkerID                                    string
	RequestType                                 string
	AssignedPrincipalID                         *string
	Status                                      *string
	RestrictedDetailsDocumentVersionReferenceID string
	DueAt                                       *time.Time
	ResolvedAt                                  *time.Time
	ResolutionDocumentVersionReferenceID        *string
	IdempotencyKey                              string
}

// UpdateServiceRequestInput membedakan field absen, null dan nilai eksplisit.
type UpdateServiceRequestInput struct {
	RequestNumber                               operation.Field[string]
	WorkerID                                    operation.Field[string]
	RequestType                                 operation.Field[string]
	AssignedPrincipalID                         operation.Field[string]
	Status                                      operation.Field[string]
	RestrictedDetailsDocumentVersionReferenceID operation.Field[string]
	DueAt                                       operation.Field[time.Time]
	ResolvedAt                                  operation.Field[time.Time]
	ResolutionDocumentVersionReferenceID        operation.Field[string]
	IdempotencyKey                              operation.Field[string]
}

// CreateServiceRequestEventInput adalah object input aplikasi, bukan entity ORM.
type CreateServiceRequestEventInput struct {
	ServiceRequestID                   string
	EventType                          string
	Decision                           *entity.ApprovalDecision
	EvidenceDocumentVersionReferenceID *string
	IdempotencyKey                     string
}

// UpdateServiceRequestEventInput membedakan field absen, null dan nilai eksplisit.
type UpdateServiceRequestEventInput struct {
	ServiceRequestID                   operation.Field[string]
	EventType                          operation.Field[string]
	Decision                           operation.Field[entity.ApprovalDecision]
	EvidenceDocumentVersionReferenceID operation.Field[string]
	IdempotencyKey                     operation.Field[string]
}
