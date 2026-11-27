package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func TestAuthorizer_SensitiveReauthBubblesUnavailableWithoutMasking(t *testing.T) {
	t.Parallel()

	now := time.Now()
	store := newFakeStore()
	loader := NewSnapshotLoader(&fakeRefresher{}, 2*time.Second, 30*time.Second)
	mgr, err := NewAuthManager(store, loader, SessionConfig{CookieSecure: true})
	if err != nil {
		t.Fatalf("NewAuthManager: %v", err)
	}

	rawTok := "tok-sensitive-unavailable"
	tokHash := authsession.HashOpaqueToken(rawTok)
	snap := buildActiveSnapshot("user-sec-1", now, 1, 1, authz.PermissionGrant{
		Code:  "iam.users.write",
		Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
	})
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-sec-1",
			UserID:       "user-sec-1",
			TokenHashHex: tokHash,
			UserKind:     "employee",
			AuthEpoch:    1,
			IssuedAtUnix: now.Add(-15 * time.Minute).Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: "user-sec-1", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: snap,
	}

	ctx := WithRequestMetadata(context.Background(), RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})

	// Ketika layanan IAM tidak tersedia saat memeriksa sensitive reauth, Authorizer wajib mengembalikan UNAVAILABLE, bukan FORBIDDEN.
	authzUnavailable := NewAuthorizer(mgr, &fakeSensitiveChecker{err: ErrUnavailable("identity service unavailable", nil)}, 10*time.Minute)
	_, errUnavail := authzUnavailable.RequirePermission(ctx, "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true)
	if gwErr, ok := AsGatewayError(errUnavail); !ok || gwErr.Code != CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE when VerifySensitiveSession reports IAM unavailable, got %v", errUnavail)
	}

	// Error non-Unavailable / non-InternalError tetap dipetakan menjadi FORBIDDEN dengan menyimpan cause internal.
	authzForbidden := NewAuthorizer(mgr, &fakeSensitiveChecker{err: errors.New("reauth expired")}, 10*time.Minute)
	_, errForb := authzForbidden.RequirePermission(ctx, "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true)
	gwErr, ok := AsGatewayError(errForb)
	if !ok || gwErr.Code != CodeForbidden || gwErr.Err == nil {
		t.Fatalf("expected FORBIDDEN with preserved cause for expired reauth, got %v", errForb)
	}

	// Galat internal IAM tidak boleh dimask menjadi FORBIDDEN.
	authzInternal := NewAuthorizer(mgr, &fakeSensitiveChecker{err: NewGatewayError(CodeInternalError, "internal server error", errors.New("db error"))}, 10*time.Minute)
	_, errInt := authzInternal.RequirePermission(ctx, "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true)
	if gwInt, ok := AsGatewayError(errInt); !ok || gwInt.Code != CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR to bubble without FORBIDDEN masking, got %v", errInt)
	}
}
