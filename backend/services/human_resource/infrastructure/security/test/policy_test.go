package security_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/metadata"
)

type sessionReader struct {
	state *authsession.SessionValidationContext
	err   error
}

func (r *sessionReader) ReadSessionContext(context.Context, string, time.Time) (*authsession.SessionValidationContext, error) {
	return r.state, r.err
}

func TestPolicyRejectsUnverifiedState(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"allowed", "revoked", "redis unavailable", "stale snapshot", "wrong principal", "wrong revision", "wrong recovery", "missing permission", "missing mfa", "scoped grant"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			now := time.Now().Unix()
			reader := &sessionReader{state: &authsession.SessionValidationContext{
				Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
				Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session"},
				Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
				Snapshot: &authz.AccessSnapshot{
					PrincipalID: "principal", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
					SyncedAtUnix: now, FreshUntilUnix: now + 30,
					Permissions: []authz.PermissionGrant{{Code: "hrm.jobs:read", Scope: authz.ScopeGrant{Type: "global"}}},
				},
			}}
			want := hrerrors.ErrForbidden
			switch name {
			case "allowed":
				want = nil
			case "revoked":
				reader.err = authsession.ErrSessionRevoked
				want = hrerrors.ErrUnauthenticated
			case "redis unavailable":
				reader.err = authsession.ErrRedisUnavailable
				want = hrerrors.ErrUnavailable
			case "stale snapshot":
				reader.state.Snapshot.FreshUntilUnix = now - 1
				want = hrerrors.ErrUnavailable
			case "wrong principal":
				reader.state.Token.UserID = "other"
			case "wrong revision":
				reader.state.Security.PermRevision++
			case "wrong recovery":
				reader.state.Gate.RecoveryEpoch++
			case "missing permission":
				reader.state.Snapshot.Permissions = nil
			case "missing mfa":
				reader.state.Snapshot.Permissions[0].IsSensitive = true
			case "scoped grant":
				reader.state.Snapshot.Permissions[0].Scope = authz.ScopeGrant{Type: "legal_entity", ID: "restricted"}
			}
			policy, err := security.NewPolicy(reader, map[string]string{"hr_jobs.read": "hrm.jobs:read"})
			if err != nil {
				t.Fatal(err)
			}
			ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
			ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
				Authenticated: true, PrincipalID: "principal", SessionID: "session", AuthEpoch: 1, PermRevision: 2,
			})
			_, err = policy.Authorize(ctx, operation.Access{Resource: "hr_jobs", Action: "read"})
			if !errors.Is(err, want) {
				t.Fatalf("authorize: got %v, want %v", err, want)
			}
		})
	}
}

func TestPolicyRequiresRecentAuthenticationForSensitiveOperations(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name          string
		authAge       time.Duration
		missingTime   bool
		mfa           bool
		catalogAction string
		grantFlag     bool
		wantForbidden bool
	}{
		{name: "recent catalog-sensitive action", authAge: 5 * time.Minute, mfa: true, catalogAction: "approve"},
		{name: "expired catalog-sensitive action", authAge: 11 * time.Minute, mfa: true, catalogAction: "approve", wantForbidden: true},
		{name: "future authentication time", authAge: -time.Minute, mfa: true, catalogAction: "approve", wantForbidden: true},
		{name: "missing authentication time", missingTime: true, mfa: true, catalogAction: "approve", wantForbidden: true},
		{name: "missing mfa", authAge: time.Minute, catalogAction: "approve", wantForbidden: true},
		{name: "recent sensitive grant", authAge: 5 * time.Minute, mfa: true, catalogAction: "read", grantFlag: true},
		{name: "expired sensitive grant", authAge: 11 * time.Minute, mfa: true, catalogAction: "read", grantFlag: true, wantForbidden: true},
		{name: "ordinary read keeps working", authAge: 11 * time.Minute, catalogAction: "read"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			now := time.Now()
			issuedAt := now.Add(-test.authAge).Unix()
			if test.missingTime {
				issuedAt = 0
			}
			reader := &sessionReader{state: &authsession.SessionValidationContext{
				Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
				Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", MFAAuthenticated: test.mfa, IssuedAtUnix: issuedAt},
				Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
				Snapshot: &authz.AccessSnapshot{
					PrincipalID: "principal", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
					SyncedAtUnix: now.Unix(), FreshUntilUnix: now.Unix() + 30,
					Permissions: []authz.PermissionGrant{{Code: "human_resource.jobs." + test.catalogAction, IsSensitive: test.grantFlag, Scope: authz.ScopeGrant{Type: "global"}}},
				},
			}}
			policy, err := security.NewPolicy(reader, map[string]string{
				"hr_jobs." + test.catalogAction: "human_resource.jobs." + test.catalogAction,
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
			ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
				Authenticated: true, PrincipalID: "principal", SessionID: "session", AuthEpoch: 1, PermRevision: 2,
			})
			_, err = policy.Authorize(ctx, operation.Access{Resource: "hr_jobs", Action: test.catalogAction})
			if test.wantForbidden && !errors.Is(err, hrerrors.ErrForbidden) {
				t.Fatalf("authorize error = %v, want forbidden", err)
			}
			if !test.wantForbidden && err != nil {
				t.Fatalf("authorize error = %v, want allowed", err)
			}
		})
	}
}

func TestPolicyWorkerLegalEntityScope(t *testing.T) {
	t.Parallel()
	const legalA = "11111111-1111-4111-8111-111111111111"
	const legalB = "22222222-2222-4222-8222-222222222222"
	const legalC = "33333333-3333-4333-8333-333333333333"
	now := time.Now().Unix()
	reader := &sessionReader{state: &authsession.SessionValidationContext{
		Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
		Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session"},
		Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "principal", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
			SyncedAtUnix: now, FreshUntilUnix: now + 30,
			Permissions: []authz.PermissionGrant{
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalB}},
				{Code: "human_resource.jobs.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
			},
		},
	}}
	policy, err := security.NewPolicy(reader, security.PermissionCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
	ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
		Authenticated: true, PrincipalID: "principal", SessionID: "session", AuthEpoch: 1, PermRevision: 2,
	})
	for _, id := range []string{legalA, legalB} {
		if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": id}}); err != nil {
			t.Fatalf("worker in legal entity %s denied: %v", id, err)
		}
	}
	for _, access := range []operation.Access{
		{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalC}},
		{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": "invalid"}},
		{Resource: "hr_workers", Action: "create", Attributes: map[string]string{"primary_legal_entity_reference_id": legalA}},
		{Resource: "hr_jobs", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalA}},
	} {
		if _, err := policy.Authorize(ctx, access); !errors.Is(err, hrerrors.ErrForbidden) {
			t.Fatalf("unsupported or outside scope %+v: %v", access, err)
		}
	}
	scope, err := policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"})
	if err != nil || scope.All || len(scope.Filters) != 1 || scope.Filters[0].Column != "primary_legal_entity_reference_id" || len(scope.Filters[0].Values) != 2 {
		t.Fatalf("worker scope = %+v, error = %v", scope, err)
	}
	reader.state.Snapshot.Status = "suspended"
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read"}); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("suspended scoped worker read = %v, want unavailable", err)
	}
	reader.state.Snapshot.Status = "active"
	reader.state.Snapshot.Permissions = append(reader.state.Snapshot.Permissions,
		authz.PermissionGrant{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "global"}})
	scope, err = policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"})
	if err != nil || !scope.All || len(scope.Filters) != 0 {
		t.Fatalf("global worker scope = %+v, error = %v", scope, err)
	}
}
