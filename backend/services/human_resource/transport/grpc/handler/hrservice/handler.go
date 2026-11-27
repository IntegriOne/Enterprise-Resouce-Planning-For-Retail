// Package hrservice memetakan permintaan layanan HR dan riwayatnya ke RPC bertipe.
package hrservice

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/hrservice/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/hrservice"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	// Kontrak privat wajib memiliki implementasi eksplisit untuk setiap RPC.
	pb.UnsafeHRServiceServer
	service *service.Service
}

// Register mengikat handler pada service yang sudah memiliki dependency dan policy wajib.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterHRServiceServer(registrar, &server{service: svc})
	return nil
}
func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func decision(value *string) *entity.ApprovalDecision {
	if value == nil {
		return nil
	}
	result := entity.ApprovalDecision(*value)
	return &result
}
func decisionString(value *entity.ApprovalDecision) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}
func record(value *entity.ServiceRequest) *pb.ServiceRequest {
	if value == nil {
		return nil
	}
	return &pb.ServiceRequest{
		Id: value.ID, RequestNumber: value.RequestNumber, WorkerId: value.WorkerID, RequestType: value.RequestType,
		RequestedByPrincipalId: value.RequestedByPrincipalID, AssignedPrincipalId: value.AssignedPrincipalID, Status: value.Status,
		RestrictedDetailsDocumentVersionReferenceId: value.RestrictedDetailsDocumentVersionReferenceID,
		OpenedAt: timestamppb.New(value.OpenedAt), DueAt: timestamp(value.DueAt), ResolvedAt: timestamp(value.ResolvedAt),
		ResolutionDocumentVersionReferenceId: value.ResolutionDocumentVersionReferenceID, IdempotencyKey: value.IdempotencyKey,
	}
}
func event(value *entity.ServiceRequestEvent) *pb.ServiceRequestEvent {
	if value == nil {
		return nil
	}
	return &pb.ServiceRequestEvent{
		Id: value.ID, ServiceRequestId: value.ServiceRequestID, EventType: value.EventType, Decision: decisionString(value.Decision),
		ActorPrincipalId: value.ActorPrincipalID, OccurredAt: timestamppb.New(value.OccurredAt), EvidenceDocumentVersionReferenceId: value.EvidenceDocumentVersionReferenceID, IdempotencyKey: value.IdempotencyKey,
	}
}
func recordResponse(value *entity.ServiceRequest, err error) (*pb.ServiceRequestResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ServiceRequestResponse{Value: record(value)}, nil
}
func eventResponse(value *entity.ServiceRequestEvent, err error) (*pb.ServiceRequestEventResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ServiceRequestEventResponse{Value: event(value)}, nil
}
func records(page operation.Page[entity.ServiceRequest], err error) (*pb.ServiceRequestsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.ServiceRequest, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, record(&page.Items[index]))
	}
	return &pb.ServiceRequestsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func events(page operation.Page[entity.ServiceRequestEvent], err error) (*pb.ServiceRequestEventsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.ServiceRequestEvent, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, event(&page.Items[index]))
	}
	return &pb.ServiceRequestEventsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (s *server) CreateServiceRequest(ctx context.Context, request *pb.CreateServiceRequestRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	resolved, err := response.OptionalTimestamp(in.ResolvedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return recordResponse(s.service.CreateServiceRequest(ctx, service.CreateServiceRequestInput{
		RequestNumber: in.GetRequestNumber(), WorkerID: in.GetWorkerId(), RequestType: in.GetRequestType(), AssignedPrincipalID: in.AssignedPrincipalId,
		Status: in.Status, RestrictedDetailsDocumentVersionReferenceID: in.GetRestrictedDetailsDocumentVersionReferenceId(), DueAt: due, ResolvedAt: resolved,
		ResolutionDocumentVersionReferenceID: in.ResolutionDocumentVersionReferenceId, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (s *server) UpdateServiceRequest(ctx context.Context, request *pb.UpdateServiceRequestRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"request_number", "worker_id", "request_type", "assigned_principal_id", "status", "restricted_details_document_version_reference_id", "due_at", "resolved_at", "resolution_document_version_reference_id", "idempotency_key",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	due, err := timeField(selected["due_at"], in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	resolved, err := timeField(selected["resolved_at"], in.ResolvedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return recordResponse(s.service.UpdateServiceRequest(ctx, request.Id, service.UpdateServiceRequestInput{
		RequestNumber: operation.Field[string]{Set: selected["request_number"], Value: in.RequestNumber}, WorkerID: operation.Field[string]{Set: selected["worker_id"], Value: in.WorkerId},
		RequestType: operation.Field[string]{Set: selected["request_type"], Value: in.RequestType}, AssignedPrincipalID: operation.Field[string]{Set: selected["assigned_principal_id"], Value: in.AssignedPrincipalId},
		Status: operation.Field[string]{Set: selected["status"], Value: in.Status}, RestrictedDetailsDocumentVersionReferenceID: operation.Field[string]{Set: selected["restricted_details_document_version_reference_id"], Value: in.RestrictedDetailsDocumentVersionReferenceId},
		DueAt: due, ResolvedAt: resolved, ResolutionDocumentVersionReferenceID: operation.Field[string]{Set: selected["resolution_document_version_reference_id"], Value: in.ResolutionDocumentVersionReferenceId}, IdempotencyKey: operation.Field[string]{Set: selected["idempotency_key"], Value: in.IdempotencyKey},
	}))
}
func (s *server) DeleteServiceRequest(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteServiceRequest(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetServiceRequestByID(ctx context.Context, request *commonv1.IDRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(s.service.GetServiceRequestByID(ctx, request.Id))
}
func (s *server) GetAllServiceRequests(ctx context.Context, request *commonv1.PageRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.GetAllServiceRequests(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByRequestNumber(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByRequestNumber(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByWorkerID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByWorkerID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByRequestType(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByRequestType(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByRequestedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByRequestedByPrincipalID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByAssignedPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByAssignedPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByStatus(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByStatus(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByResolutionDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByResolutionDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestByIdempotencyKey(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return records(s.service.FindServiceRequestByIdempotencyKey(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}

func (s *server) CreateServiceRequestEvent(ctx context.Context, request *pb.CreateServiceRequestEventRequest) (*pb.ServiceRequestEventResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	return eventResponse(s.service.CreateServiceRequestEvent(ctx, service.CreateServiceRequestEventInput{
		ServiceRequestID: in.GetServiceRequestId(), EventType: in.GetEventType(), Decision: decision(in.Decision), EvidenceDocumentVersionReferenceID: in.EvidenceDocumentVersionReferenceId, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (s *server) UpdateServiceRequestEvent(ctx context.Context, request *pb.UpdateServiceRequestEventRequest) (*pb.ServiceRequestEventResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask, "service_request_id", "event_type", "decision", "evidence_document_version_reference_id", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	return eventResponse(s.service.UpdateServiceRequestEvent(ctx, request.Id, service.UpdateServiceRequestEventInput{
		ServiceRequestID: operation.Field[string]{Set: selected["service_request_id"], Value: in.ServiceRequestId}, EventType: operation.Field[string]{Set: selected["event_type"], Value: in.EventType},
		Decision: operation.Field[entity.ApprovalDecision]{Set: selected["decision"], Value: decision(in.Decision)}, EvidenceDocumentVersionReferenceID: operation.Field[string]{Set: selected["evidence_document_version_reference_id"], Value: in.EvidenceDocumentVersionReferenceId}, IdempotencyKey: operation.Field[string]{Set: selected["idempotency_key"], Value: in.IdempotencyKey},
	}))
}
func (s *server) DeleteServiceRequestEvent(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteServiceRequestEvent(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetServiceRequestEventByID(ctx context.Context, request *commonv1.IDRequest) (*pb.ServiceRequestEventResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return eventResponse(s.service.GetServiceRequestEventByID(ctx, request.Id))
}
func (s *server) GetAllServiceRequestEvents(ctx context.Context, request *commonv1.PageRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.GetAllServiceRequestEvents(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestEventByServiceRequestID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindServiceRequestEventByServiceRequestID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestEventByEventType(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindServiceRequestEventByEventType(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestEventByActorPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindServiceRequestEventByActorPrincipalID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestEventByEvidenceDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindServiceRequestEventByEvidenceDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindServiceRequestEventByIdempotencyKey(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.ServiceRequestEventsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return events(s.service.FindServiceRequestEventByIdempotencyKey(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}

func (s *server) AssignServiceRequest(ctx context.Context, request *commonv1.IDRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(s.service.AssignServiceRequest(ctx, request.Id))
}

func (s *server) ResolveServiceRequest(ctx context.Context, request *pb.ResolveServiceRequestRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(s.service.ResolveServiceRequest(ctx, request.Id, request.ResolutionDocumentVersionReferenceId))
}

func (s *server) CloseServiceRequest(ctx context.Context, request *commonv1.IDRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(s.service.CloseServiceRequest(ctx, request.Id))
}

func (s *server) CancelServiceRequest(ctx context.Context, request *commonv1.IDRequest) (*pb.ServiceRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return recordResponse(s.service.CancelServiceRequest(ctx, request.Id))
}
