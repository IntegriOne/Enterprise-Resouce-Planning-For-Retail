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

// LeaveEntitlementRepository menyediakan kueri bertipe untuk hr_leave_entitlements.
type LeaveEntitlementRepository struct {
	db *gorm.DB
}

var _ domainrepo.LeaveEntitlementRepository = (*LeaveEntitlementRepository)(nil)

// NewLeaveEntitlementRepository memvalidasi koneksi dan menyiapkan repository.
func NewLeaveEntitlementRepository(db *gorm.DB) (*LeaveEntitlementRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LeaveEntitlementRepository{db: db}, nil
}

func (r *LeaveEntitlementRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LeaveEntitlement{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *LeaveEntitlementRepository) Insert(ctx context.Context, value *entity.LeaveEntitlement) error {
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
func (r *LeaveEntitlementRepository) Create(ctx context.Context, value *entity.LeaveEntitlement) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *LeaveEntitlementRepository) FindByID(ctx context.Context, key string) (*entity.LeaveEntitlement, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *LeaveEntitlementRepository) SelectByID(ctx context.Context, key string) (*entity.LeaveEntitlement, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *LeaveEntitlementRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LeaveEntitlement, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *LeaveEntitlementRepository) find(ctx context.Context, key string, lock bool) (*entity.LeaveEntitlement, error) {
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
	value := new(entity.LeaveEntitlement)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *LeaveEntitlementRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *LeaveEntitlementRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *LeaveEntitlementRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *LeaveEntitlementRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveEntitlement, error) {
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
	rows := make([]entity.LeaveEntitlement, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *LeaveEntitlementRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *LeaveEntitlementRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LeaveEntitlementRepository) Update(ctx context.Context, value *entity.LeaveEntitlement) error {
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
func (r *LeaveEntitlementRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LeaveEntitlement{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *LeaveEntitlementRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *LeaveEntitlementRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByLeavePolicyID mencari kecocokan leave_policy_id dengan paginasi dan urutan primary key.
func (r *LeaveEntitlementRepository) FindByLeavePolicyID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectBy(ctx, "leave_policy_id", value, limit, offset)
}

// FindByUOMCode mencari kecocokan uom_code dengan paginasi dan urutan primary key.
func (r *LeaveEntitlementRepository) FindByUOMCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectBy(ctx, "uom_code", value, limit, offset)
}

// FindBySourceWorkforceSnapshotReferenceID mencari kecocokan source_workforce_snapshot_reference_id dengan paginasi dan urutan primary key.
func (r *LeaveEntitlementRepository) FindBySourceWorkforceSnapshotReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveEntitlement, error) {
	return r.SelectBy(ctx, "source_workforce_snapshot_reference_id", value, limit, offset)
}
