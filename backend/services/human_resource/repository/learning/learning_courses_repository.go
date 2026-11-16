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

// LearningCourseRepository menyediakan kueri bertipe untuk hr_learning_courses.
type LearningCourseRepository struct {
	db *gorm.DB
}

var _ domainrepo.LearningCourseRepository = (*LearningCourseRepository)(nil)

// NewLearningCourseRepository memvalidasi koneksi dan menyiapkan repository.
func NewLearningCourseRepository(db *gorm.DB) (*LearningCourseRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LearningCourseRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak kursus yang eksplisit.
func (r *LearningCourseRepository) InsertLearningCourses(ctx context.Context, value *entity.LearningCourse) error {
	return r.Insert(ctx, value)
}
func (r *LearningCourseRepository) UpdateLearningCourses(ctx context.Context, value *entity.LearningCourse) error {
	return r.Update(ctx, value)
}
func (r *LearningCourseRepository) DeleteLearningCoursesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *LearningCourseRepository) FindLearningCoursesByUUID(ctx context.Context, id string) (*entity.LearningCourse, error) {
	return r.FindByID(ctx, id)
}
func (r *LearningCourseRepository) FindLearningCoursesByUUIDForUpdate(ctx context.Context, id string) (*entity.LearningCourse, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *LearningCourseRepository) GetAllLearningCourses(ctx context.Context, limit, offset int) ([]entity.LearningCourse, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningCourseRepository) CountLearningCourses(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LearningCourseRepository) SelectLearningCoursesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LearningCourseRepository) CountLearningCoursesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LearningCourseRepository) FindLearningCoursesByCourseCode(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.FindByCourseCode(ctx, value, limit, offset)
}
func (r *LearningCourseRepository) FindLearningCoursesByCourseType(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.FindByCourseType(ctx, value, limit, offset)
}
func (r *LearningCourseRepository) FindLearningCoursesByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.FindByProviderPartyReferenceID(ctx, value, limit, offset)
}

func (r *LearningCourseRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LearningCourse{})
}

// Insert menyimpan kursus sesuai constraint database.
func (r *LearningCourseRepository) Insert(ctx context.Context, value *entity.LearningCourse) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *LearningCourseRepository) Create(ctx context.Context, value *entity.LearningCourse) error {
	return r.Insert(ctx, value)
}
func (r *LearningCourseRepository) FindByID(ctx context.Context, key string) (*entity.LearningCourse, error) {
	return r.find(ctx, key, false)
}
func (r *LearningCourseRepository) SelectByID(ctx context.Context, key string) (*entity.LearningCourse, error) {
	return r.FindByID(ctx, key)
}
func (r *LearningCourseRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LearningCourse, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *LearningCourseRepository) find(ctx context.Context, key string, lock bool) (*entity.LearningCourse, error) {
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
	var value entity.LearningCourse
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *LearningCourseRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *LearningCourseRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LearningCourse, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningCourseRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *LearningCourseRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningCourse, error) {
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
	rows := make([]entity.LearningCourse, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *LearningCourseRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *LearningCourseRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LearningCourseRepository) Update(ctx context.Context, value *entity.LearningCourse) error {
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
func (r *LearningCourseRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LearningCourse{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *LearningCourseRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCourseCode mencari kecocokan course_code dengan paginasi dan urutan primary key.
func (r *LearningCourseRepository) FindByCourseCode(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectBy(ctx, "course_code", value, limit, offset)
}

// FindByCourseType mencari kecocokan course_type dengan paginasi dan urutan primary key.
func (r *LearningCourseRepository) FindByCourseType(ctx context.Context, value string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectBy(ctx, "course_type", value, limit, offset)
}

// FindByProviderPartyReferenceID mencari kecocokan provider_party_reference_id dengan paginasi dan urutan primary key.
func (r *LearningCourseRepository) FindByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningCourse, error) {
	return r.SelectBy(ctx, "provider_party_reference_id", value, limit, offset)
}
