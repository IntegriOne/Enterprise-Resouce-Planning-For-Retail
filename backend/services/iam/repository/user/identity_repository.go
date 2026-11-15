package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.UserIdentityRepository = (*IdentityRepository)(nil)

// IdentityRepository mengimplementasikan kontrak UserIdentityRepository menggunakan PostgreSQL dan GORM.
type IdentityRepository struct {
	db *gorm.DB
}

// NewIdentityRepository membangun IdentityRepository dengan validasi koneksi database fail-closed.
func NewIdentityRepository(db *gorm.DB) (*IdentityRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &IdentityRepository{db: db}, nil
}

// SelectByID mengambil identitas login pengguna berdasarkan UUID identitas.
func (r *IdentityRepository) SelectByID(ctx context.Context, id string) (*entity.UserIdentity, error) {
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
	var m userIdentityModel
	if err := db.Where("id = ?", uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrIdentityNotFound)
	}
	return identityToDomain(&m)
}

// SelectByKindAndValue mengambil identitas login berdasarkan pasangan jenis (email/username) dan nilai bernormalisasi.
func (r *IdentityRepository) SelectByKindAndValue(ctx context.Context, kind entity.IdentityKind, value string) (*entity.UserIdentity, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	normalized, err := entity.NormalizeIdentityValue(kind, value)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m userIdentityModel
	if err := db.Where("kind = ? AND value = ?", kind.String(), normalized).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrIdentityNotFound)
	}
	return identityToDomain(&m)
}

// FindUserIdentityByEmailKind mengambil identitas login bertipe email setelah normalisasi dan validasi format.
func (r *IdentityRepository) FindUserIdentityByEmailKind(ctx context.Context, email string) (*entity.UserIdentity, error) {
	return r.SelectByKindAndValue(ctx, entity.IdentityKindEmail, email)
}

// FindUserIdentityByUsernameKind mengambil identitas login bertipe username setelah normalisasi dan validasi format.
func (r *IdentityRepository) FindUserIdentityByUsernameKind(ctx context.Context, username string) (*entity.UserIdentity, error) {
	return r.SelectByKindAndValue(ctx, entity.IdentityKindUsername, username)
}

// SelectByEmailAndUsername mencari identitas pengguna yang sudah terdaftar untuk email dan/atau username tertentu guna mencegah duplikasi.
func (r *IdentityRepository) SelectByEmailAndUsername(ctx context.Context, email, username string) ([]*entity.UserIdentity, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	var normEmail, normUsername string
	var err error
	if strings.TrimSpace(email) != "" {
		normEmail, err = entity.ValidateEmail(email)
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(username) != "" {
		normUsername, err = entity.ValidateUsername(username)
		if err != nil {
			return nil, err
		}
	}
	if normEmail == "" && normUsername == "" {
		return []*entity.UserIdentity{}, nil
	}

	var emails, usernames []string
	if normEmail != "" {
		emails = []string{normEmail}
	}
	if normUsername != "" {
		usernames = []string{normUsername}
	}

	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := findIdentitiesByEmailAndUsername(db, emails, usernames)
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	result := make([]*entity.UserIdentity, 0, len(rows))
	for i := range rows {
		mapped, mapErr := identityToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// SelectByUserID mengambil seluruh identitas login yang terhubung ke satu pengguna.
func (r *IdentityRepository) SelectByUserID(ctx context.Context, userID string) ([]*entity.UserIdentity, error) {
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
	var rows []userIdentityModel
	if err := db.Where("user_id = ?", uid).
		Order("is_primary DESC, kind ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	result := make([]*entity.UserIdentity, 0, len(rows))
	for i := range rows {
		mapped, mapErr := identityToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// CheckExistsByKindAndValue memeriksa apakah identitas dengan jenis dan nilai tersebut sudah dipakai.
func (r *IdentityRepository) CheckExistsByKindAndValue(ctx context.Context, kind entity.IdentityKind, value string) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	normalized, err := entity.NormalizeIdentityValue(kind, value)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	var count int64
	if err := db.Model(&userIdentityModel{}).
		Where("kind = ? AND value = ?", kind.String(), normalized).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, pgstore.MapError(err, nil)
	}
	return count > 0, nil
}

// Insert menyimpan identitas login baru setelah memastikan pasangan kind/value belum digunakan.
func (r *IdentityRepository) Insert(ctx context.Context, identity *entity.UserIdentity) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := identityToModel(identity)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Transaction(func(tx *gorm.DB) error {
		var existing userIdentityModel
		findErr := tx.Where("kind = ? AND value = ?", model.Kind, model.Value).Take(&existing).Error
		if findErr == nil {
			return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, model.Kind, model.Kind+" identity already exists")
		}
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		return tx.Create(&model).Error
	}), nil)
}

// Update memperbarui data identitas login sambil menjaga keunikan global pasangan kind/value.
func (r *IdentityRepository) Update(ctx context.Context, identity *entity.UserIdentity) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := identityToModel(identity)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Transaction(func(tx *gorm.DB) error {
		var existing userIdentityModel
		findErr := tx.Where("kind = ? AND value = ? AND id <> ?", model.Kind, model.Value, model.ID).Take(&existing).Error
		if findErr == nil {
			return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, model.Kind, model.Kind+" identity already exists")
		}
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return findErr
		}
		res := tx.Model(&userIdentityModel{}).
			Where("id = ?", model.ID).
			Updates(map[string]any{
				"kind":        model.Kind,
				"value":       model.Value,
				"is_primary":  model.IsPrimary,
				"is_verified": model.IsVerified,
				"verified_at": model.VerifiedAt,
				"updated_at":  model.UpdatedAt,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return iamerrors.ErrIdentityNotFound
		}
		return nil
	}), iamerrors.ErrIdentityNotFound)
}

// Delete menghapus identitas login pengguna berdasarkan UUID identitas.
func (r *IdentityRepository) Delete(ctx context.Context, id string) error {
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
	res := db.Where("id = ?", uid).Delete(&userIdentityModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrIdentityNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrIdentityNotFound
	}
	return nil
}

// findIdentitiesByEmailAndUsername menanyakan user_identities berdasarkan daftar email dan/atau username dalam satu kueri terparameter.
func findIdentitiesByEmailAndUsername(tx *gorm.DB, normEmails, normUsernames []string) ([]userIdentityModel, error) {
	var rows []userIdentityModel
	switch {
	case len(normEmails) > 0 && len(normUsernames) > 0:
		err := tx.Where(
			"(kind = ? AND value IN ?) OR (kind = ? AND value IN ?)",
			entity.IdentityKindEmail.String(), normEmails,
			entity.IdentityKindUsername.String(), normUsernames,
		).Find(&rows).Error
		return rows, err
	case len(normEmails) > 0:
		err := tx.Where("kind = ? AND value IN ?", entity.IdentityKindEmail.String(), normEmails).Find(&rows).Error
		return rows, err
	case len(normUsernames) > 0:
		err := tx.Where("kind = ? AND value IN ?", entity.IdentityKindUsername.String(), normUsernames).Find(&rows).Error
		return rows, err
	default:
		return rows, nil
	}
}
