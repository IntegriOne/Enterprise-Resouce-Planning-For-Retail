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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func payrollResult(v *entity.PayrollResult) *pb.PayrollResult {
	return &pb.PayrollResult{
		Id: v.ID, PayrollRunId: v.PayrollRunID, PayrollGroupId: v.PayrollGroupID, WorkerId: v.WorkerID, PayrollProfileReferenceId: v.PayrollProfileReferenceID,
		CurrencyCode: v.CurrencyCode, GrossAmount: string(v.GrossAmount), ReimbursementAmount: string(v.ReimbursementAmount), TotalDeductions: string(v.TotalDeductions),
		EmployerCostAmount: string(v.EmployerCostAmount), NetPayAmount: string(v.NetPayAmount), ResultStatus: v.ResultStatus,
		PaymentReferenceId: v.PaymentReferenceID, PayslipDocumentReferenceId: v.PayslipDocumentReferenceID, CalculationDigest: append([]byte(nil), v.CalculationDigest...), CreatedAt: timestamppb.New(v.CreatedAt),
	}
}
func payrollResults(page operation.Page[entity.PayrollResult], err error) (*pb.PayrollResultsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollResult, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollResult(&page.Items[i]))
	}
	return &pb.PayrollResultsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollResult(ctx context.Context, r *pb.CreatePayrollResultRequest) (*pb.PayrollResultResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	v, err := h.service.CreatePayrollResult(ctx, service.CreatePayrollResultInput{
		PayrollRunID: in.GetPayrollRunId(), PayrollGroupID: in.GetPayrollGroupId(), WorkerID: in.GetWorkerId(), PayrollProfileReferenceID: in.GetPayrollProfileReferenceId(),
		CurrencyCode: in.GetCurrencyCode(), GrossAmount: entity.Decimal(in.GetGrossAmount()), ReimbursementAmount: decimal(in.ReimbursementAmount),
		TotalDeductions: entity.Decimal(in.GetTotalDeductions()), EmployerCostAmount: entity.Decimal(in.GetEmployerCostAmount()), NetPayAmount: entity.Decimal(in.GetNetPayAmount()),
		ResultStatus: in.ResultStatus, PaymentReferenceID: in.PaymentReferenceId, PayslipDocumentReferenceID: in.PayslipDocumentReferenceId, CalculationDigest: append([]byte(nil), in.CalculationDigest...),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultResponse{Item: payrollResult(v)}, nil
}
func (h *Handler) UpdatePayrollResult(ctx context.Context, r *pb.UpdatePayrollResultRequest) (*pb.PayrollResultResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "payroll_run_id", "payroll_group_id", "worker_id", "payroll_profile_reference_id", "currency_code", "gross_amount", "reimbursement_amount", "total_deductions", "employer_cost_amount", "net_pay_amount", "result_status", "payment_reference_id", "payslip_document_reference_id", "calculation_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	v, err := h.service.UpdatePayrollResult(ctx, r.Id, service.UpdatePayrollResultInput{
		PayrollRunID: field(m["payroll_run_id"], in.PayrollRunId), PayrollGroupID: field(m["payroll_group_id"], in.PayrollGroupId), WorkerID: field(m["worker_id"], in.WorkerId),
		PayrollProfileReferenceID: field(m["payroll_profile_reference_id"], in.PayrollProfileReferenceId), CurrencyCode: field(m["currency_code"], in.CurrencyCode),
		GrossAmount: field(m["gross_amount"], decimal(in.GrossAmount)), ReimbursementAmount: field(m["reimbursement_amount"], decimal(in.ReimbursementAmount)),
		TotalDeductions: field(m["total_deductions"], decimal(in.TotalDeductions)), EmployerCostAmount: field(m["employer_cost_amount"], decimal(in.EmployerCostAmount)),
		NetPayAmount: field(m["net_pay_amount"], decimal(in.NetPayAmount)), ResultStatus: field(m["result_status"], in.ResultStatus),
		PaymentReferenceID: field(m["payment_reference_id"], in.PaymentReferenceId), PayslipDocumentReferenceID: field(m["payslip_document_reference_id"], in.PayslipDocumentReferenceId), CalculationDigest: bytesField(m["calculation_digest"], in.CalculationDigest),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultResponse{Item: payrollResult(v)}, nil
}
func (h *Handler) DeletePayrollResult(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollResult(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollResultByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollResultResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollResultByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollResultResponse{Item: payrollResult(v)}, nil
}
func (h *Handler) GetAllPayrollResults(ctx context.Context, r *common.PageRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.GetAllPayrollResults(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByPayrollRunID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByPayrollRunID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByPayrollGroupID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByPayrollGroupID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByWorkerID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByPayrollProfileReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByPayrollProfileReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByResultStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByResultStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByPaymentReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByPaymentReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollResultByPayslipDocumentReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollResultsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollResults(h.service.FindPayrollResultByPayslipDocumentReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
