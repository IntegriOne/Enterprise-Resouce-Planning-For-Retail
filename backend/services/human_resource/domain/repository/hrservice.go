// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// ServiceRequestRepository adalah port persistensi hr_service_requests.
type ServiceRequestRepository interface {
	Repository[entity.ServiceRequest, string]
	InsertServiceRequests(context.Context, *entity.ServiceRequest) error
	UpdateServiceRequests(context.Context, *entity.ServiceRequest) error
	DeleteServiceRequestsByUUID(context.Context, string) error
	FindServiceRequestsByUUID(context.Context, string) (*entity.ServiceRequest, error)
	FindServiceRequestsByUUIDForUpdate(context.Context, string) (*entity.ServiceRequest, error)
	GetAllServiceRequests(context.Context, int, int) ([]entity.ServiceRequest, error)
	CountServiceRequests(context.Context) (int64, error)
	SelectServiceRequestsWhere(context.Context, []Filter, int, int) ([]entity.ServiceRequest, error)
	CountServiceRequestsWhere(context.Context, []Filter) (int64, error)
	FindServiceRequestsByRequestNumber(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByWorkerID(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByRequestType(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByRequestedByPrincipalID(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByAssignedPrincipalID(context.Context, *string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByStatus(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByRestrictedDetailsDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByResolutionDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.ServiceRequest, error)
	FindServiceRequestsByIdempotencyKey(context.Context, string, int, int) ([]entity.ServiceRequest, error)
	FindByRequestNumber(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByRequestType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByRestrictedDetailsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByResolutionDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error)
}

// ServiceRequestEventRepository adalah port persistensi hr_service_request_events.
type ServiceRequestEventRepository interface {
	Repository[entity.ServiceRequestEvent, string]
	InsertServiceRequestEvents(context.Context, *entity.ServiceRequestEvent) error
	UpdateServiceRequestEvents(context.Context, *entity.ServiceRequestEvent) error
	DeleteServiceRequestEventsByUUID(context.Context, string) error
	FindServiceRequestEventsByUUID(context.Context, string) (*entity.ServiceRequestEvent, error)
	FindServiceRequestEventsByUUIDForUpdate(context.Context, string) (*entity.ServiceRequestEvent, error)
	GetAllServiceRequestEvents(context.Context, int, int) ([]entity.ServiceRequestEvent, error)
	CountServiceRequestEvents(context.Context) (int64, error)
	SelectServiceRequestEventsWhere(context.Context, []Filter, int, int) ([]entity.ServiceRequestEvent, error)
	CountServiceRequestEventsWhere(context.Context, []Filter) (int64, error)
	FindServiceRequestEventsByServiceRequestID(context.Context, string, int, int) ([]entity.ServiceRequestEvent, error)
	FindServiceRequestEventsByEventType(context.Context, string, int, int) ([]entity.ServiceRequestEvent, error)
	FindServiceRequestEventsByActorPrincipalID(context.Context, string, int, int) ([]entity.ServiceRequestEvent, error)
	FindServiceRequestEventsByEvidenceDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.ServiceRequestEvent, error)
	FindServiceRequestEventsByIdempotencyKey(context.Context, string, int, int) ([]entity.ServiceRequestEvent, error)
	FindByServiceRequestID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error)
	FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error)
	FindByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequestEvent, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error)
}
