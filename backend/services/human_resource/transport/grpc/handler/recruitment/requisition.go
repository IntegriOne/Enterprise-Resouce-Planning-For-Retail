package recruitment

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func requisition(v *entity.Requisition) (*pb.Requisition, error) {
	minimum, err := response.JSONValue(v.MinimumQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	desired, err := response.JSONValue(v.DesiredQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.Requisition{
		Id: v.ID, RequisitionNumber: v.RequisitionNumber, WorkforcePlanLineId: v.WorkforcePlanLineID, PositionId: v.PositionID, JobId: v.JobID,
		OrganizationUnitReferenceId: v.OrganizationUnitReferenceID, HiringManagerPrincipalId: v.HiringManagerPrincipalID, RecruiterPrincipalId: v.RecruiterPrincipalID,
		RequisitionType: v.RequisitionType, Status: v.Status, Headcount: v.Headcount, TargetStartDate: dateString(v.TargetStartDate), BudgetReferenceId: v.BudgetReferenceID,
		WorkLocationReferenceId: v.WorkLocationReferenceID, EmploymentCategory: string(v.EmploymentCategory), MinimumQualifications: minimum, DesiredQualifications: desired,
		ApprovedAt: timestamp(v.ApprovedAt), ApprovedByPrincipalId: v.ApprovedByPrincipalID, OpenedAt: timestamp(v.OpenedAt), ClosedAt: timestamp(v.ClosedAt), IdempotencyKey: v.IdempotencyKey,
	}, nil
}
func requisitionResponse(v *entity.Requisition, err error) (*pb.RequisitionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := requisition(v)
	if err != nil {
		return nil, err
	}
	return &pb.RequisitionResponse{Item: item}, nil
}
func requisitions(page operation.Page[entity.Requisition], err error) (*pb.RequisitionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Requisition, 0, len(page.Items))
	for i := range page.Items {
		item, err := requisition(&page.Items[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &pb.RequisitionsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateRequisition(ctx context.Context, r *pb.CreateRequisitionRequest) (*pb.RequisitionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	target, err := optionalDate(in.TargetStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	minimum, err := response.JSON(in.MinimumQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	desired, err := response.JSON(in.DesiredQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	opened, err := response.OptionalTimestamp(in.OpenedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := response.OptionalTimestamp(in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return requisitionResponse(h.service.CreateRequisition(ctx, service.CreateRequisitionInput{
		RequisitionNumber: in.GetRequisitionNumber(), WorkforcePlanLineID: in.WorkforcePlanLineId, PositionID: in.PositionId, JobID: in.GetJobId(),
		OrganizationUnitReferenceID: in.GetOrganizationUnitReferenceId(), HiringManagerPrincipalID: in.GetHiringManagerPrincipalId(), RecruiterPrincipalID: in.RecruiterPrincipalId,
		RequisitionType: in.GetRequisitionType(), Status: in.Status, Headcount: in.Headcount, TargetStartDate: target, BudgetReferenceID: in.BudgetReferenceId,
		WorkLocationReferenceID: in.WorkLocationReferenceId, EmploymentCategory: entity.WorkerCategory(in.GetEmploymentCategory()), MinimumQualifications: minimum, DesiredQualifications: desired,
		ApprovedAt: approved, ApprovedByPrincipalID: in.ApprovedByPrincipalId, OpenedAt: opened, ClosedAt: closed, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (h *Handler) UpdateRequisition(ctx context.Context, r *pb.UpdateRequisitionRequest) (*pb.RequisitionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "requisition_number", "workforce_plan_line_id", "position_id", "job_id", "organization_unit_reference_id", "hiring_manager_principal_id", "recruiter_principal_id", "requisition_type", "status", "headcount", "target_start_date", "budget_reference_id", "work_location_reference_id", "employment_category", "minimum_qualifications", "desired_qualifications", "approved_at", "approved_by_principal_id", "opened_at", "closed_at", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	target, err := dateField(m["target_start_date"], in.TargetStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	minimum, err := jsonField(m["minimum_qualifications"], in.MinimumQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	desired, err := jsonField(m["desired_qualifications"], in.DesiredQualifications)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	opened, err := timeField(m["opened_at"], in.OpenedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := timeField(m["closed_at"], in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return requisitionResponse(h.service.UpdateRequisition(ctx, r.Id, service.UpdateRequisitionInput{
		RequisitionNumber: field(m["requisition_number"], in.RequisitionNumber), WorkforcePlanLineID: field(m["workforce_plan_line_id"], in.WorkforcePlanLineId), PositionID: field(m["position_id"], in.PositionId), JobID: field(m["job_id"], in.JobId),
		OrganizationUnitReferenceID: field(m["organization_unit_reference_id"], in.OrganizationUnitReferenceId), HiringManagerPrincipalID: field(m["hiring_manager_principal_id"], in.HiringManagerPrincipalId), RecruiterPrincipalID: field(m["recruiter_principal_id"], in.RecruiterPrincipalId),
		RequisitionType: field(m["requisition_type"], in.RequisitionType), Status: field(m["status"], in.Status), Headcount: field(m["headcount"], in.Headcount), TargetStartDate: target,
		BudgetReferenceID: field(m["budget_reference_id"], in.BudgetReferenceId), WorkLocationReferenceID: field(m["work_location_reference_id"], in.WorkLocationReferenceId), EmploymentCategory: field(m["employment_category"], enumPointer[entity.WorkerCategory](in.EmploymentCategory)),
		MinimumQualifications: minimum, DesiredQualifications: desired, ApprovedAt: approved, ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), OpenedAt: opened, ClosedAt: closed, IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	}))
}
func (h *Handler) DeleteRequisition(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteRequisition)
}
func (h *Handler) GetRequisitionByID(ctx context.Context, r *common.IDRequest) (*pb.RequisitionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitionResponse(h.service.GetRequisitionByID(ctx, r.Id))
}
func (h *Handler) GetAllRequisitions(ctx context.Context, r *common.PageRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.GetAllRequisitions(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByRequisitionNumber(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByRequisitionNumber(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByWorkforcePlanLineID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByWorkforcePlanLineID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByPositionID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByPositionID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByJobID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByJobID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByOrganizationUnitReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByOrganizationUnitReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByHiringManagerPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByHiringManagerPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByRecruiterPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByRecruiterPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByRequisitionType(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByRequisitionType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByBudgetReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByBudgetReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByWorkLocationReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByWorkLocationReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindRequisitionByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.RequisitionsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return requisitions(h.service.FindRequisitionByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
