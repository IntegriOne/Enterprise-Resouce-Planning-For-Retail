package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

const payrollResultReadPermission = "human_resource.payroll_results.read"

// PayrollResult membaca satu hasil kalkulasi untuk reviewer payroll.
func (a *Adapter) PayrollResult(ctx context.Context, rawID string) (*model.PayrollResult, error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, payrollResultReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-result")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll result ID", nil)
	}
	response, err := a.payroll.GetPayrollResultByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll result")
	}
	row := response.GetItem()
	if row.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll result response", nil)
	}
	return toModelPayrollResult(row)
}

// PayrollResults membaca nominal hasil kalkulasi untuk satu run.
func (a *Adapter) PayrollResults(ctx context.Context, rawRunID string, limit, offset int) (*Page[model.PayrollResult], error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, payrollResultReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawRunID, "payroll-run")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll run ID", nil)
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll result pagination", nil)
	}
	runID := id.String()
	response, err := a.payroll.FindPayrollResultByPayrollRunID(rpcCtx, &common.StringFinderRequest{
		Value: &runID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "payroll result")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll result page", nil)
	}
	result := &Page[model.PayrollResult]{
		Items: make([]*model.PayrollResult, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range items {
		if row.GetPayrollRunId() != runID {
			return nil, middleware.ErrUnavailable("invalid human resource payroll result run", nil)
		}
		item, err := toModelPayrollResult(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollResult(row *payroll.PayrollResult) (*model.PayrollResult, error) {
	for _, raw := range []string{row.GetId(), row.GetPayrollRunId(), row.GetWorkerId()} {
		if parsed, err := uuid.Parse(raw); err != nil || parsed == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll result reference", nil)
		}
	}
	return &model.PayrollResult{
		ID: nodeID("payroll-result", row.GetId()), RawID: row.GetId(), PayrollRunID: row.GetPayrollRunId(), WorkerID: row.GetWorkerId(), CurrencyCode: row.GetCurrencyCode(),
		GrossAmount: row.GetGrossAmount(), ReimbursementAmount: row.GetReimbursementAmount(),
		TotalDeductions: row.GetTotalDeductions(), EmployerCostAmount: row.GetEmployerCostAmount(),
		NetPayAmount: row.GetNetPayAmount(), ResultStatus: row.GetResultStatus(),
	}, nil
}
