package human_resource

import (
	"context"
	"math"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const workforcePlanLineReadPermission = "human_resource.workforce_plan_lines.read"

// WorkforcePlanLine membaca satu baris dengan cakupan badan hukum rencana induk.
func (a *Adapter) WorkforcePlanLine(ctx context.Context, rawID string) (*model.WorkforcePlanLine, error) {
	rpcCtx, err := a.authorizeRead(ctx, workforcePlanLineReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "workforce-plan-line")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan line ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetWorkforcePlanLineByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "workforce plan line")
	}
	line := response.GetItem()
	if line.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan line", nil)
	}
	return toModelWorkforcePlanLine(line), nil
}

// WorkforcePlanLines membaca halaman terotorisasi dengan filter rencana opsional.
func (a *Adapter) WorkforcePlanLines(ctx context.Context, rawPlanID *string, limit, offset int) (*Page[model.WorkforcePlanLine], error) {
	rpcCtx, err := a.authorizeRead(ctx, workforcePlanLineReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan line pagination", nil)
	}
	var planID string
	if rawPlanID != nil {
		id, err := parseNodeID(*rawPlanID, "workforce-plan")
		if err != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid workforce plan ID", nil)
		}
		planID = id.String()
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *workforce.WorkforcePlanLinesResponse
	if rawPlanID == nil {
		response, err = a.workforce.GetAllWorkforcePlanLines(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		response, err = a.workforce.FindWorkforcePlanLineByWorkforcePlanID(rpcCtx, &common.StringFinderRequest{
			Value: &planID, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "workforce plan line")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan line page", nil)
	}
	result := &Page[model.WorkforcePlanLine]{
		Items: make([]*model.WorkforcePlanLine, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, line := range items {
		id, err := uuid.Parse(line.GetId())
		if err != nil || id == uuid.Nil || (planID != "" && line.GetWorkforcePlanId() != planID) {
			return nil, middleware.ErrUnavailable("invalid human resource workforce plan line", nil)
		}
		result.Items = append(result.Items, toModelWorkforcePlanLine(line))
	}
	return result, nil
}

func toModelWorkforcePlanLine(line *workforce.WorkforcePlanLine) *model.WorkforcePlanLine {
	return &model.WorkforcePlanLine{
		ID: nodeID("workforce-plan-line", line.GetId()), RawID: line.GetId(), WorkforcePlanID: line.GetWorkforcePlanId(), LineNumber: int(line.GetLineNumber()),
		OrganizationUnitReferenceID: line.GetOrganizationUnitReferenceId(), JobID: line.GetJobId(),
		WorkLocationReferenceID: line.WorkLocationReferenceId, WorkerCategory: line.GetWorkerCategory(),
		CurrentFte: line.GetCurrentFte(), RequiredFte: line.GetRequiredFte(),
		NeededFrom: line.GetNeededFrom(), NeededUntil: line.NeededUntil, ActionCode: line.GetActionCode(),
	}
}

// CreateWorkforcePlanLine menambah kebutuhan pada draf rencana melalui HRM.
func (a *Adapter) CreateWorkforcePlanLine(ctx context.Context, input model.CreateWorkforcePlanLineInput) (*model.WorkforcePlanLine, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workforce_plan_lines.create")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(input.WorkforcePlanID, "workforce-plan")
	if err != nil || id == uuid.Nil || input.LineNumber < 1 || input.LineNumber > math.MaxInt32 {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan line request", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	planID, lineNumber := id.String(), int32(input.LineNumber)
	response, err := a.workforce.CreateWorkforcePlanLine(rpcCtx, &workforce.CreateWorkforcePlanLineRequest{
		Input: &workforce.WorkforcePlanLineInput{
			WorkforcePlanId: &planID, LineNumber: &lineNumber,
			OrganizationUnitReferenceId: optionalRawNodeID(&input.OrganizationUnitReferenceID, "organization-unit"), JobId: optionalRawNodeID(&input.JobID, "job"),
			WorkLocationReferenceId: optionalRawNodeID(input.WorkLocationReferenceID, "location"), WorkerCategory: &input.WorkerCategory,
			CurrentFte: &input.CurrentFte, RequiredFte: &input.RequiredFte,
			EstimatedCost: input.EstimatedCost, CurrencyCode: input.CurrencyCode,
			NeededFrom: &input.NeededFrom, NeededUntil: input.NeededUntil,
			ActionCode: &input.ActionCode, Rationale: &input.Rationale,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("workforce plan line conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "workforce plan line")
	}
	line := response.GetItem()
	lineID, err := uuid.Parse(line.GetId())
	if err != nil || lineID == uuid.Nil || line.GetWorkforcePlanId() != planID || line.GetLineNumber() != lineNumber {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan line", nil)
	}
	return toModelWorkforcePlanLine(line), nil
}

// DeleteWorkforcePlanLine menghapus rincian melalui pemeriksaan status dan referensi HRM.
func (a *Adapter) DeleteWorkforcePlanLine(ctx context.Context, id string) (*bool, error) {
	return a.deleteWorkforceRecord(ctx, id, "human_resource.workforce_plan_lines.delete", "workforce plan line", workforce.WorkforcePlanningServiceClient.DeleteWorkforcePlanLine)
}

// UpdateWorkforcePlanLine mengoreksi field yang dikirim pada rincian draf.
func (a *Adapter) UpdateWorkforcePlanLine(ctx context.Context, rawID string, input model.UpdateWorkforcePlanLineInput) (*model.WorkforcePlanLine, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workforce_plan_lines.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "workforce-plan-line")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan line ID", nil)
	}
	patch := &workforce.WorkforcePlanLineInput{}
	mask := &fieldmaskpb.FieldMask{}
	if number, set := input.LineNumber.ValueOK(); set {
		if number == nil || *number < 1 || *number > math.MaxInt32 {
			return nil, middleware.ErrInvalidArgument("invalid workforce plan line number", nil)
		}
		value := int32(*number)
		patch.LineNumber = &value
		mask.Paths = append(mask.Paths, "line_number")
	}
	for _, field := range []struct {
		path   string
		input  graphql.Omittable[*string]
		target **string
	}{
		{"organization_unit_reference_id", input.OrganizationUnitReferenceID, &patch.OrganizationUnitReferenceId},
		{"job_id", input.JobID, &patch.JobId},
		{"work_location_reference_id", input.WorkLocationReferenceID, &patch.WorkLocationReferenceId},
		{"worker_category", input.WorkerCategory, &patch.WorkerCategory},
		{"current_fte", input.CurrentFte, &patch.CurrentFte},
		{"required_fte", input.RequiredFte, &patch.RequiredFte},
		{"estimated_cost", input.EstimatedCost, &patch.EstimatedCost},
		{"currency_code", input.CurrencyCode, &patch.CurrencyCode},
		{"needed_from", input.NeededFrom, &patch.NeededFrom},
		{"needed_until", input.NeededUntil, &patch.NeededUntil},
		{"action_code", input.ActionCode, &patch.ActionCode},
		{"rationale", input.Rationale, &patch.Rationale},
	} {
		if value, set := field.input.ValueOK(); set {
			switch field.path {
			case "job_id":
				value = optionalRawNodeID(value, "job")
			case "organization_unit_reference_id":
				value = optionalRawNodeID(value, "organization-unit")
			case "work_location_reference_id":
				value = optionalRawNodeID(value, "location")
			}
			*field.target = value
			mask.Paths = append(mask.Paths, field.path)
		}
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("workforce plan line update is empty", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.UpdateWorkforcePlanLine(rpcCtx, &workforce.UpdateWorkforcePlanLineRequest{
		Id: id.String(), Input: patch, UpdateMask: mask,
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("workforce plan line conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "workforce plan line")
	}
	line := response.GetItem()
	if line.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan line", nil)
	}
	return toModelWorkforcePlanLine(line), nil
}
