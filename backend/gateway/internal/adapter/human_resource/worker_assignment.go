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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ApproveWorkerAssignment meneruskan persetujuan assignment ke HRM.
func (a *Adapter) ApproveWorkerAssignment(ctx context.Context, rawID string) (*model.WorkerAssignment, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.worker_assignments.approve")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "worker-assignment")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.ApproveWorkerAssignment(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("worker assignment does not meet approval requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "worker assignment")
	}
	item := response.GetItem()
	approvedAt := item.GetApprovedAt()
	if item.GetId() != id.String() || approvedAt.CheckValid() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker assignment response", nil)
	}
	if approver, err := uuid.Parse(item.GetApprovedByPrincipalId()); err != nil || approver == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker assignment approver", nil)
	}
	return toModelWorkerAssignment(item)
}

// WorkerAssignment membaca rincian assignment melalui izin baca khusus.
func (a *Adapter) WorkerAssignment(ctx context.Context, rawID string) (*model.WorkerAssignment, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignments.read")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "worker-assignment")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignment ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.GetWorkerAssignmentByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "worker assignment")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource worker assignment response", nil)
	}
	return toModelWorkerAssignment(item)
}

// WorkerAssignments membaca halaman assignment untuk satu pekerja atau posisi.
func (a *Adapter) WorkerAssignments(ctx context.Context, rawWorkerID, rawPositionID *string, limit, offset int) (*Page[model.WorkerAssignment], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignments.read")
	if err != nil {
		return nil, err
	}
	if (rawWorkerID == nil) == (rawPositionID == nil) || !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("provide one worker assignment filter and valid pagination", nil)
	}
	rawID, resource := "", "position"
	if rawWorkerID != nil {
		rawID, resource = *rawWorkerID, "hrm-worker"
	} else {
		rawID = *rawPositionID
	}
	id, err := parseNodeID(rawID, resource)
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker assignments filter ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	filterID := id.String()
	request := &common.StringFinderRequest{Value: &filterID, Limit: int32(limit), Offset: int32(offset)}
	var response *personnel.WorkerAssignmentsResponse
	if rawWorkerID != nil {
		response, err = a.personnel.FindWorkerAssignmentByWorkerID(rpcCtx, request)
	} else {
		response, err = a.personnel.FindWorkerAssignmentByPositionID(rpcCtx, request)
	}
	if err != nil {
		return nil, mapReadError(err, "worker assignment")
	}
	items, page := response.GetItems(), response.GetPage()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource worker assignment page", nil)
	}
	result := &Page[model.WorkerAssignment]{
		Items: make([]*model.WorkerAssignment, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, item := range items {
		if (rawWorkerID != nil && item.GetWorkerId() != filterID) || (rawPositionID != nil && item.GetPositionId() != filterID) {
			return nil, middleware.ErrUnavailable("invalid human resource worker assignment reference", nil)
		}
		mapped, err := toModelWorkerAssignment(item)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, mapped)
	}
	return result, nil
}

func toModelWorkerAssignment(item *personnel.WorkerAssignment) (*model.WorkerAssignment, error) {
	for _, raw := range []string{item.GetId(), item.GetWorkerId(), item.GetPositionId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource worker assignment reference", nil)
		}
	}
	result := &model.WorkerAssignment{
		ID: nodeID("worker-assignment", item.GetId()), RawID: item.GetId(), WorkerID: workerGIDPrefix + item.GetWorkerId(), PositionID: item.GetPositionId(),
		AssignmentType: item.GetAssignmentType(), FtePercent: item.GetFtePercent(),
		ValidFrom: item.GetValidFrom(), ValidTo: item.ValidTo,
	}
	for _, field := range []struct {
		value  *timestamppb.Timestamp
		target **time.Time
	}{
		{item.RequestedAt, &result.RequestedAt},
		{item.ApprovedAt, &result.ApprovedAt},
		{item.CancelledAt, &result.CancelledAt},
	} {
		if field.value == nil {
			continue
		}
		if field.value.CheckValid() != nil {
			return nil, middleware.ErrUnavailable("invalid human resource worker assignment timestamp", nil)
		}
		value := field.value.AsTime()
		*field.target = &value
	}
	return result, nil
}

// CreateWorkerAssignment mencatat pengajuan assignment untuk diperiksa HRM.
func (a *Adapter) CreateWorkerAssignment(ctx context.Context, input model.CreateWorkerAssignmentInput) (*model.WorkerAssignment, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.worker_assignments.create")
	if err != nil {
		return nil, err
	}
	worker, err := uuid.Parse(strings.TrimPrefix(input.WorkerID, workerGIDPrefix))
	if err != nil || worker == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	position, err := parseNodeID(input.PositionID, "position")
	if err != nil || position == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid position ID", nil)
	}
	var managerID *string
	if input.ManagerWorkerID != nil {
		manager, err := uuid.Parse(strings.TrimPrefix(*input.ManagerWorkerID, workerGIDPrefix))
		if err != nil || manager == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid manager worker ID", nil)
		}
		id := manager.String()
		managerID = &id
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	workerID, positionID := worker.String(), position.String()
	response, err := a.personnel.CreateWorkerAssignment(rpcCtx, &personnel.CreateWorkerAssignmentRequest{
		Input: &personnel.WorkerAssignmentInput{
			WorkerId: &workerID, PositionId: &positionID, ManagerWorkerId: managerID,
			AssignmentType: &input.AssignmentType, FtePercent: input.FtePercent,
			ValidFrom: &input.ValidFrom, ValidTo: input.ValidTo, AssignmentReason: input.AssignmentReason,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("worker assignment conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "worker assignment")
	}
	item := response.GetItem()
	if item.GetWorkerId() != workerID || item.GetPositionId() != positionID || item.GetRequestedAt().CheckValid() != nil ||
		item.GetApprovedAt() != nil || item.GetApprovedByPrincipalId() != "" || item.GetCancelledAt() != nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker assignment response", nil)
	}
	return toModelWorkerAssignment(item)
}
