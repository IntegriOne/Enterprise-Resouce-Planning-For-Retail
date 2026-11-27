package organizationunit

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.OrganizationUnitRepository = (*Repository)(nil)

// Repository mengimplementasikan OrganizationUnitRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun OrganizationUnitRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil unit organisasi aktif (belum dihapus lunak) berdasarkan UUID.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return r.selectByID(ctx, id, false)
}

// SelectByIDForUpdate membaca unit dan mengunci baris di dalam transaksi service.
func (r *Repository) SelectByIDForUpdate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return r.selectByID(ctx, id, true)
}

func (r *Repository) selectByID(ctx context.Context, id string, lock bool) (*entity.OrganizationUnit, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var m organizationUnitModel
	if err := db.Where("id = ? AND deleted_at IS NULL", uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrOrgUnitNotFound)
	}
	return organizationUnitToDomain(&m)
}

// SelectByIDs mengambil beberapa unit organisasi aktif sekaligus berdasarkan daftar UUID untuk HRM dan IAM scope.
func (r *Repository) SelectByIDs(ctx context.Context, ids []string) ([]*entity.OrganizationUnit, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	cleanIDs, err := pgstore.NormalizeUUIDSlice("ids", ids)
	if err != nil {
		return nil, err
	}
	if len(cleanIDs) == 0 {
		return []*entity.OrganizationUnit{}, nil
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []organizationUnitModel
	if err := db.Where("id IN ? AND deleted_at IS NULL", cleanIDs).
		Order("unit_code ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, organizationUnitToDomain)
}

// SelectByEntityAndCode mengambil unit organisasi aktif berdasarkan pasangan legal_entity_id dan unit_code.
func (r *Repository) SelectByEntityAndCode(ctx context.Context, legalEntityID, unitCode string) (*entity.OrganizationUnit, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("legal_entity_id", legalEntityID)
	if err != nil {
		return nil, err
	}
	code, err := entity.ValidateCode(unitCode, entity.MaxCodeLength, eferrors.ErrInvalidOrgUnitCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m organizationUnitModel
	if err := db.Where("legal_entity_id = ? AND unit_code = ? AND deleted_at IS NULL", eid, code).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrOrgUnitNotFound)
	}
	return organizationUnitToDomain(&m)
}

// SelectByHeadPersonID mengambil unit organisasi aktif yang dipimpin oleh worker_id HRM tertentu secara berhalaman.
func (r *Repository) SelectByHeadPersonID(ctx context.Context, headPersonID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, pgstore.CheckDB(nil)
	}
	hid, err := pgstore.ValidateUUIDParam("head_person_id", headPersonID)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	q := db.Model(&organizationUnitModel{}).Where("head_person_id = ? AND deleted_at IS NULL", hid)
	return pgstore.Paginate(q, page, "unit_code ASC, id ASC", organizationUnitToDomain)
}

// SelectByPrimaryLocationID mengambil unit organisasi aktif yang berkedudukan utama pada suatu lokasi fisik secara berhalaman.
func (r *Repository) SelectByPrimaryLocationID(ctx context.Context, primaryLocationID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, pgstore.CheckDB(nil)
	}
	lid, err := pgstore.ValidateUUIDParam("primary_location_id", primaryLocationID)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	q := db.Model(&organizationUnitModel{}).Where("primary_location_id = ? AND deleted_at IS NULL", lid)
	return pgstore.Paginate(q, page, "unit_code ASC, id ASC", organizationUnitToDomain)
}

