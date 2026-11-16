// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package hrservice

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// ServiceRequestRepository menyediakan kueri bertipe untuk hr_service_requests.
type ServiceRequestRepository struct {
	db *gorm.DB
}

var _ domainrepo.ServiceRequestRepository = (*ServiceRequestRepository)(nil)

// NewServiceRequestRepository memvalidasi koneksi dan menyiapkan repository.
func NewServiceRequestRepository(db *gorm.DB) (*ServiceRequestRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &ServiceRequestRepository{db: db}, nil
}

func (r *ServiceRequestRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.ServiceRequest{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *ServiceRequestRepository) Insert(ctx context.Context, value *entity.ServiceRequest) error {
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
func (r *ServiceRequestRepository) Create(ctx context.Context, value *entity.ServiceRequest) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *ServiceRequestRepository) FindByID(ctx context.Context, key string) (*entity.ServiceRequest, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *ServiceRequestRepository) SelectByID(ctx context.Context, key string) (*entity.ServiceRequest, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *ServiceRequestRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.ServiceRequest, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *ServiceRequestRepository) find(ctx context.Context, key string, lock bool) (*entity.ServiceRequest, error) {
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
	value := new(entity.ServiceRequest)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *ServiceRequestRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *ServiceRequestRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *ServiceRequestRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *ServiceRequestRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ServiceRequest, error) {
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
	rows := make([]entity.ServiceRequest, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *ServiceRequestRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *ServiceRequestRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *ServiceRequestRepository) Update(ctx context.Context, value *entity.ServiceRequest) error {
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
func (r *ServiceRequestRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.ServiceRequest{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *ServiceRequestRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByRequestNumber mencari kecocokan request_number dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByRequestNumber(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "request_number", value, limit, offset)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByRequestType mencari kecocokan request_type dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByRequestType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "request_type", value, limit, offset)
}

// FindByRequestedByPrincipalID mencari kecocokan requested_by_principal_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "requested_by_principal_id", value, limit, offset)
}

// FindByAssignedPrincipalID mencari kecocokan assigned_principal_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByAssignedPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "assigned_principal_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByRestrictedDetailsDocumentVersionReferenceID mencari kecocokan restricted_details_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByRestrictedDetailsDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "restricted_details_document_version_reference_id", value, limit, offset)
}

// FindByResolutionDocumentVersionReferenceID mencari kecocokan resolution_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByResolutionDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "resolution_document_version_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *ServiceRequestRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequest, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
