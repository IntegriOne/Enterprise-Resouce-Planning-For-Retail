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
	for _, name := range []string{"allowed", "revoked", "redis unavailable", "stale snapshot", "wrong principal", "wrong revision", "wrong recovery", "missing caller kind", "wrong caller kind", "wrong token kind", "wrong snapshot kind", "customer kind", "missing permission", "missing mfa", "scoped grant"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			now := time.Now().Unix()
			reader := &sessionReader{state: &authsession.SessionValidationContext{
				Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
				Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", UserKind: "employee"},
				Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
				Snapshot: &authz.AccessSnapshot{
					PrincipalID: "principal", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
					SyncedAtUnix: now, FreshUntilUnix: now + 30,
					Permissions: []authz.PermissionGrant{{Code: "hrm.jobs:read", Scope: authz.ScopeGrant{Type: "global"}}},
				},
			}}
			want := hrerrors.ErrForbidden
			callerKind := "employee"
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
			case "missing caller kind":
				callerKind = ""
			case "wrong caller kind":
				callerKind = "vendor"
			case "wrong token kind":
				reader.state.Token.UserKind = "customer"
			case "wrong snapshot kind":
				reader.state.Snapshot.UserKind = "supplier"
			case "customer kind":
				callerKind = "customer"
				reader.state.Token.UserKind = callerKind
				reader.state.Snapshot.UserKind = callerKind
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
				Authenticated: true, PrincipalID: "principal", SessionID: "session", UserKind: callerKind, AuthEpoch: 1, PermRevision: 2,
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
		resource      string
		scope         authz.ScopeGrant
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
		{name: "recent staff payroll read", authAge: time.Minute, mfa: true, resource: "payroll_results", catalogAction: "read"},
		{name: "staff payroll read without MFA", authAge: time.Minute, resource: "payroll_results", catalogAction: "read", wantForbidden: true},
		{name: "own request cancel scope", resource: "service_requests", catalogAction: "cancel", authAge: time.Minute, mfa: true, scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "principal"}},
		{name: "wrong request cancel scope", resource: "service_requests", catalogAction: "cancel", authAge: time.Minute, mfa: true, scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "other"}, wantForbidden: true},
		{name: "request cancel global grant", resource: "service_requests", catalogAction: "cancel", authAge: time.Minute, mfa: true},
		{name: "own payslip scope", resource: "payslips", catalogAction: "read_self", scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "principal"}},
		{name: "wrong payslip self scope", resource: "payslips", catalogAction: "read_self", scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "other"}, wantForbidden: true},
		{name: "global permission cannot read payslip as self", resource: "payslips", catalogAction: "read_self", wantForbidden: true},
		{name: "self scope cannot read staff payroll results", authAge: time.Minute, mfa: true, resource: "payroll_results", catalogAction: "read", scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "principal"}, wantForbidden: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			resource, scope := test.resource, test.scope
			if resource == "" {
				resource = "jobs"
			}
			if scope.Type == "" {
				scope.Type = authz.ScopeGlobal
			}
			permission := "human_resource." + resource + "." + test.catalogAction
			now := time.Now()
			issuedAt := now.Add(-test.authAge).Unix()
			if test.missingTime {
				issuedAt = 0
			}
			reader := &sessionReader{state: &authsession.SessionValidationContext{
				Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
				Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", UserKind: "employee", MFAAuthenticated: test.mfa, IssuedAtUnix: issuedAt},
				Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
				Snapshot: &authz.AccessSnapshot{
					PrincipalID: "principal", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
					SyncedAtUnix: now.Unix(), FreshUntilUnix: now.Unix() + 30,
					Permissions: []authz.PermissionGrant{{Code: permission, IsSensitive: test.grantFlag, Scope: scope}},
				},
			}}
			policy, err := security.NewPolicy(reader, map[string]string{
				"hr_" + resource + "." + test.catalogAction: permission,
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
			ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
				Authenticated: true, PrincipalID: "principal", SessionID: "session", UserKind: "employee", AuthEpoch: 1, PermRevision: 2,
			})
			_, err = policy.Authorize(ctx, operation.Access{Resource: "hr_" + resource, Action: test.catalogAction})
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
		Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", UserKind: "employee"},
		Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "principal", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
			SyncedAtUnix: now, FreshUntilUnix: now + 30,
			Permissions: []authz.PermissionGrant{
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalB}},
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: " " + legalC}},
				{Code: "human_resource.workforce_plans.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
				{Code: "human_resource.workforce_plan_lines.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
				{Code: "human_resource.workforce_plans.submit", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}},
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
		Authenticated: true, PrincipalID: "principal", SessionID: "session", UserKind: "employee", AuthEpoch: 1, PermRevision: 2,
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
	planScope, err := policy.ReadScope(ctx, operation.Access{Resource: "hr_workforce_plans", Action: "list"})
	if err != nil || planScope.All || len(planScope.Filters) != 1 ||
		planScope.Filters[0].Column != "legal_entity_reference_id" ||
		len(planScope.Filters[0].Values) != 1 || planScope.Filters[0].Values[0] != legalA {
		t.Fatalf("workforce plan scope = %+v, error = %v", planScope, err)
	}
	if _, err := policy.Authorize(ctx, operation.Access{
		Resource: "hr_workforce_plans", Action: "read", Attributes: map[string]string{"legal_entity_reference_id": legalA},
	}); err != nil {
		t.Fatalf("authorize workforce plan in scope: %v", err)
	}
	for _, access := range []operation.Access{
		{Resource: "hr_workforce_plans", Action: "read", Attributes: map[string]string{"legal_entity_reference_id": legalB}},
		{Resource: "hr_workforce_plans", Action: "submit", Attributes: map[string]string{"legal_entity_reference_id": legalA}},
		{Resource: "hr_workforce_plan_lines", Action: "read", Attributes: map[string]string{"workforce_plan_id": legalA}},
	} {
		if _, err := policy.Authorize(ctx, access); !errors.Is(err, hrerrors.ErrForbidden) {
			t.Fatalf("unsupported workforce scope %+v: %v", access, err)
		}
	}
	lineScope, err := policy.ReadScope(ctx, operation.Access{Resource: "hr_workforce_plan_lines", Action: "list"})
	if err != nil || lineScope.All || len(lineScope.Filters) != 1 ||
		lineScope.Filters[0].Column != "plan_legal_entity_reference_id" || len(lineScope.Filters[0].Values) != 1 ||
		lineScope.Filters[0].Values[0] != legalA {
		t.Fatalf("workforce line scope = %+v, error = %v", lineScope, err)
	}
	reader.state.Snapshot.Permissions[0].IsSensitive = true
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalA}}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("sensitive grant without MFA must deny legal entity A: %v", err)
	}
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalB}}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("sensitive permission code must also deny legal entity B without MFA: %v", err)
	}
	if _, err := policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("sensitive permission code must deny list without MFA: %v", err)
	}
	reader.state.Snapshot.Permissions = append(reader.state.Snapshot.Permissions,
		authz.PermissionGrant{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: legalA}})
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalA}}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("duplicate non-sensitive grant must not bypass MFA: %v", err)
	}
	reader.state.Token.MFAAuthenticated = true
	reader.state.Token.IssuedAtUnix = now
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read", Attributes: map[string]string{"primary_legal_entity_reference_id": legalA}}); err != nil {
		t.Fatalf("valid MFA must allow legal entity A: %v", err)
	}
	scope, err = policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"})
	if err != nil || len(scope.Filters) != 1 || len(scope.Filters[0].Values) != 2 {
		t.Fatalf("valid MFA must include both legal entities: %+v, error = %v", scope, err)
	}
	reader.state.Snapshot.Status = "suspended"
	if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: "read"}); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("suspended scoped worker read = %v, want unavailable", err)
	}
	reader.state.Snapshot.Status = "active"
	reader.state.Snapshot.Permissions = append(reader.state.Snapshot.Permissions,
		authz.PermissionGrant{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "global"}})
	reader.state.Token.MFAAuthenticated = false
	if _, err := policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("global non-sensitive duplicate must not bypass code-level MFA: %v", err)
	}
	reader.state.Token.MFAAuthenticated = true
	scope, err = policy.ReadScope(ctx, operation.Access{Resource: "hr_workers", Action: "list"})
	if err != nil || !scope.All || len(scope.Filters) != 0 {
		t.Fatalf("global worker scope = %+v, error = %v", scope, err)
	}
}

