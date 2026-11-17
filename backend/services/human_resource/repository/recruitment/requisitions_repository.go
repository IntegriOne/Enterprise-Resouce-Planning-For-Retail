// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package recruitment

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// RequisitionRepository menyediakan kueri bertipe untuk hr_requisitions.
type RequisitionRepository struct {
	db *gorm.DB
}

var _ domainrepo.RequisitionRepository = (*RequisitionRepository)(nil)

// Method bernama objek memperjelas tabel requisitions serta kolom pencariannya.
func (r *RequisitionRepository) InsertRequisitions(ctx context.Context, v *entity.Requisition) error {
	return r.Insert(ctx, v)
}
func (r *RequisitionRepository) UpdateRequisitions(ctx context.Context, v *entity.Requisition) error {
	return r.Update(ctx, v)
}
func (r *RequisitionRepository) DeleteRequisitionsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *RequisitionRepository) FindRequisitionsByUUID(ctx context.Context, id string) (*entity.Requisition, error) {
	return r.FindByID(ctx, id)
}
func (r *RequisitionRepository) FindRequisitionsByUUIDForUpdate(ctx context.Context, id string) (*entity.Requisition, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *RequisitionRepository) GetAllRequisitions(ctx context.Context, limit, offset int) ([]entity.Requisition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *RequisitionRepository) CountRequisitions(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *RequisitionRepository) SelectRequisitionsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *RequisitionRepository) CountRequisitionsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *RequisitionRepository) FindRequisitionsByRequisitionNumber(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByRequisitionNumber(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByWorkforcePlanLineID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByWorkforcePlanLineID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByPositionID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByPositionID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByJobID(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByJobID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByOrganizationUnitReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByOrganizationUnitReferenceID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByHiringManagerPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByHiringManagerPrincipalID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByRecruiterPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByRecruiterPrincipalID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByRequisitionType(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByRequisitionType(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByStatus(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByBudgetReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByBudgetReferenceID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByWorkLocationReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByWorkLocationReferenceID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByRejectionReasonCode(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByRejectionReasonCode(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByRejectedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByRejectedByPrincipalID(ctx, v, limit, offset)
}
func (r *RequisitionRepository) FindRequisitionsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.Requisition, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

// NewRequisitionRepository memvalidasi koneksi dan menyiapkan repository.
func NewRequisitionRepository(db *gorm.DB) (*RequisitionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &RequisitionRepository{db: db}, nil
}

func (r *RequisitionRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Requisition{})
}

// Insert menyimpan permintaan rekrutmen tanpa keputusan approval di repository.
func (r *RequisitionRepository) Insert(ctx context.Context, value *entity.Requisition) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *RequisitionRepository) Create(ctx context.Context, value *entity.Requisition) error {
	return r.Insert(ctx, value)
}
func (r *RequisitionRepository) FindByID(ctx context.Context, key string) (*entity.Requisition, error) {
	return r.find(ctx, key, false)
}
func (r *RequisitionRepository) SelectByID(ctx context.Context, key string) (*entity.Requisition, error) {
	return r.FindByID(ctx, key)
}
func (r *RequisitionRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Requisition, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *RequisitionRepository) find(ctx context.Context, key string, lock bool) (*entity.Requisition, error) {
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
	value := new(entity.Requisition)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *RequisitionRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *RequisitionRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Requisition, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *RequisitionRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter permintaan sebelum pagination.
func (r *RequisitionRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Requisition, error) {
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
	rows := make([]entity.Requisition, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *RequisitionRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *RequisitionRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot non-key tanpa upsert atau perubahan primary key.
func (r *RequisitionRepository) Update(ctx context.Context, value *entity.Requisition) error {
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
func (r *RequisitionRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Requisition{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *RequisitionRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRequisitionNumber mencari kecocokan requisition_number dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByRequisitionNumber(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "requisition_number", value, limit, offset)
}

// FindByWorkforcePlanLineID mencari kecocokan workforce_plan_line_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByWorkforcePlanLineID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "workforce_plan_line_id", value, limit, offset)
}

// FindByPositionID mencari kecocokan position_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByPositionID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "position_id", value, limit, offset)
}

// FindByJobID mencari kecocokan job_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByJobID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "job_id", value, limit, offset)
}

// FindByOrganizationUnitReferenceID mencari kecocokan organization_unit_reference_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByOrganizationUnitReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "organization_unit_reference_id", value, limit, offset)
}

// FindByHiringManagerPrincipalID mencari kecocokan hiring_manager_principal_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByHiringManagerPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "hiring_manager_principal_id", value, limit, offset)
}

// FindByRecruiterPrincipalID mencari kecocokan recruiter_principal_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByRecruiterPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "recruiter_principal_id", value, limit, offset)
}

// FindByRequisitionType mencari kecocokan requisition_type dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByRequisitionType(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "requisition_type", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByBudgetReferenceID mencari kecocokan budget_reference_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByBudgetReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "budget_reference_id", value, limit, offset)
}

// FindByWorkLocationReferenceID mencari kecocokan work_location_reference_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByWorkLocationReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "work_location_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByRejectionReasonCode mencari alasan penolakan dengan paginasi.
func (r *RequisitionRepository) FindByRejectionReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "rejection_reason_code", value, limit, offset)
}

// FindByRejectedByPrincipalID mencari pengambil keputusan dengan paginasi.
func (r *RequisitionRepository) FindByRejectedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "rejected_by_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *RequisitionRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.Requisition, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
