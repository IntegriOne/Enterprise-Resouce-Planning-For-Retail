package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

func TestRedirectResolver(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	resolver := service.NewDefaultRedirectResolver()

	whRole, _ := entity.NewRole("11111111-1111-4111-8111-111111111111", "warehouse_staff", "Warehouse", "", "wms", "/warehouse", 30, false, now)
	iamRole, _ := entity.NewRole("22222222-2222-4222-8222-222222222222", "iam_admin", "IAM Admin", "", "iam", "/identity", 10, true, now)
	hrRole, _ := entity.NewRole("33333333-3333-4333-8333-333333333333", "hr_manager", "HR", "", "hrm", "/human-resource", 10, false, now)
	inactiveRole, _ := entity.NewRole("44444444-4444-4444-8444-444444444444", "super_admin", "Super Admin", "", "iam", "/identity", 1, false, now)
	_ = inactiveRole.Deactivate(now)

	t.Run("lowest RedirectPriority wins among active roles", func(t *testing.T) {
		t.Parallel()
		path, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, []entity.Role{*whRole, *iamRole})
		if err != nil {
			t.Fatalf("ResolvePostLoginPath() unexpected error: %v", err)
		}
		if path != "/identity" {
			t.Fatalf("path = %q, want /identity", path)
		}
	})

	t.Run("inactive high-priority role is ignored", func(t *testing.T) {
		t.Parallel()
		path, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, []entity.Role{*inactiveRole, *whRole})
		if err != nil {
			t.Fatalf("ResolvePostLoginPath() unexpected error: %v", err)
		}
		if path != "/warehouse" {
			t.Fatalf("path = %q, want /warehouse", path)
		}
	})

	t.Run("equal priority uses deterministic tie-breaker", func(t *testing.T) {
		t.Parallel()
		p1, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, []entity.Role{*iamRole, *hrRole})
		if err != nil {
			t.Fatalf("ResolvePostLoginPath() error: %v", err)
		}
		p2, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, []entity.Role{*hrRole, *iamRole})
		if err != nil {
			t.Fatalf("ResolvePostLoginPath() error: %v", err)
		}
		if p1 != p2 || p1 != "/human-resource" {
			t.Fatalf("tie-breaker not deterministic: p1=%q p2=%q want /human-resource", p1, p2)
		}
	})

	t.Run("fallback path by user kind when no role defines redirect", func(t *testing.T) {
		t.Parallel()
		noRedirectRole, _ := entity.NewRole("55555555-5555-4555-8555-555555555555", "viewer", "Viewer", "", "iam", "", 50, false, now)

		custPath, err := resolver.ResolvePostLoginPath(entity.UserKindCustomer, []entity.Role{*noRedirectRole})
		if err != nil || custPath != "/" {
			t.Fatalf("customer fallback = (%q, %v), want /", custPath, err)
		}
		vendorPath, err := resolver.ResolvePostLoginPath(entity.UserKindVendor, nil)
		if err != nil || vendorPath != "/vendor" {
			t.Fatalf("vendor fallback = (%q, %v), want /vendor", vendorPath, err)
		}
		supPath, err := resolver.ResolvePostLoginPath(entity.UserKindSupplier, nil)
		if err != nil || supPath != "/supplier" {
			t.Fatalf("supplier fallback = (%q, %v), want /supplier", supPath, err)
		}
		empPath, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, nil)
		if err != nil || empPath != "/account" {
			t.Fatalf("employee fallback = (%q, %v), want /account", empPath, err)
		}
	})

	t.Run("non-allowlisted internal path is rejected fail-closed", func(t *testing.T) {
		t.Parallel()
		unlistedRole, _ := entity.NewRole("66666666-6666-4666-8666-666666666666", "weird", "Weird", "", "other", "/hidden-admin", 5, false, now)
		if _, err := resolver.ResolvePostLoginPath(entity.UserKindEmployee, []entity.Role{*unlistedRole}); !errors.Is(err, iamerrors.ErrRedirectPathInvalid) {
			t.Fatalf("unlisted redirect error = %v, want ErrRedirectPathInvalid", err)
		}
	})

	t.Run("custom allowlist and subpath matching", func(t *testing.T) {
		t.Parallel()
		custom, err := service.NewRedirectResolver([]string{"/warehouse", "/finance"})
		if err != nil {
			t.Fatalf("NewRedirectResolver() error: %v", err)
		}
		if err := custom.ValidateAllowedPath("/warehouse/bins"); err != nil {
			t.Fatalf("ValidateAllowedPath(/warehouse/bins) error: %v", err)
		}
		if _, err := service.NewRedirectResolver([]string{"https://evil.com"}); !errors.Is(err, iamerrors.ErrRedirectPathInvalid) {
			t.Fatalf("NewRedirectResolver(evil) error = %v, want ErrRedirectPathInvalid", err)
		}
		if _, err := resolver.ResolvePostLoginPath(entity.UserKind("invalid"), nil); !errors.Is(err, iamerrors.ErrInvalidUserKind) {
			t.Fatalf("invalid UserKind error = %v, want ErrInvalidUserKind", err)
		}
	})
}
