package security

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	leavepolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leaveEntitlementValidator struct {
	workers     leaveWorkerReader
	policies    leavePolicyReader
	adjustments leaveAdjustmentReader
	ledger      leaveLedgerReader
}

// ValidateMutation membatasi saldo hak cuti pada draft kosong atau proyeksi ledger HRM sah.
func (v *leaveEntitlementValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_entitlements" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.policies) != nil || operation.RequireDependency(v.adjustments) != nil ||
		operation.RequireDependency(v.ledger) != nil {
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
	if previous.Revision != count || entitlement.Revision != previous.Revision+1 ||
		previous.WorkerID != entitlement.WorkerID || previous.LeavePolicyID != entitlement.LeavePolicyID ||
		previous.LeaveTypeCodeSnapshot != entitlement.LeaveTypeCodeSnapshot ||
		!previous.EntitlementPeriodStart.Equal(entitlement.EntitlementPeriodStart) ||
		!previous.EntitlementPeriodEnd.Equal(entitlement.EntitlementPeriodEnd) ||
		previous.UOMCode != entitlement.UOMCode || count > 10000 {
		return hrerrors.ErrWorkflowRequired
	}
	entries := make([]entity.LeaveLedgerEntry, 0, int(count))
	for offset := 0; offset < int(count); offset += 100 {
		page, err := v.ledger.SelectWhere(ctx, filter, 100, offset)
		if err != nil {
			return err
		}
		entries = append(entries, page...)
	}
	if int64(len(entries)) != count {
		return hrerrors.ErrWorkflowRequired
	}
	adjustments := make([]entity.LeaveAdjustment, 0)
	seen := make(map[string]bool)
	lastRecorded := false
	for _, entry := range entries {
		if entry.EntryType != "opening" && entry.EntryType != "adjustment" && entry.EntryType != "expiry" {
			return hrerrors.ErrWorkflowRequired
		}
		if entry.EntryType == "expiry" {
			if entry.EntrySequence != count || entity.CompareDates(entry.EffectiveDate, entitlement.EntitlementPeriodEnd) != 0 {
				return hrerrors.ErrWorkflowRequired
			}
		}
		if entry.EntrySequence == count && entry.RecordedAt.Truncate(time.Microsecond).Equal(entitlement.CalculatedAt.Truncate(time.Microsecond)) {
			lastRecorded = true
		}
		if entry.AdjustmentID == nil {
			continue
		}
		key := strings.ToLower(*entry.AdjustmentID)
		if seen[key] {
			return hrerrors.ErrWorkflowRequired
		}
		adjustment, err := v.adjustments.FindByID(ctx, key)
		if err != nil {
			return err
		}
		if adjustment == nil {
			return hrerrors.ErrReferenceConflict
		}
		adjustments = append(adjustments, *adjustment)
		seen[key] = true
	}
	if !lastRecorded {
		return hrerrors.ErrWorkflowRequired
	}
	expected, err := leavepolicy.RebuildBalance(leavepolicy.BalanceInput{
		Entitlement: previous, Entries: entries, Adjustments: adjustments, CalculatedAt: entitlement.CalculatedAt,
	})
	if err != nil {
		return err
	}
	if policy.MaxBalance != nil {
		maximum, err := policy.MaxBalance.Rat()
		if err != nil {
			return err
		}
		closing, err := expected.ClosingBalanceSnapshot.Rat()
		if err != nil {
			return err
		}
		if closing.Cmp(maximum) > 0 {
			return hrerrors.Invalid("closing_balance_snapshot", "balance exceeds policy maximum")
		}
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
