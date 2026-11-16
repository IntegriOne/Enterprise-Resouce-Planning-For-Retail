// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package talent

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// TalentPoolRepository menyediakan kueri bertipe untuk hr_talent_pools.
type TalentPoolRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentPoolRepository = (*TalentPoolRepository)(nil)

// NewTalentPoolRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentPoolRepository(db *gorm.DB) (*TalentPoolRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentPoolRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak pool yang eksplisit.
func (r *TalentPoolRepository) InsertTalentPools(ctx context.Context, value *entity.TalentPool) error {
	return r.Insert(ctx, value)
}
func (r *TalentPoolRepository) UpdateTalentPools(ctx context.Context, value *entity.TalentPool) error {
	return r.Update(ctx, value)
}
func (r *TalentPoolRepository) DeleteTalentPoolsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentPoolRepository) FindTalentPoolsByUUID(ctx context.Context, id string) (*entity.TalentPool, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentPoolRepository) FindTalentPoolsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentPool, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentPoolRepository) GetAllTalentPools(ctx context.Context, limit, offset int) ([]entity.TalentPool, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentPoolRepository) CountTalentPools(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentPoolRepository) SelectTalentPoolsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentPoolRepository) CountTalentPoolsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentPoolRepository) FindTalentPoolsByPoolCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.FindByPoolCode(ctx, value, limit, offset)
}
func (r *TalentPoolRepository) FindTalentPoolsByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *TalentPoolRepository) FindTalentPoolsByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPool, error) {
	return r.FindByTargetJobID(ctx, value, limit, offset)
}
func (r *TalentPoolRepository) FindTalentPoolsByEligibilityDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.FindByEligibilityDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *TalentPoolRepository) FindTalentPoolsByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.FindByOwnerPrincipalID(ctx, value, limit, offset)
}

func (r *TalentPoolRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentPool{})
}

// Insert menyimpan pool sesuai constraint database.
func (r *TalentPoolRepository) Insert(ctx context.Context, value *entity.TalentPool) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentPoolRepository) Create(ctx context.Context, value *entity.TalentPool) error {
	return r.Insert(ctx, value)
}
func (r *TalentPoolRepository) FindByID(ctx context.Context, key string) (*entity.TalentPool, error) {
	return r.find(ctx, key, false)
}
func (r *TalentPoolRepository) SelectByID(ctx context.Context, key string) (*entity.TalentPool, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentPoolRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentPool, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentPoolRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentPool, error) {
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
	var value entity.TalentPool
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentPoolRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentPoolRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentPool, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentPoolRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentPoolRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentPool, error) {
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
	rows := make([]entity.TalentPool, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentPoolRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentPoolRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key, termasuk nol/NULL, tanpa upsert.
func (r *TalentPoolRepository) Update(ctx context.Context, value *entity.TalentPool) error {
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
func (r *TalentPoolRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentPool{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentPoolRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPoolCode mencari kecocokan pool_code dengan paginasi dan urutan primary key.
func (r *TalentPoolRepository) FindByPoolCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectBy(ctx, "pool_code", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *TalentPoolRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByTargetJobID mencari kecocokan target_job_id dengan paginasi dan urutan primary key.
func (r *TalentPoolRepository) FindByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectBy(ctx, "target_job_id", value, limit, offset)
}

// FindByEligibilityDocumentVersionReferenceID mencari kecocokan eligibility_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentPoolRepository) FindByEligibilityDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectBy(ctx, "eligibility_document_version_reference_id", value, limit, offset)
}

// FindByOwnerPrincipalID mencari kecocokan owner_principal_id dengan paginasi dan urutan primary key.
func (r *TalentPoolRepository) FindByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentPool, error) {
	return r.SelectBy(ctx, "owner_principal_id", value, limit, offset)
}
