package entity

import (
	"bytes"
	"encoding/json"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Pola regex dikompilasi satu kali pada level package untuk menghindari alokasi berulang di jalur panas.
var (
	uuidPattern        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	codePattern        = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)
	countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)
	languagePattern    = regexp.MustCompile(`^[a-zA-Z]{2,3}(-[a-zA-Z0-9]{2,8})*$`)
)

// IsValidUUID memeriksa apakah string mengikuti format UUID standar 36 karakter.
func IsValidUUID(raw string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(raw))
}

// NormalizeUUID menormalkan string UUID menjadi huruf kecil tanpa spasi tepi.
func NormalizeUUID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidateUUID memvalidasi string UUID wajib dan mengembalikan error validasi dengan konteks field bila tidak valid.
func ValidateUUID(field, raw string) error {
	if !IsValidUUID(raw) {
		return eferrors.NewValidation(field, eferrors.Code("enterprise_foundation.invalid_uuid"), "must be a valid uuid")
	}
	return nil
}

// ValidateOptionalUUID memvalidasi pointer UUID opsional dan menormalkan ke huruf kecil bila tersedia.
func ValidateOptionalUUID(field string, raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	if !IsValidUUID(*raw) {
		return nil, eferrors.NewValidation(field, eferrors.Code("enterprise_foundation.invalid_uuid"), "must be a valid uuid")
	}
	normalized := NormalizeUUID(*raw)
	return &normalized, nil
}

// NormalizeCode menormalkan kode master data ke huruf kecil tanpa spasi tepi agar unik secara konsisten.
func NormalizeCode(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// ValidateCode memvalidasi format, panjang maksimum, dan karakter kode bisnis wajib.
func ValidateCode(raw string, maxLen int, invalidErr error) (string, error) {
	normalized := NormalizeCode(raw)
	if normalized == "" || utf8.RuneCountInString(normalized) > maxLen || !codePattern.MatchString(normalized) {
		return "", invalidErr
	}
	return normalized, nil
}

// ValidateRequiredText memastikan teks wajib tidak kosong, bebas karakter kontrol, dan dalam batas panjang rune.
func ValidateRequiredText(raw string, maxLen int, invalidErr error) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > maxLen || HasControlChar(trimmed) {
		return "", invalidErr
	}
	return trimmed, nil
}

// ValidateOptionalText memvalidasi pointer teks opsional; nilai kosong atau hanya spasi dinormalkan menjadi nil agar selaras dengan constraint NULL PostgreSQL.
func ValidateOptionalText(field string, raw *string, maxLen int, invalidErr error) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if HasControlChar(trimmed) || (maxLen > 0 && utf8.RuneCountInString(trimmed) > maxLen) {
		return nil, eferrors.WithField(invalidErr, field, "exceeds length bounds or contains control characters")
	}
	return &trimmed, nil
}

// ValidateOptionalMultilineText memvalidasi pointer teks bebas (seperti notes atau alasan pajak) dengan mengizinkan baris baru sambil menolak karakter kontrol berbahaya lainnya.
func ValidateOptionalMultilineText(field string, raw *string, invalidErr error) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return nil, eferrors.WithField(invalidErr, field, "contains invalid control characters")
		}
	}
	return &trimmed, nil
}

// ValidateCountryCode memvalidasi kode negara ISO 3166-1 alpha-2 dan mengembalikan huruf kapital dua karakter.
func ValidateCountryCode(field, raw string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if !countryCodePattern.MatchString(normalized) {
		return "", eferrors.WithField(eferrors.ErrInvalidCountryCode, field, "country code must be 2 uppercase letters")
	}
	return normalized, nil
}

// NormalizeCountryCodeOrDefault menerapkan DefaultCountryCode bila kosong lalu memvalidasi format ISO 3166-1 alpha-2.
func NormalizeCountryCodeOrDefault(field, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultCountryCode, nil
	}
	return ValidateCountryCode(field, trimmed)
}

// ValidateFiscalMonth memastikan bulan awal tahun buku berada pada rentang 1 sampai 12.
func ValidateFiscalMonth(month int16) error {
	if month < 1 || month > 12 {
		return eferrors.ErrInvalidFiscalMonth
	}
	return nil
}

