package entity_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

func TestValidateUUID(t *testing.T) {
	t.Parallel()

	valid := "550e8400-e29b-41d4-a716-446655440000"
	if err := entity.ValidateUUID("user_id", valid); err != nil {
		t.Fatalf("ValidateUUID(%q) unexpected error: %v", valid, err)
	}

	invalidCases := []string{
		"",
		"not-a-uuid",
		"550e8400e29b41d4a716446655440000",
		"550e8400-e29b-41d4-a716-44665544000g",
		"550e8400-e29b-41d4-a716-44665544000",
	}
	for _, tc := range invalidCases {
		err := entity.ValidateUUID("user_id", tc)
		if !errors.Is(err, iamerrors.ErrInvalidInput) {
			t.Fatalf("ValidateUUID(%q) = %v, want ErrInvalidInput", tc, err)
		}
		if got := iamerrors.FieldOf(err); got != "user_id" {
			t.Fatalf("ValidateUUID(%q) field = %q, want user_id", tc, got)
		}
	}
}

func TestNormalizeCode(t *testing.T) {
	t.Parallel()

	if got := entity.NormalizeCode("  Super_Admin "); got != "super_admin" {
		t.Fatalf("NormalizeCode() = %q, want super_admin", got)
	}
}

func TestValidateInternalRedirectPath(t *testing.T) {
	t.Parallel()

	validPaths := []string{
		"/",
		"/identity",
		"/human-resource",
		"/warehouse/bins",
		"/vendor/portal",
	}
	for _, p := range validPaths {
		if err := entity.ValidateInternalRedirectPath(p); err != nil {
			t.Fatalf("ValidateInternalRedirectPath(%q) unexpected error: %v", p, err)
		}
	}

	invalidPaths := []string{
		"",
		"identity",
		"//evil.example.com",
		"/\\evil.example.com",
		"https://evil.example.com/identity",
		"javascript:alert(1)",
		"/identity?next=https://evil.com",
		"/identity#frag",
		"/identity/../admin",
		"/identity\nSet-Cookie: a=b",
		"/" + strings.Repeat("a", 120),
	}
	for _, p := range invalidPaths {
		err := entity.ValidateInternalRedirectPath(p)
		if !errors.Is(err, iamerrors.ErrRedirectPathInvalid) {
			t.Fatalf("ValidateInternalRedirectPath(%q) = %v, want ErrRedirectPathInvalid", p, err)
		}
	}
}

func TestValidateOptionalURL(t *testing.T) {
	t.Parallel()

	for _, u := range []string{"", "https://cdn.integrione.local/avatar.png", "http://localhost:3333/a.png"} {
		if err := entity.ValidateOptionalURL("avatar_url", u, entity.MaxAvatarURLLength); err != nil {
			t.Fatalf("ValidateOptionalURL(%q) unexpected error: %v", u, err)
		}
	}

	for _, u := range []string{"javascript:alert(1)", "ftp://files.local/a.png", "https://cdn.local/" + strings.Repeat("a", 1000)} {
		if err := entity.ValidateOptionalURL("avatar_url", u, entity.MaxAvatarURLLength); !errors.Is(err, iamerrors.ErrInvalidAvatarURL) {
			t.Fatalf("ValidateOptionalURL(%q) = %v, want ErrInvalidAvatarURL", u, err)
		}
	}
}
