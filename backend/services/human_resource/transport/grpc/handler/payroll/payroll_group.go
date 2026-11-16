package payroll

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func payrollGroup(v *entity.PayrollGroup) *pb.PayrollGroup {
	return &pb.PayrollGroup{
		Id: v.ID, GroupCode: v.GroupCode, GroupName: v.GroupName, EmployingLegalEntityReferenceId: v.EmployingLegalEntityReferenceID,
		JurisdictionCode: v.JurisdictionCode, CurrencyCode: v.CurrencyCode, CurrencyScale: int32(v.CurrencyScale), PayFrequency: v.PayFrequency, Timezone: v.Timezone, IsActive: v.IsActive,
	}
}
func payrollGroups(page operation.Page[entity.PayrollGroup], err error) (*pb.PayrollGroupsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollGroup, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollGroup(&page.Items[i]))
	}
	return &pb.PayrollGroupsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollGroup(ctx context.Context, r *pb.CreatePayrollGroupRequest) (*pb.PayrollGroupResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	if in.GetCurrencyScale() < -32768 || in.GetCurrencyScale() > 32767 {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.CreatePayrollGroup(ctx, service.CreatePayrollGroupInput{
		GroupCode: in.GetGroupCode(), GroupName: in.GetGroupName(), EmployingLegalEntityReferenceID: in.GetEmployingLegalEntityReferenceId(),
		JurisdictionCode: in.GetJurisdictionCode(), CurrencyCode: in.GetCurrencyCode(), CurrencyScale: int16(in.GetCurrencyScale()), PayFrequency: in.GetPayFrequency(), Timezone: in.GetTimezone(), IsActive: in.IsActive,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollGroupResponse{Item: payrollGroup(v)}, nil
}
func (h *Handler) UpdatePayrollGroup(ctx context.Context, r *pb.UpdatePayrollGroupRequest) (*pb.PayrollGroupResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "group_code", "group_name", "employing_legal_entity_reference_id", "jurisdiction_code", "currency_code", "currency_scale", "pay_frequency", "timezone", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var scale *int16
	if m["currency_scale"] && in.CurrencyScale != nil {
		if *in.CurrencyScale < -32768 || *in.CurrencyScale > 32767 {
			return nil, response.Error(hrerrors.ErrInvalidInput)
		}
		v := int16(*in.CurrencyScale)
		scale = &v
	}
	v, err := h.service.UpdatePayrollGroup(ctx, r.Id, service.UpdatePayrollGroupInput{
		GroupCode: field(m["group_code"], in.GroupCode), GroupName: field(m["group_name"], in.GroupName),
		EmployingLegalEntityReferenceID: field(m["employing_legal_entity_reference_id"], in.EmployingLegalEntityReferenceId),
		JurisdictionCode:                field(m["jurisdiction_code"], in.JurisdictionCode), CurrencyCode: field(m["currency_code"], in.CurrencyCode),
		CurrencyScale: field(m["currency_scale"], scale), PayFrequency: field(m["pay_frequency"], in.PayFrequency), Timezone: field(m["timezone"], in.Timezone), IsActive: field(m["is_active"], in.IsActive),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollGroupResponse{Item: payrollGroup(v)}, nil
}
func (h *Handler) DeletePayrollGroup(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollGroup(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollGroupByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollGroupResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollGroupByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollGroupResponse{Item: payrollGroup(v)}, nil
}
func (h *Handler) GetAllPayrollGroups(ctx context.Context, r *common.PageRequest) (*pb.PayrollGroupsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollGroups(h.service.GetAllPayrollGroups(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollGroupByGroupCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollGroupsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollGroups(h.service.FindPayrollGroupByGroupCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollGroupByEmployingLegalEntityReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollGroupsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollGroups(h.service.FindPayrollGroupByEmployingLegalEntityReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollGroupByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollGroupsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollGroups(h.service.FindPayrollGroupByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollGroupByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollGroupsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollGroups(h.service.FindPayrollGroupByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
