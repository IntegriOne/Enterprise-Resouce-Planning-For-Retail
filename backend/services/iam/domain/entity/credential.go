package entity

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const redactedSecret = "[REDACTED]"

// UserCredential menyimpan hash kata sandi dan status penguncian percobaan login pada tabel user_credentials.
type UserCredential struct {
	UserID              string
	passwordHash        string
	MustChangePassword  bool
	PasswordChangedAt   time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// NewUserCredential membentuk entitas kredensial baru dengan validasi UUID pengguna dan hash non-kosong.
func NewUserCredential(userID, passwordHash string, mustChange bool, now time.Time) (*UserCredential, error) {
	ts := normalizeLocalTime(now)
	return RehydrateUserCredential(userID, passwordHash, mustChange, ts, 0, nil, ts, ts)
}

// RehydrateUserCredential membangun kembali entitas kredensial dari lapisan persistensi sambil menjaga constraint percobaan gagal.
func RehydrateUserCredential(
	userID, passwordHash string,
	mustChange bool,
	passwordChangedAt time.Time,
	failedAttempts int,
	lockedUntil *time.Time,
	createdAt, updatedAt time.Time,
) (*UserCredential, error) {
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(passwordHash) == "" {
		return nil, iamerrors.NewValidation("password_hash", iamerrors.Code("iam.credential.empty_hash"), "password hash must not be empty")
	}
	if failedAttempts < 0 {
		return nil, iamerrors.NewValidation("failed_login_attempts", iamerrors.Code("iam.credential.negative_attempts"), "failed login attempts must be non-negative")
	}
	return &UserCredential{
		UserID:              strings.TrimSpace(userID),
		passwordHash:        passwordHash,
		MustChangePassword:  mustChange,
		PasswordChangedAt:   normalizeLocalTime(passwordChangedAt),
		FailedLoginAttempts: failedAttempts,
		LockedUntil:         cloneTimePtr(lockedUntil),
		CreatedAt:           normalizeLocalTime(createdAt),
		UpdatedAt:           normalizeLocalTime(updatedAt),
	}, nil
}

// PasswordHash mengembalikan hash Argon2id untuk keperluan verifikasi oleh layanan autentikasi.
func (c *UserCredential) PasswordHash() string {
	if c == nil {
		return ""
	}
	return c.passwordHash
}

// IsLocked melaporkan apakah akun masih berada dalam jendela penguncian sementara.
func (c *UserCredential) IsLocked(now time.Time) bool {
	if c == nil || c.LockedUntil == nil {
		return false
	}
	return normalizeLocalTime(now).Before(normalizeLocalTime(*c.LockedUntil))
}

// RecordFailedAttempt menambah penghitung gagal login dan mengunci akun jika ambang batas tercapai.
func (c *UserCredential) RecordFailedAttempt(maxAttempts int, lockDuration time.Duration, now time.Time) {
	if c == nil {
		return
	}
	ts := normalizeLocalTime(now)
	c.FailedLoginAttempts++
	if maxAttempts > 0 && lockDuration > 0 && c.FailedLoginAttempts >= maxAttempts {
		until := ts.Add(lockDuration)
		c.LockedUntil = &until
	}
	c.UpdatedAt = ts
}

// ResetFailedAttempts mereset penghitung kegagalan dan kunci sementara setelah login berhasil.
func (c *UserCredential) ResetFailedAttempts(now time.Time) {
	if c == nil {
		return
	}
	c.FailedLoginAttempts = 0
	c.LockedUntil = nil
	c.UpdatedAt = normalizeLocalTime(now)
}

// UpdatePasswordHash mengganti hash kata sandi, mereset penguncian, dan menghapus kewajiban rotasi kata sandi.
func (c *UserCredential) UpdatePasswordHash(newHash string, now time.Time) error {
	if c == nil {
		return iamerrors.ErrCredentialNotFound
	}
	if strings.TrimSpace(newHash) == "" {
		return iamerrors.NewValidation("password_hash", iamerrors.Code("iam.credential.empty_hash"), "password hash must not be empty")
	}
	ts := normalizeLocalTime(now)
	c.passwordHash = newHash
	c.MustChangePassword = false
	c.PasswordChangedAt = ts
	c.FailedLoginAttempts = 0
	c.LockedUntil = nil
	c.UpdatedAt = ts
	return nil
}

// String menyembunyikan hash kata sandi saat dicetak ke log atau string debug.
func (c UserCredential) String() string {
	return fmt.Sprintf("UserCredential{UserID:%s, MustChange:%t, FailedAttempts:%d, PasswordHash:%s}",
		c.UserID, c.MustChangePassword, c.FailedLoginAttempts, redactedSecret)
}

// GoString menyamarkan hash Argon2id pada format %#v.
func (c UserCredential) GoString() string {
	return c.String()
}

// LogValue memancarkan representasi slog aman tanpa hash kata sandi.
func (c UserCredential) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user_id", c.UserID),
		slog.Bool("must_change_password", c.MustChangePassword),
		slog.Int("failed_login_attempts", c.FailedLoginAttempts),
		slog.String("password_hash", redactedSecret),
	)
}

