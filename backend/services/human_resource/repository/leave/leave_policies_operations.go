package leave

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Nama operasi menyebut tabel dan kolom; implementasi GORM berada pada repository yang sama.
func (r *LeavePolicyRepository) InsertLeavePolicies(ctx context.Context, value *entity.LeavePolicy) error {
	return r.Insert(ctx, value)
}
func (r *LeavePolicyRepository) UpdateLeavePolicies(ctx context.Context, value *entity.LeavePolicy) error {
	return r.Update(ctx, value)
}
func (r *LeavePolicyRepository) FindLeavePoliciesByUUID(ctx context.Context, uuid string) (*entity.LeavePolicy, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeavePolicyRepository) FindLeavePoliciesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeavePolicy, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeavePolicyRepository) DeleteLeavePoliciesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeavePolicyRepository) GetAllLeavePolicies(ctx context.Context, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeavePolicyRepository) CountLeavePolicies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeavePolicyRepository) SelectLeavePoliciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeavePolicyRepository) CountLeavePoliciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeavePolicyRepository) FindLeavePoliciesByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.FindByPolicyCode(ctx, value, limit, offset)
}
func (r *LeavePolicyRepository) FindLeavePoliciesByLeaveTypeCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.FindByLeaveTypeCode(ctx, value, limit, offset)
}
func (r *LeavePolicyRepository) FindLeavePoliciesByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
