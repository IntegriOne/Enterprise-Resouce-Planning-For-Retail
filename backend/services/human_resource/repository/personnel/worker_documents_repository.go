// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package personnel

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkerDocumentRepository menyediakan kueri bertipe untuk hr_worker_documents.
type WorkerDocumentRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerDocumentRepository = (*WorkerDocumentRepository)(nil)

// Method bernama objek menyediakan kontrak eksplisit dokumen pekerja.
func (r *WorkerDocumentRepository) InsertWorkerDocuments(ctx context.Context, v *entity.WorkerDocument) error {
	return r.Insert(ctx, v)
}
func (r *WorkerDocumentRepository) UpdateWorkerDocuments(ctx context.Context, v *entity.WorkerDocument) error {
	return r.Update(ctx, v)
}
func (r *WorkerDocumentRepository) DeleteWorkerDocumentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByUUID(ctx context.Context, id string) (*entity.WorkerDocument, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkerDocument, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerDocumentRepository) GetAllWorkerDocuments(ctx context.Context, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerDocumentRepository) CountWorkerDocuments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerDocumentRepository) SelectWorkerDocumentsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkerDocumentRepository) CountWorkerDocumentsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByDocumentType(ctx context.Context, v string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByDocumentType(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByDocumentReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByDocumentReferenceID(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByIssuingCountryCode(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByIssuingCountryCode(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByVerificationStatus(ctx context.Context, v string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByVerificationStatus(ctx, v, limit, offset)
}
func (r *WorkerDocumentRepository) FindWorkerDocumentsByVerifiedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.FindByVerifiedByPrincipalID(ctx, v, limit, offset)
}

// NewWorkerDocumentRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerDocumentRepository(db *gorm.DB) (*WorkerDocumentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerDocumentRepository{db: db}, nil
}

func (r *WorkerDocumentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerDocument{})
}

// Insert menyimpan referensi dokumen pekerja, bukan berkas milik DRM.
func (r *WorkerDocumentRepository) Insert(ctx context.Context, value *entity.WorkerDocument) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerDocumentRepository) Create(ctx context.Context, value *entity.WorkerDocument) error {
	return r.Insert(ctx, value)
}
func (r *WorkerDocumentRepository) FindByID(ctx context.Context, key string) (*entity.WorkerDocument, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerDocumentRepository) SelectByID(ctx context.Context, key string) (*entity.WorkerDocument, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerDocumentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkerDocument, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerDocumentRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkerDocument, error) {
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
	value := new(entity.WorkerDocument)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkerDocumentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerDocumentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerDocumentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere membatasi referensi dokumen sesuai filter sebelum pagination.
func (r *WorkerDocumentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerDocument, error) {
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
	rows := make([]entity.WorkerDocument, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerDocumentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerDocumentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key tanpa upsert atau penggantian primary key.
func (r *WorkerDocumentRepository) Update(ctx context.Context, value *entity.WorkerDocument) error {
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
func (r *WorkerDocumentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerDocument{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerDocumentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByDocumentType mencari kecocokan document_type dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByDocumentType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "document_type", value, limit, offset)
}

// FindByDocumentReferenceID mencari kecocokan document_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByDocumentReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "document_reference_id", value, limit, offset)
}

// FindByDocumentVersionReferenceID mencari kecocokan document_version_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "document_version_reference_id", value, limit, offset)
}

// FindByIssuingCountryCode mencari kecocokan issuing_country_code dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByIssuingCountryCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "issuing_country_code", value, limit, offset)
}

// FindByVerificationStatus mencari kecocokan verification_status dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByVerificationStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "verification_status", value, limit, offset)
}

// FindByVerifiedByPrincipalID mencari kecocokan verified_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerDocumentRepository) FindByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerDocument, error) {
	return r.SelectBy(ctx, "verified_by_principal_id", value, limit, offset)
}
