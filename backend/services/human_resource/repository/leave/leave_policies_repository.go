// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package leave

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// LeavePolicyRepository menyediakan kueri bertipe untuk hr_leave_policies.
type LeavePolicyRepository struct {
	db *gorm.DB
}

var _ domainrepo.LeavePolicyRepository = (*LeavePolicyRepository)(nil)

// NewLeavePolicyRepository memvalidasi koneksi dan menyiapkan repository.
func NewLeavePolicyRepository(db *gorm.DB) (*LeavePolicyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LeavePolicyRepository{db: db}, nil
}

func (r *LeavePolicyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LeavePolicy{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *LeavePolicyRepository) Insert(ctx context.Context, value *entity.LeavePolicy) error {
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
func (r *LeavePolicyRepository) Create(ctx context.Context, value *entity.LeavePolicy) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *LeavePolicyRepository) FindByID(ctx context.Context, key string) (*entity.LeavePolicy, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *LeavePolicyRepository) SelectByID(ctx context.Context, key string) (*entity.LeavePolicy, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *LeavePolicyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LeavePolicy, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *LeavePolicyRepository) find(ctx context.Context, key string, lock bool) (*entity.LeavePolicy, error) {
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
	value := new(entity.LeavePolicy)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *LeavePolicyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *LeavePolicyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *LeavePolicyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *LeavePolicyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeavePolicy, error) {
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
	rows := make([]entity.LeavePolicy, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *LeavePolicyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *LeavePolicyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LeavePolicyRepository) Update(ctx context.Context, value *entity.LeavePolicy) error {
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
func (r *LeavePolicyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LeavePolicy{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *LeavePolicyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPolicyCode mencari kecocokan policy_code dengan paginasi dan urutan primary key.
func (r *LeavePolicyRepository) FindByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectBy(ctx, "policy_code", value, limit, offset)
}

// FindByLeaveTypeCode mencari kecocokan leave_type_code dengan paginasi dan urutan primary key.
func (r *LeavePolicyRepository) FindByLeaveTypeCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectBy(ctx, "leave_type_code", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *LeavePolicyRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}
