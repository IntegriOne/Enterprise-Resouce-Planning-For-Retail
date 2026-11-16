// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package workforceplanning

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// JobRepository menyediakan kueri bertipe untuk hr_jobs.
type JobRepository struct {
	db *gorm.DB
}

var _ domainrepo.JobRepository = (*JobRepository)(nil)

// Method bernama objek menjadi API eksplisit master jabatan HRM.
func (r *JobRepository) InsertJobs(ctx context.Context, v *entity.Job) error { return r.Insert(ctx, v) }
func (r *JobRepository) UpdateJobs(ctx context.Context, v *entity.Job) error { return r.Update(ctx, v) }
func (r *JobRepository) DeleteJobsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *JobRepository) FindJobsByUUID(ctx context.Context, id string) (*entity.Job, error) {
	return r.FindByID(ctx, id)
}
func (r *JobRepository) FindJobsByUUIDForUpdate(ctx context.Context, id string) (*entity.Job, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *JobRepository) GetAllJobs(ctx context.Context, limit, offset int) ([]entity.Job, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobRepository) CountJobs(ctx context.Context) (int64, error) { return r.Count(ctx) }
func (r *JobRepository) SelectJobsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Job, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *JobRepository) CountJobsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *JobRepository) FindJobsByJobFamilyID(ctx context.Context, v string, limit, offset int) ([]entity.Job, error) {
	return r.FindByJobFamilyID(ctx, v, limit, offset)
}
func (r *JobRepository) FindJobsByJobCode(ctx context.Context, v string, limit, offset int) ([]entity.Job, error) {
	return r.FindByJobCode(ctx, v, limit, offset)
}
func (r *JobRepository) FindJobsByJobLevelCode(ctx context.Context, v *string, limit, offset int) ([]entity.Job, error) {
	return r.FindByJobLevelCode(ctx, v, limit, offset)
}

// NewJobRepository memvalidasi koneksi dan menyiapkan repository.
func NewJobRepository(db *gorm.DB) (*JobRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &JobRepository{db: db}, nil
}

func (r *JobRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Job{})
}

// Insert menyimpan master jabatan HRM tanpa menulis master organisasi Foundation.
func (r *JobRepository) Insert(ctx context.Context, value *entity.Job) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *JobRepository) Create(ctx context.Context, value *entity.Job) error {
	return r.Insert(ctx, value)
}
func (r *JobRepository) FindByID(ctx context.Context, key string) (*entity.Job, error) {
	return r.find(ctx, key, false)
}
func (r *JobRepository) SelectByID(ctx context.Context, key string) (*entity.Job, error) {
	return r.FindByID(ctx, key)
}
func (r *JobRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Job, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *JobRepository) find(ctx context.Context, key string, lock bool) (*entity.Job, error) {
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
	value := new(entity.Job)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *JobRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Job, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *JobRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Job, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Job, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere membatasi hasil kueri sebelum limit/offset dengan urutan ID stabil.
func (r *JobRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Job, error) {
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
	rows := make([]entity.Job, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *JobRepository) Count(ctx context.Context) (int64, error) { return r.CountWhere(ctx, nil) }
func (r *JobRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot non-key lengkap tanpa upsert.
func (r *JobRepository) Update(ctx context.Context, value *entity.Job) error {
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
func (r *JobRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Job{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *JobRepository) Delete(ctx context.Context, key string) error { return r.DeleteByID(ctx, key) }

// FindByJobFamilyID mencari kecocokan job_family_id dengan paginasi dan urutan primary key.
func (r *JobRepository) FindByJobFamilyID(ctx context.Context, value string, limit, offset int) ([]entity.Job, error) {
	return r.SelectBy(ctx, "job_family_id", value, limit, offset)
}

// FindByJobCode mencari kecocokan job_code dengan paginasi dan urutan primary key.
func (r *JobRepository) FindByJobCode(ctx context.Context, value string, limit, offset int) ([]entity.Job, error) {
	return r.SelectBy(ctx, "job_code", value, limit, offset)
}

// FindByJobLevelCode mencari kecocokan job_level_code dengan paginasi dan urutan primary key.
func (r *JobRepository) FindByJobLevelCode(ctx context.Context, value *string, limit, offset int) ([]entity.Job, error) {
	return r.SelectBy(ctx, "job_level_code", value, limit, offset)
}
