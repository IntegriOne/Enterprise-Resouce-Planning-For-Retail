// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package payroll

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// PayrollAdjustmentRepository menyediakan kueri bertipe untuk hr_payroll_adjustments.
type PayrollAdjustmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollAdjustmentRepository = (*PayrollAdjustmentRepository)(nil)

// NewPayrollAdjustmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollAdjustmentRepository(db *gorm.DB) (*PayrollAdjustmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollAdjustmentRepository{db: db}, nil
}

func (r *PayrollAdjustmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollAdjustment{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollAdjustmentRepository) Insert(ctx context.Context, value *entity.PayrollAdjustment) error {
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
func (r *PayrollAdjustmentRepository) Create(ctx context.Context, value *entity.PayrollAdjustment) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollAdjustmentRepository) FindByID(ctx context.Context, key string) (*entity.PayrollAdjustment, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollAdjustmentRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollAdjustment, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollAdjustmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollAdjustment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollAdjustmentRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollAdjustment, error) {
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
	value := new(entity.PayrollAdjustment)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollAdjustmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollAdjustmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollAdjustmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollAdjustmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollAdjustment, error) {
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
	rows := make([]entity.PayrollAdjustment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollAdjustmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollAdjustmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollAdjustmentRepository) Update(ctx context.Context, value *entity.PayrollAdjustment) error {
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
func (r *PayrollAdjustmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollAdjustment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollAdjustmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPayrollPeriodID mencari kecocokan payroll_period_id dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByPayrollPeriodID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "payroll_period_id", value, limit, offset)
}

// FindByComponentID mencari kecocokan component_id dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "component_id", value, limit, offset)
}

// FindByAdjustmentType mencari kecocokan adjustment_type dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByAdjustmentType(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "adjustment_type", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByRequestedByPrincipalID mencari kecocokan requested_by_principal_id dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "requested_by_principal_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *PayrollAdjustmentRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PayrollAdjustment, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
