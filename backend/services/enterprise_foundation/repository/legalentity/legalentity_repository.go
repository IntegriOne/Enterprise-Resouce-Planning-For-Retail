package legalentity

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

var _ domainrepo.LegalEntityRepository = (*Repository)(nil)

// Repository mengimplementasikan LegalEntityRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun LegalEntityRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil badan hukum aktif (belum dihapus lunak) berdasarkan UUID.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.LegalEntity, error) {
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
	var m legalEntityModel
	if err := db.Where("id = ? AND deleted_at IS NULL", lid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLegalEntityNotFound)
	}
	return legalEntityToDomain(&m)
}

// SelectByIDs mengambil beberapa badan hukum aktif sekaligus berdasarkan daftar UUID untuk resolusi scope IAM dan konsolidasi lintas modul.
func (r *Repository) SelectByIDs(ctx context.Context, ids []string) ([]*entity.LegalEntity, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cleanIDs, err := pgstore.NormalizeUUIDSlice("ids", ids)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*entity.LegalEntity{}, nil
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []legalEntityModel
	if err := db.Where("id IN ? AND deleted_at IS NULL", cleanIDs).
		Order("is_primary_operating_entity DESC, entity_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, legalEntityToDomain)
}

// SelectByCode mengambil badan hukum aktif berdasarkan entity_code yang telah dinormalisasi.
func (r *Repository) SelectByCode(ctx context.Context, entityCode string) (*entity.LegalEntity, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(entityCode, entity.MaxCodeLength, eferrors.ErrInvalidLegalEntityCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m legalEntityModel
	if err := db.Where("entity_code = ? AND deleted_at IS NULL", code).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLegalEntityNotFound)
	}
	return legalEntityToDomain(&m)
}

// SelectByTaxIDNumber mengambil badan hukum aktif berdasarkan NPWP/Tax ID unik.
func (r *Repository) SelectByTaxIDNumber(ctx context.Context, taxIDNumber string) (*entity.LegalEntity, error) {
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
	var m legalEntityModel
	if err := db.Where("tax_id_number = ? AND deleted_at IS NULL", taxID).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLegalEntityNotFound)
	}
	return legalEntityToDomain(&m)
}

// SelectByNIB mengambil daftar badan hukum aktif yang memiliki Nomor Induk Berusaha (NIB) tertentu secara berhalaman.
func (r *Repository) SelectByNIB(ctx context.Context, nib string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, pgstore.CheckDB(nil)
	}
	trimmed, err := entity.ValidateRequiredText(nib, entity.MaxCodeLength, eferrors.ErrInvalidInput)
	if err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, eferrors.WithField(eferrors.ErrInvalidInput, "nib", "nib must be non-empty and within length bounds")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	q := db.Model(&legalEntityModel{}).Where("nib = ? AND deleted_at IS NULL", trimmed)
	return pgstore.Paginate(q, page, "entity_code ASC, id ASC", legalEntityToDomain)
}

// SelectPrimaryOperating mengambil badan hukum operasional utama perusahaan (is_primary_operating_entity = true).
func (r *Repository) SelectPrimaryOperating(ctx context.Context) (*entity.LegalEntity, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m legalEntityModel
	if err := db.Where("is_primary_operating_entity = true AND deleted_at IS NULL").Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLegalEntityNotFound)
	}
	return legalEntityToDomain(&m)
}

// SelectList menampilkan daftar badan hukum secara berhalaman berdasarkan kata kunci, induk, status, dan penanda utama.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.Status != nil && !f.Status.IsValid() {
		return domainrepo.PageResult[*entity.LegalEntity]{}, eferrors.ErrInvalidLegalEntityStatus
	}
	if f.ParentLegalEntityID != nil {
		if _, err := pgstore.ValidateUUIDParam("parent_legal_entity_id", *f.ParentLegalEntityID); err != nil {
			return domainrepo.PageResult[*entity.LegalEntity]{}, err
		}
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}

	q := db.Model(&legalEntityModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.ParentLegalEntityID != nil {
		q = q.Where("parent_legal_entity_id = ?", *f.ParentLegalEntityID)
	}
	if f.Status != nil {
		q = q.Where("status = ?", f.Status.String())
	}
	if f.PrimaryOnly != nil {
		q = q.Where("is_primary_operating_entity = ?", *f.PrimaryOnly)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(entity_code ILIKE ? ESCAPE '\' OR legal_name ILIKE ? ESCAPE '\' OR COALESCE(trading_name, '') ILIKE ? ESCAPE '\' OR COALESCE(tax_id_number, '') ILIKE ? ESCAPE '\' OR COALESCE(nib, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "is_primary_operating_entity DESC, entity_code ASC, id ASC", legalEntityToDomain)
}

// SelectParentID mengambil parent_legal_entity_id dari badan hukum aktif untuk deteksi siklus hierarki.
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
		ParentID *string `gorm:"column:parent_legal_entity_id"`
	}
	if err := db.Model(&legalEntityModel{}).
		Select("parent_legal_entity_id").
		Where("id = ? AND deleted_at IS NULL", lid).
		Take(&row).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrLegalEntityNotFound)
	}
	if row.ParentID == nil {
		return nil, nil
	}
	norm := entity.NormalizeUUID(*row.ParentID)
	return &norm, nil
}

