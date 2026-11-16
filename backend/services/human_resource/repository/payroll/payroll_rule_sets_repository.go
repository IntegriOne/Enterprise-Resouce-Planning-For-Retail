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

// PayrollRuleSetRepository menyediakan kueri bertipe untuk hr_payroll_rule_sets.
type PayrollRuleSetRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollRuleSetRepository = (*PayrollRuleSetRepository)(nil)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollRuleSetRepository) InsertPayrollRuleSets(ctx context.Context, v *entity.PayrollRuleSet) error {
	return r.Insert(ctx, v)
}
func (r *PayrollRuleSetRepository) UpdatePayrollRuleSets(ctx context.Context, v *entity.PayrollRuleSet) error {
	return r.Update(ctx, v)
}
func (r *PayrollRuleSetRepository) DeletePayrollRuleSetsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByUUID(ctx context.Context, id string) (*entity.PayrollRuleSet, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollRuleSet, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollRuleSetRepository) GetAllPayrollRuleSets(ctx context.Context, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollRuleSetRepository) CountPayrollRuleSets(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollRuleSetRepository) SelectPayrollRuleSetsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollRuleSetRepository) CountPayrollRuleSetsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByRuleCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByRuleCode(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByArtifactDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByArtifactDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByLegalBasisDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByLegalBasisDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRuleSetRepository) FindPayrollRuleSetsByApprovedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}

// NewPayrollRuleSetRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollRuleSetRepository(db *gorm.DB) (*PayrollRuleSetRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollRuleSetRepository{db: db}, nil
}

func (r *PayrollRuleSetRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollRuleSet{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollRuleSetRepository) Insert(ctx context.Context, value *entity.PayrollRuleSet) error {
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
func (r *PayrollRuleSetRepository) Create(ctx context.Context, value *entity.PayrollRuleSet) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollRuleSetRepository) FindByID(ctx context.Context, key string) (*entity.PayrollRuleSet, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollRuleSetRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollRuleSet, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollRuleSetRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollRuleSet, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollRuleSetRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollRuleSet, error) {
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
	value := new(entity.PayrollRuleSet)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollRuleSetRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollRuleSetRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollRuleSetRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollRuleSetRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRuleSet, error) {
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
	rows := make([]entity.PayrollRuleSet, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollRuleSetRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollRuleSetRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollRuleSetRepository) Update(ctx context.Context, value *entity.PayrollRuleSet) error {
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
func (r *PayrollRuleSetRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollRuleSet{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollRuleSetRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRuleCode mencari kecocokan rule_code dengan paginasi dan urutan primary key.
func (r *PayrollRuleSetRepository) FindByRuleCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectBy(ctx, "rule_code", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayrollRuleSetRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByArtifactDocumentVersionReferenceID mencari kecocokan artifact_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRuleSetRepository) FindByArtifactDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectBy(ctx, "artifact_document_version_reference_id", value, limit, offset)
}

// FindByLegalBasisDocumentVersionReferenceID mencari kecocokan legal_basis_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRuleSetRepository) FindByLegalBasisDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectBy(ctx, "legal_basis_document_version_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *PayrollRuleSetRepository) FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRuleSet, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
