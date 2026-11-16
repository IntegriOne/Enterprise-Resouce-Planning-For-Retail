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

// ValidateMutation mengizinkan opening, koreksi, kedaluwarsa, dan pembalikan opening HRM.
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
	if entry.SourceManagement != "hrm" ||
		(entry.EntryType != "opening" && entry.EntryType != "adjustment" &&
			entry.EntryType != "expiry" && entry.EntryType != "reversal") {
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
	expired, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{
		{Column: "entitlement_id", Values: []any{entry.EntitlementID}},
		{Column: "entry_type", Values: []any{"expiry"}},
	})
	if err != nil {
		return err
	}
	if expired != 0 {
		return hrerrors.ErrWorkflowRequired
	}
	if entry.EntryType == "reversal" {
		if count == 0 || entry.ReversesEntryID == nil || entry.AdjustmentID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		originals, err := v.ledger.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "id", Values: []any{*entry.ReversesEntryID}},
			{Column: "entitlement_id", Values: []any{entry.EntitlementID}},
		}, 1, 0)
		if err != nil {
			return err
		}
		if len(originals) != 1 || originals[0].EntryType != "opening" ||
			originals[0].SourceManagement != "hrm" ||
			originals[0].ActorPrincipalID == entry.ActorPrincipalID ||
			entity.CompareDates(entry.EffectiveDate, originals[0].EffectiveDate) < 0 {
			return hrerrors.ErrWorkflowRequired
		}
		reversals, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{
			{Column: "reverses_entry_id", Values: []any{*entry.ReversesEntryID}},
		})
		if err != nil {
			return err
		}
		if reversals != 0 {
			return hrerrors.ErrAlreadyExists
		}
		negated, err := entity.DecimalNegate(originals[0].QuantityDelta)
		if err != nil {
			return err
		}
		expected, err := negated.Rat()
		if err != nil {
			return err
		}
		actual, err := entry.QuantityDelta.Rat()
		if err != nil {
			return err
		}
		if actual.Cmp(expected) != 0 {
			return hrerrors.Invalid("quantity_delta", "reversal must exactly negate the opening entry")
		}
		return nil
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
	if entry.EntryType == "expiry" {
		if count == 0 || entry.AdjustmentID != nil ||
			entity.CompareDates(entry.EffectiveDate, entitlement.EntitlementPeriodEnd) != 0 ||
			entity.CompareDates(entry.RecordedAt, entry.EffectiveDate) < 0 {
			return hrerrors.ErrWorkflowRequired
		}
		remaining, err := entitlement.ClosingBalanceSnapshot.Rat()
		if err != nil {
			return err
		}
		negative, err := entity.DecimalNegate(entitlement.ClosingBalanceSnapshot)
		if err != nil {
			return err
		}
		expected, err := negative.Rat()
		if err != nil {
			return err
		}
		quantity, err := entry.QuantityDelta.Rat()
		if err != nil {
			return err
		}
		if remaining.Sign() <= 0 || quantity.Cmp(expected) != 0 {
			return hrerrors.Invalid("quantity_delta", "expiry must consume the remaining balance")
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
