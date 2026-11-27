package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrservice "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/hrservice/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelServiceRequest(row *hrservice.ServiceRequest) (*model.ServiceRequest, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource service-request response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource service-request ID", nil)
	}
	out := &model.ServiceRequest{ID: nodeID("service-request", row.GetId()), RawID: row.GetId()}
	out.RequestNumber = row.GetRequestNumber()
	out.WorkerID = row.GetWorkerId()
	out.RequestType = row.GetRequestType()
	out.RequestedByPrincipalID = row.GetRequestedByPrincipalId()
	out.AssignedPrincipalID = row.AssignedPrincipalId
	out.Status = row.GetStatus()
	out.RestrictedDetailsDocumentVersionReferenceID = row.GetRestrictedDetailsDocumentVersionReferenceId()
	out.OpenedAt = pbTimePtr(row.GetOpenedAt())
	out.DueAt = pbTimePtr(row.GetDueAt())
	out.ResolvedAt = pbTimePtr(row.GetResolvedAt())
	out.ResolutionDocumentVersionReferenceID = row.ResolutionDocumentVersionReferenceId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// ServiceRequest membaca detail service-request berdasarkan ID.
func (a *Adapter) ServiceRequest(ctx context.Context, rawID string) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_service_requests", "human_resource.service_requests.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.GetServiceRequestByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// ServiceRequests mendaftar dan memfilter service-request dengan paginasi.
func (a *Adapter) ServiceRequests(ctx context.Context, limit, offset int, f_request_number *string, f_worker_id *string, f_request_type *string, f_requested_by_principal_id *string, f_assigned_principal_id *string, f_status *string, f_restricted_details_document_version_reference_id *string, f_resolution_document_version_reference_id *string, f_idempotency_key *string) (*Page[model.ServiceRequest], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_service_requests", "human_resource.service_requests.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid service-request pagination", nil)
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_request_number != nil, f_worker_id != nil, f_request_type != nil, f_requested_by_principal_id != nil, f_assigned_principal_id != nil, f_status != nil, f_restricted_details_document_version_reference_id != nil, f_resolution_document_version_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *hrservice.ServiceRequestsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_request_number, "request_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByRequestNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_request_type, "request_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByRequestType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_requested_by_principal_id, "requested_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByRequestedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_assigned_principal_id, "assigned_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByAssignedPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_restricted_details_document_version_reference_id, "restricted_details_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_resolution_document_version_reference_id, "resolution_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByResolutionDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.hrService.GetAllServiceRequests(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "service-request")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "service-request", toModelServiceRequest)
}

// CreateServiceRequest membuat service-request baru melalui HRM.
func (a *Adapter) CreateServiceRequest(ctx context.Context, input model.CreateServiceRequestInput) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "create", "human_resource.service_requests.create")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &hrservice.ServiceRequestInput{}
	patch.RequestNumber = input.RequestNumber
	patch.WorkerId = input.WorkerID
	patch.RequestType = input.RequestType
	patch.AssignedPrincipalId = input.AssignedPrincipalID
	patch.Status = input.Status
	patch.RestrictedDetailsDocumentVersionReferenceId = input.RestrictedDetailsDocumentVersionReferenceID
	patch.DueAt = optionalTimestamp(input.DueAt)
	patch.ResolvedAt = optionalTimestamp(input.ResolvedAt)
	patch.ResolutionDocumentVersionReferenceId = input.ResolutionDocumentVersionReferenceID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.hrService.CreateServiceRequest(rpcCtx, &hrservice.CreateServiceRequestRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// UpdateServiceRequest memperbarui service-request berdasarkan FieldMask.
func (a *Adapter) UpdateServiceRequest(ctx context.Context, rawID string, input model.UpdateServiceRequestInput) (*model.ServiceRequest, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_requests", "update", "human_resource.service_requests.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request", "service-request")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &hrservice.ServiceRequestInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "request_number", input.RequestNumber, &patch.RequestNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "request_type", input.RequestType, &patch.RequestType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_principal_id", input.AssignedPrincipalID, &patch.AssignedPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "restricted_details_document_version_reference_id", input.RestrictedDetailsDocumentVersionReferenceID, &patch.RestrictedDetailsDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "resolved_at", input.ResolvedAt, &patch.ResolvedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "resolution_document_version_reference_id", input.ResolutionDocumentVersionReferenceID, &patch.ResolutionDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("service-request update is empty", nil)
	}
	resp, err := a.hrService.UpdateServiceRequest(rpcCtx, &hrservice.UpdateServiceRequestRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request")
	}
	return toModelServiceRequest(resp.GetValue())
}

