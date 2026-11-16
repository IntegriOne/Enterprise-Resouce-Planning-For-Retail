// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createLeavePolicy(in CreateLeavePolicyInput) (*entity.LeavePolicy, error) {
	v := &entity.LeavePolicy{}
	v.PolicyCode = in.PolicyCode
	v.PolicyVersion = in.PolicyVersion
	v.PolicyName = in.PolicyName
	v.LeaveTypeCode = in.LeaveTypeCode
	v.JurisdictionCode = in.JurisdictionCode
	v.AccrualMethod = in.AccrualMethod
	v.AnnualEntitlement = in.AnnualEntitlement
	v.CarryoverLimit = in.CarryoverLimit
	v.MaxBalance = in.MaxBalance
	v.MinimumServiceDays = in.MinimumServiceDays
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	if in.IsPaid != nil {
		v.IsPaid = *in.IsPaid
	} else {
		v.IsPaid = true
	}
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchLeavePolicy(v *entity.LeavePolicy, in UpdateLeavePolicyInput) error {
	if in.PolicyCode.Set {
		if in.PolicyCode.Value == nil {
			return hrerrors.Invalid("policy_code", "must not be null")
		}
		v.PolicyCode = *in.PolicyCode.Value
	}
	if in.PolicyVersion.Set {
		if in.PolicyVersion.Value == nil {
			return hrerrors.Invalid("policy_version", "must not be null")
		}
		v.PolicyVersion = *in.PolicyVersion.Value
	}
	if in.PolicyName.Set {
		if in.PolicyName.Value == nil {
			return hrerrors.Invalid("policy_name", "must not be null")
		}
		v.PolicyName = *in.PolicyName.Value
	}
	if in.LeaveTypeCode.Set {
		if in.LeaveTypeCode.Value == nil {
			return hrerrors.Invalid("leave_type_code", "must not be null")
		}
		v.LeaveTypeCode = *in.LeaveTypeCode.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
	}
	if in.AccrualMethod.Set {
		if in.AccrualMethod.Value == nil {
			return hrerrors.Invalid("accrual_method", "must not be null")
		}
		v.AccrualMethod = *in.AccrualMethod.Value
	}
	if in.AnnualEntitlement.Set {
		v.AnnualEntitlement = in.AnnualEntitlement.Value
	}
	if in.CarryoverLimit.Set {
		v.CarryoverLimit = in.CarryoverLimit.Value
	}
	if in.MaxBalance.Set {
		v.MaxBalance = in.MaxBalance.Value
	}
	if in.MinimumServiceDays.Set {
		v.MinimumServiceDays = in.MinimumServiceDays.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	if in.IsPaid.Set {
		if in.IsPaid.Value == nil {
			return hrerrors.Invalid("is_paid", "must not be null")
		}
		v.IsPaid = *in.IsPaid.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specLeavePolicy() operation.Specification[entity.LeavePolicy, string] {
	return operation.Specification[entity.LeavePolicy, string]{
		Domain: "leave", Table: "hr_leave_policies",
		Key:         func(v *entity.LeavePolicy) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LeavePolicy) string { return v.ID },
		Initialize: func(v *entity.LeavePolicy, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.LeavePolicy, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LeavePolicy) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			return a
		},
		Guard: func(action string, old, next *entity.LeavePolicy) error {
			return nil
		},
	}
}
func createLeaveEntitlement(in CreateLeaveEntitlementInput) (*entity.LeaveEntitlement, error) {
	v := &entity.LeaveEntitlement{}
	v.WorkerID = in.WorkerID
	v.LeavePolicyID = in.LeavePolicyID
	v.LeaveTypeCodeSnapshot = in.LeaveTypeCodeSnapshot
	v.EntitlementPeriodStart = in.EntitlementPeriodStart
	v.EntitlementPeriodEnd = in.EntitlementPeriodEnd
	if in.OpeningBalance != nil {
		v.OpeningBalance = *in.OpeningBalance
	} else {
		v.OpeningBalance = entity.Decimal("0")
	}
	if in.AccruedQuantity != nil {
		v.AccruedQuantity = *in.AccruedQuantity
	} else {
		v.AccruedQuantity = entity.Decimal("0")
	}
	if in.UsedQuantitySnapshot != nil {
		v.UsedQuantitySnapshot = *in.UsedQuantitySnapshot
	} else {
		v.UsedQuantitySnapshot = entity.Decimal("0")
	}
	if in.AdjustedQuantity != nil {
		v.AdjustedQuantity = *in.AdjustedQuantity
	} else {
		v.AdjustedQuantity = entity.Decimal("0")
	}
	if in.ClosingBalanceSnapshot != nil {
		v.ClosingBalanceSnapshot = *in.ClosingBalanceSnapshot
	} else {
		v.ClosingBalanceSnapshot = entity.Decimal("0")
	}
	if in.UOMCode != nil {
		v.UOMCode = *in.UOMCode
	} else {
		v.UOMCode = "day"
	}
	if in.Revision != nil {
		v.Revision = *in.Revision
	} else {
		v.Revision = 1
	}
	v.SourceWorkforceSnapshotReferenceID = in.SourceWorkforceSnapshotReferenceID
	return v, nil
}
func patchLeaveEntitlement(v *entity.LeaveEntitlement, in UpdateLeaveEntitlementInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.LeavePolicyID.Set {
		if in.LeavePolicyID.Value == nil {
			return hrerrors.Invalid("leave_policy_id", "must not be null")
		}
		v.LeavePolicyID = *in.LeavePolicyID.Value
	}
	if in.LeaveTypeCodeSnapshot.Set {
		if in.LeaveTypeCodeSnapshot.Value == nil {
			return hrerrors.Invalid("leave_type_code_snapshot", "must not be null")
		}
		v.LeaveTypeCodeSnapshot = *in.LeaveTypeCodeSnapshot.Value
	}
	if in.EntitlementPeriodStart.Set {
		if in.EntitlementPeriodStart.Value == nil {
			return hrerrors.Invalid("entitlement_period_start", "must not be null")
		}
		v.EntitlementPeriodStart = *in.EntitlementPeriodStart.Value
	}
	if in.EntitlementPeriodEnd.Set {
		if in.EntitlementPeriodEnd.Value == nil {
			return hrerrors.Invalid("entitlement_period_end", "must not be null")
		}
		v.EntitlementPeriodEnd = *in.EntitlementPeriodEnd.Value
	}
	if in.OpeningBalance.Set {
		if in.OpeningBalance.Value == nil {
			return hrerrors.Invalid("opening_balance", "must not be null")
		}
		v.OpeningBalance = *in.OpeningBalance.Value
	}
	if in.AccruedQuantity.Set {
		if in.AccruedQuantity.Value == nil {
			return hrerrors.Invalid("accrued_quantity", "must not be null")
		}
		v.AccruedQuantity = *in.AccruedQuantity.Value
	}
	if in.UsedQuantitySnapshot.Set {
		if in.UsedQuantitySnapshot.Value == nil {
			return hrerrors.Invalid("used_quantity_snapshot", "must not be null")
		}
		v.UsedQuantitySnapshot = *in.UsedQuantitySnapshot.Value
	}
	if in.AdjustedQuantity.Set {
		if in.AdjustedQuantity.Value == nil {
			return hrerrors.Invalid("adjusted_quantity", "must not be null")
		}
		v.AdjustedQuantity = *in.AdjustedQuantity.Value
	}
	if in.ClosingBalanceSnapshot.Set {
		if in.ClosingBalanceSnapshot.Value == nil {
			return hrerrors.Invalid("closing_balance_snapshot", "must not be null")
		}
		v.ClosingBalanceSnapshot = *in.ClosingBalanceSnapshot.Value
	}
	if in.UOMCode.Set {
		if in.UOMCode.Value == nil {
			return hrerrors.Invalid("uom_code", "must not be null")
		}
		v.UOMCode = *in.UOMCode.Value
	}
	if in.Revision.Set {
		if in.Revision.Value == nil {
			return hrerrors.Invalid("revision", "must not be null")
		}
		v.Revision = *in.Revision.Value
	}
	if in.SourceWorkforceSnapshotReferenceID.Set {
		v.SourceWorkforceSnapshotReferenceID = in.SourceWorkforceSnapshotReferenceID.Value
	}
	return nil
}
func specLeaveEntitlement() operation.Specification[entity.LeaveEntitlement, string] {
	return operation.Specification[entity.LeaveEntitlement, string]{
		Domain: "leave", Table: "hr_leave_entitlements",
		Key:         func(v *entity.LeaveEntitlement) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LeaveEntitlement) string { return v.ID },
		Initialize: func(v *entity.LeaveEntitlement, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CalculatedAt = now
		},
		Touch: func(v *entity.LeaveEntitlement, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LeaveEntitlement) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["leave_policy_id"] = fmt.Sprint(v.LeavePolicyID)
			if v.SourceWorkforceSnapshotReferenceID != nil {
				a["source_workforce_snapshot_reference_id"] = fmt.Sprint(*v.SourceWorkforceSnapshotReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.LeaveEntitlement) error {
			return nil
		},
	}
}
func createLeaveAdjustment(in CreateLeaveAdjustmentInput) (*entity.LeaveAdjustment, error) {
	v := &entity.LeaveAdjustment{}
	v.EntitlementID = in.EntitlementID
	v.QuantityDelta = in.QuantityDelta
	v.ReasonCode = in.ReasonCode
	v.Reason = in.Reason
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchLeaveAdjustment(v *entity.LeaveAdjustment, in UpdateLeaveAdjustmentInput) error {
	if in.EntitlementID.Set {
		if in.EntitlementID.Value == nil {
			return hrerrors.Invalid("entitlement_id", "must not be null")
		}
		v.EntitlementID = *in.EntitlementID.Value
	}
	if in.QuantityDelta.Set {
		if in.QuantityDelta.Value == nil {
			return hrerrors.Invalid("quantity_delta", "must not be null")
		}
		v.QuantityDelta = *in.QuantityDelta.Value
	}
	if in.ReasonCode.Set {
		if in.ReasonCode.Value == nil {
			return hrerrors.Invalid("reason_code", "must not be null")
		}
		v.ReasonCode = *in.ReasonCode.Value
	}
	if in.Reason.Set {
		if in.Reason.Value == nil {
			return hrerrors.Invalid("reason", "must not be null")
		}
		v.Reason = *in.Reason.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specLeaveAdjustment() operation.Specification[entity.LeaveAdjustment, string] {
	return operation.Specification[entity.LeaveAdjustment, string]{
		Domain: "leave", Table: "hr_leave_adjustments",
		Key:         func(v *entity.LeaveAdjustment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LeaveAdjustment) string { return v.ID },
		Initialize: func(v *entity.LeaveAdjustment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RequestedByPrincipalID = actor.PrincipalID
			v.CreatedAt = now
		},
		Touch: func(v *entity.LeaveAdjustment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LeaveAdjustment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["entitlement_id"] = fmt.Sprint(v.EntitlementID)
			a["requested_by_principal_id"] = fmt.Sprint(v.RequestedByPrincipalID)
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.LeaveAdjustment) error {
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createLeaveLedgerEntry(in CreateLeaveLedgerEntryInput) (*entity.LeaveLedgerEntry, error) {
	v := &entity.LeaveLedgerEntry{}
	v.EntitlementID = in.EntitlementID
	v.EntrySequence = in.EntrySequence
	v.EntryType = in.EntryType
	v.QuantityDelta = in.QuantityDelta
	v.EffectiveDate = in.EffectiveDate
	v.AdjustmentID = in.AdjustmentID
	v.ReversesEntryID = in.ReversesEntryID
	v.SourceManagement = in.SourceManagement
	v.SourceEventID = in.SourceEventID
	v.SourceLineKey = in.SourceLineKey
	v.SourceDigest = in.SourceDigest
	v.ReasonCode = in.ReasonCode
	return v, nil
}
func patchLeaveLedgerEntry(v *entity.LeaveLedgerEntry, in UpdateLeaveLedgerEntryInput) error {
	if in.EntitlementID.Set {
		if in.EntitlementID.Value == nil {
			return hrerrors.Invalid("entitlement_id", "must not be null")
		}
		v.EntitlementID = *in.EntitlementID.Value
	}
	if in.EntrySequence.Set {
		if in.EntrySequence.Value == nil {
			return hrerrors.Invalid("entry_sequence", "must not be null")
		}
		v.EntrySequence = *in.EntrySequence.Value
	}
	if in.EntryType.Set {
		if in.EntryType.Value == nil {
			return hrerrors.Invalid("entry_type", "must not be null")
		}
		v.EntryType = *in.EntryType.Value
	}
	if in.QuantityDelta.Set {
		if in.QuantityDelta.Value == nil {
			return hrerrors.Invalid("quantity_delta", "must not be null")
		}
		v.QuantityDelta = *in.QuantityDelta.Value
	}
	if in.EffectiveDate.Set {
		if in.EffectiveDate.Value == nil {
			return hrerrors.Invalid("effective_date", "must not be null")
		}
		v.EffectiveDate = *in.EffectiveDate.Value
	}
	if in.AdjustmentID.Set {
		v.AdjustmentID = in.AdjustmentID.Value
	}
	if in.ReversesEntryID.Set {
		v.ReversesEntryID = in.ReversesEntryID.Value
	}
	if in.SourceManagement.Set {
		if in.SourceManagement.Value == nil {
			return hrerrors.Invalid("source_management", "must not be null")
		}
		v.SourceManagement = *in.SourceManagement.Value
	}
	if in.SourceEventID.Set {
		if in.SourceEventID.Value == nil {
			return hrerrors.Invalid("source_event_id", "must not be null")
		}
		v.SourceEventID = *in.SourceEventID.Value
	}
	if in.SourceLineKey.Set {
		if in.SourceLineKey.Value == nil {
			return hrerrors.Invalid("source_line_key", "must not be null")
		}
		v.SourceLineKey = *in.SourceLineKey.Value
	}
	if in.SourceDigest.Set {
		if in.SourceDigest.Value == nil {
			v.SourceDigest = nil
		} else {
			v.SourceDigest = *in.SourceDigest.Value
		}
	}
	if in.ReasonCode.Set {
		if in.ReasonCode.Value == nil {
			return hrerrors.Invalid("reason_code", "must not be null")
		}
		v.ReasonCode = *in.ReasonCode.Value
	}
	return nil
}
func specLeaveLedgerEntry() operation.Specification[entity.LeaveLedgerEntry, string] {
	return operation.Specification[entity.LeaveLedgerEntry, string]{
		Domain: "leave", Table: "hr_leave_ledger_entries",
		Key:         func(v *entity.LeaveLedgerEntry) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LeaveLedgerEntry) string { return v.ID },
		Initialize: func(v *entity.LeaveLedgerEntry, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.ActorPrincipalID = actor.PrincipalID
			v.RecordedAt = now
		},
		Touch: func(v *entity.LeaveLedgerEntry, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LeaveLedgerEntry) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["entitlement_id"] = fmt.Sprint(v.EntitlementID)
			if v.AdjustmentID != nil {
				a["adjustment_id"] = fmt.Sprint(*v.AdjustmentID)
			}
			if v.ReversesEntryID != nil {
				a["reverses_entry_id"] = fmt.Sprint(*v.ReversesEntryID)
			}
			a["actor_principal_id"] = fmt.Sprint(v.ActorPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.LeaveLedgerEntry) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
