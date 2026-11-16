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

// PolicyAcknowledgementRepository menyediakan kueri bertipe untuk hr_policy_acknowledgements.
type PolicyAcknowledgementRepository struct {
	db *gorm.DB
}

var _ domainrepo.PolicyAcknowledgementRepository = (*PolicyAcknowledgementRepository)(nil)

// NewPolicyAcknowledgementRepository memvalidasi koneksi dan menyiapkan repository.
func NewPolicyAcknowledgementRepository(db *gorm.DB) (*PolicyAcknowledgementRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PolicyAcknowledgementRepository{db: db}, nil
}

func (r *PolicyAcknowledgementRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PolicyAcknowledgement{})
}

// Insert menyimpan acknowledgement dan mengembalikan hasil persistensi.
func (r *PolicyAcknowledgementRepository) Insert(ctx context.Context, value *entity.PolicyAcknowledgement) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create mempertahankan alias operasi INSERT.
func (r *PolicyAcknowledgementRepository) Create(ctx context.Context, value *entity.PolicyAcknowledgement) error {
	return r.Insert(ctx, value)
}

// FindByID membaca acknowledgement berdasarkan ID.
func (r *PolicyAcknowledgementRepository) FindByID(ctx context.Context, key string) (*entity.PolicyAcknowledgement, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *PolicyAcknowledgementRepository) SelectByID(ctx context.Context, key string) (*entity.PolicyAcknowledgement, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengambil row lock dalam transaksi aktif.
func (r *PolicyAcknowledgementRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PolicyAcknowledgement, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PolicyAcknowledgementRepository) find(ctx context.Context, key string, lock bool) (*entity.PolicyAcknowledgement, error) {
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
	value := &entity.PolicyAcknowledgement{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman acknowledgement.
func (r *PolicyAcknowledgementRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias pembacaan berhalaman.
func (r *PolicyAcknowledgementRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy menambahkan satu filter kolom yang diverifikasi terhadap metadata model.
func (r *PolicyAcknowledgementRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum urutan, limit dan offset.
func (r *PolicyAcknowledgementRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
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
	rows := make([]entity.PolicyAcknowledgement, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh acknowledgement.
func (r *PolicyAcknowledgementRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung baris dengan filter yang sama dengan daftar.
func (r *PolicyAcknowledgementRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key lengkap, tanpa upsert.
func (r *PolicyAcknowledgementRepository) Update(ctx context.Context, value *entity.PolicyAcknowledgement) error {
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

// DeleteByID menghapus satu baris dengan tetap mengikuti constraint database.
func (r *PolicyAcknowledgementRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PolicyAcknowledgement{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *PolicyAcknowledgementRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPolicyID mencari kecocokan policy_id dengan paginasi dan urutan primary key.
func (r *PolicyAcknowledgementRepository) FindByPolicyID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectBy(ctx, "policy_id", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *PolicyAcknowledgementRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByAcknowledgedByPrincipalID mencari kecocokan acknowledged_by_principal_id dengan paginasi dan urutan primary key.
func (r *PolicyAcknowledgementRepository) FindByAcknowledgedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectBy(ctx, "acknowledged_by_principal_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *PolicyAcknowledgementRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}
