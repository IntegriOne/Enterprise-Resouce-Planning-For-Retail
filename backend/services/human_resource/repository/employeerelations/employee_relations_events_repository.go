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

// EmployeeRelationsEventRepository menyediakan kueri bertipe untuk hr_employee_relations_events.
type EmployeeRelationsEventRepository struct {
	db *gorm.DB
}

var _ domainrepo.EmployeeRelationsEventRepository = (*EmployeeRelationsEventRepository)(nil)

// NewEmployeeRelationsEventRepository memvalidasi koneksi dan menyiapkan repository.
func NewEmployeeRelationsEventRepository(db *gorm.DB) (*EmployeeRelationsEventRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EmployeeRelationsEventRepository{db: db}, nil
}

// ID event berupa int64, sehingga nama pencarian memakai ByID, bukan ByUUID.
func (r *EmployeeRelationsEventRepository) InsertEmployeeRelationsEvents(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
	return r.Insert(ctx, value)
}
func (r *EmployeeRelationsEventRepository) UpdateEmployeeRelationsEvents(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
	return r.Update(ctx, value)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByID(ctx context.Context, id int64) (*entity.EmployeeRelationsEvent, error) {
	return r.FindByID(ctx, id)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByIDForUpdate(ctx context.Context, id int64) (*entity.EmployeeRelationsEvent, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EmployeeRelationsEventRepository) DeleteEmployeeRelationsEventsByID(ctx context.Context, id int64) error {
	return r.DeleteByID(ctx, id)
}
func (r *EmployeeRelationsEventRepository) GetAllEmployeeRelationsEvents(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmployeeRelationsEventRepository) CountEmployeeRelationsEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EmployeeRelationsEventRepository) SelectEmployeeRelationsEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EmployeeRelationsEventRepository) CountEmployeeRelationsEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByCaseID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByCaseID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByActorPrincipalID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByEventReferenceID(ctx, value, limit, offset)
}

func (r *EmployeeRelationsEventRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EmployeeRelationsEvent{})
}

// Insert menyimpan event dan menerima ID identity dari PostgreSQL.
func (r *EmployeeRelationsEventRepository) Insert(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
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
func (r *EmployeeRelationsEventRepository) Create(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
	return r.Insert(ctx, value)
}

// FindByID membaca event berdasarkan ID integer.
func (r *EmployeeRelationsEventRepository) FindByID(ctx context.Context, key int64) (*entity.EmployeeRelationsEvent, error) {
	return r.find(ctx, key, false)
}

// SelectByID mempertahankan alias SELECT berdasarkan ID.
func (r *EmployeeRelationsEventRepository) SelectByID(ctx context.Context, key int64) (*entity.EmployeeRelationsEvent, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci event dalam transaksi aktif.
func (r *EmployeeRelationsEventRepository) FindByIDForUpdate(ctx context.Context, key int64) (*entity.EmployeeRelationsEvent, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *EmployeeRelationsEventRepository) find(ctx context.Context, key int64, lock bool) (*entity.EmployeeRelationsEvent, error) {
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
	value := &entity.EmployeeRelationsEvent{}
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman event dengan urutan ID ascending.
func (r *EmployeeRelationsEventRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}

// SelectList mempertahankan alias daftar berhalaman.
func (r *EmployeeRelationsEventRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy memakai kolom yang diverifikasi terhadap model.
func (r *EmployeeRelationsEventRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere menerapkan filter sebelum pagination.
func (r *EmployeeRelationsEventRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
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
	rows := make([]entity.EmployeeRelationsEvent, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}

// Count menghitung seluruh event.
func (r *EmployeeRelationsEventRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}

// CountWhere menghitung event yang memenuhi filter.
func (r *EmployeeRelationsEventRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti seluruh field non-key tanpa upsert; trigger immutable tetap berlaku.
func (r *EmployeeRelationsEventRepository) Update(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
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

// DeleteByID menghapus satu ID dengan tetap mengikuti trigger/FK database.
func (r *EmployeeRelationsEventRepository) DeleteByID(ctx context.Context, key int64) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EmployeeRelationsEvent{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete mempertahankan alias penghapusan berdasarkan ID.
func (r *EmployeeRelationsEventRepository) Delete(ctx context.Context, key int64) error {
	return r.DeleteByID(ctx, key)
}

// FindByCaseID mencari kecocokan case_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsEventRepository) FindByCaseID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectBy(ctx, "case_id", value, limit, offset)
}

// FindByEventType mencari kecocokan event_type dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsEventRepository) FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectBy(ctx, "event_type", value, limit, offset)
}

// FindByActorPrincipalID mencari kecocokan actor_principal_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsEventRepository) FindByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectBy(ctx, "actor_principal_id", value, limit, offset)
}

// FindByEventReferenceID mencari kecocokan event_reference_id dengan paginasi dan urutan primary key.
func (r *EmployeeRelationsEventRepository) FindByEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectBy(ctx, "event_reference_id", value, limit, offset)
}