func TestRequisitionApprovalRequiresMatchingUnitGrant(t *testing.T) {
	testSensitiveScopedAction(t, "hr_requisitions", "approve", "human_resource.requisitions.approve",
		"organization_unit", "organization_unit_reference_id")
}

func TestPayslipReissueRequiresMatchingLegalEntityGrant(t *testing.T) {
	testSensitiveScopedAction(t, "hr_payslips", "reissue", "human_resource.payslips.reissue",
		"legal_entity", "employing_legal_entity_reference_id")
}

func testSensitiveScopedAction(t *testing.T, resource, action, permission, scope, column string) {
	t.Helper()
	t.Parallel()
	const unit = "00000000-0000-4000-8000-000000000001"
	const otherUnit = "00000000-0000-4000-8000-000000000002"
	global := authz.PermissionGrant{Code: permission, Scope: authz.ScopeGrant{Type: "global"}}
	scoped := authz.PermissionGrant{Code: permission, Scope: authz.ScopeGrant{Type: scope, ID: unit}}
	for _, tc := range []struct {
		name       string
		grants     []authz.PermissionGrant
		attributes map[string]string
		mfa        bool
		want       error
	}{
		{name: "precheck with a scoped grant", grants: []authz.PermissionGrant{scoped}, mfa: true},
		{name: "global alone", grants: []authz.PermissionGrant{global}, mfa: true, want: hrerrors.ErrForbidden},
		{name: "match despite global", grants: []authz.PermissionGrant{global, scoped}, attributes: map[string]string{column: unit}, mfa: true},
		{name: "other scope", grants: []authz.PermissionGrant{global, scoped}, attributes: map[string]string{column: otherUnit}, mfa: true, want: hrerrors.ErrForbidden},
		{name: "record missing scope", grants: []authz.PermissionGrant{scoped}, attributes: map[string]string{"id": otherUnit}, mfa: true, want: hrerrors.ErrForbidden},
		{name: "no mfa", grants: []authz.PermissionGrant{scoped}, attributes: map[string]string{column: unit}, want: hrerrors.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			now := time.Now().Unix()
			reader := &sessionReader{state: &authsession.SessionValidationContext{
				Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
				Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", UserKind: "employee", MFAAuthenticated: tc.mfa, IssuedAtUnix: now},
				Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
				Snapshot: &authz.AccessSnapshot{
					PrincipalID: "principal", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
					SyncedAtUnix: now, FreshUntilUnix: now + 30, Permissions: tc.grants,
				},
			}}
			policy, err := security.NewPolicy(reader, map[string]string{resource + "." + action: permission})
			if err != nil {
				t.Fatal(err)
			}
			ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
			ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
				Authenticated: true, PrincipalID: "principal", SessionID: "session", UserKind: "employee", AuthEpoch: 1, PermRevision: 2,
			})
			_, err = policy.Authorize(ctx, operation.Access{Resource: resource, Action: action, Attributes: tc.attributes})
			if !errors.Is(err, tc.want) {
				t.Fatalf("authorize: got %v, want %v", err, tc.want)
			}
		})
	}
}
