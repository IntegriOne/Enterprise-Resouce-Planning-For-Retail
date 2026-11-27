// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package engagement

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// EngagementSurveyResultRepository menyediakan kueri bertipe untuk hr_engagement_survey_results.
type EngagementSurveyResultRepository struct {
	db *gorm.DB
}

var _ domainrepo.EngagementSurveyResultRepository = (*EngagementSurveyResultRepository)(nil)

// NewEngagementSurveyResultRepository memvalidasi koneksi dan menyiapkan repository.
func NewEngagementSurveyResultRepository(db *gorm.DB) (*EngagementSurveyResultRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EngagementSurveyResultRepository{db: db}, nil
}

func (r *EngagementSurveyResultRepository) InsertEngagementSurveyResults(ctx context.Context, value *entity.EngagementSurveyResult) error {
	return r.Insert(ctx, value)
}
func (r *EngagementSurveyResultRepository) UpdateEngagementSurveyResults(ctx context.Context, value *entity.EngagementSurveyResult) error {
	return r.Update(ctx, value)
}
func (r *EngagementSurveyResultRepository) DeleteEngagementSurveyResultsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByUUID(ctx context.Context, id string) (*entity.EngagementSurveyResult, error) {
	return r.FindByID(ctx, id)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByUUIDForUpdate(ctx context.Context, id string) (*entity.EngagementSurveyResult, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EngagementSurveyResultRepository) GetAllEngagementSurveyResults(ctx context.Context, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EngagementSurveyResultRepository) CountEngagementSurveyResults(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EngagementSurveyResultRepository) SelectEngagementSurveyResultsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EngagementSurveyResultRepository) CountEngagementSurveyResultsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsBySurveyID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindBySurveyID(ctx, value, limit, offset)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, value, limit, offset)
}
func (r *EngagementSurveyResultRepository) FindEngagementSurveyResultsByQuestionCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.FindByQuestionCode(ctx, value, limit, offset)
}

func (r *EngagementSurveyResultRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EngagementSurveyResult{})
}

// reportableQuery menyaring hasil sebelum pembacaan, pagination, dan penghitungan.
func (r *EngagementSurveyResultRepository) reportableQuery(ctx context.Context) (*gorm.DB, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	return db.Where(`EXISTS (SELECT 1 FROM hr_engagement_surveys AS survey
		WHERE survey.id = hr_engagement_survey_results.survey_id
		AND hr_engagement_survey_results.respondent_count >= survey.minimum_reporting_population)`), nil
}

// Insert menyimpan record baru tanpa upsert.
func (r *EngagementSurveyResultRepository) Insert(ctx context.Context, value *entity.EngagementSurveyResult) error {
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
func (r *EngagementSurveyResultRepository) Create(ctx context.Context, value *entity.EngagementSurveyResult) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *EngagementSurveyResultRepository) FindByID(ctx context.Context, key string) (*entity.EngagementSurveyResult, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *EngagementSurveyResultRepository) SelectByID(ctx context.Context, key string) (*entity.EngagementSurveyResult, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *EngagementSurveyResultRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.EngagementSurveyResult, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *EngagementSurveyResultRepository) find(ctx context.Context, key string, lock bool) (*entity.EngagementSurveyResult, error) {
	db, err := r.reportableQuery(ctx)
	if err != nil {
		return nil, err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	value := new(entity.EngagementSurveyResult)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *EngagementSurveyResultRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *EngagementSurveyResultRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *EngagementSurveyResultRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *EngagementSurveyResultRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	limit, offset, err := pgstore.NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.reportableQuery(ctx)
	if err != nil {
		return nil, err
	}
	db, err = pgstore.ApplyFilters(db, filters)
	if err != nil {
		return nil, err
	}
	rows := make([]entity.EngagementSurveyResult, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *EngagementSurveyResultRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *EngagementSurveyResultRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	db, err := r.reportableQuery(ctx)
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

// CountStoredBySurveyID menghitung semua hasil tersimpan untuk menjaga survei tetap immutable.
func (r *EngagementSurveyResultRepository) CountStoredBySurveyID(ctx context.Context, surveyID string) (int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Where("survey_id = ?", surveyID).Count(&total).Error
	return total, pgstore.MapError(err)
}

// Update menyimpan snapshot lengkap non-key termasuk nilai nol dan NULL.
func (r *EngagementSurveyResultRepository) Update(ctx context.Context, value *entity.EngagementSurveyResult) error {
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
func (r *EngagementSurveyResultRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EngagementSurveyResult{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *EngagementSurveyResultRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindBySurveyID mencari kecocokan survey_id dengan paginasi dan urutan primary key.
func (r *EngagementSurveyResultRepository) FindBySurveyID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectBy(ctx, "survey_id", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *EngagementSurveyResultRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByQuestionCode mencari kecocokan question_code dengan paginasi dan urutan primary key.
func (r *EngagementSurveyResultRepository) FindByQuestionCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurveyResult, error) {
	return r.SelectBy(ctx, "question_code", value, limit, offset)
}
