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

// WorkerCareerAspirationRepository menyediakan kueri bertipe untuk hr_worker_career_aspirations.
type WorkerCareerAspirationRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerCareerAspirationRepository = (*WorkerCareerAspirationRepository)(nil)

// NewWorkerCareerAspirationRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerCareerAspirationRepository(db *gorm.DB) (*WorkerCareerAspirationRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerCareerAspirationRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak aspirasi karier yang eksplisit.
func (r *WorkerCareerAspirationRepository) InsertWorkerCareerAspirations(ctx context.Context, value *entity.WorkerCareerAspiration) error {
	return r.Insert(ctx, value)
}
func (r *WorkerCareerAspirationRepository) UpdateWorkerCareerAspirations(ctx context.Context, value *entity.WorkerCareerAspiration) error {
	return r.Update(ctx, value)
}
func (r *WorkerCareerAspirationRepository) DeleteWorkerCareerAspirationsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByUUID(ctx context.Context, id string) (*entity.WorkerCareerAspiration, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkerCareerAspiration, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerCareerAspirationRepository) GetAllWorkerCareerAspirations(ctx context.Context, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerCareerAspirationRepository) CountWorkerCareerAspirations(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerCareerAspirationRepository) SelectWorkerCareerAspirationsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *WorkerCareerAspirationRepository) CountWorkerCareerAspirationsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByTargetJobID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.FindByTargetJobID(ctx, value, limit, offset)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByCareerPathTransitionID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.FindByCareerPathTransitionID(ctx, value, limit, offset)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.FindByRecordedByPrincipalID(ctx, value, limit, offset)
}
func (r *WorkerCareerAspirationRepository) FindWorkerCareerAspirationsByPreferredLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.FindByPreferredLocationReferenceID(ctx, value, limit, offset)
}

func (r *WorkerCareerAspirationRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerCareerAspiration{})
}

// Insert menyimpan aspirasi sesuai constraint database.
func (r *WorkerCareerAspirationRepository) Insert(ctx context.Context, value *entity.WorkerCareerAspiration) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerCareerAspirationRepository) Create(ctx context.Context, value *entity.WorkerCareerAspiration) error {
	return r.Insert(ctx, value)
}
func (r *WorkerCareerAspirationRepository) FindByID(ctx context.Context, key string) (*entity.WorkerCareerAspiration, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerCareerAspirationRepository) SelectByID(ctx context.Context, key string) (*entity.WorkerCareerAspiration, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerCareerAspirationRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkerCareerAspiration, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerCareerAspirationRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkerCareerAspiration, error) {
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
	var value entity.WorkerCareerAspiration
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *WorkerCareerAspirationRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerCareerAspirationRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerCareerAspirationRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *WorkerCareerAspirationRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
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
	rows := make([]entity.WorkerCareerAspiration, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerCareerAspirationRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerCareerAspirationRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *WorkerCareerAspirationRepository) Update(ctx context.Context, value *entity.WorkerCareerAspiration) error {
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
func (r *WorkerCareerAspirationRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerCareerAspiration{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerCareerAspirationRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerCareerAspirationRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByTargetJobID mencari kecocokan target_job_id dengan paginasi dan urutan primary key.
func (r *WorkerCareerAspirationRepository) FindByTargetJobID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectBy(ctx, "target_job_id", value, limit, offset)
}

// FindByCareerPathTransitionID mencari kecocokan career_path_transition_id dengan paginasi dan urutan primary key.
func (r *WorkerCareerAspirationRepository) FindByCareerPathTransitionID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectBy(ctx, "career_path_transition_id", value, limit, offset)
}

// FindByRecordedByPrincipalID mencari kecocokan recorded_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerCareerAspirationRepository) FindByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectBy(ctx, "recorded_by_principal_id", value, limit, offset)
}

// FindByPreferredLocationReferenceID mencari kecocokan preferred_location_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerCareerAspirationRepository) FindByPreferredLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCareerAspiration, error) {
	return r.SelectBy(ctx, "preferred_location_reference_id", value, limit, offset)
}
