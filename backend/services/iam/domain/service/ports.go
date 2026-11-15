// Package service menyediakan kebijakan domain murni dan port layanan kriptografi/waktu untuk IAM IntegriOne.
package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// Clock menyediakan sumber waktu terabstraksi agar logika domain tetap deterministik saat diuji.
type Clock interface {
	Now() time.Time
}

// SystemClock mengimplementasikan Clock menggunakan waktu lokal sistem operasi.
type SystemClock struct{}

// Now mengembalikan waktu lokal sistem operasi saat ini melalui time.Now().
func (SystemClock) Now() time.Time {
	return time.Now()
}

// IDGenerator membuat pengenal unik UUID baru untuk entitas IAM.
type IDGenerator interface {
	NewUUID() (string, error)
}

// PasswordHasher mengabstraksikan hashing dan verifikasi kata sandi seperti Argon2id tanpa mengikat domain ke library tertentu.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (bool, error)
}

// TokenGenerator menghasilkan token buram acak, digest SHA-256, dan kode pemulihan 2FA.
type TokenGenerator interface {
	GenerateOpaqueToken(byteLen int) (plaintext string, digest []byte, err error)
	Digest(plaintext string) []byte
	GenerateRecoveryCodes(count int) ([]string, error)
}

// SecretCipher mengenkripsi dan mendekripsi secret TOTP sebelum disimpan di user_totp_factors.
type SecretCipher interface {
	Encrypt(plaintext []byte) (ciphertext []byte, keyVersion int16, err error)
	Decrypt(ciphertext []byte, keyVersion int16) ([]byte, error)
}

// TOTPProvider membangkitkan secret TOTP Google Authenticator dan memvalidasi kode OTP 6 digit beserta time step.
type TOTPProvider interface {
	GenerateSecret(accountName, issuer string) (secret []byte, otpAuthURI string, err error)
	VerifyCode(secret []byte, code string, at time.Time) (matched bool, step int64, err error)
}

// CSRFDeriver membentuk token CSRF deterministik multi-tab yang terikat ke sesi aktif melalui HMAC-SHA256.
type CSRFDeriver interface {
	ActiveKeyVersion() int
	DeriveToken(keyVersion int, sessionID string, tokenDigest []byte, generation int64) (plaintext string, digest []byte, err error)
}

// RuntimeSecurityProjector memproyeksikan sesi, pencabutan, dan AccessSnapshot IAM ke cache Redis keamanan.
type RuntimeSecurityProjector interface {
	IsEnabled() bool
	ProjectSessionIssued(ctx context.Context, user *entity.User, sess *entity.Session, recoveryEpoch int64) error
	ProjectSessionUpdated(ctx context.Context, user *entity.User, sess *entity.Session, recoveryEpoch int64) error
	ProjectSessionRevoked(ctx context.Context, user *entity.User, tokenDigest []byte, recoveryEpoch int64) error
	ProjectUserSessionsRevoked(ctx context.Context, user *entity.User, tokenDigests [][]byte, recoveryEpoch int64) error
	ProjectUserSecurityChanged(ctx context.Context, user *entity.User, recoveryEpoch int64, invalidateAuthz bool) error
	ProjectAccessSnapshot(
		ctx context.Context,
		snapshot *entity.AccessSnapshot,
		canAccessIdentity bool,
		defaultRedirectPath string,
		allowedRedirectPaths []string,
	) error
}
