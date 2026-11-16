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

// OnboardingPlanRepository menyediakan kueri bertipe untuk hr_onboarding_plans.
type OnboardingPlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.OnboardingPlanRepository = (*OnboardingPlanRepository)(nil)

// Method bernama objek menjadi API review eksplisit; alias lama menjaga wiring service.
func (r *OnboardingPlanRepository) InsertOnboardingPlans(ctx context.Context, v *entity.OnboardingPlan) error {
	return r.Insert(ctx, v)
}
func (r *OnboardingPlanRepository) UpdateOnboardingPlans(ctx context.Context, v *entity.OnboardingPlan) error {
	return r.Update(ctx, v)
}
func (r *OnboardingPlanRepository) DeleteOnboardingPlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByUUID(ctx context.Context, id string) (*entity.OnboardingPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByUUIDForUpdate(ctx context.Context, id string) (*entity.OnboardingPlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *OnboardingPlanRepository) GetAllOnboardingPlans(ctx context.Context, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OnboardingPlanRepository) CountOnboardingPlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OnboardingPlanRepository) SelectOnboardingPlansWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *OnboardingPlanRepository) CountOnboardingPlansWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByPlanCode(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.FindByPlanCode(ctx, v, limit, offset)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByStatus(ctx context.Context, v string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByTemplateReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.FindByTemplateReferenceID(ctx, v, limit, offset)
}
func (r *OnboardingPlanRepository) FindOnboardingPlansByCreatedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}

// NewOnboardingPlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewOnboardingPlanRepository(db *gorm.DB) (*OnboardingPlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OnboardingPlanRepository{db: db}, nil
}

func (r *OnboardingPlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.OnboardingPlan{})
}

// Insert menyimpan rencana onboarding beserta primary key hasil RETURNING.
func (r *OnboardingPlanRepository) Insert(ctx context.Context, value *entity.OnboardingPlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

func (r *OnboardingPlanRepository) Create(ctx context.Context, value *entity.OnboardingPlan) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu rencana onboarding berdasarkan primary key.
func (r *OnboardingPlanRepository) FindByID(ctx context.Context, key string) (*entity.OnboardingPlan, error) {
	return r.find(ctx, key, false)
}
func (r *OnboardingPlanRepository) SelectByID(ctx context.Context, key string) (*entity.OnboardingPlan, error) {
	return r.FindByID(ctx, key)
}
func (r *OnboardingPlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.OnboardingPlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *OnboardingPlanRepository) find(ctx context.Context, key string, lock bool) (*entity.OnboardingPlan, error) {
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
	value := new(entity.OnboardingPlan)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

func (r *OnboardingPlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *OnboardingPlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OnboardingPlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination dengan urutan primary key stabil.
func (r *OnboardingPlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OnboardingPlan, error) {
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
	rows := make([]entity.OnboardingPlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *OnboardingPlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *OnboardingPlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh snapshot non-key termasuk nilai nol dan NULL, tanpa upsert.
func (r *OnboardingPlanRepository) Update(ctx context.Context, value *entity.OnboardingPlan) error {
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
func (r *OnboardingPlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.OnboardingPlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OnboardingPlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *OnboardingPlanRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPlanCode mencari kecocokan plan_code dengan paginasi dan urutan primary key.
func (r *OnboardingPlanRepository) FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectBy(ctx, "plan_code", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *OnboardingPlanRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByTemplateReferenceID mencari kecocokan template_reference_id dengan paginasi dan urutan primary key.
func (r *OnboardingPlanRepository) FindByTemplateReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectBy(ctx, "template_reference_id", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *OnboardingPlanRepository) FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OnboardingPlan, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}
