package human_resource

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SubmitWorkerAssignmentTransfer mencatat permintaan transfer dengan kunci idempotensi.
func (a *Adapter) SubmitWorkerAssignmentTransfer(ctx context.Context, input model.SubmitWorkerAssignmentTransferInput) (*model.WorkerAssignmentTransfer, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignments.create")
	if err != nil {
		return nil, err
	}
	workerID, workerErr := uuid.Parse(strings.TrimPrefix(input.WorkerID, workerGIDPrefix))
	sourceID, sourceErr := parseNodeID(input.SourceAssignmentID, "worker-assignment")
	targetID, targetErr := parseNodeID(input.TargetPositionID, "position")
	if workerErr != nil || sourceErr != nil || targetErr != nil || workerID == uuid.Nil || sourceID == uuid.Nil || targetID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment transfer references", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.SubmitWorkerAssignmentTransfer(rpcCtx, &personnel.SubmitWorkerAssignmentTransferRequest{
		WorkerId: workerID.String(), SourceAssignmentId: sourceID.String(), TargetPositionId: targetID.String(),
		EffectiveDate: input.EffectiveDate, ReasonCode: input.ReasonCode, IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return nil, mapWorkerTransferError(err)
	}
	item := response.GetItem()
	if item.GetWorkerId() != workerID.String() || item.GetSourceAssignmentId() != sourceID.String() || item.GetTargetPositionId() != targetID.String() {
		return nil, middleware.ErrUnavailable("invalid human resource transfer response", nil)
	}
	return toModelWorkerTransfer(item)
}

// DecideWorkerAssignmentTransfer meneruskan keputusan yang diterapkan secara atomik oleh HRM.
func (a *Adapter) DecideWorkerAssignmentTransfer(ctx context.Context, rawID string, decision model.WorkerAssignmentTransferDecision) (*model.WorkerAssignmentTransfer, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.worker_assignments.transfer")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "worker-assignment-transfer")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment transfer ID", nil)
	}
	var rpcDecision personnel.TransferDecision
	var expectedStatus string
	switch decision {
	case model.WorkerAssignmentTransferDecisionApprove:
		rpcDecision, expectedStatus = personnel.TransferDecision_TRANSFER_DECISION_APPROVE, "applied"
	case model.WorkerAssignmentTransferDecisionReject:
		rpcDecision, expectedStatus = personnel.TransferDecision_TRANSFER_DECISION_REJECT, "rejected"
	default:
		return nil, middleware.ErrInvalidArgument("invalid worker assignment transfer decision", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.DecideWorkerAssignmentTransfer(rpcCtx, &personnel.DecideWorkerAssignmentTransferRequest{
		TransferId: id.String(), Decision: rpcDecision,
	})
	if err != nil {
		return nil, mapWorkerTransferError(err)
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource transfer response", nil)
	}
	return toModelWorkerTransfer(item)
}

func mapWorkerTransferError(err error) error {
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return middleware.ErrConflict("worker assignment transfer conflicts with existing state", err)
	default:
		return mapReadError(err, "worker assignment transfer")
	}
}

func toModelWorkerTransfer(item *personnel.WorkerAssignmentTransfer) (*model.WorkerAssignmentTransfer, error) {
	for _, raw := range []string{item.GetId(), item.GetWorkerId(), item.GetSourceAssignmentId(), item.GetTargetPositionId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource transfer reference", nil)
		}
	}
	switch item.GetStatus() {
	case "applied":
		if id, err := uuid.Parse(item.GetAppliedAssignmentId()); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource applied assignment", nil)
		}
	case "submitted", "rejected":
		if item.AppliedAssignmentId != nil {
			return nil, middleware.ErrUnavailable("invalid human resource transfer response", nil)
		}
	default:
		return nil, middleware.ErrUnavailable("invalid human resource transfer status", nil)
	}
	requested := item.GetRequestedAt()
	if requested.CheckValid() != nil || strings.TrimSpace(item.GetReasonCode()) == "" {
		return nil, middleware.ErrUnavailable("invalid human resource transfer metadata", nil)
	}
	var decidedAt *time.Time
	decided := item.GetDecidedAt()
	if item.GetStatus() == "submitted" {
		if decided != nil {
			return nil, middleware.ErrUnavailable("invalid human resource transfer decision time", nil)
		}
	} else {
		if decided.CheckValid() != nil || decided.AsTime().Before(requested.AsTime()) {
			return nil, middleware.ErrUnavailable("invalid human resource transfer decision time", nil)
		}
		at := decided.AsTime()
		decidedAt = &at
	}

	return &model.WorkerAssignmentTransfer{
		ID: nodeID("worker-assignment-transfer", item.GetId()), RawID: item.GetId(), WorkerID: workerGIDPrefix + item.GetWorkerId(), SourceAssignmentID: item.GetSourceAssignmentId(),
		TargetPositionID: item.GetTargetPositionId(), EffectiveDate: item.GetEffectiveDate(),
		Status: item.GetStatus(), AppliedAssignmentID: item.AppliedAssignmentId,
		ReasonCode: item.GetReasonCode(), RequestedAt: requested.AsTime(), DecidedAt: decidedAt,
	}, nil
}

// WorkerAssignmentTransfer membaca status dan referensi transfer.
func (a *Adapter) WorkerAssignmentTransfer(ctx context.Context, rawID string) (*model.WorkerAssignmentTransfer, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignment_transfers.read")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "worker-assignment-transfer")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment transfer ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.GetWorkerAssignmentTransfer(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "worker assignment transfer")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource transfer response", nil)
	}
	return toModelWorkerTransfer(item)
}

// WorkerAssignmentTransfers membaca riwayat transfer pekerja.
func (a *Adapter) WorkerAssignmentTransfers(ctx context.Context, rawWorkerID string, limit, offset int) (*Page[model.WorkerAssignmentTransfer], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignment_transfers.read")
	if err != nil {
		return nil, err
	}
	workerID, err := parseNodeID(rawWorkerID, "hrm-worker")
	if err != nil || workerID == uuid.Nil || !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment transfer history arguments", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.ListWorkerAssignmentTransfers(rpcCtx, &personnel.ListWorkerAssignmentTransfersRequest{
		WorkerId: workerID.String(), Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, mapReadError(err, "worker assignment transfer history")
	}
	items, meta := response.GetItems(), response.GetPage()
	if !validPage(meta, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource transfer history page", nil)
	}
	page := &Page[model.WorkerAssignmentTransfer]{Items: make([]*model.WorkerAssignmentTransfer, 0, len(items)), TotalCount: int(meta.GetTotal()), Limit: limit, Offset: offset}
	for _, item := range items {
		if item.GetWorkerId() != workerID.String() {
			return nil, middleware.ErrUnavailable("invalid worker in transfer history", nil)
		}
		node, err := toModelWorkerTransfer(item)
		if err != nil {
			return nil, err
		}
		page.Items = append(page.Items, node)
	}
	return page, nil
}
