package group

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.GroupRoleRepository = (*GroupRoleRepository)(nil)

// GroupRoleRepository mengimplementasikan kontrak GroupRoleRepository menggunakan PostgreSQL dan GORM.
type GroupRoleRepository struct {
	db *gorm.DB
}

// NewGroupRoleRepository membangun GroupRoleRepository dengan validasi koneksi database fail-closed.
func NewGroupRoleRepository(db *gorm.DB) (*GroupRoleRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &GroupRoleRepository{db: db}, nil
}

// SelectByGroupID mengambil daftar penugasan peran milik grup dengan opsi menyertakan yang kedaluwarsa.
func (r *GroupRoleRepository) SelectByGroupID(ctx context.Context, groupID string, includeExpired bool, now time.Time) ([]*entity.GroupRole, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}

	q := db.Where("group_id = ?", gid)
	if !includeExpired {
		checkTime := pgstore.MustLocalNow(now)
		q = q.Where("(expires_at IS NULL OR expires_at > ?)", checkTime)
	}

	var rows []groupRoleModel
	if err := q.Order("created_at ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	result := make([]*entity.GroupRole, 0, len(rows))
	for i := range rows {
		mapped, mapErr := groupRoleToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// CheckExists memeriksa apakah peran sudah ditugaskan ke grup pada scope target (termasuk global).
func (r *GroupRoleRepository) CheckExists(ctx context.Context, groupID, roleID string, scope entity.Scope) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
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

	q, err := pgstore.ApplyScopeWhere(db.Model(&groupRoleModel{}).Where("group_id = ? AND role_id = ?", gid, rid), scope)
	if err != nil {
		return false, err
	}
	var count int64
	if err := q.Limit(1).Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menyimpan penugasan peran kolektif grup pada tabel group_roles.
func (r *GroupRoleRepository) Insert(ctx context.Context, assignment *entity.GroupRole) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := groupRoleToModel(assignment)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// DeleteByID mencabut penugasan peran grup berdasarkan UUID penugasan.
func (r *GroupRoleRepository) DeleteByID(ctx context.Context, id string) error {
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
	res := db.Where("id = ?", aid).Delete(&groupRoleModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrGroupRoleNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrGroupRoleNotFound
	}
	return nil
}
