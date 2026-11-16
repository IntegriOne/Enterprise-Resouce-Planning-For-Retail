package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayComponentRepository) InsertPayComponents(ctx context.Context, v *entity.PayComponent) error {
	return r.Insert(ctx, v)
}
func (r *PayComponentRepository) UpdatePayComponents(ctx context.Context, v *entity.PayComponent) error {
	return r.Update(ctx, v)
}
func (r *PayComponentRepository) DeletePayComponentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayComponentRepository) FindPayComponentsByUUID(ctx context.Context, id string) (*entity.PayComponent, error) {
	return r.FindByID(ctx, id)
}
func (r *PayComponentRepository) FindPayComponentsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayComponent, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayComponentRepository) GetAllPayComponents(ctx context.Context, limit, offset int) ([]entity.PayComponent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayComponentRepository) CountPayComponents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayComponentRepository) SelectPayComponentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayComponentRepository) CountPayComponentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayComponentRepository) FindPayComponentsByComponentCode(ctx context.Context, v string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByComponentCode(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByComponentType(ctx context.Context, v string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByComponentType(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByTaxTreatmentCode(ctx context.Context, v *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByTaxTreatmentCode(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByCurrencyCode(ctx context.Context, v *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByRuleSetID(ctx context.Context, v *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByRuleSetID(ctx, v, limit, offset)
}
func (r *PayComponentRepository) FindPayComponentsByFormulaReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.FindByFormulaReferenceID(ctx, v, limit, offset)
}
