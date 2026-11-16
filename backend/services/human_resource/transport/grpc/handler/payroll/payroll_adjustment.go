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

func payrollAdjustment(v *entity.PayrollAdjustment) *pb.PayrollAdjustment {
	return &pb.PayrollAdjustment{
		Id: v.ID, WorkerId: v.WorkerID, PayrollPeriodId: v.PayrollPeriodID, ComponentId: v.ComponentID, AdjustmentType: v.AdjustmentType,
		Amount: string(v.Amount), CurrencyCode: v.CurrencyCode, Reason: v.Reason, SourceReference: v.SourceReference, RequestedByPrincipalId: v.RequestedByPrincipalID,
		ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamp(v.ApprovedAt), IdempotencyKey: v.IdempotencyKey, CreatedAt: timestamppb.New(v.CreatedAt),
	}
}
func payrollAdjustments(page operation.Page[entity.PayrollAdjustment], err error) (*pb.PayrollAdjustmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollAdjustment, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollAdjustment(&page.Items[i]))
	}
	return &pb.PayrollAdjustmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollAdjustment(ctx context.Context, r *pb.CreatePayrollAdjustmentRequest) (*pb.PayrollAdjustmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreatePayrollAdjustment(ctx, service.CreatePayrollAdjustmentInput{
		WorkerID: in.GetWorkerId(), PayrollPeriodID: in.GetPayrollPeriodId(), ComponentID: in.GetComponentId(), AdjustmentType: in.GetAdjustmentType(),
		Amount: entity.Decimal(in.GetAmount()), CurrencyCode: in.GetCurrencyCode(), Reason: in.GetReason(), SourceReference: in.SourceReference,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, IdempotencyKey: in.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollAdjustmentResponse{Item: payrollAdjustment(v)}, nil
}
func (h *Handler) UpdatePayrollAdjustment(ctx context.Context, r *pb.UpdatePayrollAdjustmentRequest) (*pb.PayrollAdjustmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "worker_id", "payroll_period_id", "component_id", "adjustment_type", "amount", "currency_code", "reason", "source_reference", "approved_by_principal_id", "approved_at", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdatePayrollAdjustment(ctx, r.Id, service.UpdatePayrollAdjustmentInput{
		WorkerID: field(m["worker_id"], in.WorkerId), PayrollPeriodID: field(m["payroll_period_id"], in.PayrollPeriodId), ComponentID: field(m["component_id"], in.ComponentId),
		AdjustmentType: field(m["adjustment_type"], in.AdjustmentType), Amount: field(m["amount"], decimal(in.Amount)), CurrencyCode: field(m["currency_code"], in.CurrencyCode),
		Reason: field(m["reason"], in.Reason), SourceReference: field(m["source_reference"], in.SourceReference), ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId),
		ApprovedAt: approved, IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollAdjustmentResponse{Item: payrollAdjustment(v)}, nil
}
func (h *Handler) DeletePayrollAdjustment(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollAdjustment(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollAdjustmentByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollAdjustmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollAdjustmentByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollAdjustmentResponse{Item: payrollAdjustment(v)}, nil
}
func (h *Handler) GetAllPayrollAdjustments(ctx context.Context, r *common.PageRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.GetAllPayrollAdjustments(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByWorkerID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByPayrollPeriodID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByPayrollPeriodID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByComponentID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByComponentID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByAdjustmentType(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByAdjustmentType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByRequestedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByRequestedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollAdjustmentByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollAdjustments(h.service.FindPayrollAdjustmentByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
