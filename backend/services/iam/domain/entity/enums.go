// Package entity mendefinisikan entitas domain inti, value object, enum, dan invarian IAM IntegriOne.
package entity

import (
	"fmt"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// Konstanta kebijakan keamanan dan batas panjang kolom yang selaras dengan skema PostgreSQL IAM.
const (
	// SessionIdleTimeout adalah batas idle sesi maksimum selama 30 menit sesuai D01-D12.
	SessionIdleTimeout = 30 * time.Minute
	// SessionAbsoluteTimeout adalah batas umur sesi mutlak maksimum selama 8 jam tanpa perpanjangan.
	SessionAbsoluteTimeout = 8 * time.Hour
	// SnapshotMaxFreshness adalah batas kesegaran snapshot akses selama 30 detik.
	SnapshotMaxFreshness = 30 * time.Second
	// DefaultMaxFailedLoginAttempts adalah batas kegagalan login sebelum akun dikunci sementara.
	DefaultMaxFailedLoginAttempts = 5
	// DefaultLoginLockDuration adalah durasi penguncian sementara setelah percobaan login berulang gagal.
	DefaultLoginLockDuration = 15 * time.Minute
	// DefaultSensitiveReauthWindow adalah jendela waktu kelayakan re-autentikasi untuk izin sensitif.
	DefaultSensitiveReauthWindow = 10 * time.Minute

	// MaxDisplayNameLength membatasi panjang kolom users.display_name.
	MaxDisplayNameLength = 200
	// MaxAvatarURLLength membatasi panjang kolom users.avatar_url.
	MaxAvatarURLLength = 1000
	// DefaultUserLocale adalah lokal bawaan akun baru sesuai skema database IAM.
	DefaultUserLocale = "id-ID"
	// DefaultUserTimezone adalah zona waktu bawaan akun baru sesuai skema database IAM.
	DefaultUserTimezone = "Asia/Jakarta"
	// MaxLocaleLength membatasi panjang kolom users.locale.
	MaxLocaleLength = 20
	// MaxTimezoneLength membatasi panjang kolom users.timezone.
	MaxTimezoneLength = 80
	// MaxIdentityValueLength membatasi panjang kolom user_identities.value.
	MaxIdentityValueLength = 320
	// MaxRoleCodeLength membatasi panjang kolom roles.code.
	MaxRoleCodeLength = 100
	// MaxRoleNameLength membatasi panjang kolom roles.name.
	MaxRoleNameLength = 160
	// MaxManagementDomainLength membatasi panjang kolom roles.management_domain.
	MaxManagementDomainLength = 100
	// MaxRedirectPathLength membatasi panjang kolom roles.default_redirect_path.
	MaxRedirectPathLength = 120
	// MaxPermissionCodeLength membatasi panjang kolom permissions.code.
	MaxPermissionCodeLength = 160
	// MaxPermissionResourceLength membatasi panjang kolom permissions.resource.
	MaxPermissionResourceLength = 100
	// MaxPermissionActionLength membatasi panjang kolom permissions.action.
	MaxPermissionActionLength = 60
	// MaxGroupCodeLength membatasi panjang kolom groups.code.
	MaxGroupCodeLength = 100
	// MaxGroupNameLength membatasi panjang kolom groups.name.
	MaxGroupNameLength = 160
)

// UserKind merepresentasikan kategori akun utama dalam satu tabel users sesuai enum PostgreSQL user_kind.
type UserKind string

const (
	// UserKindCustomer adalah akun pembeli pada Web Commerce.
	UserKindCustomer UserKind = "customer"
	// UserKindEmployee adalah akun pegawai internal ERP IntegriOne.
	UserKindEmployee UserKind = "employee"
	// UserKindVendor adalah akun mitra vendor eksternal.
	UserKindVendor UserKind = "vendor"
	// UserKindSupplier adalah akun mitra pemasok eksternal.
	UserKindSupplier UserKind = "supplier"
	// UserKindSystem adalah akun layanan non-interaktif internal.
	UserKindSystem UserKind = "system"
)

// String mengembalikan representasi string enum user_kind.
func (k UserKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah UserKind termasuk nilai enum yang didukung skema.
func (k UserKind) IsValid() bool {
	switch k {
	case UserKindCustomer, UserKindEmployee, UserKindVendor, UserKindSupplier, UserKindSystem:
		return true
	default:
		return false
	}
}

// ParseUserKind mengonversi string menjadi UserKind tervalidasi.
func ParseUserKind(raw string) (UserKind, error) {
	k := UserKind(raw)
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidUserKind, raw)
	}
	return k, nil
}

