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

var _ domainrepo.UserRoleRepository = (*UserRoleRepository)(nil)

// UserRoleRepository mengimplementasikan kontrak UserRoleRepository menggunakan PostgreSQL dan GORM.
type UserRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository membangun UserRoleRepository dengan validasi koneksi database fail-closed.
func NewUserRoleRepository(db *gorm.DB) (*UserRoleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &UserRoleRepository{db: db}, nil
}

// SelectByUserID mengambil penugasan peran langsung milik pengguna dengan opsi menyertakan yang kedaluwarsa.
func (r *UserRoleRepository) SelectByUserID(ctx context.Context, userID string, includeExpired bool, now time.Time) ([]*entity.UserRole, error) {
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

	q := db.Where("user_id = ?", uid)
	if !includeExpired {
		checkTime := pgstore.MustLocalNow(now)
		q = q.Where("(expires_at IS NULL OR expires_at > ?)", checkTime)
	}

	var rows []userRoleModel
	if err := q.Order("created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	result := make([]*entity.UserRole, 0, len(rows))
	for i := range rows {
		mapped, mapErr := userRoleToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// CheckExists memeriksa apakah pengguna sudah memiliki penugasan peran pada scope target (termasuk global).
func (r *UserRoleRepository) CheckExists(ctx context.Context, userID, roleID string, scope entity.Scope) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return false, err
	}
	rid, err := pgstore.ValidateUUIDParam("role_id", roleID)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}

	q, err := pgstore.ApplyScopeWhere(db.Model(&userRoleModel{}).Where("user_id = ? AND role_id = ?", uid, rid), scope)
	if err != nil {
		return false, err
	}
	var count int64
	if err := q.Limit(1).Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menyimpan penugasan peran langsung pengguna pada tabel user_roles.
func (r *UserRoleRepository) Insert(ctx context.Context, assignment *entity.UserRole) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := userRoleToModel(assignment)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// DeleteByID mencabut penugasan peran langsung pengguna berdasarkan UUID penugasan.
func (r *UserRoleRepository) DeleteByID(ctx context.Context, id string) error {
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
	res := db.Where("id = ?", aid).Delete(&userRoleModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrUserRoleNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrUserRoleNotFound
	}
	return nil
}
