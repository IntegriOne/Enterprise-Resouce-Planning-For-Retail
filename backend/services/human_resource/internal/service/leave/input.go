// Input aplikasi cuti dipelihara langsung dan terpisah dari entity persistensi.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CreateLeavePolicyInput adalah object input aplikasi, bukan entity ORM.
type CreateLeavePolicyInput struct {
	PolicyCode         string
	PolicyVersion      int32
	PolicyName         string
	LeaveTypeCode      string
	JurisdictionCode   string
	AccrualMethod      string
	AnnualEntitlement  *entity.Decimal
	CarryoverLimit     *entity.Decimal
	MaxBalance         *entity.Decimal
	MinimumServiceDays *int32
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	IsPaid             *bool
	IsActive           *bool
}

// UpdateLeavePolicyInput membedakan field absen, null dan nilai eksplisit.
type UpdateLeavePolicyInput struct {
	PolicyCode         operation.Field[string]
	PolicyVersion      operation.Field[int32]
	PolicyName         operation.Field[string]
	LeaveTypeCode      operation.Field[string]
	JurisdictionCode   operation.Field[string]
	AccrualMethod      operation.Field[string]
	AnnualEntitlement  operation.Field[entity.Decimal]
	CarryoverLimit     operation.Field[entity.Decimal]
	MaxBalance         operation.Field[entity.Decimal]
	MinimumServiceDays operation.Field[int32]
	EffectiveFrom      operation.Field[time.Time]
	EffectiveTo        operation.Field[time.Time]
	IsPaid             operation.Field[bool]
	IsActive           operation.Field[bool]
}

// CreateLeaveEntitlementInput adalah object input aplikasi, bukan entity ORM.
type CreateLeaveEntitlementInput struct {
	WorkerID                           string
	LeavePolicyID                      string
	LeaveTypeCodeSnapshot              string
	EntitlementPeriodStart             time.Time
	EntitlementPeriodEnd               time.Time
	OpeningBalance                     *entity.Decimal
	AccruedQuantity                    *entity.Decimal
	UsedQuantitySnapshot               *entity.Decimal
	AdjustedQuantity                   *entity.Decimal
	ClosingBalanceSnapshot             *entity.Decimal
	UOMCode                            *string
	Revision                           *int64
	SourceWorkforceSnapshotReferenceID *string
}

// UpdateLeaveEntitlementInput membedakan field absen, null dan nilai eksplisit.
type UpdateLeaveEntitlementInput struct {
	WorkerID                           operation.Field[string]
	LeavePolicyID                      operation.Field[string]
	LeaveTypeCodeSnapshot              operation.Field[string]
	EntitlementPeriodStart             operation.Field[time.Time]
	EntitlementPeriodEnd               operation.Field[time.Time]
	OpeningBalance                     operation.Field[entity.Decimal]
	AccruedQuantity                    operation.Field[entity.Decimal]
	UsedQuantitySnapshot               operation.Field[entity.Decimal]
	AdjustedQuantity                   operation.Field[entity.Decimal]
	ClosingBalanceSnapshot             operation.Field[entity.Decimal]
	UOMCode                            operation.Field[string]
	Revision                           operation.Field[int64]
	SourceWorkforceSnapshotReferenceID operation.Field[string]
}

// CreateLeaveAdjustmentInput adalah object input aplikasi, bukan entity ORM.
type CreateLeaveAdjustmentInput struct {
	EntitlementID         string
	QuantityDelta         entity.Decimal
	ReasonCode            string
	Reason                string
	ApprovedByPrincipalID *string
	ApprovedAt            *time.Time
	IdempotencyKey        string
}

// UpdateLeaveAdjustmentInput membedakan field absen, null dan nilai eksplisit.
type UpdateLeaveAdjustmentInput struct {
	EntitlementID         operation.Field[string]
	QuantityDelta         operation.Field[entity.Decimal]
	ReasonCode            operation.Field[string]
	Reason                operation.Field[string]
	ApprovedByPrincipalID operation.Field[string]
	ApprovedAt            operation.Field[time.Time]
	IdempotencyKey        operation.Field[string]
}

// CreateLeaveLedgerEntryInput adalah object input aplikasi, bukan entity ORM.
type CreateLeaveLedgerEntryInput struct {
	EntitlementID    string
	EntrySequence    int64
	EntryType        string
	QuantityDelta    entity.Decimal
	EffectiveDate    time.Time
	AdjustmentID     *string
	ReversesEntryID  *string
	SourceManagement string
	SourceEventID    string
	SourceLineKey    string
	SourceDigest     []byte
	ReasonCode       string
}

// UpdateLeaveLedgerEntryInput membedakan field absen, null dan nilai eksplisit.
type UpdateLeaveLedgerEntryInput struct {
	EntitlementID    operation.Field[string]
	EntrySequence    operation.Field[int64]
	EntryType        operation.Field[string]
	QuantityDelta    operation.Field[entity.Decimal]
	EffectiveDate    operation.Field[time.Time]
	AdjustmentID     operation.Field[string]
	ReversesEntryID  operation.Field[string]
	SourceManagement operation.Field[string]
	SourceEventID    operation.Field[string]
	SourceLineKey    operation.Field[string]
	SourceDigest     operation.Field[[]byte]
	ReasonCode       operation.Field[string]
}

// PostLedgerEntryInput adalah permintaan posting, bukan perubahan saldo secara langsung.
// ID, urutan ledger, actor dan waktu pencatatan ditentukan server.
type PostLedgerEntryInput struct {
	EntryType        string
	QuantityDelta    entity.Decimal
	EffectiveDate    time.Time
	AdjustmentID     *string
	ReversesEntryID  *string
	SourceManagement string
	SourceEventID    string
	SourceLineKey    string
	SourceDigest     []byte
	ReasonCode       string
}
