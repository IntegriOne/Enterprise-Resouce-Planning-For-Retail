package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leaveAdjustmentValidator struct {
	entitlements leaveEntitlementReader
	adjustments  leaveAdjustmentReader
	ledger       leaveLedgerReader
}

// ValidateMutation menjaga pengajuan dan persetujuan koreksi sebelum ledger diposting.
func (v *leaveAdjustmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_adjustments" || operation.RequireDependency(v.entitlements) != nil ||
		operation.RequireDependency(v.adjustments) != nil || operation.RequireDependency(v.ledger) != nil {
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
	if expired != 0 {
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
			previous.RequestedByPrincipalID != adjustment.RequestedByPrincipalID ||
			!previous.CreatedAt.Equal(adjustment.CreatedAt) || previous.IdempotencyKey != adjustment.IdempotencyKey {
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
