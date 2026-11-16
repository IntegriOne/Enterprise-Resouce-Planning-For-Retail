package outbox

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// OutboxEventRepository menyediakan kueri bertipe untuk hr_outbox_events.
type OutboxEventRepository struct {
	db *gorm.DB
}

var _ domainrepo.OutboxEventRepository = (*OutboxEventRepository)(nil)

// NewOutboxEventRepository memvalidasi koneksi dan menyiapkan repository.
func NewOutboxEventRepository(db *gorm.DB) (*OutboxEventRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &OutboxEventRepository{db: db}, nil
}

// Operasi bernama objek menyimpan/membaca outbox lokal tanpa publikasi jaringan.
func (r *OutboxEventRepository) InsertOutboxEvents(ctx context.Context, value *entity.OutboxEvent) error {
	return r.Insert(ctx, value)
}
func (r *OutboxEventRepository) UpdateOutboxEvents(ctx context.Context, value *entity.OutboxEvent) error {
	return r.Update(ctx, value)
}
func (r *OutboxEventRepository) FindOutboxEventsByUUID(ctx context.Context, uuid string) (*entity.OutboxEvent, error) {
	return r.FindByID(ctx, uuid)
}
func (r *OutboxEventRepository) FindOutboxEventsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.OutboxEvent, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *OutboxEventRepository) DeleteOutboxEventsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *OutboxEventRepository) GetAllOutboxEvents(ctx context.Context, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OutboxEventRepository) CountOutboxEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OutboxEventRepository) SelectOutboxEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *OutboxEventRepository) CountOutboxEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateType(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateID(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateID(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateVersion(ctx context.Context, value int64, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateVersion(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByLastErrorCode(ctx context.Context, value *string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByLastErrorCode(ctx, value, limit, offset)
}

func (r *OutboxEventRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.OutboxEvent{})
}

// Insert menyimpan event dalam transaksi perubahan bisnis, tanpa mengirim ke jaringan.
func (r *OutboxEventRepository) Insert(ctx context.Context, value *entity.OutboxEvent) error {
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
func (r *OutboxEventRepository) Create(ctx context.Context, value *entity.OutboxEvent) error {
	return r.Insert(ctx, value)
}

// FindByID membaca event berdasarkan ID.
func (r *OutboxEventRepository) FindByID(ctx context.Context, key string) (*entity.OutboxEvent, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *OutboxEventRepository) SelectByID(ctx context.Context, key string) (*entity.OutboxEvent, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengambil row lock dalam transaksi aktif.
func (r *OutboxEventRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.OutboxEvent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *OutboxEventRepository) find(ctx context.Context, key string, lock bool) (*entity.OutboxEvent, error) {
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
	value := &entity.OutboxEvent{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman event yang tersimpan.
func (r *OutboxEventRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *OutboxEventRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memvalidasi nama kolom dan mengikat nilai sebagai parameter SQL.
func (r *OutboxEventRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter scope sebelum pagination.
func (r *OutboxEventRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OutboxEvent, error) {
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
	rows := make([]entity.OutboxEvent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh event yang masih disimpan.
func (r *OutboxEventRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung hanya event yang memenuhi filter.
func (r *OutboxEventRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot non-key lengkap tanpa upsert.
func (r *OutboxEventRepository) Update(ctx context.Context, value *entity.OutboxEvent) error {
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

// DeleteByID menghapus satu event berdasarkan ID lengkap.
func (r *OutboxEventRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.OutboxEvent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *OutboxEventRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByAggregateType mencari kecocokan aggregate_type dengan paginasi dan urutan primary key.
func (r *OutboxEventRepository) FindByAggregateType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectBy(ctx, "aggregate_type", value, limit, offset)
}

// FindByAggregateID mencari kecocokan aggregate_id dengan paginasi dan urutan primary key.
func (r *OutboxEventRepository) FindByAggregateID(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectBy(ctx, "aggregate_id", value, limit, offset)
}

// FindByAggregateVersion mencari kecocokan aggregate_version dengan paginasi dan urutan primary key.
func (r *OutboxEventRepository) FindByAggregateVersion(ctx context.Context, value int64, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectBy(ctx, "aggregate_version", value, limit, offset)
}

// FindByEventType mencari kecocokan event_type dengan paginasi dan urutan primary key.
func (r *OutboxEventRepository) FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectBy(ctx, "event_type", value, limit, offset)
}

// FindByLastErrorCode mencari kecocokan last_error_code dengan paginasi dan urutan primary key.
func (r *OutboxEventRepository) FindByLastErrorCode(ctx context.Context, value *string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectBy(ctx, "last_error_code", value, limit, offset)
}
