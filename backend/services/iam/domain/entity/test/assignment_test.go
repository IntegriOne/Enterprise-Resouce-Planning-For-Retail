package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	testAssignmentID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	testFacilityID1  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	testFacilityID2  = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
)

func TestScopeInvariantsAndMatching(t *testing.T) {
	t.Parallel()

	global := entity.GlobalScope()
	fac1ID := testFacilityID1
	fac2ID := testFacilityID2

	fac1, err := entity.NewScope(entity.ScopeTypeFacility, &fac1ID)
	if err != nil {
		t.Fatalf("NewScope(facility) error = %v", err)
	}
	fac2, err := entity.NewScope(entity.ScopeTypeFacility, &fac2ID)
	if err != nil {
		t.Fatalf("NewScope(facility2) error = %v", err)
	}

	// Global scope tidak boleh membawa scope_id.
	if _, err := entity.NewScope(entity.ScopeTypeGlobal, &fac1ID); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("global scope with ID error = %v, want ErrInvalidScope", err)
	}
	// Non-global scope wajib membawa scope_id UUID valid.
	if _, err := entity.NewScope(entity.ScopeTypeFacility, nil); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("facility scope without ID error = %v, want ErrInvalidScope", err)
	}
	badUUID := "not-uuid"
	if _, err := entity.NewScope(entity.ScopeTypeFacility, &badUUID); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("facility scope invalid UUID error = %v, want ErrInvalidScope", err)
	}

	if !global.Matches(fac1) {
		t.Fatal("global granted scope must satisfy facility required scope")
	}
	if !fac1.Matches(fac1) {
		t.Fatal("identical facility scope must match")
	}
	if fac1.Matches(fac2) {
		t.Fatal("different facility scopes must not match")
	}
	if fac1.Matches(global) {
		t.Fatal("facility-scoped grant must not satisfy global required scope")
	}
}

func TestUserRoleGroupRoleAndAccessSnapshotClone(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	exp := now.Add(time.Hour)
	fac1ID := testFacilityID1
	fac1, err := entity.NewScope(entity.ScopeTypeFacility, &fac1ID)
	if err != nil {
		t.Fatalf("NewScope() error = %v", err)
	}

	ur, err := entity.NewUserRole(testAssignmentID, testUserID, testRoleID, fac1, nil, &exp, now)
	if err != nil {
		t.Fatalf("NewUserRole() error = %v", err)
	}
	if !ur.IsEffective(now.Add(30 * time.Minute)) {
		t.Fatal("UserRole should be effective before ExpiresAt")
	}
	if ur.IsEffective(now.Add(time.Hour)) {
		t.Fatal("UserRole should be expired at ExpiresAt boundary")
	}

	gr, err := entity.NewGroupRole(testAssignmentID, testGroupID, testRoleID, entity.GlobalScope(), nil, nil, now)
	if err != nil {
		t.Fatalf("NewGroupRole() error = %v", err)
	}
	if !gr.IsEffective(now.Add(24 * time.Hour)) {
		t.Fatal("non-expiring GroupRole should remain effective")
	}

	snap := entity.AccessSnapshot{
		UserID:     testUserID,
		UserKind:   entity.UserKindEmployee,
		UserStatus: entity.UserStatusActive,
		MFAEnabled: true,
		Permissions: []entity.ScopedPermission{
			{
				PermissionCode: "iam.user.read",
				IsSensitive:    false,
				Scope:          fac1,
			},
		},
		CheckedAt: now,
	}

	if _, ok := snap.HasPermission("IAM.User.Read", fac1); !ok {
		t.Fatal("HasPermission() should match case-insensitively on matching scope")
	}
	if _, ok := snap.HasPermission("iam.user.read", entity.GlobalScope()); ok {
		t.Fatal("facility permission must not satisfy global requirement")
	}

	cloned := snap.Clone()
	cloned.Permissions[0].PermissionCode = "mutated"
	if _, ok := snap.HasPermission("iam.user.read", fac1); !ok {
		t.Fatal("mutating cloned snapshot must not mutate original snapshot")
	}
}

func TestSelfScopeRoleAndSnapshotGuards(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.Local)
	userA := testUserID
	userB := testFacilityID2

	scopeA, err := entity.NewScope(entity.ScopeTypeSelf, &userA)
	if err != nil {
		t.Fatalf("NewScope(self, userA) error = %v", err)
	}
	scopeB, err := entity.NewScope(entity.ScopeTypeSelf, &userB)
	if err != nil {
		t.Fatalf("NewScope(self, userB) error = %v", err)
	}

	if _, err := entity.NewUserRole(testAssignmentID, userA, testRoleID, scopeA, nil, nil, now); err != nil {
		t.Fatalf("NewUserRole matching self scope error = %v", err)
	}
	if _, err := entity.NewUserRole(testAssignmentID, userA, testRoleID, scopeB, nil, nil, now); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("NewUserRole mismatched self scope error = %v, want ErrInvalidScope", err)
	}
	if _, err := entity.NewGroupRole(testAssignmentID, testGroupID, testRoleID, scopeA, nil, nil, now); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("NewGroupRole(ScopeTypeSelf) error = %v, want ErrInvalidScope", err)
	}

	snap := entity.AccessSnapshot{
		UserID:     userA,
		UserKind:   entity.UserKindCustomer,
		UserStatus: entity.UserStatusActive,
		Permissions: []entity.ScopedPermission{
			{
				PermissionCode: "iam.users.readself",
				Scope:          scopeA,
			},
			{
				PermissionCode: "iam.users.updateself",
				IsSensitive:    false,
				Scope:          entity.GlobalScope(),
			},
		},
		CheckedAt: now,
	}

	if _, ok := snap.HasPermission("iam.users.readself", entity.GlobalScope()); ok {
		t.Fatal("*self permission must reject Global required scope")
	}
	if _, ok := snap.HasPermission("iam.users.updateself", entity.GlobalScope()); ok {
		t.Fatal("*self permission must reject Global required scope even with global grant")
	}
	if _, ok := snap.HasPermission("iam.users.readself", scopeB); ok {
		t.Fatal("*self permission must reject another user's self scope")
	}
	if _, ok := snap.HasPermission("iam.users.readself", scopeA); !ok {
		t.Fatal("*self permission must allow matching own self scope")
	}
}
