package iam_test

import (
	"strings"
	"testing"

	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
)

// TestParseTypedID_RoleAndPermissionGIDAndRawUUID memverifikasi ParseTypedID menormalkan GID dan UUID mentah untuk ResourceRole dan ResourcePermission serta menolak resource yang salah.
func TestParseTypedID_RoleAndPermissionGIDAndRawUUID(t *testing.T) {
	t.Parallel()

	const wantUUID = "0b95f8c5-5e2c-4980-a33a-034a60c98886"
	tests := []struct {
		name     string
		expected iamadapter.GIDResource
		raw      string
		want     string
		wantErr  bool
	}{
		{
			name:     "role_gid_lowercase",
			expected: iamadapter.ResourceRole,
			raw:      "gid://integrione/role/" + wantUUID,
			want:     wantUUID,
		},
		{
			name:     "role_gid_uppercase_padded",
			expected: iamadapter.ResourceRole,
			raw:      "  gid://integrione/role/" + strings.ToUpper(wantUUID) + "  ",
			want:     wantUUID,
		},
		{
			name:     "role_raw_uuid_lowercase",
			expected: iamadapter.ResourceRole,
			raw:      wantUUID,
			want:     wantUUID,
		},
		{
			name:     "role_raw_uuid_uppercase_padded",
			expected: iamadapter.ResourceRole,
			raw:      " " + strings.ToUpper(wantUUID) + " ",
			want:     wantUUID,
		},
		{
			name:     "permission_gid_lowercase",
			expected: iamadapter.ResourcePermission,
			raw:      "gid://integrione/permission/" + wantUUID,
			want:     wantUUID,
		},
		{
			name:     "permission_gid_uppercase_padded",
			expected: iamadapter.ResourcePermission,
			raw:      "  gid://integrione/permission/" + strings.ToUpper(wantUUID) + "  ",
			want:     wantUUID,
		},
		{
			name:     "permission_raw_uuid_lowercase",
			expected: iamadapter.ResourcePermission,
			raw:      wantUUID,
			want:     wantUUID,
		},
		{
			name:     "permission_raw_uuid_uppercase_padded",
			expected: iamadapter.ResourcePermission,
			raw:      "\t" + strings.ToUpper(wantUUID) + "\n",
			want:     wantUUID,
		},
		{
			name:     "role_rejects_permission_gid",
			expected: iamadapter.ResourceRole,
			raw:      "gid://integrione/permission/" + wantUUID,
			wantErr:  true,
		},
		{
			name:     "permission_rejects_role_gid",
			expected: iamadapter.ResourcePermission,
			raw:      "gid://integrione/role/" + wantUUID,
			wantErr:  true,
		},
		{
			name:     "role_rejects_user_gid",
			expected: iamadapter.ResourceRole,
			raw:      "gid://integrione/user/" + wantUUID,
			wantErr:  true,
		},
		{
			name:     "permission_rejects_wrong_prefix",
			expected: iamadapter.ResourcePermission,
			raw:      "gid://other/permission/" + wantUUID,
			wantErr:  true,
		},
		{
			name:     "role_rejects_unsafe_raw_chars",
			expected: iamadapter.ResourceRole,
			raw:      "not a valid/uuid",
			wantErr:  true,
		},
		{
			name:     "permission_rejects_unsafe_query_chars",
			expected: iamadapter.ResourcePermission,
			raw:      "perm?bad=1",
			wantErr:  true,
		},
		{
			name:     "permission_rejects_empty",
			expected: iamadapter.ResourcePermission,
			raw:      "   ",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := iamadapter.ParseTypedID(tc.raw, tc.expected)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseTypedID(%q, %q) expected error, got %q", tc.raw, tc.expected, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTypedID(%q, %q) unexpected error: %v", tc.raw, tc.expected, err)
			}
			if got != tc.want {
				t.Fatalf("ParseTypedID(%q, %q) = %q, want %q", tc.raw, tc.expected, got, tc.want)
			}
		})
	}
}

