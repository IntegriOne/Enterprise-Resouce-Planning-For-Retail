// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package benefits

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// BenefitPlanRepository menyediakan kueri bertipe untuk hr_benefit_plans.
type BenefitPlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.BenefitPlanRepository = (*BenefitPlanRepository)(nil)

// NewBenefitPlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewBenefitPlanRepository(db *gorm.DB) (*BenefitPlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &BenefitPlanRepository{db: db}, nil
}

func (r *BenefitPlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.BenefitPlan{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *BenefitPlanRepository) Insert(ctx context.Context, value *entity.BenefitPlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias INSERT.
func (r *BenefitPlanRepository) Create(ctx context.Context, value *entity.BenefitPlan) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *BenefitPlanRepository) FindByID(ctx context.Context, key string) (*entity.BenefitPlan, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *BenefitPlanRepository) SelectByID(ctx context.Context, key string) (*entity.BenefitPlan, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *BenefitPlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.BenefitPlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *BenefitPlanRepository) find(ctx context.Context, key string, lock bool) (*entity.BenefitPlan, error) {
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
	value := new(entity.BenefitPlan)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *BenefitPlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *BenefitPlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *BenefitPlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *BenefitPlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.BenefitPlan, error) {
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
	rows := make([]entity.BenefitPlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *BenefitPlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *BenefitPlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot lengkap non-key termasuk nilai nol dan NULL.
func (r *BenefitPlanRepository) Update(ctx context.Context, value *entity.BenefitPlan) error {
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

// DeleteByID menghapus satu record sesuai constraint database.
func (r *BenefitPlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.BenefitPlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *BenefitPlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPlanCode mencari kecocokan plan_code dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "plan_code", value, limit, offset)
}

// FindByBenefitType mencari kecocokan benefit_type dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByBenefitType(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "benefit_type", value, limit, offset)
}

// FindByProviderPartyReferenceID mencari kecocokan provider_party_reference_id dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "provider_party_reference_id", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByEmployeeContributionRuleReferenceID mencari kecocokan employee_contribution_rule_reference_id dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByEmployeeContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "employee_contribution_rule_reference_id", value, limit, offset)
}

// FindByEmployerContributionRuleReferenceID mencari kecocokan employer_contribution_rule_reference_id dengan paginasi dan urutan primary key.
func (r *BenefitPlanRepository) FindByEmployerContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectBy(ctx, "employer_contribution_rule_reference_id", value, limit, offset)
}
