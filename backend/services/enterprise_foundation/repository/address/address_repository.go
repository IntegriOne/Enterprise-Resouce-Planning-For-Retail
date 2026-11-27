package address

import (
	"context"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.AddressRepository = (*Repository)(nil)

// Repository mengimplementasikan AddressRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun AddressRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil satu master alamat berdasarkan UUID.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.Address, error) {
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
	var m addressModel
	if err := db.Where("id = ?", aid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrAddressNotFound)
	}
	return addressToDomain(&m)
}

// SelectByIDs mengambil beberapa master alamat sekaligus berdasarkan daftar UUID untuk hidrasi batch lintas entitas.
func (r *Repository) SelectByIDs(ctx context.Context, ids []string) ([]*entity.Address, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cleanIDs, err := pgstore.NormalizeUUIDSlice("ids", ids)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*entity.Address{}, nil
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []addressModel
	if err := db.Where("id IN ?", cleanIDs).Order("created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, addressToDomain)
}

// SelectList menampilkan daftar master alamat secara berhalaman berdasarkan kata kunci, kota, provinsi, kode pos, dan negara.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Address]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.CountryCode != "" {
		normalizedCountry, err := entity.ValidateCountryCode("country_code", f.CountryCode)
		if err != nil {
			return domainrepo.PageResult[*entity.Address]{}, err
		}
		f.CountryCode = normalizedCountry
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Address]{}, err
	}

	q := db.Model(&addressModel{})
	if f.CountryCode != "" {
		q = q.Where("country_code = ?", f.CountryCode)
	}
	if f.City != "" {
		q = q.Where(`city ILIKE ? ESCAPE '\'`, pgstore.LikePattern(f.City))
	}
	if f.Province != "" {
		q = q.Where(`COALESCE(province, '') ILIKE ? ESCAPE '\'`, pgstore.LikePattern(f.Province))
	}
	if f.PostalCode != "" {
		q = q.Where("LOWER(COALESCE(postal_code, '')) = LOWER(?)", f.PostalCode)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(COALESCE(label, '') ILIKE ? ESCAPE '\' OR address_line_1 ILIKE ? ESCAPE '\' OR COALESCE(address_line_2, '') ILIKE ? ESCAPE '\' OR city ILIKE ? ESCAPE '\' OR COALESCE(province, '') ILIKE ? ESCAPE '\' OR COALESCE(postal_code, '') ILIKE ? ESCAPE '\' OR COALESCE(formatted_address, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern, pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "city ASC, address_line_1 ASC, id ASC", addressToDomain)
}

// CheckReferenced memeriksa apakah alamat sedang dirujuk oleh badan hukum aktif, lokasi internal, atau relasi alamat mitra.
func (r *Repository) CheckReferenced(ctx context.Context, id string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}

	var referenced bool
	err = db.Raw(
		`SELECT (
			EXISTS (SELECT 1 FROM legal_entities WHERE registered_address_id = ? AND deleted_at IS NULL)
			OR EXISTS (SELECT 1 FROM locations WHERE address_id = ?)
			OR EXISTS (SELECT 1 FROM party_addresses WHERE address_id = ?)
		)`,
		aid, aid, aid,
	).Scan(&referenced).Error
	if err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return referenced, nil
}

// Insert menyimpan master alamat baru ke tabel addresses.
func (r *Repository) Insert(ctx context.Context, addr *entity.Address) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := addressToModel(addr)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui seluruh atribut master alamat pada tabel addresses.
func (r *Repository) Update(ctx context.Context, addr *entity.Address) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := addressToModel(addr)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&addressModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"label":             model.Label,
			"address_line_1":    model.AddressLine1,
			"address_line_2":    model.AddressLine2,
			"city":              model.City,
			"province":          model.Province,
			"postal_code":       model.PostalCode,
			"country_code":      model.CountryCode,
			"formatted_address": model.FormattedAddress,
			"latitude":          model.Latitude,
			"longitude":         model.Longitude,
			"updated_at":        model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrAddressNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrAddressNotFound
	}
	return nil
}

// DeleteByID menghapus master alamat secara permanen dan memetakan konflik RESTRICT menjadi ErrAddressInUse.
func (r *Repository) DeleteByID(ctx context.Context, id string) error {
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
	res := db.Where("id = ?", aid).Delete(&addressModel{})
	if res.Error != nil {
		return pgstore.MapDeleteError(res.Error, eferrors.ErrAddressNotFound, eferrors.ErrAddressInUse)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrAddressNotFound
	}
	return nil
}
