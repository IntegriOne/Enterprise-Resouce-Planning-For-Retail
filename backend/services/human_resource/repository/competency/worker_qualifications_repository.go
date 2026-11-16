// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package competency

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkerQualificationRepository menyediakan kueri bertipe untuk hr_worker_qualifications.
type WorkerQualificationRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerQualificationRepository = (*WorkerQualificationRepository)(nil)

// NewWorkerQualificationRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerQualificationRepository(db *gorm.DB) (*WorkerQualificationRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerQualificationRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak kualifikasi yang eksplisit.
func (r *WorkerQualificationRepository) InsertWorkerQualifications(ctx context.Context, value *entity.WorkerQualification) error {
	return r.Insert(ctx, value)
}
func (r *WorkerQualificationRepository) UpdateWorkerQualifications(ctx context.Context, value *entity.WorkerQualification) error {
	return r.Update(ctx, value)
}
func (r *WorkerQualificationRepository) DeleteWorkerQualificationsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByUUID(ctx context.Context, id string) (*entity.WorkerQualification, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkerQualification, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerQualificationRepository) GetAllWorkerQualifications(ctx context.Context, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerQualificationRepository) CountWorkerQualifications(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerQualificationRepository) SelectWorkerQualificationsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *WorkerQualificationRepository) CountWorkerQualificationsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByQualificationType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByQualificationType(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByQualificationCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByQualificationCode(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByIssuerPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByIssuerPartyReferenceID(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByCredentialReferenceID(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByVerificationStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByVerificationStatus(ctx, value, limit, offset)
}
func (r *WorkerQualificationRepository) FindWorkerQualificationsByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.FindByVerifiedByPrincipalID(ctx, value, limit, offset)
}

func (r *WorkerQualificationRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerQualification{})
}

// Insert menyimpan kualifikasi sesuai constraint database.
func (r *WorkerQualificationRepository) Insert(ctx context.Context, value *entity.WorkerQualification) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerQualificationRepository) Create(ctx context.Context, value *entity.WorkerQualification) error {
	return r.Insert(ctx, value)
}
func (r *WorkerQualificationRepository) FindByID(ctx context.Context, key string) (*entity.WorkerQualification, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerQualificationRepository) SelectByID(ctx context.Context, key string) (*entity.WorkerQualification, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerQualificationRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkerQualification, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerQualificationRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkerQualification, error) {
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
	var value entity.WorkerQualification
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *WorkerQualificationRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerQualificationRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerQualificationRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *WorkerQualificationRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerQualification, error) {
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
	rows := make([]entity.WorkerQualification, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerQualificationRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerQualificationRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key, termasuk nol/NULL, tanpa upsert.
func (r *WorkerQualificationRepository) Update(ctx context.Context, value *entity.WorkerQualification) error {
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
func (r *WorkerQualificationRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerQualification{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerQualificationRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByQualificationType mencari kecocokan qualification_type dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByQualificationType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "qualification_type", value, limit, offset)
}

// FindByQualificationCode mencari kecocokan qualification_code dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByQualificationCode(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "qualification_code", value, limit, offset)
}

// FindByIssuerPartyReferenceID mencari kecocokan issuer_party_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByIssuerPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "issuer_party_reference_id", value, limit, offset)
}

// FindByCredentialReferenceID mencari kecocokan credential_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "credential_reference_id", value, limit, offset)
}

// FindByVerificationStatus mencari kecocokan verification_status dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByVerificationStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "verification_status", value, limit, offset)
}

// FindByVerifiedByPrincipalID mencari kecocokan verified_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerQualificationRepository) FindByVerifiedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerQualification, error) {
	return r.SelectBy(ctx, "verified_by_principal_id", value, limit, offset)
}
