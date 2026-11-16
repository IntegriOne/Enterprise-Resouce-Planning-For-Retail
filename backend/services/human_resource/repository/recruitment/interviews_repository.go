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

// InterviewRepository menyediakan kueri bertipe untuk hr_interviews.
type InterviewRepository struct {
	db *gorm.DB
}

var _ domainrepo.InterviewRepository = (*InterviewRepository)(nil)

// Method bernama objek menyediakan API eksplisit data wawancara.
func (r *InterviewRepository) InsertInterviews(ctx context.Context, v *entity.Interview) error {
	return r.Insert(ctx, v)
}
func (r *InterviewRepository) UpdateInterviews(ctx context.Context, v *entity.Interview) error {
	return r.Update(ctx, v)
}
func (r *InterviewRepository) DeleteInterviewsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *InterviewRepository) FindInterviewsByUUID(ctx context.Context, id string) (*entity.Interview, error) {
	return r.FindByID(ctx, id)
}
func (r *InterviewRepository) FindInterviewsByUUIDForUpdate(ctx context.Context, id string) (*entity.Interview, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *InterviewRepository) GetAllInterviews(ctx context.Context, limit, offset int) ([]entity.Interview, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewRepository) CountInterviews(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *InterviewRepository) SelectInterviewsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Interview, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *InterviewRepository) CountInterviewsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *InterviewRepository) FindInterviewsByApplicationID(ctx context.Context, v string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByApplicationID(ctx, v, limit, offset)
}
func (r *InterviewRepository) FindInterviewsByInterviewType(ctx context.Context, v string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByInterviewType(ctx, v, limit, offset)
}
func (r *InterviewRepository) FindInterviewsByLocationReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByLocationReferenceID(ctx, v, limit, offset)
}
func (r *InterviewRepository) FindInterviewsByMeetingReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByMeetingReferenceID(ctx, v, limit, offset)
}
func (r *InterviewRepository) FindInterviewsByStatus(ctx context.Context, v string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *InterviewRepository) FindInterviewsByCreatedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Interview, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}

// NewInterviewRepository memvalidasi koneksi dan menyiapkan repository.
func NewInterviewRepository(db *gorm.DB) (*InterviewRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &InterviewRepository{db: db}, nil
}

func (r *InterviewRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Interview{})
}

// Insert menyimpan wawancara tanpa efek samping di luar database HRM.
func (r *InterviewRepository) Insert(ctx context.Context, value *entity.Interview) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *InterviewRepository) Create(ctx context.Context, value *entity.Interview) error {
	return r.Insert(ctx, value)
}
func (r *InterviewRepository) FindByID(ctx context.Context, key string) (*entity.Interview, error) {
	return r.find(ctx, key, false)
}
func (r *InterviewRepository) SelectByID(ctx context.Context, key string) (*entity.Interview, error) {
	return r.FindByID(ctx, key)
}
func (r *InterviewRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Interview, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *InterviewRepository) find(ctx context.Context, key string, lock bool) (*entity.Interview, error) {
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
	value := new(entity.Interview)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *InterviewRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Interview, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *InterviewRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Interview, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Interview, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter wawancara sebelum pagination berurutan ID.
func (r *InterviewRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Interview, error) {
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
	rows := make([]entity.Interview, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *InterviewRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *InterviewRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot non-key termasuk NULL tanpa upsert.
func (r *InterviewRepository) Update(ctx context.Context, value *entity.Interview) error {
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
func (r *InterviewRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Interview{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *InterviewRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByApplicationID mencari kecocokan application_id dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByApplicationID(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "application_id", value, limit, offset)
}

// FindByInterviewType mencari kecocokan interview_type dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByInterviewType(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "interview_type", value, limit, offset)
}

// FindByLocationReferenceID mencari kecocokan location_reference_id dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "location_reference_id", value, limit, offset)
}

// FindByMeetingReferenceID mencari kecocokan meeting_reference_id dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByMeetingReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "meeting_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *InterviewRepository) FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Interview, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}
