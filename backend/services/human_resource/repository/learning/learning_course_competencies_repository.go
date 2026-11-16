// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package learning

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// LearningCourseCompetencyRepository menyediakan kueri bertipe untuk hr_learning_course_competencies.
type LearningCourseCompetencyRepository struct {
	db *gorm.DB
}

var _ domainrepo.LearningCourseCompetencyRepository = (*LearningCourseCompetencyRepository)(nil)

// NewLearningCourseCompetencyRepository memvalidasi koneksi dan menyiapkan repository.
func NewLearningCourseCompetencyRepository(db *gorm.DB) (*LearningCourseCompetencyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LearningCourseCompetencyRepository{db: db}, nil
}

// Operasi bernama objek menyebut kedua bagian primary key secara eksplisit.
func (r *LearningCourseCompetencyRepository) InsertLearningCourseCompetencies(ctx context.Context, value *entity.LearningCourseCompetency) error {
	return r.Insert(ctx, value)
}
func (r *LearningCourseCompetencyRepository) UpdateLearningCourseCompetencies(ctx context.Context, value *entity.LearningCourseCompetency) error {
	return r.Update(ctx, value)
}
func (r *LearningCourseCompetencyRepository) DeleteLearningCourseCompetenciesByCourseIDAndCompetencyID(ctx context.Context, key entity.LearningCourseCompetencyKey) error {
	return r.DeleteByID(ctx, key)
}
func (r *LearningCourseCompetencyRepository) FindLearningCourseCompetenciesByCourseIDAndCompetencyID(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	return r.FindByID(ctx, key)
}
func (r *LearningCourseCompetencyRepository) FindLearningCourseCompetenciesByCourseIDAndCompetencyIDForUpdate(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	return r.FindByIDForUpdate(ctx, key)
}
func (r *LearningCourseCompetencyRepository) GetAllLearningCourseCompetencies(ctx context.Context, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningCourseCompetencyRepository) CountLearningCourseCompetencies(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LearningCourseCompetencyRepository) SelectLearningCourseCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LearningCourseCompetencyRepository) CountLearningCourseCompetenciesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LearningCourseCompetencyRepository) FindLearningCourseCompetenciesByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.FindByCourseID(ctx, value, limit, offset)
}
func (r *LearningCourseCompetencyRepository) FindLearningCourseCompetenciesByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.FindByCompetencyID(ctx, value, limit, offset)
}

func (r *LearningCourseCompetencyRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LearningCourseCompetency{})
}

// Insert menyimpan hubungan kursus dan kompetensi sesuai constraint database.
func (r *LearningCourseCompetencyRepository) Insert(ctx context.Context, value *entity.LearningCourseCompetency) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *LearningCourseCompetencyRepository) Create(ctx context.Context, value *entity.LearningCourseCompetency) error {
	return r.Insert(ctx, value)
}
func (r *LearningCourseCompetencyRepository) FindByID(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	return r.find(ctx, key, false)
}
func (r *LearningCourseCompetencyRepository) SelectByID(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	return r.FindByID(ctx, key)
}
func (r *LearningCourseCompetencyRepository) FindByIDForUpdate(ctx context.Context, key entity.LearningCourseCompetencyKey) (*entity.LearningCourseCompetency, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *LearningCourseCompetencyRepository) find(ctx context.Context, key entity.LearningCourseCompetencyKey, lock bool) (*entity.LearningCourseCompetency, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	if err := pgstore.ValidateKeys(key.CourseID, key.CompetencyID); err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var value entity.LearningCourseCompetency
	if err := db.Where("course_id = ? AND competency_id = ?", key.CourseID, key.CompetencyID).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *LearningCourseCompetencyRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *LearningCourseCompetencyRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningCourseCompetencyRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman dengan urutan kedua bagian primary key.
func (r *LearningCourseCompetencyRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningCourseCompetency, error) {
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
	rows := make([]entity.LearningCourseCompetency, 0)
	if err := db.Order("course_id ASC, competency_id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *LearningCourseCompetencyRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *LearningCourseCompetencyRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LearningCourseCompetencyRepository) Update(ctx context.Context, value *entity.LearningCourseCompetency) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(value.CourseID, value.CompetencyID); err != nil {
		return err
	}
	result := db.Where("course_id = ? AND competency_id = ?", value.CourseID, value.CompetencyID).Select("*").Omit("course_id", "competency_id").Updates(value)
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *LearningCourseCompetencyRepository) DeleteByID(ctx context.Context, key entity.LearningCourseCompetencyKey) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key.CourseID, key.CompetencyID); err != nil {
		return err
	}
	result := db.Where("course_id = ? AND competency_id = ?", key.CourseID, key.CompetencyID).Delete(&entity.LearningCourseCompetency{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *LearningCourseCompetencyRepository) Delete(ctx context.Context, key entity.LearningCourseCompetencyKey) error {
	return r.DeleteByID(ctx, key)
}

// FindByCourseID mencari kecocokan course_id dengan paginasi dan urutan primary key.
func (r *LearningCourseCompetencyRepository) FindByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.SelectBy(ctx, "course_id", value, limit, offset)
}

// FindByCompetencyID mencari kecocokan competency_id dengan paginasi dan urutan primary key.
func (r *LearningCourseCompetencyRepository) FindByCompetencyID(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourseCompetency, error) {
	return r.SelectBy(ctx, "competency_id", value, limit, offset)
}
