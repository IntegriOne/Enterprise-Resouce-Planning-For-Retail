package personnel

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func contact(value *entity.EmergencyContact) *pb.EmergencyContact {
	item := &pb.EmergencyContact{Id: value.ID, WorkerId: value.WorkerID, ContactPartyReferenceId: value.ContactPartyReferenceID,
		RelationshipCode: value.RelationshipCode, PriorityOrder: value.PriorityOrder, IsActive: value.IsActive}
	if value.AuthorizationConfirmedAt != nil {
		item.AuthorizationConfirmedAt = timestamppb.New(*value.AuthorizationConfirmedAt)
	}
	return item
}
func contactResponse(value *entity.EmergencyContact, err error) (*pb.EmergencyContactResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EmergencyContactResponse{Item: contact(value)}, nil
}
func contacts(page operation.Page[entity.EmergencyContact], err error) (*pb.EmergencyContactsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.EmergencyContact, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, contact(&page.Items[index]))
	}
	return &pb.EmergencyContactsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateEmergencyContact(ctx context.Context, request *pb.CreateEmergencyContactRequest) (*pb.EmergencyContactResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	confirmed, err := response.OptionalTimestamp(in.AuthorizationConfirmedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return contactResponse(h.service.CreateEmergencyContact(ctx, service.CreateEmergencyContactInput{WorkerID: in.GetWorkerId(), ContactPartyReferenceID: in.GetContactPartyReferenceId(),
		RelationshipCode: in.GetRelationshipCode(), PriorityOrder: in.GetPriorityOrder(), AuthorizationConfirmedAt: confirmed, IsActive: in.IsActive}))
}
func (h *Handler) UpdateEmergencyContact(ctx context.Context, request *pb.UpdateEmergencyContactRequest) (*pb.EmergencyContactResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "contact_party_reference_id", "relationship_code", "priority_order", "authorization_confirmed_at", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	patch := service.UpdateEmergencyContactInput{WorkerID: field(mask["worker_id"], in.WorkerId), ContactPartyReferenceID: field(mask["contact_party_reference_id"], in.ContactPartyReferenceId),
		RelationshipCode: field(mask["relationship_code"], in.RelationshipCode), PriorityOrder: field(mask["priority_order"], in.PriorityOrder), IsActive: field(mask["is_active"], in.IsActive)}
	if mask["authorization_confirmed_at"] {
		value, err := response.OptionalTimestamp(in.AuthorizationConfirmedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.AuthorizationConfirmedAt = field(true, value)
	}
	return contactResponse(h.service.UpdateEmergencyContact(ctx, request.Id, patch))
}
func (h *Handler) DeleteEmergencyContact(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteEmergencyContact(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetEmergencyContactByID(ctx context.Context, request *common.IDRequest) (*pb.EmergencyContactResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contactResponse(h.service.GetEmergencyContactByID(ctx, request.Id))
}
func (h *Handler) GetAllEmergencyContacts(ctx context.Context, request *common.PageRequest) (*pb.EmergencyContactsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contacts(h.service.GetAllEmergencyContacts(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmergencyContactByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmergencyContactsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contacts(h.service.FindEmergencyContactByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmergencyContactByContactPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmergencyContactsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contacts(h.service.FindEmergencyContactByContactPartyReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmergencyContactByRelationshipCode(ctx context.Context, request *common.StringFinderRequest) (*pb.EmergencyContactsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contacts(h.service.FindEmergencyContactByRelationshipCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
