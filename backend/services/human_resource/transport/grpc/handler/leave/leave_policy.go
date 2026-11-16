package leave

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func leavePolicy(v *entity.LeavePolicy) *pb.LeavePolicy {
	return &pb.LeavePolicy{
		Id: v.ID, PolicyCode: v.PolicyCode, PolicyVersion: v.PolicyVersion, PolicyName: v.PolicyName,
		LeaveTypeCode: v.LeaveTypeCode, JurisdictionCode: v.JurisdictionCode, AccrualMethod: v.AccrualMethod,
		AnnualEntitlement: decimalString(v.AnnualEntitlement), CarryoverLimit: decimalString(v.CarryoverLimit),
		MaxBalance: decimalString(v.MaxBalance), MinimumServiceDays: v.MinimumServiceDays,
		EffectiveFrom: v.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(v.EffectiveTo), IsPaid: v.IsPaid, IsActive: v.IsActive,
	}
}
func leavePolicies(page operation.Page[entity.LeavePolicy], err error) (*pb.LeavePoliciesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LeavePolicy, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, leavePolicy(&page.Items[i]))
	}
	return &pb.LeavePoliciesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (h *Handler) CreateLeavePolicy(ctx context.Context, r *pb.CreateLeavePolicyRequest) (*pb.LeavePolicyResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreateLeavePolicy(ctx, service.CreateLeavePolicyInput{
		PolicyCode: in.GetPolicyCode(), PolicyVersion: in.GetPolicyVersion(), PolicyName: in.GetPolicyName(),
		LeaveTypeCode: in.GetLeaveTypeCode(), JurisdictionCode: in.GetJurisdictionCode(), AccrualMethod: in.GetAccrualMethod(),
		AnnualEntitlement: decimal(in.AnnualEntitlement), CarryoverLimit: decimal(in.CarryoverLimit),
		MaxBalance: decimal(in.MaxBalance), MinimumServiceDays: in.MinimumServiceDays,
		EffectiveFrom: from, EffectiveTo: to, IsPaid: in.IsPaid, IsActive: in.IsActive,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeavePolicyResponse{Item: leavePolicy(v)}, nil
}
func (h *Handler) UpdateLeavePolicy(ctx context.Context, r *pb.UpdateLeavePolicyRequest) (*pb.LeavePolicyResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "policy_code", "policy_version", "policy_name", "leave_type_code", "jurisdiction_code", "accrual_method", "annual_entitlement", "carryover_limit", "max_balance", "minimum_service_days", "effective_from", "effective_to", "is_paid", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := dateField(m["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(m["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdateLeavePolicy(ctx, r.Id, service.UpdateLeavePolicyInput{
		PolicyCode: field(m["policy_code"], in.PolicyCode), PolicyVersion: field(m["policy_version"], in.PolicyVersion),
		PolicyName: field(m["policy_name"], in.PolicyName), LeaveTypeCode: field(m["leave_type_code"], in.LeaveTypeCode),
		JurisdictionCode: field(m["jurisdiction_code"], in.JurisdictionCode), AccrualMethod: field(m["accrual_method"], in.AccrualMethod),
		AnnualEntitlement: field(m["annual_entitlement"], decimal(in.AnnualEntitlement)), CarryoverLimit: field(m["carryover_limit"], decimal(in.CarryoverLimit)),
		MaxBalance: field(m["max_balance"], decimal(in.MaxBalance)), MinimumServiceDays: field(m["minimum_service_days"], in.MinimumServiceDays),
		EffectiveFrom: from, EffectiveTo: to, IsPaid: field(m["is_paid"], in.IsPaid), IsActive: field(m["is_active"], in.IsActive),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeavePolicyResponse{Item: leavePolicy(v)}, nil
}
func (h *Handler) DeleteLeavePolicy(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLeavePolicy(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLeavePolicyByID(ctx context.Context, r *common.IDRequest) (*pb.LeavePolicyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetLeavePolicyByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeavePolicyResponse{Item: leavePolicy(v)}, nil
}
func (h *Handler) GetAllLeavePolicies(ctx context.Context, r *common.PageRequest) (*pb.LeavePoliciesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leavePolicies(h.service.GetAllLeavePolicies(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeavePolicyByPolicyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeavePoliciesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leavePolicies(h.service.FindLeavePolicyByPolicyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeavePolicyByLeaveTypeCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeavePoliciesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leavePolicies(h.service.FindLeavePolicyByLeaveTypeCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeavePolicyByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeavePoliciesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leavePolicies(h.service.FindLeavePolicyByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
