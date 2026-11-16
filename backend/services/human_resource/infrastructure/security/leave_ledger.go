package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leaveLedgerValidator struct {
	entitlements leaveEntitlementReader
	policies     leavePolicyReader
	adjustments  leaveAdjustmentReader
	ledger       leaveLedgerReader
}

// ValidateMutation mengizinkan opening serta koreksi yang telah disetujui HRM.
func (v *leaveLedgerValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_ledger_entries" || operation.RequireDependency(v.entitlements) != nil ||
		operation.RequireDependency(v.policies) != nil || operation.RequireDependency(v.adjustments) != nil ||
		operation.RequireDependency(v.ledger) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	entry, ok := after.(*entity.LeaveLedgerEntry)
	if !ok || entry == nil {
		return hrerrors.ErrInvalidInput
	}
	if entry.SourceManagement != "hrm" || entry.ReversesEntryID != nil ||
		(entry.EntryType != "opening" && entry.EntryType != "adjustment") {
		return hrerrors.ErrWorkflowRequired
	}
	entitlement, err := v.entitlements.FindByIDForUpdate(ctx, entry.EntitlementID)
	if err != nil {
		return err
	}
	if entitlement == nil ||
		entity.CompareDates(entry.EffectiveDate, entitlement.EntitlementPeriodStart) < 0 ||
		entity.CompareDates(entry.EffectiveDate, entitlement.EntitlementPeriodEnd) > 0 {
		return hrerrors.ErrReferenceConflict
	}
	policy, err := v.policies.FindByIDForUpdate(ctx, entitlement.LeavePolicyID)
	if err != nil {
		return err
	}
	if policy == nil || !policy.IsActive || policy.AccrualMethod != "annual_grant" || policy.AnnualEntitlement == nil {
		return hrerrors.ErrReferenceConflict
	}
	count, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{{Column: "entitlement_id", Values: []any{entry.EntitlementID}}})
	if err != nil {
		return err
	}
	if entry.EntrySequence != count+1 || entitlement.Revision != count+1 {
		return hrerrors.ErrWorkflowRequired
	}
	if entry.EntryType == "opening" {
		if count != 0 || !emptyLeaveBalance(entitlement) || entry.AdjustmentID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		quantity, err := entry.QuantityDelta.Rat()
		if err != nil {
			return err
		}
		annual, err := policy.AnnualEntitlement.Rat()
		if err != nil {
			return err
		}
		if quantity.Cmp(annual) > 0 {
			return hrerrors.Invalid("quantity_delta", "opening grant exceeds annual entitlement")
		}
		if policy.MaxBalance != nil {
			maximum, err := policy.MaxBalance.Rat()
			if err != nil {
				return err
			}
			if quantity.Cmp(maximum) > 0 {
				return hrerrors.Invalid("quantity_delta", "opening grant exceeds maximum balance")
			}
		}
		return nil
	}
	if count == 0 || entry.AdjustmentID == nil {
		return hrerrors.ErrWorkflowRequired
	}
	adjustment, err := v.adjustments.FindByID(ctx, *entry.AdjustmentID)
	if err != nil {
		return err
	}
	if adjustment == nil || adjustment.EntitlementID != entry.EntitlementID ||
		adjustment.ApprovedByPrincipalID == nil || adjustment.ApprovedAt == nil ||
		adjustment.RequestedByPrincipalID == *adjustment.ApprovedByPrincipalID {
		return hrerrors.ErrReferenceConflict
	}
	posted, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{{Column: "adjustment_id", Values: []any{adjustment.ID}}})
	if err != nil {
		return err
	}
	if posted != 0 {
		return hrerrors.ErrAlreadyExists
	}
	quantity, err := entry.QuantityDelta.Rat()
	if err != nil {
		return err
	}
	approved, err := adjustment.QuantityDelta.Rat()
	if err != nil {
		return err
	}
	if quantity.Cmp(approved) != 0 {
		return hrerrors.Invalid("quantity_delta", "entry amount differs from approved adjustment")
	}
	return nil
}
