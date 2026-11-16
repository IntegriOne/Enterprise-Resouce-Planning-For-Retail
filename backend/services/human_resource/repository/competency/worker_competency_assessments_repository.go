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

// WorkerCompetencyAssessmentRepository menyediakan kueri bertipe untuk hr_worker_competency_assessments.
type WorkerCompetencyAssessmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerCompetencyAssessmentRepository = (*WorkerCompetencyAssessmentRepository)(nil)

// NewWorkerCompetencyAssessmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerCompetencyAssessmentRepository(db *gorm.DB) (*WorkerCompetencyAssessmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerCompetencyAssessmentRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak asesmen kompetensi yang eksplisit.
func (r *WorkerCompetencyAssessmentRepository) InsertWorkerCompetencyAssessments(ctx context.Context, value *entity.WorkerCompetencyAssessment) error {
	return r.Insert(ctx, value)
}
func (r *WorkerCompetencyAssessmentRepository) UpdateWorkerCompetencyAssessments(ctx context.Context, value *entity.WorkerCompetencyAssessment) error {
	return r.Update(ctx, value)
}
func (r *WorkerCompetencyAssessmentRepository) DeleteWorkerCompetencyAssessmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByUUID(ctx context.Context, id string) (*entity.WorkerCompetencyAssessment, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkerCompetencyAssessment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerCompetencyAssessmentRepository) GetAllWorkerCompetencyAssessments(ctx context.Context, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) CountWorkerCompetencyAssessments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerCompetencyAssessmentRepository) SelectWorkerCompetencyAssessmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) CountWorkerCompetencyAssessmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.FindByCompetencyID(ctx, value, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.FindByAssessedByPrincipalID(ctx, value, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) FindWorkerCompetencyAssessmentsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *WorkerCompetencyAssessmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerCompetencyAssessment{})
}

// Insert menyimpan asesmen kompetensi sesuai constraint database.
func (r *WorkerCompetencyAssessmentRepository) Insert(ctx context.Context, value *entity.WorkerCompetencyAssessment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerCompetencyAssessmentRepository) Create(ctx context.Context, value *entity.WorkerCompetencyAssessment) error {
	return r.Insert(ctx, value)
}
func (r *WorkerCompetencyAssessmentRepository) FindByID(ctx context.Context, key string) (*entity.WorkerCompetencyAssessment, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerCompetencyAssessmentRepository) SelectByID(ctx context.Context, key string) (*entity.WorkerCompetencyAssessment, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerCompetencyAssessmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkerCompetencyAssessment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerCompetencyAssessmentRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkerCompetencyAssessment, error) {
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
	var value entity.WorkerCompetencyAssessment
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *WorkerCompetencyAssessmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerCompetencyAssessmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *WorkerCompetencyAssessmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
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
	rows := make([]entity.WorkerCompetencyAssessment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerCompetencyAssessmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerCompetencyAssessmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *WorkerCompetencyAssessmentRepository) Update(ctx context.Context, value *entity.WorkerCompetencyAssessment) error {
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
func (r *WorkerCompetencyAssessmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerCompetencyAssessment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerCompetencyAssessmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerCompetencyAssessmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByCompetencyID mencari kecocokan competency_id dengan paginasi dan urutan primary key.
func (r *WorkerCompetencyAssessmentRepository) FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectBy(ctx, "competency_id", value, limit, offset)
}

// FindByAssessedByPrincipalID mencari kecocokan assessed_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerCompetencyAssessmentRepository) FindByAssessedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectBy(ctx, "assessed_by_principal_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerCompetencyAssessmentRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *WorkerCompetencyAssessmentRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.WorkerCompetencyAssessment, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
