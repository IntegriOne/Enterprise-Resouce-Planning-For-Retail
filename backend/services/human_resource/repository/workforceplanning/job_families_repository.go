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

// JobFamilyRepository menyediakan kueri bertipe untuk hr_job_families.
type JobFamilyRepository struct {
	db *gorm.DB
}

var _ domainrepo.JobFamilyRepository = (*JobFamilyRepository)(nil)

// Method bernama objek memperjelas operasi rumpun jabatan.
func (r *JobFamilyRepository) InsertJobFamilies(ctx context.Context, v *entity.JobFamily) error {
	return r.Insert(ctx, v)
}
func (r *JobFamilyRepository) UpdateJobFamilies(ctx context.Context, v *entity.JobFamily) error {
	return r.Update(ctx, v)
}
func (r *JobFamilyRepository) DeleteJobFamiliesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *JobFamilyRepository) FindJobFamiliesByUUID(ctx context.Context, id string) (*entity.JobFamily, error) {
	return r.FindByID(ctx, id)
}
func (r *JobFamilyRepository) FindJobFamiliesByUUIDForUpdate(ctx context.Context, id string) (*entity.JobFamily, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *JobFamilyRepository) GetAllJobFamilies(ctx context.Context, limit, offset int) ([]entity.JobFamily, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobFamilyRepository) CountJobFamilies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *JobFamilyRepository) SelectJobFamiliesWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.JobFamily, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *JobFamilyRepository) CountJobFamiliesWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *JobFamilyRepository) FindJobFamiliesByFamilyCode(ctx context.Context, v string, limit, offset int) ([]entity.JobFamily, error) {
	return r.FindByFamilyCode(ctx, v, limit, offset)
}

// NewJobFamilyRepository memvalidasi koneksi dan menyiapkan repository.
func NewJobFamilyRepository(db *gorm.DB) (*JobFamilyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &JobFamilyRepository{db: db}, nil
}

func (r *JobFamilyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.JobFamily{})
}

// Insert menyimpan rumpun jabatan dengan seluruh nilai field eksplisit.
func (r *JobFamilyRepository) Insert(ctx context.Context, value *entity.JobFamily) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *JobFamilyRepository) Create(ctx context.Context, value *entity.JobFamily) error {
	return r.Insert(ctx, value)
}
func (r *JobFamilyRepository) FindByID(ctx context.Context, key string) (*entity.JobFamily, error) {
	return r.find(ctx, key, false)
}
func (r *JobFamilyRepository) SelectByID(ctx context.Context, key string) (*entity.JobFamily, error) {
	return r.FindByID(ctx, key)
}
func (r *JobFamilyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.JobFamily, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *JobFamilyRepository) find(ctx context.Context, key string, lock bool) (*entity.JobFamily, error) {
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
	value := new(entity.JobFamily)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *JobFamilyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.JobFamily, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *JobFamilyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.JobFamily, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobFamilyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.JobFamily, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination dengan urutan ID ascending.
func (r *JobFamilyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.JobFamily, error) {
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
	rows := make([]entity.JobFamily, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *JobFamilyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *JobFamilyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mempertahankan false/nol/NULL tanpa mengubah primary key atau melakukan upsert.
func (r *JobFamilyRepository) Update(ctx context.Context, value *entity.JobFamily) error {
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
func (r *JobFamilyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.JobFamily{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *JobFamilyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByFamilyCode mencari kecocokan family_code dengan paginasi dan urutan primary key.
func (r *JobFamilyRepository) FindByFamilyCode(ctx context.Context, value string, limit, offset int) ([]entity.JobFamily, error) {
	return r.SelectBy(ctx, "family_code", value, limit, offset)
}
