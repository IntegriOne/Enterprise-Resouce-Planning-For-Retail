package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollResultLineRepository) InsertPayrollResultLines(ctx context.Context, v *entity.PayrollResultLine) error {
	return r.Insert(ctx, v)
}
func (r *PayrollResultLineRepository) UpdatePayrollResultLines(ctx context.Context, v *entity.PayrollResultLine) error {
	return r.Update(ctx, v)
}
func (r *PayrollResultLineRepository) DeletePayrollResultLinesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollResultLineRepository) FindPayrollResultLinesByUUID(ctx context.Context, id string) (*entity.PayrollResultLine, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollResultLineRepository) FindPayrollResultLinesByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollResultLine, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollResultLineRepository) GetAllPayrollResultLines(ctx context.Context, limit, offset int) ([]entity.PayrollResultLine, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollResultLineRepository) CountPayrollResultLines(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollResultLineRepository) SelectPayrollResultLinesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollResultLine, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollResultLineRepository) CountPayrollResultLinesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollResultLineRepository) FindPayrollResultLinesByPayrollResultID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResultLine, error) {
	return r.FindByPayrollResultID(ctx, v, limit, offset)
}
func (r *PayrollResultLineRepository) FindPayrollResultLinesByComponentID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResultLine, error) {
	return r.FindByComponentID(ctx, v, limit, offset)
}
