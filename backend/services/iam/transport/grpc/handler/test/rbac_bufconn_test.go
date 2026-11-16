package handler_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
)

func TestRegisterAll_FailClosedOnInvalidInputs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := grpc.NewServer()
	defer srv.Stop()

	if err := handler.RegisterAll(nil, &services.Registry{}, logger); err == nil {
		t.Fatal("expected error when grpc server is nil")
	}
	if err := handler.RegisterAll(srv, nil, logger); err == nil {
		t.Fatal("expected error when service registry is nil")
	}
	if err := handler.RegisterAll(srv, &services.Registry{}, logger); err == nil {
		t.Fatal("expected error when service registry has missing dependencies")
	}
}

func TestRBACHandlers_BufconnEndToEnd(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := t.Context()

	healthClient := healthpb.NewHealthClient(env.conn)
	healthResp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: "integrione-iam"})
	if err != nil {
		t.Fatalf("Health.Check error = %v", err)
	}
	if healthResp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", healthResp.GetStatus())
	}

	permClient := iamv1.NewPermissionServiceClient(env.conn)
	roleClient := iamv1.NewRoleServiceClient(env.conn)
	groupClient := iamv1.NewGroupServiceClient(env.conn)
	userClient := iamv1.NewUserServiceClient(env.conn)

	// Buat permission sensitif.
	permResp, err := permClient.CreatePermission(ctx, &iamv1.CreatePermissionRequest{
		Code:        "iam.user.manage",
		Resource:    "iam.user",
		Action:      "manage",
		Description: "Manage IAM accounts",
		IsSensitive: true,
	})
	if err != nil {
		t.Fatalf("CreatePermission error = %v", err)
	}
	perm := permResp.GetPermission()
	if perm.GetCode() != "iam.user.manage" || !perm.GetIsSensitive() {
		t.Fatalf("unexpected created permission: %+v", perm)
	}

	// Query permission by ID & Code & List.
	gotPermByID, err := permClient.GetPermissionById(ctx, &iamv1.GetPermissionByIdRequest{Id: perm.GetId()})
	if err != nil || gotPermByID.GetPermission().GetId() != perm.GetId() {
		t.Fatalf("GetPermissionById failed: resp=%+v err=%v", gotPermByID, err)
	}
	gotPermByCode, err := permClient.GetPermissionByCode(ctx, &iamv1.GetPermissionByCodeRequest{Code: "iam.user.manage"})
	if err != nil || gotPermByCode.GetPermission().GetId() != perm.GetId() {
		t.Fatalf("GetPermissionByCode failed: resp=%+v err=%v", gotPermByCode, err)
	}
	sens := true
	listPerms, err := permClient.ListPermissions(ctx, &iamv1.ListPermissionsRequest{
		Resource:    "iam.user",
		IsSensitive: &sens,
		Page:        &iamv1.PageRequest{Limit: 10, Offset: 0},
	})
	if err != nil || len(listPerms.GetItems()) != 1 {
		t.Fatalf("ListPermissions failed: resp=%+v err=%v", listPerms, err)
	}

	// Buat role dengan redirect path valid dan ikat permission.
	roleResp, err := roleClient.CreateRole(ctx, &iamv1.CreateRoleRequest{
		Code:                "iam_admin",
		Name:                "IAM Administrator",
		Description:         "Manages identity and access controls",
		ManagementDomain:    "iam",
		DefaultRedirectPath: "/identity",
		RedirectPriority:    10,
		IsSystem:            false,
	})
	if err != nil {
		t.Fatalf("CreateRole error = %v", err)
	}
	r := roleResp.GetRole()
	if _, err := roleClient.AssignRolePermission(ctx, &iamv1.AssignRolePermissionRequest{
		RoleId:       r.GetId(),
		PermissionId: perm.GetId(),
	}); err != nil {
		t.Fatalf("AssignRolePermission error = %v", err)
	}
	rolePerms, err := roleClient.ListRolePermissions(ctx, &iamv1.ListRolePermissionsRequest{RoleId: r.GetId()})
	if err != nil || len(rolePerms.GetPermissions()) != 1 {
		t.Fatalf("ListRolePermissions failed: resp=%+v err=%v", rolePerms, err)
	}
	byRolePerms, err := permClient.ListPermissionsByRole(ctx, &iamv1.ListPermissionsByRoleRequest{RoleId: r.GetId()})
	if err != nil || len(byRolePerms.GetPermissions()) != 1 {
		t.Fatalf("ListPermissionsByRole failed: resp=%+v err=%v", byRolePerms, err)
	}

	// Buat user & grup, tambahkan keanggotaan & peran berscope.
	createdUser, err := userClient.CreateUser(ctx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_EMPLOYEE,
		Status:                 iamv1.UserStatus_USER_STATUS_ACTIVE,
		DisplayName:            "Rina Admin",
		Email:                  "rina.admin@integrione.local",
		Username:               "rina.admin",
		Password:               "StrongPass!2026#",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser error = %v", err)
	}
	uid := createdUser.GetUser().GetId()
	externalScopeID := "77777777-7777-4777-8777-777777777777"
	_, err = roleClient.AssignUserRole(ctx, &iamv1.AssignUserRoleRequest{
		UserId: uid, RoleId: r.GetId(), ScopeType: iamv1.ScopeType_SCOPE_TYPE_LEGAL_ENTITY,
		ScopeId: &externalScopeID,
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("external user role scope status = %s, want Unavailable", status.Code(err))
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	assignUR, err := roleClient.AssignUserRole(ctx, &iamv1.AssignUserRoleRequest{
		UserId:    uid,
		RoleId:    r.GetId(),
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
		ExpiresAt: timestamppb.New(expiresAt),
	})
	if err != nil {
		t.Fatalf("AssignUserRole error = %v", err)
	}
	if assignUR.GetUserRole().GetExpiresAt().AsTime().Unix() != expiresAt.Unix() {
		t.Fatalf("expected expires_at unix %d, got %d", expiresAt.Unix(), assignUR.GetUserRole().GetExpiresAt().AsTime().Unix())
	}

	grpResp, err := groupClient.CreateGroup(ctx, &iamv1.CreateGroupRequest{
		Code:        "iam_ops",
		Name:        "IAM Operations",
		Description: "Identity operations team",
	})
	if err != nil {
		t.Fatalf("CreateGroup error = %v", err)
	}
	gid := grpResp.GetGroup().GetId()
	_, err = groupClient.AssignGroupRole(ctx, &iamv1.AssignGroupRoleRequest{
		GroupId: gid, RoleId: r.GetId(), ScopeType: iamv1.ScopeType_SCOPE_TYPE_FACILITY,
		ScopeId: &externalScopeID,
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("external group role scope status = %s, want Unavailable", status.Code(err))
	}
	if _, err := groupClient.AddGroupMember(ctx, &iamv1.AddGroupMemberRequest{
		GroupId: gid,
		UserId:  uid,
	}); err != nil {
		t.Fatalf("AddGroupMember error = %v", err)
	}
	if _, err := groupClient.AssignGroupRole(ctx, &iamv1.AssignGroupRoleRequest{
		GroupId:   gid,
		RoleId:    r.GetId(),
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
	}); err != nil {
		t.Fatalf("AssignGroupRole error = %v", err)
	}
	membersResp, err := groupClient.ListGroupMembers(ctx, &iamv1.ListGroupMembersRequest{GroupId: gid})
	if err != nil || len(membersResp.GetItems()) != 1 {
		t.Fatalf("ListGroupMembers failed: resp=%+v err=%v", membersResp, err)
	}
	effRoles, err := roleClient.ListEffectiveUserRoles(ctx, &iamv1.ListEffectiveUserRolesRequest{UserId: uid})
	if err != nil || len(effRoles.GetRoles()) != 1 {
		t.Fatalf("ListEffectiveUserRoles failed: resp=%+v err=%v", effRoles, err)
	}

	// Uji error status gRPC untuk ID tidak valid dan duplikasi.
	_, err = roleClient.GetRoleById(ctx, &iamv1.GetRoleByIdRequest{Id: "invalid-uuid"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument on invalid UUID, got %s (%v)", status.Code(err), err)
	}
	_, err = roleClient.CreateRole(ctx, &iamv1.CreateRoleRequest{
		Code:                "iam_admin",
		Name:                "Dup",
		DefaultRedirectPath: "/identity",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists on duplicate role code, got %s (%v)", status.Code(err), err)
	}
}
