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

// CompensationBandRepository menyediakan kueri bertipe untuk hr_compensation_bands.
type CompensationBandRepository struct {
	db *gorm.DB
}

var _ domainrepo.CompensationBandRepository = (*CompensationBandRepository)(nil)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *CompensationBandRepository) InsertCompensationBands(ctx context.Context, value *entity.CompensationBand) error {
	return r.Insert(ctx, value)
}
func (r *CompensationBandRepository) UpdateCompensationBands(ctx context.Context, value *entity.CompensationBand) error {
	return r.Update(ctx, value)
}
func (r *CompensationBandRepository) DeleteCompensationBandsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *CompensationBandRepository) FindCompensationBandsByUUID(ctx context.Context, id string) (*entity.CompensationBand, error) {
	return r.FindByID(ctx, id)
}
func (r *CompensationBandRepository) FindCompensationBandsByUUIDForUpdate(ctx context.Context, id string) (*entity.CompensationBand, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *CompensationBandRepository) GetAllCompensationBands(ctx context.Context, limit, offset int) ([]entity.CompensationBand, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *CompensationBandRepository) CountCompensationBands(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *CompensationBandRepository) SelectCompensationBandsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *CompensationBandRepository) CountCompensationBandsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *CompensationBandRepository) FindCompensationBandsByBandCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByBandCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByJobLevelCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByJobLevelCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}
func (r *CompensationBandRepository) FindCompensationBandsByApprovalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.FindByApprovalReferenceID(ctx, value, limit, offset)
}

// NewCompensationBandRepository memvalidasi koneksi dan menyiapkan repository.
func NewCompensationBandRepository(db *gorm.DB) (*CompensationBandRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CompensationBandRepository{db: db}, nil
}

func (r *CompensationBandRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.CompensationBand{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *CompensationBandRepository) Insert(ctx context.Context, value *entity.CompensationBand) error {
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
func (r *CompensationBandRepository) Create(ctx context.Context, value *entity.CompensationBand) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *CompensationBandRepository) FindByID(ctx context.Context, key string) (*entity.CompensationBand, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *CompensationBandRepository) SelectByID(ctx context.Context, key string) (*entity.CompensationBand, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *CompensationBandRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.CompensationBand, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *CompensationBandRepository) find(ctx context.Context, key string, lock bool) (*entity.CompensationBand, error) {
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
	value := new(entity.CompensationBand)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *CompensationBandRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *CompensationBandRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.CompensationBand, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *CompensationBandRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *CompensationBandRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.CompensationBand, error) {
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
	rows := make([]entity.CompensationBand, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *CompensationBandRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *CompensationBandRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *CompensationBandRepository) Update(ctx context.Context, value *entity.CompensationBand) error {
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
func (r *CompensationBandRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.CompensationBand{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *CompensationBandRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByBandCode mencari kecocokan band_code dengan paginasi dan urutan primary key.
func (r *CompensationBandRepository) FindByBandCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectBy(ctx, "band_code", value, limit, offset)
}

// FindByJobLevelCode mencari kecocokan job_level_code dengan paginasi dan urutan primary key.
func (r *CompensationBandRepository) FindByJobLevelCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectBy(ctx, "job_level_code", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *CompensationBandRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *CompensationBandRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByApprovalReferenceID mencari kecocokan approval_reference_id dengan paginasi dan urutan primary key.
func (r *CompensationBandRepository) FindByApprovalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.CompensationBand, error) {
	return r.SelectBy(ctx, "approval_reference_id", value, limit, offset)
}
