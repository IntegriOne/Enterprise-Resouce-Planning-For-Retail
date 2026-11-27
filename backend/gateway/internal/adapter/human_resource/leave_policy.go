package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	leave "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
)

const leavePolicyReadPermission = "human_resource.leave_policies.read"

// LeavePolicy membaca detail kebijakan cuti berdasarkan ID.
func (a *Adapter) LeavePolicy(ctx context.Context, rawID string) (*model.LeavePolicy, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leavePolicyReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "leave-policy")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid leave policy ID", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.GetLeavePolicyByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "leave policy")
	}
	policy := response.GetItem()
	if policy.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource leave policy response", nil)
	}
	return toModelLeavePolicy(policy)
}

// LeavePolicies mendaftar kebijakan cuti dengan izin baca tersendiri.
func (a *Adapter) LeavePolicies(ctx context.Context, limit, offset int) (*Page[model.LeavePolicy], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, leavePolicyReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid leave policy pagination", nil)
	}
	if a.leave == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.leave.GetAllLeavePolicies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, mapReadError(err, "leave policy")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource leave policy page", nil)
	}
	result := &Page[model.LeavePolicy]{
		Items: make([]*model.LeavePolicy, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, policy := range items {
		item, err := toModelLeavePolicy(policy)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelLeavePolicy(policy *leave.LeavePolicy) (*model.LeavePolicy, error) {
	if id, err := uuid.Parse(policy.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource leave policy reference", nil)
	}
	item := &model.LeavePolicy{
		ID: nodeID("leave-policy", policy.GetId()), RawID: policy.GetId(), PolicyCode: policy.GetPolicyCode(), PolicyVersion: int(policy.GetPolicyVersion()),
		PolicyName: policy.GetPolicyName(), LeaveTypeCode: policy.GetLeaveTypeCode(), JurisdictionCode: policy.GetJurisdictionCode(),
		AccrualMethod: policy.GetAccrualMethod(), AnnualEntitlement: policy.AnnualEntitlement, CarryoverLimit: policy.CarryoverLimit,
		MaxBalance: policy.MaxBalance, EffectiveFrom: policy.GetEffectiveFrom(), EffectiveTo: policy.EffectiveTo,
		IsPaid: policy.GetIsPaid(), IsActive: policy.GetIsActive(),
	}
	if policy.MinimumServiceDays != nil {
		days := int(policy.GetMinimumServiceDays())
		item.MinimumServiceDays = &days
	}
	return item, nil
}
