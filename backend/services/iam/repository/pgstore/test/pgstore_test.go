package pgstore_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		notFound   error
		wantTarget error
		wantKind   iamerrors.Kind
		wantNil    bool
	}{
		{
			name:    "nil error stays nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:       "existing domain error returned untouched",
			err:        iamerrors.ErrInvalidEmail,
			wantTarget: iamerrors.ErrInvalidEmail,
			wantKind:   iamerrors.KindInvalidInput,
		},
		{
			name:       "gorm record not found uses custom sentinel",
			err:        gorm.ErrRecordNotFound,
			notFound:   iamerrors.ErrUserNotFound,
			wantTarget: iamerrors.ErrUserNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "gorm record not found falls back to ErrNotFound",
			err:        gorm.ErrRecordNotFound,
			notFound:   nil,
			wantTarget: iamerrors.ErrNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "context deadline maps to unavailable",
			err:        context.DeadlineExceeded,
			wantTarget: iamerrors.ErrUnavailable,
			wantKind:   iamerrors.KindUnavailable,
		},
		{
			name:       "context canceled maps to unavailable",
			err:        context.Canceled,
			wantTarget: iamerrors.ErrUnavailable,
			wantKind:   iamerrors.KindUnavailable,
		},
		{
			name:       "unique identity violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "uq_user_identities_kind_value"},
			wantTarget: iamerrors.ErrIdentityAlreadyExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "unique role code violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "roles_code_key"},
			wantTarget: iamerrors.ErrRoleAlreadyExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "unique permission resource_action violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "uq_permissions_resource_action"},
			wantTarget: iamerrors.ErrPermissionAlreadyExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "unique group member violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "uq_group_members_group_user"},
			wantTarget: iamerrors.ErrGroupMemberExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "unique user role assignment violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "uq_user_roles_assignment"},
			wantTarget: iamerrors.ErrUserRoleExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "unique group role assignment violation",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "uq_group_roles_assignment"},
			wantTarget: iamerrors.ErrGroupRoleExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "fk missing user maps to ErrUserNotFound",
			err:        &pgconn.PgError{Code: "23503", ConstraintName: "sessions_user_id_fkey"},
			wantTarget: iamerrors.ErrUserNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "fk missing role maps to ErrRoleNotFound",
			err:        &pgconn.PgError{Code: "23503", ConstraintName: "user_roles_role_id_fkey"},
			wantTarget: iamerrors.ErrRoleNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "fk missing permission maps to ErrPermissionNotFound",
			err:        &pgconn.PgError{Code: "23503", ConstraintName: "role_permissions_permission_id_fkey"},
			wantTarget: iamerrors.ErrPermissionNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "fk missing group maps to ErrGroupNotFound",
			err:        &pgconn.PgError{Code: "23503", ConstraintName: "group_members_group_id_fkey"},
			wantTarget: iamerrors.ErrGroupNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "check scope violation maps to ErrInvalidScope",
			err:        &pgconn.PgError{Code: "23514", ConstraintName: "ck_user_roles_scope"},
			wantTarget: iamerrors.ErrInvalidScope,
			wantKind:   iamerrors.KindInvalidInput,
		},
		{
			name:       "check redirect priority maps to ErrInvalidRedirectPriority",
			err:        &pgconn.PgError{Code: "23514", ConstraintName: "ck_roles_redirect_priority"},
			wantTarget: iamerrors.ErrInvalidRedirectPriority,
			wantKind:   iamerrors.KindInvalidInput,
		},
		{
			name:       "invalid uuid text maps to ErrInvalidInput",
			err:        &pgconn.PgError{Code: "22P02"},
			wantTarget: iamerrors.ErrInvalidInput,
			wantKind:   iamerrors.KindInvalidInput,
		},
		{
			name:       "numeric out of range maps to ErrInvalidInput",
			err:        &pgconn.PgError{Code: "22003"},
			wantTarget: iamerrors.ErrInvalidInput,
			wantKind:   iamerrors.KindInvalidInput,
		},
		{
			name:       "duplicate token digest maps to ErrAlreadyExists",
			err:        &pgconn.PgError{Code: "23505", ConstraintName: "user_tokens_token_digest_key"},
			wantTarget: iamerrors.ErrAlreadyExists,
			wantKind:   iamerrors.KindAlreadyExists,
		},
		{
			name:       "fk missing identity maps to ErrIdentityNotFound",
			err:        &pgconn.PgError{Code: "23503", ConstraintName: "user_tokens_identity_id_fkey"},
			wantTarget: iamerrors.ErrIdentityNotFound,
			wantKind:   iamerrors.KindNotFound,
		},
		{
			name:       "deadlock maps to conflict",
			err:        &pgconn.PgError{Code: "40P01"},
			wantTarget: iamerrors.ErrConflict,
			wantKind:   iamerrors.KindConflict,
		},
		{
			name:       "connection failure maps to unavailable",
			err:        &pgconn.PgError{Code: "08006"},
			wantTarget: iamerrors.ErrUnavailable,
			wantKind:   iamerrors.KindUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pgstore.MapError(tc.err, tc.notFound)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !errors.Is(got, tc.wantTarget) {
				t.Fatalf("expected error matching %v, got %v", tc.wantTarget, got)
			}
			if iamerrors.KindOf(got) != tc.wantKind {
				t.Fatalf("expected kind %v, got %v", tc.wantKind, iamerrors.KindOf(got))
			}
		})
	}
}

func TestMapDeleteError(t *testing.T) {
	t.Parallel()

	inUseErr := iamerrors.New(iamerrors.KindConflict, "iam.role.in_use", "role is currently assigned")
	for _, code := range []string{"23503", "23001"} {
		pgErr := &pgconn.PgError{Code: code, ConstraintName: "user_roles_role_id_fkey"}

		got := pgstore.MapDeleteError(pgErr, iamerrors.ErrRoleNotFound, inUseErr)
		if !errors.Is(got, inUseErr) {
			t.Fatalf("code %s: expected inUseErr, got %v", code, got)
		}
		if iamerrors.KindOf(got) != iamerrors.KindConflict {
			t.Fatalf("code %s: expected conflict kind, got %v", code, iamerrors.KindOf(got))
		}
	}
}

func TestMapError_IdentityUniqueField(t *testing.T) {
	t.Parallel()

	unameErr := pgstore.MapError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uq_user_identities_kind_value",
		Detail:         "Key (kind, value)=(username, kibo) already exists.",
	}, nil)
	if domErr, ok := errors.AsType[*iamerrors.DomainError](unameErr); !ok || domErr.Field != "username" || domErr.Message != "username has been taken" {
		t.Fatalf("expected username field conflict, got %+v", unameErr)
	}

	emailErr := pgstore.MapError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uq_user_identities_kind_value",
		Detail:         "Key (kind, value)=(email, user@integrione.local) already exists.",
	}, nil)
	if domErr, ok := errors.AsType[*iamerrors.DomainError](emailErr); !ok || domErr.Field != "email" {
		t.Fatalf("expected email field conflict, got %+v", emailErr)
	}
}
