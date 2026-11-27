package human_resource

import (
	"context"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const workforcePlanReadPermission = "human_resource.workforce_plans.read"

// WorkforcePlan membaca ringkasan rencana dengan cakupan yang diperiksa HRM.
func (a *Adapter) WorkforcePlan(ctx context.Context, rawID string) (*model.WorkforcePlan, error) {
	rpcCtx, err := a.authorizeRead(ctx, workforcePlanReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "workforce-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetWorkforcePlanByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "workforce plan")
	}
	plan := response.GetItem()
	if plan.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource response", nil)
	}
	return toModelWorkforcePlan(plan), nil
}

// WorkforcePlans membaca halaman dan total rencana yang telah dibatasi cakupan oleh HRM.
func (a *Adapter) WorkforcePlans(ctx context.Context, limit, offset int) (*Page[model.WorkforcePlan], error) {
	rpcCtx, err := a.authorizeRead(ctx, workforcePlanReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.GetAllWorkforcePlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, mapReadError(err, "workforce plan")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan page", nil)
	}
	result := &Page[model.WorkforcePlan]{
		Items: make([]*model.WorkforcePlan, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, plan := range items {
		if id, err := uuid.Parse(plan.GetId()); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource workforce plan page", nil)
		}
		result.Items = append(result.Items, toModelWorkforcePlan(plan))
	}
	return result, nil
}

func toModelWorkforcePlan(plan *workforce.WorkforcePlan) *model.WorkforcePlan {
	return &model.WorkforcePlan{
		ID: nodeID("workforce-plan", plan.GetId()), RawID: plan.GetId(), PlanCode: plan.GetPlanCode(), PlanVersion: int(plan.GetPlanVersion()),
		LegalEntityReferenceID: plan.GetLegalEntityReferenceId(), PeriodStart: plan.GetPeriodStart(),
		PeriodEnd: plan.GetPeriodEnd(), ScenarioCode: plan.GetScenarioCode(), Status: plan.GetStatus(),
		SupersededByPlanID: plan.SupersededByPlanId,
	}
}

// CreateWorkforcePlan membuat draf dengan identitas pembuat dari sesi.
func (a *Adapter) CreateWorkforcePlan(ctx context.Context, input model.CreateWorkforcePlanInput) (*model.WorkforcePlan, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workforce_plans.create")
	if err != nil {
		return nil, err
	}
	entityID, err := parseNodeID(input.LegalEntityReferenceID, "legal-entity")
	if err != nil || entityID == uuid.Nil || input.PlanVersion < 1 || input.PlanVersion > math.MaxInt32 {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan request", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	version, legalEntityID := int32(input.PlanVersion), entityID.String()
	response, err := a.workforce.CreateWorkforcePlan(rpcCtx, &workforce.CreateWorkforcePlanRequest{
		Input: &workforce.WorkforcePlanInput{
			PlanCode: &input.PlanCode, PlanVersion: &version, LegalEntityReferenceId: &legalEntityID,
			PeriodStart: &input.PeriodStart, PeriodEnd: &input.PeriodEnd, ScenarioCode: input.ScenarioCode,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("workforce plan conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "workforce plan")
	}
	plan := response.GetItem()
	id, err := uuid.Parse(plan.GetId())
	if err != nil || id == uuid.Nil || plan.GetStatus() != "draft" || plan.GetLegalEntityReferenceId() != legalEntityID {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan response", nil)
	}
	return toModelWorkforcePlan(plan), nil
}

// SubmitWorkforcePlan mengajukan draf rencana untuk review.
func (a *Adapter) SubmitWorkforcePlan(ctx context.Context, id string) (*model.WorkforcePlan, error) {
	return a.runWorkforcePlanCommand(ctx, id, "human_resource.workforce_plans.submit", "review", workforce.WorkforcePlanningServiceClient.SubmitWorkforcePlan)
}

// CancelWorkforcePlan menarik draf atau rencana yang sedang direview.
func (a *Adapter) CancelWorkforcePlan(ctx context.Context, id string) (*model.WorkforcePlan, error) {
	return a.runWorkforcePlanCommand(ctx, id, "human_resource.workforce_plans.cancel", "cancelled", workforce.WorkforcePlanningServiceClient.CancelWorkforcePlan)
}

// ApproveWorkforcePlan meneruskan keputusan dengan verifikasi anggaran di HRM.
func (a *Adapter) ApproveWorkforcePlan(ctx context.Context, id string) (*model.WorkforcePlan, error) {
	return a.runWorkforcePlanCommand(ctx, id, "human_resource.workforce_plans.approve", "approved", workforce.WorkforcePlanningServiceClient.ApproveWorkforcePlan)
}

// SupersedeWorkforcePlan menautkan rencana yang disetujui ke versi penggantinya.
func (a *Adapter) SupersedeWorkforcePlan(ctx context.Context, id, rawReplacementID, reasonCode string) (*model.WorkforcePlan, error) {
	return a.runWorkforcePlanCommand(ctx, id, "human_resource.workforce_plans.supersede", "superseded",
		func(client workforce.WorkforcePlanningServiceClient, rpcCtx context.Context, request *common.IDRequest, opts ...grpc.CallOption) (*workforce.WorkforcePlanResponse, error) {
			replacementID, err := parseNodeID(rawReplacementID, "workforce-plan")
			reason := strings.TrimSpace(reasonCode)
			if err != nil || replacementID == uuid.Nil || replacementID.String() == request.Id || reason == "" || len(reason) > 80 {
				return nil, status.Error(codes.InvalidArgument, "invalid workforce plan replacement")
			}
			response, err := client.SupersedeWorkforcePlan(rpcCtx, &workforce.SupersedeWorkforcePlanRequest{
				Id: request.Id, ReplacementPlanId: replacementID.String(), ReasonCode: reason,
			}, opts...)
			if err == nil && response.GetItem().GetSupersededByPlanId() != replacementID.String() {
				return nil, status.Error(codes.Internal, "invalid workforce plan replacement response")
			}
			return response, err
		},
	)
}

func (a *Adapter) runWorkforcePlanCommand(
	ctx context.Context,
	rawID, permission, expectedStatus string,
	command func(workforce.WorkforcePlanningServiceClient, context.Context, *common.IDRequest, ...grpc.CallOption) (*workforce.WorkforcePlanResponse, error),
) (*model.WorkforcePlan, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "workforce-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := command(a.workforce, rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("workforce plan does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "workforce plan")
	}
	plan := response.GetItem()
	if plan.GetId() != id.String() || plan.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan response", nil)
	}
	return toModelWorkforcePlan(plan), nil
}

// UpdateWorkforcePlan mengoreksi periode dan skenario pada draf.
func (a *Adapter) UpdateWorkforcePlan(ctx context.Context, rawID string, input model.UpdateWorkforcePlanInput) (*model.WorkforcePlan, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workforce_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "workforce-plan")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid workforce plan ID", nil)
	}
	patch := &workforce.WorkforcePlanInput{}
	mask := &fieldmaskpb.FieldMask{}
	for _, field := range []struct {
		path   string
		value  *string
		set    bool
		target **string
	}{
		{"period_start", input.PeriodStart.Value(), input.PeriodStart.IsSet(), &patch.PeriodStart},
		{"period_end", input.PeriodEnd.Value(), input.PeriodEnd.IsSet(), &patch.PeriodEnd},
		{"scenario_code", input.ScenarioCode.Value(), input.ScenarioCode.IsSet(), &patch.ScenarioCode},
	} {
		if !field.set {
			continue
		}
		if field.value == nil {
			return nil, middleware.ErrInvalidArgument("workforce plan update fields cannot be null", nil)
		}
		*field.target = field.value
		mask.Paths = append(mask.Paths, field.path)
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("workforce plan update is empty", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.workforce.UpdateWorkforcePlan(rpcCtx, &workforce.UpdateWorkforcePlanRequest{
		Id: id.String(), Input: patch, UpdateMask: mask,
	})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("workforce plan update conflicts with existing state", err)
	}
	if err != nil {
		return nil, mapReadError(err, "workforce plan")
	}
	plan := response.GetItem()
	if plan.GetId() != id.String() || plan.GetStatus() != "draft" {
		return nil, middleware.ErrUnavailable("invalid human resource workforce plan response", nil)
	}
	return toModelWorkforcePlan(plan), nil
}

// DeleteWorkforcePlan menghapus draf kosong yang lolos validasi HRM.
func (a *Adapter) DeleteWorkforcePlan(ctx context.Context, id string) (*bool, error) {
	return a.deleteWorkforceRecord(ctx, id, "human_resource.workforce_plans.delete", "workforce plan", workforce.WorkforcePlanningServiceClient.DeleteWorkforcePlan)
}

func (a *Adapter) deleteWorkforceRecord(
	ctx context.Context,
	rawID, permission, resource string,
	command func(workforce.WorkforcePlanningServiceClient, context.Context, *common.IDRequest, ...grpc.CallOption) (*common.EmptyResponse, error),
) (*bool, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, resource)
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid "+resource+" ID", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := command(a.workforce, rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict(resource+" does not meet deletion requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, resource)
	}
	if response == nil {
		return nil, middleware.ErrUnavailable("invalid human resource "+resource+" response", nil)
	}
	deleted := true
	return &deleted, nil
}
