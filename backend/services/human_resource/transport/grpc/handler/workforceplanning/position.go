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

func position(value *entity.Position) *pb.Position {
	return &pb.Position{Id: value.ID, PositionCode: value.PositionCode, PositionTitle: value.PositionTitle, JobId: value.JobID,
		OrganizationUnitReferenceId: value.OrganizationUnitReferenceID, LegalEntityReferenceId: value.LegalEntityReferenceID,
		ReportsToPositionId: value.ReportsToPositionID, PositionStatus: value.PositionStatus, HeadcountAuthorized: string(value.HeadcountAuthorized),
		WorkLocationReferenceId: value.WorkLocationReferenceID, CostCenterReferenceId: value.CostCenterReferenceID,
		ValidFrom: value.ValidFrom.Format("2006-01-02"), ValidTo: dateString(value.ValidTo)}
}
func positionResponse(value *entity.Position, err error) (*pb.PositionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PositionResponse{Item: position(value)}, nil
}
func positions(page operation.Page[entity.Position], err error) (*pb.PositionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Position, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, position(&page.Items[index]))
	}
	return &pb.PositionsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePosition(ctx context.Context, request *pb.CreatePositionRequest) (*pb.PositionResponse, error) {
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
	return positionResponse(h.service.CreatePosition(ctx, service.CreatePositionInput{
		PositionCode: in.GetPositionCode(), PositionTitle: in.GetPositionTitle(), JobID: in.GetJobId(),
		OrganizationUnitReferenceID: in.GetOrganizationUnitReferenceId(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		ReportsToPositionID: in.ReportsToPositionId, PositionStatus: in.PositionStatus, HeadcountAuthorized: decimalPointer(in.HeadcountAuthorized),
		WorkLocationReferenceID: in.WorkLocationReferenceId, CostCenterReferenceID: in.CostCenterReferenceId, ValidFrom: from, ValidTo: to,
	}))
}

// ActivatePosition meneruskan aktivasi posisi ke command HRM.
func (h *Handler) ActivatePosition(ctx context.Context, request *common.IDRequest) (*pb.PositionResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positionResponse(h.service.ActivatePosition(ctx, request.Id))
}
func (h *Handler) UpdatePosition(ctx context.Context, request *pb.UpdatePositionRequest) (*pb.PositionResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "position_code", "position_title", "job_id", "organization_unit_reference_id", "legal_entity_reference_id", "reports_to_position_id", "position_status", "headcount_authorized", "work_location_reference_id", "cost_center_reference_id", "valid_from", "valid_to")
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
	return positionResponse(h.service.UpdatePosition(ctx, request.Id, service.UpdatePositionInput{
		PositionCode: field(mask["position_code"], in.PositionCode), PositionTitle: field(mask["position_title"], in.PositionTitle), JobID: field(mask["job_id"], in.JobId),
		OrganizationUnitReferenceID: field(mask["organization_unit_reference_id"], in.OrganizationUnitReferenceId), LegalEntityReferenceID: field(mask["legal_entity_reference_id"], in.LegalEntityReferenceId),
		ReportsToPositionID: field(mask["reports_to_position_id"], in.ReportsToPositionId), PositionStatus: field(mask["position_status"], in.PositionStatus),
		HeadcountAuthorized: field(mask["headcount_authorized"], decimalPointer(in.HeadcountAuthorized)), WorkLocationReferenceID: field(mask["work_location_reference_id"], in.WorkLocationReferenceId),
		CostCenterReferenceID: field(mask["cost_center_reference_id"], in.CostCenterReferenceId), ValidFrom: from, ValidTo: to,
	}))
}
func (h *Handler) DeletePosition(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePosition(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPositionByID(ctx context.Context, request *common.IDRequest) (*pb.PositionResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positionResponse(h.service.GetPositionByID(ctx, request.Id))
}
func (h *Handler) GetAllPositions(ctx context.Context, request *common.PageRequest) (*pb.PositionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.GetAllPositions(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByPositionCode(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByPositionCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByJobID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByJobID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByOrganizationUnitReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByOrganizationUnitReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByLegalEntityReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByLegalEntityReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByReportsToPositionID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByReportsToPositionID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByPositionStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByPositionStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByWorkLocationReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByWorkLocationReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindPositionByCostCenterReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.PositionsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return positions(h.service.FindPositionByCostCenterReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
