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

// PositionRepository menyediakan kueri bertipe untuk hr_positions.
type PositionRepository struct {
	db *gorm.DB
}

var _ domainrepo.PositionRepository = (*PositionRepository)(nil)

// Method bernama objek memperjelas tabel positions serta kolom pencarian.
func (r *PositionRepository) InsertPositions(ctx context.Context, v *entity.Position) error {
	return r.Insert(ctx, v)
}
func (r *PositionRepository) UpdatePositions(ctx context.Context, v *entity.Position) error {
	return r.Update(ctx, v)
}
func (r *PositionRepository) DeletePositionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PositionRepository) FindPositionsByUUID(ctx context.Context, id string) (*entity.Position, error) {
	return r.FindByID(ctx, id)
}
func (r *PositionRepository) FindPositionsByUUIDForUpdate(ctx context.Context, id string) (*entity.Position, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PositionRepository) GetAllPositions(ctx context.Context, limit, offset int) ([]entity.Position, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PositionRepository) CountPositions(ctx context.Context) (int64, error) { return r.Count(ctx) }
func (r *PositionRepository) SelectPositionsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Position, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *PositionRepository) CountPositionsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *PositionRepository) FindPositionsByPositionCode(ctx context.Context, v string, limit, offset int) ([]entity.Position, error) {
	return r.FindByPositionCode(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByJobID(ctx context.Context, v string, limit, offset int) ([]entity.Position, error) {
	return r.FindByJobID(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByOrganizationUnitReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Position, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Position, error) {
	return r.FindByLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByReportsToPositionID(ctx context.Context, v *string, limit, offset int) ([]entity.Position, error) {
	return r.FindByReportsToPositionID(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByPositionStatus(ctx context.Context, v string, limit, offset int) ([]entity.Position, error) {
	return r.FindByPositionStatus(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByWorkLocationReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Position, error) {
	return r.FindByWorkLocationReferenceID(ctx, v, limit, offset)
}
func (r *PositionRepository) FindPositionsByCostCenterReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Position, error) {
	return r.FindByCostCenterReferenceID(ctx, v, limit, offset)
}

// NewPositionRepository memvalidasi koneksi dan menyiapkan repository.
func NewPositionRepository(db *gorm.DB) (*PositionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PositionRepository{db: db}, nil
}

func (r *PositionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Position{})
}

// Insert menyimpan posisi dengan referensi master yang sudah divalidasi service.
func (r *PositionRepository) Insert(ctx context.Context, value *entity.Position) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *PositionRepository) Create(ctx context.Context, value *entity.Position) error {
	return r.Insert(ctx, value)
}
func (r *PositionRepository) FindByID(ctx context.Context, key string) (*entity.Position, error) {
	return r.find(ctx, key, false)
}
func (r *PositionRepository) SelectByID(ctx context.Context, key string) (*entity.Position, error) {
	return r.FindByID(ctx, key)
}
func (r *PositionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Position, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *PositionRepository) find(ctx context.Context, key string, lock bool) (*entity.Position, error) {
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
	value := new(entity.Position)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *PositionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Position, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *PositionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Position, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PositionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Position, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter posisi sebelum limit dan offset.
func (r *PositionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Position, error) {
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
	rows := make([]entity.Position, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *PositionRepository) Count(ctx context.Context) (int64, error) { return r.CountWhere(ctx, nil) }
func (r *PositionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key dan tidak membuat posisi baru bila ID hilang.
func (r *PositionRepository) Update(ctx context.Context, value *entity.Position) error {
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
func (r *PositionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Position{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *PositionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPositionCode mencari kecocokan position_code dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByPositionCode(ctx context.Context, value string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "position_code", value, limit, offset)
}

// FindByJobID mencari kecocokan job_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "job_id", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByReportsToPositionID mencari kecocokan reports_to_position_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByReportsToPositionID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "reports_to_position_id", value, limit, offset)
}

// FindByPositionStatus mencari kecocokan position_status dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByPositionStatus(ctx context.Context, value string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "position_status", value, limit, offset)
}

// FindByWorkLocationReferenceID mencari kecocokan work_location_reference_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "work_location_reference_id", value, limit, offset)
}

// FindByCostCenterReferenceID mencari kecocokan cost_center_reference_id dengan paginasi dan urutan primary key.
func (r *PositionRepository) FindByCostCenterReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Position, error) {
	return r.SelectBy(ctx, "cost_center_reference_id", value, limit, offset)
}
