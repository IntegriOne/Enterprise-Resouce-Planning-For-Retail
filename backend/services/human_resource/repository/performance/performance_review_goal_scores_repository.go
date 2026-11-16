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

// PerformanceReviewGoalScoreRepository menyediakan kueri bertipe untuk hr_performance_review_goal_scores.
type PerformanceReviewGoalScoreRepository struct {
	db *gorm.DB
}

var _ domainrepo.PerformanceReviewGoalScoreRepository = (*PerformanceReviewGoalScoreRepository)(nil)

// NewPerformanceReviewGoalScoreRepository memvalidasi koneksi dan menyiapkan repository.
func NewPerformanceReviewGoalScoreRepository(db *gorm.DB) (*PerformanceReviewGoalScoreRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PerformanceReviewGoalScoreRepository{db: db}, nil
}

func (r *PerformanceReviewGoalScoreRepository) InsertPerformanceReviewGoalScores(ctx context.Context, v *entity.PerformanceReviewGoalScore) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceReviewGoalScoreRepository) UpdatePerformanceReviewGoalScores(ctx context.Context, v *entity.PerformanceReviewGoalScore) error {
	return r.Update(ctx, v)
}
func (r *PerformanceReviewGoalScoreRepository) DeletePerformanceReviewGoalScoresByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByUUID(ctx context.Context, id string) (*entity.PerformanceReviewGoalScore, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceReviewGoalScore, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceReviewGoalScoreRepository) GetAllPerformanceReviewGoalScores(ctx context.Context, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) CountPerformanceReviewGoalScores(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceReviewGoalScoreRepository) SelectPerformanceReviewGoalScoresWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) CountPerformanceReviewGoalScoresWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByPerformanceReviewID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByPerformanceReviewID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByGoalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByGoalID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByPerformanceCycleID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByPerformanceCycleID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByRubricDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByRubricDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PerformanceReviewGoalScoreRepository) FindPerformanceReviewGoalScoresByEvidenceDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, v, limit, offset)
}

func (r *PerformanceReviewGoalScoreRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PerformanceReviewGoalScore{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PerformanceReviewGoalScoreRepository) Insert(ctx context.Context, value *entity.PerformanceReviewGoalScore) error {
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
func (r *PerformanceReviewGoalScoreRepository) Create(ctx context.Context, value *entity.PerformanceReviewGoalScore) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PerformanceReviewGoalScoreRepository) FindByID(ctx context.Context, key string) (*entity.PerformanceReviewGoalScore, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PerformanceReviewGoalScoreRepository) SelectByID(ctx context.Context, key string) (*entity.PerformanceReviewGoalScore, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PerformanceReviewGoalScoreRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PerformanceReviewGoalScore, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PerformanceReviewGoalScoreRepository) find(ctx context.Context, key string, lock bool) (*entity.PerformanceReviewGoalScore, error) {
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
	value := new(entity.PerformanceReviewGoalScore)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PerformanceReviewGoalScoreRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PerformanceReviewGoalScoreRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PerformanceReviewGoalScoreRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PerformanceReviewGoalScoreRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
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
	rows := make([]entity.PerformanceReviewGoalScore, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PerformanceReviewGoalScoreRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PerformanceReviewGoalScoreRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PerformanceReviewGoalScoreRepository) Update(ctx context.Context, value *entity.PerformanceReviewGoalScore) error {
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
func (r *PerformanceReviewGoalScoreRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PerformanceReviewGoalScore{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PerformanceReviewGoalScoreRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPerformanceReviewID mencari kecocokan performance_review_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByPerformanceReviewID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "performance_review_id", value, limit, offset)
}

// FindByGoalID mencari kecocokan goal_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByGoalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "goal_id", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPerformanceCycleID mencari kecocokan performance_cycle_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "performance_cycle_id", value, limit, offset)
}

// FindByRubricDocumentVersionReferenceID mencari kecocokan rubric_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "rubric_document_version_reference_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PerformanceReviewGoalScoreRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceReviewGoalScore, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}
