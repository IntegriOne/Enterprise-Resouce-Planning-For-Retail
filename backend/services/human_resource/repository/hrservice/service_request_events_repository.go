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

// ServiceRequestEventRepository menyediakan kueri bertipe untuk hr_service_request_events.
type ServiceRequestEventRepository struct {
	db *gorm.DB
}

var _ domainrepo.ServiceRequestEventRepository = (*ServiceRequestEventRepository)(nil)

// NewServiceRequestEventRepository memvalidasi koneksi dan menyiapkan repository.
func NewServiceRequestEventRepository(db *gorm.DB) (*ServiceRequestEventRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &ServiceRequestEventRepository{db: db}, nil
}

func (r *ServiceRequestEventRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.ServiceRequestEvent{})
}

// Insert menyimpan record baru tanpa upsert.
func (r *ServiceRequestEventRepository) Insert(ctx context.Context, value *entity.ServiceRequestEvent) error {
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
func (r *ServiceRequestEventRepository) Create(ctx context.Context, value *entity.ServiceRequestEvent) error {
	return r.Insert(ctx, value)
}

// FindByID membaca satu record.
func (r *ServiceRequestEventRepository) FindByID(ctx context.Context, key string) (*entity.ServiceRequestEvent, error) {
	return r.find(ctx, key, false)
}

// SelectByID adalah alias pembacaan primary key.
func (r *ServiceRequestEventRepository) SelectByID(ctx context.Context, key string) (*entity.ServiceRequestEvent, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci record dalam transaksi aktif.
func (r *ServiceRequestEventRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.ServiceRequestEvent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *ServiceRequestEventRepository) find(ctx context.Context, key string, lock bool) (*entity.ServiceRequestEvent, error) {
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
	value := new(entity.ServiceRequestEvent)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca halaman dengan urutan primary key stabil.
func (r *ServiceRequestEventRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList adalah alias pembacaan berhalaman.
func (r *ServiceRequestEventRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi kolom dan mengikat nilai termasuk NULL.
func (r *ServiceRequestEventRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination.
func (r *ServiceRequestEventRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.ServiceRequestEvent, error) {
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
	rows := make([]entity.ServiceRequestEvent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh record.
func (r *ServiceRequestEventRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung record dalam scope yang sama dengan list.
func (r *ServiceRequestEventRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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
func (r *ServiceRequestEventRepository) Update(ctx context.Context, value *entity.ServiceRequestEvent) error {
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
func (r *ServiceRequestEventRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.ServiceRequestEvent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan primary key.
func (r *ServiceRequestEventRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByServiceRequestID mencari kecocokan service_request_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestEventRepository) FindByServiceRequestID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectBy(ctx, "service_request_id", value, limit, offset)
}

// FindByEventType mencari kecocokan event_type dengan paginasi dan urutan primary key.
func (r *ServiceRequestEventRepository) FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectBy(ctx, "event_type", value, limit, offset)
}

// FindByActorPrincipalID mencari kecocokan actor_principal_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestEventRepository) FindByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectBy(ctx, "actor_principal_id", value, limit, offset)
}

// FindByEvidenceDocumentVersionReferenceID mencari kecocokan evidence_document_version_reference_id dengan paginasi dan urutan primary key.
func (r *ServiceRequestEventRepository) FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectBy(ctx, "evidence_document_version_reference_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *ServiceRequestEventRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.ServiceRequestEvent, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
