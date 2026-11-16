// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package benefits

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// BenefitEnrollmentRepository menyediakan kueri bertipe untuk hr_benefit_enrollments.
type BenefitEnrollmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.BenefitEnrollmentRepository = (*BenefitEnrollmentRepository)(nil)

func (r *BenefitEnrollmentRepository) InsertBenefitEnrollments(ctx context.Context, value *entity.BenefitEnrollment) error {
	return r.Insert(ctx, value)
}
func (r *BenefitEnrollmentRepository) UpdateBenefitEnrollments(ctx context.Context, value *entity.BenefitEnrollment) error {
	return r.Update(ctx, value)
}
func (r *BenefitEnrollmentRepository) DeleteBenefitEnrollmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByUUID(ctx context.Context, id string) (*entity.BenefitEnrollment, error) {
	return r.FindByID(ctx, id)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.BenefitEnrollment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *BenefitEnrollmentRepository) GetAllBenefitEnrollments(ctx context.Context, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *BenefitEnrollmentRepository) CountBenefitEnrollments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *BenefitEnrollmentRepository) SelectBenefitEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *BenefitEnrollmentRepository) CountBenefitEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByBenefitPlanID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByBenefitPlanID(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByEnrollmentStatus(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByEnrollmentStatus(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}

// NewBenefitEnrollmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewBenefitEnrollmentRepository(db *gorm.DB) (*BenefitEnrollmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &BenefitEnrollmentRepository{db: db}, nil
}

func (r *BenefitEnrollmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.BenefitEnrollment{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *BenefitEnrollmentRepository) Insert(ctx context.Context, value *entity.BenefitEnrollment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias INSERT.
func (r *BenefitEnrollmentRepository) Create(ctx context.Context, value *entity.BenefitEnrollment) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *BenefitEnrollmentRepository) FindByID(ctx context.Context, key string) (*entity.BenefitEnrollment, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *BenefitEnrollmentRepository) SelectByID(ctx context.Context, key string) (*entity.BenefitEnrollment, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *BenefitEnrollmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.BenefitEnrollment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *BenefitEnrollmentRepository) find(ctx context.Context, key string, lock bool) (*entity.BenefitEnrollment, error) {
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
	value := new(entity.BenefitEnrollment)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *BenefitEnrollmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *BenefitEnrollmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *BenefitEnrollmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *BenefitEnrollmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.BenefitEnrollment, error) {
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
	rows := make([]entity.BenefitEnrollment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *BenefitEnrollmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *BenefitEnrollmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot lengkap non-key termasuk nilai nol dan NULL.
func (r *BenefitEnrollmentRepository) Update(ctx context.Context, value *entity.BenefitEnrollment) error {
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

// DeleteByID menghapus satu record sesuai constraint database.
func (r *BenefitEnrollmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.BenefitEnrollment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *BenefitEnrollmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *BenefitEnrollmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByBenefitPlanID mencari kecocokan benefit_plan_id dengan paginasi dan urutan primary key.
func (r *BenefitEnrollmentRepository) FindByBenefitPlanID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectBy(ctx, "benefit_plan_id", value, limit, offset)
}

// FindByEnrollmentStatus mencari kecocokan enrollment_status dengan paginasi dan urutan primary key.
func (r *BenefitEnrollmentRepository) FindByEnrollmentStatus(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectBy(ctx, "enrollment_status", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *BenefitEnrollmentRepository) FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}
