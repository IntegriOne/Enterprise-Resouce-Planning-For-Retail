package eferrors_test

import (
	"errors"
	"strings"
	"testing"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestSentinels_BaseMatchingAndUniqueCodes(t *testing.T) {
	t.Parallel()

	sentinels := []*eferrors.DomainError{
		eferrors.ErrAddressNotFound,
		eferrors.ErrAddressInUse,
		eferrors.ErrInvalidAddressLine,
		eferrors.ErrInvalidCity,
		eferrors.ErrInvalidCountryCode,
		eferrors.ErrInvalidCoordinates,
		eferrors.ErrInvalidAddressUsage,
		eferrors.ErrLegalEntityNotFound,
		eferrors.ErrLegalEntityDeleted,
		eferrors.ErrLegalEntityInactive,
		eferrors.ErrLegalEntityCodeExists,
		eferrors.ErrLegalEntityTaxIDExists,
		eferrors.ErrPrimaryLegalEntityExists,
		eferrors.ErrInvalidLegalEntityCode,
		eferrors.ErrInvalidLegalEntityName,
		eferrors.ErrInvalidLegalEntityStatus,
		eferrors.ErrLegalEntitySelfParent,
		eferrors.ErrLegalEntityCycleDetected,
		eferrors.ErrInvalidFiscalMonth,
		eferrors.ErrInvalidEntityDates,
		eferrors.ErrLegalEntityHasDependents,
		eferrors.ErrLocationNotFound,
		eferrors.ErrLocationDeleted,
		eferrors.ErrLocationInactive,
		eferrors.ErrLocationCodeExists,
		eferrors.ErrInvalidLocationCode,
		eferrors.ErrInvalidLocationName,
		eferrors.ErrInvalidLocationKind,
		eferrors.ErrInvalidTimezone,
		eferrors.ErrLocationSelfParent,
		eferrors.ErrLocationParentMismatch,
		eferrors.ErrLocationCycleDetected,
		eferrors.ErrOrgUnitNotFound,
		eferrors.ErrOrgUnitDeleted,
		eferrors.ErrOrgUnitInactive,
		eferrors.ErrOrgUnitCodeExists,
		eferrors.ErrInvalidOrgUnitCode,
		eferrors.ErrInvalidOrgUnitName,
		eferrors.ErrInvalidOrgUnitKind,
		eferrors.ErrOrgUnitSelfParent,
		eferrors.ErrOrgUnitParentMismatch,
		eferrors.ErrOrgUnitLocationMismatch,
		eferrors.ErrOrgUnitCycleDetected,
		eferrors.ErrPartyNotFound,
		eferrors.ErrPartyDeleted,
		eferrors.ErrPartyInactive,
		eferrors.ErrPartyCodeExists,
		eferrors.ErrPartyTaxIDExists,
		eferrors.ErrInvalidPartyCode,
		eferrors.ErrInvalidPartyName,
		eferrors.ErrInvalidPartyKind,
		eferrors.ErrInvalidPartyStatus,
		eferrors.ErrTaxExemptionReasonRequired,
		eferrors.ErrPartyRoleNotFound,
		eferrors.ErrPartyRoleExists,
		eferrors.ErrPartyRoleInactive,
		eferrors.ErrInvalidPartyRoleKind,
		eferrors.ErrPartyContactNotFound,
		eferrors.ErrPartyContactDeleted,
		eferrors.ErrPartyContactInactive,
		eferrors.ErrPartyContactIAMUserExists,
		eferrors.ErrContactChannelRequired,
		eferrors.ErrInvalidPortalAccessStatus,
		eferrors.ErrPortalLinkInvariantViolation,
		eferrors.ErrPortalAccessNotActive,
		eferrors.ErrPartyAddressNotFound,
		eferrors.ErrPartyAddressExists,
		eferrors.ErrPrimaryPartyAddressConflict,
		eferrors.ErrOutboxEventNotFound,
		eferrors.ErrInvalidOutboxStatus,
		eferrors.ErrInvalidOutboxPayload,
		eferrors.ErrInvalidEmail,
		eferrors.ErrInvalidPhone,
		eferrors.ErrInvalidURL,
		eferrors.ErrInvalidScopeTarget,
		eferrors.ErrScopeTargetNotEligible,
	}

	seenCodes := make(map[eferrors.Code]struct{}, len(sentinels))
	for _, s := range sentinels {
		if s == nil {
			t.Fatal("encountered nil sentinel")
		}
		if s.Code == "" || !strings.HasPrefix(string(s.Code), "enterprise_foundation.") {
			t.Fatalf("sentinel code %q must start with enterprise_foundation.", s.Code)
		}
		if _, dup := seenCodes[s.Code]; dup {
			t.Fatalf("duplicate sentinel code detected: %q", s.Code)
		}
		seenCodes[s.Code] = struct{}{}

		if msg := s.Error(); msg == "" || msg != strings.ToLower(msg) {
			t.Fatalf("sentinel %q message %q must be non-empty lowercase", s.Code, msg)
		}

		// Pastikan sentinel bukan base error agar tidak saling cocok lintas kode berbeda.
		for _, other := range sentinels {
			if s != other && errors.Is(s, other) {
				t.Fatalf("sentinel %q unexpectedly matched distinct sentinel %q", s.Code, other.Code)
			}
		}
	}

	cases := []struct {
		name string
		err  error
		base error
	}{
		{name: "not found", err: eferrors.ErrPartyNotFound, base: eferrors.ErrNotFound},
		{name: "already exists", err: eferrors.ErrLegalEntityCodeExists, base: eferrors.ErrAlreadyExists},
		{name: "invalid input", err: eferrors.ErrPortalLinkInvariantViolation, base: eferrors.ErrInvalidInput},
		{name: "conflict", err: eferrors.ErrLegalEntityCycleDetected, base: eferrors.ErrConflict},
		{name: "forbidden", err: eferrors.ErrPortalAccessNotActive, base: eferrors.ErrForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !errors.Is(tc.err, tc.base) {
				t.Fatalf("expected %v to match %v", tc.err, tc.base)
			}
		})
	}
}

func TestFieldOfExtraction(t *testing.T) {
	t.Parallel()

	withField := eferrors.WithField(eferrors.ErrInvalidCountryCode, "jurisdiction_country_code", "must be ID or ISO alpha-2")
	if got := eferrors.FieldOf(withField); got != "jurisdiction_country_code" {
		t.Fatalf("FieldOf() = %q, want jurisdiction_country_code", got)
	}
	if !errors.Is(withField, eferrors.ErrInvalidCountryCode) {
		t.Fatal("WithField error must still match wrapped sentinel")
	}
	if !errors.Is(withField, eferrors.ErrInvalidInput) {
		t.Fatal("WithField error must still match ErrInvalidInput category")
	}
	if got := eferrors.FieldOf(errors.New("plain error")); got != "" {
		t.Fatalf("FieldOf(plain) = %q, want empty string", got)
	}
}
