package group

import (
	"context"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.GroupRepository = (*GroupRepository)(nil)

// GroupRepository mengimplementasikan kontrak GroupRepository menggunakan PostgreSQL dan GORM.
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository membangun GroupRepository dengan validasi koneksi database fail-closed.
func NewGroupRepository(db *gorm.DB) (*GroupRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &GroupRepository{db: db}, nil
}

// SelectByID mengambil grup pengguna berdasarkan UUID grup.
func (r *GroupRepository) SelectByID(ctx context.Context, id string) (*entity.Group, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m groupModel
	if err := db.Where("id = ?", gid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrGroupNotFound)
	}
	return groupToDomain(&m)
}

// SelectByCode mengambil grup pengguna berdasarkan kode grup yang telah dinormalisasi.
func (r *GroupRepository) SelectByCode(ctx context.Context, code string) (*entity.Group, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	normCode, err := entity.ValidateEntityCode(code, entity.MaxGroupCodeLength, iamerrors.ErrInvalidGroupCode)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m groupModel
	if err := db.Where("code = ?", normCode).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrGroupNotFound)
	}
	return groupToDomain(&m)
}

// SelectList menampilkan daftar grup secara berhalaman berdasarkan status aktif dan kata kunci pencarian.
func (r *GroupRepository) SelectList(ctx context.Context, filter domainrepo.GroupFilter) (domainrepo.PageResult[*entity.Group], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Group]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Group]{}, err
	}

	q := db.Model(&groupModel{})
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

	return pgstore.Paginate(q, f.Page, "code ASC, id ASC", groupToDomain)
}

// Insert menyimpan grup baru ke tabel groups.
func (r *GroupRepository) Insert(ctx context.Context, group *entity.Group) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := groupToModel(group)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui nama, deskripsi, dan status aktif grup.
func (r *GroupRepository) Update(ctx context.Context, group *entity.Group) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := groupToModel(group)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&groupModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"name":        model.Name,
			"description": model.Description,
			"is_active":   model.IsActive,
			"updated_at":  model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrGroupNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrGroupNotFound
	}
	return nil
}

// Delete menghapus grup pengguna berdasarkan UUID grup.
func (r *GroupRepository) Delete(ctx context.Context, id string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Where("id = ?", gid).Delete(&groupModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrGroupNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrGroupNotFound
	}
	return nil
}
