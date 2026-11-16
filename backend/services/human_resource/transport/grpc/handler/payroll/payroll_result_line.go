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

func payrollResultLine(v *entity.PayrollResultLine) *pb.PayrollResultLine {
	return &pb.PayrollResultLine{
		Id: v.ID, PayrollResultId: v.PayrollResultID, ComponentId: v.ComponentID, ComponentCodeSnapshot: v.ComponentCodeSnapshot,
		ComponentNameSnapshot: v.ComponentNameSnapshot, ComponentTypeSnapshot: v.ComponentTypeSnapshot, Amount: string(v.Amount),
		Quantity: decimalString(v.Quantity), Rate: decimalString(v.Rate), TaxableBase: decimalString(v.TaxableBase), CalculationReference: v.CalculationReference, LineSequence: v.LineSequence,
	}
}
func payrollResultLines(page operation.Page[entity.PayrollResultLine], err error) (*pb.PayrollResultLinesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollResultLine, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollResultLine(&page.Items[i]))
	}
	return &pb.PayrollResultLinesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollResultLine(ctx context.Context, r *pb.CreatePayrollResultLineRequest) (*pb.PayrollResultLineResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	v, err := h.service.CreatePayrollResultLine(ctx, service.CreatePayrollResultLineInput{
		PayrollResultID: in.GetPayrollResultId(), ComponentID: in.GetComponentId(), ComponentCodeSnapshot: in.GetComponentCodeSnapshot(),
		ComponentNameSnapshot: in.GetComponentNameSnapshot(), ComponentTypeSnapshot: in.GetComponentTypeSnapshot(), Amount: entity.Decimal(in.GetAmount()),
		Quantity: decimal(in.Quantity), Rate: decimal(in.Rate), TaxableBase: decimal(in.TaxableBase), CalculationReference: in.CalculationReference, LineSequence: in.GetLineSequence(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultLineResponse{Item: payrollResultLine(v)}, nil
}
func (h *Handler) UpdatePayrollResultLine(ctx context.Context, r *pb.UpdatePayrollResultLineRequest) (*pb.PayrollResultLineResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "payroll_result_id", "component_id", "component_code_snapshot", "component_name_snapshot", "component_type_snapshot", "amount", "quantity", "rate", "taxable_base", "calculation_reference", "line_sequence")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	v, err := h.service.UpdatePayrollResultLine(ctx, r.Id, service.UpdatePayrollResultLineInput{
		PayrollResultID: field(m["payroll_result_id"], in.PayrollResultId), ComponentID: field(m["component_id"], in.ComponentId),
		ComponentCodeSnapshot: field(m["component_code_snapshot"], in.ComponentCodeSnapshot), ComponentNameSnapshot: field(m["component_name_snapshot"], in.ComponentNameSnapshot),
		ComponentTypeSnapshot: field(m["component_type_snapshot"], in.ComponentTypeSnapshot), Amount: field(m["amount"], decimal(in.Amount)),
		Quantity: field(m["quantity"], decimal(in.Quantity)), Rate: field(m["rate"], decimal(in.Rate)), TaxableBase: field(m["taxable_base"], decimal(in.TaxableBase)),
		CalculationReference: field(m["calculation_reference"], in.CalculationReference), LineSequence: field(m["line_sequence"], in.LineSequence),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultLineResponse{Item: payrollResultLine(v)}, nil
}
func (h *Handler) DeletePayrollResultLine(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollResultLine(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollResultLineByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollResultLineResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollResultLineByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultLineResponse{Item: payrollResultLine(v)}, nil
}
func (h *Handler) GetAllPayrollResultLines(ctx context.Context, r *common.PageRequest) (*pb.PayrollResultLinesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResultLines(h.service.GetAllPayrollResultLines(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultLineByPayrollResultID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultLinesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResultLines(h.service.FindPayrollResultLineByPayrollResultID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultLineByComponentID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultLinesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResultLines(h.service.FindPayrollResultLineByComponentID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
