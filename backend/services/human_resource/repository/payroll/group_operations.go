package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollGroupRepository) InsertPayrollGroups(ctx context.Context, v *entity.PayrollGroup) error {
	return r.Insert(ctx, v)
}
func (r *PayrollGroupRepository) UpdatePayrollGroups(ctx context.Context, v *entity.PayrollGroup) error {
	return r.Update(ctx, v)
}
func (r *PayrollGroupRepository) DeletePayrollGroupsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByUUID(ctx context.Context, id string) (*entity.PayrollGroup, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollGroup, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollGroupRepository) GetAllPayrollGroups(ctx context.Context, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollGroupRepository) CountPayrollGroups(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollGroupRepository) SelectPayrollGroupsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollGroupRepository) CountPayrollGroupsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByGroupCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByGroupCode(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
