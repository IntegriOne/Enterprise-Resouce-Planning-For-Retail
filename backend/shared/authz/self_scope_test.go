package authz_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/shared/authz"
)

func TestIsSelfServicePermissionCodeAndScopeMatching(t *testing.T) {
	t.Parallel()

	selfCodes := []string{
		authz.PermUsersReadSelf,
		authz.PermUsersUpdateSelf,
		authz.PermUsersDeleteSelf,
		authz.PermSessionsReadSelf,
		authz.PermSessionsDeleteSelf,
		"human_resource.payslips.read_self",
		" IAM.Users.ReadSelf ",
	}
	for _, code := range selfCodes {
		if !authz.IsSelfServicePermissionCode(code) {
			t.Fatalf("expected IsSelfServicePermissionCode(%q) = true", code)
		}
	}

	nonSelfCodes := []string{
		"",
		authz.PermUsersRead,
		authz.PermUsersWrite,
		authz.PermSessionsRead,
		authz.PermSessionsRevoke,
		authz.PermDashboardView,
	}
	for _, code := range nonSelfCodes {
		if authz.IsSelfServicePermissionCode(code) {
			t.Fatalf("expected IsSelfServicePermissionCode(%q) = false", code)
		}
	}

	selfA := authz.ScopeGrant{Type: authz.ScopeSelf, ID: "11111111-1111-4111-8111-111111111111"}
	selfAUpper := authz.ScopeGrant{Type: " SELF ", ID: "11111111-1111-4111-8111-111111111111"}
	selfB := authz.ScopeGrant{Type: authz.ScopeSelf, ID: "22222222-2222-4222-8222-222222222222"}
	global := authz.ScopeGrant{Type: authz.ScopeGlobal}

	if !selfA.Matches(selfAUpper) {
		t.Fatal("self scope must match case-insensitively on same user ID")
	}
	if selfA.Matches(selfB) {
		t.Fatal("self scope for user A must not match user B")
	}
	if selfA.Matches(global) {
		t.Fatal("self scope must not satisfy global scope")
	}
	if !global.Matches(selfA) {
		t.Fatal("global grant must still satisfy scoped check unless blocked by self-service rule")
	}
}

func TestPayslipPermissionRequiresOwnSelfScope(t *testing.T) {
	t.Parallel()

	const permission = authz.PermPayslipsReadSelf
	owner := "11111111-1111-4111-8111-111111111111"
	other := "22222222-2222-4222-8222-222222222222"
	snapshot := &authz.AccessSnapshot{
		PrincipalID: owner,
		Status:      authz.StatusActive,
		Permissions: []authz.PermissionGrant{
			{Code: permission, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		},
	}
	if _, ok := snapshot.HasPermission(permission); ok {
		t.Fatal("payslip permission must not authorize global access")
	}
	if _, ok := snapshot.HasScopedPermission(permission, authz.ScopeGrant{Type: authz.ScopeSelf, ID: other}); ok {
		t.Fatal("payslip permission must not authorize access to another principal")
	}
	ownScope := authz.ScopeGrant{Type: authz.ScopeSelf, ID: owner}
	if _, ok := snapshot.HasScopedPermission(permission, ownScope); ok {
		t.Fatal("global grant must not authorize the owner's payslip")
	}
	snapshot.Permissions = append(snapshot.Permissions, authz.PermissionGrant{Code: permission, Scope: ownScope})
	if grant, ok := snapshot.HasScopedPermission(permission, ownScope); !ok || grant.Scope != ownScope {
		t.Fatal("payslip permission must authorize the owner")
	}
	snapshot.PrincipalID = ""
	if _, ok := snapshot.HasScopedPermission(permission, ownScope); ok {
		t.Fatal("payslip permission must require a snapshot principal")
	}
}

func TestAccessSnapshotHasScopedPermissionSelfGuard(t *testing.T) {
	t.Parallel()

	userA := "11111111-1111-4111-8111-111111111111"
	userB := "22222222-2222-4222-8222-222222222222"

	// Walau ada grant global (misal superadmin) atau grant self milik userA, evaluasi *self wajib fail-closed terhadap scope global maupun ID orang lain.
	snap := &authz.AccessSnapshot{
		PrincipalID: userA,
		Status:      authz.StatusActive,
		Permissions: []authz.PermissionGrant{
			{
				Code:        authz.PermUsersReadSelf,
				IsSensitive: false,
				Scope:       authz.ScopeGrant{Type: authz.ScopeSelf, ID: userA},
			},
			{
				Code:        authz.PermSessionsDeleteSelf,
				IsSensitive: true,
				Scope:       authz.ScopeGrant{Type: authz.ScopeGlobal},
			},
		},
	}

	if _, ok := snap.HasPermission(authz.PermUsersReadSelf); ok {
		t.Fatal("*self permission must never satisfy HasPermission(Global)")
	}
	if _, ok := snap.HasScopedPermission(authz.PermSessionsDeleteSelf, authz.ScopeGrant{Type: authz.ScopeGlobal}); ok {
		t.Fatal("*self permission must fail even with Global grant when evaluated against ScopeGlobal")
	}
	if _, ok := snap.HasScopedPermission(authz.PermUsersReadSelf, authz.ScopeGrant{Type: authz.ScopeSelf, ID: userB}); ok {
		t.Fatal("user A must not satisfy *self permission evaluated for user B")
	}
	if _, ok := snap.HasScopedPermission(authz.PermSessionsDeleteSelf, authz.ScopeGrant{Type: authz.ScopeSelf, ID: userB}); ok {
		t.Fatal("global grant on *self must still fail when required self ID belongs to user B")
	}

	if _, ok := snap.HasScopedPermission(authz.PermUsersReadSelf, authz.ScopeGrant{Type: authz.ScopeSelf, ID: userA}); !ok {
		t.Fatal("matching own user ID with self grant must succeed")
	}
	if grant, ok := snap.HasScopedPermission(authz.PermSessionsDeleteSelf, authz.ScopeGrant{Type: authz.ScopeSelf, ID: userA}); !ok || !grant.IsSensitive {
		t.Fatalf("superadmin global grant evaluated on own user ID must succeed and preserve IsSensitive=true, got ok=%v grant=%+v", ok, grant)
	}
}
