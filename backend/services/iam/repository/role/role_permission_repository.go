package role

import (
	"context"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.RolePermissionRepository = (*RolePermissionRepository)(nil)

// RolePermissionRepository mengimplementasikan kontrak RolePermissionRepository menggunakan PostgreSQL dan GORM.
type RolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository membangun RolePermissionRepository dengan validasi koneksi database fail-closed.
func NewRolePermissionRepository(db *gorm.DB) (*RolePermissionRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &RolePermissionRepository{db: db}, nil
}

// CheckExists memeriksa apakah pasangan peran dan izin sudah terikat di tabel role_permissions.
func (r *RolePermissionRepository) CheckExists(ctx context.Context, roleID, permissionID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("role_id", roleID)
	if err != nil {
		return false, err
	}
	pid, err := pgstore.ValidateUUIDParam("permission_id", permissionID)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	var count int64
	if err := db.Model(&rolePermissionModel{}).
		Where("role_id = ? AND permission_id = ?", rid, pid).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menautkan izin atomik ke suatu peran.
func (r *RolePermissionRepository) Insert(ctx context.Context, rolePermission *entity.RolePermission) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := rolePermissionToModel(rolePermission)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Delete melepas ikatan izin atomik dari suatu peran.
func (r *RolePermissionRepository) Delete(ctx context.Context, roleID, permissionID string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("role_id", roleID)
	if err != nil {
		return err
	}
	pid, err := pgstore.ValidateUUIDParam("permission_id", permissionID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Where("role_id = ? AND permission_id = ?", rid, pid).Delete(&rolePermissionModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrRolePermissionNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrRolePermissionNotFound
	}
	return nil
}
