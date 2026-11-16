// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package recruitment

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// OfferRepository menyediakan kueri bertipe untuk hr_offers.
type OfferRepository struct {
	db *gorm.DB
}

var _ domainrepo.OfferRepository = (*OfferRepository)(nil)

// Method bernama objek memperjelas operasi penawaran dan kolom pencarian.
func (r *OfferRepository) InsertOffers(ctx context.Context, v *entity.Offer) error {
	return r.Insert(ctx, v)
}
func (r *OfferRepository) UpdateOffers(ctx context.Context, v *entity.Offer) error {
	return r.Update(ctx, v)
}
func (r *OfferRepository) DeleteOffersByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *OfferRepository) FindOffersByUUID(ctx context.Context, id string) (*entity.Offer, error) {
	return r.FindByID(ctx, id)
}
func (r *OfferRepository) FindOffersByUUIDForUpdate(ctx context.Context, id string) (*entity.Offer, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *OfferRepository) GetAllOffers(ctx context.Context, limit, offset int) ([]entity.Offer, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OfferRepository) CountOffers(ctx context.Context) (int64, error) { return r.Count(ctx) }
func (r *OfferRepository) SelectOffersWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Offer, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *OfferRepository) CountOffersWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *OfferRepository) FindOffersByApplicationID(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByApplicationID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByOfferNumber(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByOfferNumber(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByOfferStatus(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByOfferStatus(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByProposedJobID(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByProposedJobID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByProposedLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByProposedLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByProposedContractType(ctx context.Context, v string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByProposedContractType(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByProposedCompensationDocumentVersionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByProposedCompensationDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByCurrencyCode(ctx context.Context, v *string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByOfferDocumentReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByOfferDocumentReferenceID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByOfferDocumentVersionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByOfferDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *OfferRepository) FindOffersByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Offer, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}

// NewOfferRepository memvalidasi koneksi dan menyiapkan repository.
func NewOfferRepository(db *gorm.DB) (*OfferRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OfferRepository{db: db}, nil
}

func (r *OfferRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Offer{})
}

// Insert menyimpan penawaran kerja dengan audit dan status dari service.
func (r *OfferRepository) Insert(ctx context.Context, value *entity.Offer) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *OfferRepository) Create(ctx context.Context, value *entity.Offer) error {
	return r.Insert(ctx, value)
}
func (r *OfferRepository) FindByID(ctx context.Context, key string) (*entity.Offer, error) {
	return r.find(ctx, key, false)
}
func (r *OfferRepository) SelectByID(ctx context.Context, key string) (*entity.Offer, error) {
	return r.FindByID(ctx, key)
}
func (r *OfferRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Offer, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *OfferRepository) find(ctx context.Context, key string, lock bool) (*entity.Offer, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	value := new(entity.Offer)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *OfferRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Offer, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *OfferRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Offer, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OfferRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Offer, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter penawaran sebelum pembatasan halaman.
func (r *OfferRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Offer, error) {
	limit, offset, err := pgstore.NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	db, err = pgstore.ApplyFilters(db, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]entity.Offer, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *OfferRepository) Count(ctx context.Context) (int64, error) { return r.CountWhere(ctx, nil) }
func (r *OfferRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return 0, err
	}
	db, err = pgstore.ApplyFilters(db, filters)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Count(&total).Error
	return total, pgstore.MapError(err)
}

// Update mengganti snapshot penawaran non-key tanpa upsert atau penggantian ID.
func (r *OfferRepository) Update(ctx context.Context, value *entity.Offer) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(value.ID); err != nil {
		return err
	}
	result := db.Where("id = ?", value.ID).Select("*").Omit("id").Updates(value)
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OfferRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Offer{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OfferRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByApplicationID mencari kecocokan application_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "application_id", value, limit, offset)
}

// FindByOfferNumber mencari kecocokan offer_number dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByOfferNumber(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "offer_number", value, limit, offset)
}

// FindByOfferStatus mencari kecocokan offer_status dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByOfferStatus(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "offer_status", value, limit, offset)
}

// FindByProposedJobID mencari kecocokan proposed_job_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByProposedJobID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "proposed_job_id", value, limit, offset)
}

// FindByProposedLegalEntityReferenceID mencari kecocokan proposed_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByProposedLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "proposed_legal_entity_reference_id", value, limit, offset)
}

// FindByProposedContractType mencari kecocokan proposed_contract_type dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByProposedContractType(ctx context.Context, value string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "proposed_contract_type", value, limit, offset)
}

// FindByProposedCompensationDocumentVersionReferenceID mencari kecocokan proposed_compensation_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByProposedCompensationDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "proposed_compensation_document_version_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByOfferDocumentReferenceID mencari kecocokan offer_document_reference_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByOfferDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "offer_document_reference_id", value, limit, offset)
}

// FindByOfferDocumentVersionReferenceID mencari kecocokan offer_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByOfferDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "offer_document_version_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *OfferRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Offer, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
