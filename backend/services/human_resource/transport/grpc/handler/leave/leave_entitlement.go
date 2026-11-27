package leave

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func leaveEntitlement(v *entity.LeaveEntitlement) *pb.LeaveEntitlement {
	return &pb.LeaveEntitlement{
		Id: v.ID, WorkerId: v.WorkerID, LeavePolicyId: v.LeavePolicyID, LeaveTypeCodeSnapshot: v.LeaveTypeCodeSnapshot,
		EntitlementPeriodStart: v.EntitlementPeriodStart.Format("2006-01-02"), EntitlementPeriodEnd: v.EntitlementPeriodEnd.Format("2006-01-02"),
		OpeningBalance: string(v.OpeningBalance), AccruedQuantity: string(v.AccruedQuantity), UsedQuantitySnapshot: string(v.UsedQuantitySnapshot),
		AdjustedQuantity: string(v.AdjustedQuantity), ClosingBalanceSnapshot: string(v.ClosingBalanceSnapshot), UomCode: v.UOMCode,
		ClosedAt: timestamp(v.ClosedAt), ClosedByPrincipalId: v.ClosedByPrincipalID,
		Revision: v.Revision, CalculatedAt: timestamppb.New(v.CalculatedAt), SourceWorkforceSnapshotReferenceId: v.SourceWorkforceSnapshotReferenceID,
	}
}
func leaveEntitlements(page operation.Page[entity.LeaveEntitlement], err error) (*pb.LeaveEntitlementsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.LeaveEntitlement, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, leaveEntitlement(&page.Items[i]))
	}
	return &pb.LeaveEntitlementsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateLeaveEntitlement(ctx context.Context, r *pb.CreateLeaveEntitlementRequest) (*pb.LeaveEntitlementResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := response.Date(in.GetEntitlementPeriodStart())
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Date(in.GetEntitlementPeriodEnd())
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreateLeaveEntitlement(ctx, service.CreateLeaveEntitlementInput{
		WorkerID: in.GetWorkerId(), LeavePolicyID: in.GetLeavePolicyId(), LeaveTypeCodeSnapshot: in.GetLeaveTypeCodeSnapshot(),
		EntitlementPeriodStart: start, EntitlementPeriodEnd: end, OpeningBalance: decimal(in.OpeningBalance), AccruedQuantity: decimal(in.AccruedQuantity),
		UsedQuantitySnapshot: decimal(in.UsedQuantitySnapshot), AdjustedQuantity: decimal(in.AdjustedQuantity), ClosingBalanceSnapshot: decimal(in.ClosingBalanceSnapshot),
		UOMCode: in.UomCode, Revision: in.Revision, SourceWorkforceSnapshotReferenceID: in.SourceWorkforceSnapshotReferenceId,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveEntitlementResponse{Item: leaveEntitlement(v)}, nil
}
func (h *Handler) UpdateLeaveEntitlement(ctx context.Context, r *pb.UpdateLeaveEntitlementRequest) (*pb.LeaveEntitlementResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "worker_id", "leave_policy_id", "leave_type_code_snapshot", "entitlement_period_start", "entitlement_period_end", "opening_balance", "accrued_quantity", "used_quantity_snapshot", "adjusted_quantity", "closing_balance_snapshot", "uom_code", "revision", "source_workforce_snapshot_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := dateField(m["entitlement_period_start"], in.EntitlementPeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := dateField(m["entitlement_period_end"], in.EntitlementPeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdateLeaveEntitlement(ctx, r.Id, service.UpdateLeaveEntitlementInput{
		WorkerID: field(m["worker_id"], in.WorkerId), LeavePolicyID: field(m["leave_policy_id"], in.LeavePolicyId),
		LeaveTypeCodeSnapshot: field(m["leave_type_code_snapshot"], in.LeaveTypeCodeSnapshot), EntitlementPeriodStart: start, EntitlementPeriodEnd: end,
		OpeningBalance: field(m["opening_balance"], decimal(in.OpeningBalance)), AccruedQuantity: field(m["accrued_quantity"], decimal(in.AccruedQuantity)),
		UsedQuantitySnapshot: field(m["used_quantity_snapshot"], decimal(in.UsedQuantitySnapshot)), AdjustedQuantity: field(m["adjusted_quantity"], decimal(in.AdjustedQuantity)),
		ClosingBalanceSnapshot: field(m["closing_balance_snapshot"], decimal(in.ClosingBalanceSnapshot)), UOMCode: field(m["uom_code"], in.UomCode),
		Revision: field(m["revision"], in.Revision), SourceWorkforceSnapshotReferenceID: field(m["source_workforce_snapshot_reference_id"], in.SourceWorkforceSnapshotReferenceId),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveEntitlementResponse{Item: leaveEntitlement(v)}, nil
}
func (h *Handler) DeleteLeaveEntitlement(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteLeaveEntitlement(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetLeaveEntitlementByID(ctx context.Context, r *common.IDRequest) (*pb.LeaveEntitlementResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetLeaveEntitlementByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveEntitlementResponse{Item: leaveEntitlement(v)}, nil
}
func (h *Handler) GetAllLeaveEntitlements(ctx context.Context, r *common.PageRequest) (*pb.LeaveEntitlementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveEntitlements(h.service.GetAllLeaveEntitlements(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveEntitlementByWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveEntitlementsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveEntitlements(h.service.FindLeaveEntitlementByWorkerID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveEntitlementByLeavePolicyID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveEntitlementsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveEntitlements(h.service.FindLeaveEntitlementByLeavePolicyID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveEntitlementByUOMCode(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveEntitlementsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveEntitlements(h.service.FindLeaveEntitlementByUOMCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.LeaveEntitlementsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return leaveEntitlements(h.service.FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}

func (h *Handler) CloseLeaveEntitlement(ctx context.Context, request *common.IDRequest) (*pb.LeaveEntitlementResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	closed, err := h.service.CloseLeaveEntitlement(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.LeaveEntitlementResponse{Item: leaveEntitlement(closed)}, nil
}