// TestAdapter_AssignAndRevokeRolePermission_AcceptsGIDAndRawUUID memverifikasi assignRolePermission dan revokeRolePermission menerima kombinasi GID dan UUID mentah (lowercase/uppercase), meneruskan UUID kanonik ke IAM gRPC, tetap mengembalikan GID standar Relay pada output, serta menolak ID tidak valid sebelum RPC.
func TestAdapter_AssignAndRevokeRolePermission_AcceptsGIDAndRawUUID(t *testing.T) {
	t.Parallel()

	const (
		selfUUID  = "11111111-1111-4111-8111-111111111111"
		adminUUID = "22222222-2222-4222-8222-222222222222"
		roleUUID  = "0b95f8c5-5e2c-4980-a33a-034a60c98886"
		permUUID  = "818542c6-d7eb-4e86-a4f6-cfed4b1bada1"
	)
	wantRoleGID := "gid://integrione/role/" + roleUUID
	wantPermGID := "gid://integrione/permission/" + permUUID

	h := newGIDTestHarness(t, selfUUID, adminUUID)

	validCases := []struct {
		name   string
		roleID string
		permID string
	}{
		{
			name:   "gid_role_and_gid_perm",
			roleID: wantRoleGID,
			permID: wantPermGID,
		},
		{
			name:   "raw_lower_role_and_raw_lower_perm",
			roleID: roleUUID,
			permID: permUUID,
		},
		{
			name:   "gid_role_and_raw_lower_perm",
			roleID: wantRoleGID,
			permID: permUUID,
		},
		{
			name:   "raw_lower_role_and_gid_perm",
			roleID: roleUUID,
			permID: wantPermGID,
		},
		{
			name:   "raw_upper_padded_role_and_perm",
			roleID: "  " + strings.ToUpper(roleUUID) + "  ",
			permID: " " + strings.ToUpper(permUUID) + " ",
		},
		{
			name:   "gid_upper_role_and_raw_upper_perm",
			roleID: "gid://integrione/role/" + strings.ToUpper(roleUUID),
			permID: strings.ToUpper(permUUID),
		},
	}

	for _, tc := range validCases {
		t.Run("Assign/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			payload, err := h.adapter.AssignRolePermission(h.adminCtx, tc.roleID, tc.permID)
			if err != nil {
				t.Fatalf("AssignRolePermission(%q, %q) unexpected Go error: %v", tc.roleID, tc.permID, err)
			}
			if payload == nil || !payload.WasSuccessful || len(payload.Errors) != 0 {
				t.Fatalf("AssignRolePermission(%q, %q) unexpected failure payload: %+v", tc.roleID, tc.permID, payload)
			}
			if h.roles.callCount != 1 || h.roles.lastRoleID != roleUUID || h.roles.lastPermID != permUUID {
				t.Fatalf("AssignRolePermission(%q, %q) forwarded role=%q perm=%q calls=%d, want role=%q perm=%q calls=1",
					tc.roleID, tc.permID, h.roles.lastRoleID, h.roles.lastPermID, h.roles.callCount, roleUUID, permUUID)
			}
			if h.perms.callCount != 1 || h.perms.lastPermID != permUUID {
				t.Fatalf("GetPermissionById forwarded perm=%q calls=%d, want %q calls=1",
					h.perms.lastPermID, h.perms.callCount, permUUID)
			}
			edge := payload.RolePermissionEdge
			if edge == nil || edge.RoleID != wantRoleGID {
				t.Fatalf("RolePermissionEdge.RoleID = %#v, want %q", edge, wantRoleGID)
			}
			if edge.Node == nil || edge.Node.ID != wantPermGID || edge.Node.RawID != permUUID {
				t.Fatalf("RolePermissionEdge.Node = %+v, want id=%q _id=%q", edge.Node, wantPermGID, permUUID)
			}
		})

		t.Run("Revoke/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			payload, err := h.adapter.RevokeRolePermission(h.adminCtx, tc.roleID, tc.permID)
			if err != nil {
				t.Fatalf("RevokeRolePermission(%q, %q) unexpected Go error: %v", tc.roleID, tc.permID, err)
			}
			if payload == nil || !payload.WasSuccessful || len(payload.Errors) != 0 || payload.Revoked == nil || !*payload.Revoked {
				t.Fatalf("RevokeRolePermission(%q, %q) unexpected failure payload: %+v", tc.roleID, tc.permID, payload)
			}
			if h.roles.callCount != 1 || h.roles.lastRoleID != roleUUID || h.roles.lastPermID != permUUID {
				t.Fatalf("RevokeRolePermission(%q, %q) forwarded role=%q perm=%q calls=%d, want role=%q perm=%q calls=1",
					tc.roleID, tc.permID, h.roles.lastRoleID, h.roles.lastPermID, h.roles.callCount, roleUUID, permUUID)
			}
		})
	}

	invalidCases := []struct {
		name   string
		roleID string
		permID string
	}{
		{
			name:   "swapped_permission_gid_for_role_id",
			roleID: wantPermGID,
			permID: wantPermGID,
		},
		{
			name:   "swapped_role_gid_for_permission_id",
			roleID: wantRoleGID,
			permID: wantRoleGID,
		},
		{
			name:   "wrong_user_gid_for_role_id",
			roleID: "gid://integrione/user/" + roleUUID,
			permID: permUUID,
		},
		{
			name:   "wrong_user_gid_for_permission_id",
			roleID: roleUUID,
			permID: "gid://integrione/user/" + permUUID,
		},
		{
			name:   "unsafe_chars_for_role_id",
			roleID: "invalid role/id",
			permID: permUUID,
		},
		{
			name:   "unsafe_chars_for_permission_id",
			roleID: roleUUID,
			permID: "invalid?perm#id",
		},
		{
			name:   "empty_role_id",
			roleID: "   ",
			permID: permUUID,
		},
		{
			name:   "empty_permission_id",
			roleID: roleUUID,
			permID: "   ",
		},
	}

	for _, tc := range invalidCases {
		t.Run("RejectAssign/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			payload, err := h.adapter.AssignRolePermission(h.adminCtx, tc.roleID, tc.permID)
			if err != nil {
				t.Fatalf("AssignRolePermission(%q, %q) unexpected Go error: %v", tc.roleID, tc.permID, err)
			}
			if payload == nil || payload.WasSuccessful || len(payload.Errors) == 0 {
				t.Fatalf("AssignRolePermission(%q, %q) expected validation failure payload, got %+v", tc.roleID, tc.permID, payload)
			}
			if h.roles.callCount != 0 || h.perms.callCount != 0 {
				t.Fatalf("invalid ID should fail before downstream RPC (roleCalls=%d permCalls=%d)", h.roles.callCount, h.perms.callCount)
			}
		})

		t.Run("RejectRevoke/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			payload, err := h.adapter.RevokeRolePermission(h.adminCtx, tc.roleID, tc.permID)
			if err != nil {
				t.Fatalf("RevokeRolePermission(%q, %q) unexpected Go error: %v", tc.roleID, tc.permID, err)
			}
			if payload == nil || payload.WasSuccessful || len(payload.Errors) == 0 {
				t.Fatalf("RevokeRolePermission(%q, %q) expected validation failure payload, got %+v", tc.roleID, tc.permID, payload)
			}
			if h.roles.callCount != 0 || h.perms.callCount != 0 {
				t.Fatalf("invalid ID should fail before downstream RPC (roleCalls=%d permCalls=%d)", h.roles.callCount, h.perms.callCount)
			}
		})
	}
}

