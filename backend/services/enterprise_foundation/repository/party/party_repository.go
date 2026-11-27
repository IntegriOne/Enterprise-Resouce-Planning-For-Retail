package party

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.PartyRepository = (*PartyRepository)(nil)

// PartyRepository mengimplementasikan PartyRepository menggunakan PostgreSQL dan GORM.
type PartyRepository struct {
	db *gorm.DB
}

// NewPartyRepository membangun PartyRepository dengan validasi koneksi database fail-closed.
func NewPartyRepository(db *gorm.DB) (*PartyRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &PartyRepository{db: db}, nil
}

// SelectByID mengambil mitra eksternal aktif (belum dihapus lunak) berdasarkan UUID.
func (r *PartyRepository) SelectByID(ctx context.Context, id string) (*entity.Party, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyModel
	if err := db.Where("id = ? AND deleted_at IS NULL", pid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyNotFound)
	}
	return partyToDomain(&m)
}

// SelectByIDs mengambil beberapa mitra eksternal aktif sekaligus berdasarkan daftar UUID untuk PO, AP/AR, dan portal IAM.
func (r *PartyRepository) SelectByIDs(ctx context.Context, ids []string) ([]*entity.Party, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cleanIDs, err := pgstore.NormalizeUUIDSlice("ids", ids)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*entity.Party{}, nil
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []partyModel
	if err := db.Where("id IN ? AND deleted_at IS NULL", cleanIDs).
		Order("partner_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, partyToDomain)
}

// SelectByCode mengambil mitra eksternal aktif berdasarkan partner_code yang telah dinormalisasi.
func (r *PartyRepository) SelectByCode(ctx context.Context, partnerCode string) (*entity.Party, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(partnerCode, entity.MaxCodeLength, eferrors.ErrInvalidPartyCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyModel
	if err := db.Where("partner_code = ? AND deleted_at IS NULL", code).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyNotFound)
	}
	return partyToDomain(&m)
}

// SelectByTaxIDNumber mengambil mitra eksternal aktif berdasarkan NPWP/Tax ID unik.
func (r *PartyRepository) SelectByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.Party, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	taxID, err := entity.ValidateRequiredText(taxIDNumber, entity.MaxTaxIDLength, eferrors.ErrInvalidInput)
	if err != nil {
		return nil, eferrors.WithField(eferrors.ErrInvalidInput, "tax_id_number", "tax id number must be non-empty and within length bounds")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m partyModel
	if err := db.Where("tax_id_number = ? AND deleted_at IS NULL", taxID).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrPartyNotFound)
	}
	return partyToDomain(&m)
}

// SelectByEmail mengambil daftar mitra eksternal aktif berdasarkan email resmi perusahaan secara berhalaman (case-insensitive).
func (r *PartyRepository) SelectByEmail(ctx context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Party]{}, pgstore.CheckDB(nil)
	}
	normEmail, err := entity.ValidateEmail(email)
	if err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}
	q := db.Model(&partyModel{}).Where("LOWER(COALESCE(email, '')) = ? AND deleted_at IS NULL", normEmail)
	return pgstore.Paginate(q, page, "partner_code ASC, id ASC", partyToDomain)
}

