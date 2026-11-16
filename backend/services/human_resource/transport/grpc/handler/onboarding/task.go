package onboarding

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/onboarding/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func task(value *entity.OnboardingTask) *pb.OnboardingTask {
	item := &pb.OnboardingTask{Id: value.ID, OnboardingPlanId: value.OnboardingPlanID, TaskCode: value.TaskCode, TaskName: value.TaskName,
		TaskOwnerType: value.TaskOwnerType, AssignedPrincipalId: value.AssignedPrincipalID, CompletionEvidenceReferenceId: value.CompletionEvidenceReferenceID, IsMandatory: value.IsMandatory}
	if value.DueAt != nil {
		item.DueAt = timestamppb.New(*value.DueAt)
	}
	if value.CompletedAt != nil {
		item.CompletedAt = timestamppb.New(*value.CompletedAt)
	}
	return item
}
func taskResponse(value *entity.OnboardingTask, err error) (*pb.OnboardingTaskResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.OnboardingTaskResponse{Item: task(value)}, nil
}
func tasks(page operation.Page[entity.OnboardingTask], err error) (*pb.OnboardingTasksResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.OnboardingTask, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, task(&page.Items[index]))
	}
	return &pb.OnboardingTasksResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateOnboardingTask(ctx context.Context, request *pb.CreateOnboardingTaskRequest) (*pb.OnboardingTaskResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return taskResponse(h.service.CreateOnboardingTask(ctx, service.CreateOnboardingTaskInput{OnboardingPlanID: in.GetOnboardingPlanId(), TaskCode: in.GetTaskCode(),
		TaskName: in.GetTaskName(), TaskOwnerType: in.GetTaskOwnerType(), AssignedPrincipalID: in.AssignedPrincipalId, DueAt: due, CompletedAt: completed,
		CompletionEvidenceReferenceID: in.CompletionEvidenceReferenceId, IsMandatory: in.IsMandatory}))
}
func (h *Handler) UpdateOnboardingTask(ctx context.Context, request *pb.UpdateOnboardingTaskRequest) (*pb.OnboardingTaskResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "onboarding_plan_id", "task_code", "task_name", "task_owner_type", "assigned_principal_id", "due_at", "completed_at", "completion_evidence_reference_id", "is_mandatory")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	patch := service.UpdateOnboardingTaskInput{OnboardingPlanID: field(mask["onboarding_plan_id"], in.OnboardingPlanId), TaskCode: field(mask["task_code"], in.TaskCode),
		TaskName: field(mask["task_name"], in.TaskName), TaskOwnerType: field(mask["task_owner_type"], in.TaskOwnerType), AssignedPrincipalID: field(mask["assigned_principal_id"], in.AssignedPrincipalId),
		CompletionEvidenceReferenceID: field(mask["completion_evidence_reference_id"], in.CompletionEvidenceReferenceId), IsMandatory: field(mask["is_mandatory"], in.IsMandatory)}
	if mask["due_at"] {
		value, err := response.OptionalTimestamp(in.DueAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.DueAt = field(true, value)
	}
	if mask["completed_at"] {
		value, err := response.OptionalTimestamp(in.CompletedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.CompletedAt = field(true, value)
	}
	return taskResponse(h.service.UpdateOnboardingTask(ctx, request.Id, patch))
}
func (h *Handler) DeleteOnboardingTask(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteOnboardingTask(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetOnboardingTaskByID(ctx context.Context, request *common.IDRequest) (*pb.OnboardingTaskResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return taskResponse(h.service.GetOnboardingTaskByID(ctx, request.Id))
}
func (h *Handler) GetAllOnboardingTasks(ctx context.Context, request *common.PageRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.GetAllOnboardingTasks(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingTaskByOnboardingPlanID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.FindOnboardingTaskByOnboardingPlanID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingTaskByTaskCode(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.FindOnboardingTaskByTaskCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingTaskByTaskOwnerType(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.FindOnboardingTaskByTaskOwnerType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingTaskByAssignedPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.FindOnboardingTaskByAssignedPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingTaskByCompletionEvidenceReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(h.service.FindOnboardingTaskByCompletionEvidenceReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