// SelectList menampilkan daftar unit organisasi secara berhalaman berdasarkan badan hukum, induk, lokasi utama, jenis, dan status aktif.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.OrganizationUnitFilter) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.LegalEntityID != "" {
		if _, err := pgstore.ValidateUUIDParam("legal_entity_id", f.LegalEntityID); err != nil {
			return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
		}
	}
	if f.ParentUnitID != nil {
		if _, err := pgstore.ValidateUUIDParam("parent_unit_id", *f.ParentUnitID); err != nil {
			return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
		}
	}
	if f.PrimaryLocationID != nil {
		if _, err := pgstore.ValidateUUIDParam("primary_location_id", *f.PrimaryLocationID); err != nil {
			return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
		}
	}
	if f.Kind != nil && !f.Kind.IsValid() {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, eferrors.ErrInvalidOrgUnitKind
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}

	q := db.Model(&organizationUnitModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.LegalEntityID != "" {
		q = q.Where("legal_entity_id = ?", f.LegalEntityID)
	}
	if f.ParentUnitID != nil {
		q = q.Where("parent_unit_id = ?", *f.ParentUnitID)
	}
	if f.PrimaryLocationID != nil {
		q = q.Where("primary_location_id = ?", *f.PrimaryLocationID)
	}
	if f.Kind != nil {
		q = q.Where("unit_kind = ?", f.Kind.String())
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(unit_code ILIKE ? ESCAPE '\' OR name ILIKE ? ESCAPE '\' OR COALESCE(cost_center_code, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "unit_code ASC, id ASC", organizationUnitToDomain)
}

// SelectParentID mengambil parent_unit_id dari unit organisasi aktif untuk pemeriksaan siklus hierarki.
func (r *Repository) SelectParentID(ctx context.Context, id string) (*string, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var row struct {
		ParentID *string `gorm:"column:parent_unit_id"`
	}
	if err := db.Model(&organizationUnitModel{}).
		Select("parent_unit_id").
		Where("id = ? AND deleted_at IS NULL", uid).
		Take(&row).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrOrgUnitNotFound)
	}
	if row.ParentID == nil {
		return nil, nil
	}
	norm := entity.NormalizeUUID(*row.ParentID)
	return &norm, nil
}

// SelectChildrenByParentID mengambil unit organisasi anak langsung yang belum dihapus lunak secara berhalaman.
func (r *Repository) SelectChildrenByParentID(ctx context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, pgstore.CheckDB(nil)
	}
	pid, err := pgstore.ValidateUUIDParam("parent_unit_id", parentID)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.OrganizationUnit]{}, err
	}
	q := db.Model(&organizationUnitModel{}).Where("parent_unit_id = ? AND deleted_at IS NULL", pid)
	return pgstore.Paginate(q, page, "unit_code ASC, id ASC", organizationUnitToDomain)
}

// CountActiveDependents menghitung jumlah sub-unit organisasi aktif yang masih menunjuk unit ini sebagai induk.
func (r *Repository) CountActiveDependents(ctx context.Context, id string) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return 0, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := db.Model(&organizationUnitModel{}).
		Where("parent_unit_id = ? AND deleted_at IS NULL", uid).
		Count(&total).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return total, nil
}

// CheckCodeExistsInEntity memeriksa keunikan unit_code dalam satu badan hukum di antara unit yang belum dihapus lunak.
func (r *Repository) CheckCodeExistsInEntity(ctx context.Context, legalEntityID, unitCode, excludeID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("legal_entity_id", legalEntityID)
	if err != nil {
		return false, err
	}
	code, err := entity.ValidateCode(unitCode, entity.MaxCodeLength, eferrors.ErrInvalidOrgUnitCode)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	q := db.Model(&organizationUnitModel{}).
		Where("legal_entity_id = ? AND unit_code = ? AND deleted_at IS NULL", eid, code)
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

// Insert menyimpan unit organisasi baru ke tabel organization_units.
func (r *Repository) Insert(ctx context.Context, unit *entity.OrganizationUnit) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := organizationUnitToModel(unit)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui profil, hierarki, penanggung jawab, dan status aktif unit organisasi yang belum dihapus lunak.
func (r *Repository) Update(ctx context.Context, unit *entity.OrganizationUnit) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := organizationUnitToModel(unit)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&organizationUnitModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"parent_unit_id":      model.ParentUnitID,
			"primary_location_id": model.PrimaryLocationID,
			"name":                model.Name,
			"unit_kind":           model.Kind,
			"cost_center_code":    model.CostCenterCode,
			"head_person_id":      model.HeadPersonID,
			"is_active":           model.IsActive,
			"updated_at":          model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrOrgUnitNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrOrgUnitNotFound
	}
	return nil
}

// SoftDeleteByID menonaktifkan dan menghapus lunak unit organisasi berdasarkan UUID.
func (r *Repository) SoftDeleteByID(ctx context.Context, id string, deletedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(deletedAt)
	res := db.Model(&organizationUnitModel{}).
		Where("id = ? AND deleted_at IS NULL", uid).
		Updates(map[string]any{
			"is_active":  false,
			"deleted_at": ts,
			"updated_at": ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrOrgUnitNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrOrgUnitNotFound
	}
	return nil
}
