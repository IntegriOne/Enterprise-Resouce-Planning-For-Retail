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

// TalentDimensionRepository menyediakan kueri bertipe untuk hr_talent_dimensions.
type TalentDimensionRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentDimensionRepository = (*TalentDimensionRepository)(nil)

// NewTalentDimensionRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentDimensionRepository(db *gorm.DB) (*TalentDimensionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentDimensionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak dimensi yang eksplisit.
func (r *TalentDimensionRepository) InsertTalentDimensions(ctx context.Context, value *entity.TalentDimension) error {
	return r.Insert(ctx, value)
}
func (r *TalentDimensionRepository) UpdateTalentDimensions(ctx context.Context, value *entity.TalentDimension) error {
	return r.Update(ctx, value)
}
func (r *TalentDimensionRepository) DeleteTalentDimensionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentDimensionRepository) FindTalentDimensionsByUUID(ctx context.Context, id string) (*entity.TalentDimension, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentDimensionRepository) FindTalentDimensionsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentDimension, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentDimensionRepository) GetAllTalentDimensions(ctx context.Context, limit, offset int) ([]entity.TalentDimension, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentDimensionRepository) CountTalentDimensions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentDimensionRepository) SelectTalentDimensionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentDimension, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentDimensionRepository) CountTalentDimensionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentDimensionRepository) FindTalentDimensionsByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error) {
	return r.FindByFrameworkID(ctx, value, limit, offset)
}
func (r *TalentDimensionRepository) FindTalentDimensionsByDimensionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error) {
	return r.FindByDimensionCode(ctx, value, limit, offset)
}

func (r *TalentDimensionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentDimension{})
}

// Insert menyimpan snapshot dimensi.
func (r *TalentDimensionRepository) Insert(ctx context.Context, value *entity.TalentDimension) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentDimensionRepository) Create(ctx context.Context, value *entity.TalentDimension) error {
	return r.Insert(ctx, value)
}
func (r *TalentDimensionRepository) FindByID(ctx context.Context, key string) (*entity.TalentDimension, error) {
	return r.find(ctx, key, false)
}
func (r *TalentDimensionRepository) SelectByID(ctx context.Context, key string) (*entity.TalentDimension, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentDimensionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentDimension, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentDimensionRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentDimension, error) {
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
	var value entity.TalentDimension
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentDimensionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentDimension, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentDimensionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentDimension, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentDimensionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentDimension, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum urutan dan pagination.
func (r *TalentDimensionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentDimension, error) {
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
	rows := make([]entity.TalentDimension, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentDimensionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentDimensionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh kolom non-key, termasuk nol/NULL, tanpa upsert.
func (r *TalentDimensionRepository) Update(ctx context.Context, value *entity.TalentDimension) error {
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
func (r *TalentDimensionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentDimension{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentDimensionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByFrameworkID mencari kecocokan framework_id dengan paginasi dan urutan primary key.
func (r *TalentDimensionRepository) FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error) {
	return r.SelectBy(ctx, "framework_id", value, limit, offset)
}

// FindByDimensionCode mencari kecocokan dimension_code dengan paginasi dan urutan primary key.
func (r *TalentDimensionRepository) FindByDimensionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentDimension, error) {
	return r.SelectBy(ctx, "dimension_code", value, limit, offset)
}
