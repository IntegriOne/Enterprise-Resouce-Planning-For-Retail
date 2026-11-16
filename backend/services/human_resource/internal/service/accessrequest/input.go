// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateAccessRequestInput adalah object input aplikasi, bukan entity ORM.
type CreateAccessRequestInput struct {
	WorkerID                    string
	LifecycleEventID            int64
	RequestKind                 string
	RequestedScopeType          string
	RequestedScopeReferenceID   string
	ApprovedCatalogReferenceID  *string
	IAMRequestReferenceID       *string
	IAMDecisionEventReferenceID *string
	Status                      *string
	LastReconciledAt            *time.Time
	IdempotencyKey              string
}

// UpdateAccessRequestInput membedakan field absen, null dan nilai eksplisit.
type UpdateAccessRequestInput struct {
	WorkerID                    operation.Field[string]
	LifecycleEventID            operation.Field[int64]
	RequestKind                 operation.Field[string]
	RequestedScopeType          operation.Field[string]
	RequestedScopeReferenceID   operation.Field[string]
	ApprovedCatalogReferenceID  operation.Field[string]
	IAMRequestReferenceID       operation.Field[string]
	IAMDecisionEventReferenceID operation.Field[string]
	Status                      operation.Field[string]
	LastReconciledAt            operation.Field[time.Time]
	IdempotencyKey              operation.Field[string]
}
