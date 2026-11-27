package party

import (
	"context"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.PartyRoleRepository = (*PartyRoleRepository)(nil)

// PartyRoleRepository mengimplementasikan PartyRoleRepository menggunakan PostgreSQL dan GORM.
type PartyRoleRepository struct {
	db *gorm.DB
}

// NewPartyRoleRepository membangun PartyRoleRepository dengan validasi koneksi database fail-closed.
func NewPartyRoleRepository(db *gorm.DB) (*PartyRoleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PartyRoleRepository{db: db}, nil
}

// SelectByID mengambil penetapan peran mitra berdasarkan UUID.
func (r *PartyRoleRepository) SelectByID(ctx context.Context, id string) (*entity.PartyRole, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyRoleModel
	if err := db.Where("id = ?", rid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyRoleNotFound)
	}
	return partyRoleToDomain(&m)
}

// SelectByPartyID mengambil peran bisnis milik satu mitra secara berhalaman.
func (r *PartyRoleRepository) SelectByPartyID(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	q := db.Model(&partyRoleModel{}).Where("party_id = ?", pid)
	return pgstore.Paginate(q, page, "role ASC, id ASC", partyRoleToDomain)
}

// SelectByPartyAndRole mengambil penetapan peran spesifik milik mitra untuk validasi Procurement, OMS, atau Finance.
func (r *PartyRoleRepository) SelectByPartyAndRole(ctx context.Context, partyID string, role entity.PartyRoleKind) (*entity.PartyRole, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return nil, err
	}
	if !role.IsValid() {
		return nil, eferrors.ErrInvalidPartyRoleKind
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyRoleModel
	if err := db.Where("party_id = ? AND role = ?", pid, role.String()).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyRoleNotFound)
	}
	return partyRoleToDomain(&m)
}

// SelectByRoleKind mengambil daftar penetapan peran mitra berdasarkan jenis peran bisnis secara berhalaman.
func (r *PartyRoleRepository) SelectByRoleKind(ctx context.Context, role entity.PartyRoleKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, pgstore.CheckDB(nil)
	}
	if !role.IsValid() {
		return domainrepo.PageResult[*entity.PartyRole]{}, eferrors.ErrInvalidPartyRoleKind
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}
	q := db.Model(&partyRoleModel{}).Where("role = ?", role.String())
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	return pgstore.Paginate(q, page, "created_at DESC, id ASC", partyRoleToDomain)
}

// SelectList menampilkan daftar penetapan peran mitra secara berhalaman berdasarkan mitra, jenis peran, kata kunci catatan, dan keaktifan.
func (r *PartyRoleRepository) SelectList(ctx context.Context, filter domainrepo.PartyRoleFilter) (domainrepo.PageResult[*entity.PartyRole], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.PartyID != "" {
		if _, err := pgstore.ValidateUUIDParam("party_id", f.PartyID); err != nil {
			return domainrepo.PageResult[*entity.PartyRole]{}, err
		}
	}
	if f.Role != nil && !f.Role.IsValid() {
		return domainrepo.PageResult[*entity.PartyRole]{}, eferrors.ErrInvalidPartyRoleKind
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyRole]{}, err
	}

	q := db.Model(&partyRoleModel{})
	if f.PartyID != "" {
		q = q.Where("party_id = ?", f.PartyID)
	}
	if f.Role != nil {
		q = q.Where("role = ?", f.Role.String())
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(`COALESCE(reference_note, '') ILIKE ? ESCAPE '\'`, pattern)
	}

	return pgstore.Paginate(q, f.Page, "created_at DESC, id ASC", partyRoleToDomain)
}

// CheckExists memeriksa apakah pasangan mitra dan peran bisnis sudah terdaftar di tabel party_roles.
func (r *PartyRoleRepository) CheckExists(ctx context.Context, partyID string, role entity.PartyRoleKind) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return false, err
	}
	if !role.IsValid() {
		return false, eferrors.ErrInvalidPartyRoleKind
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	var count int64
	if err := db.Model(&partyRoleModel{}).
		Where("party_id = ? AND role = ?", pid, role.String()).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menyimpan penetapan peran mitra baru ke tabel party_roles.
func (r *PartyRoleRepository) Insert(ctx context.Context, role *entity.PartyRole) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyRoleToModel(role)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui status aktif dan catatan referensi peran mitra.
func (r *PartyRoleRepository) Update(ctx context.Context, role *entity.PartyRole) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyRoleToModel(role)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&partyRoleModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"is_active":      model.IsActive,
			"reference_note": model.ReferenceNote,
			"updated_at":     model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyRoleNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyRoleNotFound
	}
	return nil
}

// DeleteByID menghapus penetapan peran mitra berdasarkan UUID.
func (r *PartyRoleRepository) DeleteByID(ctx context.Context, id string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Where("id = ?", rid).Delete(&partyRoleModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyRoleNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyRoleNotFound
	}
	return nil
}
