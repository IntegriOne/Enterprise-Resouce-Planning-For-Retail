package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func TestAuthorizer_RequireSelfPermission(t *testing.T) {
	t.Parallel()

	const (
		callerID    = "11111111-1111-4111-8111-111111111111"
		otherUserID = "22222222-2222-4222-8222-222222222222"
	)

	makeCtx := func(
		t *testing.T,
		authenticated bool,
		tokenUserID string,
		issuedAt time.Time,
		mfaEnabled bool,
		mfaAuthenticated bool,
		perms []authz.PermissionGrant,
		checkerErr error,
	) (*Authorizer, context.Context, *fakeSensitiveChecker) {
		t.Helper()
		now := time.Now()
		store := newFakeStore()
		loader := NewSnapshotLoader(&fakeRefresher{}, 2*time.Second, 30*time.Second)
		mgr, err := NewAuthManager(store, loader, SessionConfig{CookieSecure: true})
		if err != nil {
			t.Fatalf("NewAuthManager failed: %v", err)
		}

		var checker *fakeSensitiveChecker
		if checkerErr != nil {
			checker = &fakeSensitiveChecker{err: checkerErr}
		} else {
			checker = &fakeSensitiveChecker{}
		}
		authzSvc := NewAuthorizer(mgr, checker, 10*time.Minute)
		if !authenticated {
			return authzSvc, context.Background(), checker
		}

		rawTok := "self-authz-" + t.Name()
		tokHash := authsession.HashOpaqueToken(rawTok)
		snapshotUserID := tokenUserID
		if snapshotUserID == "" {
			snapshotUserID = callerID
		}
		store.sessions[tokHash] = &authsession.SessionValidationContext{
			Token: &authsession.SessionTokenRecord{
				SessionID:        "sess-self-authz",
				UserID:           tokenUserID,
				TokenHashHex:     tokHash,
				UserKind:         "customer",
				AuthEpoch:        1,
				IssuedAtUnix:     issuedAt.Unix(),
				MFAAuthenticated: mfaAuthenticated,
			},
			Security: &authsession.UserSecurityRecord{
				UserID:        snapshotUserID,
				Status:        authz.StatusActive,
				AuthEpoch:     1,
				PermRevision:  1,
				RecoveryEpoch: 1,
			},
			Snapshot: &authz.AccessSnapshot{
				PrincipalID:    snapshotUserID,
				UserKind:       "customer",
				Status:         authz.StatusActive,
				MFAEnabled:     mfaEnabled,
				AuthEpoch:      1,
				PermRevision:   1,
				RecoveryEpoch:  1,
				SyncedAtUnix:   now.Unix(),
				FreshUntilUnix: now.Add(30 * time.Second).Unix(),
				Permissions:    perms,
			},
		}

		ctx := WithRequestMetadata(context.Background(), RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})
		return authzSvc, ctx, checker
	}

	now := time.Now()
	cases := []struct {
		name              string
		authenticated     bool
		tokenUserID       string
		issuedAt          time.Time
		mfaEnabled        bool
		mfaAuthenticated  bool
		perms             []authz.PermissionGrant
		permCode          string
		requireSensitive  bool
		checkerErr        error
		wantErrCode       ErrorCode
		wantCheckerCalled bool
	}{
		{
			name:             "rejects_empty_perm_code",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}}},
			permCode:         "   ",
			requireSensitive: false,
			wantErrCode:      CodeForbidden,
		},
		{
			name:             "rejects_unauthenticated_caller",
			authenticated:    false,
			permCode:         authz.PermUsersReadSelf,
			requireSensitive: false,
			wantErrCode:      CodeUnauthenticated,
		},
		{
			name:             "allows_matching_scope_self",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: callerID}}},
			permCode:         authz.PermUsersReadSelf,
			requireSensitive: false,
		},
		{
			name:             "allows_scope_global_self_permission",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}}},
			permCode:         authz.PermUsersReadSelf,
			requireSensitive: false,
		},
		{
			name:             "rejects_foreign_scope_self_idor",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: otherUserID}}},
			permCode:         authz.PermUsersReadSelf,
			requireSensitive: false,
			wantErrCode:      CodeForbidden,
		},
		{
			name:             "rejects_admin_perm_without_self_perm",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersRead, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}}},
			permCode:         authz.PermUsersReadSelf,
			requireSensitive: false,
			wantErrCode:      CodeForbidden,
		},
		{
			name:             "allows_non_sensitive_self_update_on_older_session",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now.Add(-20 * time.Minute),
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: callerID}}},
			permCode:         authz.PermUsersUpdateSelf,
			requireSensitive: false,
			checkerErr:       errors.New("reauth expired"),
		},
		{
			name:              "rejects_stale_session_on_sensitive_self_action",
			authenticated:     true,
			tokenUserID:       callerID,
			issuedAt:          now.Add(-20 * time.Minute),
			perms:             []authz.PermissionGrant{{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: callerID}}},
			permCode:          authz.PermUsersUpdateSelf,
			requireSensitive:  true,
			checkerErr:        errors.New("reauth expired"),
			wantErrCode:       CodeForbidden,
			wantCheckerCalled: true,
		},
		{
			name:             "rejects_unverified_mfa_on_sensitive_self_action",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			mfaEnabled:       true,
			mfaAuthenticated: false,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: callerID}}},
			permCode:         authz.PermUsersUpdateSelf,
			requireSensitive: true,
			checkerErr:       errors.New("mfa required"),
			wantErrCode:      CodeForbidden,
		},
		{
			name:             "allows_fresh_authenticated_sensitive_self_action",
			authenticated:    true,
			tokenUserID:      callerID,
			issuedAt:         now,
			mfaEnabled:       true,
			mfaAuthenticated: true,
			perms:            []authz.PermissionGrant{{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: callerID}}},
			permCode:         authz.PermUsersUpdateSelf,
			requireSensitive: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			authzSvc, ctx, checker := makeCtx(t, tc.authenticated, tc.tokenUserID, tc.issuedAt, tc.mfaEnabled, tc.mfaAuthenticated, tc.perms, tc.checkerErr)

			sess, err := authzSvc.RequireSelfPermission(ctx, tc.permCode, tc.requireSensitive)
			if tc.wantErrCode == "" {
				if err != nil || sess == nil {
					t.Fatalf("RequireSelfPermission(%q, %v) unexpected error: %v", tc.permCode, tc.requireSensitive, err)
				}
				if tc.requireSensitive {
					if _, sensErr := authzSvc.RequireSelfSensitive(ctx, tc.permCode); sensErr != nil {
						t.Fatalf("RequireSelfSensitive(%q) unexpected error: %v", tc.permCode, sensErr)
					}
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error code %s, got nil", tc.wantErrCode)
			}
			gwErr, ok := AsGatewayError(err)
			if !ok || gwErr.Code != tc.wantErrCode {
				t.Fatalf("expected error code %s, got %v", tc.wantErrCode, err)
			}
			if tc.wantCheckerCalled && checker.calls.Load() == 0 {
				t.Fatalf("expected sensitiveChecker to be consulted for %s", tc.name)
			}
		})
	}
}
