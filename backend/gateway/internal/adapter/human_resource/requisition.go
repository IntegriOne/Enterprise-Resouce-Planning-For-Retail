package human_resource

import (
	"context"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const requisitionReadPermission = "human_resource.requisitions.read"

// Requisition membaca detail kandidat berdasarkan ID.
func (a *Adapter) Requisition(ctx context.Context, rawID string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, requisitionReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.GetRequisitionByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}

// Requisitions membaca satu halaman kandidat dengan filter nomor opsional.
func (a *Adapter) Requisitions(ctx context.Context, number *string, limit, offset int) (*Page[model.Requisition], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, requisitionReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid requisition pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *recruitment.RequisitionsResponse
	if number == nil {
		response, err = a.recruitment.GetAllRequisitions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		value := *number
		if strings.TrimSpace(value) == "" || !utf8.ValidString(value) || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > 120 {
			return nil, middleware.ErrInvalidArgument("invalid requisition number", nil)
		}
		response, err = a.recruitment.FindRequisitionByRequisitionNumber(rpcCtx, &common.StringFinderRequest{
			Value: &value, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource requisition page", nil)
	}
	result := &Page[model.Requisition]{
		Items: make([]*model.Requisition, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range items {
		item, err := toModelRequisition(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelRequisition(item *recruitment.Requisition) (*model.Requisition, error) {
	id, err := uuid.Parse(item.GetId())
	if err != nil || id == uuid.Nil || item.GetRequisitionNumber() == "" || item.GetStatus() == "" ||
		item.GetHeadcount() <= 0 || item.GetJobId() == "" || item.GetOrganizationUnitReferenceId() == "" ||
		item.GetEmploymentCategory() == "" || item.GetRequisitionType() == "" {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	var managerWorkerID *string
	if raw := item.HiringManagerWorkerId; raw != nil {
		workerID, err := uuid.Parse(*raw)
		if err != nil || workerID == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource hiring manager worker ID", nil)
		}
		gid := workerGIDPrefix + workerID.String()
		managerWorkerID = &gid
	}
	return &model.Requisition{
		ID: nodeID("requisition", id.String()), RawID: id.String(), RequisitionNumber: item.GetRequisitionNumber(), Status: item.GetStatus(),
		Headcount: int(item.GetHeadcount()), JobID: item.GetJobId(),
		OrganizationUnitReferenceID: item.GetOrganizationUnitReferenceId(), RequisitionType: item.GetRequisitionType(),
		TargetStartDate: item.TargetStartDate, EmploymentCategory: item.GetEmploymentCategory(),
		RejectionReasonCode:     item.RejectionReasonCode,
		HiringManagerWorkerID:   managerWorkerID,
		WorkLocationReferenceID: item.WorkLocationReferenceId, PositionID: item.PositionId, WorkforcePlanLineID: item.WorkforcePlanLineId,
	}, nil
}

// CreateRequisition mencatat draf dengan hiring manager yang dipilih berdasarkan pekerja.
func (a *Adapter) CreateRequisition(ctx context.Context, input model.CreateRequisitionInput) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.requisitions.create")
	if err != nil {
		return nil, err
	}
	workerID, err := uuid.Parse(strings.TrimPrefix(input.HiringManagerWorkerID, workerGIDPrefix))
	if err != nil || workerID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid hiring manager worker ID", nil)
	}
	if input.Headcount < 1 || input.Headcount > math.MaxInt32 {
		return nil, middleware.ErrInvalidArgument("invalid requisition headcount", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	manager, headcount := workerID.String(), int32(input.Headcount)
	jobID := rawNodeID(input.JobID, "job")
	response, err := a.recruitment.CreateRequisition(rpcCtx, &recruitment.CreateRequisitionRequest{
		HiringManagerWorkerId: &manager,
		Input: &recruitment.RequisitionInput{
			RequisitionNumber: &input.RequisitionNumber, JobId: &jobID,
			OrganizationUnitReferenceId: optionalRawNodeID(&input.OrganizationUnitReferenceID, "organization-unit"), RequisitionType: &input.RequisitionType,
			EmploymentCategory: &input.EmploymentCategory, Headcount: &headcount, TargetStartDate: input.TargetStartDate,
			PositionId: optionalRawNodeID(input.PositionID, "position"), WorkforcePlanLineId: optionalRawNodeID(input.WorkforcePlanLineID, "workforce-plan-line"),
			WorkLocationReferenceId: optionalRawNodeID(input.WorkLocationReferenceID, "location"), IdempotencyKey: &input.IdempotencyKey,
		},
	})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("requisition conflicts with existing state or references", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetStatus() != "draft" {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}

// SubmitRequisition mengajukan draf milik aktor ke tahap persetujuan.
func (a *Adapter) SubmitRequisition(ctx context.Context, rawID string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.requisitions.submit")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.SubmitRequisition(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("requisition does not meet submission requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetStatus() != "submitted" {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}

// RejectRequisition mencatat penolakan oleh peninjau requisition.
func (a *Adapter) RejectRequisition(ctx context.Context, rawID, reasonCode string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.requisitions.reject")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" || len(reason) > 80 {
		return nil, middleware.ErrInvalidArgument("invalid requisition rejection reason", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.RejectRequisition(rpcCtx, &recruitment.RejectRequisitionRequest{Id: id.String(), ReasonCode: reason})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("requisition does not meet rejection requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetStatus() != "rejected" || item.GetRejectionReasonCode() != reason {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}

// DeleteRequisition menghapus requisition yang masih berstatus draf.
func (a *Adapter) DeleteRequisition(ctx context.Context, rawID string) (*bool, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.requisitions.delete")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.DeleteRequisition(rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("requisition does not meet deletion requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	if response == nil {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	deleted := true
	return &deleted, nil
}

// UpdateRequisition mengubah rincian kebutuhan pada requisition draf.
func (a *Adapter) UpdateRequisition(ctx context.Context, rawID string, input model.UpdateRequisitionInput) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.requisitions.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	var hiringManagerID *string
	if value, set := input.HiringManagerWorkerID.ValueOK(); set {
		if value == nil {
			return nil, middleware.ErrInvalidArgument("hiring manager worker ID cannot be null", nil)
		}
		manager, parseErr := parseNodeID(*value, "hrm-worker")
		if parseErr != nil || manager == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid hiring manager worker ID", nil)
		}
		canonical := manager.String()
		hiringManagerID = &canonical
	}
	patch, mask, err := requisitionPatch(input)
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.UpdateRequisition(rpcCtx, &recruitment.UpdateRequisitionRequest{Id: id.String(), Input: patch, UpdateMask: mask, HiringManagerWorkerId: hiringManagerID})
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("requisition update conflicts with existing state or references", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetStatus() != "draft" {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}

func requisitionPatch(input model.UpdateRequisitionInput) (*recruitment.RequisitionInput, *fieldmaskpb.FieldMask, error) {
	patch := &recruitment.RequisitionInput{}
	mask := &fieldmaskpb.FieldMask{}
	for _, field := range []struct {
		path, resource string
		value          *string
		set, nullable  bool
		target         **string
	}{
		{"target_start_date", "", input.TargetStartDate.Value(), input.TargetStartDate.IsSet(), true, &patch.TargetStartDate},
		{"requisition_type", "", input.RequisitionType.Value(), input.RequisitionType.IsSet(), false, &patch.RequisitionType},
		{"employment_category", "", input.EmploymentCategory.Value(), input.EmploymentCategory.IsSet(), false, &patch.EmploymentCategory},
		{"job_id", "job", input.JobID.Value(), input.JobID.IsSet(), false, &patch.JobId},
		{"organization_unit_reference_id", "organization-unit", input.OrganizationUnitReferenceID.Value(), input.OrganizationUnitReferenceID.IsSet(), false, &patch.OrganizationUnitReferenceId},
		{"work_location_reference_id", "location", input.WorkLocationReferenceID.Value(), input.WorkLocationReferenceID.IsSet(), true, &patch.WorkLocationReferenceId},
		{"position_id", "position", input.PositionID.Value(), input.PositionID.IsSet(), true, &patch.PositionId},
		{"workforce_plan_line_id", "workforce-plan-line", input.WorkforcePlanLineID.Value(), input.WorkforcePlanLineID.IsSet(), true, &patch.WorkforcePlanLineId},
	} {
		if !field.set {
			continue
		}
		if field.value == nil && !field.nullable {
			return nil, nil, middleware.ErrInvalidArgument(field.path+" cannot be null", nil)
		}
		value := field.value
		if value != nil && field.resource != "" {
			id, err := parseNodeID(*value, field.resource)
			if err != nil || id == uuid.Nil {
				return nil, nil, middleware.ErrInvalidArgument("invalid "+field.path, nil)
			}
			canonical := id.String()
			value = &canonical
		}
		*field.target = value
		mask.Paths = append(mask.Paths, field.path)
	}
	if input.Headcount.IsSet() {
		value := input.Headcount.Value()
		if value == nil || *value < 1 || *value > math.MaxInt32 {
			return nil, nil, middleware.ErrInvalidArgument("invalid requisition headcount", nil)
		}
		count := int32(*value)
		patch.Headcount = &count
		mask.Paths = append(mask.Paths, "headcount")
	}
	if input.HiringManagerWorkerID.IsSet() {
		mask.Paths = append(mask.Paths, "hiring_manager_worker_id")
	}
	if len(mask.Paths) == 0 {
		return nil, nil, middleware.ErrInvalidArgument("requisition update is empty", nil)
	}
	return patch, mask, nil
}
