package security

import (
	"context"
	"math/big"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leaveAdjustmentValidator struct {
	entitlements leaveEntitlementReader
	policies     leavePolicyReader
	adjustments  leaveAdjustmentReader
	ledger       leaveLedgerReader
}

// ValidateMutation menjaga pengajuan dan persetujuan koreksi sebelum ledger diposting.
func (v *leaveAdjustmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_adjustments" || operation.RequireDependency(v.entitlements) != nil ||
		operation.RequireDependency(v.policies) != nil || operation.RequireDependency(v.adjustments) != nil ||
		operation.RequireDependency(v.ledger) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LeaveAdjustment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LeaveAdjustment)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	var adjustment *entity.LeaveAdjustment
	if after != nil {
		var ok bool
		adjustment, ok = after.(*entity.LeaveAdjustment)
		if !ok || adjustment == nil || (previous != nil && previous.ID != adjustment.ID) {
			return hrerrors.ErrInvalidInput
		}
	} else {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		adjustment = previous
	}
	entitlement, err := v.entitlements.FindByIDForUpdate(ctx, adjustment.EntitlementID)
	if err != nil {
		return err
	}
	if entitlement == nil || entitlement.Revision < 2 || (previous != nil && previous.EntitlementID != adjustment.EntitlementID) {
		return hrerrors.ErrReferenceConflict
	}
	expired, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{
		{Column: "entitlement_id", Values: []any{adjustment.EntitlementID}},
		{Column: "entry_type", Values: []any{"expiry"}},
	})
	if err != nil {
		return err
	}
	rejecting := after != nil && previous != nil && adjustment.RejectedAt != nil
	if (expired != 0 || entitlement.ClosedAt != nil) && !rejecting {
		return hrerrors.ErrWorkflowRequired
	}
	posted, err := v.ledger.CountWhere(ctx, []domainrepo.Filter{{Column: "adjustment_id", Values: []any{adjustment.ID}}})
	if err != nil {
		return err
	}
	if posted > 0 {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil {
		if previous.ApprovedAt != nil || previous.ApprovedByPrincipalID != nil ||
			previous.RejectedAt != nil || previous.RejectedByPrincipalID != nil || previous.RejectionReasonCode != nil ||
			previous.RequestedByPrincipalID != adjustment.RequestedByPrincipalID ||
			!previous.CreatedAt.Equal(adjustment.CreatedAt) || previous.IdempotencyKey != adjustment.IdempotencyKey {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if adjustment.RejectedAt != nil || adjustment.RejectedByPrincipalID != nil || adjustment.RejectionReasonCode != nil {
		if !rejecting || adjustment.RejectedByPrincipalID == nil || adjustment.RejectionReasonCode == nil ||
			strings.TrimSpace(*adjustment.RejectionReasonCode) == "" || len(*adjustment.RejectionReasonCode) > 80 ||
			strings.EqualFold(adjustment.RequestedByPrincipalID, *adjustment.RejectedByPrincipalID) ||
			adjustment.RejectedAt.Before(adjustment.CreatedAt) {
			return hrerrors.ErrWorkflowRequired
		}
		expected := *previous
		expected.RejectedAt = adjustment.RejectedAt
		expected.RejectedByPrincipalID = adjustment.RejectedByPrincipalID
		expected.RejectionReasonCode = adjustment.RejectionReasonCode
		if !reflect.DeepEqual(&expected, adjustment) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if adjustment.ApprovedAt != nil || adjustment.ApprovedByPrincipalID != nil {
		if previous == nil || adjustment.ApprovedAt == nil || adjustment.ApprovedByPrincipalID == nil ||
			strings.EqualFold(adjustment.RequestedByPrincipalID, *adjustment.ApprovedByPrincipalID) ||
			adjustment.ApprovedAt.Before(adjustment.CreatedAt) {
			return hrerrors.ErrWorkflowRequired
		}
		withoutApproval := *adjustment
		withoutApproval.ApprovedAt = nil
		withoutApproval.ApprovedByPrincipalID = nil
		if !reflect.DeepEqual(previous, &withoutApproval) {
			return hrerrors.ErrWorkflowRequired
		}
		if err := checkLeaveBalanceWithPendingAdjustments(ctx, entitlement, adjustment.QuantityDelta, v.policies, v.adjustments, v.ledger); err != nil {
			return err
		}
	}
	if after == nil {
		return nil
	}
	matches, err := v.adjustments.SelectWhere(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{adjustment.IdempotencyKey}}}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != adjustment.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// checkLeaveBalanceWithPendingAdjustments melindungi saldo yang dialokasikan untuk koreksi disetujui.
func checkLeaveBalanceWithPendingAdjustments(
	ctx context.Context, entitlement *entity.LeaveEntitlement, delta entity.Decimal,
	policies leavePolicyReader, adjustments leaveAdjustmentReader, ledger leaveLedgerReader,
) error {
	balance, err := entitlement.ClosingBalanceSnapshot.Rat()
	if err != nil {
		return err
	}
	quantity, err := delta.Rat()
	if err != nil {
		return err
	}
	filter := []domainrepo.Filter{{Column: "entitlement_id", Values: []any{entitlement.ID}}}
	for offset := 0; ; {
		rows, err := adjustments.SelectWhere(ctx, filter, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		ids := make([]any, 0, len(rows))
		unposted := make(map[string]*big.Rat, len(rows))
		for _, existing := range rows {
			if existing.ApprovedAt == nil || existing.ApprovedByPrincipalID == nil {
				continue
			}
			pending, err := existing.QuantityDelta.Rat()
			if err != nil {
				return err
			}
			// Koreksi searah harus dapat dibukukan tanpa bergantung pada urutan posting.
			if pending.Sign() != quantity.Sign() {
				continue
			}
			ids = append(ids, existing.ID)
			unposted[strings.ToLower(existing.ID)] = pending
		}
		if len(ids) != 0 {
			posted, err := ledger.SelectWhere(ctx, []domainrepo.Filter{
				{Column: "entitlement_id", Values: []any{entitlement.ID}},
				{Column: "adjustment_id", Values: ids},
			}, domainrepo.MaxLimit, 0)
			if err != nil {
				return err
			}
			// Satu adjustment_id hanya dapat muncul sekali pada ledger.
			for _, entry := range posted {
				if entry.AdjustmentID != nil {
					delete(unposted, strings.ToLower(*entry.AdjustmentID))
				}
			}
		}
		for _, pending := range unposted {
			balance.Add(balance, pending)
		}
		if len(rows) < domainrepo.MaxLimit {
			break
		}
		offset += len(rows)
	}
	balance.Add(balance, quantity)
	if balance.Sign() < 0 {
		return hrerrors.ErrReferenceConflict
	}
	policy, err := policies.FindByIDForUpdate(ctx, entitlement.LeavePolicyID)
	if err != nil {
		return err
	}
	if policy == nil {
		return hrerrors.ErrReferenceConflict
	}
	if policy.MaxBalance != nil {
		maximum, err := policy.MaxBalance.Rat()
		if err != nil {
			return err
		}
		if balance.Cmp(maximum) > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