// UserStatus merepresentasikan siklus hidup akun sesuai enum PostgreSQL user_status.
type UserStatus string

const (
	// UserStatusPending menandakan akun baru menunggu aktivasi atau verifikasi awal.
	UserStatusPending UserStatus = "pending"
	// UserStatusActive menandakan akun aktif dan diizinkan melakukan autentikasi.
	UserStatusActive UserStatus = "active"
	// UserStatusSuspended menandakan akun ditangguhkan sementara oleh admin.
	UserStatusSuspended UserStatus = "suspended"
	// UserStatusDisabled menandakan akun dinonaktifkan secara permanen atau keluar dari organisasi.
	UserStatusDisabled UserStatus = "disabled"
)

// String mengembalikan representasi string enum user_status.
func (s UserStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah UserStatus termasuk nilai enum yang didukung skema.
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPending, UserStatusActive, UserStatusSuspended, UserStatusDisabled:
		return true
	default:
		return false
	}
}

// ParseUserStatus mengonversi string menjadi UserStatus tervalidasi.
func ParseUserStatus(raw string) (UserStatus, error) {
	s := UserStatus(raw)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidUserStatus, raw)
	}
	return s, nil
}

// IdentityKind merepresentasikan tipe pengenal login sesuai enum PostgreSQL user_identities_kind.
type IdentityKind string

const (
	// IdentityKindEmail adalah pengenal berbasis alamat surel.
	IdentityKindEmail IdentityKind = "email"
	// IdentityKindUsername adalah pengenal berbasis nama pengguna internal atau mitra.
	IdentityKindUsername IdentityKind = "username"
)

// String mengembalikan representasi string enum user_identities_kind.
func (k IdentityKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah IdentityKind sesuai dengan enum user_identities_kind.
func (k IdentityKind) IsValid() bool {
	switch k {
	case IdentityKindEmail, IdentityKindUsername:
		return true
	default:
		return false
	}
}

// ParseIdentityKind mengonversi string menjadi IdentityKind tervalidasi.
func ParseIdentityKind(raw string) (IdentityKind, error) {
	k := IdentityKind(raw)
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidIdentityKind, raw)
	}
	return k, nil
}

// TokenPurpose merepresentasikan tujuan token sekali pakai sesuai enum PostgreSQL token_purpose.
type TokenPurpose string

const (
	// TokenPurposeEmailVerification adalah token verifikasi kepemilikan email.
	TokenPurposeEmailVerification TokenPurpose = "email_verification"
	// TokenPurposePasswordReset adalah token pengaturan ulang kata sandi.
	TokenPurposePasswordReset TokenPurpose = "password_reset"
	// TokenPurposeAccountInvitation adalah token undangan aktivasi akun baru.
	TokenPurposeAccountInvitation TokenPurpose = "account_invitation"
	// TokenPurposeMFAChallenge adalah token tantangan verifikasi 2FA sementara.
	TokenPurposeMFAChallenge TokenPurpose = "mfa_challenge"
)

// String mengembalikan representasi string enum token_purpose.
func (p TokenPurpose) String() string {
	return string(p)
}

// IsValid memeriksa apakah TokenPurpose sesuai enum PostgreSQL token_purpose.
func (p TokenPurpose) IsValid() bool {
	switch p {
	case TokenPurposeEmailVerification, TokenPurposePasswordReset, TokenPurposeAccountInvitation, TokenPurposeMFAChallenge:
		return true
	default:
		return false
	}
}

// ParseTokenPurpose mengonversi string menjadi TokenPurpose tervalidasi.
func ParseTokenPurpose(raw string) (TokenPurpose, error) {
	p := TokenPurpose(raw)
	if !p.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrTokenInvalidPurpose, raw)
	}
	return p, nil
}

// TokenStatus merepresentasikan status siklus hidup token sekali pakai sesuai enum PostgreSQL token_status.
type TokenStatus string

const (
	// TokenStatusActive menandakan token masih aktif dan dapat dipakai satu kali.
	TokenStatusActive TokenStatus = "active"
	// TokenStatusConsumed menandakan token sudah digunakan.
	TokenStatusConsumed TokenStatus = "consumed"
	// TokenStatusExpired menandakan token telah kedaluwarsa atau digantikan token baru.
	TokenStatusExpired TokenStatus = "expired"
	// TokenStatusRevoked menandakan token dibatalkan sebelum dipakai.
	TokenStatusRevoked TokenStatus = "revoked"
)

