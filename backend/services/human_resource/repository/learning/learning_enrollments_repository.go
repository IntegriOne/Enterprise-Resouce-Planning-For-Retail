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

// LearningEnrollmentRepository menyediakan kueri bertipe untuk hr_learning_enrollments.
type LearningEnrollmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.LearningEnrollmentRepository = (*LearningEnrollmentRepository)(nil)

// NewLearningEnrollmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewLearningEnrollmentRepository(db *gorm.DB) (*LearningEnrollmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &LearningEnrollmentRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak enrollment yang eksplisit.
func (r *LearningEnrollmentRepository) InsertLearningEnrollments(ctx context.Context, value *entity.LearningEnrollment) error {
	return r.Insert(ctx, value)
}
func (r *LearningEnrollmentRepository) UpdateLearningEnrollments(ctx context.Context, value *entity.LearningEnrollment) error {
	return r.Update(ctx, value)
}
func (r *LearningEnrollmentRepository) DeleteLearningEnrollmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByUUID(ctx context.Context, id string) (*entity.LearningEnrollment, error) {
	return r.FindByID(ctx, id)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.LearningEnrollment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *LearningEnrollmentRepository) GetAllLearningEnrollments(ctx context.Context, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningEnrollmentRepository) CountLearningEnrollments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LearningEnrollmentRepository) SelectLearningEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LearningEnrollmentRepository) CountLearningEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByCourseID(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByEnrollmentStatus(ctx context.Context, value entity.LearningStatus, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByEnrollmentStatus(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByAssignedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByAssignedByPrincipalID(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByCredentialReferenceID(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByFundingReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByFundingReferenceID(ctx, value, limit, offset)
}
func (r *LearningEnrollmentRepository) FindLearningEnrollmentsByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}

func (r *LearningEnrollmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.LearningEnrollment{})
}

// Insert menyimpan enrollment sesuai constraint database.
func (r *LearningEnrollmentRepository) Insert(ctx context.Context, value *entity.LearningEnrollment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *LearningEnrollmentRepository) Create(ctx context.Context, value *entity.LearningEnrollment) error {
	return r.Insert(ctx, value)
}
func (r *LearningEnrollmentRepository) FindByID(ctx context.Context, key string) (*entity.LearningEnrollment, error) {
	return r.find(ctx, key, false)
}
func (r *LearningEnrollmentRepository) SelectByID(ctx context.Context, key string) (*entity.LearningEnrollment, error) {
	return r.FindByID(ctx, key)
}
func (r *LearningEnrollmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.LearningEnrollment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *LearningEnrollmentRepository) find(ctx context.Context, key string, lock bool) (*entity.LearningEnrollment, error) {
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
	var value entity.LearningEnrollment
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *LearningEnrollmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *LearningEnrollmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LearningEnrollmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *LearningEnrollmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LearningEnrollment, error) {
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
	rows := make([]entity.LearningEnrollment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *LearningEnrollmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *LearningEnrollmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *LearningEnrollmentRepository) Update(ctx context.Context, value *entity.LearningEnrollment) error {
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
func (r *LearningEnrollmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.LearningEnrollment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *LearningEnrollmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByCourseID mencari kecocokan course_id dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByCourseID(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "course_id", value, limit, offset)
}

// FindByEnrollmentStatus mencari kecocokan enrollment_status dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByEnrollmentStatus(ctx context.Context, value entity.LearningStatus, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "enrollment_status", value, limit, offset)
}

// FindByAssignedByPrincipalID mencari kecocokan assigned_by_principal_id dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByAssignedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "assigned_by_principal_id", value, limit, offset)
}

// FindByCredentialReferenceID mencari kecocokan credential_reference_id dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByCredentialReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "credential_reference_id", value, limit, offset)
}

// FindByFundingReferenceID mencari kecocokan funding_reference_id dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByFundingReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "funding_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *LearningEnrollmentRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LearningEnrollment, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
