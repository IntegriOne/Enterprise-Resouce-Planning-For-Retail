package entity_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

func TestUserKindEnum(t *testing.T) {
	t.Parallel()

	valid := []entity.UserKind{
		entity.UserKindCustomer,
		entity.UserKindEmployee,
		entity.UserKindVendor,
		entity.UserKindSupplier,
		entity.UserKindSystem,
	}
	for _, k := range valid {
		parsed, err := entity.ParseUserKind(k.String())
		if err != nil {
			t.Fatalf("ParseUserKind(%q) unexpected error: %v", k.String(), err)
		}
		if parsed != k || !parsed.IsValid() {
			t.Fatalf("ParseUserKind(%q) = %q, valid=%v", k.String(), parsed, parsed.IsValid())
		}
	}

	if _, err := entity.ParseUserKind("partner"); !errors.Is(err, iamerrors.ErrInvalidUserKind) {
		t.Fatalf("ParseUserKind(partner) error = %v, want ErrInvalidUserKind", err)
	}
}

func TestUserStatusEnum(t *testing.T) {
	t.Parallel()

	valid := []entity.UserStatus{
		entity.UserStatusPending,
		entity.UserStatusActive,
		entity.UserStatusSuspended,
		entity.UserStatusDisabled,
	}
	for _, s := range valid {
		parsed, err := entity.ParseUserStatus(s.String())
		if err != nil {
			t.Fatalf("ParseUserStatus(%q) unexpected error: %v", s.String(), err)
		}
		if parsed != s || !parsed.IsValid() {
			t.Fatalf("ParseUserStatus(%q) = %q, valid=%v", s.String(), parsed, parsed.IsValid())
		}
	}

	if _, err := entity.ParseUserStatus("deleted"); !errors.Is(err, iamerrors.ErrInvalidUserStatus) {
		t.Fatalf("ParseUserStatus(deleted) error = %v, want ErrInvalidUserStatus", err)
	}
}

func TestIdentityKindEnum(t *testing.T) {
	t.Parallel()

	for _, k := range []entity.IdentityKind{entity.IdentityKindEmail, entity.IdentityKindUsername} {
		parsed, err := entity.ParseIdentityKind(k.String())
		if err != nil || parsed != k {
			t.Fatalf("ParseIdentityKind(%q) = (%q, %v), want %q", k, parsed, err, k)
		}
	}

	if _, err := entity.ParseIdentityKind("phone"); !errors.Is(err, iamerrors.ErrInvalidIdentityKind) {
		t.Fatalf("ParseIdentityKind(phone) error = %v, want ErrInvalidIdentityKind", err)
	}
}

func TestTokenPurposeAndSessionEndReasonEnums(t *testing.T) {
	t.Parallel()

	purposes := []entity.TokenPurpose{
		entity.TokenPurposeEmailVerification,
		entity.TokenPurposePasswordReset,
		entity.TokenPurposeAccountInvitation,
		entity.TokenPurposeMFAChallenge,
	}
	for _, p := range purposes {
		parsed, err := entity.ParseTokenPurpose(p.String())
		if err != nil || parsed != p {
			t.Fatalf("ParseTokenPurpose(%q) = (%q, %v)", p, parsed, err)
		}
	}
	if _, err := entity.ParseTokenPurpose("magic_link"); !errors.Is(err, iamerrors.ErrTokenInvalidPurpose) {
		t.Fatalf("ParseTokenPurpose(magic_link) error = %v, want ErrTokenInvalidPurpose", err)
	}

	reasons := []entity.SessionEndReason{
		entity.SessionEndLogout,
		entity.SessionEndIdleTimeout,
		entity.SessionEndAbsoluteTimeout,
		entity.SessionEndPasswordChanged,
		entity.SessionEndAdminRevoked,
		entity.SessionEndUserDisabled,
	}
	for _, r := range reasons {
		parsed, err := entity.ParseSessionEndReason(r.String())
		if err != nil || parsed != r {
			t.Fatalf("ParseSessionEndReason(%q) = (%q, %v)", r, parsed, err)
		}
	}
	if _, err := entity.ParseSessionEndReason("unknown"); !errors.Is(err, iamerrors.ErrInvalidSessionEndReason) {
		t.Fatalf("ParseSessionEndReason(unknown) error = %v, want ErrInvalidSessionEndReason", err)
	}
}

func TestScopeTypeEnum(t *testing.T) {
	t.Parallel()

	scopes := []entity.ScopeType{
		entity.ScopeTypeGlobal,
		entity.ScopeTypeLegalEntity,
		entity.ScopeTypeOrganizationUnit,
		entity.ScopeTypeFacility,
		entity.ScopeTypeSupplierParty,
		entity.ScopeTypeVendorParty,
		entity.ScopeTypeSelf,
	}
	for _, s := range scopes {
		parsed, err := entity.ParseScopeType(s.String())
		if err != nil || parsed != s {
			t.Fatalf("ParseScopeType(%q) = (%q, %v)", s, parsed, err)
		}
	}
	if _, err := entity.ParseScopeType("tenant"); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("ParseScopeType(tenant) error = %v, want ErrInvalidScope", err)
	}
}
