package entity_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestValidationHelpers(t *testing.T) {
	t.Parallel()

	t.Run("UUID", func(t *testing.T) {
		t.Parallel()
		if !entity.IsValidUUID(" 11111111-1111-4111-8111-AAAAAAAAAAAA ") {
			t.Fatal("expected uppercase padded UUID to be valid")
		}
		if got := entity.NormalizeUUID(" 11111111-1111-4111-8111-AAAAAAAAAAAA "); got != "11111111-1111-4111-8111-aaaaaaaaaaaa" {
			t.Fatalf("NormalizeUUID() = %q", got)
		}
		if err := entity.ValidateUUID("id", "not-a-uuid"); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
		if opt, err := entity.ValidateOptionalUUID("parent_id", nil); err != nil || opt != nil {
			t.Fatalf("ValidateOptionalUUID(nil) = (%v, %v)", opt, err)
		}
		if _, err := entity.ValidateOptionalUUID("parent_id", strPtr("")); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Fatalf("expected error on non-nil empty UUID pointer, got %v", err)
		}
	})

	t.Run("CountryAndLanguageAndTimezone", func(t *testing.T) {
		t.Parallel()
		if got, err := entity.NormalizeCountryCodeOrDefault("country_code", ""); err != nil || got != "ID" {
			t.Fatalf("NormalizeCountryCodeOrDefault empty = (%q, %v)", got, err)
		}
		if got, err := entity.NormalizeCountryCodeOrDefault("country_code", " sg "); err != nil || got != "SG" {
			t.Fatalf("NormalizeCountryCodeOrDefault sg = (%q, %v)", got, err)
		}
		if _, err := entity.NormalizeCountryCodeOrDefault("country_code", "IDN"); !errors.Is(err, eferrors.ErrInvalidCountryCode) {
			t.Fatalf("expected ErrInvalidCountryCode, got %v", err)
		}

		if got, err := entity.ValidateTimezone(""); err != nil || got != entity.DefaultTimezone {
			t.Fatalf("ValidateTimezone empty = (%q, %v)", got, err)
		}
		if got, err := entity.ValidateTimezone("Asia/Makassar"); err != nil || got != "Asia/Makassar" {
			t.Fatalf("ValidateTimezone Asia/Makassar = (%q, %v)", got, err)
		}
		if _, err := entity.ValidateTimezone("Mars/Olympus"); !errors.Is(err, eferrors.ErrInvalidTimezone) {
			t.Fatalf("expected ErrInvalidTimezone, got %v", err)
		}

		if got, err := entity.ValidateLanguageCode(""); err != nil || got != entity.DefaultPreferredLanguage {
			t.Fatalf("ValidateLanguageCode empty = (%q, %v)", got, err)
		}
		if got, err := entity.ValidateLanguageCode("ID-id"); err != nil || got != "id-id" {
			t.Fatalf("ValidateLanguageCode ID-id = (%q, %v)", got, err)
		}
		if _, err := entity.ValidateLanguageCode("bad lang!"); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Fatalf("expected invalid input error, got %v", err)
		}
	})

	t.Run("EmailPhoneURLAndJSON", func(t *testing.T) {
		t.Parallel()
		email, err := entity.ValidateOptionalEmail(strPtr("  Vendor.PIC@Partner.CO.ID "))
		if err != nil || email == nil || *email != "vendor.pic@partner.co.id" {
			t.Fatalf("ValidateOptionalEmail = (%v, %v)", email, err)
		}
		if _, err := entity.ValidateOptionalEmail(strPtr("Name <user@example.com>")); !errors.Is(err, eferrors.ErrInvalidEmail) {
			t.Fatalf("expected ErrInvalidEmail for display-name email, got %v", err)
		}
		if nilPhone, err := entity.ValidateOptionalPhone(strPtr("   ")); err != nil || nilPhone != nil {
			t.Fatalf("ValidateOptionalPhone blank should return nil, got (%v, %v)", nilPhone, err)
		}
		if _, err := entity.ValidateOptionalPhone(strPtr("123")); !errors.Is(err, eferrors.ErrInvalidPhone) {
			t.Fatalf("expected ErrInvalidPhone for short phone, got %v", err)
		}
		if _, err := entity.ValidateOptionalURL("website_url", strPtr("ftp://files.example.com"), entity.MaxWebsiteURLLength); !errors.Is(err, eferrors.ErrInvalidURL) {
			t.Fatalf("expected ErrInvalidURL for ftp scheme, got %v", err)
		}
		if raw, err := entity.ValidateJSONObject("payload", nil); err != nil || string(raw) != "{}" {
			t.Fatalf("ValidateJSONObject(nil) = (%q, %v)", string(raw), err)
		}
		if _, err := entity.ValidateJSONObject("payload", []byte(`["not-an-object"]`)); !errors.Is(err, eferrors.ErrInvalidOutboxPayload) {
			t.Fatalf("expected ErrInvalidOutboxPayload for json array, got %v", err)
		}
	})
}
