package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CreateLeavePolicy membuat kebijakan cuti baru melalui HRM.
func (a *Adapter) CreateLeavePolicy(ctx context.Context, input model.CreateLeavePolicyInput) (*model.LeavePolicy, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_leave_policies", "create", "human_resource.leave_policies.create")
	if err != nil {
		return nil, err
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &leave.LeavePolicyInput{
		PolicyCode:         input.PolicyCode,
		PolicyVersion:      optionalInt32Ptr(input.PolicyVersion),
		PolicyName:         input.PolicyName,
		LeaveTypeCode:      input.LeaveTypeCode,
		JurisdictionCode:   input.JurisdictionCode,
		AccrualMethod:      input.AccrualMethod,
		AnnualEntitlement:  input.AnnualEntitlement,
		CarryoverLimit:     input.CarryoverLimit,
		MaxBalance:         input.MaxBalance,
		MinimumServiceDays: optionalInt32Ptr(input.MinimumServiceDays),
		EffectiveFrom:      input.EffectiveFrom,
		EffectiveTo:        input.EffectiveTo,
		IsPaid:             input.IsPaid,
		IsActive:           input.IsActive,
	}
	resp, err := a.leave.CreateLeavePolicy(rpcCtx, &leave.CreateLeavePolicyRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "leave-policy")
	}
	return toModelLeavePolicy(resp.GetItem())
}

// UpdateLeavePolicy memperbarui kebijakan cuti berdasarkan FieldMask.
func (a *Adapter) UpdateLeavePolicy(ctx context.Context, rawID string, input model.UpdateLeavePolicyInput) (*model.LeavePolicy, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_leave_policies", "update", "human_resource.leave_policies.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "leave-policy", "leave-policy")
	if err != nil {
		return nil, err
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &leave.LeavePolicyInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "policy_code", input.PolicyCode, &patch.PolicyCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "policy_version", input.PolicyVersion, &patch.PolicyVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "policy_name", input.PolicyName, &patch.PolicyName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "leave_type_code", input.LeaveTypeCode, &patch.LeaveTypeCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "accrual_method", input.AccrualMethod, &patch.AccrualMethod, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "annual_entitlement", input.AnnualEntitlement, &patch.AnnualEntitlement, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "carryover_limit", input.CarryoverLimit, &patch.CarryoverLimit, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "max_balance", input.MaxBalance, &patch.MaxBalance, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "minimum_service_days", input.MinimumServiceDays, &patch.MinimumServiceDays, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_paid", input.IsPaid, &patch.IsPaid, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("leave-policy update is empty", nil)
	}
	resp, err := a.leave.UpdateLeavePolicy(rpcCtx, &leave.UpdateLeavePolicyRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "leave-policy")
	}
	return toModelLeavePolicy(resp.GetItem())
}

// DeleteLeavePolicy menghapus kebijakan cuti berdasarkan ID.
func (a *Adapter) DeleteLeavePolicy(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_leave_policies", "human_resource.leave_policies.delete", "leave-policy", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.leave.DeleteLeavePolicy(c, req)
	})
}

// CreateLeaveEntitlement membuat jatah cuti pekerja baru melalui HRM.
func (a *Adapter) CreateLeaveEntitlement(ctx context.Context, input model.CreateLeaveEntitlementInput) (*model.LeaveEntitlement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_leave_entitlements", "create", "human_resource.leave_entitlements.create")
	if err != nil {
		return nil, err
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	rev, err := optionalInt64FromStr(input.Revision, "revision")
	if err != nil {
		return nil, err
	}
	patch := &leave.LeaveEntitlementInput{
		WorkerId:                           input.WorkerID,
		LeavePolicyId:                      input.LeavePolicyID,
		LeaveTypeCodeSnapshot:              input.LeaveTypeCodeSnapshot,
		EntitlementPeriodStart:             input.EntitlementPeriodStart,
		EntitlementPeriodEnd:               input.EntitlementPeriodEnd,
		OpeningBalance:                     input.OpeningBalance,
		AccruedQuantity:                    input.AccruedQuantity,
		UsedQuantitySnapshot:               input.UsedQuantitySnapshot,
		AdjustedQuantity:                   input.AdjustedQuantity,
		ClosingBalanceSnapshot:             input.ClosingBalanceSnapshot,
		UomCode:                            input.UomCode,
		Revision:                           rev,
		SourceWorkforceSnapshotReferenceId: input.SourceWorkforceSnapshotReferenceID,
	}
	resp, err := a.leave.CreateLeaveEntitlement(rpcCtx, &leave.CreateLeaveEntitlementRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "leave-entitlement")
	}
	return toModelLeaveEntitlement(resp.GetItem())
}

