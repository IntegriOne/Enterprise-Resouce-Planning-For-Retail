package employeerelations

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// EmployeeRelationsCaseRepository menyediakan kueri bertipe untuk hr_employee_relations_cases.
type EmployeeRelationsCaseRepository struct {
	db *gorm.DB
}

var _ domainrepo.EmployeeRelationsCaseRepository = (*EmployeeRelationsCaseRepository)(nil)

// NewEmployeeRelationsCaseRepository memvalidasi koneksi dan menyiapkan repository.
func NewEmployeeRelationsCaseRepository(db *gorm.DB) (*EmployeeRelationsCaseRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EmployeeRelationsCaseRepository{db: db}, nil
}

func (r *EmployeeRelationsCaseRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EmployeeRelationsCase{})
}

// Insert menyimpan satu kasus hubungan industrial.
func (r *EmployeeRelationsCaseRepository) Insert(ctx context.Context, value *entity.EmployeeRelationsCase) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create mempertahankan alias INSERT.
func (r *EmployeeRelationsCaseRepository) Create(ctx context.Context, value *entity.EmployeeRelationsCase) error {
	return r.Insert(ctx, value)
}

// FindByID membaca kasus berdasarkan ID.
func (r *EmployeeRelationsCaseRepository) FindByID(ctx context.Context, key string) (*entity.EmployeeRelationsCase, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *EmployeeRelationsCaseRepository) SelectByID(ctx context.Context, key string) (*entity.EmployeeRelationsCase, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci kasus dalam transaksi aktif.
func (r *EmployeeRelationsCaseRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.EmployeeRelationsCase, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *EmployeeRelationsCaseRepository) find(ctx context.Context, key string, lock bool) (*entity.EmployeeRelationsCase, error) {
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
	value := &entity.EmployeeRelationsCase{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman kasus.
func (r *EmployeeRelationsCaseRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *EmployeeRelationsCaseRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memakai kolom model yang diverifikasi untuk pencarian berhalaman.
func (r *EmployeeRelationsCaseRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum limit/offset.
func (r *EmployeeRelationsCaseRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
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
	rows := make([]entity.EmployeeRelationsCase, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh kasus.
func (r *EmployeeRelationsCaseRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung hanya kasus yang memenuhi filter.
func (r *EmployeeRelationsCaseRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot kasus non-key lengkap tanpa upsert.
func (r *EmployeeRelationsCaseRepository) Update(ctx context.Context, value *entity.EmployeeRelationsCase) error {
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

// DeleteByID menghapus satu kasus dengan tetap mengikuti constraint database.
func (r *EmployeeRelationsCaseRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EmployeeRelationsCase{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *EmployeeRelationsCaseRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByCaseNumber mencari kecocokan case_number dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByCaseNumber(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "case_number", value, limit, offset)
}

// FindByCaseType mencari kecocokan case_type dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByCaseType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "case_type", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByStatus(ctx context.Context, value entity.CaseStatus, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindBySubjectWorkerID mencari kecocokan subject_worker_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindBySubjectWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "subject_worker_id", value, limit, offset)
}

// FindByReportingPartyReferenceID mencari kecocokan reporting_party_reference_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByReportingPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "reporting_party_reference_id", value, limit, offset)
}

// FindByAssignedInvestigatorPrincipalID mencari kecocokan assigned_investigator_principal_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByAssignedInvestigatorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "assigned_investigator_principal_id", value, limit, offset)
}

// FindByOutcomeCode mencari kecocokan outcome_code dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByOutcomeCode(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "outcome_code", value, limit, offset)
}

// FindByRestrictedSummaryReferenceID mencari kecocokan restricted_summary_reference_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByRestrictedSummaryReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "restricted_summary_reference_id", value, limit, offset)
}

// FindByLegalHoldReferenceID mencari kecocokan legal_hold_reference_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByLegalHoldReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "legal_hold_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsCaseRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
