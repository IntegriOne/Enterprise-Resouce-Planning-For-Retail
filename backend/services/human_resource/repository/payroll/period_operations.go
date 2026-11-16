package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollPeriodRepository) InsertPayrollPeriods(ctx context.Context, v *entity.PayrollPeriod) error {
	return r.Insert(ctx, v)
}
func (r *PayrollPeriodRepository) UpdatePayrollPeriods(ctx context.Context, v *entity.PayrollPeriod) error {
	return r.Update(ctx, v)
}
func (r *PayrollPeriodRepository) DeletePayrollPeriodsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByUUID(ctx context.Context, id string) (*entity.PayrollPeriod, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollPeriod, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollPeriodRepository) GetAllPayrollPeriods(ctx context.Context, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollPeriodRepository) CountPayrollPeriods(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollPeriodRepository) SelectPayrollPeriodsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollPeriodRepository) CountPayrollPeriodsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByPeriodCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByPeriodCode(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByPeriodStatus(ctx context.Context, v string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByPeriodStatus(ctx, v, limit, offset)
}
func (r *PayrollPeriodRepository) FindPayrollPeriodsByCalendarReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.FindByCalendarReferenceID(ctx, v, limit, offset)
}
