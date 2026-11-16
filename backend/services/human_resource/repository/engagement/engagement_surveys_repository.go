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

// EngagementSurveyRepository menyediakan kueri bertipe untuk hr_engagement_surveys.
type EngagementSurveyRepository struct {
	db *gorm.DB
}

var _ domainrepo.EngagementSurveyRepository = (*EngagementSurveyRepository)(nil)

// NewEngagementSurveyRepository memvalidasi koneksi dan menyiapkan repository.
func NewEngagementSurveyRepository(db *gorm.DB) (*EngagementSurveyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EngagementSurveyRepository{db: db}, nil
}

func (r *EngagementSurveyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EngagementSurvey{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *EngagementSurveyRepository) Insert(ctx context.Context, value *entity.EngagementSurvey) error {
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
func (r *EngagementSurveyRepository) Create(ctx context.Context, value *entity.EngagementSurvey) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *EngagementSurveyRepository) FindByID(ctx context.Context, key string) (*entity.EngagementSurvey, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *EngagementSurveyRepository) SelectByID(ctx context.Context, key string) (*entity.EngagementSurvey, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *EngagementSurveyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.EngagementSurvey, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *EngagementSurveyRepository) find(ctx context.Context, key string, lock bool) (*entity.EngagementSurvey, error) {
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
	value := new(entity.EngagementSurvey)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *EngagementSurveyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *EngagementSurveyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *EngagementSurveyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *EngagementSurveyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EngagementSurvey, error) {
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
	rows := make([]entity.EngagementSurvey, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *EngagementSurveyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *EngagementSurveyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *EngagementSurveyRepository) Update(ctx context.Context, value *entity.EngagementSurvey) error {
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
func (r *EngagementSurveyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EngagementSurvey{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *EngagementSurveyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindBySurveyCode mencari kecocokan survey_code dengan paginasi dan urutan primary key.
func (r *EngagementSurveyRepository) FindBySurveyCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectBy(ctx, "survey_code", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *EngagementSurveyRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByTopicCode mencari kecocokan topic_code dengan paginasi dan urutan primary key.
func (r *EngagementSurveyRepository) FindByTopicCode(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectBy(ctx, "topic_code", value, limit, offset)
}

// FindByQuestionnaireDocumentVersionReferenceID mencari kecocokan questionnaire_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *EngagementSurveyRepository) FindByQuestionnaireDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EngagementSurvey, error) {
	return r.SelectBy(ctx, "questionnaire_document_version_reference_id", value, limit, offset)
}
