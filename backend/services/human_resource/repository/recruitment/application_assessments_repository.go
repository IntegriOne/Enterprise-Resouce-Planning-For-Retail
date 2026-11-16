// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package recruitment

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// ApplicationAssessmentRepository menyediakan kueri bertipe untuk hr_application_assessments.
type ApplicationAssessmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.ApplicationAssessmentRepository = (*ApplicationAssessmentRepository)(nil)

// Method bernama objek memperjelas operasi asesmen lamaran dan filter kolom.
func (r *ApplicationAssessmentRepository) InsertApplicationAssessments(ctx context.Context, v *entity.ApplicationAssessment) error {
	return r.Insert(ctx, v)
}
func (r *ApplicationAssessmentRepository) UpdateApplicationAssessments(ctx context.Context, v *entity.ApplicationAssessment) error {
	return r.Update(ctx, v)
}
func (r *ApplicationAssessmentRepository) DeleteApplicationAssessmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByUUID(ctx context.Context, id string) (*entity.ApplicationAssessment, error) {
	return r.FindByID(ctx, id)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.ApplicationAssessment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *ApplicationAssessmentRepository) GetAllApplicationAssessments(ctx context.Context, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationAssessmentRepository) CountApplicationAssessments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *ApplicationAssessmentRepository) SelectApplicationAssessmentsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *ApplicationAssessmentRepository) CountApplicationAssessmentsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByApplicationID(ctx context.Context, v string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByApplicationID(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByAssessmentCode(ctx context.Context, v string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByAssessmentCode(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByAttemptNumber(ctx context.Context, v int32, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByAttemptNumber(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByRubricDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByRubricDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByStatus(ctx context.Context, v string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByAssessorPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByAssessorPrincipalID(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByRecommendationCode(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByRecommendationCode(ctx, v, limit, offset)
}
func (r *ApplicationAssessmentRepository) FindApplicationAssessmentsByEvidenceDocumentVersionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, v, limit, offset)
}

// NewApplicationAssessmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewApplicationAssessmentRepository(db *gorm.DB) (*ApplicationAssessmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &ApplicationAssessmentRepository{db: db}, nil
}

func (r *ApplicationAssessmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.ApplicationAssessment{})
}

// Insert menyimpan asesmen lamaran tanpa menentukan keputusan penerimaan.
func (r *ApplicationAssessmentRepository) Insert(ctx context.Context, value *entity.ApplicationAssessment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *ApplicationAssessmentRepository) Create(ctx context.Context, value *entity.ApplicationAssessment) error {
	return r.Insert(ctx, value)
}
func (r *ApplicationAssessmentRepository) FindByID(ctx context.Context, key string) (*entity.ApplicationAssessment, error) {
	return r.find(ctx, key, false)
}
func (r *ApplicationAssessmentRepository) SelectByID(ctx context.Context, key string) (*entity.ApplicationAssessment, error) {
	return r.FindByID(ctx, key)
}
func (r *ApplicationAssessmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.ApplicationAssessment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *ApplicationAssessmentRepository) find(ctx context.Context, key string, lock bool) (*entity.ApplicationAssessment, error) {
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
	value := new(entity.ApplicationAssessment)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *ApplicationAssessmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *ApplicationAssessmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationAssessmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter asesmen sebelum limit/offset.
func (r *ApplicationAssessmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ApplicationAssessment, error) {
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
	rows := make([]entity.ApplicationAssessment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *ApplicationAssessmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *ApplicationAssessmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot asesmen non-key tanpa upsert dan tanpa mengubah ID.
func (r *ApplicationAssessmentRepository) Update(ctx context.Context, value *entity.ApplicationAssessment) error {
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
func (r *ApplicationAssessmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.ApplicationAssessment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *ApplicationAssessmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByApplicationID mencari kecocokan application_id dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "application_id", value, limit, offset)
}

// FindByAssessmentCode mencari kecocokan assessment_code dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByAssessmentCode(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "assessment_code", value, limit, offset)
}

// FindByAttemptNumber mencari kecocokan attempt_number dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByAttemptNumber(ctx context.Context, value int32, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "attempt_number", value, limit, offset)
}

// FindByRubricDocumentVersionReferenceID mencari kecocokan rubric_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "rubric_document_version_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByAssessorPrincipalID mencari kecocokan assessor_principal_id dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByAssessorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "assessor_principal_id", value, limit, offset)
}

// FindByRecommendationCode mencari kecocokan recommendation_code dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByRecommendationCode(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "recommendation_code", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *ApplicationAssessmentRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ApplicationAssessment, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}
