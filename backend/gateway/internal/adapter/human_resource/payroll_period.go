package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

const payrollPeriodReadPermission = "human_resource.payroll_periods.read"

// PayrollPeriod membaca detail periode payroll berdasarkan ID.
func (a *Adapter) PayrollPeriod(ctx context.Context, rawID string) (*model.PayrollPeriod, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollPeriodReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-period")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll period ID", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetPayrollPeriodByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll period")
	}
	period := response.GetItem()
	if period.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll period response", nil)
	}
	return toModelPayrollPeriod(period)
}

// PayrollPeriods mendaftar periode payroll dengan izin baca tersendiri.
func (a *Adapter) PayrollPeriods(ctx context.Context, limit, offset int, groupID *string) (*Page[model.PayrollPeriod], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollPeriodReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll period pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *payroll.PayrollPeriodsResponse
	var groupFilter string
	if groupID == nil {
		response, err = a.payroll.GetAllPayrollPeriods(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		id, parseErr := parseNodeID(*groupID, "payroll-group")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid payroll group ID", nil)
		}
		groupFilter = id.String()
		response, err = a.payroll.FindPayrollPeriodByPayrollGroupID(rpcCtx, &common.StringFinderRequest{
			Value: &groupFilter, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll period")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll period page", nil)
	}
	result := &Page[model.PayrollPeriod]{
		Items: make([]*model.PayrollPeriod, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, period := range items {
		if groupID != nil && period.GetPayrollGroupId() != groupFilter {
			return nil, middleware.ErrUnavailable("invalid human resource payroll period group", nil)
		}
		item, err := toModelPayrollPeriod(period)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollPeriod(period *payroll.PayrollPeriod) (*model.PayrollPeriod, error) {
	for _, raw := range []string{period.GetId(), period.GetPayrollGroupId(), period.GetEmployingLegalEntityReferenceId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll period reference", nil)
		}
	}
	return &model.PayrollPeriod{
		ID: nodeID("payroll-period", period.GetId()), RawID: period.GetId(), PayrollGroupID: period.GetPayrollGroupId(),
		EmployingLegalEntityReferenceID: period.GetEmployingLegalEntityReferenceId(),
		CurrencyCode:                    period.GetCurrencyCode(), JurisdictionCode: period.GetJurisdictionCode(),
		PeriodCode: period.GetPeriodCode(), PeriodStart: period.GetPeriodStart(), PeriodEnd: period.GetPeriodEnd(),
		PayDate: period.GetPayDate(), PeriodStatus: period.GetPeriodStatus(),
	}, nil
}
