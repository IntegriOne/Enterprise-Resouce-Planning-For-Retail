// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package workforceplanning

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkforcePlanRepository menyediakan kueri bertipe untuk hr_workforce_plans.
type WorkforcePlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkforcePlanRepository = (*WorkforcePlanRepository)(nil)

// Method bernama objek memperjelas kontrak rencana tenaga kerja.
func (r *WorkforcePlanRepository) InsertWorkforcePlans(ctx context.Context, v *entity.WorkforcePlan) error {
	return r.Insert(ctx, v)
}
func (r *WorkforcePlanRepository) UpdateWorkforcePlans(ctx context.Context, v *entity.WorkforcePlan) error {
	return r.Update(ctx, v)
}
func (r *WorkforcePlanRepository) DeleteWorkforcePlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByUUID(ctx context.Context, id string) (*entity.WorkforcePlan, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkforcePlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkforcePlanRepository) GetAllWorkforcePlans(ctx context.Context, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkforcePlanRepository) CountWorkforcePlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkforcePlanRepository) SelectWorkforcePlansWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkforcePlanRepository) CountWorkforcePlansWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByPlanCode(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByPlanCode(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByScenarioCode(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByScenarioCode(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByStatus(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByFinanceBudgetReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByFinanceBudgetReferenceID(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByCreatedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}
func (r *WorkforcePlanRepository) FindWorkforcePlansByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}

// NewWorkforcePlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkforcePlanRepository(db *gorm.DB) (*WorkforcePlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkforcePlanRepository{db: db}, nil
}

func (r *WorkforcePlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkforcePlan{})
}

// Insert menyimpan rencana tenaga kerja tanpa keputusan approval di repository.
func (r *WorkforcePlanRepository) Insert(ctx context.Context, value *entity.WorkforcePlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkforcePlanRepository) Create(ctx context.Context, value *entity.WorkforcePlan) error {
	return r.Insert(ctx, value)
}
func (r *WorkforcePlanRepository) FindByID(ctx context.Context, key string) (*entity.WorkforcePlan, error) {
	return r.find(ctx, key, false)
}
func (r *WorkforcePlanRepository) SelectByID(ctx context.Context, key string) (*entity.WorkforcePlan, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkforcePlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkforcePlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkforcePlanRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkforcePlan, error) {
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
	value := new(entity.WorkforcePlan)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkforcePlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkforcePlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkforcePlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter rencana sebelum pagination dan menjaga urutan primary key.
func (r *WorkforcePlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkforcePlan, error) {
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
	rows := make([]entity.WorkforcePlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkforcePlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkforcePlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan seluruh snapshot non-key tanpa upsert dan tanpa melewati constraint SQL.
func (r *WorkforcePlanRepository) Update(ctx context.Context, value *entity.WorkforcePlan) error {
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
func (r *WorkforcePlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkforcePlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkforcePlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPlanCode mencari kecocokan plan_code dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "plan_code", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByScenarioCode mencari kecocokan scenario_code dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByScenarioCode(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "scenario_code", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByFinanceBudgetReferenceID mencari kecocokan finance_budget_reference_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByFinanceBudgetReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "finance_budget_reference_id", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkforcePlanRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkforcePlan, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
