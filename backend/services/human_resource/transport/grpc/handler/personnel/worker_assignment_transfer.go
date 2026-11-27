package personnel

import (
	"context"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SubmitWorkerAssignmentTransfer meneruskan pengajuan transfer ke service Personnel.
func (h *Handler) SubmitWorkerAssignmentTransfer(ctx context.Context, request *pb.SubmitWorkerAssignmentTransferRequest) (*pb.SubmitWorkerAssignmentTransferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	effectiveDate, err := time.Parse(time.DateOnly, request.EffectiveDate)
	if err != nil {
		return nil, response.Error(hrerrors.Invalid("effective_date", "must use YYYY-MM-DD"))
	}
	result, err := h.service.SubmitWorkerAssignmentTransfer(ctx, service.SubmitWorkerAssignmentTransferInput{
		WorkerID: request.WorkerId, SourceAssignmentID: request.SourceAssignmentId,
		TargetPositionID: request.TargetPositionId, EffectiveDate: effectiveDate,
		ReasonCode: request.ReasonCode, IdempotencyKey: request.IdempotencyKey,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SubmitWorkerAssignmentTransferResponse{Item: assignmentTransfer(result)}, nil
}

// DecideWorkerAssignmentTransfer memutuskan permintaan transfer Personnel.
func (h *Handler) DecideWorkerAssignmentTransfer(ctx context.Context, request *pb.DecideWorkerAssignmentTransferRequest) (*pb.DecideWorkerAssignmentTransferResponse, error) {
	if request == nil || (request.Decision != pb.TransferDecision_TRANSFER_DECISION_APPROVE &&
		request.Decision != pb.TransferDecision_TRANSFER_DECISION_REJECT) {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	result, err := h.service.DecideWorkerAssignmentTransfer(ctx, request.TransferId,
		request.Decision == pb.TransferDecision_TRANSFER_DECISION_APPROVE)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.DecideWorkerAssignmentTransferResponse{Item: assignmentTransfer(result)}, nil
}

// GetWorkerAssignmentTransfer membaca detail transfer dari service Personnel.
func (h *Handler) GetWorkerAssignmentTransfer(ctx context.Context, request *common.IDRequest) (*pb.GetWorkerAssignmentTransferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	result, err := h.service.GetWorkerAssignmentTransfer(ctx, request.GetId())
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.GetWorkerAssignmentTransferResponse{Item: assignmentTransfer(result)}, nil
}

// ListWorkerAssignmentTransfers membaca halaman transfer untuk satu pekerja.
func (h *Handler) ListWorkerAssignmentTransfers(ctx context.Context, request *pb.ListWorkerAssignmentTransfersRequest) (*pb.ListWorkerAssignmentTransfersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := h.service.ListWorkerAssignmentTransfers(ctx, request.GetWorkerId(), int(request.GetLimit()), int(request.GetOffset()))
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerAssignmentTransfer, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, assignmentTransfer(&item))
	}
	return &pb.ListWorkerAssignmentTransfersResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func assignmentTransfer(result *entity.WorkerAssignmentTransfer) *pb.WorkerAssignmentTransfer {
	item := &pb.WorkerAssignmentTransfer{
		Id: result.ID, WorkerId: result.WorkerID, SourceAssignmentId: result.SourceAssignmentID,
		TargetPositionId: result.TargetPositionID, EffectiveDate: result.EffectiveDate.Format(time.DateOnly),
		ReasonCode: result.ReasonCode, Status: result.Status, RequestedByPrincipalId: result.RequestedByPrincipalID,
		RequestedAt: timestamppb.New(result.RequestedAt), DecidedByPrincipalId: result.DecidedByPrincipalID,
		AppliedAssignmentId: result.AppliedAssignmentID,
	}
	if result.DecidedAt != nil {
		item.DecidedAt = timestamppb.New(*result.DecidedAt)
	}
	return item
}
