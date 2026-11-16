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

// LeaveLedgerEntryRepository menyediakan kueri bertipe untuk hr_leave_ledger_entries.
type LeaveLedgerEntryRepository struct {
	db *gorm.DB
}

var _ domainrepo.LeaveLedgerEntryRepository = (*LeaveLedgerEntryRepository)(nil)

// NewLeaveLedgerEntryRepository memvalidasi koneksi dan menyiapkan repository.
func NewLeaveLedgerEntryRepository(db *gorm.DB) (*LeaveLedgerEntryRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LeaveLedgerEntryRepository{db: db}, nil
}

// Operasi bernama objek tidak mengubah aturan append-only ledger pada database.
func (r *LeaveLedgerEntryRepository) InsertLeaveLedgerEntries(ctx context.Context, value *entity.LeaveLedgerEntry) error {
	return r.Insert(ctx, value)
}
func (r *LeaveLedgerEntryRepository) UpdateLeaveLedgerEntries(ctx context.Context, value *entity.LeaveLedgerEntry) error {
	return r.Update(ctx, value)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByUUID(ctx context.Context, uuid string) (*entity.LeaveLedgerEntry, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeaveLedgerEntry, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) DeleteLeaveLedgerEntriesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) GetAllLeaveLedgerEntries(ctx context.Context, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeaveLedgerEntryRepository) CountLeaveLedgerEntries(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeaveLedgerEntryRepository) SelectLeaveLedgerEntriesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeaveLedgerEntryRepository) CountLeaveLedgerEntriesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByEntitlementID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByEntryType(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByEntryType(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByAdjustmentID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByAdjustmentID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByReversesEntryID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByReversesEntryID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindBySourceManagement(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindBySourceEventID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByActorPrincipalID(ctx, value, limit, offset)
}

func (r *LeaveLedgerEntryRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LeaveLedgerEntry{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *LeaveLedgerEntryRepository) Insert(ctx context.Context, value *entity.LeaveLedgerEntry) error {
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
func (r *LeaveLedgerEntryRepository) Create(ctx context.Context, value *entity.LeaveLedgerEntry) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *LeaveLedgerEntryRepository) FindByID(ctx context.Context, key string) (*entity.LeaveLedgerEntry, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *LeaveLedgerEntryRepository) SelectByID(ctx context.Context, key string) (*entity.LeaveLedgerEntry, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *LeaveLedgerEntryRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LeaveLedgerEntry, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *LeaveLedgerEntryRepository) find(ctx context.Context, key string, lock bool) (*entity.LeaveLedgerEntry, error) {
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
	value := new(entity.LeaveLedgerEntry)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *LeaveLedgerEntryRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *LeaveLedgerEntryRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *LeaveLedgerEntryRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *LeaveLedgerEntryRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
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
	rows := make([]entity.LeaveLedgerEntry, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *LeaveLedgerEntryRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *LeaveLedgerEntryRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengikuti pembatasan immutable milik database.
func (r *LeaveLedgerEntryRepository) Update(ctx context.Context, value *entity.LeaveLedgerEntry) error {
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

// DeleteByID tetap mengikuti pembatasan immutable database.
func (r *LeaveLedgerEntryRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LeaveLedgerEntry{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *LeaveLedgerEntryRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByEntitlementID mencari kecocokan entitlement_id dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "entitlement_id", value, limit, offset)
}

// FindByEntryType mencari kecocokan entry_type dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByEntryType(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "entry_type", value, limit, offset)
}

// FindByAdjustmentID mencari kecocokan adjustment_id dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByAdjustmentID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "adjustment_id", value, limit, offset)
}

// FindByReversesEntryID mencari kecocokan reverses_entry_id dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByReversesEntryID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "reverses_entry_id", value, limit, offset)
}

// FindBySourceManagement mencari kecocokan source_management dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "source_management", value, limit, offset)
}

// FindBySourceEventID mencari kecocokan source_event_id dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "source_event_id", value, limit, offset)
}

// FindByReasonCode mencari kecocokan reason_code dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "reason_code", value, limit, offset)
}

// FindByActorPrincipalID mencari kecocokan actor_principal_id dengan paginasi dan urutan primary key.
func (r *LeaveLedgerEntryRepository) FindByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectBy(ctx, "actor_principal_id", value, limit, offset)
}
