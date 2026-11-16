// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package succession

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// SuccessionPlanRepository menyediakan kueri bertipe untuk hr_succession_plans.
type SuccessionPlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.SuccessionPlanRepository = (*SuccessionPlanRepository)(nil)

// NewSuccessionPlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewSuccessionPlanRepository(db *gorm.DB) (*SuccessionPlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &SuccessionPlanRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak rencana suksesi yang eksplisit.
func (r *SuccessionPlanRepository) InsertSuccessionPlans(ctx context.Context, value *entity.SuccessionPlan) error {
	return r.Insert(ctx, value)
}
func (r *SuccessionPlanRepository) UpdateSuccessionPlans(ctx context.Context, value *entity.SuccessionPlan) error {
	return r.Update(ctx, value)
}
func (r *SuccessionPlanRepository) DeleteSuccessionPlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *SuccessionPlanRepository) FindSuccessionPlansByUUID(ctx context.Context, id string) (*entity.SuccessionPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *SuccessionPlanRepository) FindSuccessionPlansByUUIDForUpdate(ctx context.Context, id string) (*entity.SuccessionPlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *SuccessionPlanRepository) GetAllSuccessionPlans(ctx context.Context, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *SuccessionPlanRepository) CountSuccessionPlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *SuccessionPlanRepository) SelectSuccessionPlansWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *SuccessionPlanRepository) CountSuccessionPlansWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *SuccessionPlanRepository) FindSuccessionPlansByPositionID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.FindByPositionID(ctx, value, limit, offset)
}
func (r *SuccessionPlanRepository) FindSuccessionPlansByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.FindByPlanCode(ctx, value, limit, offset)
}
func (r *SuccessionPlanRepository) FindSuccessionPlansByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}

func (r *SuccessionPlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.SuccessionPlan{})
}

// Insert menyimpan rencana suksesi sesuai constraint database.
func (r *SuccessionPlanRepository) Insert(ctx context.Context, value *entity.SuccessionPlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *SuccessionPlanRepository) Create(ctx context.Context, value *entity.SuccessionPlan) error {
	return r.Insert(ctx, value)
}
func (r *SuccessionPlanRepository) FindByID(ctx context.Context, key string) (*entity.SuccessionPlan, error) {
	return r.find(ctx, key, false)
}
func (r *SuccessionPlanRepository) SelectByID(ctx context.Context, key string) (*entity.SuccessionPlan, error) {
	return r.FindByID(ctx, key)
}
func (r *SuccessionPlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.SuccessionPlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *SuccessionPlanRepository) find(ctx context.Context, key string, lock bool) (*entity.SuccessionPlan, error) {
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
	var value entity.SuccessionPlan
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *SuccessionPlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *SuccessionPlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *SuccessionPlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *SuccessionPlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.SuccessionPlan, error) {
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
	rows := make([]entity.SuccessionPlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *SuccessionPlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *SuccessionPlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *SuccessionPlanRepository) Update(ctx context.Context, value *entity.SuccessionPlan) error {
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
func (r *SuccessionPlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.SuccessionPlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *SuccessionPlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPositionID mencari kecocokan position_id dengan paginasi dan urutan primary key.
func (r *SuccessionPlanRepository) FindByPositionID(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectBy(ctx, "position_id", value, limit, offset)
}

// FindByPlanCode mencari kecocokan plan_code dengan paginasi dan urutan primary key.
func (r *SuccessionPlanRepository) FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectBy(ctx, "plan_code", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *SuccessionPlanRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.SuccessionPlan, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
