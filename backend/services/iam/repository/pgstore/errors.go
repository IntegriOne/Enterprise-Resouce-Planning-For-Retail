package pgstore

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// Pemetaan nama constraint unik ke sentinel error domain IAM.
var uniqueConstraintMap = map[string]error{
	"uq_user_identities_kind_value":                  iamerrors.ErrIdentityAlreadyExists,
	"uq_user_identities_one_pending_email_per_user":  iamerrors.ErrIdentityLimitExceeded,
	"uq_user_identities_one_verified_email_per_user": iamerrors.ErrIdentityLimitExceeded,
	"uq_user_identities_one_username_per_user":       iamerrors.ErrIdentityLimitExceeded,
	"user_credentials_user_id_key":                   iamerrors.ErrCredentialAlreadyExists,
	"user_totp_factors_user_id_key":                  iamerrors.ErrTOTPAlreadyExists,
	"roles_code_key":                                 iamerrors.ErrRoleAlreadyExists,
	"permissions_code_key":                           iamerrors.ErrPermissionAlreadyExists,
	"uq_permissions_resource_action":                 iamerrors.ErrPermissionAlreadyExists,
	"pk_role_permissions":                            iamerrors.ErrRolePermissionExists,
	"groups_code_key":                                iamerrors.ErrGroupAlreadyExists,
	"uq_group_members_group_user":                    iamerrors.ErrGroupMemberExists,
	"uq_user_roles_assignment":                       iamerrors.ErrUserRoleExists,
	"uq_group_roles_assignment":                      iamerrors.ErrGroupRoleExists,
	"user_recovery_codes_code_hash_key":              iamerrors.ErrAlreadyExists,
	"user_tokens_token_digest_key":                   iamerrors.ErrAlreadyExists,
	"uq_user_tokens_one_active_per_purpose":          iamerrors.ErrAlreadyExists,
	"sessions_token_digest_key":                      iamerrors.ErrAlreadyExists,
}

// Pemetaan nama constraint check ke sentinel error domain IAM.
var checkConstraintMap = map[string]error{
	"ck_user_roles_scope":        iamerrors.ErrInvalidScope,
	"ck_group_roles_scope":       iamerrors.ErrInvalidScope,
	"ck_roles_redirect_priority": iamerrors.ErrInvalidRedirectPriority,
}

// MapError menerjemahkan error GORM/PostgreSQL menjadi DomainError terstruktur dengan aman tanpa membocorkan detail SQL internal.
func MapError(err, notFoundErr error) error {
	return mapPGError(err, notFoundErr, nil, false)
}

// MapDeleteError menerjemahkan error operasi penghapusan termasuk pelanggaran FK RESTRICT menjadi konflik domain bila relasi masih dipakai.
func MapDeleteError(err, notFoundErr, inUseErr error) error {
	return mapPGError(err, notFoundErr, inUseErr, true)
}

func mapPGError(err, notFoundErr, inUseErr error, isDelete bool) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*iamerrors.DomainError](err); ok {
		return err
	}
	if errors.Is(err, context.Canceled) {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "database operation cancelled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "database operation timed out")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if notFoundErr != nil {
			return notFoundErr
		}
		return iamerrors.ErrNotFound
	}

	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "database operation failed")
	}

	return mapSQLState(pgErr, inUseErr, isDelete)
}

func mapSQLState(pgErr *pgconn.PgError, inUseErr error, isDelete bool) error {
	switch pgErr.Code {
	case "23505": // unique_violation
		if pgErr.ConstraintName == "uq_user_identities_kind_value" {
			return mapIdentityUniqueConstraint(pgErr.Detail)
		}
		if mapped, ok := uniqueConstraintMap[pgErr.ConstraintName]; ok {
			return mapped
		}
		return iamerrors.ErrAlreadyExists
	case "23503", "23001": // foreign_key_violation, restrict_violation
		if isDelete && inUseErr != nil {
			return inUseErr
		}
		return mapForeignKeyConstraint(pgErr.ConstraintName)
	case "23514": // check_violation
		if mapped, ok := checkConstraintMap[pgErr.ConstraintName]; ok {
			return mapped
		}
		return iamerrors.ErrInvalidInput
	case "22P02", "23502", "22001", "22003": // invalid_text_representation, not_null_violation, string_data_right_truncation, numeric_value_out_of_range
		return iamerrors.ErrInvalidInput
	case "40001", "40P01": // serialization_failure, deadlock_detected
		return iamerrors.Wrap(pgErr, iamerrors.KindConflict, iamerrors.ErrConflict.Code, "concurrent database transaction conflict")
	case "08000", "08001", "08003", "08004", "08006", "57P01", "57P02", "57P03", "57014", "53300":
		return iamerrors.Wrap(pgErr, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "database service is unavailable")
	default:
		if strings.HasPrefix(pgErr.Code, "08") || strings.HasPrefix(pgErr.Code, "53") || strings.HasPrefix(pgErr.Code, "57") {
			return iamerrors.Wrap(pgErr, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "database service is unavailable")
		}
		return iamerrors.Wrap(pgErr, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, "unexpected database error")
	}
}

func mapForeignKeyConstraint(constraint string) error {
	switch {
	case strings.Contains(constraint, "role_id_fkey"):
		return iamerrors.ErrRoleNotFound
	case strings.Contains(constraint, "permission_id_fkey"):
		return iamerrors.ErrPermissionNotFound
	case strings.Contains(constraint, "group_id_fkey"):
		return iamerrors.ErrGroupNotFound
	case strings.Contains(constraint, "identity_id_fkey"):
		return iamerrors.ErrIdentityNotFound
	case strings.Contains(constraint, "user_id_fkey"),
		strings.Contains(constraint, "assigned_by_fkey"),
		strings.Contains(constraint, "added_by_fkey"):
		return iamerrors.ErrUserNotFound
	default:
		return iamerrors.ErrNotFound
	}
}

func mapIdentityUniqueConstraint(detail string) error {
	lower := strings.ToLower(detail)
	switch {
	case strings.Contains(lower, "(username,"):
		return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "username", "username has been taken")
	case strings.Contains(lower, "(email,"):
		return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "email", "email is already registered")
	default:
		return iamerrors.ErrIdentityAlreadyExists
	}
}
