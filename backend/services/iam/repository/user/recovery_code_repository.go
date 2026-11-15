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

var _ domainrepo.UserRecoveryCodeRepository = (*RecoveryCodeRepository)(nil)

// RecoveryCodeRepository mengimplementasikan kontrak UserRecoveryCodeRepository menggunakan PostgreSQL dan GORM.
type RecoveryCodeRepository struct {
	db *gorm.DB
}

// NewRecoveryCodeRepository membangun RecoveryCodeRepository dengan validasi koneksi database fail-closed.
func NewRecoveryCodeRepository(db *gorm.DB) (*RecoveryCodeRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &RecoveryCodeRepository{db: db}, nil
}

// SelectUnusedByUserID mengambil daftar kode pemulihan 2FA yang belum pernah dipakai untuk pengguna terkait.
func (r *RecoveryCodeRepository) SelectUnusedByUserID(ctx context.Context, userID string) ([]*entity.UserRecoveryCode, error) {
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
	var rows []userRecoveryCodeModel
	if err := db.Where("user_id = ? AND used_at IS NULL", uid).
		Order("created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	result := make([]*entity.UserRecoveryCode, 0, len(rows))
	for i := range rows {
		mapped, mapErr := recoveryCodeToDomain(&rows[i])
		if mapErr != nil {
			return nil, mapErr
		}
		result = append(result, mapped)
	}
	return result, nil
}

// CountUnusedByUserID menghitung sisa kode pemulihan 2FA aktif yang belum digunakan.
func (r *RecoveryCodeRepository) CountUnusedByUserID(ctx context.Context, userID string) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return 0, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := db.Model(&userRecoveryCodeModel{}).
		Where("user_id = ? AND used_at IS NULL", uid).
		Count(&count).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return count, nil
}

// InsertBatch menyimpan sekumpulan kode pemulihan 2FA baru dalam satu batch atomik.
func (r *RecoveryCodeRepository) InsertBatch(ctx context.Context, codes []*entity.UserRecoveryCode) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	if len(codes) == 0 {
		return nil
	}
	models := make([]userRecoveryCodeModel, 0, len(codes))
	for _, code := range codes {
		model, err := recoveryCodeToModel(code)
		if err != nil {
			return err
		}
		models = append(models, model)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&models).Error, nil)
}

// UpdateUsedAtIfUnused menandai satu kode pemulihan terpakai secara atomik hanya jika used_at masih NULL.
func (r *RecoveryCodeRepository) UpdateUsedAtIfUnused(ctx context.Context, id string, usedAt time.Time) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	cid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	ts := pgstore.MustLocalNow(usedAt)
	res := db.Model(&userRecoveryCodeModel{}).
		Where("id = ? AND used_at IS NULL", cid).
		Update("used_at", ts)
	if res.Error != nil {
		return false, pgstore.MapError(res.Error, iamerrors.ErrRecoveryCodeNotFound)
	}
	return res.RowsAffected > 0, nil
}

// DeleteByUserID menghapus seluruh kode pemulihan milik satu pengguna saat regenerasi atau penonaktifan 2FA.
func (r *RecoveryCodeRepository) DeleteByUserID(ctx context.Context, userID string) error {
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
	return pgstore.MapError(db.Where("user_id = ?", uid).Delete(&userRecoveryCodeModel{}).Error, nil)
}
