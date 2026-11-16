// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package career

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// DevelopmentPlanRepository menyediakan kueri bertipe untuk hr_development_plans.
type DevelopmentPlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.DevelopmentPlanRepository = (*DevelopmentPlanRepository)(nil)

// NewDevelopmentPlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewDevelopmentPlanRepository(db *gorm.DB) (*DevelopmentPlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &DevelopmentPlanRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak rencana pengembangan yang eksplisit.
func (r *DevelopmentPlanRepository) InsertDevelopmentPlans(ctx context.Context, value *entity.DevelopmentPlan) error {
	return r.Insert(ctx, value)
}
func (r *DevelopmentPlanRepository) UpdateDevelopmentPlans(ctx context.Context, value *entity.DevelopmentPlan) error {
	return r.Update(ctx, value)
}
func (r *DevelopmentPlanRepository) DeleteDevelopmentPlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByUUID(ctx context.Context, id string) (*entity.DevelopmentPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByUUIDForUpdate(ctx context.Context, id string) (*entity.DevelopmentPlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *DevelopmentPlanRepository) GetAllDevelopmentPlans(ctx context.Context, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *DevelopmentPlanRepository) CountDevelopmentPlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *DevelopmentPlanRepository) SelectDevelopmentPlansWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *DevelopmentPlanRepository) CountDevelopmentPlansWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByPlanCode(ctx, value, limit, offset)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByTargetJobID(ctx, value, limit, offset)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByStatus(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByMentorWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByMentorWorkerID(ctx, value, limit, offset)
}
func (r *DevelopmentPlanRepository) FindDevelopmentPlansByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.FindByApprovedByPrincipalID(ctx, value, limit, offset)
}

func (r *DevelopmentPlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.DevelopmentPlan{})
}

// Insert menyimpan rencana sesuai constraint database.
func (r *DevelopmentPlanRepository) Insert(ctx context.Context, value *entity.DevelopmentPlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *DevelopmentPlanRepository) Create(ctx context.Context, value *entity.DevelopmentPlan) error {
	return r.Insert(ctx, value)
}
func (r *DevelopmentPlanRepository) FindByID(ctx context.Context, key string) (*entity.DevelopmentPlan, error) {
	return r.find(ctx, key, false)
}
func (r *DevelopmentPlanRepository) SelectByID(ctx context.Context, key string) (*entity.DevelopmentPlan, error) {
	return r.FindByID(ctx, key)
}
func (r *DevelopmentPlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.DevelopmentPlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *DevelopmentPlanRepository) find(ctx context.Context, key string, lock bool) (*entity.DevelopmentPlan, error) {
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
	var value entity.DevelopmentPlan
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *DevelopmentPlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *DevelopmentPlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *DevelopmentPlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *DevelopmentPlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.DevelopmentPlan, error) {
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
	rows := make([]entity.DevelopmentPlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *DevelopmentPlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *DevelopmentPlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *DevelopmentPlanRepository) Update(ctx context.Context, value *entity.DevelopmentPlan) error {
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
func (r *DevelopmentPlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.DevelopmentPlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *DevelopmentPlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPlanCode mencari kecocokan plan_code dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "plan_code", value, limit, offset)
}

// FindByTargetJobID mencari kecocokan target_job_id dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByTargetJobID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "target_job_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByMentorWorkerID mencari kecocokan mentor_worker_id dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByMentorWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "mentor_worker_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *DevelopmentPlanRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.DevelopmentPlan, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
