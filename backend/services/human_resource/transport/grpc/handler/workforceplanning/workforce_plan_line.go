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
)

func workforcePlanLine(value *entity.WorkforcePlanLine) *pb.WorkforcePlanLine {
	return &pb.WorkforcePlanLine{Id: value.ID, WorkforcePlanId: value.WorkforcePlanID, LineNumber: value.LineNumber,
		OrganizationUnitReferenceId: value.OrganizationUnitReferenceID, JobId: value.JobID, WorkLocationReferenceId: value.WorkLocationReferenceID,
		WorkerCategory: string(value.WorkerCategory), CurrentFte: string(value.CurrentFTE), RequiredFte: string(value.RequiredFTE),
		EstimatedCost: decimalString(value.EstimatedCost), CurrencyCode: value.CurrencyCode, NeededFrom: value.NeededFrom.Format("2006-01-02"),
		NeededUntil: dateString(value.NeededUntil), ActionCode: value.ActionCode, Rationale: value.Rationale}
}
func workforcePlanLineResponse(value *entity.WorkforcePlanLine, err error) (*pb.WorkforcePlanLineResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkforcePlanLineResponse{Item: workforcePlanLine(value)}, nil
}
func workforcePlanLines(page operation.Page[entity.WorkforcePlanLine], err error) (*pb.WorkforcePlanLinesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkforcePlanLine, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, workforcePlanLine(&page.Items[index]))
	}
	return &pb.WorkforcePlanLinesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkforcePlanLine(ctx context.Context, request *pb.CreateWorkforcePlanLineRequest) (*pb.WorkforcePlanLineResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetNeededFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	until, err := optionalDate(in.NeededUntil)
	if err != nil {
		return nil, response.Error(err)
	}
	return workforcePlanLineResponse(h.service.CreateWorkforcePlanLine(ctx, service.CreateWorkforcePlanLineInput{
		WorkforcePlanID: in.GetWorkforcePlanId(), LineNumber: in.GetLineNumber(), OrganizationUnitReferenceID: in.GetOrganizationUnitReferenceId(),
		JobID: in.GetJobId(), WorkLocationReferenceID: in.WorkLocationReferenceId, WorkerCategory: entity.WorkerCategory(in.GetWorkerCategory()),
		CurrentFTE: entity.Decimal(in.GetCurrentFte()), RequiredFTE: entity.Decimal(in.GetRequiredFte()), EstimatedCost: decimalPointer(in.EstimatedCost),
		CurrencyCode: in.CurrencyCode, NeededFrom: from, NeededUntil: until, ActionCode: in.GetActionCode(), Rationale: in.GetRationale(),
	}))
}
func (h *Handler) UpdateWorkforcePlanLine(ctx context.Context, request *pb.UpdateWorkforcePlanLineRequest) (*pb.WorkforcePlanLineResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "workforce_plan_id", "line_number", "organization_unit_reference_id", "job_id", "work_location_reference_id", "worker_category", "current_fte", "required_fte", "estimated_cost", "currency_code", "needed_from", "needed_until", "action_code", "rationale")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["needed_from"], in.NeededFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	until, err := dateField(mask["needed_until"], in.NeededUntil)
	if err != nil {
		return nil, response.Error(err)
	}
	var category *entity.WorkerCategory
	if in.WorkerCategory != nil {
		value := entity.WorkerCategory(*in.WorkerCategory)
		category = &value
	}
	return workforcePlanLineResponse(h.service.UpdateWorkforcePlanLine(ctx, request.Id, service.UpdateWorkforcePlanLineInput{
		WorkforcePlanID: field(mask["workforce_plan_id"], in.WorkforcePlanId), LineNumber: field(mask["line_number"], in.LineNumber),
		OrganizationUnitReferenceID: field(mask["organization_unit_reference_id"], in.OrganizationUnitReferenceId), JobID: field(mask["job_id"], in.JobId),
		WorkLocationReferenceID: field(mask["work_location_reference_id"], in.WorkLocationReferenceId), WorkerCategory: field(mask["worker_category"], category),
		CurrentFTE: field(mask["current_fte"], decimalPointer(in.CurrentFte)), RequiredFTE: field(mask["required_fte"], decimalPointer(in.RequiredFte)),
		EstimatedCost: field(mask["estimated_cost"], decimalPointer(in.EstimatedCost)), CurrencyCode: field(mask["currency_code"], in.CurrencyCode),
		NeededFrom: from, NeededUntil: until, ActionCode: field(mask["action_code"], in.ActionCode), Rationale: field(mask["rationale"], in.Rationale),
	}))
}
func (h *Handler) DeleteWorkforcePlanLine(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkforcePlanLine(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkforcePlanLineByID(ctx context.Context, request *common.IDRequest) (*pb.WorkforcePlanLineResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLineResponse(h.service.GetWorkforcePlanLineByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkforcePlanLines(ctx context.Context, request *common.PageRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.GetAllWorkforcePlanLines(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByWorkforcePlanID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByWorkforcePlanID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByLineNumber(ctx context.Context, request *common.Int32FinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByLineNumber(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByOrganizationUnitReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByOrganizationUnitReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByJobID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByJobID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByWorkLocationReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByWorkLocationReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByCurrencyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByCurrencyCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkforcePlanLineByActionCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkforcePlanLinesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workforcePlanLines(h.service.FindWorkforcePlanLineByActionCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
