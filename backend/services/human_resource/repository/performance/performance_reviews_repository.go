// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package performance

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// PerformanceReviewRepository menyediakan kueri bertipe untuk hr_performance_reviews.
type PerformanceReviewRepository struct {
	db *gorm.DB
}

var _ domainrepo.PerformanceReviewRepository = (*PerformanceReviewRepository)(nil)

// NewPerformanceReviewRepository memvalidasi koneksi dan menyiapkan repository.
func NewPerformanceReviewRepository(db *gorm.DB) (*PerformanceReviewRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PerformanceReviewRepository{db: db}, nil
}

func (r *PerformanceReviewRepository) InsertPerformanceReviews(ctx context.Context, v *entity.PerformanceReview) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceReviewRepository) UpdatePerformanceReviews(ctx context.Context, v *entity.PerformanceReview) error {
	return r.Update(ctx, v)
}
func (r *PerformanceReviewRepository) DeletePerformanceReviewsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByUUID(ctx context.Context, id string) (*entity.PerformanceReview, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceReview, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceReviewRepository) GetAllPerformanceReviews(ctx context.Context, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceReviewRepository) CountPerformanceReviews(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceReviewRepository) SelectPerformanceReviewsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceReviewRepository) CountPerformanceReviewsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByPerformanceCycleID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByPerformanceCycleID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByReviewerPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByReviewerPrincipalID(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByOverallRatingCode(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByOverallRatingCode(ctx, v, limit, offset)
}
func (r *PerformanceReviewRepository) FindPerformanceReviewsByDevelopmentPlanID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.FindByDevelopmentPlanID(ctx, v, limit, offset)
}

func (r *PerformanceReviewRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PerformanceReview{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PerformanceReviewRepository) Insert(ctx context.Context, value *entity.PerformanceReview) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias INSERT.
func (r *PerformanceReviewRepository) Create(ctx context.Context, value *entity.PerformanceReview) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PerformanceReviewRepository) FindByID(ctx context.Context, key string) (*entity.PerformanceReview, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PerformanceReviewRepository) SelectByID(ctx context.Context, key string) (*entity.PerformanceReview, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PerformanceReviewRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PerformanceReview, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PerformanceReviewRepository) find(ctx context.Context, key string, lock bool) (*entity.PerformanceReview, error) {
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
	value := new(entity.PerformanceReview)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PerformanceReviewRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PerformanceReviewRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PerformanceReviewRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PerformanceReviewRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReview, error) {
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
	rows := make([]entity.PerformanceReview, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PerformanceReviewRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PerformanceReviewRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot lengkap non-key termasuk nilai nol dan NULL.
func (r *PerformanceReviewRepository) Update(ctx context.Context, value *entity.PerformanceReview) error {
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

// DeleteByID menghapus satu record sesuai constraint database.
func (r *PerformanceReviewRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PerformanceReview{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PerformanceReviewRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPerformanceCycleID mencari kecocokan performance_cycle_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewRepository) FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectBy(ctx, "performance_cycle_id", value, limit, offset)
}

// FindByReviewerPrincipalID mencari kecocokan reviewer_principal_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewRepository) FindByReviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectBy(ctx, "reviewer_principal_id", value, limit, offset)
}

// FindByOverallRatingCode mencari kecocokan overall_rating_code dengan paginasi dan urutan primary key.
func (r *PerformanceReviewRepository) FindByOverallRatingCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectBy(ctx, "overall_rating_code", value, limit, offset)
}

// FindByDevelopmentPlanID mencari kecocokan development_plan_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewRepository) FindByDevelopmentPlanID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceReview, error) {
	return r.SelectBy(ctx, "development_plan_id", value, limit, offset)
}
