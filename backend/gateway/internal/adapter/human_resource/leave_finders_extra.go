package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
)

// LeavePoliciesFiltered mendaftar kebijakan cuti dengan filter konsolidasi.
func (a *Adapter) LeavePoliciesFiltered(ctx context.Context, limit, offset int, policyCode, leaveTypeCode, jurisdictionCode *string) (*Page[model.LeavePolicy], error) {
	if policyCode == nil && leaveTypeCode == nil && jurisdictionCode == nil {
		return a.LeavePolicies(ctx, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_leave_policies", "human_resource.leave_policies.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave-policy pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(policyCode != nil, leaveTypeCode != nil, jurisdictionCode != nil)
	if err != nil {
		return nil, err
	}
	var resp *leave.LeavePoliciesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(policyCode, "policy_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeavePolicyByPolicyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(leaveTypeCode, "leave_type_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeavePolicyByLeaveTypeCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(jurisdictionCode, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeavePolicyByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.leave.GetAllLeavePolicies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "leave-policy")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "leave-policy", toModelLeavePolicy)
}

// LeaveEntitlementsFiltered mendaftar hak cuti pekerja dengan filter konsolidasi.
func (a *Adapter) LeaveEntitlementsFiltered(ctx context.Context, limit, offset int, workerID, leavePolicyID, uomCode, sourceWorkforceSnapshotRefID *string) (*Page[model.LeaveEntitlement], error) {
	if workerID != nil && leavePolicyID == nil && uomCode == nil && sourceWorkforceSnapshotRefID == nil {
		return a.LeaveEntitlements(ctx, *workerID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_leave_entitlements", "human_resource.leave_entitlements.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave-entitlement pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(workerID != nil, leavePolicyID != nil, uomCode != nil, sourceWorkforceSnapshotRefID != nil)
	if err != nil {
		return nil, err
	}
	var resp *leave.LeaveEntitlementsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(workerID, "worker", "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveEntitlementByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(leavePolicyID, "leave-policy", "leave_policy_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveEntitlementByLeavePolicyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(uomCode, "uom_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveEntitlementByUOMCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(sourceWorkforceSnapshotRefID, "workforce-snapshot", "source_workforce_snapshot_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.leave.GetAllLeaveEntitlements(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "leave-entitlement")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "leave-entitlement", toModelLeaveEntitlement)
}

// LeaveLedgerEntriesFiltered mendaftar entri buku besar cuti dengan filter konsolidasi.
func (a *Adapter) LeaveLedgerEntriesFiltered(ctx context.Context, limit, offset int, entitlementID, adjustmentID, entryType, sourceManagement, sourceEventID, reasonCode, reversesEntryID, actorPrincipalID *string) (*Page[model.LeaveLedgerEntry], error) {
	if entryType == nil && sourceManagement == nil && sourceEventID == nil && reasonCode == nil && reversesEntryID == nil && actorPrincipalID == nil {
		return a.LeaveLedgerEntries(ctx, entitlementID, adjustmentID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_leave_ledger_entries", "human_resource.leave_ledger_entries.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave-ledger-entry pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		entitlementID != nil,
		adjustmentID != nil,
		entryType != nil,
		sourceManagement != nil,
		sourceEventID != nil,
		reasonCode != nil,
		reversesEntryID != nil,
		actorPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *leave.LeaveLedgerEntriesResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(entitlementID, "leave-entitlement", "entitlement_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByEntitlementID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(adjustmentID, "leave-adjustment", "adjustment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByAdjustmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(entryType, "entry_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByEntryType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(sourceManagement, "source_management")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryBySourceManagement(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(sourceEventID, "source-event", "source_event_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryBySourceEventID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(reasonCode, "reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(reversesEntryID, "leave-ledger-entry", "reverses_entry_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByReversesEntryID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(actorPrincipalID, "principal", "actor_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveLedgerEntryByActorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.leave.GetAllLeaveLedgerEntries(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "leave-ledger-entry")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "leave-ledger-entry", toModelLeaveLedgerEntry)
}

// LeaveAdjustmentsFiltered mendaftar penyesuaian saldo cuti dengan filter konsolidasi.
func (a *Adapter) LeaveAdjustmentsFiltered(ctx context.Context, limit, offset int, entitlementID, idempotencyKey, reasonCode, requestedByPrincipalID, approvedByPrincipalID *string) (*Page[model.LeaveAdjustment], error) {
	if reasonCode == nil && requestedByPrincipalID == nil && approvedByPrincipalID == nil {
		return a.LeaveAdjustments(ctx, entitlementID, idempotencyKey, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_leave_adjustments", "human_resource.leave_adjustments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave-adjustment pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		entitlementID != nil,
		idempotencyKey != nil,
		reasonCode != nil,
		requestedByPrincipalID != nil,
		approvedByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *leave.LeaveAdjustmentsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(entitlementID, "leave-entitlement", "entitlement_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveAdjustmentByEntitlementID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(idempotencyKey, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveAdjustmentByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(reasonCode, "reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveAdjustmentByReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(requestedByPrincipalID, "principal", "requested_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveAdjustmentByRequestedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.leave.FindLeaveAdjustmentByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.leave.GetAllLeaveAdjustments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "leave-adjustment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "leave-adjustment", toModelLeaveAdjustment)
}
