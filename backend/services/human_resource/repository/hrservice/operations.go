package hrservice

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *ServiceRequestRepository) InsertServiceRequests(ctx context.Context, value *entity.ServiceRequest) error {
	return r.Insert(ctx, value)
}
func (r *ServiceRequestRepository) UpdateServiceRequests(ctx context.Context, value *entity.ServiceRequest) error {
	return r.Update(ctx, value)
}
func (r *ServiceRequestRepository) DeleteServiceRequestsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *ServiceRequestRepository) FindServiceRequestsByUUID(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	return r.FindByID(ctx, id)
}
func (r *ServiceRequestRepository) FindServiceRequestsByUUIDForUpdate(ctx context.Context, id string) (*entity.ServiceRequest, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *ServiceRequestRepository) GetAllServiceRequests(ctx context.Context, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ServiceRequestRepository) CountServiceRequests(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *ServiceRequestRepository) SelectServiceRequestsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *ServiceRequestRepository) CountServiceRequestsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *ServiceRequestRepository) FindServiceRequestsByRequestNumber(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByRequestNumber(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByRequestType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByRequestType(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByRequestedByPrincipalID(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByAssignedPrincipalID(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByRestrictedDetailsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByRestrictedDetailsDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByResolutionDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByResolutionDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *ServiceRequestRepository) FindServiceRequestsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *ServiceRequestEventRepository) InsertServiceRequestEvents(ctx context.Context, value *entity.ServiceRequestEvent) error {
	return r.Insert(ctx, value)
}
func (r *ServiceRequestEventRepository) UpdateServiceRequestEvents(ctx context.Context, value *entity.ServiceRequestEvent) error {
	return r.Update(ctx, value)
}
func (r *ServiceRequestEventRepository) DeleteServiceRequestEventsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByUUID(ctx context.Context, id string) (*entity.ServiceRequestEvent, error) {
	return r.FindByID(ctx, id)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByUUIDForUpdate(ctx context.Context, id string) (*entity.ServiceRequestEvent, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *ServiceRequestEventRepository) GetAllServiceRequestEvents(ctx context.Context, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ServiceRequestEventRepository) CountServiceRequestEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *ServiceRequestEventRepository) SelectServiceRequestEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *ServiceRequestEventRepository) CountServiceRequestEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByServiceRequestID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.FindByServiceRequestID(ctx, value, limit, offset)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.FindByActorPrincipalID(ctx, value, limit, offset)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *ServiceRequestEventRepository) FindServiceRequestEventsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}
