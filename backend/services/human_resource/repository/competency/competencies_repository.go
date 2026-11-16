// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package competency

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// CompetencyRepository menyediakan kueri bertipe untuk hr_competencies.
type CompetencyRepository struct {
	db *gorm.DB
}

var _ domainrepo.CompetencyRepository = (*CompetencyRepository)(nil)

// NewCompetencyRepository memvalidasi koneksi dan menyiapkan repository.
func NewCompetencyRepository(db *gorm.DB) (*CompetencyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CompetencyRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak kompetensi yang eksplisit.
func (r *CompetencyRepository) InsertCompetencies(ctx context.Context, value *entity.Competency) error {
	return r.Insert(ctx, value)
}
func (r *CompetencyRepository) UpdateCompetencies(ctx context.Context, value *entity.Competency) error {
	return r.Update(ctx, value)
}
func (r *CompetencyRepository) DeleteCompetenciesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CompetencyRepository) FindCompetenciesByUUID(ctx context.Context, id string) (*entity.Competency, error) {
	return r.FindByID(ctx, id)
}
func (r *CompetencyRepository) FindCompetenciesByUUIDForUpdate(ctx context.Context, id string) (*entity.Competency, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CompetencyRepository) GetAllCompetencies(ctx context.Context, limit, offset int) ([]entity.Competency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompetencyRepository) CountCompetencies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CompetencyRepository) SelectCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Competency, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CompetencyRepository) CountCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CompetencyRepository) FindCompetenciesByCompetencyCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error) {
	return r.FindByCompetencyCode(ctx, value, limit, offset)
}
func (r *CompetencyRepository) FindCompetenciesByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error) {
	return r.FindByCategoryCode(ctx, value, limit, offset)
}

func (r *CompetencyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Competency{})
}

// Insert menyimpan kompetensi sesuai constraint database.
func (r *CompetencyRepository) Insert(ctx context.Context, value *entity.Competency) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *CompetencyRepository) Create(ctx context.Context, value *entity.Competency) error {
	return r.Insert(ctx, value)
}
func (r *CompetencyRepository) FindByID(ctx context.Context, key string) (*entity.Competency, error) {
	return r.find(ctx, key, false)
}
func (r *CompetencyRepository) SelectByID(ctx context.Context, key string) (*entity.Competency, error) {
	return r.FindByID(ctx, key)
}
func (r *CompetencyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Competency, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *CompetencyRepository) find(ctx context.Context, key string, lock bool) (*entity.Competency, error) {
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
	var value entity.Competency
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *CompetencyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Competency, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *CompetencyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Competency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompetencyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Competency, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *CompetencyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Competency, error) {
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
	rows := make([]entity.Competency, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *CompetencyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *CompetencyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *CompetencyRepository) Update(ctx context.Context, value *entity.Competency) error {
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
func (r *CompetencyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Competency{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *CompetencyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCompetencyCode mencari kecocokan competency_code dengan paginasi dan urutan primary key.
func (r *CompetencyRepository) FindByCompetencyCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error) {
	return r.SelectBy(ctx, "competency_code", value, limit, offset)
}

// FindByCategoryCode mencari kecocokan category_code dengan paginasi dan urutan primary key.
func (r *CompetencyRepository) FindByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Competency, error) {
	return r.SelectBy(ctx, "category_code", value, limit, offset)
}
