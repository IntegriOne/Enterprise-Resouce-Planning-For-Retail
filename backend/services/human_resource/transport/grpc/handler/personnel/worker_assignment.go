package personnel

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func assignment(value *entity.WorkerAssignment) *pb.WorkerAssignment {
	item := &pb.WorkerAssignment{Id: value.ID, WorkerId: value.WorkerID, PositionId: value.PositionID, ManagerWorkerId: value.ManagerWorkerID,
		AssignmentType: value.AssignmentType, EmploymentStatusSnapshot: value.EmploymentStatusSnapshot, FtePercent: string(value.FTEPercent),
		ValidFrom: value.ValidFrom.Format("2006-01-02"), ValidTo: dateString(value.ValidTo), AssignmentReason: value.AssignmentReason, ApprovedByPrincipalId: value.ApprovedByPrincipalID}
	if value.ApprovedAt != nil {
		item.ApprovedAt = timestamppb.New(*value.ApprovedAt)
	}
	return item
}
func assignmentResponse(value *entity.WorkerAssignment, err error) (*pb.WorkerAssignmentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerAssignmentResponse{Item: assignment(value)}, nil
}
func assignments(page operation.Page[entity.WorkerAssignment], err error) (*pb.WorkerAssignmentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerAssignment, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, assignment(&page.Items[index]))
	}
	return &pb.WorkerAssignmentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkerAssignment(ctx context.Context, request *pb.CreateWorkerAssignmentRequest) (*pb.WorkerAssignmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetValidFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return assignmentResponse(h.service.CreateWorkerAssignment(ctx, service.CreateWorkerAssignmentInput{WorkerID: in.GetWorkerId(), PositionID: in.GetPositionId(),
		ManagerWorkerID: in.ManagerWorkerId, AssignmentType: in.GetAssignmentType(), EmploymentStatusSnapshot: in.EmploymentStatusSnapshot,
		FTEPercent: decimalPointer(in.FtePercent), ValidFrom: from, ValidTo: to, AssignmentReason: in.AssignmentReason, ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved}))
}
func (h *Handler) UpdateWorkerAssignment(ctx context.Context, request *pb.UpdateWorkerAssignmentRequest) (*pb.WorkerAssignmentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "position_id", "manager_worker_id", "assignment_type", "employment_status_snapshot", "fte_percent", "valid_from", "valid_to", "assignment_reason", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateWorkerAssignmentInput{WorkerID: field(mask["worker_id"], in.WorkerId), PositionID: field(mask["position_id"], in.PositionId),
		ManagerWorkerID: field(mask["manager_worker_id"], in.ManagerWorkerId), AssignmentType: field(mask["assignment_type"], in.AssignmentType),
		EmploymentStatusSnapshot: field(mask["employment_status_snapshot"], in.EmploymentStatusSnapshot), FTEPercent: field(mask["fte_percent"], decimalPointer(in.FtePercent)),
		ValidFrom: from, ValidTo: to, AssignmentReason: field(mask["assignment_reason"], in.AssignmentReason), ApprovedByPrincipalID: field(mask["approved_by_principal_id"], in.ApprovedByPrincipalId)}
	if mask["approved_at"] {
		value, err := response.OptionalTimestamp(in.ApprovedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.ApprovedAt = field(true, value)
	}
	return assignmentResponse(h.service.UpdateWorkerAssignment(ctx, request.Id, patch))
}
func (h *Handler) DeleteWorkerAssignment(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkerAssignment(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerAssignmentByID(ctx context.Context, request *common.IDRequest) (*pb.WorkerAssignmentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignmentResponse(h.service.GetWorkerAssignmentByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkerAssignments(ctx context.Context, request *common.PageRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.GetAllWorkerAssignments(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerAssignmentByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.FindWorkerAssignmentByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerAssignmentByPositionID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.FindWorkerAssignmentByPositionID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerAssignmentByManagerWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.FindWorkerAssignmentByManagerWorkerID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerAssignmentByAssignmentType(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.FindWorkerAssignmentByAssignmentType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerAssignmentByApprovedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerAssignmentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return assignments(h.service.FindWorkerAssignmentByApprovedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
