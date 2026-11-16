package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollRuleSetRepository) InsertPayrollRuleSets(ctx context.Context, v *entity.PayrollRuleSet) error {
	return r.Insert(ctx, v)
}
func (r *PayrollRuleSetRepository) UpdatePayrollRuleSets(ctx context.Context, v *entity.PayrollRuleSet) error {
	return r.Update(ctx, v)
}
func (r *PayrollRuleSetRepository) DeletePayrollRuleSetsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByUUID(ctx context.Context, id string) (*entity.PayrollRuleSet, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollRuleSet, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollRuleSetRepository) GetAllPayrollRuleSets(ctx context.Context, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollRuleSetRepository) CountPayrollRuleSets(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollRuleSetRepository) SelectPayrollRuleSetsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollRuleSetRepository) CountPayrollRuleSetsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByRuleCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByRuleCode(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByArtifactDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByArtifactDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByLegalBasisDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByLegalBasisDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByApprovedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
