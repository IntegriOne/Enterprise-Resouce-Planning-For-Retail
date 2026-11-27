package pgstore

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Pemetaan nama unique index/constraint PostgreSQL Enterprise Foundation ke sentinel error domain.
var uniqueConstraintMap = map[string]error{
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
}

// Pemetaan nama check constraint PostgreSQL Enterprise Foundation ke sentinel error domain.
var checkConstraintMap = map[string]error{
	"ck_addresses_line_1_nonempty":                 eferrors.ErrInvalidAddressLine,
	"ck_addresses_city_nonempty":                   eferrors.ErrInvalidCity,
	"ck_addresses_country_code_len":                eferrors.ErrInvalidCountryCode,
	"ck_addresses_latitude":                        eferrors.ErrInvalidCoordinates,
	"ck_addresses_longitude":                       eferrors.ErrInvalidCoordinates,
	"ck_legal_entities_code_nonempty":              eferrors.ErrInvalidLegalEntityCode,
	"ck_legal_entities_legal_name_nonempty":        eferrors.ErrInvalidLegalEntityName,
	"ck_legal_entities_country_code_len":           eferrors.ErrInvalidCountryCode,
	"ck_legal_entities_tax_id_nonempty":            eferrors.ErrInvalidInput,
	"ck_legal_entities_not_self_parent":            eferrors.ErrLegalEntitySelfParent,
	"ck_legal_entities_fiscal_month":               eferrors.ErrInvalidFiscalMonth,
	"ck_legal_entities_dates":                      eferrors.ErrInvalidEntityDates,
	"ck_locations_code_nonempty":                   eferrors.ErrInvalidLocationCode,
	"ck_locations_name_nonempty":                   eferrors.ErrInvalidLocationName,
	"ck_locations_timezone_nonempty":               eferrors.ErrInvalidTimezone,
	"ck_locations_not_self_parent":                 eferrors.ErrLocationSelfParent,
	"ck_organization_units_code_nonempty":          eferrors.ErrInvalidOrgUnitCode,
	"ck_organization_units_name_nonempty":          eferrors.ErrInvalidOrgUnitName,
	"ck_organization_units_not_self_parent":        eferrors.ErrOrgUnitSelfParent,
	"ck_parties_partner_code_nonempty":             eferrors.ErrInvalidPartyCode,
	"ck_parties_display_name_nonempty":             eferrors.ErrInvalidPartyName,
	"ck_parties_tax_id_nonempty":                   eferrors.ErrInvalidInput,
	"ck_parties_tax_exemption_reason":              eferrors.ErrTaxExemptionReasonRequired,
	"ck_party_contacts_full_name_nonempty":         eferrors.ErrInvalidPartyName,
	"ck_party_contacts_contact_role_nonempty":      eferrors.ErrInvalidInput,
	"ck_party_contacts_has_channel_or_user":        eferrors.ErrContactChannelRequired,
	"ck_party_contacts_portal_link_invariant":      eferrors.ErrPortalLinkInvariantViolation,
	"ck_foundation_outbox_event_type_nonempty":     eferrors.ErrInvalidInput,
	"ck_foundation_outbox_aggregate_type_nonempty": eferrors.ErrInvalidInput,
	"ck_foundation_outbox_attempt_count":           eferrors.ErrInvalidInput,
}

// MapError menerjemahkan error GORM/PostgreSQL menjadi DomainError terstruktur tanpa membocorkan detail SQL internal.
func MapError(err, notFoundErr error) error {
	return mapPGError(err, notFoundErr, nil, false)
}

// MapDeleteError menerjemahkan error operasi penghapusan termasuk pelanggaran FK RESTRICT menjadi konflik domain saat relasi masih digunakan.
func MapDeleteError(err, notFoundErr, inUseErr error) error {
	return mapPGError(err, notFoundErr, inUseErr, true)
}

func mapPGError(err, notFoundErr, inUseErr error, isDelete bool) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*eferrors.DomainError](err); ok {
		return err
	}
	if errors.Is(err, context.Canceled) {
		return eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database operation cancelled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database operation timed out")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if notFoundErr != nil {
			return notFoundErr
		}
		return eferrors.ErrNotFound
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return eferrors.Wrap(err, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database operation failed")
	}

	return mapSQLState(pgErr, inUseErr, isDelete)
}

func mapSQLState(pgErr *pgconn.PgError, inUseErr error, isDelete bool) error {
	switch pgErr.Code {
	case "23505": // unique_violation
		if mapped, ok := uniqueConstraintMap[pgErr.ConstraintName]; ok {
			return mapped
		}
		return eferrors.ErrAlreadyExists
	case "23503", "23001": // foreign_key_violation, restrict_violation
		if isDelete && inUseErr != nil {
			return inUseErr
		}
		return mapForeignKeyConstraint(pgErr.ConstraintName)
	case "23514": // check_violation
		if mapped, ok := checkConstraintMap[pgErr.ConstraintName]; ok {
			return mapped
		}
		return eferrors.ErrInvalidInput
	case "22P02", "23502", "22001": // invalid_text_representation, not_null_violation, string_data_right_truncation
		return eferrors.ErrInvalidInput
	case "40001", "40P01": // serialization_failure, deadlock_detected
		return eferrors.Wrap(pgErr, eferrors.KindConflict, eferrors.ErrConflict.Code, "concurrent database transaction conflict")
	case "08000", "08001", "08003", "08004", "08006", "57P01", "57P02", "57P03", "57014", "53300":
		return eferrors.Wrap(pgErr, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database service is unavailable")
	default:
		if strings.HasPrefix(pgErr.Code, "08") || strings.HasPrefix(pgErr.Code, "53") || strings.HasPrefix(pgErr.Code, "57") {
			return eferrors.Wrap(pgErr, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database service is unavailable")
		}
		return eferrors.Wrap(pgErr, eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "unexpected database error")
	}
}

func mapForeignKeyConstraint(constraint string) error {
	switch {
	case strings.Contains(constraint, "registered_address_id_fkey"),
		strings.Contains(constraint, "address_id_fkey"):
		return eferrors.ErrAddressNotFound
	case strings.Contains(constraint, "parent_legal_entity_id_fkey"),
		strings.Contains(constraint, "legal_entity_id_fkey"):
		return eferrors.ErrLegalEntityNotFound
	case strings.Contains(constraint, "parent_location_id_fkey"),
		strings.Contains(constraint, "primary_location_id_fkey"):
		return eferrors.ErrLocationNotFound
	case strings.Contains(constraint, "parent_unit_id_fkey"):
		return eferrors.ErrOrgUnitNotFound
	case strings.Contains(constraint, "party_id_fkey"):
		return eferrors.ErrPartyNotFound
	default:
		return eferrors.ErrNotFound
	}
}
