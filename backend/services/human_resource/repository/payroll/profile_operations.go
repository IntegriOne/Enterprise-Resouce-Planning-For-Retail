package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollProfileRepository) InsertPayrollProfiles(ctx context.Context, v *entity.PayrollProfile) error {
	return r.Insert(ctx, v)
}
func (r *PayrollProfileRepository) UpdatePayrollProfiles(ctx context.Context, v *entity.PayrollProfile) error {
	return r.Update(ctx, v)
}
func (r *PayrollProfileRepository) DeletePayrollProfilesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByUUID(ctx context.Context, id string) (*entity.PayrollProfile, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollProfile, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollProfileRepository) GetAllPayrollProfiles(ctx context.Context, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollProfileRepository) CountPayrollProfiles(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollProfileRepository) SelectPayrollProfilesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollProfileRepository) CountPayrollProfilesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollCountryCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollCountryCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByTaxProfileReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByTaxProfileReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByBankAccountTokenReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByBankAccountTokenReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByTaxStatusCode(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByTaxStatusCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByWithholdingElectionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByWithholdingElectionReferenceID(ctx, v, limit, offset)
}
