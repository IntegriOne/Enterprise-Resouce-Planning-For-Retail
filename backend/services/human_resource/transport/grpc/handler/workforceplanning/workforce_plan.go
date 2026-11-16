package workforceplanning

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func workforcePlan(value *entity.WorkforcePlan) *pb.WorkforcePlan {
	item := &pb.WorkforcePlan{Id: value.ID, PlanCode: value.PlanCode, PlanVersion: value.PlanVersion, LegalEntityReferenceId: value.LegalEntityReferenceID,
		PeriodStart: value.PeriodStart.Format("2006-01-02"), PeriodEnd: value.PeriodEnd.Format("2006-01-02"), ScenarioCode: value.ScenarioCode,
		Status: value.Status, FinanceBudgetReferenceId: value.FinanceBudgetReferenceID, CreatedByPrincipalId: value.CreatedByPrincipalID,
		CreatedAt: timestamppb.New(value.CreatedAt), ApprovedByPrincipalId: value.ApprovedByPrincipalID}
	if value.ApprovedAt != nil {
		item.ApprovedAt = timestamppb.New(*value.ApprovedAt)
	}
	return item
}
func workforcePlanResponse(value *entity.WorkforcePlan, err error) (*pb.WorkforcePlanResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkforcePlanResponse{Item: workforcePlan(value)}, nil
}
func workforcePlans(page operation.Page[entity.WorkforcePlan], err error) (*pb.WorkforcePlansResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkforcePlan, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, workforcePlan(&page.Items[index]))
	}
	return &pb.WorkforcePlansResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkforcePlan(ctx context.Context, request *pb.CreateWorkforcePlanRequest) (*pb.WorkforcePlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	start, err := response.Date(in.GetPeriodStart())
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Date(in.GetPeriodEnd())
	if err != nil {
		return nil, response.Error(err)
	}
	approvedAt, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return workforcePlanResponse(h.service.CreateWorkforcePlan(ctx, service.CreateWorkforcePlanInput{
		PlanCode: in.GetPlanCode(), PlanVersion: in.GetPlanVersion(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		PeriodStart: start, PeriodEnd: end, ScenarioCode: in.ScenarioCode, Status: in.Status,
		FinanceBudgetReferenceID: in.FinanceBudgetReferenceId, ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approvedAt,
	}))
}
func (h *Handler) UpdateWorkforcePlan(ctx context.Context, request *pb.UpdateWorkforcePlanRequest) (*pb.WorkforcePlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "plan_code", "plan_version", "legal_entity_reference_id", "period_start", "period_end", "scenario_code", "status", "finance_budget_reference_id", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	start, err := dateField(mask["period_start"], in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := dateField(mask["period_end"], in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateWorkforcePlanInput{PlanCode: field(mask["plan_code"], in.PlanCode), PlanVersion: field(mask["plan_version"], in.PlanVersion),
		LegalEntityReferenceID: field(mask["legal_entity_reference_id"], in.LegalEntityReferenceId), PeriodStart: start, PeriodEnd: end,
		ScenarioCode: field(mask["scenario_code"], in.ScenarioCode), Status: field(mask["status"], in.Status),
		FinanceBudgetReferenceID: field(mask["finance_budget_reference_id"], in.FinanceBudgetReferenceId), ApprovedByPrincipalID: field(mask["approved_by_principal_id"], in.ApprovedByPrincipalId)}
	if mask["approved_at"] {
		value, err := response.OptionalTimestamp(in.ApprovedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.ApprovedAt = field(true, value)
	}
	return workforcePlanResponse(h.service.UpdateWorkforcePlan(ctx, request.Id, patch))
}
func (h *Handler) DeleteWorkforcePlan(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkforcePlan(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkforcePlanByID(ctx context.Context, request *common.IDRequest) (*pb.WorkforcePlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanResponse(h.service.GetWorkforcePlanByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkforcePlans(ctx context.Context, request *common.PageRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.GetAllWorkforcePlans(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByPlanCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByPlanCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByLegalEntityReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByLegalEntityReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByScenarioCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByScenarioCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByFinanceBudgetReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByFinanceBudgetReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByCreatedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByCreatedByPrincipalID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanByApprovedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlans(h.service.FindWorkforcePlanByApprovedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