// UserRecoveryCode merepresentasikan satu kode cadangan pemulihan 2FA sekali pakai pada tabel user_recovery_codes.
type UserRecoveryCode struct {
	ID        string
	UserID    string
	codeHash  string
	UsedAt    *time.Time
	CreatedAt time.Time
}

// NewUserRecoveryCode membuat entitas kode pemulihan 2FA baru dengan hash tervalidasi.
func NewUserRecoveryCode(id, userID, codeHash string, now time.Time) (*UserRecoveryCode, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(codeHash) == "" {
		return nil, iamerrors.NewValidation("code_hash", iamerrors.Code("iam.recovery_code.empty_hash"), "recovery code hash must not be empty")
	}
	return &UserRecoveryCode{
		ID:        strings.TrimSpace(id),
		UserID:    strings.TrimSpace(userID),
		codeHash:  codeHash,
		CreatedAt: normalizeLocalTime(now),
	}, nil
}

// RehydrateUserRecoveryCode membangun kembali entitas kode pemulihan dari lapisan persistensi.
func RehydrateUserRecoveryCode(id, userID, codeHash string, usedAt *time.Time, createdAt time.Time) (*UserRecoveryCode, error) {
	rc, err := NewUserRecoveryCode(id, userID, codeHash, createdAt)
	if err != nil {
		return nil, err
	}
	rc.UsedAt = cloneTimePtr(usedAt)
	return rc, nil
}

// CodeHash mengembalikan hash kode pemulihan untuk pencocokan kriptografis.
func (r *UserRecoveryCode) CodeHash() string {
	if r == nil {
		return ""
	}
	return r.codeHash
}

// IsUsed melaporkan apakah kode pemulihan sudah pernah dipakai.
func (r *UserRecoveryCode) IsUsed() bool {
	return r != nil && r.UsedAt != nil
}

// Consume menandai kode pemulihan telah dipakai dan menolak penggunaan ulang.
func (r *UserRecoveryCode) Consume(now time.Time) error {
	if r == nil {
		return iamerrors.ErrRecoveryCodeNotFound
	}
	if r.UsedAt != nil {
		return iamerrors.ErrRecoveryCodeAlreadyUsed
	}
	ts := normalizeLocalTime(now)
	r.UsedAt = &ts
	return nil
}

// String menyembunyikan hash kode pemulihan pada keluaran string.
func (r UserRecoveryCode) String() string {
	return fmt.Sprintf("UserRecoveryCode{ID:%s, UserID:%s, Used:%t, CodeHash:%s}",
		r.ID, r.UserID, r.UsedAt != nil, redactedSecret)
}

// GoString menyamarkan hash kode pemulihan pada format %#v.
func (r UserRecoveryCode) GoString() string {
	return r.String()
}

// LogValue memancarkan representasi slog aman untuk kode pemulihan.
func (r UserRecoveryCode) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", r.ID),
		slog.String("user_id", r.UserID),
		slog.Bool("used", r.UsedAt != nil),
		slog.String("code_hash", redactedSecret),
	)
}

func normalizeLocalTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.Local()
}

func cloneTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := normalizeLocalTime(*t)
	return &v
}
