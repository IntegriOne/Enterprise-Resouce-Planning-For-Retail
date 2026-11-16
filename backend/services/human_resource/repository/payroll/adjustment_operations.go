package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollAdjustmentRepository) InsertPayrollAdjustments(ctx context.Context, v *entity.PayrollAdjustment) error {
	return r.Insert(ctx, v)
}
func (r *PayrollAdjustmentRepository) UpdatePayrollAdjustments(ctx context.Context, v *entity.PayrollAdjustment) error {
	return r.Update(ctx, v)
}
func (r *PayrollAdjustmentRepository) DeletePayrollAdjustmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByUUID(ctx context.Context, id string) (*entity.PayrollAdjustment, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollAdjustment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollAdjustmentRepository) GetAllPayrollAdjustments(ctx context.Context, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollAdjustmentRepository) CountPayrollAdjustments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollAdjustmentRepository) SelectPayrollAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollAdjustmentRepository) CountPayrollAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByPayrollPeriodID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByPayrollPeriodID(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByComponentID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByComponentID(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByAdjustmentType(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByAdjustmentType(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByRequestedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByRequestedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollAdjustmentRepository) FindPayrollAdjustmentsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}
