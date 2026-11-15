package group

import (
	"context"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.GroupMemberRepository = (*MemberRepository)(nil)

// MemberRepository mengimplementasikan kontrak GroupMemberRepository menggunakan PostgreSQL dan GORM.
type MemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository membangun MemberRepository dengan validasi koneksi database fail-closed.
func NewMemberRepository(db *gorm.DB) (*MemberRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &MemberRepository{db: db}, nil
}

// SelectUsersByGroupID menampilkan anggota aktif grup secara berhalaman.
func (r *MemberRepository) SelectUsersByGroupID(ctx context.Context, groupID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.User], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.User]{}, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
	if err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}

	q := db.Table("users u").
		Select("u.id, u.user_kind, u.status, u.display_name, u.avatar_url, u.locale, u.timezone, u.last_login_at, u.created_at, u.updated_at, u.deleted_at").
		Joins("JOIN group_members gm ON gm.user_id = u.id").
		Where("gm.group_id = ? AND u.deleted_at IS NULL", gid)

	return pgstore.Paginate(q, page, "u.display_name ASC, u.id ASC", groupUserToDomain)
}

// SelectGroupsByUserID mengambil seluruh grup yang diikuti oleh seorang pengguna.
func (r *MemberRepository) SelectGroupsByUserID(ctx context.Context, userID string) ([]*entity.Group, error) {
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

	var rows []groupModel
	if err := db.Table("groups g").
		Select("g.id, g.code, g.name, g.description, g.is_active, g.created_at, g.updated_at").
		Joins("JOIN group_members gm ON gm.group_id = g.id").
		Where("gm.user_id = ?", uid).
		Order("g.code ASC, g.id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	result := make([]*entity.Group, 0, len(rows))
	for i := range rows {
		mapped, mapErr := groupToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// CheckExists memeriksa apakah pengguna sudah terdaftar sebagai anggota grup.
func (r *MemberRepository) CheckExists(ctx context.Context, groupID, userID string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
	if err != nil {
		return false, err
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}

	var count int64
	if err := db.Model(&groupMemberModel{}).
		Where("group_id = ? AND user_id = ?", gid, uid).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menambahkan pengguna ke dalam grup tanpa mengisi kolom id agar dihasilkan otomatis oleh PostgreSQL.
func (r *MemberRepository) Insert(ctx context.Context, member *entity.GroupMember) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := groupMemberToModel(member)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Omit("id").Create(&model).Error, nil)
}

// Delete menghapus keanggotaan pengguna dari suatu grup.
func (r *MemberRepository) Delete(ctx context.Context, groupID, userID string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
	if err != nil {
		return err
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Where("group_id = ? AND user_id = ?", gid, uid).Delete(&groupMemberModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrGroupMemberNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrGroupMemberNotFound
	}
	return nil
}
