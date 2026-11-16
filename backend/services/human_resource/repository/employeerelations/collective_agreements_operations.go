package employeerelations

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek meneruskan ke implementasi GORM lokal yang sama.
func (r *CollectiveAgreementRepository) InsertCollectiveAgreements(ctx context.Context, value *entity.CollectiveAgreement) error {
	return r.Insert(ctx, value)
}
func (r *CollectiveAgreementRepository) UpdateCollectiveAgreements(ctx context.Context, value *entity.CollectiveAgreement) error {
	return r.Update(ctx, value)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByUUID(ctx context.Context, uuid string) (*entity.CollectiveAgreement, error) {
	return r.FindByID(ctx, uuid)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.CollectiveAgreement, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *CollectiveAgreementRepository) DeleteCollectiveAgreementsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *CollectiveAgreementRepository) GetAllCollectiveAgreements(ctx context.Context, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CollectiveAgreementRepository) CountCollectiveAgreements(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CollectiveAgreementRepository) SelectCollectiveAgreementsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CollectiveAgreementRepository) CountCollectiveAgreementsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByAgreementCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.FindByAgreementCode(ctx, value, limit, offset)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByRepresentativePartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.FindByRepresentativePartyReferenceID(ctx, value, limit, offset)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
func (r *CollectiveAgreementRepository) FindCollectiveAgreementsByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.FindByDocumentVersionReferenceID(ctx, value, limit, offset)
}
