package entity_test

import (
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

func TestAccessSnapshot_HasPermissionRejectsNonActiveSelfUser(t *testing.T) {
	t.Parallel()

	userID := "11111111-1111-4111-8111-111111111111"
	otherID := "22222222-2222-4222-8222-222222222222"
	selfScope, err := entity.NewScope(entity.ScopeTypeSelf, &userID)
	if err != nil {
		t.Fatalf("NewScope(self) error = %v", err)
	}
	otherScope, err := entity.NewScope(entity.ScopeTypeSelf, &otherID)
	if err != nil {
		t.Fatalf("NewScope(other) error = %v", err)
	}

	now := time.Now()
	perms := []entity.ScopedPermission{
		{PermissionCode: "iam.users.readself", Scope: selfScope},
		{PermissionCode: "iam.users.updateself", Scope: selfScope},
		{PermissionCode: "iam.users.deleteself", Scope: selfScope},
		{PermissionCode: "iam.sessions.readself", Scope: selfScope},
		{PermissionCode: "iam.sessions.deleteself", IsSensitive: true, Scope: selfScope},
	}

	// Akun PENDING, SUSPENDED, maupun DISABLED tidak boleh lulus HasPermission meski role/permission *self sudah terpasang.
	for _, st := range []entity.UserStatus{
		entity.UserStatusPending,
		entity.UserStatusSuspended,
		entity.UserStatusDisabled,
	} {
		snap := entity.AccessSnapshot{
			UserID:      userID,
			UserKind:    entity.UserKindCustomer,
			UserStatus:  st,
			Permissions: perms,
			CheckedAt:   now,
		}
		for _, p := range perms {
			if _, ok := snap.HasPermission(p.PermissionCode, selfScope); ok {
				t.Fatalf("status=%s HasPermission(%s) = true, want false", st, p.PermissionCode)
			}
		}
	}

	activeSnap := entity.AccessSnapshot{
		UserID:      userID,
		UserKind:    entity.UserKindCustomer,
		UserStatus:  entity.UserStatusActive,
		Permissions: perms,
		CheckedAt:   now,
	}
	for _, p := range perms {
		if _, ok := activeSnap.HasPermission(p.PermissionCode, selfScope); !ok {
			t.Fatalf("active HasPermission(%s, self) = false, want true", p.PermissionCode)
		}
		if _, ok := activeSnap.HasPermission(p.PermissionCode, otherScope); ok {
			t.Fatalf("active HasPermission(%s, other) = true, want false", p.PermissionCode)
		}
	}
}
