package iamerrors_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

func TestKindString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		kind iamerrors.Kind
		want string
	}{
		{name: "unknown", kind: iamerrors.KindUnknown, want: "unknown"},
		{name: "invalid input", kind: iamerrors.KindInvalidInput, want: "invalid_input"},
		{name: "not found", kind: iamerrors.KindNotFound, want: "not_found"},
		{name: "already exists", kind: iamerrors.KindAlreadyExists, want: "already_exists"},
		{name: "unauthorized", kind: iamerrors.KindUnauthorized, want: "unauthorized"},
		{name: "forbidden", kind: iamerrors.KindForbidden, want: "forbidden"},
		{name: "conflict", kind: iamerrors.KindConflict, want: "conflict"},
		{name: "rate limited", kind: iamerrors.KindRateLimited, want: "rate_limited"},
		{name: "unavailable", kind: iamerrors.KindUnavailable, want: "unavailable"},
		{name: "out of range", kind: iamerrors.Kind(255), want: "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.kind.String(); got != tc.want {
				t.Fatalf("Kind.String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBaseSentinelsAndDomainSentinels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      error
		base     error
		wantKind iamerrors.Kind
		wantCode iamerrors.Code
	}{
		{
			name:     "user not found matches base not found",
			err:      iamerrors.ErrUserNotFound,
			base:     iamerrors.ErrNotFound,
			wantKind: iamerrors.KindNotFound,
			wantCode: iamerrors.Code("iam.user.not_found"),
		},
		{
			name:     "identity already exists matches base already exists",
			err:      iamerrors.ErrIdentityAlreadyExists,
			base:     iamerrors.ErrAlreadyExists,
			wantKind: iamerrors.KindAlreadyExists,
			wantCode: iamerrors.Code("iam.identity.already_exists"),
		},
		{
			name:     "invalid credentials matches base unauthorized",
			err:      iamerrors.ErrInvalidCredentials,
			base:     iamerrors.ErrUnauthorized,
			wantKind: iamerrors.KindUnauthorized,
			wantCode: iamerrors.Code("iam.credential.invalid"),
		},
		{
			name:     "account locked matches base rate limited",
			err:      iamerrors.ErrAccountLocked,
			base:     iamerrors.ErrRateLimited,
			wantKind: iamerrors.KindRateLimited,
			wantCode: iamerrors.Code("iam.credential.locked"),
		},
		{
			name:     "session idle expired matches base unauthorized",
			err:      iamerrors.ErrSessionIdleExpired,
			base:     iamerrors.ErrUnauthorized,
			wantKind: iamerrors.KindUnauthorized,
			wantCode: iamerrors.Code("iam.session.idle_expired"),
		},
		{
			name:     "system role immutable matches base conflict",
			err:      iamerrors.ErrSystemRoleImmutable,
			base:     iamerrors.ErrConflict,
			wantKind: iamerrors.KindConflict,
			wantCode: iamerrors.Code("iam.role.system_immutable"),
		},
		{
			name:     "permission denied matches base forbidden",
			err:      iamerrors.ErrPermissionDenied,
			base:     iamerrors.ErrForbidden,
			wantKind: iamerrors.KindForbidden,
			wantCode: iamerrors.Code("iam.rbac.permission_denied"),
		},
		{
			name:     "invalid scope matches base invalid input",
			err:      iamerrors.ErrInvalidScope,
			base:     iamerrors.ErrInvalidInput,
			wantKind: iamerrors.KindInvalidInput,
			wantCode: iamerrors.Code("iam.rbac.invalid_scope"),
		},
		{
			name:     "base unavailable has unavailable kind",
			err:      iamerrors.ErrUnavailable,
			base:     iamerrors.ErrUnavailable,
			wantKind: iamerrors.KindUnavailable,
			wantCode: iamerrors.Code("iam.unavailable"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !errors.Is(tc.err, tc.base) {
				t.Fatalf("errors.Is(%v, %v) = false, want true", tc.err, tc.base)
			}
			if got := iamerrors.KindOf(tc.err); got != tc.wantKind {
				t.Fatalf("KindOf() = %v, want %v", got, tc.wantKind)
			}
			if got := iamerrors.CodeOf(tc.err); got != tc.wantCode {
				t.Fatalf("CodeOf() = %q, want %q", got, tc.wantCode)
			}
			msg := tc.err.Error()
			if msg == "" || msg != strings.ToLower(msg) {
				t.Fatalf("Error() = %q, want non-empty lowercase message", msg)
			}
			if strings.HasSuffix(msg, ".") || strings.HasSuffix(msg, "!") {
				t.Fatalf("Error() = %q, must not have trailing punctuation", msg)
			}
		})
	}
}

func TestSentinelIsolation(t *testing.T) {
	t.Parallel()

	// Pastikan sentinel berbeda tidak saling cocok meskipun Kind-nya sama.
	if errors.Is(iamerrors.ErrUserNotFound, iamerrors.ErrSessionNotFound) {
		t.Fatal("ErrUserNotFound must not match ErrSessionNotFound")
	}
	if errors.Is(iamerrors.ErrInvalidEmail, iamerrors.ErrInvalidScope) {
		t.Fatal("ErrInvalidEmail must not match ErrInvalidScope")
	}
}

func TestDomainErrorWrappingAndInspection(t *testing.T) {
	t.Parallel()

	t.Run("Wrap preserves sentinel, code, kind and underlying error", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("connection refused")
		wrapped := iamerrors.Wrap(cause, iamerrors.KindUnavailable, iamerrors.Code("iam.store.unavailable"), "identity store is unavailable")

		if !errors.Is(wrapped, cause) {
			t.Fatal("wrapped error must preserve underlying cause")
		}
		if !errors.Is(wrapped, iamerrors.ErrUnavailable) {
			t.Fatal("wrapped error must match base KindUnavailable sentinel")
		}
		if got := iamerrors.KindOf(wrapped); got != iamerrors.KindUnavailable {
			t.Fatalf("KindOf() = %v, want %v", got, iamerrors.KindUnavailable)
		}
		if got := iamerrors.CodeOf(wrapped); got != iamerrors.Code("iam.store.unavailable") {
			t.Fatalf("CodeOf() = %q, want %q", got, "iam.store.unavailable")
		}
		de, ok := errors.AsType[*iamerrors.DomainError](wrapped)
		if !ok {
			t.Fatal("errors.AsType[*DomainError] = false, want true")
		}
		if de.Message != "identity store is unavailable" {
			t.Fatalf("DomainError.Message = %q, want %q", de.Message, "identity store is unavailable")
		}
	})

	t.Run("NewValidation attaches field and invalid_input kind", func(t *testing.T) {
		t.Parallel()
		err := iamerrors.NewValidation("display_name", iamerrors.Code("iam.user.invalid_display_name"), "display name must not be empty")

		if !errors.Is(err, iamerrors.ErrInvalidInput) {
			t.Fatal("validation error must match ErrInvalidInput")
		}
		if got := iamerrors.KindOf(err); got != iamerrors.KindInvalidInput {
			t.Fatalf("KindOf() = %v, want %v", got, iamerrors.KindInvalidInput)
		}
		if got := iamerrors.FieldOf(err); got != "display_name" {
			t.Fatalf("FieldOf() = %q, want %q", got, "display_name")
		}
		if !strings.Contains(err.Error(), "display_name") {
			t.Fatalf("Error() = %q, expected field name included", err.Error())
		}
	})

	t.Run("WithField enriches domain sentinel without losing sentinel identity", func(t *testing.T) {
		t.Parallel()
		err := iamerrors.WithField(iamerrors.ErrInvalidEmail, "email", "email must be valid")
		outer := fmt.Errorf("create identity: %w", err)

		if !errors.Is(outer, iamerrors.ErrInvalidEmail) {
			t.Fatal("outer error must match ErrInvalidEmail")
		}
		if !errors.Is(outer, iamerrors.ErrInvalidInput) {
			t.Fatal("outer error must match ErrInvalidInput")
		}
		if got := iamerrors.FieldOf(outer); got != "email" {
			t.Fatalf("FieldOf() = %q, want %q", got, "email")
		}
		if got := iamerrors.CodeOf(outer); got != iamerrors.Code("iam.identity.invalid_email") {
			t.Fatalf("CodeOf() = %q, want %q", got, "iam.identity.invalid_email")
		}
	})

	t.Run("Unknown error helpers degrade safely", func(t *testing.T) {
		t.Parallel()
		plain := errors.New("plain error")
		if got := iamerrors.KindOf(nil); got != iamerrors.KindUnknown {
			t.Fatalf("KindOf(nil) = %v, want KindUnknown", got)
		}
		if got := iamerrors.KindOf(plain); got != iamerrors.KindUnknown {
			t.Fatalf("KindOf(plain) = %v, want KindUnknown", got)
		}
		if got := iamerrors.CodeOf(nil); got != "" {
			t.Fatalf("CodeOf(nil) = %q, want empty", got)
		}
		if got := iamerrors.CodeOf(plain); got != "" {
			t.Fatalf("CodeOf(plain) = %q, want empty", got)
		}
		if got := iamerrors.FieldOf(nil); got != "" {
			t.Fatalf("FieldOf(nil) = %q, want empty", got)
		}
		if got := iamerrors.FieldOf(plain); got != "" {
			t.Fatalf("FieldOf(plain) = %q, want empty", got)
		}
		if iamerrors.Wrap(nil, iamerrors.KindNotFound, "x", "y") != nil {
			t.Fatal("Wrap(nil) must return nil")
		}
		if iamerrors.WithField(nil, "f", "m") != nil {
			t.Fatal("WithField(nil) must return nil")
		}
		direct := iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.test.conflict"), "direct conflict")
		if !errors.Is(direct, iamerrors.ErrConflict) || direct.Error() != "direct conflict" {
			t.Fatalf("New() unexpected behavior: %v", direct)
		}
	})
}
