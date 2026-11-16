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

// TalentAssessmentRepository menyediakan kueri bertipe untuk hr_talent_assessments.
type TalentAssessmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentAssessmentRepository = (*TalentAssessmentRepository)(nil)

// NewTalentAssessmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentAssessmentRepository(db *gorm.DB) (*TalentAssessmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentAssessmentRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak asesmen yang eksplisit.
func (r *TalentAssessmentRepository) InsertTalentAssessments(ctx context.Context, value *entity.TalentAssessment) error {
	return r.Insert(ctx, value)
}
func (r *TalentAssessmentRepository) UpdateTalentAssessments(ctx context.Context, value *entity.TalentAssessment) error {
	return r.Update(ctx, value)
}
func (r *TalentAssessmentRepository) DeleteTalentAssessmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByUUID(ctx context.Context, id string) (*entity.TalentAssessment, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentAssessment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentAssessmentRepository) GetAllTalentAssessments(ctx context.Context, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentAssessmentRepository) CountTalentAssessments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentAssessmentRepository) SelectTalentAssessmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentAssessmentRepository) CountTalentAssessmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByTalentReviewCycleID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByFrameworkID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByPerformanceReviewID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByPerformanceReviewID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByAssessedByPrincipalID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsBySegmentCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindBySegmentCode(ctx, value, limit, offset)
}
func (r *TalentAssessmentRepository) FindTalentAssessmentsByFinalizedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.FindByFinalizedByPrincipalID(ctx, value, limit, offset)
}

func (r *TalentAssessmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentAssessment{})
}

// Insert menyimpan asesmen tanpa melewati constraint database.
func (r *TalentAssessmentRepository) Insert(ctx context.Context, value *entity.TalentAssessment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentAssessmentRepository) Create(ctx context.Context, value *entity.TalentAssessment) error {
	return r.Insert(ctx, value)
}
func (r *TalentAssessmentRepository) FindByID(ctx context.Context, key string) (*entity.TalentAssessment, error) {
	return r.find(ctx, key, false)
}
func (r *TalentAssessmentRepository) SelectByID(ctx context.Context, key string) (*entity.TalentAssessment, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentAssessmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentAssessment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentAssessmentRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentAssessment, error) {
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
	var value entity.TalentAssessment
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentAssessmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentAssessmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentAssessmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentAssessmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentAssessment, error) {
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
	rows := make([]entity.TalentAssessment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentAssessmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentAssessmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *TalentAssessmentRepository) Update(ctx context.Context, value *entity.TalentAssessment) error {
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
func (r *TalentAssessmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentAssessment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentAssessmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByTalentReviewCycleID mencari kecocokan talent_review_cycle_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "talent_review_cycle_id", value, limit, offset)
}

// FindByFrameworkID mencari kecocokan framework_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "framework_id", value, limit, offset)
}

// FindByPerformanceReviewID mencari kecocokan performance_review_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByPerformanceReviewID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "performance_review_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByAssessedByPrincipalID mencari kecocokan assessed_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "assessed_by_principal_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}

// FindBySegmentCode mencari kecocokan segment_code dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindBySegmentCode(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "segment_code", value, limit, offset)
}

// FindByFinalizedByPrincipalID mencari kecocokan finalized_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentRepository) FindByFinalizedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentAssessment, error) {
	return r.SelectBy(ctx, "finalized_by_principal_id", value, limit, offset)
}
