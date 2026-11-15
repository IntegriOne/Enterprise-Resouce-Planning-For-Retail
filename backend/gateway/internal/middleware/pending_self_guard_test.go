package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func TestAuthorizer_RequireSelfPermissionRejectsPendingEvenWithAssignedSelfGrants(t *testing.T) {
	t.Parallel()

	const userID = "11111111-1111-4111-8111-111111111111"
	now := time.Now()
	selfGrants := []authz.PermissionGrant{
		{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}},
		{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}},
		{Code: authz.PermUsersDeleteSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}},
		{Code: authz.PermSessionsReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}},
		{Code: authz.PermSessionsDeleteSelf, IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}},
	}

	buildSnap := func(status string) *authz.AccessSnapshot {
		return &authz.AccessSnapshot{
			PrincipalID:    userID,
			UserKind:       "customer",
			Status:         status,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			Roles:          []authz.RoleGrant{{Code: "customer"}},
			Permissions:    selfGrants,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		}
	}

	for _, status := range []string{"pending", "suspended", "disabled"} {
		store := newFakeStore()
		ref := &fakeRefresher{snap: buildSnap(status)}
		loader := NewSnapshotLoader(ref, 2*time.Second, 30*time.Second)
		mgr, err := NewAuthManager(store, loader, SessionConfig{CookieSecure: true})
		if err != nil {
			t.Fatalf("NewAuthManager() error = %v", err)
		}
		authzSvc := NewAuthorizer(mgr, &fakeSensitiveChecker{}, 10*time.Minute)

		rawTok := "tok-" + status
		tokHash := authsession.HashOpaqueToken(rawTok)
		store.sessions[tokHash] = &authsession.SessionValidationContext{
			Token: &authsession.SessionTokenRecord{
				SessionID:    "sess-" + status,
				UserID:       userID,
				TokenHashHex: tokHash,
				AuthEpoch:    1,
				IssuedAtUnix: now.Unix(),
			},
			Security: &authsession.UserSecurityRecord{
				UserID:        userID,
				Status:        status,
				AuthEpoch:     1,
				PermRevision:  1,
				RecoveryEpoch: 1,
			},
			Snapshot: buildSnap(status),
		}

		ctx := WithRequestMetadata(context.Background(), RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})
		for _, tc := range []struct {
			perm      string
			sensitive bool
		}{
			{perm: authz.PermUsersReadSelf, sensitive: false},
			{perm: authz.PermUsersUpdateSelf, sensitive: false},
			{perm: authz.PermUsersDeleteSelf, sensitive: false},
			{perm: authz.PermSessionsReadSelf, sensitive: false},
			{perm: authz.PermSessionsDeleteSelf, sensitive: true},
		} {
			_, err := authzSvc.RequireSelfPermission(ctx, tc.perm, tc.sensitive)
			if err == nil {
				t.Fatalf("status=%q perm=%s succeeded, want FORBIDDEN", status, tc.perm)
			}
			gwErr, ok := AsGatewayError(err)
			if !ok || gwErr.Code != CodeForbidden {
				t.Fatalf("status=%q perm=%s error = %v, want CodeForbidden", status, tc.perm, err)
			}
		}
	}

	// Saat akun berubah menjadi ACTIVE, seluruh operasi *self langsung diizinkan.
	store := newFakeStore()
	loader := NewSnapshotLoader(&fakeRefresher{snap: buildSnap(authz.StatusActive)}, 2*time.Second, 30*time.Second)
	mgr, _ := NewAuthManager(store, loader, SessionConfig{CookieSecure: true})
	authzSvc := NewAuthorizer(mgr, &fakeSensitiveChecker{}, 10*time.Minute)

	rawTok := "tok-active"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-active",
			UserID:       userID,
			TokenHashHex: tokHash,
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{
			UserID:        userID,
			Status:        authz.StatusActive,
			AuthEpoch:     1,
			PermRevision:  1,
			RecoveryEpoch: 1,
		},
		Snapshot: buildSnap(authz.StatusActive),
	}
	ctx := WithRequestMetadata(context.Background(), RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})

	if _, err := authzSvc.RequireSelfPermission(ctx, authz.PermUsersReadSelf, false); err != nil {
		t.Fatalf("active RequireSelfPermission(readself) error = %v", err)
	}
	if _, err := authzSvc.RequireSelfSensitive(ctx, authz.PermSessionsDeleteSelf); err != nil {
		t.Fatalf("active RequireSelfSensitive(deleteself) error = %v", err)
	}
}
