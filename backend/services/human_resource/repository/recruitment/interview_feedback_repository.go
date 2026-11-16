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

// InterviewFeedbackRepository menyediakan kueri bertipe untuk hr_interview_feedback.
type InterviewFeedbackRepository struct {
	db *gorm.DB
}

var _ domainrepo.InterviewFeedbackRepository = (*InterviewFeedbackRepository)(nil)

// Method bernama objek mempertahankan nama tabel feedback tanpa plural buatan.
func (r *InterviewFeedbackRepository) InsertInterviewFeedback(ctx context.Context, v *entity.InterviewFeedback) error {
	return r.Insert(ctx, v)
}
func (r *InterviewFeedbackRepository) UpdateInterviewFeedback(ctx context.Context, v *entity.InterviewFeedback) error {
	return r.Update(ctx, v)
}
func (r *InterviewFeedbackRepository) DeleteInterviewFeedbackByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *InterviewFeedbackRepository) FindInterviewFeedbackByUUID(ctx context.Context, id string) (*entity.InterviewFeedback, error) {
	return r.FindByID(ctx, id)
}
func (r *InterviewFeedbackRepository) FindInterviewFeedbackByUUIDForUpdate(ctx context.Context, id string) (*entity.InterviewFeedback, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *InterviewFeedbackRepository) GetAllInterviewFeedback(ctx context.Context, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewFeedbackRepository) CountInterviewFeedback(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *InterviewFeedbackRepository) SelectInterviewFeedbackWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *InterviewFeedbackRepository) CountInterviewFeedbackWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *InterviewFeedbackRepository) FindInterviewFeedbackByInterviewID(ctx context.Context, v string, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.FindByInterviewID(ctx, v, limit, offset)
}
func (r *InterviewFeedbackRepository) FindInterviewFeedbackByInterviewerPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.FindByInterviewerPrincipalID(ctx, v, limit, offset)
}

// NewInterviewFeedbackRepository memvalidasi koneksi dan menyiapkan repository.
func NewInterviewFeedbackRepository(db *gorm.DB) (*InterviewFeedbackRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &InterviewFeedbackRepository{db: db}, nil
}

func (r *InterviewFeedbackRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.InterviewFeedback{})
}

// Insert menyimpan bukti feedback wawancara tanpa menentukan hasil rekrutmen.
func (r *InterviewFeedbackRepository) Insert(ctx context.Context, value *entity.InterviewFeedback) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *InterviewFeedbackRepository) Create(ctx context.Context, value *entity.InterviewFeedback) error {
	return r.Insert(ctx, value)
}
func (r *InterviewFeedbackRepository) FindByID(ctx context.Context, key string) (*entity.InterviewFeedback, error) {
	return r.find(ctx, key, false)
}
func (r *InterviewFeedbackRepository) SelectByID(ctx context.Context, key string) (*entity.InterviewFeedback, error) {
	return r.FindByID(ctx, key)
}
func (r *InterviewFeedbackRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.InterviewFeedback, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *InterviewFeedbackRepository) find(ctx context.Context, key string, lock bool) (*entity.InterviewFeedback, error) {
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
	value := new(entity.InterviewFeedback)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *InterviewFeedbackRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *InterviewFeedbackRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InterviewFeedbackRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter feedback sebelum pagination berurutan ID.
func (r *InterviewFeedbackRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.InterviewFeedback, error) {
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
	rows := make([]entity.InterviewFeedback, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *InterviewFeedbackRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *InterviewFeedbackRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key dan tunduk pada pembekuan bukti di database.
func (r *InterviewFeedbackRepository) Update(ctx context.Context, value *entity.InterviewFeedback) error {
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
func (r *InterviewFeedbackRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.InterviewFeedback{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *InterviewFeedbackRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByInterviewID mencari kecocokan interview_id dengan paginasi dan urutan primary key.
func (r *InterviewFeedbackRepository) FindByInterviewID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.SelectBy(ctx, "interview_id", value, limit, offset)
}

// FindByInterviewerPrincipalID mencari kecocokan interviewer_principal_id dengan paginasi dan urutan primary key.
func (r *InterviewFeedbackRepository) FindByInterviewerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.InterviewFeedback, error) {
	return r.SelectBy(ctx, "interviewer_principal_id", value, limit, offset)
}
