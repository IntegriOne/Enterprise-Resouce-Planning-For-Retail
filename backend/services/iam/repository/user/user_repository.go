package user

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.UserRepository = (*UserRepository)(nil)

// UserRepository mengimplementasikan kontrak UserRepository menggunakan PostgreSQL dan GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository membangun UserRepository dengan validasi koneksi database fail-closed.
func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &UserRepository{db: db}, nil
}

// SelectByID mengambil akun pengguna aktif maupun belum terhapus berdasarkan UUID.
func (r *UserRepository) SelectByID(ctx context.Context, id string) (*entity.User, error) {
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
	var m userModel
	if err := db.Where("id = ? AND deleted_at IS NULL", uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrUserNotFound)
	}
	return userToDomain(&m)
}

// SelectList menampilkan daftar pengguna berhalaman berdasarkan filter jenis, status, kata kunci, dan penghapusan lunak.
func (r *UserRepository) SelectList(ctx context.Context, filter domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.User]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	if f.Kind != nil && !f.Kind.IsValid() {
		return domainrepo.PageResult[*entity.User]{}, iamerrors.ErrInvalidUserKind
	}
	if f.Status != nil && !f.Status.IsValid() {
		return domainrepo.PageResult[*entity.User]{}, iamerrors.ErrInvalidUserStatus
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.User]{}, err
	}

	q := db.Model(&userModel{})
	if !f.IncludeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	if f.Kind != nil {
		q = q.Where("user_kind = ?", f.Kind.String())
	}
	if f.Status != nil {
		q = q.Where("status = ?", f.Status.String())
	}
	if f.Query != "" {
		pattern := pgstore.LikePattern(f.Query)
		q = q.Where(
			`(display_name ILIKE ? ESCAPE '\' OR EXISTS (SELECT 1 FROM user_identities ui WHERE ui.user_id = users.id AND ui.value ILIKE ? ESCAPE '\'))`,
			pattern, pattern,
		)
	}

	return pgstore.Paginate(q, f.Page, "created_at DESC, id ASC", userToDomain)
}

// Insert menyimpan baris pengguna baru ke tabel users.
func (r *UserRepository) Insert(ctx context.Context, user *entity.User) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := userToModel(user)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// InsertWithIdentities mencari user_identities untuk email dan username terlebih dahulu sebelum membuat pengguna secara atomik.
func (r *UserRepository) InsertWithIdentities(ctx context.Context, user *entity.User, identities []*entity.UserIdentity) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	uModel, err := userToModel(user)
	if err != nil {
		return err
	}
	emails, usernames, identityModels, err := prepareUserIdentities(uModel.ID, identities)
	if err != nil {
		return err
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		// Wajib memeriksa tabel user_identities berdasarkan email & username sebelum menyentuh tabel users.
		conflicts, findErr := findIdentitiesByEmailAndUsername(tx, emails, usernames)
		if findErr != nil {
			return findErr
		}
		if len(conflicts) > 0 {
			conflict := pickIdentityConflict(conflicts)
			if conflict.Kind == entity.IdentityKindUsername.String() {
				return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, conflict.Kind, "username has been taken")
			}
			if conflict.Kind == entity.IdentityKindEmail.String() {
				return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, conflict.Kind, "email is already registered")
			}
			return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, conflict.Kind, conflict.Kind+" identity already exists")
		}
		if err := tx.Create(&uModel).Error; err != nil {
			return err
		}
		return tx.Create(&identityModels).Error
	})
	return pgstore.MapError(txErr, nil)
}

// Update memperbarui atribut profil dan status pengguna yang belum dihapus lunak.
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := userToModel(user)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	res := db.Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", model.ID).
		Updates(map[string]any{
			"user_kind":     model.UserKind,
			"status":        model.Status,
			"display_name":  model.DisplayName,
			"avatar_url":    model.AvatarURL,
			"locale":        model.Locale,
			"timezone":      model.Timezone,
			"last_login_at": model.LastLoginAt,
			"updated_at":    model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrUserNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrUserNotFound
	}
	return nil
}

// UpdateLastLogin memperbarui cap waktu login terakhir pengguna memakai waktu lokal OS.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string, loginAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	ts := pgstore.MustLocalNow(loginAt)
	res := db.Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", uid).
		Updates(map[string]any{
			"last_login_at": ts,
			"updated_at":    ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrUserNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrUserNotFound
	}
	return nil
}

// SoftDelete menonaktifkan pengguna dan mengisi deleted_at memakai waktu lokal OS secara atomik.
func (r *UserRepository) SoftDelete(ctx context.Context, id string, deletedAt time.Time) error {
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
	res := db.Model(&userModel{}).
		Where("id = ? AND deleted_at IS NULL", uid).
		Updates(map[string]any{
			"status":     entity.UserStatusDisabled.String(),
			"deleted_at": ts,
			"updated_at": ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrUserNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrUserNotFound
	}
	return nil
}

func prepareUserIdentities(userID string, identities []*entity.UserIdentity) ([]string, []string, []userIdentityModel, error) {
	if len(identities) == 0 {
		return nil, nil, nil, iamerrors.NewValidation("identities", iamerrors.Code("iam.user.identities_required"), "at least one email or username identity is required")
	}

	emails := make([]string, 0, len(identities))
	usernames := make([]string, 0, len(identities))
	models := make([]userIdentityModel, 0, len(identities))
	seen := make(map[string]struct{}, len(identities))

	for _, item := range identities {
		if item == nil {
			return nil, nil, nil, iamerrors.NewValidation("identities", iamerrors.Code("iam.identity.nil"), "user identity must not be nil")
		}
		if !strings.EqualFold(strings.TrimSpace(item.UserID), userID) {
			return nil, nil, nil, iamerrors.NewValidation("identities.user_id", iamerrors.Code("iam.identity.user_mismatch"), "identity user_id must match user id")
		}
		model, err := identityToModel(item)
		if err != nil {
			return nil, nil, nil, err
		}
		key := model.Kind + ":" + model.Value
		if _, dup := seen[key]; dup {
			return nil, nil, nil, iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, model.Kind, "duplicate "+model.Kind+" in user identities")
		}
		seen[key] = struct{}{}
		if model.Kind == entity.IdentityKindEmail.String() {
			// Setiap pengguna hanya boleh memiliki maksimal satu identitas email awal saat pembuatan akun.
			if len(emails) > 0 {
				return nil, nil, nil, iamerrors.WithField(iamerrors.ErrIdentityLimitExceeded, "email", "user may only have one email identity")
			}
			emails = append(emails, model.Value)
		} else if model.Kind == entity.IdentityKindUsername.String() {
			// Setiap pengguna hanya boleh memiliki maksimal satu identitas username.
			if len(usernames) > 0 {
				return nil, nil, nil, iamerrors.WithField(iamerrors.ErrIdentityLimitExceeded, "username", "user may only have one username identity")
			}
			usernames = append(usernames, model.Value)
		}
		models = append(models, model)
	}
	return emails, usernames, models, nil
}

func pickIdentityConflict(conflicts []userIdentityModel) userIdentityModel {
	// Utamakan konflik username terlebih dahulu agar username terpakai tidak menjadi oracle eksistensi email.
	for _, c := range conflicts {
		if c.Kind == entity.IdentityKindUsername.String() {
			return c
		}
	}
	return conflicts[0]
}