// String mengembalikan representasi string enum token_status.
func (s TokenStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah TokenStatus sesuai enum PostgreSQL token_status.
func (s TokenStatus) IsValid() bool {
	switch s {
	case TokenStatusActive, TokenStatusConsumed, TokenStatusExpired, TokenStatusRevoked:
		return true
	default:
		return false
	}
}

// ParseTokenStatus mengonversi string menjadi TokenStatus tervalidasi.
func ParseTokenStatus(raw string) (TokenStatus, error) {
	s := TokenStatus(raw)
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidTokenStatus, raw)
	}
	return s, nil
}

// SessionEndReason merepresentasikan alasan pengakhiran sesi sesuai enum PostgreSQL session_end_reason.
type SessionEndReason string

const (
	// SessionEndLogout menandakan pengguna mengakhiri sesi melalui logout eksplisit.
	SessionEndLogout SessionEndReason = "logout"
	// SessionEndIdleTimeout menandakan sesi berakhir karena tidak aktif melebihi batas idle.
	SessionEndIdleTimeout SessionEndReason = "idle_timeout"
	// SessionEndAbsoluteTimeout menandakan sesi berakhir karena melewati batas umur maksimum 8 jam.
	SessionEndAbsoluteTimeout SessionEndReason = "absolute_timeout"
	// SessionEndPasswordChanged menandakan sesi dicabut akibat pergantian kata sandi.
	SessionEndPasswordChanged SessionEndReason = "password_changed"
	// SessionEndAdminRevoked menandakan sesi dicabut secara administratif.
	SessionEndAdminRevoked SessionEndReason = "admin_revoked"
	// SessionEndUserDisabled menandakan sesi diakhiri karena akun pengguna dinonaktifkan atau ditangguhkan.
	SessionEndUserDisabled SessionEndReason = "user_disabled"
)

// String mengembalikan representasi string enum session_end_reason.
func (r SessionEndReason) String() string {
	return string(r)
}

// IsValid memeriksa apakah SessionEndReason sesuai enum PostgreSQL session_end_reason.
func (r SessionEndReason) IsValid() bool {
	switch r {
	case SessionEndLogout, SessionEndIdleTimeout, SessionEndAbsoluteTimeout, SessionEndPasswordChanged, SessionEndAdminRevoked, SessionEndUserDisabled:
		return true
	default:
		return false
	}
}

// ParseSessionEndReason mengonversi string menjadi SessionEndReason tervalidasi.
func ParseSessionEndReason(raw string) (SessionEndReason, error) {
	r := SessionEndReason(raw)
	if !r.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidSessionEndReason, raw)
	}
	return r, nil
}

// ScopeType merepresentasikan tipe lingkup penugasan peran sesuai enum PostgreSQL scope_type.
type ScopeType string

const (
	// ScopeTypeGlobal menandakan peran berlaku lintas seluruh unit tanpa ID lingkup spesifik.
	ScopeTypeGlobal ScopeType = "global"
	// ScopeTypeLegalEntity membatasi peran pada ID entitas hukum tertentu.
	ScopeTypeLegalEntity ScopeType = "legal_entity"
	// ScopeTypeOrganizationUnit membatasi peran pada ID unit organisasi tertentu.
	ScopeTypeOrganizationUnit ScopeType = "organization_unit"
	// ScopeTypeFacility membatasi peran pada ID gudang atau toko tertentu.
	ScopeTypeFacility ScopeType = "facility"
	// ScopeTypeSupplierParty membatasi peran pada ID mitra pemasok tertentu.
	ScopeTypeSupplierParty ScopeType = "supplier_party"
	// ScopeTypeVendorParty membatasi peran pada ID mitra vendor tertentu.
	ScopeTypeVendorParty ScopeType = "vendor_party"
	// ScopeTypeSelf membatasi peran hanya pada sumber daya milik pengguna sendiri.
	ScopeTypeSelf ScopeType = "self"
)

// String mengembalikan representasi string enum scope_type.
func (t ScopeType) String() string {
	return string(t)
}

// IsValid memeriksa apakah ScopeType sesuai enum PostgreSQL scope_type.
func (t ScopeType) IsValid() bool {
	switch t {
	case ScopeTypeGlobal, ScopeTypeLegalEntity, ScopeTypeOrganizationUnit, ScopeTypeFacility, ScopeTypeSupplierParty, ScopeTypeVendorParty, ScopeTypeSelf:
		return true
	default:
		return false
	}
}

// ParseScopeType mengonversi string menjadi ScopeType tervalidasi.
func ParseScopeType(raw string) (ScopeType, error) {
	t := ScopeType(raw)
	if !t.IsValid() {
		return "", fmt.Errorf("%w: %q", iamerrors.ErrInvalidScope, raw)
	}
	return t, nil
}
