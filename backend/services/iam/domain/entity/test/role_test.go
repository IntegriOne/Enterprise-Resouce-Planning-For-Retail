package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	testRoleID       = "77777777-7777-4777-8777-777777777777"
	testPermissionID = "88888888-8888-4888-8888-888888888888"
)

func TestRoleCreationAndSystemProtection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	role, err := entity.NewRole(testRoleID, " Super_Admin ", "Super Administrator", "Full IAM control", "iam", "/identity", 10, true, now)
	if err != nil {
		t.Fatalf("NewRole() unexpected error: %v", err)
	}
	if role.Code != "super_admin" {
		t.Fatalf("Role.Code = %q, want super_admin", role.Code)
	}
	if !role.IsActive {
		t.Fatal("new role should default to active")
	}
	if err := role.Deactivate(now.Add(time.Minute)); !errors.Is(err, iamerrors.ErrSystemRoleImmutable) {
		t.Fatalf("deactivating system role = %v, want ErrSystemRoleImmutable", err)
	}

	customRole, err := entity.NewRole(testRoleID, "warehouse_staff", "Warehouse Staff", "Ops", "wms", "/warehouse", 50, false, now)
	if err != nil {
		t.Fatalf("NewRole(custom) error = %v", err)
	}
	if err := customRole.UpdateDetails("Warehouse Lead", "Updated ops", "wms", "/warehouse/bins", 25, now.Add(30*time.Second)); err != nil {
		t.Fatalf("UpdateDetails() unexpected error: %v", err)
	}
	if customRole.DefaultRedirectPath != "/warehouse/bins" || customRole.RedirectPriority != 25 {
		t.Fatalf("UpdateDetails() did not apply expected changes: %+v", customRole)
	}
	customRole.Deactivate(now.Add(time.Minute))
	if customRole.IsActive {
		t.Fatal("customRole should be inactive after Deactivate")
	}
	customRole.Activate(now.Add(2 * time.Minute))
	if !customRole.IsActive {
		t.Fatal("customRole should be active after Activate")
	}

	if _, err := entity.NewRole(testRoleID, "bad code!", "Name", "", "iam", "/identity", 10, false, now); !errors.Is(err, iamerrors.ErrInvalidRoleCode) {
		t.Fatalf("invalid role code error = %v, want ErrInvalidRoleCode", err)
	}
	if _, err := entity.NewRole(testRoleID, "role_a", "Name", "", "iam", "/identity", -1, false, now); !errors.Is(err, iamerrors.ErrInvalidRedirectPriority) {
		t.Fatalf("negative redirect priority error = %v, want ErrInvalidRedirectPriority", err)
	}
	if _, err := entity.NewRole(testRoleID, "role_b", "Name", "", "iam", "https://evil.com", 10, false, now); !errors.Is(err, iamerrors.ErrRedirectPathInvalid) {
		t.Fatalf("external redirect path error = %v, want ErrRedirectPathInvalid", err)
	}
}

func TestPermissionAndRolePermission(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)

	perm, err := entity.NewPermission(testPermissionID, " IAM.Role:Grant ", " role ", " GRANT ", "Grant role to principal", true, now)
	if err != nil {
		t.Fatalf("NewPermission() unexpected error: %v", err)
	}
	if perm.Code != "iam.role:grant" || perm.Resource != "role" || perm.Action != "grant" {
		t.Fatalf("normalized permission = %+v", perm)
	}

	if _, err := entity.NewPermission(testPermissionID, "invalid code!", "role", "grant", "", false, now); !errors.Is(err, iamerrors.ErrInvalidPermissionCode) {
		t.Fatalf("invalid permission code error = %v, want ErrInvalidPermissionCode", err)
	}

	rp, err := entity.NewRolePermission(testRoleID, testPermissionID, now)
	if err != nil {
		t.Fatalf("NewRolePermission() unexpected error: %v", err)
	}
	if rp.RoleID != testRoleID || rp.PermissionID != testPermissionID {
		t.Fatalf("RolePermission mismatch: %+v", rp)
	}
}
