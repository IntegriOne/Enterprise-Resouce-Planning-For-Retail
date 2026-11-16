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

// EmergencyContactRepository menyediakan kueri bertipe untuk hr_emergency_contacts.
type EmergencyContactRepository struct {
	db *gorm.DB
}

var _ domainrepo.EmergencyContactRepository = (*EmergencyContactRepository)(nil)

// Method bernama objek menjadi kontrak eksplisit kontak darurat.
func (r *EmergencyContactRepository) InsertEmergencyContacts(ctx context.Context, v *entity.EmergencyContact) error {
	return r.Insert(ctx, v)
}
func (r *EmergencyContactRepository) UpdateEmergencyContacts(ctx context.Context, v *entity.EmergencyContact) error {
	return r.Update(ctx, v)
}
func (r *EmergencyContactRepository) DeleteEmergencyContactsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *EmergencyContactRepository) FindEmergencyContactsByUUID(ctx context.Context, id string) (*entity.EmergencyContact, error) {
	return r.FindByID(ctx, id)
}
func (r *EmergencyContactRepository) FindEmergencyContactsByUUIDForUpdate(ctx context.Context, id string) (*entity.EmergencyContact, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EmergencyContactRepository) GetAllEmergencyContacts(ctx context.Context, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmergencyContactRepository) CountEmergencyContacts(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EmergencyContactRepository) SelectEmergencyContactsWhere(ctx context.Context, f []domainrepo.Filter, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectWhere(ctx, f, limit, offset)
}
func (r *EmergencyContactRepository) CountEmergencyContactsWhere(ctx context.Context, f []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, f)
}
func (r *EmergencyContactRepository) FindEmergencyContactsByWorkerID(ctx context.Context, v string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.FindByWorkerID(ctx, v, limit, offset)
}
func (r *EmergencyContactRepository) FindEmergencyContactsByContactPartyReferenceID(ctx context.Context, v string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.FindByContactPartyReferenceID(ctx, v, limit, offset)
}
func (r *EmergencyContactRepository) FindEmergencyContactsByRelationshipCode(ctx context.Context, v string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.FindByRelationshipCode(ctx, v, limit, offset)
}

// NewEmergencyContactRepository memvalidasi koneksi dan menyiapkan repository.
func NewEmergencyContactRepository(db *gorm.DB) (*EmergencyContactRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &EmergencyContactRepository{db: db}, nil
}

func (r *EmergencyContactRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.EmergencyContact{})
}

// Insert menyimpan kontak darurat dengan nilai nullable tetap eksplisit.
func (r *EmergencyContactRepository) Insert(ctx context.Context, value *entity.EmergencyContact) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}
func (r *EmergencyContactRepository) Create(ctx context.Context, value *entity.EmergencyContact) error {
	return r.Insert(ctx, value)
}
func (r *EmergencyContactRepository) FindByID(ctx context.Context, key string) (*entity.EmergencyContact, error) {
	return r.find(ctx, key, false)
}
func (r *EmergencyContactRepository) SelectByID(ctx context.Context, key string) (*entity.EmergencyContact, error) {
	return r.FindByID(ctx, key)
}
func (r *EmergencyContactRepository) FindByIDForUpdate(ctx context.Context, key string) (*entity.EmergencyContact, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}
func (r *EmergencyContactRepository) find(ctx context.Context, key string, lock bool) (*entity.EmergencyContact, error) {
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
	value := new(entity.EmergencyContact)
	if err := db.Where("id = ?", key).First(value).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return value, nil
}
func (r *EmergencyContactRepository) GetAll(ctx context.Context, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectWhere(ctx, nil, limit, offset)
}
func (r *EmergencyContactRepository) SelectList(ctx context.Context, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmergencyContactRepository) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectWhere(ctx, []domainrepo.Filter{{Column: column, Values: []any{value}}}, limit, offset)
}

// SelectWhere memfilter kontak sebelum pagination dan mengembalikan slice kosong non-nil.
func (r *EmergencyContactRepository) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmergencyContact, error) {
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
	rows := make([]entity.EmergencyContact, 0)
	if err := db.Order("id ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return rows, nil
}
func (r *EmergencyContactRepository) Count(ctx context.Context) (int64, error) {
	return r.CountWhere(ctx, nil)
}
func (r *EmergencyContactRepository) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
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

// Update mengganti snapshot lengkap dan tidak menyisipkan ulang ID yang hilang.
func (r *EmergencyContactRepository) Update(ctx context.Context, value *entity.EmergencyContact) error {
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
func (r *EmergencyContactRepository) DeleteByID(ctx context.Context, key string) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	if err := pgstore.ValidateKeys(key); err != nil {
		return err
	}
	result := db.Where("id = ?", key).Delete(&entity.EmergencyContact{})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}
func (r *EmergencyContactRepository) Delete(ctx context.Context, key string) error {
	return r.DeleteByID(ctx, key)
}

// FindByWorkerID mencari kecocokan worker_id dengan paginasi dan urutan primary key.
func (r *EmergencyContactRepository) FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectBy(ctx, "worker_id", value, limit, offset)
}

// FindByContactPartyReferenceID mencari kecocokan contact_party_reference_id dengan paginasi dan urutan primary key.
func (r *EmergencyContactRepository) FindByContactPartyReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectBy(ctx, "contact_party_reference_id", value, limit, offset)
}

// FindByRelationshipCode mencari kecocokan relationship_code dengan paginasi dan urutan primary key.
func (r *EmergencyContactRepository) FindByRelationshipCode(ctx context.Context, value string, limit, offset int) ([]entity.EmergencyContact, error) {
	return r.SelectBy(ctx, "relationship_code", value, limit, offset)
}
