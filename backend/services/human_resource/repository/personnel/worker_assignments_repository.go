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

// WorkerAssignmentRepository menyediakan kueri bertipe untuk hr_worker_assignments.
type WorkerAssignmentRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerAssignmentRepository = (*WorkerAssignmentRepository)(nil)

// Method bernama objek menjadi API eksplisit penempatan pekerja.
func (r *WorkerAssignmentRepository) InsertWorkerAssignments(ctx context.Context, v *entity.WorkerAssignment) error {
	return r.Insert(ctx, v)
}
func (r *WorkerAssignmentRepository) UpdateWorkerAssignments(ctx context.Context, v *entity.WorkerAssignment) error {
	return r.Update(ctx, v)
}
func (r *WorkerAssignmentRepository) DeleteWorkerAssignmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByUUID(ctx context.Context, id string) (*entity.WorkerAssignment, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.WorkerAssignment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerAssignmentRepository) GetAllWorkerAssignments(ctx context.Context, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerAssignmentRepository) CountWorkerAssignments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerAssignmentRepository) SelectWorkerAssignmentsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkerAssignmentRepository) CountWorkerAssignmentsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByPositionID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByPositionID(ctx, v, limit, offset)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByManagerWorkerID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByManagerWorkerID(ctx, v, limit, offset)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByAssignmentType(ctx context.Context, v string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByAssignmentType(ctx, v, limit, offset)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByRequestedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByRequestedByPrincipalID(ctx, v, limit, offset)
}
func (r *WorkerAssignmentRepository) FindWorkerAssignmentsByApprovedByPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.FindByApprovedByPrincipalID(ctx, v, limit, offset)
}

// NewWorkerAssignmentRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerAssignmentRepository(db *gorm.DB) (*WorkerAssignmentRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerAssignmentRepository{db: db}, nil
}

func (r *WorkerAssignmentRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerAssignment{})
}

// Insert menyimpan penempatan pekerja tanpa keputusan bisnis di repository.
func (r *WorkerAssignmentRepository) Insert(ctx context.Context, value *entity.WorkerAssignment) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerAssignmentRepository) Create(ctx context.Context, value *entity.WorkerAssignment) error {
	return r.Insert(ctx, value)
}
func (r *WorkerAssignmentRepository) FindByID(ctx context.Context, key string) (*entity.WorkerAssignment, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerAssignmentRepository) SelectByID(ctx context.Context, key string) (*entity.WorkerAssignment, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerAssignmentRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.WorkerAssignment, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerAssignmentRepository) find(ctx context.Context, key string, lock bool) (*entity.WorkerAssignment, error) {
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
	value := new(entity.WorkerAssignment)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkerAssignmentRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerAssignmentRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerAssignmentRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter penempatan sebelum pembatasan halaman.
func (r *WorkerAssignmentRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerAssignment, error) {
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
	rows := make([]entity.WorkerAssignment, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerAssignmentRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerAssignmentRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key dan mempertahankan nilai nol serta NULL.
func (r *WorkerAssignmentRepository) Update(ctx context.Context, value *entity.WorkerAssignment) error {
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
func (r *WorkerAssignmentRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerAssignment{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerAssignmentRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerAssignmentRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByPositionID mencari kecocokan position_id dengan paginasi dan urutan primary key.
func (r *WorkerAssignmentRepository) FindByPositionID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "position_id", value, limit, offset)
}

// FindByManagerWorkerID mencari kecocokan manager_worker_id dengan paginasi dan urutan primary key.
func (r *WorkerAssignmentRepository) FindByManagerWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "manager_worker_id", value, limit, offset)
}

// FindByAssignmentType mencari kecocokan assignment_type dengan paginasi dan urutan primary key.
func (r *WorkerAssignmentRepository) FindByAssignmentType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "assignment_type", value, limit, offset)
}

// FindByRequestedByPrincipalID mencari kecocokan requested_by_principal_id dengan paginasi.
func (r *WorkerAssignmentRepository) FindByRequestedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "requested_by_principal_id", value, limit, offset)
}

// FindByApprovedByPrincipalID mencari kecocokan approved_by_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerAssignmentRepository) FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerAssignment, error) {
	return r.SelectBy(ctx, "approved_by_principal_id", value, limit, offset)
}
