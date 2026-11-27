package location

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

var _ domainrepo.LocationRepository = (*Repository)(nil)

// Repository mengimplementasikan LocationRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun LocationRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil lokasi internal aktif (belum dihapus lunak) berdasarkan UUID.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.Location, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	lid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m locationModel
	if err := db.Where("id = ? AND deleted_at IS NULL", lid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLocationNotFound)
	}
	return locationToDomain(&m)
}

// SelectByIDs mengambil beberapa lokasi internal aktif sekaligus berdasarkan daftar UUID untuk WMS, POS, dan IAM scope.
func (r *Repository) SelectByIDs(ctx context.Context, ids []string) ([]*entity.Location, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cleanIDs, err := pgstore.NormalizeUUIDSlice("ids", ids)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*entity.Location{}, nil
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []locationModel
	if err := db.Where("id IN ? AND deleted_at IS NULL", cleanIDs).
		Order("location_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, locationToDomain)
}

// SelectByCode mengambil lokasi internal aktif berdasarkan location_code yang telah dinormalisasi.
func (r *Repository) SelectByCode(ctx context.Context, locationCode string) (*entity.Location, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(locationCode, entity.MaxCodeLength, eferrors.ErrInvalidLocationCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m locationModel
	if err := db.Where("location_code = ? AND deleted_at IS NULL", code).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLocationNotFound)
	}
	return locationToDomain(&m)
}

// SelectByLegalEntityAndKind mengambil daftar lokasi menurut badan hukum dan jenis lokasi secara berhalaman.
func (r *Repository) SelectByLegalEntityAndKind(ctx context.Context, legalEntityID string, kind entity.LocationKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Location]{}, pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("legal_entity_id", legalEntityID)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	if !kind.IsValid() {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.ErrInvalidLocationKind
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	q := db.Model(&locationModel{}).Where("legal_entity_id = ? AND location_kind = ? AND deleted_at IS NULL", eid, kind.String())
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	return pgstore.Paginate(q, page, "location_code ASC, id ASC", locationToDomain)
}

// SelectByAddressID mengambil lokasi internal aktif yang merujuk ke suatu master alamat secara berhalaman.
func (r *Repository) SelectByAddressID(ctx context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Location]{}, pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("address_id", addressID)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	q := db.Model(&locationModel{}).Where("address_id = ? AND deleted_at IS NULL", aid)
	return pgstore.Paginate(q, page, "location_code ASC, id ASC", locationToDomain)
}

// SelectList menampilkan daftar lokasi internal secara berhalaman berdasarkan badan hukum, induk, alamat, jenis, dan status aktif.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Location]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.LegalEntityID != "" {
		if _, err := pgstore.ValidateUUIDParam("legal_entity_id", f.LegalEntityID); err != nil {
			return domainrepo.PageResult[*entity.Location]{}, err
		}
	}
	if f.ParentLocationID != nil {
		if _, err := pgstore.ValidateUUIDParam("parent_location_id", *f.ParentLocationID); err != nil {
			return domainrepo.PageResult[*entity.Location]{}, err
		}
	}
	if f.AddressID != nil {
		if _, err := pgstore.ValidateUUIDParam("address_id", *f.AddressID); err != nil {
			return domainrepo.PageResult[*entity.Location]{}, err
		}
	}
	if f.Kind != nil && !f.Kind.IsValid() {
		return domainrepo.PageResult[*entity.Location]{}, eferrors.ErrInvalidLocationKind
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}

	q := db.Model(&locationModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.LegalEntityID != "" {
		q = q.Where("legal_entity_id = ?", f.LegalEntityID)
	}
	if f.ParentLocationID != nil {
		q = q.Where("parent_location_id = ?", *f.ParentLocationID)
	}
	if f.AddressID != nil {
		q = q.Where("address_id = ?", *f.AddressID)
	}
	if f.Kind != nil {
		q = q.Where("location_kind = ?", f.Kind.String())
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(location_code ILIKE ? ESCAPE '\' OR name ILIKE ? ESCAPE '\' OR COALESCE(phone, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "location_code ASC, id ASC", locationToDomain)
}

// SelectParentID mengambil parent_location_id dari lokasi aktif untuk pemeriksaan siklus hierarki.
func (r *Repository) SelectParentID(ctx context.Context, id string) (*string, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	lid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var row struct {
		ParentID *string `gorm:"column:parent_location_id"`
	}
	if err := db.Model(&locationModel{}).
		Select("parent_location_id").
		Where("id = ? AND deleted_at IS NULL", lid).
		Take(&row).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLocationNotFound)
	}
	if row.ParentID == nil {
		return nil, nil
	}
	norm := entity.NormalizeUUID(*row.ParentID)
	return &norm, nil
}

// SelectChildrenByParentID mengambil sub-lokasi aktif yang berada langsung di bawah lokasi induk tertentu secara berhalaman.
func (r *Repository) SelectChildrenByParentID(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Location]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("parent_location_id", parentID)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Location]{}, err
	}
	q := db.Model(&locationModel{}).Where("parent_location_id = ? AND deleted_at IS NULL", pid)
	return pgstore.Paginate(q, page, "location_code ASC, id ASC", locationToDomain)
}

// CountActiveDependents menghitung total sub-lokasi dan unit organisasi aktif yang bergantung pada lokasi ini.
func (r *Repository) CountActiveDependents(ctx context.Context, id string) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	lid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return 0, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Raw(
		`SELECT
			(SELECT COUNT(*) FROM locations WHERE parent_location_id = ? AND deleted_at IS NULL) +
			(SELECT COUNT(*) FROM organization_units WHERE primary_location_id = ? AND deleted_at IS NULL)`,
		lid, lid,
	).Scan(&total).Error
	if err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return total, nil
}

// CheckCodeExists memeriksa apakah location_code sudah digunakan oleh lokasi aktif lain.
func (r *Repository) CheckCodeExists(ctx context.Context, locationCode, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(locationCode, entity.MaxCodeLength, eferrors.ErrInvalidLocationCode)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	q := db.Model(&locationModel{}).Where("location_code = ? AND deleted_at IS NULL", code)
	if strings.TrimSpace(excludeID) != "" {
		exID, err := pgstore.ValidateUUIDParam("exclude_id", excludeID)
		if err != nil {
			return false, err
		}
		q = q.Where("id <> ?", exID)
	}
	var count int64
	if err := q.Limit(1).Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menyimpan lokasi internal baru ke tabel locations.
func (r *Repository) Insert(ctx context.Context, loc *entity.Location) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := locationToModel(loc)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui atribut profil, alamat, atribut operasional, dan status aktif lokasi yang belum dihapus lunak.
func (r *Repository) Update(ctx context.Context, loc *entity.Location) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := locationToModel(loc)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&locationModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"parent_location_id":     model.ParentLocationID,
			"address_id":             model.AddressID,
			"name":                   model.Name,
			"location_kind":          model.Kind,
			"phone":                  model.Phone,
			"timezone":               model.Timezone,
			"operational_attributes": model.OperationalAttributes,
			"is_active":              model.IsActive,
			"updated_at":             model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrLocationNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrLocationNotFound
	}
	return nil
}

// SoftDeleteByID menonaktifkan dan menghapus lunak lokasi internal berdasarkan UUID.
func (r *Repository) SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	lid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(deletedAt)
	res := db.Model(&locationModel{}).
		Where("id = ? AND deleted_at IS NULL", lid).
		Updates(map[string]any{
			"is_active":  false,
			"deleted_at": ts,
			"updated_at": ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrLocationNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrLocationNotFound
	}
	return nil
}
