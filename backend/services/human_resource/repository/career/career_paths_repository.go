// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package career

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// CareerPathRepository menyediakan kueri bertipe untuk hr_career_paths.
type CareerPathRepository struct {
	db *gorm.DB
}

var _ domainrepo.CareerPathRepository = (*CareerPathRepository)(nil)

// NewCareerPathRepository memvalidasi koneksi dan menyiapkan repository.
func NewCareerPathRepository(db *gorm.DB) (*CareerPathRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CareerPathRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak jalur karier yang eksplisit.
func (r *CareerPathRepository) InsertCareerPaths(ctx context.Context, value *entity.CareerPath) error {
	return r.Insert(ctx, value)
}
func (r *CareerPathRepository) UpdateCareerPaths(ctx context.Context, value *entity.CareerPath) error {
	return r.Update(ctx, value)
}
func (r *CareerPathRepository) DeleteCareerPathsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CareerPathRepository) FindCareerPathsByUUID(ctx context.Context, id string) (*entity.CareerPath, error) {
	return r.FindByID(ctx, id)
}
func (r *CareerPathRepository) FindCareerPathsByUUIDForUpdate(ctx context.Context, id string) (*entity.CareerPath, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CareerPathRepository) GetAllCareerPaths(ctx context.Context, limit, offset int) ([]entity.CareerPath, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CareerPathRepository) CountCareerPaths(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CareerPathRepository) SelectCareerPathsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CareerPathRepository) CountCareerPathsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CareerPathRepository) FindCareerPathsByPathCode(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.FindByPathCode(ctx, value, limit, offset)
}
func (r *CareerPathRepository) FindCareerPathsByJobFamilyID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.FindByJobFamilyID(ctx, value, limit, offset)
}
func (r *CareerPathRepository) FindCareerPathsByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}

func (r *CareerPathRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.CareerPath{})
}

// Insert menyimpan jalur karier sesuai constraint database.
func (r *CareerPathRepository) Insert(ctx context.Context, value *entity.CareerPath) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *CareerPathRepository) Create(ctx context.Context, value *entity.CareerPath) error {
	return r.Insert(ctx, value)
}
func (r *CareerPathRepository) FindByID(ctx context.Context, key string) (*entity.CareerPath, error) {
	return r.find(ctx, key, false)
}
func (r *CareerPathRepository) SelectByID(ctx context.Context, key string) (*entity.CareerPath, error) {
	return r.FindByID(ctx, key)
}
func (r *CareerPathRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.CareerPath, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *CareerPathRepository) find(ctx context.Context, key string, lock bool) (*entity.CareerPath, error) {
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
	var value entity.CareerPath
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *CareerPathRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *CareerPathRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.CareerPath, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CareerPathRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *CareerPathRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CareerPath, error) {
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
	rows := make([]entity.CareerPath, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *CareerPathRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *CareerPathRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *CareerPathRepository) Update(ctx context.Context, value *entity.CareerPath) error {
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
func (r *CareerPathRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.CareerPath{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *CareerPathRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPathCode mencari kecocokan path_code dengan paginasi dan urutan primary key.
func (r *CareerPathRepository) FindByPathCode(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectBy(ctx, "path_code", value, limit, offset)
}

// FindByJobFamilyID mencari kecocokan job_family_id dengan paginasi dan urutan primary key.
func (r *CareerPathRepository) FindByJobFamilyID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectBy(ctx, "job_family_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *CareerPathRepository) FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.CareerPath, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
