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

// PerformanceCycleRepository menyediakan kueri bertipe untuk hr_performance_cycles.
type PerformanceCycleRepository struct {
	db *gorm.DB
}

var _ domainrepo.PerformanceCycleRepository = (*PerformanceCycleRepository)(nil)

// NewPerformanceCycleRepository memvalidasi koneksi dan menyiapkan repository.
func NewPerformanceCycleRepository(db *gorm.DB) (*PerformanceCycleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PerformanceCycleRepository{db: db}, nil
}

func (r *PerformanceCycleRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PerformanceCycle{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PerformanceCycleRepository) Insert(ctx context.Context, value *entity.PerformanceCycle) error {
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
func (r *PerformanceCycleRepository) Create(ctx context.Context, value *entity.PerformanceCycle) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PerformanceCycleRepository) FindByID(ctx context.Context, key string) (*entity.PerformanceCycle, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PerformanceCycleRepository) SelectByID(ctx context.Context, key string) (*entity.PerformanceCycle, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PerformanceCycleRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PerformanceCycle, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PerformanceCycleRepository) find(ctx context.Context, key string, lock bool) (*entity.PerformanceCycle, error) {
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
	value := new(entity.PerformanceCycle)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PerformanceCycleRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PerformanceCycleRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PerformanceCycleRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PerformanceCycleRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PerformanceCycle, error) {
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
	rows := make([]entity.PerformanceCycle, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PerformanceCycleRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PerformanceCycleRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PerformanceCycleRepository) Update(ctx context.Context, value *entity.PerformanceCycle) error {
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
func (r *PerformanceCycleRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PerformanceCycle{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PerformanceCycleRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCycleCode mencari kecocokan cycle_code dengan paginasi dan urutan primary key.
func (r *PerformanceCycleRepository) FindByCycleCode(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectBy(ctx, "cycle_code", value, limit, offset)
}

// FindByCycleType mencari kecocokan cycle_type dengan paginasi dan urutan primary key.
func (r *PerformanceCycleRepository) FindByCycleType(ctx context.Context, value string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectBy(ctx, "cycle_type", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *PerformanceCycleRepository) FindByStatus(ctx context.Context, value entity.PerformanceCycleStatus, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByLaunchedByPrincipalID mencari kecocokan launched_by_principal_id dengan paginasi dan urutan primary key.
func (r *PerformanceCycleRepository) FindByLaunchedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PerformanceCycle, error) {
	return r.SelectBy(ctx, "launched_by_principal_id", value, limit, offset)
}
