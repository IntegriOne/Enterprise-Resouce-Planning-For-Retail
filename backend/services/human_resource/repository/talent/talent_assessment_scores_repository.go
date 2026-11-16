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

// TalentAssessmentScoreRepository menyediakan kueri bertipe untuk hr_talent_assessment_scores.
type TalentAssessmentScoreRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentAssessmentScoreRepository = (*TalentAssessmentScoreRepository)(nil)

// NewTalentAssessmentScoreRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentAssessmentScoreRepository(db *gorm.DB) (*TalentAssessmentScoreRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentAssessmentScoreRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak skor yang eksplisit.
func (r *TalentAssessmentScoreRepository) InsertTalentAssessmentScores(ctx context.Context, value *entity.TalentAssessmentScore) error {
	return r.Insert(ctx, value)
}
func (r *TalentAssessmentScoreRepository) UpdateTalentAssessmentScores(ctx context.Context, value *entity.TalentAssessmentScore) error {
	return r.Update(ctx, value)
}
func (r *TalentAssessmentScoreRepository) DeleteTalentAssessmentScoresByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByUUID(ctx context.Context, id string) (*entity.TalentAssessmentScore, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentAssessmentScore, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentAssessmentScoreRepository) GetAllTalentAssessmentScores(ctx context.Context, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentAssessmentScoreRepository) CountTalentAssessmentScores(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentAssessmentScoreRepository) SelectTalentAssessmentScoresWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentAssessmentScoreRepository) CountTalentAssessmentScoresWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.FindByAssessmentID(ctx, value, limit, offset)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.FindByFrameworkID(ctx, value, limit, offset)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.FindByDimensionID(ctx, value, limit, offset)
}
func (r *TalentAssessmentScoreRepository) FindTalentAssessmentScoresByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, value, limit, offset)
}

func (r *TalentAssessmentScoreRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentAssessmentScore{})
}

// Insert menyimpan skor tanpa melewati constraint database.
func (r *TalentAssessmentScoreRepository) Insert(ctx context.Context, value *entity.TalentAssessmentScore) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentAssessmentScoreRepository) Create(ctx context.Context, value *entity.TalentAssessmentScore) error {
	return r.Insert(ctx, value)
}
func (r *TalentAssessmentScoreRepository) FindByID(ctx context.Context, key string) (*entity.TalentAssessmentScore, error) {
	return r.find(ctx, key, false)
}
func (r *TalentAssessmentScoreRepository) SelectByID(ctx context.Context, key string) (*entity.TalentAssessmentScore, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentAssessmentScoreRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentAssessmentScore, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentAssessmentScoreRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentAssessmentScore, error) {
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
	var value entity.TalentAssessmentScore
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentAssessmentScoreRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentAssessmentScoreRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentAssessmentScoreRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentAssessmentScoreRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentAssessmentScore, error) {
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
	rows := make([]entity.TalentAssessmentScore, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentAssessmentScoreRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentAssessmentScoreRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *TalentAssessmentScoreRepository) Update(ctx context.Context, value *entity.TalentAssessmentScore) error {
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
func (r *TalentAssessmentScoreRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentAssessmentScore{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentAssessmentScoreRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByAssessmentID mencari kecocokan assessment_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentScoreRepository) FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectBy(ctx, "assessment_id", value, limit, offset)
}

// FindByFrameworkID mencari kecocokan framework_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentScoreRepository) FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectBy(ctx, "framework_id", value, limit, offset)
}

// FindByDimensionID mencari kecocokan dimension_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentScoreRepository) FindByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectBy(ctx, "dimension_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentAssessmentScoreRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentAssessmentScore, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}