// ValidateEmail memvalidasi format alamat surel bisnis dalam batas panjang PostgreSQL.
func ValidateEmail(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" || utf8.RuneCountInString(normalized) > MaxEmailLength {
		return "", eferrors.ErrInvalidEmail
	}
	if strings.ContainsAny(normalized, " \t\r\n<>") {
		return "", eferrors.ErrInvalidEmail
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized || parsed.Name != "" {
		return "", eferrors.ErrInvalidEmail
	}
	at := strings.LastIndex(normalized, "@")
	if at <= 0 || at == len(normalized)-1 {
		return "", eferrors.ErrInvalidEmail
	}
	domain := normalized[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return "", eferrors.ErrInvalidEmail
	}
	return normalized, nil
}

// ValidateOptionalEmail menormalkan dan memvalidasi alamat surel opsional; nilai kosong menjadi nil.
func ValidateOptionalEmail(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	if strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	normalized, err := ValidateEmail(*raw)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

// ValidateOptionalPhone memvalidasi format nomor telepon operasional atau mitra; nilai kosong menjadi nil.
func ValidateOptionalPhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxCodeLength || HasControlChar(trimmed) {
		return nil, eferrors.ErrInvalidPhone
	}
	digits := 0
	for _, r := range trimmed {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+', r == '-', r == ' ', r == '(', r == ')':
			// Karakter pemisah nomor telepon yang sah.
		default:
			return nil, eferrors.ErrInvalidPhone
		}
	}
	if digits < 6 {
		return nil, eferrors.ErrInvalidPhone
	}
	return &trimmed, nil
}

// ValidateOptionalURL memvalidasi URL http/https opsional dalam batas panjang skema; nilai kosong menjadi nil.
func ValidateOptionalURL(field string, raw *string, maxLen int) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxLen || hasControlOrWhitespace(trimmed) {
		return nil, eferrors.WithField(eferrors.ErrInvalidURL, field, "url exceeds bounds or contains invalid characters")
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Host == "" {
		return nil, eferrors.WithField(eferrors.ErrInvalidURL, field, "url must be valid http or https")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, eferrors.WithField(eferrors.ErrInvalidURL, field, "url scheme must be http or https")
	}
	return &trimmed, nil
}

// ValidateTimezone memastikan nama zona waktu IANA valid dan tidak melebihi batas panjang skema; string kosong memakai DefaultTimezone.
func ValidateTimezone(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultTimezone, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxTimezoneLength || hasControlOrWhitespace(trimmed) {
		return "", eferrors.ErrInvalidTimezone
	}
	if _, err := time.LoadLocation(trimmed); err != nil {
		return "", eferrors.ErrInvalidTimezone
	}
	return trimmed, nil
}

// ValidateLanguageCode menormalkan dan memvalidasi kode bahasa komunikasi mitra; string kosong memakai DefaultPreferredLanguage.
func ValidateLanguageCode(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return DefaultPreferredLanguage, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxPreferredLanguageLength || !languagePattern.MatchString(trimmed) {
		return "", eferrors.NewValidation("preferred_language", eferrors.Code("enterprise_foundation.party.invalid_language"), "invalid preferred language code")
	}
	return strings.ToLower(trimmed), nil
}

// ValidateJSONObject memastikan slice JSON merepresentasikan objek JSON yang sah (atau {} bila kosong).
func ValidateJSONObject(field string, raw []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return []byte("{}"), nil
	}
	var obj map[string]any
	if err := json.Unmarshal(trimmed, &obj); err != nil || obj == nil {
		return nil, eferrors.WithField(eferrors.ErrInvalidOutboxPayload, field, "must be a valid json object")
	}
	return bytes.Clone(trimmed), nil
}

// HasControlChar memeriksa apakah string mengandung karakter kontrol Unicode.
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

func normalizeCalendarDatePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	year, month, day := t.Date()
	v := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &v
}

func cloneStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}

func cloneFloat64Ptr(f *float64) *float64 {
	if f == nil {
		return nil
	}
	v := *f
	return &v
}

func equalUUIDPtr(a *string, b string) bool {
	if a == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*a), strings.TrimSpace(b))
}
