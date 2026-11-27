package entity

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// UserTOTPFactor menyimpan faktor 2FA Google Authenticator pada tabel user_totp_factors.
type UserTOTPFactor struct {
	ID              string
	UserID          string
	encryptedSecret []byte
	KeyVersion      int16
	IsEnabled       bool
	ConfirmedAt     *time.Time
	LastUsedStep    *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUserTOTPFactor membentuk faktor TOTP baru dengan salinan defensif secret terenkripsi.
func NewUserTOTPFactor(id, userID string, encryptedSecret []byte, keyVersion int16, now time.Time) (*UserTOTPFactor, error) {
	ts := normalizeLocalTime(now)
	return RehydrateUserTOTPFactor(id, userID, encryptedSecret, keyVersion, false, nil, nil, ts, ts)
}

// RehydrateUserTOTPFactor membangun kembali faktor TOTP dari lapisan persistensi sambil menegakkan ck_user_totp_confirmed_at.
func RehydrateUserTOTPFactor(
	id, userID string,
	encryptedSecret []byte,
	keyVersion int16,
	isEnabled bool,
	confirmedAt *time.Time,
	lastUsedStep *int64,
	createdAt, updatedAt time.Time,
) (*UserTOTPFactor, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if len(encryptedSecret) == 0 {
		return nil, iamerrors.NewValidation("encrypted_secret", iamerrors.Code("iam.totp.empty_secret"), "encrypted secret must not be empty")
	}
	if keyVersion < 1 {
		return nil, iamerrors.NewValidation("key_version", iamerrors.Code("iam.totp.invalid_key_version"), "key version must be at least 1")
	}
	if isEnabled && confirmedAt == nil {
		return nil, iamerrors.ErrTOTPNotConfirmed
	}

	var stepCopy *int64
	if lastUsedStep != nil {
		v := *lastUsedStep
		stepCopy = &v
	}

	return &UserTOTPFactor{
		ID:              strings.TrimSpace(id),
		UserID:          strings.TrimSpace(userID),
		encryptedSecret: bytes.Clone(encryptedSecret),
		KeyVersion:      keyVersion,
		IsEnabled:       isEnabled,
		ConfirmedAt:     cloneTimePtr(confirmedAt),
		LastUsedStep:    stepCopy,
		CreatedAt:       normalizeLocalTime(createdAt),
		UpdatedAt:       normalizeLocalTime(updatedAt),
	}, nil
}

// EncryptedSecret mengembalikan salinan byte secret terenkripsi agar caller tidak dapat memutasi state internal.
func (f *UserTOTPFactor) EncryptedSecret() []byte {
	if f == nil {
		return nil
	}
	return bytes.Clone(f.encryptedSecret)
}

// Confirm mengaktifkan faktor TOTP setelah pengguna membuktikan kepemilikan kode OTP pertama.
func (f *UserTOTPFactor) Confirm(now time.Time) error {
	if f == nil {
		return iamerrors.ErrTOTPNotFound
	}
	if f.IsEnabled {
		return iamerrors.ErrTOTPAlreadyConfirmed
	}
	ts := normalizeLocalTime(now)
	f.IsEnabled = true
	f.ConfirmedAt = &ts
	f.UpdatedAt = ts
	return nil
}

// CanUseStep memastikan faktor sudah aktif dan time step belum pernah dipakai (anti-replay).
func (f *UserTOTPFactor) CanUseStep(step int64) error {
	if f == nil {
		return iamerrors.ErrTOTPNotFound
	}
	if !f.IsEnabled || f.ConfirmedAt == nil {
		return iamerrors.ErrTOTPNotConfirmed
	}
	if f.LastUsedStep != nil && step <= *f.LastUsedStep {
		return iamerrors.ErrTOTPReplayDetected
	}
	return nil
}

// MarkStepUsed mencatat time step TOTP yang telah terverifikasi secara monotonik.
func (f *UserTOTPFactor) MarkStepUsed(step int64, now time.Time) error {
	if err := f.CanUseStep(step); err != nil {
		return err
	}
	v := step
	f.LastUsedStep = &v
	f.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Disable menonaktifkan faktor TOTP pengguna.
func (f *UserTOTPFactor) Disable(now time.Time) {
	if f == nil {
		return
	}
	f.IsEnabled = false
	f.UpdatedAt = normalizeLocalTime(now)
}

// String menyembunyikan payload secret TOTP terenkripsi dari log teks.
func (f UserTOTPFactor) String() string {
	return fmt.Sprintf("UserTOTPFactor{ID:%s, UserID:%s, KeyVersion:%d, Enabled:%t, Secret:%s}",
		f.ID, f.UserID, f.KeyVersion, f.IsEnabled, redactedSecret)
}

// GoString menyembunyikan secret TOTP pada format %#v.
func (f UserTOTPFactor) GoString() string {
	return f.String()
}

// LogValue memancarkan atribut faktor TOTP aman untuk slog.
func (f UserTOTPFactor) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", f.ID),
		slog.String("user_id", f.UserID),
		slog.Int("key_version", int(f.KeyVersion)),
		slog.Bool("is_enabled", f.IsEnabled),
		slog.String("encrypted_secret", redactedSecret),
	)
}
