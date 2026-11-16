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

// TalentCalibrationSessionRepository menyediakan kueri bertipe untuk hr_talent_calibration_sessions.
type TalentCalibrationSessionRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentCalibrationSessionRepository = (*TalentCalibrationSessionRepository)(nil)

// NewTalentCalibrationSessionRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentCalibrationSessionRepository(db *gorm.DB) (*TalentCalibrationSessionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentCalibrationSessionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak sesi kalibrasi yang eksplisit.
func (r *TalentCalibrationSessionRepository) InsertTalentCalibrationSessions(ctx context.Context, value *entity.TalentCalibrationSession) error {
	return r.Insert(ctx, value)
}
func (r *TalentCalibrationSessionRepository) UpdateTalentCalibrationSessions(ctx context.Context, value *entity.TalentCalibrationSession) error {
	return r.Update(ctx, value)
}
func (r *TalentCalibrationSessionRepository) DeleteTalentCalibrationSessionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByUUID(ctx context.Context, id string) (*entity.TalentCalibrationSession, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentCalibrationSession, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentCalibrationSessionRepository) GetAllTalentCalibrationSessions(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentCalibrationSessionRepository) CountTalentCalibrationSessions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentCalibrationSessionRepository) SelectTalentCalibrationSessionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentCalibrationSessionRepository) CountTalentCalibrationSessionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.FindByTalentReviewCycleID(ctx, value, limit, offset)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsBySessionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.FindBySessionCode(ctx, value, limit, offset)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByChairPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.FindByChairPrincipalID(ctx, value, limit, offset)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *TalentCalibrationSessionRepository) FindTalentCalibrationSessionsByMinutesDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.FindByMinutesDocumentVersionReferenceID(ctx, value, limit, offset)
}

func (r *TalentCalibrationSessionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentCalibrationSession{})
}

// Insert menyimpan sesi sesuai constraint database.
func (r *TalentCalibrationSessionRepository) Insert(ctx context.Context, value *entity.TalentCalibrationSession) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentCalibrationSessionRepository) Create(ctx context.Context, value *entity.TalentCalibrationSession) error {
	return r.Insert(ctx, value)
}
func (r *TalentCalibrationSessionRepository) FindByID(ctx context.Context, key string) (*entity.TalentCalibrationSession, error) {
	return r.find(ctx, key, false)
}
func (r *TalentCalibrationSessionRepository) SelectByID(ctx context.Context, key string) (*entity.TalentCalibrationSession, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentCalibrationSessionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentCalibrationSession, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentCalibrationSessionRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentCalibrationSession, error) {
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
	var value entity.TalentCalibrationSession
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentCalibrationSessionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentCalibrationSessionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentCalibrationSessionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *TalentCalibrationSessionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentCalibrationSession, error) {
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
	rows := make([]entity.TalentCalibrationSession, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentCalibrationSessionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentCalibrationSessionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *TalentCalibrationSessionRepository) Update(ctx context.Context, value *entity.TalentCalibrationSession) error {
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
func (r *TalentCalibrationSessionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentCalibrationSession{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentCalibrationSessionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByTalentReviewCycleID mencari kecocokan talent_review_cycle_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationSessionRepository) FindByTalentReviewCycleID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectBy(ctx, "talent_review_cycle_id", value, limit, offset)
}

// FindBySessionCode mencari kecocokan session_code dengan paginasi dan urutan primary key.
func (r *TalentCalibrationSessionRepository) FindBySessionCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectBy(ctx, "session_code", value, limit, offset)
}

// FindByChairPrincipalID mencari kecocokan chair_principal_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationSessionRepository) FindByChairPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectBy(ctx, "chair_principal_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *TalentCalibrationSessionRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByMinutesDocumentVersionReferenceID mencari kecocokan minutes_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentCalibrationSessionRepository) FindByMinutesDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentCalibrationSession, error) {
	return r.SelectBy(ctx, "minutes_document_version_reference_id", value, limit, offset)
}
