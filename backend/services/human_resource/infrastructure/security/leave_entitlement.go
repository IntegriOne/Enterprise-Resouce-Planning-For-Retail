package security

import (
	"context"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	leavepolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leaveEntitlementValidator struct {
	workers  leaveWorkerReader
	policies leavePolicyReader
	ledger   leaveLedgerReader
}

// ValidateMutation membatasi saldo hak cuti pada draft kosong atau proyeksi posting opening.
func (v *leaveEntitlementValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_entitlements" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.policies) != nil || operation.RequireDependency(v.ledger) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LeaveEntitlement
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LeaveEntitlement)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	var entitlement *entity.LeaveEntitlement
	if after != nil {
		var ok bool
		entitlement, ok = after.(*entity.LeaveEntitlement)
		if !ok || entitlement == nil || (previous != nil && previous.ID != entitlement.ID) {
			return hrerrors.ErrInvalidInput
		}
	} else {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		entitlement = previous
	}
	if entitlement.SourceWorkforceSnapshotReferenceID != nil || entitlement.UOMCode != "day" {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, entitlement.WorkerID)
	if err != nil {
		return err
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		(worker.HireDate != nil && entity.CompareDates(*worker.HireDate, entitlement.EntitlementPeriodStart) > 0) ||
		(worker.TerminationDate != nil && entity.CompareDates(*worker.TerminationDate, entitlement.EntitlementPeriodStart) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	policy, err := v.policies.FindByIDForUpdate(ctx, entitlement.LeavePolicyID)
	if err != nil {
		return err
	}
	if policy == nil || !policy.IsActive || policy.AccrualMethod != "annual_grant" ||
		policy.LeaveTypeCode != entitlement.LeaveTypeCodeSnapshot ||
		entity.CompareDates(policy.EffectiveFrom, entitlement.EntitlementPeriodStart) > 0 ||
		(policy.EffectiveTo != nil && entity.CompareDates(*policy.EffectiveTo, entitlement.EntitlementPeriodEnd) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	if policy.MinimumServiceDays != nil && *policy.MinimumServiceDays > 0 {
		serviceStart := worker.ServiceStartDate
		if serviceStart == nil {
			serviceStart = worker.HireDate
		}
		if serviceStart == nil || entity.CompareDates(entitlement.EntitlementPeriodStart,
			serviceStart.AddDate(0, 0, int(*policy.MinimumServiceDays))) < 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	filter := []domainrepo.Filter{{Column: "entitlement_id", Values: []any{entitlement.ID}}}
	count, err := v.ledger.CountWhere(ctx, filter)
	if err != nil {
		return err
	}
	if after == nil || previous == nil {
		if count != 0 || entitlement.Revision != 1 || !emptyLeaveBalance(entitlement) {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	if previous.Revision != 1 || !emptyLeaveBalance(previous) || entitlement.Revision != 2 ||
		previous.WorkerID != entitlement.WorkerID || previous.LeavePolicyID != entitlement.LeavePolicyID ||
		previous.LeaveTypeCodeSnapshot != entitlement.LeaveTypeCodeSnapshot ||
		!previous.EntitlementPeriodStart.Equal(entitlement.EntitlementPeriodStart) ||
		!previous.EntitlementPeriodEnd.Equal(entitlement.EntitlementPeriodEnd) ||
		previous.UOMCode != entitlement.UOMCode || count != 1 {
		return hrerrors.ErrWorkflowRequired
	}
	entries, err := v.ledger.SelectWhere(ctx, filter, 2, 0)
	if err != nil {
		return err
	}
	if len(entries) != 1 || entries[0].EntryType != "opening" ||
		!entries[0].RecordedAt.Truncate(time.Microsecond).Equal(entitlement.CalculatedAt.Truncate(time.Microsecond)) {
		return hrerrors.ErrWorkflowRequired
	}
	expected, err := leavepolicy.RebuildBalance(leavepolicy.BalanceInput{
		Entitlement: previous, Entries: entries, CalculatedAt: entitlement.CalculatedAt,
	})
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, entitlement) {
		return hrerrors.ErrWorkflowRequired
	}
	return nil
}

func emptyLeaveBalance(entitlement *entity.LeaveEntitlement) bool {
	for _, value := range []entity.Decimal{
		entitlement.OpeningBalance, entitlement.AccruedQuantity, entitlement.UsedQuantitySnapshot,
		entitlement.AdjustedQuantity, entitlement.ClosingBalanceSnapshot,
	} {
		number, err := value.Rat()
		if err != nil || number.Sign() != 0 {
			return false
		}
	}
	return true
}
