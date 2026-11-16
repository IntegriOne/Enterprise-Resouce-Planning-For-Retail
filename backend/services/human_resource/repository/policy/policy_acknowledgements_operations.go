package policy

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek tidak memintas validasi atau transaksi repository lokal.
func (r *PolicyAcknowledgementRepository) InsertPolicyAcknowledgements(ctx context.Context, value *entity.PolicyAcknowledgement) error {
	return r.Insert(ctx, value)
}
func (r *PolicyAcknowledgementRepository) UpdatePolicyAcknowledgements(ctx context.Context, value *entity.PolicyAcknowledgement) error {
	return r.Update(ctx, value)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByUUID(ctx context.Context, uuid string) (*entity.PolicyAcknowledgement, error) {
	return r.FindByID(ctx, uuid)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.PolicyAcknowledgement, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *PolicyAcknowledgementRepository) DeletePolicyAcknowledgementsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *PolicyAcknowledgementRepository) GetAllPolicyAcknowledgements(ctx context.Context, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PolicyAcknowledgementRepository) CountPolicyAcknowledgements(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PolicyAcknowledgementRepository) SelectPolicyAcknowledgementsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PolicyAcknowledgementRepository) CountPolicyAcknowledgementsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByPolicyID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.FindByPolicyID(ctx, value, limit, offset)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByAcknowledgedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.FindByAcknowledgedByPrincipalID(ctx, value, limit, offset)
}
func (r *PolicyAcknowledgementRepository) FindPolicyAcknowledgementsByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}
