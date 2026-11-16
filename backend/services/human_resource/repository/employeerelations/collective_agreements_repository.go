package employeerelations

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// CollectiveAgreementRepository menyediakan kueri bertipe untuk hr_collective_agreements.
type CollectiveAgreementRepository struct {
	db *gorm.DB
}

var _ domainrepo.CollectiveAgreementRepository = (*CollectiveAgreementRepository)(nil)

// NewCollectiveAgreementRepository memvalidasi koneksi dan menyiapkan repository.
func NewCollectiveAgreementRepository(db *gorm.DB) (*CollectiveAgreementRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CollectiveAgreementRepository{db: db}, nil
}

func (r *CollectiveAgreementRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.CollectiveAgreement{})
}

// Insert menyimpan satu perjanjian bersama.
func (r *CollectiveAgreementRepository) Insert(ctx context.Context, value *entity.CollectiveAgreement) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create mempertahankan alias INSERT.
func (r *CollectiveAgreementRepository) Create(ctx context.Context, value *entity.CollectiveAgreement) error {
	return r.Insert(ctx, value)
}

// FindByID membaca perjanjian berdasarkan ID.
func (r *CollectiveAgreementRepository) FindByID(ctx context.Context, key string) (*entity.CollectiveAgreement, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *CollectiveAgreementRepository) SelectByID(ctx context.Context, key string) (*entity.CollectiveAgreement, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengambil row lock dalam transaksi aktif.
func (r *CollectiveAgreementRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.CollectiveAgreement, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *CollectiveAgreementRepository) find(ctx context.Context, key string, lock bool) (*entity.CollectiveAgreement, error) {
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
	value := &entity.CollectiveAgreement{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca perjanjian secara berhalaman.
func (r *CollectiveAgreementRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *CollectiveAgreementRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memakai kolom model yang diverifikasi dan nilai parameter terikat.
func (r *CollectiveAgreementRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum limit/offset.
func (r *CollectiveAgreementRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CollectiveAgreement, error) {
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
	rows := make([]entity.CollectiveAgreement, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh perjanjian.
func (r *CollectiveAgreementRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung perjanjian yang memenuhi filter.
func (r *CollectiveAgreementRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key lengkap tanpa mengubah primary key atau melakukan upsert.
func (r *CollectiveAgreementRepository) Update(ctx context.Context, value *entity.CollectiveAgreement) error {
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

// DeleteByID menghapus satu perjanjian sesuai constraint database.
func (r *CollectiveAgreementRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.CollectiveAgreement{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *CollectiveAgreementRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByAgreementCode mencari kecocokan agreement_code dengan paginasi dan urutan primary key.
func (r *CollectiveAgreementRepository) FindByAgreementCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectBy(ctx, "agreement_code", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *CollectiveAgreementRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByRepresentativePartyReferenceID mencari kecocokan representative_party_reference_id dengan paginasi dan urutan primary key.
func (r *CollectiveAgreementRepository) FindByRepresentativePartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectBy(ctx, "representative_party_reference_id", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *CollectiveAgreementRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByDocumentVersionReferenceID mencari kecocokan document_version_reference_id dengan paginasi dan urutan primary key.
func (r *CollectiveAgreementRepository) FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.CollectiveAgreement, error) {
	return r.SelectBy(ctx, "document_version_reference_id", value, limit, offset)
}
