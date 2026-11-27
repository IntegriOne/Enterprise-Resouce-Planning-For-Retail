package payroll

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetMyPayslip mengembalikan metadata slip setelah pemeriksaan kepemilikan.
func (h *Handler) GetMyPayslip(ctx context.Context, request *pb.GetMyPayslipRequest) (*pb.GetMyPayslipResponse, error) {
	if request.GetIncludePdf() {
		slip, err := h.service.GetMyPayslipWithPDF(ctx, request.GetPayrollResultId())
		if err != nil {
			return nil, response.Error(err)
		}
		result := payslipResponse(slip)
		result.Pdf = slip.PDF
		return result, nil
	}
	slip, err := h.service.GetMyPayslip(ctx, request.GetPayrollResultId())
	if err != nil {
		return nil, response.Error(err)
	}
	return payslipResponse(slip), nil
}

func payslipResponse(slip service.PublishedPayslip) *pb.GetMyPayslipResponse {
	return &pb.GetMyPayslipResponse{
		PayrollResultId: slip.ResultID, DocumentVersionId: slip.DocumentVersionID,
		PublishedAt: timestamppb.New(slip.PublishedAt), ScheduledPayDate: slip.ScheduledPayDate.Format(time.DateOnly),
		PaymentStatus:   slip.PaymentStatus,
		PayrollPeriodId: slip.PayrollPeriodID, PeriodCode: slip.PeriodCode,
		PeriodStart: slip.PeriodStart.Format(time.DateOnly), PeriodEnd: slip.PeriodEnd.Format(time.DateOnly),
	}
}

// ListMyPayslips mengembalikan halaman slip milik pekerja yang terautentikasi.
func (h *Handler) ListMyPayslips(ctx context.Context, request *pb.ListMyPayslipsRequest) (*pb.ListMyPayslipsResponse, error) {
	page, err := h.service.ListMyPayslips(ctx, request.GetPayrollPeriodId(), int(request.GetLimit()), int(request.GetOffset()))
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.GetMyPayslipResponse, 0, len(page.Items))
	for _, slip := range page.Items {
		items = append(items, payslipResponse(slip))
	}
	return &pb.ListMyPayslipsResponse{
		Items: items,
		Page:  &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)},
	}, nil
}

// ReissueEmployeePayslip membuka PDF untuk petugas HR dengan cakupan yang sesuai.
func (h *Handler) ReissueEmployeePayslip(ctx context.Context, request *pb.ReissueEmployeePayslipRequest) (*pb.ReissueEmployeePayslipResponse, error) {
	pdf, err := h.service.ReissueEmployeePayslip(ctx, request.GetWorkerId(), request.GetPayrollResultId())
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ReissueEmployeePayslipResponse{DocumentVersionId: pdf.DocumentVersionID, Pdf: pdf.Content}, nil
}
