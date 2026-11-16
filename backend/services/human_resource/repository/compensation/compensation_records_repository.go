// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package compensation

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// CompensationRecordRepository menyediakan kueri bertipe untuk hr_compensation_records.
type CompensationRecordRepository struct {
	db *gorm.DB
}

var _ domainrepo.CompensationRecordRepository = (*CompensationRecordRepository)(nil)

func (r *CompensationRecordRepository) InsertCompensationRecords(ctx context.Context, value *entity.CompensationRecord) error {
	return r.Insert(ctx, value)
}
func (r *CompensationRecordRepository) UpdateCompensationRecords(ctx context.Context, value *entity.CompensationRecord) error {
	return r.Update(ctx, value)
}
func (r *CompensationRecordRepository) DeleteCompensationRecordsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByUUID(ctx context.Context, id string) (*entity.CompensationRecord, error) {
	return r.FindByID(ctx, id)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByUUIDForUpdate(ctx context.Context, id string) (*entity.CompensationRecord, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CompensationRecordRepository) GetAllCompensationRecords(ctx context.Context, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompensationRecordRepository) CountCompensationRecords(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CompensationRecordRepository) SelectCompensationRecordsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CompensationRecordRepository) CountCompensationRecordsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByAssignmentID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByComponentID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCompensationBandID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCompensationBandID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCompensationType(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCompensationType(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByStatus(ctx context.Context, value entity.CompensationStatus, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}
func (r *CompensationRecordRepository) FindCompensationRecordsBySourceChangeReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.FindBySourceChangeReferenceID(ctx, value, limit, offset)
}

// NewCompensationRecordRepository memvalidasi koneksi dan menyiapkan repository.
func NewCompensationRecordRepository(db *gorm.DB) (*CompensationRecordRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CompensationRecordRepository{db: db}, nil
}

func (r *CompensationRecordRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.CompensationRecord{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *CompensationRecordRepository) Insert(ctx context.Context, value *entity.CompensationRecord) error {
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
func (r *CompensationRecordRepository) Create(ctx context.Context, value *entity.CompensationRecord) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *CompensationRecordRepository) FindByID(ctx context.Context, key string) (*entity.CompensationRecord, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *CompensationRecordRepository) SelectByID(ctx context.Context, key string) (*entity.CompensationRecord, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *CompensationRecordRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.CompensationRecord, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *CompensationRecordRepository) find(ctx context.Context, key string, lock bool) (*entity.CompensationRecord, error) {
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
	value := new(entity.CompensationRecord)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *CompensationRecordRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *CompensationRecordRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *CompensationRecordRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *CompensationRecordRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationRecord, error) {
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
	rows := make([]entity.CompensationRecord, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *CompensationRecordRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *CompensationRecordRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *CompensationRecordRepository) Update(ctx context.Context, value *entity.CompensationRecord) error {
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
func (r *CompensationRecordRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.CompensationRecord{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *CompensationRecordRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByAssignmentID mencari kecocokan assignment_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByAssignmentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "assignment_id", value, limit, offset)
}

// FindByComponentID mencari kecocokan component_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByComponentID(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "component_id", value, limit, offset)
}

// FindByCompensationBandID mencari kecocokan compensation_band_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByCompensationBandID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "compensation_band_id", value, limit, offset)
}

// FindByCompensationType mencari kecocokan compensation_type dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByCompensationType(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "compensation_type", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByStatus(ctx context.Context, value entity.CompensationStatus, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByReasonCode mencari kecocokan reason_code dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "reason_code", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindBySourceChangeReferenceID mencari kecocokan source_change_reference_id dengan paginasi dan urutan primary key.
func (r *CompensationRecordRepository) FindBySourceChangeReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationRecord, error) {
	return r.SelectBy(ctx, "source_change_reference_id", value, limit, offset)
}
