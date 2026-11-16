// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package analytics

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// HumanCapitalMetricDefinitionRepository menyediakan kueri bertipe untuk hr_human_capital_metric_definitions.
type HumanCapitalMetricDefinitionRepository struct {
	db *gorm.DB
}

var _ domainrepo.HumanCapitalMetricDefinitionRepository = (*HumanCapitalMetricDefinitionRepository)(nil)

// NewHumanCapitalMetricDefinitionRepository memvalidasi koneksi dan menyiapkan repository.
func NewHumanCapitalMetricDefinitionRepository(db *gorm.DB) (*HumanCapitalMetricDefinitionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &HumanCapitalMetricDefinitionRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak definisi metrik yang eksplisit.
func (r *HumanCapitalMetricDefinitionRepository) InsertHumanCapitalMetricDefinitions(ctx context.Context, value *entity.HumanCapitalMetricDefinition) error {
	return r.Insert(ctx, value)
}
func (r *HumanCapitalMetricDefinitionRepository) UpdateHumanCapitalMetricDefinitions(ctx context.Context, value *entity.HumanCapitalMetricDefinition) error {
	return r.Update(ctx, value)
}
func (r *HumanCapitalMetricDefinitionRepository) DeleteHumanCapitalMetricDefinitionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *HumanCapitalMetricDefinitionRepository) FindHumanCapitalMetricDefinitionsByUUID(ctx context.Context, id string) (*entity.HumanCapitalMetricDefinition, error) {
	return r.FindByID(ctx, id)
}
func (r *HumanCapitalMetricDefinitionRepository) FindHumanCapitalMetricDefinitionsByUUIDForUpdate(ctx context.Context, id string) (*entity.HumanCapitalMetricDefinition, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *HumanCapitalMetricDefinitionRepository) GetAllHumanCapitalMetricDefinitions(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *HumanCapitalMetricDefinitionRepository) CountHumanCapitalMetricDefinitions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *HumanCapitalMetricDefinitionRepository) SelectHumanCapitalMetricDefinitionsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *HumanCapitalMetricDefinitionRepository) CountHumanCapitalMetricDefinitionsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *HumanCapitalMetricDefinitionRepository) FindHumanCapitalMetricDefinitionsByMetricCode(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.FindByMetricCode(ctx, value, limit, offset)
}
func (r *HumanCapitalMetricDefinitionRepository) FindHumanCapitalMetricDefinitionsByUnitCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.FindByUnitCode(ctx, value, limit, offset)
}

func (r *HumanCapitalMetricDefinitionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.HumanCapitalMetricDefinition{})
}

// Insert menyimpan definisi metrik sesuai constraint database.
func (r *HumanCapitalMetricDefinitionRepository) Insert(ctx context.Context, value *entity.HumanCapitalMetricDefinition) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *HumanCapitalMetricDefinitionRepository) Create(ctx context.Context, value *entity.HumanCapitalMetricDefinition) error {
	return r.Insert(ctx, value)
}
func (r *HumanCapitalMetricDefinitionRepository) FindByID(ctx context.Context, key string) (*entity.HumanCapitalMetricDefinition, error) {
	return r.find(ctx, key, false)
}
func (r *HumanCapitalMetricDefinitionRepository) SelectByID(ctx context.Context, key string) (*entity.HumanCapitalMetricDefinition, error) {
	return r.FindByID(ctx, key)
}
func (r *HumanCapitalMetricDefinitionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.HumanCapitalMetricDefinition, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *HumanCapitalMetricDefinitionRepository) find(ctx context.Context, key string, lock bool) (*entity.HumanCapitalMetricDefinition, error) {
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
	var value entity.HumanCapitalMetricDefinition
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *HumanCapitalMetricDefinitionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *HumanCapitalMetricDefinitionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *HumanCapitalMetricDefinitionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *HumanCapitalMetricDefinitionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
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
	rows := make([]entity.HumanCapitalMetricDefinition, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *HumanCapitalMetricDefinitionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *HumanCapitalMetricDefinitionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *HumanCapitalMetricDefinitionRepository) Update(ctx context.Context, value *entity.HumanCapitalMetricDefinition) error {
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
func (r *HumanCapitalMetricDefinitionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.HumanCapitalMetricDefinition{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *HumanCapitalMetricDefinitionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByMetricCode mencari kecocokan metric_code dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricDefinitionRepository) FindByMetricCode(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.SelectBy(ctx, "metric_code", value, limit, offset)
}

// FindByUnitCode mencari kecocokan unit_code dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricDefinitionRepository) FindByUnitCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricDefinition, error) {
	return r.SelectBy(ctx, "unit_code", value, limit, offset)
}
