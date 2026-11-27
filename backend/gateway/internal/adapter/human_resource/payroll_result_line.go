package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

const payrollResultLineReadPermission = "human_resource.payroll_result_lines.read"

// PayrollResultLine membaca detail satu komponen hasil payroll.
func (a *Adapter) PayrollResultLine(ctx context.Context, rawID string) (*model.PayrollResultLine, error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, payrollResultLineReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-result-line")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll result line ID", nil)
	}
	response, err := a.payroll.GetPayrollResultLineByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll result line")
	}
	line := response.GetItem()
	if line.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll result line response", nil)
	}
	return toModelPayrollResultLine(line)
}

// PayrollResultLines membaca komponen untuk satu hasil kalkulasi payroll.
func (a *Adapter) PayrollResultLines(ctx context.Context, rawResultID string, limit, offset int) (*Page[model.PayrollResultLine], error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, payrollResultLineReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawResultID, "payroll-result")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll result ID", nil)
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll result line pagination", nil)
	}
	resultID := id.String()
	response, err := a.payroll.FindPayrollResultLineByPayrollResultID(rpcCtx, &common.StringFinderRequest{
		Value: &resultID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "payroll result line")
	}
	page, lines := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(lines)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll result line page", nil)
	}
	result := &Page[model.PayrollResultLine]{
		Items: make([]*model.PayrollResultLine, 0, len(lines)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, line := range lines {
		if line.GetPayrollResultId() != resultID {
			return nil, middleware.ErrUnavailable("invalid human resource payroll result line reference", nil)
		}
		item, err := toModelPayrollResultLine(line)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollResultLine(line *payroll.PayrollResultLine) (*model.PayrollResultLine, error) {
	for _, raw := range []string{line.GetId(), line.GetPayrollResultId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll result line reference", nil)
		}
	}
	return &model.PayrollResultLine{
		ID: nodeID("payroll-result-line", line.GetId()), RawID: line.GetId(), PayrollResultID: line.GetPayrollResultId(),
		ComponentCode: line.GetComponentCodeSnapshot(), ComponentName: line.GetComponentNameSnapshot(),
		ComponentType: line.GetComponentTypeSnapshot(), Amount: line.GetAmount(), Quantity: line.Quantity,
		Rate: line.Rate, TaxableBase: line.TaxableBase, LineSequence: int(line.GetLineSequence()),
	}, nil
}
