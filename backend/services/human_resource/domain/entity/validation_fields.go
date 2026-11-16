package entity

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// Validatable adalah kontrak entity yang diterima aturan domain sebelum persistensi.
type Validatable interface{ Validate() error }

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidateUUID memeriksa UUID logis tanpa menganggap keberadaan ID sebagai bukti akses.
func ValidateUUID(field, value string) error {
	if !uuidPattern.MatchString(value) || strings.EqualFold(value, "00000000-0000-0000-0000-000000000000") {
		return hrerrors.Invalid(field, "must be a nonzero uuid")
	}
	return nil
}

func validateText(field, value string, required bool, maxLength, exactLength int) error {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return hrerrors.Invalid(field, "must be valid text")
	}
	size := utf8.RuneCountInString(value)
	if required && strings.TrimSpace(value) == "" {
		return hrerrors.Invalid(field, "must not be blank")
	}
	if maxLength > 0 && size > maxLength {
		return hrerrors.Invalid(field, "exceeds maximum length")
	}
	if exactLength > 0 && size != exactLength {
		return hrerrors.Invalid(field, "has invalid length")
	}
	return nil
}

func validateTime(field string, value time.Time) error {
	if value.IsZero() || value.Year() < 1 || value.Year() > 9999 {
		return hrerrors.Invalid(field, "must be a valid date or timestamp")
	}
	return nil
}
func validateJSON(field string, value JSON, required bool) error {
	if value == nil {
		if required {
			return hrerrors.Invalid(field, "must not be null")
		}
		return nil
	}
	if !json.Valid(value) {
		return hrerrors.Invalid(field, "must be valid json")
	}
	return nil
}
func validateBytes(field string, value []byte, required bool) error {
	if value == nil && !required {
		return nil
	}
	if len(value) == 0 {
		return hrerrors.Invalid(field, "must contain evidence bytes")
	}
	return nil
}

func validateEnum(field string, valid bool) error {
	if !valid {
		return hrerrors.Invalid(field, "has invalid enum value")
	}
	return nil
}
func validateIdentity(field string, value int64, generated bool) error {
	if value < 0 || (!generated && value == 0) {
		return hrerrors.Invalid(field, "must be a positive identity")
	}
	return nil
}
