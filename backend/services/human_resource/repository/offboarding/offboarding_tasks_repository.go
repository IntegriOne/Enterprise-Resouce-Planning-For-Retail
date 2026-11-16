// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package offboarding

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// OffboardingTaskRepository menyediakan kueri bertipe untuk hr_offboarding_tasks.
type OffboardingTaskRepository struct {
	db *gorm.DB
}

var _ domainrepo.OffboardingTaskRepository = (*OffboardingTaskRepository)(nil)

// Method bernama objek mempertahankan kontrak kueri tanpa embedding struct.
func (r *OffboardingTaskRepository) InsertOffboardingTasks(ctx context.Context, v *entity.OffboardingTask) error {
	return r.Insert(ctx, v)
}
func (r *OffboardingTaskRepository) UpdateOffboardingTasks(ctx context.Context, v *entity.OffboardingTask) error {
	return r.Update(ctx, v)
}
func (r *OffboardingTaskRepository) DeleteOffboardingTasksByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByUUID(ctx context.Context, id string) (*entity.OffboardingTask, error) {
	return r.FindByID(ctx, id)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByUUIDForUpdate(ctx context.Context, id string) (*entity.OffboardingTask, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *OffboardingTaskRepository) GetAllOffboardingTasks(ctx context.Context, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OffboardingTaskRepository) CountOffboardingTasks(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OffboardingTaskRepository) SelectOffboardingTasksWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *OffboardingTaskRepository) CountOffboardingTasksWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByOffboardingPlanID(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByOffboardingPlanID(ctx, v, limit, offset)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByTaskCode(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByTaskCode(ctx, v, limit, offset)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByAssignedPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByAssignedPrincipalID(ctx, v, limit, offset)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByStatus(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByOwnerConfirmationReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByOwnerConfirmationReferenceID(ctx, v, limit, offset)
}
func (r *OffboardingTaskRepository) FindOffboardingTasksByWaivedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.FindByWaivedByPrincipalID(ctx, v, limit, offset)
}

// NewOffboardingTaskRepository memvalidasi koneksi dan menyiapkan repository.
func NewOffboardingTaskRepository(db *gorm.DB) (*OffboardingTaskRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OffboardingTaskRepository{db: db}, nil
}

func (r *OffboardingTaskRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.OffboardingTask{})
}

// Insert menyimpan tugas clearance dengan nilai field eksplisit.
func (r *OffboardingTaskRepository) Insert(ctx context.Context, value *entity.OffboardingTask) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *OffboardingTaskRepository) Create(ctx context.Context, value *entity.OffboardingTask) error {
	return r.Insert(ctx, value)
}
func (r *OffboardingTaskRepository) FindByID(ctx context.Context, key string) (*entity.OffboardingTask, error) {
	return r.find(ctx, key, false)
}
func (r *OffboardingTaskRepository) SelectByID(ctx context.Context, key string) (*entity.OffboardingTask, error) {
	return r.FindByID(ctx, key)
}
func (r *OffboardingTaskRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.OffboardingTask, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *OffboardingTaskRepository) find(ctx context.Context, key string, lock bool) (*entity.OffboardingTask, error) {
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
	value := new(entity.OffboardingTask)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *OffboardingTaskRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *OffboardingTaskRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OffboardingTaskRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter scope sebelum pembatasan halaman.
func (r *OffboardingTaskRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OffboardingTask, error) {
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
	rows := make([]entity.OffboardingTask, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *OffboardingTaskRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *OffboardingTaskRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key tanpa membuat record baru saat ID hilang.
func (r *OffboardingTaskRepository) Update(ctx context.Context, value *entity.OffboardingTask) error {
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
func (r *OffboardingTaskRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.OffboardingTask{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OffboardingTaskRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByOffboardingPlanID mencari kecocokan offboarding_plan_id dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByOffboardingPlanID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "offboarding_plan_id", value, limit, offset)
}

// FindByTaskCode mencari kecocokan task_code dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByTaskCode(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "task_code", value, limit, offset)
}

// FindByAssignedPrincipalID mencari kecocokan assigned_principal_id dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "assigned_principal_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByOwnerConfirmationReferenceID mencari kecocokan owner_confirmation_reference_id dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByOwnerConfirmationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "owner_confirmation_reference_id", value, limit, offset)
}

// FindByWaivedByPrincipalID mencari kecocokan waived_by_principal_id dengan paginasi dan urutan primary key.
func (r *OffboardingTaskRepository) FindByWaivedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingTask, error) {
	return r.SelectBy(ctx, "waived_by_principal_id", value, limit, offset)
}
