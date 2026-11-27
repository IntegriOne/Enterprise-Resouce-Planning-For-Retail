package pgstore_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	t.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		if err := pgstore.MapError(nil, eferrors.ErrAddressNotFound); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("domain error passes through unchanged", func(t *testing.T) {
		t.Parallel()
		err := pgstore.MapError(eferrors.ErrPartyCodeExists, eferrors.ErrPartyNotFound)
		if !errors.Is(err, eferrors.ErrPartyCodeExists) {
			t.Fatalf("expected passthrough ErrPartyCodeExists, got %v", err)
		}
	})

	t.Run("context errors map to unavailable", func(t *testing.T) {
		t.Parallel()
		for _, ctxErr := range []error{context.Canceled, context.DeadlineExceeded} {
			mapped := pgstore.MapError(ctxErr, nil)
			if eferrors.KindOf(mapped) != eferrors.KindUnavailable {
				t.Fatalf("expected KindUnavailable for %v, got %v", ctxErr, mapped)
			}
		}
	})

	t.Run("record not found maps to custom or default sentinel", func(t *testing.T) {
		t.Parallel()
		if err := pgstore.MapError(gorm.ErrRecordNotFound, eferrors.ErrLegalEntityNotFound); !errors.Is(err, eferrors.ErrLegalEntityNotFound) {
			t.Fatalf("expected ErrLegalEntityNotFound, got %v", err)
		}
		if err := pgstore.MapError(gorm.ErrRecordNotFound, nil); !errors.Is(err, eferrors.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unique violations map to exact domain sentinels", func(t *testing.T) {
		t.Parallel()
		cases := map[string]error{
			"uq_legal_entities_code":               eferrors.ErrLegalEntityCodeExists,
			"uq_legal_entities_primary_operating":  eferrors.ErrPrimaryLegalEntityExists,
			"uq_legal_entities_tax_id":             eferrors.ErrLegalEntityTaxIDExists,
			"uq_locations_code":                    eferrors.ErrLocationCodeExists,
			"uq_organization_units_entity_code":    eferrors.ErrOrgUnitCodeExists,
			"uq_parties_partner_code":              eferrors.ErrPartyCodeExists,
			"uq_parties_tax_id_number":             eferrors.ErrPartyTaxIDExists,
			"uq_party_roles_party_role":            eferrors.ErrPartyRoleExists,
			"uq_party_contacts_party_iam_user":     eferrors.ErrPartyContactIAMUserExists,
			"uq_party_addresses_party_addr_usage":  eferrors.ErrPartyAddressExists,
			"uq_party_addresses_primary_per_usage": eferrors.ErrPrimaryPartyAddressConflict,
			"uq_unknown_constraint":                eferrors.ErrAlreadyExists,
		}
		for constraint, want := range cases {
			mapped := pgstore.MapError(&pgconn.PgError{Code: "23505", ConstraintName: constraint}, nil)
			if !errors.Is(mapped, want) {
				t.Errorf("constraint %s: got %v, want %v", constraint, mapped, want)
			}
		}
	})

	t.Run("check violations map to exact validation sentinels", func(t *testing.T) {
		t.Parallel()
		cases := map[string]error{
			"ck_addresses_line_1_nonempty":            eferrors.ErrInvalidAddressLine,
			"ck_addresses_city_nonempty":              eferrors.ErrInvalidCity,
			"ck_addresses_country_code_len":           eferrors.ErrInvalidCountryCode,
			"ck_addresses_latitude":                   eferrors.ErrInvalidCoordinates,
			"ck_addresses_longitude":                  eferrors.ErrInvalidCoordinates,
			"ck_legal_entities_code_nonempty":         eferrors.ErrInvalidLegalEntityCode,
			"ck_legal_entities_legal_name_nonempty":   eferrors.ErrInvalidLegalEntityName,
			"ck_legal_entities_not_self_parent":       eferrors.ErrLegalEntitySelfParent,
			"ck_legal_entities_fiscal_month":          eferrors.ErrInvalidFiscalMonth,
			"ck_legal_entities_dates":                 eferrors.ErrInvalidEntityDates,
			"ck_locations_code_nonempty":              eferrors.ErrInvalidLocationCode,
			"ck_locations_name_nonempty":              eferrors.ErrInvalidLocationName,
			"ck_locations_timezone_nonempty":          eferrors.ErrInvalidTimezone,
			"ck_locations_not_self_parent":            eferrors.ErrLocationSelfParent,
			"ck_organization_units_code_nonempty":     eferrors.ErrInvalidOrgUnitCode,
			"ck_organization_units_name_nonempty":     eferrors.ErrInvalidOrgUnitName,
			"ck_organization_units_not_self_parent":   eferrors.ErrOrgUnitSelfParent,
			"ck_parties_partner_code_nonempty":        eferrors.ErrInvalidPartyCode,
			"ck_parties_display_name_nonempty":        eferrors.ErrInvalidPartyName,
			"ck_parties_tax_exemption_reason":         eferrors.ErrTaxExemptionReasonRequired,
			"ck_party_contacts_has_channel_or_user":   eferrors.ErrContactChannelRequired,
			"ck_party_contacts_portal_link_invariant": eferrors.ErrPortalLinkInvariantViolation,
			"ck_unknown": eferrors.ErrInvalidInput,
		}
		for constraint, want := range cases {
			mapped := pgstore.MapError(&pgconn.PgError{Code: "23514", ConstraintName: constraint}, nil)
			if !errors.Is(mapped, want) {
				t.Errorf("constraint %s: got %v, want %v", constraint, mapped, want)
			}
		}
	})

	t.Run("foreign key and delete restrict map appropriately", func(t *testing.T) {
		t.Parallel()
		fkCases := map[string]error{
			"locations_address_id_fkey":                   eferrors.ErrAddressNotFound,
			"legal_entities_registered_address_id_fkey":   eferrors.ErrAddressNotFound,
			"locations_legal_entity_id_fkey":              eferrors.ErrLegalEntityNotFound,
			"organization_units_primary_location_id_fkey": eferrors.ErrLocationNotFound,
			"organization_units_parent_unit_id_fkey":      eferrors.ErrOrgUnitNotFound,
			"party_roles_party_id_fkey":                   eferrors.ErrPartyNotFound,
		}
		for constraint, want := range fkCases {
			mapped := pgstore.MapError(&pgconn.PgError{Code: "23503", ConstraintName: constraint}, nil)
			if !errors.Is(mapped, want) {
				t.Errorf("fk %s: got %v, want %v", constraint, mapped, want)
			}
		}

		delErr := pgstore.MapDeleteError(&pgconn.PgError{Code: "23503", ConstraintName: "locations_address_id_fkey"}, eferrors.ErrAddressNotFound, eferrors.ErrAddressInUse)
		if !errors.Is(delErr, eferrors.ErrAddressInUse) {
			t.Fatalf("expected ErrAddressInUse on delete restrict, got %v", delErr)
		}
	})

	t.Run("serialization and unavailability sqlstates map safely", func(t *testing.T) {
		t.Parallel()
		conflictErr := pgstore.MapError(&pgconn.PgError{Code: "40001"}, nil)
		if eferrors.KindOf(conflictErr) != eferrors.KindConflict {
			t.Fatalf("expected KindConflict for 40001, got %v", conflictErr)
		}
		for _, state := range []string{"08006", "57P01", "53300", "22000"} {
			if state == "22000" {
				continue
			}
			mapped := pgstore.MapError(&pgconn.PgError{Code: state}, nil)
			if eferrors.KindOf(mapped) != eferrors.KindUnavailable {
				t.Fatalf("expected KindUnavailable for %s, got %v", state, mapped)
			}
		}
		for _, badState := range []string{"22P02", "23502", "22001"} {
			mapped := pgstore.MapError(&pgconn.PgError{Code: badState}, nil)
			if !errors.Is(mapped, eferrors.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput for %s, got %v", badState, mapped)
			}
		}
	})
}
