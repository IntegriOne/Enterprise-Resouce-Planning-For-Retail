package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// UserCredentialRepository mendefinisikan operasi penyimpanan hash kata sandi dan status lockout pada tabel user_credentials.
type UserCredentialRepository interface {
	SelectByUserID(ctx context.Context, userID string) (*entity.UserCredential, error)
	Insert(ctx context.Context, credential *entity.UserCredential) error
	Update(ctx context.Context, credential *entity.UserCredential) error
	// DeleteByUserID menghapus baris kredensial kata sandi milik user tertentu secara idempoten.
	DeleteByUserID(ctx context.Context, userID string) error
	// RecordFailedAttemptAtomic menambah penghitung gagal login secara atomik dan mengembalikan jumlah serta batas waktu kunci terbaru.
	RecordFailedAttemptAtomic(ctx context.Context, userID string, maxAttempts int, lockDuration time.Duration, now time.Time) (int, *time.Time, error)
	// ResetFailedAttemptsAtomic membersihkan penghitung kegagalan login dan kunci sementara secara atomik bila masih ada.
	ResetFailedAttemptsAtomic(ctx context.Context, userID string, now time.Time) error
}

// UserTOTPFactorRepository mendefinisikan operasi faktor TOTP Google Authenticator pada tabel user_totp_factors.
type UserTOTPFactorRepository interface {
	SelectByUserID(ctx context.Context, userID string) (*entity.UserTOTPFactor, error)
	Insert(ctx context.Context, factor *entity.UserTOTPFactor) error
	Update(ctx context.Context, factor *entity.UserTOTPFactor) error
	// UpdateLastUsedStepIfGreater memperbarui last_used_step secara atomik hanya jika step > last_used_step saat ini.
	UpdateLastUsedStepIfGreater(ctx context.Context, factorID string, step int64, updatedAt time.Time) (bool, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

// UserRecoveryCodeRepository mendefinisikan operasi kode pemulihan 2FA pada tabel user_recovery_codes.
type UserRecoveryCodeRepository interface {
	SelectUnusedByUserID(ctx context.Context, userID string) ([]*entity.UserRecoveryCode, error)
	CountUnusedByUserID(ctx context.Context, userID string) (int64, error)
	InsertBatch(ctx context.Context, codes []*entity.UserRecoveryCode) error
	// UpdateUsedAtIfUnused menandai used_at secara atomik hanya jika kode masih belum terpakai.
	UpdateUsedAtIfUnused(ctx context.Context, id string, usedAt time.Time) (bool, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

// UserTokenRepository mendefinisikan operasi token sekali pakai pada tabel user_tokens.
type UserTokenRepository interface {
	SelectByDigestAndPurpose(ctx context.Context, tokenDigest []byte, purpose entity.TokenPurpose) (*entity.UserToken, error)
	Insert(ctx context.Context, token *entity.UserToken) error
	// UpdateActiveStatusByUserAndPurpose menandai token aktif milik user+purpose menjadi expired atau revoked secara atomik.
	UpdateActiveStatusByUserAndPurpose(ctx context.Context, userID string, purpose entity.TokenPurpose, nextStatus entity.TokenStatus) (int64, error)
	// UpdateUsedAtIfUnused menandai used_at dan status consumed secara atomik hanya jika token masih aktif dan belum dipakai.
	UpdateUsedAtIfUnused(ctx context.Context, id string, usedAt time.Time) (bool, error)
	DeleteExpiredOrUsed(ctx context.Context, cutoff time.Time) (int64, error)
}
