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

// HumanCapitalMetricSnapshotRepository menyediakan kueri bertipe untuk hr_human_capital_metric_snapshots.
type HumanCapitalMetricSnapshotRepository struct {
	db *gorm.DB
}

var _ domainrepo.HumanCapitalMetricSnapshotRepository = (*HumanCapitalMetricSnapshotRepository)(nil)

// NewHumanCapitalMetricSnapshotRepository memvalidasi koneksi dan menyiapkan repository.
func NewHumanCapitalMetricSnapshotRepository(db *gorm.DB) (*HumanCapitalMetricSnapshotRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &HumanCapitalMetricSnapshotRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak snapshot metrik yang eksplisit.
func (r *HumanCapitalMetricSnapshotRepository) InsertHumanCapitalMetricSnapshots(ctx context.Context, value *entity.HumanCapitalMetricSnapshot) error {
	return r.Insert(ctx, value)
}
func (r *HumanCapitalMetricSnapshotRepository) UpdateHumanCapitalMetricSnapshots(ctx context.Context, value *entity.HumanCapitalMetricSnapshot) error {
	return r.Update(ctx, value)
}
func (r *HumanCapitalMetricSnapshotRepository) DeleteHumanCapitalMetricSnapshotsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByUUID(ctx context.Context, id string) (*entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByID(ctx, id)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByUUIDForUpdate(ctx context.Context, id string) (*entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *HumanCapitalMetricSnapshotRepository) GetAllHumanCapitalMetricSnapshots(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) CountHumanCapitalMetricSnapshots(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *HumanCapitalMetricSnapshotRepository) SelectHumanCapitalMetricSnapshotsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) CountHumanCapitalMetricSnapshotsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByMetricDefinitionID(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByMetricDefinitionID(ctx, value, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, value, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByJurisdictionCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) FindHumanCapitalMetricSnapshotsByDataQualityStatus(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByDataQualityStatus(ctx, value, limit, offset)
}

func (r *HumanCapitalMetricSnapshotRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.HumanCapitalMetricSnapshot{})
}

// Insert menyimpan snapshot metrik sesuai constraint database.
func (r *HumanCapitalMetricSnapshotRepository) Insert(ctx context.Context, value *entity.HumanCapitalMetricSnapshot) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *HumanCapitalMetricSnapshotRepository) Create(ctx context.Context, value *entity.HumanCapitalMetricSnapshot) error {
	return r.Insert(ctx, value)
}
func (r *HumanCapitalMetricSnapshotRepository) FindByID(ctx context.Context, key string) (*entity.HumanCapitalMetricSnapshot, error) {
	return r.find(ctx, key, false)
}
func (r *HumanCapitalMetricSnapshotRepository) SelectByID(ctx context.Context, key string) (*entity.HumanCapitalMetricSnapshot, error) {
	return r.FindByID(ctx, key)
}
func (r *HumanCapitalMetricSnapshotRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.HumanCapitalMetricSnapshot, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *HumanCapitalMetricSnapshotRepository) find(ctx context.Context, key string, lock bool) (*entity.HumanCapitalMetricSnapshot, error) {
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
	var value entity.HumanCapitalMetricSnapshot
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *HumanCapitalMetricSnapshotRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *HumanCapitalMetricSnapshotRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter sebelum membaca halaman berurutan.
func (r *HumanCapitalMetricSnapshotRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
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
	rows := make([]entity.HumanCapitalMetricSnapshot, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *HumanCapitalMetricSnapshotRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *HumanCapitalMetricSnapshotRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *HumanCapitalMetricSnapshotRepository) Update(ctx context.Context, value *entity.HumanCapitalMetricSnapshot) error {
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
func (r *HumanCapitalMetricSnapshotRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.HumanCapitalMetricSnapshot{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *HumanCapitalMetricSnapshotRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByMetricDefinitionID mencari kecocokan metric_definition_id dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricSnapshotRepository) FindByMetricDefinitionID(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectBy(ctx, "metric_definition_id", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricSnapshotRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricSnapshotRepository) FindByJurisdictionCode(ctx context.Context, value *string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByDataQualityStatus mencari kecocokan data_quality_status dengan paginasi dan urutan primary key.
func (r *HumanCapitalMetricSnapshotRepository) FindByDataQualityStatus(ctx context.Context, value string, limit, offset int) ([]entity.HumanCapitalMetricSnapshot, error) {
	return r.SelectBy(ctx, "data_quality_status", value, limit, offset)
}
