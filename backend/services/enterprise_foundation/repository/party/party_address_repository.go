package party

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.PartyAddressRepository = (*PartyAddressRepository)(nil)

// PartyAddressRepository mengimplementasikan PartyAddressRepository menggunakan PostgreSQL dan GORM.
type PartyAddressRepository struct {
	db *gorm.DB
}

// NewPartyAddressRepository membangun PartyAddressRepository dengan validasi koneksi database fail-closed.
func NewPartyAddressRepository(db *gorm.DB) (*PartyAddressRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PartyAddressRepository{db: db}, nil
}

// SelectByID mengambil tautan alamat milik mitra berdasarkan UUID relasi.
func (r *PartyAddressRepository) SelectByID(ctx context.Context, id string) (*entity.PartyAddress, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyAddressModel
	if err := db.Where("id = ?", aid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyAddressNotFound)
	}
	return partyAddressToDomain(&m)
}

// SelectByPartyID mengambil relasi alamat milik satu mitra eksternal secara berhalaman.
func (r *PartyAddressRepository) SelectByPartyID(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	q := db.Model(&partyAddressModel{}).Where("party_id = ?", pid)
	return pgstore.Paginate(q, page, "usage ASC, is_primary DESC, created_at ASC, id ASC", partyAddressToDomain)
}

// SelectByPartyAndUsage mengambil daftar relasi alamat mitra untuk peruntukan tertentu secara berhalaman.
func (r *PartyAddressRepository) SelectByPartyAndUsage(ctx context.Context, partyID string, usage entity.AddressUsageKind, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	if !usage.IsValid() {
		return domainrepo.PageResult[*entity.PartyAddress]{}, eferrors.ErrInvalidAddressUsage
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	q := db.Model(&partyAddressModel{}).Where("party_id = ? AND usage = ?", pid, usage.String())
	return pgstore.Paginate(q, page, "is_primary DESC, created_at ASC, id ASC", partyAddressToDomain)
}

// SelectPrimaryByPartyAndUsage mengambil alamat utama mitra untuk peruntukan tertentu (default billing/shipping address).
func (r *PartyAddressRepository) SelectPrimaryByPartyAndUsage(ctx context.Context, partyID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return nil, err
	}
	if !usage.IsValid() {
		return nil, eferrors.ErrInvalidAddressUsage
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyAddressModel
	if err := db.Where("party_id = ? AND usage = ? AND is_primary = true", pid, usage.String()).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyAddressNotFound)
	}
	return partyAddressToDomain(&m)
}

// SelectByAddressID mengambil tautan mitra yang merujuk pada suatu master alamat secara berhalaman.
func (r *PartyAddressRepository) SelectByAddressID(ctx context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("address_id", addressID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}
	q := db.Model(&partyAddressModel{}).Where("address_id = ?", aid)
	return pgstore.Paginate(q, page, "created_at ASC, id ASC", partyAddressToDomain)
}

// SelectByPartyAddressUsage mengambil tautan unik berdasarkan kombinasi (party_id, address_id, usage).
func (r *PartyAddressRepository) SelectByPartyAddressUsage(ctx context.Context, partyID, addressID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return nil, err
	}
	aid, err := pgstore.ValidateUUIDParam("address_id", addressID)
	if err != nil {
		return nil, err
	}
	if !usage.IsValid() {
		return nil, eferrors.ErrInvalidAddressUsage
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyAddressModel
	if err := db.Where("party_id = ? AND address_id = ? AND usage = ?", pid, aid, usage.String()).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyAddressNotFound)
	}
	return partyAddressToDomain(&m)
}

// SelectList menampilkan daftar relasi alamat mitra secara berhalaman berdasarkan mitra, master alamat, peruntukan, dan penanda utama.
func (r *PartyAddressRepository) SelectList(ctx context.Context, filter domainrepo.PartyAddressFilter) (domainrepo.PageResult[*entity.PartyAddress], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.PartyID != "" {
		if _, err := pgstore.ValidateUUIDParam("party_id", f.PartyID); err != nil {
			return domainrepo.PageResult[*entity.PartyAddress]{}, err
		}
	}
	if f.AddressID != "" {
		if _, err := pgstore.ValidateUUIDParam("address_id", f.AddressID); err != nil {
			return domainrepo.PageResult[*entity.PartyAddress]{}, err
		}
	}
	if f.Usage != nil && !f.Usage.IsValid() {
		return domainrepo.PageResult[*entity.PartyAddress]{}, eferrors.ErrInvalidAddressUsage
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyAddress]{}, err
	}

	q := db.Model(&partyAddressModel{})
	if f.PartyID != "" {
		q = q.Where("party_id = ?", f.PartyID)
	}
	if f.AddressID != "" {
		q = q.Where("address_id = ?", f.AddressID)
	}
	if f.Usage != nil {
		q = q.Where("usage = ?", f.Usage.String())
	}
	if f.PrimaryOnly != nil {
		q = q.Where("is_primary = ?", *f.PrimaryOnly)
	}

	return pgstore.Paginate(q, f.Page, "usage ASC, is_primary DESC, created_at ASC, id ASC", partyAddressToDomain)
}

// UpdateClearPrimaryByPartyUsageExcept menghapus penanda is_primary pada relasi alamat mitra lain untuk peruntukan yang sama.
func (r *PartyAddressRepository) UpdateClearPrimaryByPartyUsageExcept(ctx context.Context, partyID string, usage entity.AddressUsageKind, keepID string, updatedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return err
	}
	if !usage.IsValid() {
		return eferrors.ErrInvalidAddressUsage
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(updatedAt)
	q := db.Model(&partyAddressModel{}).Where("party_id = ? AND usage = ? AND is_primary = true", pid, usage.String())
	if strings.TrimSpace(keepID) != "" {
		kid, err := pgstore.ValidateUUIDParam("keep_id", keepID)
		if err != nil {
			return err
		}
		q = q.Where("id <> ?", kid)
	}
	return pgstore.MapError(q.Updates(map[string]any{
		"is_primary": false,
		"updated_at": ts,
	}).Error, nil)
}

// Insert menyimpan relasi alamat mitra baru ke tabel party_addresses.
func (r *PartyAddressRepository) Insert(ctx context.Context, link *entity.PartyAddress) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyAddressToModel(link)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui peruntukan dan penanda primary pada relasi alamat milik mitra.
func (r *PartyAddressRepository) Update(ctx context.Context, link *entity.PartyAddress) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyAddressToModel(link)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&partyAddressModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"usage":      model.Usage,
			"is_primary": model.IsPrimary,
			"updated_at": model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyAddressNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyAddressNotFound
	}
	return nil
}

// DeleteByID menghapus tautan alamat mitra berdasarkan UUID relasi.
func (r *PartyAddressRepository) DeleteByID(ctx context.Context, id string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Where("id = ?", aid).Delete(&partyAddressModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyAddressNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyAddressNotFound
	}
	return nil
}
