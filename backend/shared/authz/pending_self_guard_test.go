package authz

import (
	"testing"
	"time"
)

func TestAccessSnapshot_RejectsNonActiveSelfPermissions(t *testing.T) {
	t.Parallel()

	now := time.Now()
	const userID = "11111111-1111-4111-8111-111111111111"
	const otherUserID = "22222222-2222-4222-8222-222222222222"

	selfScope := ScopeGrant{Type: ScopeSelf, ID: userID}
	otherScope := ScopeGrant{Type: ScopeSelf, ID: otherUserID}
	selfPerms := []string{
		PermUsersReadSelf,
		PermUsersUpdateSelf,
		PermUsersDeleteSelf,
		PermSessionsReadSelf,
		PermSessionsDeleteSelf,
	}

	grants := make([]PermissionGrant, 0, len(selfPerms))
	for _, perm := range selfPerms {
		grants = append(grants, PermissionGrant{
			Code:        perm,
			IsSensitive: perm == PermSessionsDeleteSelf,
			Scope:       selfScope,
		})
	}

	// Meski role dasar & izin *self sudah terpasang, akun non-active wajib ditolak.
	for _, status := range []string{"pending", "suspended", "disabled", ""} {
		snap := &AccessSnapshot{
			PrincipalID:       userID,
			UserKind:          "customer",
			Status:            status,
			AuthEpoch:         1,
			PermRevision:      1,
			RecoveryEpoch:     1,
			Roles:             []RoleGrant{{Code: "customer"}},
			Permissions:       grants,
			CanAccessIdentity: true,
			SyncedAtUnix:      now.Unix(),
			FreshUntilUnix:    now.Add(30 * time.Second).Unix(),
		}

		if snap.IsFresh(now) {
			t.Fatalf("status=%q IsFresh() = true, want false", status)
		}
		if snap.HasRole("customer") {
			t.Fatalf("status=%q HasRole(customer) = true, want false", status)
		}
		if snap.CanAccessWorkspaceIdentity() {
			t.Fatalf("status=%q CanAccessWorkspaceIdentity() = true, want false", status)
		}
		for _, perm := range selfPerms {
			if _, ok := snap.HasScopedPermission(perm, selfScope); ok {
				t.Fatalf("status=%q HasScopedPermission(%s, self) = true, want false", status, perm)
			}
		}
	}

	// Setelah status menjadi ACTIVE, seluruh izin *self milik sendiri langsung aktif dan lintas-user tetap ditolak.
	activeSnap := &AccessSnapshot{
		PrincipalID:    userID,
		UserKind:       "customer",
		Status:         StatusActive,
		AuthEpoch:      1,
		PermRevision:   1,
		RecoveryEpoch:  1,
		Roles:          []RoleGrant{{Code: "customer"}},
		Permissions:    grants,
		SyncedAtUnix:   now.Unix(),
		FreshUntilUnix: now.Add(30 * time.Second).Unix(),
	}
	if !activeSnap.IsFresh(now) {
		t.Fatal("active snapshot IsFresh() = false, want true")
	}
	if !activeSnap.HasRole("customer") {
		t.Fatal("active snapshot HasRole(customer) = false, want true")
	}
	for _, perm := range selfPerms {
		if _, ok := activeSnap.HasScopedPermission(perm, selfScope); !ok {
			t.Fatalf("active HasScopedPermission(%s, self) = false, want true", perm)
		}
		if _, ok := activeSnap.HasScopedPermission(perm, otherScope); ok {
			t.Fatalf("active HasScopedPermission(%s, otherUser) = true, want false", perm)
		}
	}
}
