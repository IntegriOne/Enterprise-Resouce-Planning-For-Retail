// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package talent

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// TalentFrameworkRepository menyediakan kueri bertipe untuk hr_talent_frameworks.
type TalentFrameworkRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentFrameworkRepository = (*TalentFrameworkRepository)(nil)

// NewTalentFrameworkRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentFrameworkRepository(db *gorm.DB) (*TalentFrameworkRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentFrameworkRepository{db: db}, nil
}

// Operasi bernama objek memudahkan penelusuran kontrak framework pada service.
func (r *TalentFrameworkRepository) InsertTalentFrameworks(ctx context.Context, value *entity.TalentFramework) error {
	return r.Insert(ctx, value)
}
func (r *TalentFrameworkRepository) UpdateTalentFrameworks(ctx context.Context, value *entity.TalentFramework) error {
	return r.Update(ctx, value)
}
func (r *TalentFrameworkRepository) DeleteTalentFrameworksByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentFrameworkRepository) FindTalentFrameworksByUUID(ctx context.Context, id string) (*entity.TalentFramework, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentFrameworkRepository) FindTalentFrameworksByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentFramework, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentFrameworkRepository) GetAllTalentFrameworks(ctx context.Context, limit, offset int) ([]entity.TalentFramework, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentFrameworkRepository) CountTalentFrameworks(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentFrameworkRepository) SelectTalentFrameworksWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentFrameworkRepository) CountTalentFrameworksWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentFrameworkRepository) FindTalentFrameworksByFrameworkCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.FindByFrameworkCode(ctx, value, limit, offset)
}
func (r *TalentFrameworkRepository) FindTalentFrameworksByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.FindByRubricDocumentVersionReferenceID(ctx, value, limit, offset)
}
func (r *TalentFrameworkRepository) FindTalentFrameworksByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}

func (r *TalentFrameworkRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentFramework{})
}

// Insert menyimpan snapshot framework dengan RETURNING dari PostgreSQL.
func (r *TalentFrameworkRepository) Insert(ctx context.Context, value *entity.TalentFramework) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias INSERT tanpa upsert.
func (r *TalentFrameworkRepository) Create(ctx context.Context, value *entity.TalentFramework) error {
	return r.Insert(ctx, value)
}

// FindByID membaca framework berdasarkan primary key.
func (r *TalentFrameworkRepository) FindByID(ctx context.Context, key string) (*entity.TalentFramework, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *TalentFrameworkRepository) SelectByID(ctx context.Context, key string) (*entity.TalentFramework, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci framework dalam transaksi aktif.
func (r *TalentFrameworkRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentFramework, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentFrameworkRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentFramework, error) {
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
	var value entity.TalentFramework
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}

// GetAll membaca satu halaman dalam urutan primary key stabil.
func (r *TalentFrameworkRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan halaman.
func (r *TalentFrameworkRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentFramework, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy membaca halaman berdasarkan satu kolom yang diverifikasi.
func (r *TalentFrameworkRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination.
func (r *TalentFrameworkRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentFramework, error) {
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
	rows := make([]entity.TalentFramework, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh framework.
func (r *TalentFrameworkRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung baris yang memenuhi filter.
func (r *TalentFrameworkRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key termasuk nol/NULL tanpa upsert.
func (r *TalentFrameworkRepository) Update(ctx context.Context, value *entity.TalentFramework) error {
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

// DeleteByID menghapus satu framework sesuai constraint database.
func (r *TalentFrameworkRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentFramework{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *TalentFrameworkRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByFrameworkCode mencari kecocokan framework_code dengan paginasi dan urutan primary key.
func (r *TalentFrameworkRepository) FindByFrameworkCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectBy(ctx, "framework_code", value, limit, offset)
}

// FindByRubricDocumentVersionReferenceID mencari kecocokan rubric_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *TalentFrameworkRepository) FindByRubricDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectBy(ctx, "rubric_document_version_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *TalentFrameworkRepository) FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentFramework, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
