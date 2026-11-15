package entity

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// Pola regex dikompilasi satu kali pada level package untuk menghindari overhead alokasi berulang.
var (
	uuidPattern           = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	usernamePattern       = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,62}[a-z0-9]$`)
	codePattern           = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)
	permissionCodePattern = regexp.MustCompile(`^[a-z0-9]+([._:-][a-z0-9]+)*$`)
	internalPathPattern   = regexp.MustCompile(`^/[a-zA-Z0-9/_-]*$`)
)

// IsValidUUID memeriksa apakah string mengikuti format UUID standar 36 karakter.
func IsValidUUID(raw string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(raw))
}

// ValidateUUID memvalidasi string UUID dan mengembalikan error domain dengan konteks nama field bila tidak valid.
func ValidateUUID(field, raw string) error {
	if !IsValidUUID(raw) {
		return iamerrors.NewValidation(field, iamerrors.Code("iam.invalid_uuid"), "must be a valid uuid")
	}
	return nil
}

// NormalizeCode menormalkan kode peran, grup, atau izin dengan memangkas spasi dan menurunkan huruf.
func NormalizeCode(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidateEntityCode memvalidasi format dan batas panjang kode untuk entitas peran atau grup.
func ValidateEntityCode(code string, maxLen int, invalidErr error) (string, error) {
	normalized := NormalizeCode(code)
	if normalized == "" || utf8.RuneCountInString(normalized) > maxLen || !codePattern.MatchString(normalized) {
		return "", invalidErr
	}
	return normalized, nil
}

// ValidatePermissionCode memvalidasi kode izin dengan pemisah titik, garis bawah, tanda hubung, atau titik dua.
func ValidatePermissionCode(code string) (string, error) {
	normalized := NormalizeCode(code)
	if normalized == "" || utf8.RuneCountInString(normalized) > MaxPermissionCodeLength || !permissionCodePattern.MatchString(normalized) {
		return "", iamerrors.ErrInvalidPermissionCode
	}
	return normalized, nil
}

// IsSelfServicePermissionCode memeriksa apakah kode izin merupakan izin mandiri (*self) yang terikat kepemilikan principal.
func IsSelfServicePermissionCode(code string) bool {
	return authz.IsSelfServicePermissionCode(code)
}

// BaselineSelfRoleCodeForUserKind mengembalikan kode peran dasar self-service sesuai kategori akun manusia non-system.
func BaselineSelfRoleCodeForUserKind(kind UserKind) string {
	switch kind {
	case UserKindCustomer:
		return "customer"
	case UserKindEmployee, UserKindVendor, UserKindSupplier:
		return "iam_selfservice"
	default:
		return ""
	}
}

// ValidateEmail memvalidasi format alamat email murni tanpa nama tampilan atau karakter kontrol.
func ValidateEmail(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || utf8.RuneCountInString(normalized) > MaxIdentityValueLength {
		return "", iamerrors.ErrInvalidEmail
	}
	if strings.ContainsAny(normalized, " \t\r\n<>") {
		return "", iamerrors.ErrInvalidEmail
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized || parsed.Name != "" {
		return "", iamerrors.ErrInvalidEmail
	}
	at := strings.LastIndex(normalized, "@")
	if at <= 0 || at == len(normalized)-1 {
		return "", iamerrors.ErrInvalidEmail
	}
	domain := normalized[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return "", iamerrors.ErrInvalidEmail
	}
	return normalized, nil
}

// ValidateUsername memvalidasi format username yang hanya memuat huruf kecil, angka, titik, garis bawah, atau tanda hubung.
func ValidateUsername(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || !usernamePattern.MatchString(normalized) {
		return "", iamerrors.ErrInvalidUsername
	}
	return normalized, nil
}

// ValidateInternalRedirectPath memastikan jalur redirect bersifat relatif internal yang aman terhadap open redirect.
func ValidateInternalRedirectPath(path string) error {
	if path == "" || len(path) > MaxRedirectPathLength {
		return iamerrors.ErrRedirectPathInvalid
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return iamerrors.ErrRedirectPathInvalid
	}
	if strings.Contains(path, "..") || !internalPathPattern.MatchString(path) {
		return iamerrors.ErrRedirectPathInvalid
	}
	return nil
}

// ValidateOptionalURL memvalidasi URL avatar opsional agar hanya menerima skema http/https dalam batas panjang skema.
func ValidateOptionalURL(field, raw string, maxLen int) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if utf8.RuneCountInString(trimmed) > maxLen || hasControlOrWhitespace(trimmed) {
		return iamerrors.WithField(iamerrors.ErrInvalidAvatarURL, field, "url exceeds bounds or contains invalid characters")
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Host == "" {
		return iamerrors.WithField(iamerrors.ErrInvalidAvatarURL, field, "url must be valid http or https")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return iamerrors.WithField(iamerrors.ErrInvalidAvatarURL, field, "url scheme must be http or https")
	}
	return nil
}

// HasControlChar memeriksa apakah string mengandung karakter kontrol Unicode seperti CR, LF, TAB, atau NUL.
func HasControlChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func hasControlOrWhitespace(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return true
		}
	}
	return false
}
