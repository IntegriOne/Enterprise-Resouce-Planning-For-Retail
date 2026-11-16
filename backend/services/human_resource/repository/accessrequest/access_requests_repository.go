package accessrequest

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// AccessRequestRepository menyediakan kueri bertipe untuk hr_access_requests.
type AccessRequestRepository struct {
	db *gorm.DB
}

var _ domainrepo.AccessRequestRepository = (*AccessRequestRepository)(nil)

// NewAccessRequestRepository memvalidasi koneksi dan menyiapkan repository.
func NewAccessRequestRepository(db *gorm.DB) (*AccessRequestRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &AccessRequestRepository{db: db}, nil
}

func (r *AccessRequestRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.AccessRequest{})
}

// Insert menyimpan permintaan akses lokal tanpa mengubah grant IAM.
func (r *AccessRequestRepository) Insert(ctx context.Context, value *entity.AccessRequest) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create mempertahankan alias operasi INSERT.
func (r *AccessRequestRepository) Create(ctx context.Context, value *entity.AccessRequest) error {
	return r.Insert(ctx, value)
}

// FindByID membaca permintaan berdasarkan ID.
func (r *AccessRequestRepository) FindByID(ctx context.Context, key string) (*entity.AccessRequest, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *AccessRequestRepository) SelectByID(ctx context.Context, key string) (*entity.AccessRequest, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci permintaan dalam transaksi aktif.
func (r *AccessRequestRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.AccessRequest, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *AccessRequestRepository) find(ctx context.Context, key string, lock bool) (*entity.AccessRequest, error) {
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
	value := &entity.AccessRequest{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman permintaan akses.
func (r *AccessRequestRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *AccessRequestRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.AccessRequest, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memakai filter kolom yang diverifikasi terhadap model.
func (r *AccessRequestRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan scope sebelum pagination dan count terpisah.
func (r *AccessRequestRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.AccessRequest, error) {
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
	rows := make([]entity.AccessRequest, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh permintaan akses lokal.
func (r *AccessRequestRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung hanya permintaan yang memenuhi filter.
func (r *AccessRequestRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update menyimpan snapshot lengkap non-key, termasuk nilai nol dan NULL.
func (r *AccessRequestRepository) Update(ctx context.Context, value *entity.AccessRequest) error {
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

// DeleteByID menghapus satu permintaan lokal tanpa menulis tabel IAM.
func (r *AccessRequestRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.AccessRequest{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *AccessRequestRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByLifecycleEventID mencari kecocokan lifecycle_event_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByLifecycleEventID(ctx context.Context, value int64, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "lifecycle_event_id", value, limit, offset)
}

// FindByRequestedScopeType mencari kecocokan requested_scope_type dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByRequestedScopeType(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "requested_scope_type", value, limit, offset)
}

// FindByRequestedScopeReferenceID mencari kecocokan requested_scope_reference_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByRequestedScopeReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "requested_scope_reference_id", value, limit, offset)
}

// FindByApprovedCatalogReferenceID mencari kecocokan approved_catalog_reference_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByApprovedCatalogReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "approved_catalog_reference_id", value, limit, offset)
}

// FindByIAMRequestReferenceID mencari kecocokan iam_request_reference_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByIAMRequestReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "iam_request_reference_id", value, limit, offset)
}

// FindByIAMDecisionEventReferenceID mencari kecocokan iam_decision_event_reference_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByIAMDecisionEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "iam_decision_event_reference_id", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByRequestedByPrincipalID mencari kecocokan requested_by_principal_id dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "requested_by_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *AccessRequestRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.AccessRequest, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
