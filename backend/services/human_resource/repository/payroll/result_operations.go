package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollResultRepository) InsertPayrollResults(ctx context.Context, v *entity.PayrollResult) error {
	return r.Insert(ctx, v)
}
func (r *PayrollResultRepository) UpdatePayrollResults(ctx context.Context, v *entity.PayrollResult) error {
	return r.Update(ctx, v)
}
func (r *PayrollResultRepository) DeletePayrollResultsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollResultRepository) FindPayrollResultsByUUID(ctx context.Context, id string) (*entity.PayrollResult, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollResultRepository) FindPayrollResultsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollResult, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollResultRepository) GetAllPayrollResults(ctx context.Context, limit, offset int) ([]entity.PayrollResult, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollResultRepository) CountPayrollResults(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollResultRepository) SelectPayrollResultsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollResultRepository) CountPayrollResultsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollResultRepository) FindPayrollResultsByPayrollRunID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByPayrollRunID(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByPayrollProfileReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByPayrollProfileReferenceID(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByResultStatus(ctx context.Context, v string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByResultStatus(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByPaymentReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByPaymentReferenceID(ctx, v, limit, offset)
}
func (r *PayrollResultRepository) FindPayrollResultsByPayslipDocumentReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.FindByPayslipDocumentReferenceID(ctx, v, limit, offset)
}
