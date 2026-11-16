// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package onboarding

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// OnboardingTaskRepository menyediakan kueri bertipe untuk hr_onboarding_tasks.
type OnboardingTaskRepository struct {
	db *gorm.DB
}

var _ domainrepo.OnboardingTaskRepository = (*OnboardingTaskRepository)(nil)

// Method bernama objek menyediakan operasi eksplisit tanpa promosi method dari embedding.
func (r *OnboardingTaskRepository) InsertOnboardingTasks(ctx context.Context, v *entity.OnboardingTask) error {
	return r.Insert(ctx, v)
}
func (r *OnboardingTaskRepository) UpdateOnboardingTasks(ctx context.Context, v *entity.OnboardingTask) error {
	return r.Update(ctx, v)
}
func (r *OnboardingTaskRepository) DeleteOnboardingTasksByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByUUID(ctx context.Context, id string) (*entity.OnboardingTask, error) {
	return r.FindByID(ctx, id)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByUUIDForUpdate(ctx context.Context, id string) (*entity.OnboardingTask, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *OnboardingTaskRepository) GetAllOnboardingTasks(ctx context.Context, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OnboardingTaskRepository) CountOnboardingTasks(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OnboardingTaskRepository) SelectOnboardingTasksWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *OnboardingTaskRepository) CountOnboardingTasksWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByOnboardingPlanID(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.FindByOnboardingPlanID(ctx, v, limit, offset)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByTaskCode(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.FindByTaskCode(ctx, v, limit, offset)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByTaskOwnerType(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.FindByTaskOwnerType(ctx, v, limit, offset)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByAssignedPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.FindByAssignedPrincipalID(ctx, v, limit, offset)
}
func (r *OnboardingTaskRepository) FindOnboardingTasksByCompletionEvidenceReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.FindByCompletionEvidenceReferenceID(ctx, v, limit, offset)
}

// NewOnboardingTaskRepository memvalidasi koneksi dan menyiapkan repository.
func NewOnboardingTaskRepository(db *gorm.DB) (*OnboardingTaskRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OnboardingTaskRepository{db: db}, nil
}

func (r *OnboardingTaskRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.OnboardingTask{})
}

// Insert menyimpan tugas onboarding dan mengembalikan primary key dari database.
func (r *OnboardingTaskRepository) Insert(ctx context.Context, value *entity.OnboardingTask) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *OnboardingTaskRepository) Create(ctx context.Context, value *entity.OnboardingTask) error {
	return r.Insert(ctx, value)
}
func (r *OnboardingTaskRepository) FindByID(ctx context.Context, key string) (*entity.OnboardingTask, error) {
	return r.find(ctx, key, false)
}
func (r *OnboardingTaskRepository) SelectByID(ctx context.Context, key string) (*entity.OnboardingTask, error) {
	return r.FindByID(ctx, key)
}
func (r *OnboardingTaskRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.OnboardingTask, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *OnboardingTaskRepository) find(ctx context.Context, key string, lock bool) (*entity.OnboardingTask, error) {
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
	value := new(entity.OnboardingTask)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *OnboardingTaskRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *OnboardingTaskRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OnboardingTaskRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter tugas sebelum limit/offset tanpa mengabaikan nilai NULL.
func (r *OnboardingTaskRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OnboardingTask, error) {
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
	rows := make([]entity.OnboardingTask, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *OnboardingTaskRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *OnboardingTaskRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot lengkap tanpa mengubah primary key atau melakukan upsert.
func (r *OnboardingTaskRepository) Update(ctx context.Context, value *entity.OnboardingTask) error {
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
func (r *OnboardingTaskRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.OnboardingTask{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OnboardingTaskRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByOnboardingPlanID mencari kecocokan onboarding_plan_id dengan paginasi dan urutan primary key.
func (r *OnboardingTaskRepository) FindByOnboardingPlanID(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectBy(ctx, "onboarding_plan_id", value, limit, offset)
}

// FindByTaskCode mencari kecocokan task_code dengan paginasi dan urutan primary key.
func (r *OnboardingTaskRepository) FindByTaskCode(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectBy(ctx, "task_code", value, limit, offset)
}

// FindByTaskOwnerType mencari kecocokan task_owner_type dengan paginasi dan urutan primary key.
func (r *OnboardingTaskRepository) FindByTaskOwnerType(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectBy(ctx, "task_owner_type", value, limit, offset)
}

// FindByAssignedPrincipalID mencari kecocokan assigned_principal_id dengan paginasi dan urutan primary key.
func (r *OnboardingTaskRepository) FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectBy(ctx, "assigned_principal_id", value, limit, offset)
}

// FindByCompletionEvidenceReferenceID mencari kecocokan completion_evidence_reference_id dengan paginasi dan urutan primary key.
func (r *OnboardingTaskRepository) FindByCompletionEvidenceReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingTask, error) {
	return r.SelectBy(ctx, "completion_evidence_reference_id", value, limit, offset)
}
