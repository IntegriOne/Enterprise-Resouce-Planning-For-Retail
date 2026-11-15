package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// User merepresentasikan entitas akun pada tabel users sebagai satu sumber identitas lintas Web Commerce dan ERP.
type User struct {
	ID           string
	Kind         UserKind
	Status       UserStatus
	DisplayName  string
	AvatarURL    string
	Locale       string
	Timezone     string
	AuthEpoch    int64
	PermRevision int64
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// NewUser membentuk entitas User baru sekaligus menegakkan invarian validasi awal skema IAM.
func NewUser(id string, kind UserKind, status UserStatus, displayName string, now time.Time) (*User, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if !kind.IsValid() {
		return nil, iamerrors.ErrInvalidUserKind
	}
	if !status.IsValid() {
		return nil, iamerrors.ErrInvalidUserStatus
	}
	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" || utf8.RuneCountInString(trimmedName) > MaxDisplayNameLength || HasControlChar(trimmedName) {
		return nil, iamerrors.ErrInvalidDisplayName
	}

	ts := normalizeLocalTime(now)
	return &User{
		ID:           strings.TrimSpace(id),
		Kind:         kind,
		Status:       status,
		DisplayName:  trimmedName,
		Locale:       DefaultUserLocale,
		Timezone:     DefaultUserTimezone,
		AuthEpoch:    1,
		PermRevision: 1,
		CreatedAt:    ts,
		UpdatedAt:    ts,
	}, nil
}

// IsDeleted melaporkan apakah pengguna telah dihapus secara lunak.
func (u *User) IsDeleted() bool {
	return u != nil && u.DeletedAt != nil
}

// CanAuthenticate mengevaluasi secara fail-closed apakah status akun mengizinkan proses login.
func (u *User) CanAuthenticate() error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	switch u.Status {
	case UserStatusActive:
		return nil
	case UserStatusSuspended:
		return iamerrors.ErrUserSuspended
	case UserStatusDisabled:
		return iamerrors.ErrUserDisabled
	case UserStatusPending:
		return iamerrors.ErrUserInactive
	default:
		return iamerrors.ErrUserInactive
	}
}

// Activate mengubah status pengguna menjadi aktif bila akun belum dihapus lunak.
func (u *User) Activate(now time.Time) error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	u.Status = UserStatusActive
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Suspend menangguhkan akun pengguna sementara bila belum dihapus lunak.
func (u *User) Suspend(now time.Time) error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	u.Status = UserStatusSuspended
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Disable menonaktifkan akun pengguna secara administratif bila belum dihapus lunak.
func (u *User) Disable(now time.Time) error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	u.Status = UserStatusDisabled
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak akun sekaligus mengubah status menjadi disabled.
func (u *User) SoftDelete(now time.Time) error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	ts := normalizeLocalTime(now)
	u.Status = UserStatusDisabled
	u.DeletedAt = &ts
	u.UpdatedAt = ts
	return nil
}

// RecordLogin memperbarui cap waktu login terakhir pada entitas pengguna.
func (u *User) RecordLogin(now time.Time) {
	if u == nil {
		return
	}
	ts := normalizeLocalTime(now)
	u.LastLoginAt = &ts
	u.UpdatedAt = ts
}

