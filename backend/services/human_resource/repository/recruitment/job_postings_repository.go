// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package recruitment

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// JobPostingRepository menyediakan kueri bertipe untuk hr_job_postings.
type JobPostingRepository struct {
	db *gorm.DB
}

var _ domainrepo.JobPostingRepository = (*JobPostingRepository)(nil)

// Method bernama objek menyediakan API eksplisit metadata lowongan.
func (r *JobPostingRepository) InsertJobPostings(ctx context.Context, v *entity.JobPosting) error {
	return r.Insert(ctx, v)
}
func (r *JobPostingRepository) UpdateJobPostings(ctx context.Context, v *entity.JobPosting) error {
	return r.Update(ctx, v)
}
func (r *JobPostingRepository) DeleteJobPostingsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *JobPostingRepository) FindJobPostingsByUUID(ctx context.Context, id string) (*entity.JobPosting, error) {
	return r.FindByID(ctx, id)
}
func (r *JobPostingRepository) FindJobPostingsByUUIDForUpdate(ctx context.Context, id string) (*entity.JobPosting, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *JobPostingRepository) GetAllJobPostings(ctx context.Context, limit, offset int) ([]entity.JobPosting, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobPostingRepository) CountJobPostings(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *JobPostingRepository) SelectJobPostingsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *JobPostingRepository) CountJobPostingsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *JobPostingRepository) FindJobPostingsByRequisitionID(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByRequisitionID(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByChannelCode(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByChannelCode(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByContentDocumentVersionReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByContentDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByStatus(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByCreatedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByWithdrawalReasonCode(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByWithdrawalReasonCode(ctx, v, limit, offset)
}
func (r *JobPostingRepository) FindJobPostingsByWithdrawnByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.JobPosting, error) {
	return r.FindByWithdrawnByPrincipalID(ctx, v, limit, offset)
}

// NewJobPostingRepository memvalidasi koneksi dan menyiapkan repository.
func NewJobPostingRepository(db *gorm.DB) (*JobPostingRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &JobPostingRepository{db: db}, nil
}

func (r *JobPostingRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.JobPosting{})
}

// Insert menyimpan metadata lowongan tanpa menerbitkan ke kanal eksternal.
func (r *JobPostingRepository) Insert(ctx context.Context, value *entity.JobPosting) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *JobPostingRepository) Create(ctx context.Context, value *entity.JobPosting) error {
	return r.Insert(ctx, value)
}
func (r *JobPostingRepository) FindByID(ctx context.Context, key string) (*entity.JobPosting, error) {
	return r.find(ctx, key, false)
}
func (r *JobPostingRepository) SelectByID(ctx context.Context, key string) (*entity.JobPosting, error) {
	return r.FindByID(ctx, key)
}
func (r *JobPostingRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.JobPosting, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *JobPostingRepository) find(ctx context.Context, key string, lock bool) (*entity.JobPosting, error) {
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
	value := new(entity.JobPosting)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *JobPostingRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *JobPostingRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.JobPosting, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *JobPostingRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter lowongan sebelum pagination dengan urutan ID stabil.
func (r *JobPostingRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.JobPosting, error) {
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
	rows := make([]entity.JobPosting, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *JobPostingRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *JobPostingRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key termasuk nilai nol dan NULL tanpa upsert.
func (r *JobPostingRepository) Update(ctx context.Context, value *entity.JobPosting) error {
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
func (r *JobPostingRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.JobPosting{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *JobPostingRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRequisitionID mencari kecocokan requisition_id dengan paginasi dan urutan primary key.
func (r *JobPostingRepository) FindByRequisitionID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "requisition_id", value, limit, offset)
}

// FindByChannelCode mencari kecocokan channel_code dengan paginasi dan urutan primary key.
func (r *JobPostingRepository) FindByChannelCode(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "channel_code", value, limit, offset)
}

// FindByContentDocumentVersionReferenceID mencari kecocokan content_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *JobPostingRepository) FindByContentDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "content_document_version_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *JobPostingRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *JobPostingRepository) FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}

// FindByWithdrawalReasonCode mencari alasan penarikan dengan paginasi stabil.
func (r *JobPostingRepository) FindByWithdrawalReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "withdrawal_reason_code", value, limit, offset)
}

// FindByWithdrawnByPrincipalID mencari actor penarikan dengan paginasi stabil.
func (r *JobPostingRepository) FindByWithdrawnByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.JobPosting, error) {
	return r.SelectBy(ctx, "withdrawn_by_principal_id", value, limit, offset)
}
