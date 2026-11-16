package policy

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek tersedia bersama alias lama untuk kompatibilitas service.
func (r *PolicyRepository) InsertPolicies(ctx context.Context, value *entity.Policy) error {
	return r.Insert(ctx, value)
}
func (r *PolicyRepository) UpdatePolicies(ctx context.Context, value *entity.Policy) error {
	return r.Update(ctx, value)
}
func (r *PolicyRepository) FindPoliciesByUUID(ctx context.Context, uuid string) (*entity.Policy, error) {
	return r.FindByID(ctx, uuid)
}
func (r *PolicyRepository) FindPoliciesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.Policy, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *PolicyRepository) DeletePoliciesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *PolicyRepository) GetAllPolicies(ctx context.Context, limit, offset int) ([]entity.Policy, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PolicyRepository) CountPolicies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PolicyRepository) SelectPoliciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Policy, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PolicyRepository) CountPoliciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PolicyRepository) FindPoliciesByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByPolicyCode(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByCategoryCode(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
