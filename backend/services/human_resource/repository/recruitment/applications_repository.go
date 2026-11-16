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

// ApplicationRepository menyediakan kueri bertipe untuk hr_applications.
type ApplicationRepository struct {
	db *gorm.DB
}

var _ domainrepo.ApplicationRepository = (*ApplicationRepository)(nil)

// Method bernama objek memperjelas operasi lamaran dan kolom pencarian.
func (r *ApplicationRepository) InsertApplications(ctx context.Context, v *entity.Application) error {
	return r.Insert(ctx, v)
}
func (r *ApplicationRepository) UpdateApplications(ctx context.Context, v *entity.Application) error {
	return r.Update(ctx, v)
}
func (r *ApplicationRepository) DeleteApplicationsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *ApplicationRepository) FindApplicationsByUUID(ctx context.Context, id string) (*entity.Application, error) {
	return r.FindByID(ctx, id)
}
func (r *ApplicationRepository) FindApplicationsByUUIDForUpdate(ctx context.Context, id string) (*entity.Application, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *ApplicationRepository) GetAllApplications(ctx context.Context, limit, offset int) ([]entity.Application, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationRepository) CountApplications(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *ApplicationRepository) SelectApplicationsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Application, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *ApplicationRepository) CountApplicationsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *ApplicationRepository) FindApplicationsByCandidateID(ctx context.Context, v string, limit, offset int) ([]entity.Application, error) {
	return r.FindByCandidateID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByRequisitionID(ctx context.Context, v string, limit, offset int) ([]entity.Application, error) {
	return r.FindByRequisitionID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByApplicationNumber(ctx context.Context, v string, limit, offset int) ([]entity.Application, error) {
	return r.FindByApplicationNumber(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByStatus(ctx context.Context, v entity.ApplicationStatus, limit, offset int) ([]entity.Application, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByCurrentStageCode(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindByCurrentStageCode(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByAssignedRecruiterPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindByAssignedRecruiterPrincipalID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByDispositionReasonCode(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindByDispositionReasonCode(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByDecisionByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindByDecisionByPrincipalID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByHiredWorkerID(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindByHiredWorkerID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsBySourcePostingID(ctx context.Context, v *string, limit, offset int) ([]entity.Application, error) {
	return r.FindBySourcePostingID(ctx, v, limit, offset)
}
func (r *ApplicationRepository) FindApplicationsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.Application, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

// NewApplicationRepository memvalidasi koneksi dan menyiapkan repository.
func NewApplicationRepository(db *gorm.DB) (*ApplicationRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &ApplicationRepository{db: db}, nil
}

func (r *ApplicationRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Application{})
}

// Insert menyimpan lamaran dengan primary key dan audit dari service.
func (r *ApplicationRepository) Insert(ctx context.Context, value *entity.Application) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *ApplicationRepository) Create(ctx context.Context, value *entity.Application) error {
	return r.Insert(ctx, value)
}
func (r *ApplicationRepository) FindByID(ctx context.Context, key string) (*entity.Application, error) {
	return r.find(ctx, key, false)
}
func (r *ApplicationRepository) SelectByID(ctx context.Context, key string) (*entity.Application, error) {
	return r.FindByID(ctx, key)
}
func (r *ApplicationRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Application, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *ApplicationRepository) find(ctx context.Context, key string, lock bool) (*entity.Application, error) {
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
	value := new(entity.Application)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *ApplicationRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Application, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *ApplicationRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Application, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *ApplicationRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Application, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter lamaran sebelum pagination dengan urutan primary key stabil.
func (r *ApplicationRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Application, error) {
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
	rows := make([]entity.Application, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *ApplicationRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *ApplicationRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot lamaran non-key, termasuk nol/NULL, tanpa upsert.
func (r *ApplicationRepository) Update(ctx context.Context, value *entity.Application) error {
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
func (r *ApplicationRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Application{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *ApplicationRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCandidateID mencari kecocokan candidate_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByCandidateID(ctx context.Context, value string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "candidate_id", value, limit, offset)
}

// FindByRequisitionID mencari kecocokan requisition_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByRequisitionID(ctx context.Context, value string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "requisition_id", value, limit, offset)
}

// FindByApplicationNumber mencari kecocokan application_number dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByApplicationNumber(ctx context.Context, value string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "application_number", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByStatus(ctx context.Context, value entity.ApplicationStatus, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByCurrentStageCode mencari kecocokan current_stage_code dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByCurrentStageCode(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "current_stage_code", value, limit, offset)
}

// FindByAssignedRecruiterPrincipalID mencari kecocokan assigned_recruiter_principal_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByAssignedRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "assigned_recruiter_principal_id", value, limit, offset)
}

// FindByDispositionReasonCode mencari kecocokan disposition_reason_code dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByDispositionReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "disposition_reason_code", value, limit, offset)
}

// FindByDecisionByPrincipalID mencari kecocokan decision_by_principal_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByDecisionByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "decision_by_principal_id", value, limit, offset)
}

// FindByHiredWorkerID mencari kecocokan hired_worker_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByHiredWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "hired_worker_id", value, limit, offset)
}

// FindBySourcePostingID mencari kecocokan source_posting_id dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindBySourcePostingID(ctx context.Context, value *string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "source_posting_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *ApplicationRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.Application, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
