package accessrequest

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek hanya mengelola permintaan lokal, bukan grant IAM.
func (r *AccessRequestRepository) InsertAccessRequests(ctx context.Context, value *entity.AccessRequest) error {
	return r.Insert(ctx, value)
}
func (r *AccessRequestRepository) UpdateAccessRequests(ctx context.Context, value *entity.AccessRequest) error {
	return r.Update(ctx, value)
}
func (r *AccessRequestRepository) FindAccessRequestsByUUID(ctx context.Context, uuid string) (*entity.AccessRequest, error) {
	return r.FindByID(ctx, uuid)
}
func (r *AccessRequestRepository) FindAccessRequestsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.AccessRequest, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *AccessRequestRepository) DeleteAccessRequestsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *AccessRequestRepository) GetAllAccessRequests(ctx context.Context, limit, offset int) ([]entity.AccessRequest, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *AccessRequestRepository) CountAccessRequests(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *AccessRequestRepository) SelectAccessRequestsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *AccessRequestRepository) CountAccessRequestsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *AccessRequestRepository) FindAccessRequestsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByLifecycleEventID(ctx context.Context, value int64, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByLifecycleEventID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByRequestedScopeType(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByRequestedScopeType(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByRequestedScopeReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByRequestedScopeReferenceID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByApprovedCatalogReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByApprovedCatalogReferenceID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByIAMRequestReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByIAMRequestReferenceID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByIAMDecisionEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByIAMDecisionEventReferenceID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByRequestedByPrincipalID(ctx, value, limit, offset)
}
func (r *AccessRequestRepository) FindAccessRequestsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}
