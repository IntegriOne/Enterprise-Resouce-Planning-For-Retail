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

// PayrollGroupRepository menyediakan kueri bertipe untuk hr_payroll_groups.
type PayrollGroupRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollGroupRepository = (*PayrollGroupRepository)(nil)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollGroupRepository) InsertPayrollGroups(ctx context.Context, v *entity.PayrollGroup) error {
	return r.Insert(ctx, v)
}
func (r *PayrollGroupRepository) UpdatePayrollGroups(ctx context.Context, v *entity.PayrollGroup) error {
	return r.Update(ctx, v)
}
func (r *PayrollGroupRepository) DeletePayrollGroupsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByUUID(ctx context.Context, id string) (*entity.PayrollGroup, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollGroup, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollGroupRepository) GetAllPayrollGroups(ctx context.Context, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollGroupRepository) CountPayrollGroups(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollGroupRepository) SelectPayrollGroupsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollGroupRepository) CountPayrollGroupsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByGroupCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByGroupCode(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollGroupRepository) FindPayrollGroupsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}

// NewPayrollGroupRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollGroupRepository(db *gorm.DB) (*PayrollGroupRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollGroupRepository{db: db}, nil
}

func (r *PayrollGroupRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollGroup{})
}

// Insert menyimpan snapshot baru dan mengembalikan identity dari PostgreSQL.
func (r *PayrollGroupRepository) Insert(ctx context.Context, value *entity.PayrollGroup) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias INSERT tanpa upsert.
func (r *PayrollGroupRepository) Create(ctx context.Context, value *entity.PayrollGroup) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu primary key lengkap.
func (r *PayrollGroupRepository) FindByID(ctx context.Context, key string) (*entity.PayrollGroup, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan berdasarkan primary key.
func (r *PayrollGroupRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollGroup, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record di dalam transaksi aktif.
func (r *PayrollGroupRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollGroup, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollGroupRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollGroup, error) {
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
	value := new(entity.PayrollGroup)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key yang stabil.
func (r *PayrollGroupRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollGroupRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai pencarian, termasuk NULL.
func (r *PayrollGroupRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan seluruh filter sebelum pagination.
func (r *PayrollGroupRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollGroup, error) {
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
	rows := make([]entity.PayrollGroup, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record tanpa pagination.
func (r *PayrollGroupRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record yang memenuhi filter scope.
func (r *PayrollGroupRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh kolom non-key, termasuk nilai nol dan NULL, tanpa upsert.
func (r *PayrollGroupRepository) Update(ctx context.Context, value *entity.PayrollGroup) error {
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

// DeleteByID menghapus satu primary key dan tetap mengikuti constraint database.
func (r *PayrollGroupRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollGroup{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan berdasarkan primary key.
func (r *PayrollGroupRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByGroupCode mencari kecocokan group_code dengan paginasi dan urutan primary key.
func (r *PayrollGroupRepository) FindByGroupCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectBy(ctx, "group_code", value, limit, offset)
}

// FindByEmployingLegalEntityReferenceID mencari kecocokan employing_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollGroupRepository) FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectBy(ctx, "employing_legal_entity_reference_id", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayrollGroupRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollGroupRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollGroup, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}
