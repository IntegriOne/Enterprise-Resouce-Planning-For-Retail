package permission

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.PermissionRepository = (*Repository)(nil)

// Repository mengimplementasikan kontrak PermissionRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun PermissionRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil izin atomik berdasarkan UUID izin.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.Permission, error) {
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
	var m permissionModel
	if err := db.Where("id = ?", pid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrPermissionNotFound)
	}
	return permissionToDomain(&m)
}

// SelectByCode mengambil izin atomik berdasarkan kode terstandardisasi.
func (r *Repository) SelectByCode(ctx context.Context, code string) (*entity.Permission, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	normCode, err := entity.ValidatePermissionCode(code)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m permissionModel
	if err := db.Where("code = ?", normCode).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrPermissionNotFound)
	}
	return permissionToDomain(&m)
}

// SelectList menampilkan daftar izin berhalaman sesuai filter kata kunci, resource, dan sensitivitas.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Permission]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Permission]{}, err
	}

	q := db.Model(&permissionModel{})
	if f.Resource != "" {
		q = q.Where("resource = ?", f.Resource)
	}
	if f.IsSensitive != nil {
		q = q.Where("is_sensitive = ?", *f.IsSensitive)
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(code ILIKE ? ESCAPE '\' OR resource ILIKE ? ESCAPE '\' OR COALESCE(description, '') ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "code ASC, id ASC", permissionToDomain)
}

// SelectByRoleID mengambil seluruh izin atomik yang terpasang pada satu peran tertentu.
func (r *Repository) SelectByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("role_id", roleID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var rows []permissionModel
	if err := db.Table("permissions p").
		Select("p.id, p.code, p.resource, p.action, p.description, p.is_sensitive, p.created_at").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Where("rp.role_id = ?", rid).
		Order("p.code ASC, p.id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	result := make([]*entity.Permission, 0, len(rows))
	for i := range rows {
		mapped, mapErr := permissionToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// Insert menyimpan definisi izin atomik baru ke katalog permissions.
func (r *Repository) Insert(ctx context.Context, permission *entity.Permission) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := permissionToModel(permission)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbaiki metadata katalog izin berdasarkan UUID tanpa mengubah kode dan sensitivitasnya.
func (r *Repository) Update(ctx context.Context, permission *entity.Permission) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := permissionToModel(permission)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&permissionModel{}).Where("id = ? AND code = ?", model.ID, model.Code).
		Updates(map[string]any{"resource": model.Resource, "action": model.Action})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrPermissionNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrPermissionNotFound
	}
	return nil
}

// Delete menghapus definisi izin atomik berdasarkan UUID izin.
func (r *Repository) Delete(ctx context.Context, id string) error {
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
	return pgstore.MapError(db.Transaction(func(tx *gorm.DB) error {
		var existing permissionModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", pid).Take(&existing).Error; err != nil {
			return pgstore.MapError(err, iamerrors.ErrPermissionNotFound)
		}
		var assigned int64
		if err := tx.Table("role_permissions").Where("permission_id = ?", pid).Limit(1).Count(&assigned).Error; err != nil {
			return pgstore.MapError(err, nil)
		}
		if assigned > 0 {
			return iamerrors.ErrPermissionInUse
		}
		if err := tx.Where("id = ?", pid).Delete(&permissionModel{}).Error; err != nil {
			return pgstore.MapError(err, iamerrors.ErrPermissionNotFound)
		}
		return nil
	}), nil)
}
