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

// JobCompetencyRepository menyediakan kueri bertipe untuk hr_job_competencies.
type JobCompetencyRepository struct {
	db *gorm.DB
}

var _ domainrepo.JobCompetencyRepository = (*JobCompetencyRepository)(nil)

// NewJobCompetencyRepository memvalidasi koneksi dan menyiapkan repository.
func NewJobCompetencyRepository(db *gorm.DB) (*JobCompetencyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &JobCompetencyRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak kebutuhan kompetensi yang eksplisit.
func (r *JobCompetencyRepository) InsertJobCompetencies(ctx context.Context, value *entity.JobCompetency) error {
	return r.Insert(ctx, value)
}
func (r *JobCompetencyRepository) UpdateJobCompetencies(ctx context.Context, value *entity.JobCompetency) error {
	return r.Update(ctx, value)
}
func (r *JobCompetencyRepository) DeleteJobCompetenciesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *JobCompetencyRepository) FindJobCompetenciesByUUID(ctx context.Context, id string) (*entity.JobCompetency, error) {
	return r.FindByID(ctx, id)
}
func (r *JobCompetencyRepository) FindJobCompetenciesByUUIDForUpdate(ctx context.Context, id string) (*entity.JobCompetency, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *JobCompetencyRepository) GetAllJobCompetencies(ctx context.Context, limit, offset int) ([]entity.JobCompetency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobCompetencyRepository) CountJobCompetencies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *JobCompetencyRepository) SelectJobCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.JobCompetency, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *JobCompetencyRepository) CountJobCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *JobCompetencyRepository) FindJobCompetenciesByJobID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error) {
	return r.FindByJobID(ctx, value, limit, offset)
}
func (r *JobCompetencyRepository) FindJobCompetenciesByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error) {
	return r.FindByCompetencyID(ctx, value, limit, offset)
}

func (r *JobCompetencyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.JobCompetency{})
}

// Insert menyimpan kebutuhan kompetensi sesuai constraint database.
func (r *JobCompetencyRepository) Insert(ctx context.Context, value *entity.JobCompetency) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *JobCompetencyRepository) Create(ctx context.Context, value *entity.JobCompetency) error {
	return r.Insert(ctx, value)
}
func (r *JobCompetencyRepository) FindByID(ctx context.Context, key string) (*entity.JobCompetency, error) {
	return r.find(ctx, key, false)
}
func (r *JobCompetencyRepository) SelectByID(ctx context.Context, key string) (*entity.JobCompetency, error) {
	return r.FindByID(ctx, key)
}
func (r *JobCompetencyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.JobCompetency, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *JobCompetencyRepository) find(ctx context.Context, key string, lock bool) (*entity.JobCompetency, error) {
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
	var value entity.JobCompetency
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *JobCompetencyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.JobCompetency, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *JobCompetencyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.JobCompetency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobCompetencyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.JobCompetency, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *JobCompetencyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.JobCompetency, error) {
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
	rows := make([]entity.JobCompetency, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *JobCompetencyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *JobCompetencyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *JobCompetencyRepository) Update(ctx context.Context, value *entity.JobCompetency) error {
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
func (r *JobCompetencyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.JobCompetency{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *JobCompetencyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByJobID mencari kecocokan job_id dengan paginasi dan urutan primary key.
func (r *JobCompetencyRepository) FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error) {
	return r.SelectBy(ctx, "job_id", value, limit, offset)
}

// FindByCompetencyID mencari kecocokan competency_id dengan paginasi dan urutan primary key.
func (r *JobCompetencyRepository) FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.JobCompetency, error) {
	return r.SelectBy(ctx, "competency_id", value, limit, offset)
}
