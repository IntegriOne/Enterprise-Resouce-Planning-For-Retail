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

func lifecycleEvent(value *entity.WorkerLifecycleEvent) (*pb.WorkerLifecycleEvent, error) {
	payload, err := response.JSONValue(value.EventPayload)
	if err != nil {
		return nil, err
	}
	return &pb.WorkerLifecycleEvent{Id: value.ID, WorkerId: value.WorkerID, EventType: value.EventType, OccurredAt: timestamppb.New(value.OccurredAt),
		EffectiveDate: value.EffectiveDate.Format("2006-01-02"), PreviousStatus: statusString(value.PreviousStatus), NewStatus: statusString(value.NewStatus),
		SourceReference: value.SourceReference, ActorPrincipalId: value.ActorPrincipalID, EventPayload: payload, IdempotencyKey: value.IdempotencyKey}, nil
}
func eventResponse(value *entity.WorkerLifecycleEvent, err error) (*pb.WorkerLifecycleEventResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := lifecycleEvent(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerLifecycleEventResponse{Item: item}, nil
}
func events(page operation.Page[entity.WorkerLifecycleEvent], err error) (*pb.WorkerLifecycleEventsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerLifecycleEvent, 0, len(page.Items))
	for index := range page.Items {
		item, err := lifecycleEvent(&page.Items[index])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.WorkerLifecycleEventsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkerLifecycleEvent(ctx context.Context, request *pb.CreateWorkerLifecycleEventRequest) (*pb.WorkerLifecycleEventResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	occurred, err := response.Timestamp(in.OccurredAt)
	if err != nil {
		return nil, response.Error(err)
	}
	effective, err := response.Date(in.GetEffectiveDate())
	if err != nil {
		return nil, response.Error(err)
	}
	payload, err := response.JSON(in.EventPayload)
	if err != nil {
		return nil, response.Error(err)
	}
	return eventResponse(h.service.CreateWorkerLifecycleEvent(ctx, service.CreateWorkerLifecycleEventInput{WorkerID: in.GetWorkerId(), EventType: in.GetEventType(), OccurredAt: occurred,
		EffectiveDate: effective, PreviousStatus: statusPointer(in.PreviousStatus), NewStatus: statusPointer(in.NewStatus), SourceReference: in.SourceReference, EventPayload: payload, IdempotencyKey: in.GetIdempotencyKey()}))
}
func (h *Handler) UpdateWorkerLifecycleEvent(ctx context.Context, request *pb.UpdateWorkerLifecycleEventRequest) (*pb.WorkerLifecycleEventResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "event_type", "occurred_at", "effective_date", "previous_status", "new_status", "source_reference", "event_payload", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	effective, err := dateField(mask["effective_date"], in.EffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateWorkerLifecycleEventInput{WorkerID: field(mask["worker_id"], in.WorkerId), EventType: field(mask["event_type"], in.EventType), EffectiveDate: effective,
		PreviousStatus: field(mask["previous_status"], statusPointer(in.PreviousStatus)), NewStatus: field(mask["new_status"], statusPointer(in.NewStatus)),
		SourceReference: field(mask["source_reference"], in.SourceReference), IdempotencyKey: field(mask["idempotency_key"], in.IdempotencyKey)}
	if mask["occurred_at"] {
		value, err := response.OptionalTimestamp(in.OccurredAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.OccurredAt = field(true, value)
	}
	if mask["event_payload"] {
		patch.EventPayload.Set = true
		if in.EventPayload != nil {
			value, err := response.JSON(in.EventPayload)
			if err != nil {
				return nil, response.Error(err)
			}
			patch.EventPayload.Value = &value
		}
	}
	return eventResponse(h.service.UpdateWorkerLifecycleEvent(ctx, request.Id, patch))
}
func (h *Handler) DeleteWorkerLifecycleEvent(ctx context.Context, request *common.IdentityIDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkerLifecycleEvent(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerLifecycleEventByID(ctx context.Context, request *common.IdentityIDRequest) (*pb.WorkerLifecycleEventResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return eventResponse(h.service.GetWorkerLifecycleEventByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkerLifecycleEvents(ctx context.Context, request *common.PageRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.GetAllWorkerLifecycleEvents(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByEventType(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByEventType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByPreviousStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByPreviousStatus(ctx, statusPointer(request.Value), int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByNewStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByNewStatus(ctx, statusPointer(request.Value), int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByActorPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByActorPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerLifecycleEventByIdempotencyKey(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerLifecycleEventsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(h.service.FindWorkerLifecycleEventByIdempotencyKey(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
