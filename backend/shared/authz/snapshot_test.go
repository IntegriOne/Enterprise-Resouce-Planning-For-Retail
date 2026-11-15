package authz

import (
	"testing"
	"time"
)

func TestAccessSnapshotFreshnessAndPermissions(t *testing.T) {
	t.Parallel()

	now := time.Now()
	snap := &AccessSnapshot{
		PrincipalID:         "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
		UserKind:            "internal",
		Status:              StatusActive,
		AuthEpoch:           2,
		PermRevision:        5,
		RecoveryEpoch:       1,
		CanAccessIdentity:   true,
		DefaultRedirectPath: "/identity",
		Roles: []RoleGrant{
			{ID: "r1", Code: "iam_operator", Name: "IAM Operator", Priority: 20, DefaultRedirectPath: "/identity"},
		},
		Permissions: []PermissionGrant{
			{Code: "iam.users.read", Scope: ScopeGrant{Type: ScopeGlobal}},
			{Code: "iam.users.update", IsSensitive: true, Scope: ScopeGrant{Type: "branch", ID: "b1"}},
		},
		SyncedAtUnix:   now.Unix(),
		FreshUntilUnix: now.Add(DefaultSnapshotMaxAge).Unix(),
	}

	if !snap.IsFresh(now) {
		t.Fatal("expected snapshot to be fresh at issue time")
	}
	if snap.IsFresh(now.Add(31 * time.Second)) {
		t.Fatal("expected snapshot to be stale after 30s")
	}
	if !snap.CanAccessWorkspaceIdentity() {
		t.Fatal("expected identity workspace access to be true")
	}
	if !snap.HasRole("IAM_OPERATOR") {
		t.Fatal("expected case-insensitive role match")
	}
	if _, ok := snap.HasPermission("iam.users.read"); !ok {
		t.Fatal("expected global permission match")
	}
	if _, ok := snap.HasPermission("iam.users.update"); ok {
		t.Fatal("did not expect branch-scoped permission to satisfy global requirement")
	}
	if perm, ok := snap.HasScopedPermission("iam.users.update", ScopeGrant{Type: "branch", ID: "b1"}); !ok || !perm.IsSensitive {
		t.Fatalf("expected scoped sensitive permission match, got (%+v, %v)", perm, ok)
	}
	if _, ok := snap.HasScopedPermission("iam.users.update", ScopeGrant{Type: "branch", ID: "b2"}); ok {
		t.Fatal("did not expect mismatched branch ID to match")
	}

	raw, err := MarshalSnapshot(snap)
	if err != nil {
		t.Fatalf("MarshalSnapshot() error = %v", err)
	}
	decoded, err := UnmarshalSnapshot(raw)
	if err != nil {
		t.Fatalf("UnmarshalSnapshot() error = %v", err)
	}
	if decoded.PrincipalID != snap.PrincipalID || decoded.PermRevision != snap.PermRevision {
		t.Fatalf("roundtrip mismatch: got %+v, want %+v", decoded, snap)
	}
}
