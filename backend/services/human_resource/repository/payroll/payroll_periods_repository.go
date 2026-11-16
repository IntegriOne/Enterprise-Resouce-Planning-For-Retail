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

// PayrollPeriodRepository menyediakan kueri bertipe untuk hr_payroll_periods.
type PayrollPeriodRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollPeriodRepository = (*PayrollPeriodRepository)(nil)

// NewPayrollPeriodRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollPeriodRepository(db *gorm.DB) (*PayrollPeriodRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollPeriodRepository{db: db}, nil
}

func (r *PayrollPeriodRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollPeriod{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollPeriodRepository) Insert(ctx context.Context, value *entity.PayrollPeriod) error {
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
func (r *PayrollPeriodRepository) Create(ctx context.Context, value *entity.PayrollPeriod) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollPeriodRepository) FindByID(ctx context.Context, key string) (*entity.PayrollPeriod, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollPeriodRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollPeriod, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollPeriodRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollPeriod, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollPeriodRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollPeriod, error) {
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
	value := new(entity.PayrollPeriod)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollPeriodRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollPeriodRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollPeriodRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollPeriodRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollPeriod, error) {
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
	rows := make([]entity.PayrollPeriod, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollPeriodRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollPeriodRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollPeriodRepository) Update(ctx context.Context, value *entity.PayrollPeriod) error {
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
func (r *PayrollPeriodRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollPeriod{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollPeriodRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPayrollGroupID mencari kecocokan payroll_group_id dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "payroll_group_id", value, limit, offset)
}

// FindByEmployingLegalEntityReferenceID mencari kecocokan employing_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "employing_legal_entity_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByPeriodCode mencari kecocokan period_code dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByPeriodCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "period_code", value, limit, offset)
}

// FindByPeriodStatus mencari kecocokan period_status dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByPeriodStatus(ctx context.Context, value string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "period_status", value, limit, offset)
}

// FindByCalendarReferenceID mencari kecocokan calendar_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollPeriodRepository) FindByCalendarReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollPeriod, error) {
	return r.SelectBy(ctx, "calendar_reference_id", value, limit, offset)
}
