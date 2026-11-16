// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package payroll

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// PayrollRunRepository menyediakan kueri bertipe untuk hr_payroll_runs.
type PayrollRunRepository struct {
	db *gorm.DB
}

var _ domainrepo.PayrollRunRepository = (*PayrollRunRepository)(nil)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *PayrollRunRepository) InsertPayrollRuns(ctx context.Context, v *entity.PayrollRun) error {
	return r.Insert(ctx, v)
}
func (r *PayrollRunRepository) UpdatePayrollRuns(ctx context.Context, v *entity.PayrollRun) error {
	return r.Update(ctx, v)
}
func (r *PayrollRunRepository) DeletePayrollRunsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *PayrollRunRepository) FindPayrollRunsByUUID(ctx context.Context, id string) (*entity.PayrollRun, error) {
	return r.FindByID(ctx, id)
}
func (r *PayrollRunRepository) FindPayrollRunsByUUIDForUpdate(ctx context.Context, id string) (*entity.PayrollRun, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *PayrollRunRepository) GetAllPayrollRuns(ctx context.Context, limit, offset int) ([]entity.PayrollRun, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *PayrollRunRepository) CountPayrollRuns(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *PayrollRunRepository) SelectPayrollRunsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *PayrollRunRepository) CountPayrollRunsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *PayrollRunRepository) FindPayrollRunsByPayrollPeriodID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPayrollPeriodID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByPayrollGroupID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPayrollGroupID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByEmployingLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByEmployingLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByCurrencyCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByCurrencyCode(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRuleSetID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRuleSetID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByOriginalRunID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByOriginalRunID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRunNumber(ctx context.Context, v int32, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRunNumber(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByRunType(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByRunType(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByStatus(ctx context.Context, v entity.PayrollRunStatus, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByJurisdictionCode(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByJurisdictionCode(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByInputSnapshotReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByInputSnapshotReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByFinanceJournalReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByFinanceJournalReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByPaymentBatchReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByPaymentBatchReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByFinanceSettlementEventReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByFinanceSettlementEventReferenceID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByStartedByPrincipalID(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByStartedByPrincipalID(ctx, v, limit, offset)
}
func (r *PayrollRunRepository) FindPayrollRunsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

// NewPayrollRunRepository memvalidasi koneksi dan menyiapkan repository.
func NewPayrollRunRepository(db *gorm.DB) (*PayrollRunRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PayrollRunRepository{db: db}, nil
}

func (r *PayrollRunRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.PayrollRun{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *PayrollRunRepository) Insert(ctx context.Context, value *entity.PayrollRun) error {
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
func (r *PayrollRunRepository) Create(ctx context.Context, value *entity.PayrollRun) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *PayrollRunRepository) FindByID(ctx context.Context, key string) (*entity.PayrollRun, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *PayrollRunRepository) SelectByID(ctx context.Context, key string) (*entity.PayrollRun, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *PayrollRunRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.PayrollRun, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *PayrollRunRepository) find(ctx context.Context, key string, lock bool) (*entity.PayrollRun, error) {
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
	value := new(entity.PayrollRun)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *PayrollRunRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *PayrollRunRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.PayrollRun, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *PayrollRunRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *PayrollRunRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.PayrollRun, error) {
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
	rows := make([]entity.PayrollRun, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *PayrollRunRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *PayrollRunRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *PayrollRunRepository) Update(ctx context.Context, value *entity.PayrollRun) error {
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
func (r *PayrollRunRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.PayrollRun{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *PayrollRunRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByPayrollPeriodID mencari kecocokan payroll_period_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByPayrollPeriodID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "payroll_period_id", value, limit, offset)
}

// FindByPayrollGroupID mencari kecocokan payroll_group_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByPayrollGroupID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "payroll_group_id", value, limit, offset)
}

// FindByEmployingLegalEntityReferenceID mencari kecocokan employing_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByEmployingLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "employing_legal_entity_reference_id", value, limit, offset)
}

// FindByCurrencyCode mencari kecocokan currency_code dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByCurrencyCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "currency_code", value, limit, offset)
}

// FindByRuleSetID mencari kecocokan rule_set_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByRuleSetID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "rule_set_id", value, limit, offset)
}

// FindByOriginalRunID mencari kecocokan original_run_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByOriginalRunID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "original_run_id", value, limit, offset)
}

// FindByRunNumber mencari kecocokan run_number dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByRunNumber(ctx context.Context, value int32, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "run_number", value, limit, offset)
}

// FindByRunType mencari kecocokan run_type dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByRunType(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "run_type", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByStatus(ctx context.Context, value entity.PayrollRunStatus, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByJurisdictionCode mencari kecocokan jurisdiction_code dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "jurisdiction_code", value, limit, offset)
}

// FindByInputSnapshotReferenceID mencari kecocokan input_snapshot_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByInputSnapshotReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "input_snapshot_reference_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}

// FindByFinanceJournalReferenceID mencari kecocokan finance_journal_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByFinanceJournalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "finance_journal_reference_id", value, limit, offset)
}

// FindByPaymentBatchReferenceID mencari kecocokan payment_batch_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByPaymentBatchReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "payment_batch_reference_id", value, limit, offset)
}

// FindByFinanceSettlementEventReferenceID mencari kecocokan finance_settlement_event_reference_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByFinanceSettlementEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "finance_settlement_event_reference_id", value, limit, offset)
}

// FindByStartedByPrincipalID mencari kecocokan started_by_principal_id dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByStartedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "started_by_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *PayrollRunRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.PayrollRun, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
