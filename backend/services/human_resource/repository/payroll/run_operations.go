package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollRunRepository) InsertPayrollRuns(ctx context.Context, v *entity.PayrollRun) error {
	return r.Insert(ctx, v)
}
func (r *PayrollRunRepository) UpdatePayrollRuns(ctx context.Context, v *entity.PayrollRun) error {
	return r.Update(ctx, v)
}
func (r *PayrollRunRepository) DeletePayrollRunsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollRunRepository) FindPayrollRunsByUUID(ctx context.Context, id string) (*entity.PayrollRun, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollRunRepository) FindPayrollRunsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollRun, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollRunRepository) GetAllPayrollRuns(ctx context.Context, limit, offset int) ([]entity.PayrollRun, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollRunRepository) CountPayrollRuns(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollRunRepository) SelectPayrollRunsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollRunRepository) CountPayrollRunsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollRunRepository) FindPayrollRunsByPayrollPeriodID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPayrollPeriodID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRuleSetID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRuleSetID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByOriginalRunID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByOriginalRunID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRunNumber(ctx context.Context, v int32, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRunNumber(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRunType(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRunType(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByStatus(ctx context.Context, v entity.PayrollRunStatus, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByInputSnapshotReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByInputSnapshotReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByFinanceJournalReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByFinanceJournalReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByPaymentBatchReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPaymentBatchReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByFinanceSettlementEventReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByFinanceSettlementEventReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByStartedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByStartedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}
