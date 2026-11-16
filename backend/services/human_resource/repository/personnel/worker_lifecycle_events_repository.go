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

// WorkerLifecycleEventRepository menyediakan kueri bertipe untuk hr_worker_lifecycle_events.
type WorkerLifecycleEventRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerLifecycleEventRepository = (*WorkerLifecycleEventRepository)(nil)

// Method bernama objek memakai ByID karena primary key event berupa identity int64.
func (r *WorkerLifecycleEventRepository) InsertWorkerLifecycleEvents(ctx context.Context, v *entity.WorkerLifecycleEvent) error {
	return r.Insert(ctx, v)
}
func (r *WorkerLifecycleEventRepository) UpdateWorkerLifecycleEvents(ctx context.Context, v *entity.WorkerLifecycleEvent) error {
	return r.Update(ctx, v)
}
func (r *WorkerLifecycleEventRepository) DeleteWorkerLifecycleEventsByID(ctx context.Context, id int64) error {
	return r.DeleteByID(ctx, id)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByID(ctx context.Context, id int64) (*entity.WorkerLifecycleEvent, error) {
	return r.FindByID(ctx, id)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByIDForUpdate(ctx context.Context, id int64) (*entity.WorkerLifecycleEvent, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *WorkerLifecycleEventRepository) GetAllWorkerLifecycleEvents(ctx context.Context, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerLifecycleEventRepository) CountWorkerLifecycleEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *WorkerLifecycleEventRepository) SelectWorkerLifecycleEventsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *WorkerLifecycleEventRepository) CountWorkerLifecycleEventsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByEventType(ctx context.Context, v string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByEventType(ctx, v, limit, offset)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByPreviousStatus(ctx context.Context, v *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByPreviousStatus(ctx, v, limit, offset)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByNewStatus(ctx context.Context, v *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByNewStatus(ctx, v, limit, offset)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByActorPrincipalID(ctx context.Context, v *string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByActorPrincipalID(ctx, v, limit, offset)
}
func (r *WorkerLifecycleEventRepository) FindWorkerLifecycleEventsByIdempotencyKey(ctx context.Context, v string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.FindByIdempotencyKey(ctx, v, limit, offset)
}

// NewWorkerLifecycleEventRepository memvalidasi koneksi dan menyiapkan repository.
func NewWorkerLifecycleEventRepository(db *gorm.DB) (*WorkerLifecycleEventRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerLifecycleEventRepository{db: db}, nil
}

func (r *WorkerLifecycleEventRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerLifecycleEvent{})
}

// Insert menyimpan event lifecycle dan menerima ID identity melalui RETURNING.
func (r *WorkerLifecycleEventRepository) Insert(ctx context.Context, value *entity.WorkerLifecycleEvent) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *WorkerLifecycleEventRepository) Create(ctx context.Context, value *entity.WorkerLifecycleEvent) error {
	return r.Insert(ctx, value)
}
func (r *WorkerLifecycleEventRepository) FindByID(ctx context.Context, key int64) (*entity.WorkerLifecycleEvent, error) {
	return r.find(ctx, key, false)
}
func (r *WorkerLifecycleEventRepository) SelectByID(ctx context.Context, key int64) (*entity.WorkerLifecycleEvent, error) {
	return r.FindByID(ctx, key)
}
func (r *WorkerLifecycleEventRepository) FindByIDForUpdate(ctx context.Context, key int64) (*entity.WorkerLifecycleEvent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *WorkerLifecycleEventRepository) find(ctx context.Context, key int64, lock bool) (*entity.WorkerLifecycleEvent, error) {
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
	value := new(entity.WorkerLifecycleEvent)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *WorkerLifecycleEventRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *WorkerLifecycleEventRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *WorkerLifecycleEventRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menyaring event sebelum pagination dengan urutan identity ascending.
func (r *WorkerLifecycleEventRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
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
	rows := make([]entity.WorkerLifecycleEvent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *WorkerLifecycleEventRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *WorkerLifecycleEventRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update tunduk pada invariant immutable database tanpa mengubah ID identity.
func (r *WorkerLifecycleEventRepository) Update(ctx context.Context, value *entity.WorkerLifecycleEvent) error {
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
func (r *WorkerLifecycleEventRepository) DeleteByID(ctx context.Context, key int64) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.WorkerLifecycleEvent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *WorkerLifecycleEventRepository) Delete(ctx context.Context, key int64) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByEventType mencari kecocokan event_type dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "event_type", value, limit, offset)
}

// FindByPreviousStatus mencari kecocokan previous_status dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByPreviousStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "previous_status", value, limit, offset)
}

// FindByNewStatus mencari kecocokan new_status dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByNewStatus(ctx context.Context, value *entity.WorkerStatus, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "new_status", value, limit, offset)
}

// FindByActorPrincipalID mencari kecocokan actor_principal_id dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "actor_principal_id", value, limit, offset)
}

// FindByIdempotencyKey mencari kecocokan idempotency_key dengan paginasi dan urutan primary key.
func (r *WorkerLifecycleEventRepository) FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.WorkerLifecycleEvent, error) {
	return r.SelectBy(ctx, "idempotency_key", value, limit, offset)
}
