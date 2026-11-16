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

// PerformanceGoalUpdateRepository menyediakan kueri bertipe untuk hr_performance_goal_updates.
type PerformanceGoalUpdateRepository struct {
	db *gorm.DB
}

var _ domainrepo.PerformanceGoalUpdateRepository = (*PerformanceGoalUpdateRepository)(nil)

// NewPerformanceGoalUpdateRepository memvalidasi koneksi dan menyiapkan repository.
func NewPerformanceGoalUpdateRepository(db *gorm.DB) (*PerformanceGoalUpdateRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PerformanceGoalUpdateRepository{db: db}, nil
}

func (r *PerformanceGoalUpdateRepository) InsertPerformanceGoalUpdates(ctx context.Context, v *entity.PerformanceGoalUpdate) error {
	return r.Insert(ctx, v)
}
func (r *PerformanceGoalUpdateRepository) UpdatePerformanceGoalUpdates(ctx context.Context, v *entity.PerformanceGoalUpdate) error {
	return r.Update(ctx, v)
}
func (r *PerformanceGoalUpdateRepository) DeletePerformanceGoalUpdatesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByUUID(ctx context.Context, id string) (*entity.PerformanceGoalUpdate, error) {
	return r.FindByID(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByUUIDForUpdate(ctx context.Context, id string) (*entity.PerformanceGoalUpdate, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PerformanceGoalUpdateRepository) GetAllPerformanceGoalUpdates(ctx context.Context, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) CountPerformanceGoalUpdates(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PerformanceGoalUpdateRepository) SelectPerformanceGoalUpdatesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) CountPerformanceGoalUpdatesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByGoalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByGoalID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByEvidenceDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByEvidenceDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByRecordedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByRecordedByPrincipalID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesBySupersedesUpdateID(ctx context.Context, v *string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindBySupersedesUpdateID(ctx, v, limit, offset)
}
func (r *PerformanceGoalUpdateRepository) FindPerformanceGoalUpdatesByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

func (r *PerformanceGoalUpdateRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PerformanceGoalUpdate{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PerformanceGoalUpdateRepository) Insert(ctx context.Context, value *entity.PerformanceGoalUpdate) error {
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
func (r *PerformanceGoalUpdateRepository) Create(ctx context.Context, value *entity.PerformanceGoalUpdate) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PerformanceGoalUpdateRepository) FindByID(ctx context.Context, key string) (*entity.PerformanceGoalUpdate, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PerformanceGoalUpdateRepository) SelectByID(ctx context.Context, key string) (*entity.PerformanceGoalUpdate, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PerformanceGoalUpdateRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PerformanceGoalUpdate, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PerformanceGoalUpdateRepository) find(ctx context.Context, key string, lock bool) (*entity.PerformanceGoalUpdate, error) {
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
	value := new(entity.PerformanceGoalUpdate)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PerformanceGoalUpdateRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PerformanceGoalUpdateRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PerformanceGoalUpdateRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PerformanceGoalUpdateRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
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
	rows := make([]entity.PerformanceGoalUpdate, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PerformanceGoalUpdateRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PerformanceGoalUpdateRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PerformanceGoalUpdateRepository) Update(ctx context.Context, value *entity.PerformanceGoalUpdate) error {
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
func (r *PerformanceGoalUpdateRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PerformanceGoalUpdate{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PerformanceGoalUpdateRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByGoalID mencari kecocokan goal_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalUpdateRepository) FindByGoalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectBy(ctx, "goal_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalUpdateRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}

// FindByRecordedByPrincipalID mencari kecocokan recorded_by_principal_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalUpdateRepository) FindByRecordedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectBy(ctx, "recorded_by_principal_id", value, limit, offset)
}

// FindBySupersedesUpdateID mencari kecocokan supersedes_update_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalUpdateRepository) FindBySupersedesUpdateID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectBy(ctx, "supersedes_update_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *PerformanceGoalUpdateRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoalUpdate, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
