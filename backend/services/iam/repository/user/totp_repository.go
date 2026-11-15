package user

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.UserTOTPFactorRepository = (*TOTPFactorRepository)(nil)

// TOTPFactorRepository mengimplementasikan kontrak UserTOTPFactorRepository menggunakan PostgreSQL dan GORM.
type TOTPFactorRepository struct {
	db *gorm.DB
}

// NewTOTPFactorRepository membangun TOTPFactorRepository dengan validasi koneksi database fail-closed.
func NewTOTPFactorRepository(db *gorm.DB) (*TOTPFactorRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TOTPFactorRepository{db: db}, nil
}

// SelectByUserID mengambil faktor TOTP Google Authenticator pengguna berdasarkan UUID pengguna.
func (r *TOTPFactorRepository) SelectByUserID(ctx context.Context, userID string) (*entity.UserTOTPFactor, error) {
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
	var m userTOTPFactorModel
	if err := db.Where("user_id = ?", uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrTOTPNotFound)
	}
	return totpToDomain(&m)
}

// Insert menyimpan faktor TOTP terenkripsi baru untuk pengguna.
func (r *TOTPFactorRepository) Insert(ctx context.Context, factor *entity.UserTOTPFactor) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := totpToModel(factor)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// Update memperbarui status aktivasi, versi kunci, atau secret faktor TOTP yang sudah terdaftar.
func (r *TOTPFactorRepository) Update(ctx context.Context, factor *entity.UserTOTPFactor) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := totpToModel(factor)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&userTOTPFactorModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"encrypted_secret": model.EncryptedSecret,
			"key_version":      model.KeyVersion,
			"is_enabled":       model.IsEnabled,
			"confirmed_at":     model.ConfirmedAt,
			"last_used_step":   model.LastUsedStep,
			"updated_at":       model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrTOTPNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrTOTPNotFound
	}
	return nil
}

// UpdateLastUsedStepIfGreater mencegah replay attack kode OTP dengan pembaruan monotonik atomik di PostgreSQL.
func (r *TOTPFactorRepository) UpdateLastUsedStepIfGreater(ctx context.Context, factorID string, step int64, updatedAt time.Time) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	fid, err := pgstore.ValidateUUIDParam("factor_id", factorID)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	ts := pgstore.MustLocalNow(updatedAt)
	res := db.Model(&userTOTPFactorModel{}).
		Where("id = ? AND is_enabled = ? AND confirmed_at IS NOT NULL AND (last_used_step IS NULL OR last_used_step < ?)", fid, true, step).
		Updates(map[string]any{
			"last_used_step": step,
			"last_used_at":   ts,
			"updated_at":     ts,
		})
	if res.Error != nil {
		return false, pgstore.MapError(res.Error, iamerrors.ErrTOTPNotFound)
	}
	return res.RowsAffected > 0, nil
}

// DeleteByUserID menghapus faktor TOTP milik pengguna tertentu saat reset atau penonaktifan 2FA.
func (r *TOTPFactorRepository) DeleteByUserID(ctx context.Context, userID string) error {
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
	res := db.Where("user_id = ?", uid).Delete(&userTOTPFactorModel{})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrTOTPNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrTOTPNotFound
	}
	return nil
}