// SelectChildrenByParentID mengambil entitas anak langsung yang belum dihapus lunak secara berhalaman.
func (r *Repository) SelectChildrenByParentID(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("parent_legal_entity_id", parentID)
	if err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.LegalEntity]{}, err
	}
	q := db.Model(&legalEntityModel{}).Where("parent_legal_entity_id = ? AND deleted_at IS NULL", pid)
	return pgstore.Paginate(q, page, "entity_code ASC, id ASC", legalEntityToDomain)
}

// CountActiveDependents menghitung total anak badan hukum, lokasi, dan unit organisasi aktif yang bergantung pada badan hukum ini.
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
			(SELECT COUNT(*) FROM legal_entities WHERE parent_legal_entity_id = ? AND deleted_at IS NULL) +
			(SELECT COUNT(*) FROM locations WHERE legal_entity_id = ? AND deleted_at IS NULL) +
			(SELECT COUNT(*) FROM organization_units WHERE legal_entity_id = ? AND deleted_at IS NULL)`,
		lid, lid, lid,
	).Scan(&total).Error
	if err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return total, nil
}

// CheckCodeExists memeriksa keunikan entity_code di antara badan hukum aktif dengan pengecualian ID opsional.
func (r *Repository) CheckCodeExists(ctx context.Context, entityCode, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	code, err := entity.ValidateCode(entityCode, entity.MaxCodeLength, eferrors.ErrInvalidLegalEntityCode)
	if err != nil {
		return false, err
	}
	return r.checkUniqueFieldExists(ctx, "entity_code", code, excludeID)
}

// CheckTaxIDExists memeriksa keunikan tax_id_number di antara badan hukum aktif dengan pengecualian ID opsional.
func (r *Repository) CheckTaxIDExists(ctx context.Context, taxIDNumber, excludeID string) (bool, error) {
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

func (r *Repository) checkUniqueFieldExists(ctx context.Context, column, value, excludeID string) (bool, error) {
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	q := db.Model(&legalEntityModel{}).Where(column+" = ? AND deleted_at IS NULL", value)
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

// Insert menyimpan badan hukum baru ke tabel legal_entities.
func (r *Repository) Insert(ctx context.Context, le *entity.LegalEntity) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := legalEntityToModel(le)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui profil, status, dan atribut legal/pajak badan hukum yang belum dihapus lunak.
func (r *Repository) Update(ctx context.Context, le *entity.LegalEntity) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := legalEntityToModel(le)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&legalEntityModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"parent_legal_entity_id":      model.ParentLegalEntityID,
			"legal_name":                  model.LegalName,
			"trading_name":                model.TradingName,
			"legal_form":                  model.LegalForm,
			"is_primary_operating_entity": model.IsPrimaryOperatingEntity,
			"status":                      model.Status,
			"jurisdiction_country_code":   model.JurisdictionCountryCode,
			"nib":                         model.NIB,
			"tax_id_number":               model.TaxIDNumber,
			"is_pkp":                      model.IsPKP,
			"registered_address_id":       model.RegisteredAddressID,
			"fiscal_year_start_month":     model.FiscalYearStartMonth,
			"incorporation_date":          model.IncorporationDate,
			"dissolution_date":            model.DissolutionDate,
			"updated_at":                  model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrLegalEntityNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrLegalEntityNotFound
	}
	return nil
}

// UpdateClearPrimaryOperatingExcept membersihkan penanda badan hukum operasional utama pada entitas lain secara atomik.
func (r *Repository) UpdateClearPrimaryOperatingExcept(ctx context.Context, keepID string, updatedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(updatedAt)
	q := db.Model(&legalEntityModel{}).Where("deleted_at IS NULL AND is_primary_operating_entity = true")
	if strings.TrimSpace(keepID) != "" {
		kid, err := pgstore.ValidateUUIDParam("keep_id", keepID)
		if err != nil {
			return err
		}
		q = q.Where("id <> ?", kid)
	}
	return pgstore.MapError(q.Updates(map[string]any{
		"is_primary_operating_entity": false,
		"updated_at":                  ts,
	}).Error, nil)
}

// SoftDeleteByID menghapus lunak badan hukum non-utama pada tabel legal_entities.
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
	res := db.Model(&legalEntityModel{}).
		Where("id = ? AND deleted_at IS NULL", lid).
		Updates(map[string]any{
			"deleted_at": ts,
			"updated_at": ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrLegalEntityNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrLegalEntityNotFound
	}
	return nil
}
