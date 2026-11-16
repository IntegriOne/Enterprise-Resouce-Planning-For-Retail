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

// TalentCalibrationDecisionRepository menyediakan kueri bertipe untuk hr_talent_calibration_decisions.
type TalentCalibrationDecisionRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentCalibrationDecisionRepository = (*TalentCalibrationDecisionRepository)(nil)

// NewTalentCalibrationDecisionRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentCalibrationDecisionRepository(db *gorm.DB) (*TalentCalibrationDecisionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentCalibrationDecisionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak keputusan kalibrasi yang eksplisit.
func (r *TalentCalibrationDecisionRepository) InsertTalentCalibrationDecisions(ctx context.Context, value *entity.TalentCalibrationDecision) error {
	return r.Insert(ctx, value)
}
func (r *TalentCalibrationDecisionRepository) UpdateTalentCalibrationDecisions(ctx context.Context, value *entity.TalentCalibrationDecision) error {
	return r.Update(ctx, value)
}
func (r *TalentCalibrationDecisionRepository) DeleteTalentCalibrationDecisionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByUUID(ctx context.Context, id string) (*entity.TalentCalibrationDecision, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentCalibrationDecision, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentCalibrationDecisionRepository) GetAllTalentCalibrationDecisions(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) CountTalentCalibrationDecisions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentCalibrationDecisionRepository) SelectTalentCalibrationDecisionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) CountTalentCalibrationDecisionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsBySessionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindBySessionID(ctx, value, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindByTalentReviewCycleID(ctx, value, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindByAssessmentID(ctx, value, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindByDimensionID(ctx, value, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindByRationaleDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) FindTalentCalibrationDecisionsByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.FindByDecidedByPrincipalID(ctx, value, limit, offset)
}

func (r *TalentCalibrationDecisionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentCalibrationDecision{})
}

// Insert menyimpan keputusan sesuai constraint database.
func (r *TalentCalibrationDecisionRepository) Insert(ctx context.Context, value *entity.TalentCalibrationDecision) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentCalibrationDecisionRepository) Create(ctx context.Context, value *entity.TalentCalibrationDecision) error {
	return r.Insert(ctx, value)
}
func (r *TalentCalibrationDecisionRepository) FindByID(ctx context.Context, key string) (*entity.TalentCalibrationDecision, error) {
	return r.find(ctx, key, false)
}
func (r *TalentCalibrationDecisionRepository) SelectByID(ctx context.Context, key string) (*entity.TalentCalibrationDecision, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentCalibrationDecisionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentCalibrationDecision, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentCalibrationDecisionRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentCalibrationDecision, error) {
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
	var value entity.TalentCalibrationDecision
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentCalibrationDecisionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentCalibrationDecisionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentCalibrationDecisionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
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
	rows := make([]entity.TalentCalibrationDecision, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentCalibrationDecisionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentCalibrationDecisionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *TalentCalibrationDecisionRepository) Update(ctx context.Context, value *entity.TalentCalibrationDecision) error {
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
func (r *TalentCalibrationDecisionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentCalibrationDecision{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentCalibrationDecisionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindBySessionID mencari kecocokan session_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindBySessionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "session_id", value, limit, offset)
}

// FindByTalentReviewCycleID mencari kecocokan talent_review_cycle_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "talent_review_cycle_id", value, limit, offset)
}

// FindByAssessmentID mencari kecocokan assessment_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindByAssessmentID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "assessment_id", value, limit, offset)
}

// FindByDimensionID mencari kecocokan dimension_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindByDimensionID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "dimension_id", value, limit, offset)
}

// FindByRationaleDocumentVersionReferenceID mencari kecocokan rationale_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindByRationaleDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "rationale_document_version_reference_id", value, limit, offset)
}

// FindByDecidedByPrincipalID mencari kecocokan decided_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationDecisionRepository) FindByDecidedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationDecision, error) {
	return r.SelectBy(ctx, "decided_by_principal_id", value, limit, offset)
}
