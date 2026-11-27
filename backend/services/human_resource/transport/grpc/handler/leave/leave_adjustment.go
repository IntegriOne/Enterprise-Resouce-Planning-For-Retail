package leave

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func leaveAdjustment(v *entity.LeaveAdjustment) *pb.LeaveAdjustment {
	return &pb.LeaveAdjustment{
		Id: v.ID, EntitlementId: v.EntitlementID, QuantityDelta: string(v.QuantityDelta), ReasonCode: v.ReasonCode, Reason: v.Reason,
		RequestedByPrincipalId: v.RequestedByPrincipalID, ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamp(v.ApprovedAt),
		CreatedAt: timestamppb.New(v.CreatedAt), IdempotencyKey: v.IdempotencyKey,
		RejectedAt: timestamp(v.RejectedAt), RejectedByPrincipalId: v.RejectedByPrincipalID, RejectionReasonCode: v.RejectionReasonCode,
	}
}
func leaveAdjustments(page operation.Page[entity.LeaveAdjustment], err error) (*pb.LeaveAdjustmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LeaveAdjustment, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, leaveAdjustment(&page.Items[i]))
	}
	return &pb.LeaveAdjustmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLeaveAdjustment(ctx context.Context, r *pb.CreateLeaveAdjustmentRequest) (*pb.LeaveAdjustmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreateLeaveAdjustment(ctx, service.CreateLeaveAdjustmentInput{
		EntitlementID: in.GetEntitlementId(), QuantityDelta: entity.Decimal(in.GetQuantityDelta()), ReasonCode: in.GetReasonCode(), Reason: in.GetReason(),
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, IdempotencyKey: in.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveAdjustmentResponse{Item: leaveAdjustment(v)}, nil
}
func (h *Handler) UpdateLeaveAdjustment(ctx context.Context, r *pb.UpdateLeaveAdjustmentRequest) (*pb.LeaveAdjustmentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "entitlement_id", "quantity_delta", "reason_code", "reason", "approved_by_principal_id", "approved_at", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	var approved operation.Field[time.Time]
	if m["approved_at"] {
		value, err := response.OptionalTimestamp(in.ApprovedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		approved = field(true, value)
	}
	v, err := h.service.UpdateLeaveAdjustment(ctx, r.Id, service.UpdateLeaveAdjustmentInput{
		EntitlementID: field(m["entitlement_id"], in.EntitlementId), QuantityDelta: field(m["quantity_delta"], decimal(in.QuantityDelta)),
		ReasonCode: field(m["reason_code"], in.ReasonCode), Reason: field(m["reason"], in.Reason),
		ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), ApprovedAt: approved, IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveAdjustmentResponse{Item: leaveAdjustment(v)}, nil
}
func (h *Handler) DeleteLeaveAdjustment(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLeaveAdjustment(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}

func (h *Handler) ApproveLeaveAdjustment(ctx context.Context, r *common.IDRequest) (*pb.LeaveAdjustmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	approved, err := h.service.ApproveLeaveAdjustment(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveAdjustmentResponse{Item: leaveAdjustment(approved)}, nil
}
func (h *Handler) GetLeaveAdjustmentByID(ctx context.Context, r *common.IDRequest) (*pb.LeaveAdjustmentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetLeaveAdjustmentByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveAdjustmentResponse{Item: leaveAdjustment(v)}, nil
}
func (h *Handler) GetAllLeaveAdjustments(ctx context.Context, r *common.PageRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.GetAllLeaveAdjustments(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveAdjustmentByEntitlementID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.FindLeaveAdjustmentByEntitlementID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveAdjustmentByReasonCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.FindLeaveAdjustmentByReasonCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveAdjustmentByRequestedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.FindLeaveAdjustmentByRequestedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveAdjustmentByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.FindLeaveAdjustmentByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveAdjustmentByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveAdjustmentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveAdjustments(h.service.FindLeaveAdjustmentByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}

func (h *Handler) RejectLeaveAdjustment(ctx context.Context, request *pb.RejectLeaveAdjustmentRequest) (*pb.LeaveAdjustmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	rejected, err := h.service.RejectLeaveAdjustment(ctx, request.Id, request.ReasonCode)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveAdjustmentResponse{Item: leaveAdjustment(rejected)}, nil
}
