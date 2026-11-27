package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

const payrollGroupReadPermission = "human_resource.payroll_groups.read"

// PayrollGroup membaca detail berdasarkan ID.
func (a *Adapter) PayrollGroup(ctx context.Context, rawID string) (*model.PayrollGroup, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollGroupReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-group")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll group ID", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetPayrollGroupByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll group")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll group response", nil)
	}
	return toModelPayrollGroup(item)
}

// PayrollGroups mendaftar grup payroll dengan izin baca tersendiri.
func (a *Adapter) PayrollGroups(ctx context.Context, limit, offset int) (*Page[model.PayrollGroup], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollGroupReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll group pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetAllPayrollGroups(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, mapReadError(err, "payroll group")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll group page", nil)
	}
	result := &Page[model.PayrollGroup]{
		Items: make([]*model.PayrollGroup, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, group := range items {
		item, err := toModelPayrollGroup(group)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollGroup(group *payroll.PayrollGroup) (*model.PayrollGroup, error) {
	for _, raw := range []string{group.GetId(), group.GetEmployingLegalEntityReferenceId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll group reference", nil)
		}
	}
	return &model.PayrollGroup{
		ID: nodeID("payroll-group", group.GetId()), RawID: group.GetId(), GroupCode: group.GetGroupCode(), GroupName: group.GetGroupName(),
		EmployingLegalEntityReferenceID: group.GetEmployingLegalEntityReferenceId(),
		JurisdictionCode:                group.GetJurisdictionCode(), CurrencyCode: group.GetCurrencyCode(),
		CurrencyScale: int(group.GetCurrencyScale()), PayFrequency: group.GetPayFrequency(),
		Timezone: group.GetTimezone(), IsActive: group.GetIsActive(),
	}, nil
}
