// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package talent

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// TalentReviewCycleRepository menyediakan kueri bertipe untuk hr_talent_review_cycles.
type TalentReviewCycleRepository struct {
	db *gorm.DB
}

var _ domainrepo.TalentReviewCycleRepository = (*TalentReviewCycleRepository)(nil)

// NewTalentReviewCycleRepository memvalidasi koneksi dan menyiapkan repository.
func NewTalentReviewCycleRepository(db *gorm.DB) (*TalentReviewCycleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TalentReviewCycleRepository{db: db}, nil
}

// Operasi bernama objek menyediakan kontrak siklus yang eksplisit.
func (r *TalentReviewCycleRepository) InsertTalentReviewCycles(ctx context.Context, value *entity.TalentReviewCycle) error {
	return r.Insert(ctx, value)
}
func (r *TalentReviewCycleRepository) UpdateTalentReviewCycles(ctx context.Context, value *entity.TalentReviewCycle) error {
	return r.Update(ctx, value)
}
func (r *TalentReviewCycleRepository) DeleteTalentReviewCyclesByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByUUID(ctx context.Context, id string) (*entity.TalentReviewCycle, error) {
	return r.FindByID(ctx, id)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByUUIDForUpdate(ctx context.Context, id string) (*entity.TalentReviewCycle, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *TalentReviewCycleRepository) GetAllTalentReviewCycles(ctx context.Context, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentReviewCycleRepository) CountTalentReviewCycles(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *TalentReviewCycleRepository) SelectTalentReviewCyclesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *TalentReviewCycleRepository) CountTalentReviewCyclesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByCycleCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByCycleCode(ctx, value, limit, offset)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByFrameworkID(ctx, value, limit, offset)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByLegalEntityReferenceID(ctx, value, limit, offset)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, value, limit, offset)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *TalentReviewCycleRepository) FindTalentReviewCyclesByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.FindByOwnerPrincipalID(ctx, value, limit, offset)
}

func (r *TalentReviewCycleRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.TalentReviewCycle{})
}

// Insert menyimpan snapshot siklus.
func (r *TalentReviewCycleRepository) Insert(ctx context.Context, value *entity.TalentReviewCycle) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *TalentReviewCycleRepository) Create(ctx context.Context, value *entity.TalentReviewCycle) error {
	return r.Insert(ctx, value)
}
func (r *TalentReviewCycleRepository) FindByID(ctx context.Context, key string) (*entity.TalentReviewCycle, error) {
	return r.find(ctx, key, false)
}
func (r *TalentReviewCycleRepository) SelectByID(ctx context.Context, key string) (*entity.TalentReviewCycle, error) {
	return r.FindByID(ctx, key)
}
func (r *TalentReviewCycleRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.TalentReviewCycle, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *TalentReviewCycleRepository) find(ctx context.Context, key string, lock bool) (*entity.TalentReviewCycle, error) {
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
	var value entity.TalentReviewCycle
	if err := db.Where("id = ?", key).First(&value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return &value, nil
}
func (r *TalentReviewCycleRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *TalentReviewCycleRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *TalentReviewCycleRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum urutan dan pagination.
func (r *TalentReviewCycleRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.TalentReviewCycle, error) {
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
	rows := make([]entity.TalentReviewCycle, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *TalentReviewCycleRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *TalentReviewCycleRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh kolom non-key, termasuk nol/NULL, tanpa upsert.
func (r *TalentReviewCycleRepository) Update(ctx context.Context, value *entity.TalentReviewCycle) error {
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
func (r *TalentReviewCycleRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.TalentReviewCycle{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *TalentReviewCycleRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCycleCode mencari kecocokan cycle_code dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByCycleCode(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "cycle_code", value, limit, offset)
}

// FindByFrameworkID mencari kecocokan framework_id dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByFrameworkID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "framework_id", value, limit, offset)
}

// FindByLegalEntityReferenceID mencari kecocokan legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "legal_entity_reference_id", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByOwnerPrincipalID mencari kecocokan owner_principal_id dengan paginasi dan urutan primary key.
func (r *TalentReviewCycleRepository) FindByOwnerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.TalentReviewCycle, error) {
	return r.SelectBy(ctx, "owner_principal_id", value, limit, offset)
}
