// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package personnel

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkerRepository menyediakan kueri bertipe untuk hr_workers.
type WorkerRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerRepository = (*WorkerRepository)(nil)

// Method bernama objek memperjelas tabel workers dan kolom pencariannya.
func (r *WorkerRepository) InsertWorkers(ctx context.Context, v *entity.Worker) error {
	return r.Insert(ctx, v)
}
func (r *WorkerRepository) UpdateWorkers(ctx context.Context, v *entity.Worker) error {
	return r.Update(ctx, v)
}
func (r *WorkerRepository) DeleteWorkersByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerRepository) FindWorkersByUUID(ctx context.Context, id string) (*entity.Worker, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerRepository) FindWorkersByUUIDForUpdate(ctx context.Context, id string) (*entity.Worker, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerRepository) GetAllWorkers(ctx context.Context, limit, offset int) ([]entity.Worker, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerRepository) CountWorkers(ctx context.Context) (int64, error) { return r.Count(ctx) }
func (r *WorkerRepository) SelectWorkersWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.Worker, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkerRepository) CountWorkersWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkerRepository) FindWorkersByWorkerNumber(ctx context.Context, v string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByWorkerNumber(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByPersonPartyReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByPersonPartyReferenceID(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByStatus(ctx context.Context, v entity.WorkerStatus, limit, offset int) ([]entity.Worker, error) {
	return r.FindByStatus(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByTerminationReasonCode(ctx context.Context, v *string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByTerminationReasonCode(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByPrimaryLegalEntityReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByPrimaryLegalEntityReferenceID(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByPrimaryAssignmentID(ctx context.Context, v *string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByPrimaryAssignmentID(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByIAMPrincipalReferenceID(ctx context.Context, v *string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByIAMPrincipalReferenceID(ctx, v, limit, offset)
}
func (r *WorkerRepository) FindWorkersByCreatedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.Worker, error) {
	return r.FindByCreatedByPrincipalID(ctx, v, limit, offset)
}

// NewWorkerRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerRepository(db *gorm.DB) (*WorkerRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerRepository{db: db}, nil
}

func (r *WorkerRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.Worker{})
}

// Insert menyimpan pekerja dengan nilai field dan timestamp dari service.
func (r *WorkerRepository) Insert(ctx context.Context, value *entity.Worker) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerRepository) Create(ctx context.Context, value *entity.Worker) error {
	return r.Insert(ctx, value)
}
func (r *WorkerRepository) FindByID(ctx context.Context, key string) (*entity.Worker, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerRepository) SelectByID(ctx context.Context, key string) (*entity.Worker, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.Worker, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerRepository) find(ctx context.Context, key string, lock bool) (*entity.Worker, error) {
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
	value := new(entity.Worker)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkerRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.Worker, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.Worker, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.Worker, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter pekerja sebelum pagination dan memakai urutan ID stabil.
func (r *WorkerRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.Worker, error) {
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
	rows := make([]entity.Worker, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerRepository) Count(ctx context.Context) (int64, error) { return r.CountWhere(ctx, nil) }
func (r *WorkerRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot pekerja lengkap tanpa mengubah ID atau melakukan upsert.
func (r *WorkerRepository) Update(ctx context.Context, value *entity.Worker) error {
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
func (r *WorkerRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.Worker{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerNumber mencari kecocokan worker_number dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByWorkerNumber(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "worker_number", value, limit, offset)
}

// FindByPersonPartyReferenceID mencari kecocokan person_party_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByPersonPartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "person_party_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByStatus(ctx context.Context, value entity.WorkerStatus, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByTerminationReasonCode mencari kecocokan termination_reason_code dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByTerminationReasonCode(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "termination_reason_code", value, limit, offset)
}

// FindByPrimaryLegalEntityReferenceID mencari kecocokan primary_legal_entity_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByPrimaryLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "primary_legal_entity_reference_id", value, limit, offset)
}

// FindByPrimaryAssignmentID mencari kecocokan primary_assignment_id dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByPrimaryAssignmentID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "primary_assignment_id", value, limit, offset)
}

// FindByIAMPrincipalReferenceID mencari kecocokan iam_principal_reference_id dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByIAMPrincipalReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "iam_principal_reference_id", value, limit, offset)
}

// FindByCreatedByPrincipalID mencari kecocokan created_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerRepository) FindByCreatedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.Worker, error) {
	return r.SelectBy(ctx, "created_by_principal_id", value, limit, offset)
}
