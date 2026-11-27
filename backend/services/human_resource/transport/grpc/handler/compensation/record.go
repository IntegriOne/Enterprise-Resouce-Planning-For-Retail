package compensation

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/compensation/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/compensation"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func record(value *entity.CompensationRecord) *pb.CompensationRecord {
	item := &pb.CompensationRecord{Id: value.ID, WorkerId: value.WorkerID, AssignmentId: value.AssignmentID, ComponentId: value.ComponentID,
		ComponentCodeSnapshot: value.ComponentCodeSnapshot, CompensationBandId: value.CompensationBandID, CompensationType: value.CompensationType,
		Amount: string(value.Amount), CurrencyCode: value.CurrencyCode, PayFrequency: value.PayFrequency, EffectiveFrom: value.EffectiveFrom.Format("2006-01-02"),
		EffectiveTo: dateString(value.EffectiveTo), Status: string(value.Status), ReasonCode: value.ReasonCode, ApprovedByPrincipalId: value.ApprovedByPrincipalID,
		SourceChangeReferenceId: value.SourceChangeReferenceID, AccessClass: value.AccessClass}
	if value.ApprovedAt != nil {
		item.ApprovedAt = timestamppb.New(*value.ApprovedAt)
	}
	return item
}
func recordResponse(value *entity.CompensationRecord, err error) (*pb.CompensationRecordResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CompensationRecordResponse{Item: record(value)}, nil
}
func records(page operation.Page[entity.CompensationRecord], err error) (*pb.CompensationRecordsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.CompensationRecord, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, record(&page.Items[index]))
	}
	return &pb.CompensationRecordsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateCompensationRecord(ctx context.Context, request *pb.CreateCompensationRecordRequest) (*pb.CompensationRecordResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return recordResponse(h.service.CreateCompensationRecord(ctx, service.CreateCompensationRecordInput{WorkerID: in.GetWorkerId(), AssignmentID: in.GetAssignmentId(), ComponentID: in.GetComponentId(),
		ComponentCodeSnapshot: in.GetComponentCodeSnapshot(), CompensationBandID: in.CompensationBandId, CompensationType: in.GetCompensationType(), Amount: entity.Decimal(in.GetAmount()),
		CurrencyCode: in.GetCurrencyCode(), PayFrequency: in.GetPayFrequency(), EffectiveFrom: from, EffectiveTo: to, Status: statusPointer(in.Status), ReasonCode: in.ReasonCode,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, SourceChangeReferenceID: in.SourceChangeReferenceId, AccessClass: in.AccessClass}))
}
func (h *Handler) UpdateCompensationRecord(ctx context.Context, request *pb.UpdateCompensationRecordRequest) (*pb.CompensationRecordResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "assignment_id", "component_id", "component_code_snapshot", "compensation_band_id", "compensation_type", "amount", "currency_code", "pay_frequency", "effective_from", "effective_to", "status", "reason_code", "approved_by_principal_id", "approved_at", "source_change_reference_id", "access_class")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateCompensationRecordInput{WorkerID: field(mask["worker_id"], in.WorkerId), AssignmentID: field(mask["assignment_id"], in.AssignmentId), ComponentID: field(mask["component_id"], in.ComponentId),
		ComponentCodeSnapshot: field(mask["component_code_snapshot"], in.ComponentCodeSnapshot), CompensationBandID: field(mask["compensation_band_id"], in.CompensationBandId), CompensationType: field(mask["compensation_type"], in.CompensationType),
		Amount: field(mask["amount"], decimalPointer(in.Amount)), CurrencyCode: field(mask["currency_code"], in.CurrencyCode), PayFrequency: field(mask["pay_frequency"], in.PayFrequency), EffectiveFrom: from, EffectiveTo: to,
		Status: field(mask["status"], statusPointer(in.Status)), ReasonCode: field(mask["reason_code"], in.ReasonCode), ApprovedByPrincipalID: field(mask["approved_by_principal_id"], in.ApprovedByPrincipalId),
		SourceChangeReferenceID: field(mask["source_change_reference_id"], in.SourceChangeReferenceId), AccessClass: field(mask["access_class"], in.AccessClass)}
	if mask["approved_at"] {
		value, err := response.OptionalTimestamp(in.ApprovedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.ApprovedAt = field(true, value)
	}
	return recordResponse(h.service.UpdateCompensationRecord(ctx, request.Id, patch))
}
func (h *Handler) DeleteCompensationRecord(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteCompensationRecord(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetCompensationRecordByID(ctx context.Context, request *common.IDRequest) (*pb.CompensationRecordResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(h.service.GetCompensationRecordByID(ctx, request.Id))
}
func (h *Handler) GetAllCompensationRecords(ctx context.Context, request *common.PageRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.GetAllCompensationRecords(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByAssignmentID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByAssignmentID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByComponentID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByComponentID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByCompensationBandID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByCompensationBandID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByCompensationType(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByCompensationType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByCurrencyCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByCurrencyCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByStatus(ctx, entity.CompensationStatus(*request.Value), int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByReasonCode(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByReasonCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordByApprovedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordByApprovedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCompensationRecordBySourceChangeReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.CompensationRecordsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(h.service.FindCompensationRecordBySourceChangeReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}

func (h *Handler) ApproveCompensationRecord(ctx context.Context, request *common.IDRequest) (*pb.CompensationRecordResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(h.service.ApproveCompensationRecord(ctx, request.Id))
}
