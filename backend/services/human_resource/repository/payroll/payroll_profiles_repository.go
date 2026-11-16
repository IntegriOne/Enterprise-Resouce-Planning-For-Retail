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

// PayrollProfileRepository menyediakan kueri bertipe untuk hr_payroll_profiles.
type PayrollProfileRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollProfileRepository = (*PayrollProfileRepository)(nil)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollProfileRepository) InsertPayrollProfiles(ctx context.Context, v *entity.PayrollProfile) error {
	return r.Insert(ctx, v)
}
func (r *PayrollProfileRepository) UpdatePayrollProfiles(ctx context.Context, v *entity.PayrollProfile) error {
	return r.Update(ctx, v)
}
func (r *PayrollProfileRepository) DeletePayrollProfilesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByUUID(ctx context.Context, id string) (*entity.PayrollProfile, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollProfile, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollProfileRepository) GetAllPayrollProfiles(ctx context.Context, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollProfileRepository) CountPayrollProfiles(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollProfileRepository) SelectPayrollProfilesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollProfileRepository) CountPayrollProfilesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollCountryCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollCountryCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByTaxProfileReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByTaxProfileReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByBankAccountTokenReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByBankAccountTokenReferenceID(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByTaxStatusCode(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByTaxStatusCode(ctx, v, limit, offset)
}
func (r *PayrollProfileRepository) FindPayrollProfilesByWithholdingElectionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.FindByWithholdingElectionReferenceID(ctx, v, limit, offset)
}

// NewPayrollProfileRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollProfileRepository(db *gorm.DB) (*PayrollProfileRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollProfileRepository{db: db}, nil
}

func (r *PayrollProfileRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollProfile{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollProfileRepository) Insert(ctx context.Context, value *entity.PayrollProfile) error {
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
func (r *PayrollProfileRepository) Create(ctx context.Context, value *entity.PayrollProfile) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollProfileRepository) FindByID(ctx context.Context, key string) (*entity.PayrollProfile, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollProfileRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollProfile, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollProfileRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollProfile, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollProfileRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollProfile, error) {
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
	value := new(entity.PayrollProfile)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollProfileRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollProfileRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollProfileRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollProfileRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollProfile, error) {
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
	rows := make([]entity.PayrollProfile, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollProfileRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollProfileRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollProfileRepository) Update(ctx context.Context, value *entity.PayrollProfile) error {
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
func (r *PayrollProfileRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollProfile{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollProfileRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPayrollCountryCode mencari kecocokan payroll_country_code dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByPayrollCountryCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "payroll_country_code", value, limit, offset)
}

// FindByPayrollJurisdictionCode mencari kecocokan payroll_jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByPayrollJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "payroll_jurisdiction_code", value, limit, offset)
}

// FindByPayrollGroupID mencari kecocokan payroll_group_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "payroll_group_id", value, limit, offset)
}

// FindByEmployingLegalEntityReferenceID mencari kecocokan employing_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "employing_legal_entity_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByTaxProfileReferenceID mencari kecocokan tax_profile_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByTaxProfileReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "tax_profile_reference_id", value, limit, offset)
}

// FindByBankAccountTokenReferenceID mencari kecocokan bank_account_token_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByBankAccountTokenReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "bank_account_token_reference_id", value, limit, offset)
}

// FindByTaxStatusCode mencari kecocokan tax_status_code dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByTaxStatusCode(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "tax_status_code", value, limit, offset)
}

// FindByWithholdingElectionReferenceID mencari kecocokan withholding_election_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollProfileRepository) FindByWithholdingElectionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollProfile, error) {
	return r.SelectBy(ctx, "withholding_election_reference_id", value, limit, offset)
}
