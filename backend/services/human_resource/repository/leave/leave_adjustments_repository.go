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

// LeaveAdjustmentRepository menyediakan kueri bertipe untuk hr_leave_adjustments.
type LeaveAdjustmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.LeaveAdjustmentRepository = (*LeaveAdjustmentRepository)(nil)

// NewLeaveAdjustmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewLeaveAdjustmentRepository(db *gorm.DB) (*LeaveAdjustmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LeaveAdjustmentRepository{db: db}, nil
}

// Operasi bernama objek tetap mengikuti validasi key dan transaksi repository lokal.
func (r *LeaveAdjustmentRepository) InsertLeaveAdjustments(ctx context.Context, value *entity.LeaveAdjustment) error {
	return r.Insert(ctx, value)
}
func (r *LeaveAdjustmentRepository) UpdateLeaveAdjustments(ctx context.Context, value *entity.LeaveAdjustment) error {
	return r.Update(ctx, value)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByUUID(ctx context.Context, uuid string) (*entity.LeaveAdjustment, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeaveAdjustment, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) DeleteLeaveAdjustmentsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeaveAdjustmentRepository) GetAllLeaveAdjustments(ctx context.Context, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeaveAdjustmentRepository) CountLeaveAdjustments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeaveAdjustmentRepository) SelectLeaveAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeaveAdjustmentRepository) CountLeaveAdjustmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByEntitlementID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByRequestedByPrincipalID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
func (r *LeaveAdjustmentRepository) FindLeaveAdjustmentsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *LeaveAdjustmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LeaveAdjustment{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *LeaveAdjustmentRepository) Insert(ctx context.Context, value *entity.LeaveAdjustment) error {
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
func (r *LeaveAdjustmentRepository) Create(ctx context.Context, value *entity.LeaveAdjustment) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *LeaveAdjustmentRepository) FindByID(ctx context.Context, key string) (*entity.LeaveAdjustment, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *LeaveAdjustmentRepository) SelectByID(ctx context.Context, key string) (*entity.LeaveAdjustment, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *LeaveAdjustmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LeaveAdjustment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *LeaveAdjustmentRepository) find(ctx context.Context, key string, lock bool) (*entity.LeaveAdjustment, error) {
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
	value := new(entity.LeaveAdjustment)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *LeaveAdjustmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *LeaveAdjustmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *LeaveAdjustmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *LeaveAdjustmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveAdjustment, error) {
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
	rows := make([]entity.LeaveAdjustment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *LeaveAdjustmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *LeaveAdjustmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LeaveAdjustmentRepository) Update(ctx context.Context, value *entity.LeaveAdjustment) error {
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
func (r *LeaveAdjustmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LeaveAdjustment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *LeaveAdjustmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByEntitlementID mencari kecocokan entitlement_id dengan paginasi dan urutan primary key.
func (r *LeaveAdjustmentRepository) FindByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectBy(ctx, "entitlement_id", value, limit, offset)
}

// FindByReasonCode mencari kecocokan reason_code dengan paginasi dan urutan primary key.
func (r *LeaveAdjustmentRepository) FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectBy(ctx, "reason_code", value, limit, offset)
}

// FindByRequestedByPrincipalID mencari kecocokan requested_by_principal_id dengan paginasi dan urutan primary key.
func (r *LeaveAdjustmentRepository) FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectBy(ctx, "requested_by_principal_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *LeaveAdjustmentRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *LeaveAdjustmentRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
