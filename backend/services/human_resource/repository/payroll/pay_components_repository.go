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

// PayComponentRepository menyediakan kueri bertipe untuk hr_pay_components.
type PayComponentRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayComponentRepository = (*PayComponentRepository)(nil)

// NewPayComponentRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayComponentRepository(db *gorm.DB) (*PayComponentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayComponentRepository{db: db}, nil
}

func (r *PayComponentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayComponent{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayComponentRepository) Insert(ctx context.Context, value *entity.PayComponent) error {
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
func (r *PayComponentRepository) Create(ctx context.Context, value *entity.PayComponent) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayComponentRepository) FindByID(ctx context.Context, key string) (*entity.PayComponent, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayComponentRepository) SelectByID(ctx context.Context, key string) (*entity.PayComponent, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayComponentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayComponent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayComponentRepository) find(ctx context.Context, key string, lock bool) (*entity.PayComponent, error) {
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
	value := new(entity.PayComponent)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayComponentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayComponentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayComponent, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayComponentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayComponentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayComponent, error) {
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
	rows := make([]entity.PayComponent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayComponentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayComponentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayComponentRepository) Update(ctx context.Context, value *entity.PayComponent) error {
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
func (r *PayComponentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayComponent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayComponentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByComponentCode mencari kecocokan component_code dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByComponentCode(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "component_code", value, limit, offset)
}

// FindByComponentType mencari kecocokan component_type dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByComponentType(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "component_type", value, limit, offset)
}

// FindByTaxTreatmentCode mencari kecocokan tax_treatment_code dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByTaxTreatmentCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "tax_treatment_code", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByRuleSetID mencari kecocokan rule_set_id dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByRuleSetID(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "rule_set_id", value, limit, offset)
}

// FindByFormulaReferenceID mencari kecocokan formula_reference_id dengan paginasi dan urutan primary key.
func (r *PayComponentRepository) FindByFormulaReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayComponent, error) {
	return r.SelectBy(ctx, "formula_reference_id", value, limit, offset)
}
