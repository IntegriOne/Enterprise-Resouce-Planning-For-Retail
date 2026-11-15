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

var _ domainrepo.UserTokenRepository = (*TokenRepository)(nil)

// TokenRepository mengimplementasikan kontrak UserTokenRepository menggunakan PostgreSQL dan GORM.
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository membangun TokenRepository dengan validasi koneksi database fail-closed.
func NewTokenRepository(db *gorm.DB) (*TokenRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &TokenRepository{db: db}, nil
}

// SelectByDigestAndPurpose mengambil token sekali pakai berdasarkan SHA-256 digest dan tujuannya.
func (r *TokenRepository) SelectByDigestAndPurpose(ctx context.Context, tokenDigest []byte, purpose entity.TokenPurpose) (*entity.UserToken, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	if len(tokenDigest) == 0 {
		return nil, iamerrors.ErrInvalidTokenDigest
	}
	if !purpose.IsValid() {
		return nil, iamerrors.ErrTokenInvalidPurpose
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m userTokenModel
	if err := db.Where("token_digest = ? AND purpose = ?", tokenDigest, purpose.String()).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrTokenNotFound)
	}
	return tokenToDomain(&m)
}

// Insert menyimpan token sekali pakai baru pada tabel user_tokens.
func (r *TokenRepository) Insert(ctx context.Context, token *entity.UserToken) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := tokenToModel(token)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// UpdateActiveStatusByUserAndPurpose menandai token berstatus active milik pengguna dan purpose tertentu menjadi expired atau revoked.
func (r *TokenRepository) UpdateActiveStatusByUserAndPurpose(ctx context.Context, userID string, purpose entity.TokenPurpose, nextStatus entity.TokenStatus) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return 0, err
	}
	if !purpose.IsValid() {
		return 0, iamerrors.ErrTokenInvalidPurpose
	}
	if nextStatus != entity.TokenStatusExpired && nextStatus != entity.TokenStatusRevoked {
		return 0, iamerrors.ErrInvalidTokenStatus
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	res := db.Model(&userTokenModel{}).
		Where("user_id = ? AND purpose = ? AND status = ?", uid, purpose.String(), entity.TokenStatusActive.String()).
		Update("status", nextStatus.String())
	if res.Error != nil {
		return 0, pgstore.MapError(res.Error, nil)
	}
	return res.RowsAffected, nil
}

// UpdateUsedAtIfUnused menandai status consumed dan kolom consumed_at secara atomik hanya jika token masih aktif dan belum pernah dikonsumsi.
func (r *TokenRepository) UpdateUsedAtIfUnused(ctx context.Context, id string, usedAt time.Time) (bool, error) {
	if r == nil {
		return false, pgstore.CheckDB(nil)
	}
	tid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return false, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return false, err
	}
	ts := pgstore.MustLocalNow(usedAt)
	res := db.Model(&userTokenModel{}).
		Where("id = ? AND status = ? AND consumed_at IS NULL AND expires_at > ?", tid, entity.TokenStatusActive.String(), ts).
		Updates(map[string]any{
			"status":      entity.TokenStatusConsumed.String(),
			"consumed_at": ts,
		})
	if res.Error != nil {
		return false, pgstore.MapError(res.Error, iamerrors.ErrTokenNotFound)
	}
	return res.RowsAffected > 0, nil
}

// DeleteExpiredOrUsed membersihkan token yang sudah terpakai atau kedaluwarsa sebelum waktu cutoff.
func (r *TokenRepository) DeleteExpiredOrUsed(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	ts := pgstore.MustLocalNow(cutoff)
	res := db.Where("consumed_at IS NOT NULL OR expires_at <= ?", ts).Delete(&userTokenModel{})
	if res.Error != nil {
		return 0, pgstore.MapError(res.Error, nil)
	}
	return res.RowsAffected, nil
}
