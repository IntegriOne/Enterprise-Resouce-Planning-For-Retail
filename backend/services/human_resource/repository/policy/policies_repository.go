package policy

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// PolicyRepository menyediakan kueri bertipe untuk hr_policies.
type PolicyRepository struct {
	db *gorm.DB
}

var _ domainrepo.PolicyRepository = (*PolicyRepository)(nil)

// NewPolicyRepository memvalidasi koneksi dan menyiapkan repository.
func NewPolicyRepository(db *gorm.DB) (*PolicyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PolicyRepository{db: db}, nil
}

// Operasi bernama objek tersedia bersama alias lama untuk kompatibilitas service.
func (r *PolicyRepository) InsertPolicies(ctx context.Context, value *entity.Policy) error {
	return r.Insert(ctx, value)
}
func (r *PolicyRepository) UpdatePolicies(ctx context.Context, value *entity.Policy) error {
	return r.Update(ctx, value)
}
func (r *PolicyRepository) FindPoliciesByUUID(ctx context.Context, uuid string) (*entity.Policy, error) {
	return r.FindByID(ctx, uuid)
}
func (r *PolicyRepository) FindPoliciesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.Policy, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *PolicyRepository) DeletePoliciesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *PolicyRepository) GetAllPolicies(ctx context.Context, limit, offset int) ([]entity.Policy, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PolicyRepository) CountPolicies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PolicyRepository) SelectPoliciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Policy, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PolicyRepository) CountPoliciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PolicyRepository) FindPoliciesByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByPolicyCode(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByCategoryCode(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *PolicyRepository) FindPoliciesByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}

func (r *PolicyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Policy{})
}

// Insert menyimpan seluruh field kebijakan tanpa menerapkan default bisnis.
func (r *PolicyRepository) Insert(ctx context.Context, value *entity.Policy) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias Insert untuk kontrak pemanggil yang sudah tersedia.
func (r *PolicyRepository) Create(ctx context.Context, value *entity.Policy) error {
	return r.Insert(ctx, value)
}

// FindByID membaca kebijakan berdasarkan primary key lengkap.
func (r *PolicyRepository) FindByID(ctx context.Context, key string) (*entity.Policy, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PolicyRepository) SelectByID(ctx context.Context, key string) (*entity.Policy, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci kebijakan dalam transaksi aktif.
func (r *PolicyRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Policy, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PolicyRepository) find(ctx context.Context, key string, lock bool) (*entity.Policy, error) {
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
	value := &entity.Policy{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman dengan urutan ID yang stabil.
func (r *PolicyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Policy, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias daftar berhalaman.
func (r *PolicyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Policy, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy membaca halaman berdasarkan satu kolom yang divalidasi.
func (r *PolicyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Policy, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination dan mengembalikan slice kosong yang eksplisit.
func (r *PolicyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Policy, error) {
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
	rows := make([]entity.Policy, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh kebijakan tanpa pagination.
func (r *PolicyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung baris yang memenuhi filter scope yang sama dengan daftar.
func (r *PolicyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh field non-key, termasuk false/nol/NULL, tanpa upsert.
func (r *PolicyRepository) Update(ctx context.Context, value *entity.Policy) error {
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

// DeleteByID menghapus satu kebijakan dengan tetap mengikuti FK dan trigger database.
func (r *PolicyRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Policy{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan berdasarkan ID.
func (r *PolicyRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPolicyCode mencari kecocokan policy_code dengan paginasi dan urutan primary key.
func (r *PolicyRepository) FindByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.SelectBy(ctx, "policy_code", value, limit, offset)
}

// FindByCategoryCode mencari kecocokan category_code dengan paginasi dan urutan primary key.
func (r *PolicyRepository) FindByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.SelectBy(ctx, "category_code", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PolicyRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByDocumentVersionReferenceID mencari kecocokan document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PolicyRepository) FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.SelectBy(ctx, "document_version_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *PolicyRepository) FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
