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

var _ domainrepo.PartyContactRepository = (*PartyContactRepository)(nil)

// PartyContactRepository mengimplementasikan PartyContactRepository menggunakan PostgreSQL dan GORM.
type PartyContactRepository struct {
	db *gorm.DB
}

// NewPartyContactRepository membangun PartyContactRepository dengan validasi koneksi database fail-closed.
func NewPartyContactRepository(db *gorm.DB) (*PartyContactRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PartyContactRepository{db: db}, nil
}

// SelectByID mengambil kontak/PIC mitra aktif (belum dihapus lunak) berdasarkan UUID.
func (r *PartyContactRepository) SelectByID(ctx context.Context, id string) (*entity.PartyContact, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyContactModel
	if err := db.Where("id = ? AND deleted_at IS NULL", cid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyContactNotFound)
	}
	return partyContactToDomain(&m)
}

// SelectByPartyID mengambil kontak/PIC aktif milik satu mitra secara berhalaman dengan urutan kontak utama lebih dahulu.
func (r *PartyContactRepository) SelectByPartyID(ctx context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	q := db.Model(&partyContactModel{}).Where("party_id = ? AND deleted_at IS NULL", pid)
	return pgstore.Paginate(q, page, "is_primary DESC, full_name ASC, id ASC", partyContactToDomain)
}

// SelectPrimaryByPartyID mengambil PIC utama aktif dari suatu mitra bisnis.
func (r *PartyContactRepository) SelectPrimaryByPartyID(ctx context.Context, partyID string) (*entity.PartyContact, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyContactModel
	if err := db.Where("party_id = ? AND is_primary = true AND deleted_at IS NULL", pid).
		Order("is_active DESC, created_at ASC, id ASC").
		Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyContactNotFound)
	}
	return partyContactToDomain(&m)
}

// SelectByIAMUserID mengambil kontak/PIC aktif yang tertaut ke principal IAM tertentu secara berhalaman untuk resolusi scope portal.
func (r *PartyContactRepository) SelectByIAMUserID(ctx context.Context, iamUserID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("iam_user_id", iamUserID)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	q := db.Model(&partyContactModel{}).Where("iam_user_id = ? AND deleted_at IS NULL", uid)
	return pgstore.Paginate(q, page, "is_primary DESC, created_at ASC, id ASC", partyContactToDomain)
}

// SelectByPartyAndIAMUserID mengambil baris PIC untuk pasangan mitra dan principal IAM tertentu.
func (r *PartyContactRepository) SelectByPartyAndIAMUserID(ctx context.Context, partyID, iamUserID string) (*entity.PartyContact, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return nil, err
	}
	uid, err := pgstore.ValidateUUIDParam("iam_user_id", iamUserID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyContactModel
	if err := db.Where("party_id = ? AND iam_user_id = ? AND deleted_at IS NULL", pid, uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyContactNotFound)
	}
	return partyContactToDomain(&m)
}

// SelectByEmail mengambil daftar kontak/PIC mitra aktif berdasarkan alamat email secara berhalaman (case-insensitive).
func (r *PartyContactRepository) SelectByEmail(ctx context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, pgstore.CheckDB(nil)
	}
	normEmail, err := entity.ValidateEmail(email)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}
	q := db.Model(&partyContactModel{}).Where("LOWER(COALESCE(email, '')) = ? AND deleted_at IS NULL", normEmail)
	return pgstore.Paginate(q, page, "is_primary DESC, full_name ASC, id ASC", partyContactToDomain)
}

// SelectList menampilkan daftar kontak/PIC mitra secara berhalaman berdasarkan mitra, kata kunci, peran, dan status portal IAM.
func (r *PartyContactRepository) SelectList(ctx context.Context, filter domainrepo.PartyContactFilter) (domainrepo.PageResult[*entity.PartyContact], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.PartyID != "" {
		if _, err := pgstore.ValidateUUIDParam("party_id", f.PartyID); err != nil {
			return domainrepo.PageResult[*entity.PartyContact]{}, err
		}
	}
	if f.PortalAccessStatus != nil && !f.PortalAccessStatus.IsValid() {
		return domainrepo.PageResult[*entity.PartyContact]{}, eferrors.ErrInvalidPortalAccessStatus
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.PartyContact]{}, err
	}

	q := db.Model(&partyContactModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.PartyID != "" {
		q = q.Where("party_id = ?", f.PartyID)
	}
	if f.ContactRole != "" {
		q = q.Where("contact_role = ?", f.ContactRole)
	}
	if f.PortalAccessStatus != nil {
		q = q.Where("portal_access_status = ?", f.PortalAccessStatus.String())
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(full_name ILIKE ? ESCAPE '\' OR COALESCE(job_title, '') ILIKE ? ESCAPE '\' OR COALESCE(email, '') ILIKE ? ESCAPE '\' OR COALESCE(phone, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "is_primary DESC, full_name ASC, id ASC", partyContactToDomain)
}

// CheckIAMUserExistsInParty memeriksa apakah principal IAM sudah tertaut ke kontak aktif lain pada mitra yang sama.
func (r *PartyContactRepository) CheckIAMUserExistsInParty(ctx context.Context, partyID, iamUserID, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return false, err
	}
	uid, err := pgstore.ValidateUUIDParam("iam_user_id", iamUserID)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	q := db.Model(&partyContactModel{}).
		Where("party_id = ? AND iam_user_id = ? AND deleted_at IS NULL", pid, uid)
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

// UpdateClearPrimaryByPartyExcept menghapus penanda is_primary pada kontak mitra selain ID yang dipertahankan.
func (r *PartyContactRepository) UpdateClearPrimaryByPartyExcept(ctx context.Context, partyID, keepID string, updatedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("party_id", partyID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(updatedAt)
	q := db.Model(&partyContactModel{}).Where("party_id = ? AND deleted_at IS NULL AND is_primary = true", pid)
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

// Insert menyimpan kontak/PIC mitra baru ke tabel party_contacts.
func (r *PartyContactRepository) Insert(ctx context.Context, contact *entity.PartyContact) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyContactToModel(contact)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui profil kontak, tautan IAM, status portal, dan status aktif PIC yang belum dihapus lunak.
func (r *PartyContactRepository) Update(ctx context.Context, contact *entity.PartyContact) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyContactToModel(contact)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&partyContactModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"iam_user_id":          model.IAMUserID,
			"full_name":            model.FullName,
			"job_title":            model.JobTitle,
			"contact_role":         model.ContactRole,
			"email":                model.Email,
			"phone":                model.Phone,
			"portal_access_status": model.PortalAccessStatus,
			"is_primary":           model.IsPrimary,
			"is_active":            model.IsActive,
			"updated_at":           model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyContactNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyContactNotFound
	}
	return nil
}

// SoftDeleteByID menghapus lunak kontak/PIC mitra sekaligus menonaktifkan dan mencabut status akses portal jika tertaut ke IAM.
func (r *PartyContactRepository) SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	cid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(deletedAt)
	res := db.Model(&partyContactModel{}).
		Where("id = ? AND deleted_at IS NULL", cid).
		Updates(map[string]any{
			"is_active":  false,
			"deleted_at": ts,
			"updated_at": ts,
			"portal_access_status": gorm.Expr(
				"CASE WHEN iam_user_id IS NOT NULL THEN 'revoked'::portal_access_status ELSE portal_access_status END",
			),
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyContactNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyContactNotFound
	}
	return nil
}
