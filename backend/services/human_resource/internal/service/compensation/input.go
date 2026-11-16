// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateCompensationBandInput adalah object input aplikasi, bukan entity ORM.
type CreateCompensationBandInput struct {
	BandCode            string
	BandVersion         int32
	JobLevelCode        string
	JurisdictionCode    string
	CurrencyCode        string
	PayFrequency        string
	MinimumAmount       entity.Decimal
	MidpointAmount      entity.Decimal
	MaximumAmount       entity.Decimal
	ValidFrom           time.Time
	ValidTo             *time.Time
	ApprovalReferenceID *string
}

// UpdateCompensationBandInput membedakan field absen, null dan nilai eksplisit.
type UpdateCompensationBandInput struct {
	BandCode            operation.Field[string]
	BandVersion         operation.Field[int32]
	JobLevelCode        operation.Field[string]
	JurisdictionCode    operation.Field[string]
	CurrencyCode        operation.Field[string]
	PayFrequency        operation.Field[string]
	MinimumAmount       operation.Field[entity.Decimal]
	MidpointAmount      operation.Field[entity.Decimal]
	MaximumAmount       operation.Field[entity.Decimal]
	ValidFrom           operation.Field[time.Time]
	ValidTo             operation.Field[time.Time]
	ApprovalReferenceID operation.Field[string]
}

// CreateCompensationRecordInput adalah object input aplikasi, bukan entity ORM.
type CreateCompensationRecordInput struct {
	WorkerID                string
	AssignmentID            string
	ComponentID             string
	ComponentCodeSnapshot   string
	CompensationBandID      *string
	CompensationType        string
	Amount                  entity.Decimal
	CurrencyCode            string
	PayFrequency            string
	EffectiveFrom           time.Time
	EffectiveTo             *time.Time
	Status                  *entity.CompensationStatus
	ReasonCode              *string
	ApprovedByPrincipalID   *string
	ApprovedAt              *time.Time
	SourceChangeReferenceID *string
	AccessClass             *string
}

// UpdateCompensationRecordInput membedakan field absen, null dan nilai eksplisit.
type UpdateCompensationRecordInput struct {
	WorkerID                operation.Field[string]
	AssignmentID            operation.Field[string]
	ComponentID             operation.Field[string]
	ComponentCodeSnapshot   operation.Field[string]
	CompensationBandID      operation.Field[string]
	CompensationType        operation.Field[string]
	Amount                  operation.Field[entity.Decimal]
	CurrencyCode            operation.Field[string]
	PayFrequency            operation.Field[string]
	EffectiveFrom           operation.Field[time.Time]
	EffectiveTo             operation.Field[time.Time]
	Status                  operation.Field[entity.CompensationStatus]
	ReasonCode              operation.Field[string]
	ApprovedByPrincipalID   operation.Field[string]
	ApprovedAt              operation.Field[time.Time]
	SourceChangeReferenceID operation.Field[string]
	AccessClass             operation.Field[string]
}
