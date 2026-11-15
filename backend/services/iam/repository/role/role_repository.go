package role

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.RoleRepository = (*RoleRepository)(nil)

// RoleRepository mengimplementasikan kontrak RoleRepository menggunakan PostgreSQL dan GORM.
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository membangun RoleRepository dengan validasi koneksi database fail-closed.
func NewRoleRepository(db *gorm.DB) (*RoleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &RoleRepository{db: db}, nil
}

// SelectByID mengambil peran berdasarkan UUID peran.
func (r *RoleRepository) SelectByID(ctx context.Context, id string) (*entity.Role, error) {
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
	var m roleModel
	if err := db.Where("id = ?", rid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrRoleNotFound)
	}
	return roleToDomain(&m)
}

// SelectByCode mengambil peran berdasarkan kode peran yang telah dinormalisasi.
func (r *RoleRepository) SelectByCode(ctx context.Context, code string) (*entity.Role, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	normCode, err := entity.ValidateEntityCode(code, entity.MaxRoleCodeLength, iamerrors.ErrInvalidRoleCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m roleModel
	if err := db.Where("code = ?", normCode).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrRoleNotFound)
	}
	return roleToDomain(&m)
}

// SelectList menampilkan daftar peran secara berhalaman berdasarkan domain management, status aktif, dan kata kunci.
func (r *RoleRepository) SelectList(ctx context.Context, filter domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Role]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Role]{}, err
	}

	q := db.Model(&roleModel{})
	if f.ManagementDomain != "" {
		q = q.Where("management_domain = ?", f.ManagementDomain)
	}
	if f.ActiveOnly {
		q = q.Where("is_active = ?", true)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(code ILIKE ? ESCAPE '\' OR name ILIKE ? ESCAPE '\' OR COALESCE(description, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "redirect_priority ASC, code ASC, id ASC", roleToDomain)
}

// SelectEffectiveByUserID mengambil seluruh peran aktif dan belum kedaluwarsa yang dimiliki pengguna secara langsung maupun via grup aktif.
func (r *RoleRepository) SelectEffectiveByUserID(ctx context.Context, userID string, now time.Time) ([]entity.Role, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}

	checkTime := pgstore.MustLocalNow(now)
	var rows []roleModel
	if err := db.Raw(`
		SELECT DISTINCT r.id::text AS id, r.code, r.name, r.description, r.management_domain,
		       r.default_redirect_path, r.redirect_priority, r.is_system, r.is_active,
		       r.created_at, r.updated_at
		FROM roles r
		WHERE r.is_active = true
		  AND r.id IN (
			SELECT ur.role_id
			FROM user_roles ur
			WHERE ur.user_id = ? AND (ur.expires_at IS NULL OR ur.expires_at > ?)
			UNION
			SELECT gr.role_id
			FROM group_members gm
			JOIN groups g ON g.id = gm.group_id
			JOIN group_roles gr ON gr.group_id = g.id
			WHERE gm.user_id = ? AND g.is_active = true AND (gr.expires_at IS NULL OR gr.expires_at > ?)
		  )
		ORDER BY r.redirect_priority ASC, r.code ASC
	`, uid, checkTime, uid, checkTime).Scan(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	result := make([]entity.Role, 0, len(rows))
	for i := range rows {
		mapped, mapErr := roleToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, *mapped)
	}
	return result, nil
}

// Insert menyimpan peran baru ke tabel roles.
func (r *RoleRepository) Insert(ctx context.Context, role *entity.Role) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := roleToModel(role)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui metadata peran termasuk rute redirect bawaan.
func (r *RoleRepository) Update(ctx context.Context, role *entity.Role) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := roleToModel(role)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&roleModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"name":                  model.Name,
			"description":           model.Description,
			"management_domain":     model.ManagementDomain,
			"default_redirect_path": model.DefaultRedirectPath,
			"redirect_priority":     model.RedirectPriority,
			"is_active":             model.IsActive,
			"updated_at":            model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrRoleNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrRoleNotFound
	}
	return nil
}

// Delete menghapus peran non-sistem sambil menolak penghapusan peran sistem atau peran yang masih ditugaskan.
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
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

	inUseErr := iamerrors.New(iamerrors.KindConflict, "iam.role.in_use", "role is currently assigned to users or groups")
	return db.Transaction(func(tx *gorm.DB) error {
		var existing roleModel
		if err := tx.Select("id, is_system").Where("id = ?", rid).Take(&existing).Error; err != nil {
			return pgstore.MapError(err, iamerrors.ErrRoleNotFound)
		}
		if existing.IsSystem {
			return iamerrors.ErrSystemRoleImmutable
		}
		res := tx.Where("id = ? AND is_system = ?", rid, false).Delete(&roleModel{})
		if res.Error != nil {
			return pgstore.MapDeleteError(res.Error, iamerrors.ErrRoleNotFound, inUseErr)
		}
		if res.RowsAffected == 0 {
			return iamerrors.ErrRoleNotFound
		}
		return nil
	})
}