// SelectList menampilkan daftar mitra eksternal berhalaman berdasarkan kata kunci, jenis entitas, status, dan peran aktif.
func (r *PartyRepository) SelectList(ctx context.Context, filter domainrepo.PartyFilter) (domainrepo.PageResult[*entity.Party], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Party]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.Kind != nil && !f.Kind.IsValid() {
		return domainrepo.PageResult[*entity.Party]{}, eferrors.ErrInvalidPartyKind
	}
	if f.Status != nil && !f.Status.IsValid() {
		return domainrepo.PageResult[*entity.Party]{}, eferrors.ErrInvalidPartyStatus
	}
	if f.Role != nil && !f.Role.IsValid() {
		return domainrepo.PageResult[*entity.Party]{}, eferrors.ErrInvalidPartyRoleKind
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Party]{}, err
	}

	q := db.Model(&partyModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.Kind != nil {
		q = q.Where("party_kind = ?", f.Kind.String())
	}
	if f.Status != nil {
		q = q.Where("status = ?", f.Status.String())
	}
	if f.Role != nil {
		q = q.Where(
			"EXISTS (SELECT 1 FROM party_roles pr WHERE pr.party_id = parties.id AND pr.role = ? AND pr.is_active = true)",
			f.Role.String(),
		)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(partner_code ILIKE ? ESCAPE '\' OR display_name ILIKE ? ESCAPE '\' OR COALESCE(legal_name, '') ILIKE ? ESCAPE '\' OR COALESCE(email, '') ILIKE ? ESCAPE '\' OR COALESCE(phone, '') ILIKE ? ESCAPE '\' OR COALESCE(tax_id_number, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "display_name ASC, partner_code ASC, id ASC", partyToDomain)
}

// CheckCodeExists memeriksa keunikan partner_code di antara mitra aktif dengan pengecualian ID opsional.
func (r *PartyRepository) CheckCodeExists(ctx context.Context, partnerCode, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(partnerCode, entity.MaxCodeLength, eferrors.ErrInvalidPartyCode)
	if err != nil {
		return false, err
	}
	return r.checkUniqueFieldExists(ctx, "partner_code", code, excludeID)
}

// CheckTaxIDExists memeriksa keunikan tax_id_number di antara mitra aktif dengan pengecualian ID opsional.
func (r *PartyRepository) CheckTaxIDExists(ctx context.Context, taxIDNumber, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	trimmed := strings.TrimSpace(taxIDNumber)
	if trimmed == "" {
		return false, nil
	}
	if utf8.RuneCountInString(trimmed) > entity.MaxTaxIDLength || entity.HasControlChar(trimmed) {
		return false, eferrors.WithField(eferrors.ErrInvalidInput, "tax_id_number", "tax id number exceeds length bounds or contains control characters")
	}
	return r.checkUniqueFieldExists(ctx, "tax_id_number", trimmed, excludeID)
}

func (r *PartyRepository) checkUniqueFieldExists(ctx context.Context, column, value, excludeID string) (bool, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	q := db.Model(&partyModel{}).Where(column+" = ? AND deleted_at IS NULL", value)
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

// Insert menyimpan mitra eksternal baru ke tabel parties.
func (r *PartyRepository) Insert(ctx context.Context, p *entity.Party) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyToModel(p)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui profil, data pajak, dan status mitra eksternal yang belum dihapus lunak.
func (r *PartyRepository) Update(ctx context.Context, p *entity.Party) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := partyToModel(p)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&partyModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"party_kind":           model.Kind,
			"display_name":         model.DisplayName,
			"legal_name":           model.LegalName,
			"email":                model.Email,
			"phone":                model.Phone,
			"website_url":          model.WebsiteURL,
			"tax_id_number":        model.TaxIDNumber,
			"tax_registered_name":  model.TaxRegisteredName,
			"is_pkp":               model.IsPKP,
			"is_tax_exempt":        model.IsTaxExempt,
			"tax_exemption_reason": model.TaxExemptionReason,
			"tax_verified_at":      model.TaxVerifiedAt,
			"preferred_language":   model.PreferredLanguage,
			"status":               model.Status,
			"notes":                model.Notes,
			"updated_at":           model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyNotFound
	}
	return nil
}

// SoftDeleteByID menghapus lunak mitra dan menurunkan statusnya ke inactive.
func (r *PartyRepository) SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(deletedAt)
	res := db.Model(&partyModel{}).
		Where("id = ? AND deleted_at IS NULL", pid).
		Updates(map[string]any{
			"status":     entity.PartyStatusInactive.String(),
			"deleted_at": ts,
			"updated_at": ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrPartyNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrPartyNotFound
	}
	return nil
}
