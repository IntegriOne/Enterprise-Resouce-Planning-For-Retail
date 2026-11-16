package inbox

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// InboxEventRepository menyediakan kueri bertipe untuk hr_inbox_events.
type InboxEventRepository struct {
	db *gorm.DB
}

var _ domainrepo.InboxEventRepository = (*InboxEventRepository)(nil)

// NewInboxEventRepository memvalidasi koneksi dan menyiapkan repository.
func NewInboxEventRepository(db *gorm.DB) (*InboxEventRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &InboxEventRepository{db: db}, nil
}

// Operasi bernama objek mempertahankan jalur transaksi yang sama dengan method lama.
func (r *InboxEventRepository) InsertInboxEvents(ctx context.Context, value *entity.InboxEvent) error {
	return r.Insert(ctx, value)
}
func (r *InboxEventRepository) UpdateInboxEvents(ctx context.Context, value *entity.InboxEvent) error {
	return r.Update(ctx, value)
}
func (r *InboxEventRepository) FindInboxEventsByUUID(ctx context.Context, uuid string) (*entity.InboxEvent, error) {
	return r.FindByID(ctx, uuid)
}
func (r *InboxEventRepository) FindInboxEventsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.InboxEvent, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *InboxEventRepository) DeleteInboxEventsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *InboxEventRepository) GetAllInboxEvents(ctx context.Context, limit, offset int) ([]entity.InboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InboxEventRepository) CountInboxEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *InboxEventRepository) SelectInboxEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *InboxEventRepository) CountInboxEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *InboxEventRepository) FindInboxEventsBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindBySourceManagement(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindBySourceEventID(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByFailureCode(ctx context.Context, value *string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByFailureCode(ctx, value, limit, offset)
}

func (r *InboxEventRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.InboxEvent{})
}

// Insert menyimpan satu event masuk dalam transaksi pemanggil.
func (r *InboxEventRepository) Insert(ctx context.Context, value *entity.InboxEvent) error {
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
func (r *InboxEventRepository) Create(ctx context.Context, value *entity.InboxEvent) error {
	return r.Insert(ctx, value)
}

// FindByID membaca event berdasarkan ID.
func (r *InboxEventRepository) FindByID(ctx context.Context, key string) (*entity.InboxEvent, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *InboxEventRepository) SelectByID(ctx context.Context, key string) (*entity.InboxEvent, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci event dalam transaksi aktif.
func (r *InboxEventRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.InboxEvent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *InboxEventRepository) find(ctx context.Context, key string, lock bool) (*entity.InboxEvent, error) {
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
	value := &entity.InboxEvent{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman event masuk.
func (r *InboxEventRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *InboxEventRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.InboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memakai kolom model yang diverifikasi sebelum membentuk filter SQL.
func (r *InboxEventRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination dengan urutan primary key stabil.
func (r *InboxEventRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.InboxEvent, error) {
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
	rows := make([]entity.InboxEvent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh event masuk.
func (r *InboxEventRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung hanya event yang memenuhi filter.
func (r *InboxEventRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key lengkap tanpa menyisipkan event yang hilang.
func (r *InboxEventRepository) Update(ctx context.Context, value *entity.InboxEvent) error {
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

// DeleteByID menghapus satu event sesuai constraint database.
func (r *InboxEventRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.InboxEvent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *InboxEventRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindBySourceManagement mencari kecocokan source_management dengan paginasi dan urutan primary key.
func (r *InboxEventRepository) FindBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectBy(ctx, "source_management", value, limit, offset)
}

// FindBySourceEventID mencari kecocokan source_event_id dengan paginasi dan urutan primary key.
func (r *InboxEventRepository) FindBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectBy(ctx, "source_event_id", value, limit, offset)
}

// FindByEventType mencari kecocokan event_type dengan paginasi dan urutan primary key.
func (r *InboxEventRepository) FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectBy(ctx, "event_type", value, limit, offset)
}

// FindByStatus mencari kecocokan status dengan paginasi dan urutan primary key.
func (r *InboxEventRepository) FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectBy(ctx, "status", value, limit, offset)
}

// FindByFailureCode mencari kecocokan failure_code dengan paginasi dan urutan primary key.
func (r *InboxEventRepository) FindByFailureCode(ctx context.Context, value *string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectBy(ctx, "failure_code", value, limit, offset)
}
