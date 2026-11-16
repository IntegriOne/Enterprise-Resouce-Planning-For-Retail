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

// PayrollResultRepository menyediakan kueri bertipe untuk hr_payroll_results.
type PayrollResultRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollResultRepository = (*PayrollResultRepository)(nil)

// NewPayrollResultRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollResultRepository(db *gorm.DB) (*PayrollResultRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollResultRepository{db: db}, nil
}

func (r *PayrollResultRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollResult{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollResultRepository) Insert(ctx context.Context, value *entity.PayrollResult) error {
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
func (r *PayrollResultRepository) Create(ctx context.Context, value *entity.PayrollResult) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollResultRepository) FindByID(ctx context.Context, key string) (*entity.PayrollResult, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollResultRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollResult, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollResultRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollResult, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollResultRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollResult, error) {
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
	value := new(entity.PayrollResult)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollResultRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollResultRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollResult, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollResultRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollResultRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollResult, error) {
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
	rows := make([]entity.PayrollResult, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollResultRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollResultRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollResultRepository) Update(ctx context.Context, value *entity.PayrollResult) error {
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
func (r *PayrollResultRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollResult{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollResultRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPayrollRunID mencari kecocokan payroll_run_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByPayrollRunID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "payroll_run_id", value, limit, offset)
}

// FindByPayrollGroupID mencari kecocokan payroll_group_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "payroll_group_id", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPayrollProfileReferenceID mencari kecocokan payroll_profile_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByPayrollProfileReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "payroll_profile_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByResultStatus mencari kecocokan result_status dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByResultStatus(ctx context.Context, value string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "result_status", value, limit, offset)
}

// FindByPaymentReferenceID mencari kecocokan payment_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByPaymentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "payment_reference_id", value, limit, offset)
}

// FindByPayslipDocumentReferenceID mencari kecocokan payslip_document_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollResultRepository) FindByPayslipDocumentReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollResult, error) {
	return r.SelectBy(ctx, "payslip_document_reference_id", value, limit, offset)
}