// UpdateProfile memperbarui atribut profil tampilan pengguna setelah memvalidasi batas panjang dan URL.
func (u *User) UpdateProfile(displayName, avatarURL, locale, timezone string, now time.Time) error {
	if u == nil {
		return iamerrors.ErrUserNotFound
	}
	if u.IsDeleted() {
		return iamerrors.ErrUserDeleted
	}
	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" || utf8.RuneCountInString(trimmedName) > MaxDisplayNameLength || HasControlChar(trimmedName) {
		return iamerrors.ErrInvalidDisplayName
	}
	if err := ValidateOptionalURL("avatar_url", avatarURL, MaxAvatarURLLength); err != nil {
		return err
	}
	trimmedLocale := strings.TrimSpace(locale)
	if utf8.RuneCountInString(trimmedLocale) > MaxLocaleLength || HasControlChar(trimmedLocale) {
		return iamerrors.ErrInvalidLocale
	}
	trimmedTimezone := strings.TrimSpace(timezone)
	if utf8.RuneCountInString(trimmedTimezone) > MaxTimezoneLength || HasControlChar(trimmedTimezone) {
		return iamerrors.ErrInvalidTimezone
	}

	u.DisplayName = trimmedName
	u.AvatarURL = strings.TrimSpace(avatarURL)
	if trimmedLocale != "" {
		u.Locale = trimmedLocale
	} else if strings.TrimSpace(u.Locale) == "" {
		u.Locale = DefaultUserLocale
	}
	if trimmedTimezone != "" {
		u.Timezone = trimmedTimezone
	} else if strings.TrimSpace(u.Timezone) == "" {
		u.Timezone = DefaultUserTimezone
	}
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// UserIdentity merepresentasikan pengenal login terhubung pada tabel user_identities.
type UserIdentity struct {
	ID         string
	UserID     string
	Kind       IdentityKind
	Value      string
	IsVerified bool
	IsPrimary  bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NormalizeIdentityValue menormalkan dan memvalidasi nilai identitas login berdasarkan jenisnya.
func NormalizeIdentityValue(kind IdentityKind, value string) (string, error) {
	switch kind {
	case IdentityKindEmail:
		return ValidateEmail(value)
	case IdentityKindUsername:
		return ValidateUsername(value)
	default:
		return "", iamerrors.ErrInvalidIdentityKind
	}
}

// NewUserIdentity membuat entitas identitas login baru dalam status awal belum terverifikasi.
func NewUserIdentity(id, userID string, kind IdentityKind, value string, isPrimary bool, now time.Time) (*UserIdentity, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	normalized, err := NormalizeIdentityValue(kind, value)
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &UserIdentity{
		ID:        strings.TrimSpace(id),
		UserID:    strings.TrimSpace(userID),
		Kind:      kind,
		Value:     normalized,
		IsPrimary: isPrimary,
		CreatedAt: ts,
		UpdatedAt: ts,
	}, nil
}

// MarkVerified menandai identitas sudah terverifikasi sekaligus memenuhi constraint verified_at IS NOT NULL.
func (i *UserIdentity) MarkVerified(now time.Time) {
	if i == nil {
		return
	}
	ts := normalizeLocalTime(now)
	i.IsVerified = true
	i.VerifiedAt = &ts
	i.UpdatedAt = ts
}

// SetPrimary memperbarui penanda identitas utama pengguna.
func (i *UserIdentity) SetPrimary(primary bool, now time.Time) {
	if i == nil {
		return
	}
	i.IsPrimary = primary
	i.UpdatedAt = normalizeLocalTime(now)
}

// UpdateValue mengganti nilai identitas yang masih tertunda sekaligus mereset status verifikasinya.
func (i *UserIdentity) UpdateValue(value string, now time.Time) error {
	if i == nil {
		return iamerrors.ErrIdentityNotFound
	}
	normalized, err := NormalizeIdentityValue(i.Kind, value)
	if err != nil {
		return err
	}
	ts := normalizeLocalTime(now)
	i.Value = normalized
	i.IsVerified = false
	i.VerifiedAt = nil
	i.UpdatedAt = ts
	return nil
}

// UpdateUsernameValue mengganti nilai username secara langsung tanpa mereset status verifikasi karena username tidak memakai alur verifikasi email.
func (i *UserIdentity) UpdateUsernameValue(value string, now time.Time) error {
	if i == nil {
		return iamerrors.ErrIdentityNotFound
	}
	if i.Kind != IdentityKindUsername {
		return iamerrors.WithField(iamerrors.ErrInvalidInput, "identity", "identity kind must be username")
	}
	normalized, err := ValidateUsername(value)
	if err != nil {
		return err
	}
	ts := normalizeLocalTime(now)
	i.Value = normalized
	// Pertahankan atau aktifkan status terverifikasi agar pengguna dapat langsung masuk dengan username baru dan username lama segera bebas dipakai pengguna lain.
	i.IsVerified = true
	if i.VerifiedAt == nil {
		verifiedAt := ts
		i.VerifiedAt = &verifiedAt
	}
	i.UpdatedAt = ts
	return nil
}
