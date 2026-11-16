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

// PerformanceGoalRepository menyediakan kueri bertipe untuk hr_performance_goals.
type PerformanceGoalRepository struct {
	db *gorm.DB
}

var _ domainrepo.PerformanceGoalRepository = (*PerformanceGoalRepository)(nil)

// NewPerformanceGoalRepository memvalidasi koneksi dan menyiapkan repository.
func NewPerformanceGoalRepository(db *gorm.DB) (*PerformanceGoalRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PerformanceGoalRepository{db: db}, nil
}

func (r *PerformanceGoalRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PerformanceGoal{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PerformanceGoalRepository) Insert(ctx context.Context, value *entity.PerformanceGoal) error {
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
func (r *PerformanceGoalRepository) Create(ctx context.Context, value *entity.PerformanceGoal) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PerformanceGoalRepository) FindByID(ctx context.Context, key string) (*entity.PerformanceGoal, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PerformanceGoalRepository) SelectByID(ctx context.Context, key string) (*entity.PerformanceGoal, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PerformanceGoalRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PerformanceGoal, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PerformanceGoalRepository) find(ctx context.Context, key string, lock bool) (*entity.PerformanceGoal, error) {
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
	value := new(entity.PerformanceGoal)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PerformanceGoalRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PerformanceGoalRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PerformanceGoalRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PerformanceGoalRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceGoal, error) {
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
	rows := make([]entity.PerformanceGoal, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PerformanceGoalRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PerformanceGoalRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PerformanceGoalRepository) Update(ctx context.Context, value *entity.PerformanceGoal) error {
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
func (r *PerformanceGoalRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PerformanceGoal{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PerformanceGoalRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPerformanceCycleID mencari kecocokan performance_cycle_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByPerformanceCycleID(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "performance_cycle_id", value, limit, offset)
}

// FindByParentGoalID mencari kecocokan parent_goal_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByParentGoalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "parent_goal_id", value, limit, offset)
}

// FindByGoalCode mencari kecocokan goal_code dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByGoalCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "goal_code", value, limit, offset)
}

// FindByGoalType mencari kecocokan goal_type dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByGoalType(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "goal_type", value, limit, offset)
}

// FindByUnitCode mencari kecocokan unit_code dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByUnitCode(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "unit_code", value, limit, offset)
}

// FindByGoalStatus mencari kecocokan goal_status dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByGoalStatus(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "goal_status", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *PerformanceGoalRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceGoal, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
