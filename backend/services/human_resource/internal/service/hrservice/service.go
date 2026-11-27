// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	ResolutionVerifier  ResolutionVerifier
	Runtime             operation.Dependencies
	ServiceRequest      domainrepo.RecordRepository[entity.ServiceRequest, string]
	ServiceRequestEvent domainrepo.RecordRepository[entity.ServiceRequestEvent, string]
}

// Service mengoordinasikan operasi domain hrservice.
type Service struct {
	resolution          ResolutionVerifier
	serviceRequest      *operation.Resource[entity.ServiceRequest, string, *entity.ServiceRequest]
	serviceRequestEvent *operation.Resource[entity.ServiceRequestEvent, string, *entity.ServiceRequestEvent]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{resolution: deps.ResolutionVerifier}
	var err error
	s.serviceRequest, err = operation.NewResource[entity.ServiceRequest, string, *entity.ServiceRequest](
		deps.ServiceRequest,
		deps.Runtime,
		specServiceRequest(),
	)
	if err != nil {
		return nil, err
	}
	s.serviceRequestEvent, err = operation.NewResource[entity.ServiceRequestEvent, string, *entity.ServiceRequestEvent](
		deps.ServiceRequestEvent,
		deps.Runtime,
		specServiceRequestEvent(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateServiceRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateServiceRequest(ctx context.Context, in CreateServiceRequestInput) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Create(ctx, func() (*entity.ServiceRequest, error) { return createServiceRequest(in) })
}

// UpdateServiceRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateServiceRequest(ctx context.Context, key string, in UpdateServiceRequestInput) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Update(ctx, key, func(value *entity.ServiceRequest) error { return patchServiceRequest(value, in) })
}

// DeleteServiceRequest menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteServiceRequest(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Delete(ctx, key)
}

// GetServiceRequestByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetServiceRequestByID(ctx context.Context, key string) (*entity.ServiceRequest, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.Get(ctx, key)
}

// GetAllServiceRequests menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllServiceRequests(ctx context.Context, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, nil, limit, offset)
}

// FindServiceRequestByRequestNumber menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByRequestNumber(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "request_number", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByRequestType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByRequestType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "request_type", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByRequestedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "requested_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByAssignedPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "assigned_principal_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "status", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "restricted_details_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByResolutionDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByResolutionDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "resolution_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequest], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequest]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequest.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateServiceRequestEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateServiceRequestEvent(ctx context.Context, in CreateServiceRequestEventInput) (*entity.ServiceRequestEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.Create(ctx, func() (*entity.ServiceRequestEvent, error) { return createServiceRequestEvent(in) })
}

// UpdateServiceRequestEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateServiceRequestEvent(ctx context.Context, key string, in UpdateServiceRequestEventInput) (*entity.ServiceRequestEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.Update(ctx, key, func(value *entity.ServiceRequestEvent) error { return patchServiceRequestEvent(value, in) })
}

// DeleteServiceRequestEvent menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteServiceRequestEvent(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.Delete(ctx, key)
}

// GetServiceRequestEventByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetServiceRequestEventByID(ctx context.Context, key string) (*entity.ServiceRequestEvent, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.Get(ctx, key)
}

// GetAllServiceRequestEvents menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllServiceRequestEvents(ctx context.Context, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, nil, limit, offset)
}

// FindServiceRequestEventByServiceRequestID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestEventByServiceRequestID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, []domainrepo.Filter{{Column: "service_request_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestEventByEventType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestEventByEventType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, []domainrepo.Filter{{Column: "event_type", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestEventByActorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestEventByActorPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, []domainrepo.Filter{{Column: "actor_principal_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestEventByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestEventByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindServiceRequestEventByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindServiceRequestEventByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.ServiceRequestEvent], error) {
	if s == nil {
		return operation.Page[entity.ServiceRequestEvent]{}, hrerrors.ErrUnavailable
	}
	return s.serviceRequestEvent.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}