// TestAdapter_RoleAndRolePermissions_AcceptGIDAndRawUUID memverifikasi query Role dan RolePermissions menerima GID maupun UUID mentah serta menolak GID resource yang salah.
func TestAdapter_RoleAndRolePermissions_AcceptGIDAndRawUUID(t *testing.T) {
	t.Parallel()

	const (
		selfUUID  = "11111111-1111-4111-8111-111111111111"
		adminUUID = "22222222-2222-4222-8222-222222222222"
		roleUUID  = "0b95f8c5-5e2c-4980-a33a-034a60c98886"
		permUUID  = "818542c6-d7eb-4e86-a4f6-cfed4b1bada1"
	)
	wantRoleGID := "gid://integrione/role/" + roleUUID
	wantPermGID := "gid://integrione/permission/" + permUUID

	h := newGIDTestHarness(t, selfUUID, adminUUID)
	h.roles.rolePermissions[roleUUID] = []*iamv1.Permission{h.perms.samplePermission(permUUID)}

	inputs := []struct {
		name string
		val  string
	}{
		{name: "gid_lowercase", val: wantRoleGID},
		{name: "gid_uppercase_padded", val: "  gid://integrione/role/" + strings.ToUpper(roleUUID) + "  "},
		{name: "raw_uuid_lowercase", val: roleUUID},
		{name: "raw_uuid_uppercase_padded", val: " " + strings.ToUpper(roleUUID) + " "},
	}

	for _, tc := range inputs {
		t.Run("Role/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			role, err := h.adapter.Role(h.adminCtx, tc.val)
			if err != nil {
				t.Fatalf("Role(%q) unexpected error: %v", tc.val, err)
			}
			if h.roles.callCount != 1 || h.roles.lastRoleID != roleUUID {
				t.Fatalf("Role(%q) forwarded role_id=%q calls=%d, want %q calls=1", tc.val, h.roles.lastRoleID, h.roles.callCount, roleUUID)
			}
			if role == nil || role.ID != wantRoleGID || role.RawID != roleUUID {
				t.Fatalf("Role(%q) returned %+v, want id=%q _id=%q", tc.val, role, wantRoleGID, roleUUID)
			}
		})

		t.Run("RolePermissions/"+tc.name, func(t *testing.T) {
			h.resetCalls()
			conn, err := h.adapter.RolePermissions(h.adminCtx, tc.val, nil, nil, nil, nil)
			if err != nil {
				t.Fatalf("RolePermissions(%q) unexpected error: %v", tc.val, err)
			}
			if h.roles.callCount != 1 || h.roles.lastRoleID != roleUUID {
				t.Fatalf("RolePermissions(%q) forwarded role_id=%q calls=%d, want %q calls=1", tc.val, h.roles.lastRoleID, h.roles.callCount, roleUUID)
			}
			if conn == nil || len(conn.Edges) != 1 || conn.Edges[0].RoleID != wantRoleGID || conn.Edges[0].Node.ID != wantPermGID {
				t.Fatalf("RolePermissions(%q) returned %+v", tc.val, conn)
			}
		})
	}

	t.Run("RejectsWrongResourceGID", func(t *testing.T) {
		h.resetCalls()
		if _, err := h.adapter.Role(h.adminCtx, wantPermGID); err == nil {
			t.Fatalf("Role(%q) expected error for permission GID", wantPermGID)
		}
		if _, err := h.adapter.RolePermissions(h.adminCtx, wantPermGID, nil, nil, nil, nil); err == nil {
			t.Fatalf("RolePermissions(%q) expected error for permission GID", wantPermGID)
		}
		if h.roles.callCount != 0 {
			t.Fatalf("wrong resource GID should fail before RPC call (calls=%d)", h.roles.callCount)
		}
	})
}
