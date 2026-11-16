// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package offboarding

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// OffboardingPlanRepository menyediakan kueri bertipe untuk hr_offboarding_plans.
type OffboardingPlanRepository struct {
	db *gorm.DB
}

var _ domainrepo.OffboardingPlanRepository = (*OffboardingPlanRepository)(nil)

// Method bernama objek memperjelas operasi tabel dan kolom kunci untuk review.
func (r *OffboardingPlanRepository) InsertOffboardingPlans(ctx context.Context, v *entity.OffboardingPlan) error {
	return r.Insert(ctx, v)
}
func (r *OffboardingPlanRepository) UpdateOffboardingPlans(ctx context.Context, v *entity.OffboardingPlan) error {
	return r.Update(ctx, v)
}
func (r *OffboardingPlanRepository) DeleteOffboardingPlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByUUID(ctx context.Context, id string) (*entity.OffboardingPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByUUIDForUpdate(ctx context.Context, id string) (*entity.OffboardingPlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *OffboardingPlanRepository) GetAllOffboardingPlans(ctx context.Context, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OffboardingPlanRepository) CountOffboardingPlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OffboardingPlanRepository) SelectOffboardingPlansWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *OffboardingPlanRepository) CountOffboardingPlansWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByPlanNumber(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByPlanNumber(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByReasonCode(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByReasonCode(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByStatus(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByExitInterviewDocumentVersionReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByExitInterviewDocumentVersionReferenceID(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByFinalPayrollRunID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByFinalPayrollRunID(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByFinanceClearanceReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByFinanceClearanceReferenceID(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
func (r *OffboardingPlanRepository) FindOffboardingPlansByCreatedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}

// NewOffboardingPlanRepository memvalidasi koneksi dan menyiapkan repository.
func NewOffboardingPlanRepository(db *gorm.DB) (*OffboardingPlanRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OffboardingPlanRepository{db: db}, nil
}

func (r *OffboardingPlanRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.OffboardingPlan{})
}

// Insert menyimpan rencana offboarding dalam transaksi pemanggil bila tersedia.
func (r *OffboardingPlanRepository) Insert(ctx context.Context, value *entity.OffboardingPlan) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *OffboardingPlanRepository) Create(ctx context.Context, value *entity.OffboardingPlan) error {
	return r.Insert(ctx, value)
}
func (r *OffboardingPlanRepository) FindByID(ctx context.Context, key string) (*entity.OffboardingPlan, error) {
	return r.find(ctx, key, false)
}
func (r *OffboardingPlanRepository) SelectByID(ctx context.Context, key string) (*entity.OffboardingPlan, error) {
	return r.FindByID(ctx, key)
}
func (r *OffboardingPlanRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.OffboardingPlan, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *OffboardingPlanRepository) find(ctx context.Context, key string, lock bool) (*entity.OffboardingPlan, error) {
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
	value := new(entity.OffboardingPlan)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *OffboardingPlanRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *OffboardingPlanRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OffboardingPlanRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter rencana sebelum pagination dengan urutan primary key stabil.
func (r *OffboardingPlanRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OffboardingPlan, error) {
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
	rows := make([]entity.OffboardingPlan, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *OffboardingPlanRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *OffboardingPlanRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mempertahankan nilai nol/NULL dan menolak perubahan primary key maupun upsert.
func (r *OffboardingPlanRepository) Update(ctx context.Context, value *entity.OffboardingPlan) error {
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
func (r *OffboardingPlanRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.OffboardingPlan{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *OffboardingPlanRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPlanNumber mencari kecocokan plan_number dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByPlanNumber(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "plan_number", value, limit, offset)
}

// FindByReasonCode mencari kecocokan reason_code dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "reason_code", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByExitInterviewDocumentVersionReferenceID mencari kecocokan exit_interview_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByExitInterviewDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "exit_interview_document_version_reference_id", value, limit, offset)
}

// FindByFinalPayrollRunID mencari kecocokan final_payroll_run_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByFinalPayrollRunID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "final_payroll_run_id", value, limit, offset)
}

// FindByFinanceClearanceReferenceID mencari kecocokan finance_clearance_reference_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByFinanceClearanceReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "finance_clearance_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *OffboardingPlanRepository) FindByCreatedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.OffboardingPlan, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}
