package user_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

func TestRequisitionApprovalMandateUsesExactActiveEmployeeGrant(t *testing.T) {
	t.Parallel()
	const principal = "00000000-0000-4000-8000-000000000001"
	unit := "00000000-0000-4000-8000-000000000002"
	otherUnit := "00000000-0000-4000-8000-000000000003"
	h := newTestHarness(t)
	account := &entity.User{ID: principal, Kind: entity.UserKindEmployee, Status: entity.UserStatusActive}
	h.store.users[principal] = account
	global := entity.ScopedPermission{
		PermissionCode: "human_resource.requisitions.approve", Scope: entity.GlobalScope(),
	}
	scoped := entity.ScopedPermission{
		PermissionCode: global.PermissionCode, Scope: entity.Scope{Type: entity.ScopeTypeOrganizationUnit, ID: &unit},
	}
	for _, tc := range []struct {
		name   string
		unit   string
		kind   entity.UserKind
		status entity.UserStatus
		grants []entity.ScopedPermission
		want   bool
	}{
		{name: "global alone", unit: unit, kind: entity.UserKindEmployee, status: entity.UserStatusActive, grants: []entity.ScopedPermission{global}},
		{name: "exact unit", unit: unit, kind: entity.UserKindEmployee, status: entity.UserStatusActive, grants: []entity.ScopedPermission{global, scoped}, want: true},
		{name: "other unit", unit: otherUnit, kind: entity.UserKindEmployee, status: entity.UserStatusActive, grants: []entity.ScopedPermission{global, scoped}},
		{name: "revoked grant", unit: unit, kind: entity.UserKindEmployee, status: entity.UserStatusActive},
		{name: "disabled account", unit: unit, kind: entity.UserKindEmployee, status: entity.UserStatusDisabled, grants: []entity.ScopedPermission{scoped}},
		{name: "vendor", unit: unit, kind: entity.UserKindVendor, status: entity.UserStatusActive, grants: []entity.ScopedPermission{scoped}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account.Kind, account.Status = tc.kind, tc.status
			h.store.permissions[principal] = tc.grants
			allowed, err := h.userSvc.HasRequisitionApprovalMandate(t.Context(), principal, tc.unit)
			if err != nil || allowed != tc.want {
				t.Fatalf("approval mandate = %v, %v; want %v", allowed, err, tc.want)
			}
		})
	}
	if _, err := h.userSvc.HasRequisitionApprovalMandate(t.Context(), principal, "invalid"); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("invalid unit: %v", err)
	}
}