// DeleteServiceRequest menghapus service-request berdasarkan ID.
func (a *Adapter) DeleteServiceRequest(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_service_requests", "human_resource.service_requests.delete", "service-request", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.hrService.DeleteServiceRequest(c, req)
	})
}

func toModelServiceRequestEvent(row *hrservice.ServiceRequestEvent) (*model.ServiceRequestEvent, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource service-request-event response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource service-request-event ID", nil)
	}
	out := &model.ServiceRequestEvent{ID: nodeID("service-request-event", row.GetId()), RawID: row.GetId()}
	out.ServiceRequestID = row.GetServiceRequestId()
	out.EventType = row.GetEventType()
	out.Decision = row.Decision
	out.ActorPrincipalID = row.GetActorPrincipalId()
	out.OccurredAt = pbTimePtr(row.GetOccurredAt())
	out.EvidenceDocumentVersionReferenceID = row.EvidenceDocumentVersionReferenceId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// ServiceRequestEvent membaca detail service-request-event berdasarkan ID.
func (a *Adapter) ServiceRequestEvent(ctx context.Context, rawID string) (*model.ServiceRequestEvent, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_service_request_events", "human_resource.service_request_events.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request-event", "service-request-event")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.hrService.GetServiceRequestEventByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "service-request-event")
	}
	return toModelServiceRequestEvent(resp.GetValue())
}

// ServiceRequestEvents mendaftar dan memfilter service-request-event dengan paginasi.
func (a *Adapter) ServiceRequestEvents(ctx context.Context, limit, offset int, f_service_request_id *string, f_event_type *string, f_actor_principal_id *string, f_evidence_document_version_reference_id *string, f_idempotency_key *string) (*Page[model.ServiceRequestEvent], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_service_request_events", "human_resource.service_request_events.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid service-request-event pagination", nil)
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_service_request_id != nil, f_event_type != nil, f_actor_principal_id != nil, f_evidence_document_version_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *hrservice.ServiceRequestEventsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_service_request_id, "service_request_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestEventByServiceRequestID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_event_type, "event_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestEventByEventType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_actor_principal_id, "actor_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestEventByActorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestEventByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.hrService.FindServiceRequestEventByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.hrService.GetAllServiceRequestEvents(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "service-request-event")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "service-request-event", toModelServiceRequestEvent)
}

// CreateServiceRequestEvent membuat service-request-event baru melalui HRM.
func (a *Adapter) CreateServiceRequestEvent(ctx context.Context, input model.CreateServiceRequestEventInput) (*model.ServiceRequestEvent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_request_events", "create", "human_resource.service_request_events.create")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &hrservice.ServiceRequestEventInput{}
	patch.ServiceRequestId = input.ServiceRequestID
	patch.EventType = input.EventType
	patch.Decision = input.Decision
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.hrService.CreateServiceRequestEvent(rpcCtx, &hrservice.CreateServiceRequestEventRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request-event")
	}
	return toModelServiceRequestEvent(resp.GetValue())
}

// UpdateServiceRequestEvent memperbarui service-request-event berdasarkan FieldMask.
func (a *Adapter) UpdateServiceRequestEvent(ctx context.Context, rawID string, input model.UpdateServiceRequestEventInput) (*model.ServiceRequestEvent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_service_request_events", "update", "human_resource.service_request_events.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "service-request-event", "service-request-event")
	if err != nil {
		return nil, err
	}
	if a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &hrservice.ServiceRequestEventInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "service_request_id", input.ServiceRequestID, &patch.ServiceRequestId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "event_type", input.EventType, &patch.EventType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "decision", input.Decision, &patch.Decision, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("service-request-event update is empty", nil)
	}
	resp, err := a.hrService.UpdateServiceRequestEvent(rpcCtx, &hrservice.UpdateServiceRequestEventRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "service-request-event")
	}
	return toModelServiceRequestEvent(resp.GetValue())
}

// DeleteServiceRequestEvent menghapus service-request-event berdasarkan ID.
func (a *Adapter) DeleteServiceRequestEvent(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.hrService == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_service_request_events", "human_resource.service_request_events.delete", "service-request-event", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.hrService.DeleteServiceRequestEvent(c, req)
	})
}