// UpdateLeaveEntitlement memperbarui jatah cuti pekerja berdasarkan FieldMask.
func (a *Adapter) UpdateLeaveEntitlement(ctx context.Context, rawID string, input model.UpdateLeaveEntitlementInput) (*model.LeaveEntitlement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_leave_entitlements", "update", "human_resource.leave_entitlements.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "leave-entitlement", "leave-entitlement")
	if err != nil {
		return nil, err
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &leave.LeaveEntitlementInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "leave_policy_id", input.LeavePolicyID, &patch.LeavePolicyId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "leave_type_code_snapshot", input.LeaveTypeCodeSnapshot, &patch.LeaveTypeCodeSnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "entitlement_period_start", input.EntitlementPeriodStart, &patch.EntitlementPeriodStart, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "entitlement_period_end", input.EntitlementPeriodEnd, &patch.EntitlementPeriodEnd, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "opening_balance", input.OpeningBalance, &patch.OpeningBalance, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "accrued_quantity", input.AccruedQuantity, &patch.AccruedQuantity, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "used_quantity_snapshot", input.UsedQuantitySnapshot, &patch.UsedQuantitySnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "adjusted_quantity", input.AdjustedQuantity, &patch.AdjustedQuantity, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "closing_balance_snapshot", input.ClosingBalanceSnapshot, &patch.ClosingBalanceSnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "uom_code", input.UomCode, &patch.UomCode, true); err != nil {
		return nil, err
	}
	if err := applyInt64Mask(mask, "revision", input.Revision, &patch.Revision, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "source_workforce_snapshot_reference_id", input.SourceWorkforceSnapshotReferenceID, &patch.SourceWorkforceSnapshotReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("leave-entitlement update is empty", nil)
	}
	resp, err := a.leave.UpdateLeaveEntitlement(rpcCtx, &leave.UpdateLeaveEntitlementRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "leave-entitlement")
	}
	return toModelLeaveEntitlement(resp.GetItem())
}

// DeleteLeaveEntitlement menghapus jatah cuti pekerja berdasarkan ID.
func (a *Adapter) DeleteLeaveEntitlement(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_leave_entitlements", "human_resource.leave_entitlements.delete", "leave-entitlement", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.leave.DeleteLeaveEntitlement(c, req)
	})
}

// UpdateLeaveAdjustment memperbarui penyesuaian saldo cuti berdasarkan FieldMask.
func (a *Adapter) UpdateLeaveAdjustment(ctx context.Context, rawID string, input model.UpdateLeaveAdjustmentInput) (*model.LeaveAdjustment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_leave_adjustments", "update", "human_resource.leave_adjustments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "leave-adjustment", "leave-adjustment")
	if err != nil {
		return nil, err
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &leave.LeaveAdjustmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "entitlement_id", input.EntitlementID, &patch.EntitlementId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "quantity_delta", input.QuantityDelta, &patch.QuantityDelta, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reason_code", input.ReasonCode, &patch.ReasonCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reason", input.Reason, &patch.Reason, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("leave-adjustment update is empty", nil)
	}
	resp, err := a.leave.UpdateLeaveAdjustment(rpcCtx, &leave.UpdateLeaveAdjustmentRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "leave-adjustment")
	}
	return toModelLeaveAdjustment(resp.GetItem())
}

// DeleteLeaveAdjustment menghapus penyesuaian saldo cuti berdasarkan ID.
func (a *Adapter) DeleteLeaveAdjustment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_leave_adjustments", "human_resource.leave_adjustments.delete", "leave-adjustment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.leave.DeleteLeaveAdjustment(c, req)
	})
}
