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

func plan(value *entity.OnboardingPlan) *pb.OnboardingPlan {
	item := &pb.OnboardingPlan{Id: value.ID, WorkerId: value.WorkerID, PlanCode: value.PlanCode, PlanVersion: value.PlanVersion,
		StartDate: value.StartDate.Format("2006-01-02"), TargetCompletionDate: dateString(value.TargetCompletionDate),
		Status: value.Status, TemplateReferenceId: value.TemplateReferenceID, CreatedByPrincipalId: value.CreatedByPrincipalID,
		StartedByPrincipalId: value.StartedByPrincipalID, CompletedByPrincipalId: value.CompletedByPrincipalID}
	if value.StartedAt != nil {
		item.StartedAt = timestamppb.New(*value.StartedAt)
	}
	if value.CompletedAt != nil {
		item.CompletedAt = timestamppb.New(*value.CompletedAt)
	}
	return item
}
func (h *Handler) StartOnboardingPlan(ctx context.Context, request *common.IDRequest) (*pb.OnboardingPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(h.service.StartOnboardingPlan(ctx, request.Id))
}
func planResponse(value *entity.OnboardingPlan, err error) (*pb.OnboardingPlanResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.OnboardingPlanResponse{Item: plan(value)}, nil
}
func plans(page operation.Page[entity.OnboardingPlan], err error) (*pb.OnboardingPlansResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.OnboardingPlan, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, plan(&page.Items[index]))
	}
	return &pb.OnboardingPlansResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateOnboardingPlan(ctx context.Context, request *pb.CreateOnboardingPlanRequest) (*pb.OnboardingPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	start, err := response.Date(in.GetStartDate())
	if err != nil {
		return nil, response.Error(err)
	}
	completion, err := optionalDate(in.TargetCompletionDate)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResponse(h.service.CreateOnboardingPlan(ctx, service.CreateOnboardingPlanInput{WorkerID: in.GetWorkerId(), PlanCode: in.GetPlanCode(),
		PlanVersion: in.PlanVersion, StartDate: start, TargetCompletionDate: completion, Status: in.Status, TemplateReferenceID: in.TemplateReferenceId}))
}
func (h *Handler) UpdateOnboardingPlan(ctx context.Context, request *pb.UpdateOnboardingPlanRequest) (*pb.OnboardingPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "plan_code", "plan_version", "start_date", "target_completion_date", "status", "template_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	start, err := dateField(mask["start_date"], in.StartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	completion, err := dateField(mask["target_completion_date"], in.TargetCompletionDate)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResponse(h.service.UpdateOnboardingPlan(ctx, request.Id, service.UpdateOnboardingPlanInput{
		WorkerID: field(mask["worker_id"], in.WorkerId), PlanCode: field(mask["plan_code"], in.PlanCode), PlanVersion: field(mask["plan_version"], in.PlanVersion),
		StartDate: start, TargetCompletionDate: completion, Status: field(mask["status"], in.Status), TemplateReferenceID: field(mask["template_reference_id"], in.TemplateReferenceId),
	}))
}
func (h *Handler) DeleteOnboardingPlan(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteOnboardingPlan(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetOnboardingPlanByID(ctx context.Context, request *common.IDRequest) (*pb.OnboardingPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(h.service.GetOnboardingPlanByID(ctx, request.Id))
}
func (h *Handler) GetAllOnboardingPlans(ctx context.Context, request *common.PageRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.GetAllOnboardingPlans(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingPlanByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindOnboardingPlanByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingPlanByPlanCode(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindOnboardingPlanByPlanCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingPlanByStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindOnboardingPlanByStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingPlanByTemplateReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindOnboardingPlanByTemplateReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindOnboardingPlanByCreatedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.OnboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindOnboardingPlanByCreatedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
