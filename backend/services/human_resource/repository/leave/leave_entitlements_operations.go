package leave

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek mempertahankan kontrak lama tanpa embedding.
func (r *LeaveEntitlementRepository) InsertLeaveEntitlements(ctx context.Context, value *entity.LeaveEntitlement) error {
	return r.Insert(ctx, value)
}
func (r *LeaveEntitlementRepository) UpdateLeaveEntitlements(ctx context.Context, value *entity.LeaveEntitlement) error {
	return r.Update(ctx, value)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsByUUID(ctx context.Context, uuid string) (*entity.LeaveEntitlement, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeaveEntitlement, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeaveEntitlementRepository) DeleteLeaveEntitlementsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeaveEntitlementRepository) GetAllLeaveEntitlements(ctx context.Context, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeaveEntitlementRepository) CountLeaveEntitlements(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeaveEntitlementRepository) SelectLeaveEntitlementsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeaveEntitlementRepository) CountLeaveEntitlementsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsByLeavePolicyID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.FindByLeavePolicyID(ctx, value, limit, offset)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsByUOMCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.FindByUOMCode(ctx, value, limit, offset)
}
func (r *LeaveEntitlementRepository) FindLeaveEntitlementsBySourceWorkforceSnapshotReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.FindBySourceWorkforceSnapshotReferenceID(ctx, value, limit, offset)
}
