package leave

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek tetap mengikuti validasi key dan transaksi repository lokal.
func (r *LeaveAdjustmentRepository) InsertLeaveAdjustments(ctx context.Context, value *entity.LeaveAdjustment) error {
	return r.Insert(ctx, value)
}
func (r *LeaveAdjustmentRepository) UpdateLeaveAdjustments(ctx context.Context, value *entity.LeaveAdjustment) error {
	return r.Update(ctx, value)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByUUID(ctx context.Context, uuid string) (*entity.LeaveAdjustment, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeaveAdjustment, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) DeleteLeaveAdjustmentsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) GetAllLeaveAdjustments(ctx context.Context, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeaveAdjustmentRepository) CountLeaveAdjustments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeaveAdjustmentRepository) SelectLeaveAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeaveAdjustmentRepository) CountLeaveAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByEntitlementID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByRequestedByPrincipalID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}
