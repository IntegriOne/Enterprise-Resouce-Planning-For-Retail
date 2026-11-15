package iam_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type gidMemStore struct {
	sessions map[string]*authsession.SessionValidationContext
	preAuths map[string]authsession.PreAuthRecord
	rates    map[string]int64
}

func newGIDMemStore() *gidMemStore {
	return &gidMemStore{
		sessions: make(map[string]*authsession.SessionValidationContext),
		preAuths: make(map[string]authsession.PreAuthRecord),
		rates:    make(map[string]int64),
	}
}

func (s *gidMemStore) GetGate(context.Context) (*authsession.SecurityGateRecord, error) {
	return &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1}, nil
}

func (s *gidMemStore) ReadSessionContext(_ context.Context, tokenHashHex string, now time.Time) (*authsession.SessionValidationContext, error) {
	if sc, ok := s.sessions[tokenHashHex]; ok {
		if sc.Token != nil && sc.Token.Revoked {
			return nil, authsession.ErrSessionRevoked
		}
		if sc.Security != nil && (strings.ToLower(strings.TrimSpace(sc.Security.Status)) != authz.StatusActive || sc.Security.LockedUntilUnix > now.Unix()) {
			return nil, authsession.ErrUserSecurityBlocked
		}
		return sc, nil
	}
	return nil, authsession.ErrSessionNotFound
}

func (s *gidMemStore) TouchActivity(context.Context, string, authsession.SessionTokenRecord, time.Duration, time.Time) error {
	return nil
}

func (s *gidMemStore) SetPreAuth(_ context.Context, rec authsession.PreAuthRecord, _ time.Duration) error {
	s.preAuths[rec.TokenHashHex] = rec
	return nil
}

func (s *gidMemStore) GetPreAuth(_ context.Context, cookieHashHex string, _ time.Time) (*authsession.PreAuthRecord, error) {
	if rec, ok := s.preAuths[cookieHashHex]; ok {
		return &rec, nil
	}
	return nil, authsession.ErrSessionNotFound
}

func (s *gidMemStore) DeletePreAuth(_ context.Context, cookieHashHex string) error {
	delete(s.preAuths, cookieHashHex)
	return nil
}

func (s *gidMemStore) CheckRateLimit(_ context.Context, scope, subject string, limit int64, _ time.Duration) (int64, error) {
	k := scope + "|" + subject
	s.rates[k]++
	if s.rates[k] > limit {
		return s.rates[k], authsession.ErrRateLimitExceeded
	}
	return s.rates[k], nil
}

type gidDummyRefresher struct{}

func (gidDummyRefresher) RefreshUserAccessSnapshot(context.Context, string, authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	return nil, nil
}

type gidMockAuthClient struct {
	iamv1.AuthServiceClient
}

type gidMockUserClient struct {
	iamv1.UserServiceClient
	now            time.Time
	lastUserID     string
	lastIdentityID string
	callCount      int
}

func (m *gidMockUserClient) resetRecorded() {
	m.lastUserID = ""
	m.lastIdentityID = ""
	m.callCount = 0
}

func (m *gidMockUserClient) sampleUser(id string, st iamv1.UserStatus) *iamv1.User {
	return &iamv1.User{
		Id:          id,
		Kind:        iamv1.UserKind_USER_KIND_EMPLOYEE,
		Status:      st,
		DisplayName: "GID Test User",
		CreatedAt:   timestamppb.New(m.now),
		UpdatedAt:   timestamppb.New(m.now),
	}
}

func (m *gidMockUserClient) GetUserById(_ context.Context, in *iamv1.GetUserByIdRequest, _ ...grpc.CallOption) (*iamv1.GetUserByIdResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.GetUserByIdResponse{User: m.sampleUser(in.GetUserId(), iamv1.UserStatus_USER_STATUS_ACTIVE)}, nil
}

func (m *gidMockUserClient) ListUserIdentities(_ context.Context, in *iamv1.ListUserIdentitiesRequest, _ ...grpc.CallOption) (*iamv1.ListUserIdentitiesResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ListUserIdentitiesResponse{
		Identities: []*iamv1.UserIdentity{
			{
				Id:         "99999999-9999-4999-8999-999999999999",
				UserId:     in.GetUserId(),
				Kind:       iamv1.IdentityKind_IDENTITY_KIND_EMAIL,
				Value:      "user@integrione.test",
				IsVerified: true,
				IsPrimary:  true,
				CreatedAt:  timestamppb.New(m.now),
				UpdatedAt:  timestamppb.New(m.now),
			},
		},
	}, nil
}

func (m *gidMockUserClient) ActivateUser(_ context.Context, in *iamv1.ActivateUserRequest, _ ...grpc.CallOption) (*iamv1.ActivateUserResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ActivateUserResponse{User: m.sampleUser(in.GetUserId(), iamv1.UserStatus_USER_STATUS_ACTIVE)}, nil
}

func (m *gidMockUserClient) SuspendUser(_ context.Context, in *iamv1.SuspendUserRequest, _ ...grpc.CallOption) (*iamv1.SuspendUserResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.SuspendUserResponse{User: m.sampleUser(in.GetUserId(), iamv1.UserStatus_USER_STATUS_SUSPENDED)}, nil
}

func (m *gidMockUserClient) DisableUser(_ context.Context, in *iamv1.DisableUserRequest, _ ...grpc.CallOption) (*iamv1.DisableUserResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.DisableUserResponse{User: m.sampleUser(in.GetUserId(), iamv1.UserStatus_USER_STATUS_DISABLED)}, nil
}

func (m *gidMockUserClient) SoftDeleteUser(_ context.Context, in *iamv1.SoftDeleteUserRequest, _ ...grpc.CallOption) (*iamv1.SoftDeleteUserResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.SoftDeleteUserResponse{}, nil
}

func (m *gidMockUserClient) SetPrimaryUserIdentity(_ context.Context, in *iamv1.SetPrimaryUserIdentityRequest, _ ...grpc.CallOption) (*iamv1.SetPrimaryUserIdentityResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	m.lastIdentityID = in.GetIdentityId()
	return &iamv1.SetPrimaryUserIdentityResponse{
		Identity: &iamv1.UserIdentity{
			Id:         in.GetIdentityId(),
			UserId:     in.GetUserId(),
			Kind:       iamv1.IdentityKind_IDENTITY_KIND_EMAIL,
			Value:      "user@integrione.test",
			IsVerified: true,
			IsPrimary:  true,
			CreatedAt:  timestamppb.New(m.now),
			UpdatedAt:  timestamppb.New(m.now),
		},
	}, nil
}

func (m *gidMockUserClient) RemoveUserIdentity(_ context.Context, in *iamv1.RemoveUserIdentityRequest, _ ...grpc.CallOption) (*iamv1.RemoveUserIdentityResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	m.lastIdentityID = in.GetIdentityId()
	return &iamv1.RemoveUserIdentityResponse{}, nil
}

type gidMockRoleClient struct {
	iamv1.RoleServiceClient
	now             time.Time
	lastUserID      string
	lastRoleID      string
	lastPermID      string
	callCount       int
	rolePermissions map[string][]*iamv1.Permission
}

func (m *gidMockRoleClient) resetRecorded() {
	m.lastUserID = ""
	m.lastRoleID = ""
	m.lastPermID = ""
	m.callCount = 0
}

func (m *gidMockRoleClient) sampleRole(id string) *iamv1.Role {
	return &iamv1.Role{
		Id:                  id,
		Code:                "finance.viewer",
		Name:                "Finance Viewer",
		ManagementDomain:    "finance",
		DefaultRedirectPath: "/finance",
		IsActive:            true,
		CreatedAt:           timestamppb.New(m.now),
		UpdatedAt:           timestamppb.New(m.now),
	}
}

func (m *gidMockRoleClient) GetRoleById(_ context.Context, in *iamv1.GetRoleByIdRequest, _ ...grpc.CallOption) (*iamv1.GetRoleByIdResponse, error) {
	m.callCount++
	m.lastRoleID = in.GetId()
	return &iamv1.GetRoleByIdResponse{Role: m.sampleRole(in.GetId())}, nil
}

func (m *gidMockRoleClient) ListRoles(context.Context, *iamv1.ListRolesRequest, ...grpc.CallOption) (*iamv1.ListRolesResponse, error) {
	return &iamv1.ListRolesResponse{}, nil
}

func (m *gidMockRoleClient) ListRolePermissions(_ context.Context, in *iamv1.ListRolePermissionsRequest, _ ...grpc.CallOption) (*iamv1.ListRolePermissionsResponse, error) {
	m.callCount++
	m.lastRoleID = in.GetRoleId()
	return &iamv1.ListRolePermissionsResponse{Permissions: m.rolePermissions[in.GetRoleId()]}, nil
}

func (m *gidMockRoleClient) AssignRolePermission(_ context.Context, in *iamv1.AssignRolePermissionRequest, _ ...grpc.CallOption) (*iamv1.AssignRolePermissionResponse, error) {
	m.callCount++
	m.lastRoleID = in.GetRoleId()
	m.lastPermID = in.GetPermissionId()
	return &iamv1.AssignRolePermissionResponse{
		RolePermission: &iamv1.RolePermission{
			RoleId:       in.GetRoleId(),
			PermissionId: in.GetPermissionId(),
			CreatedAt:    timestamppb.New(m.now),
		},
	}, nil
}

func (m *gidMockRoleClient) RevokeRolePermission(_ context.Context, in *iamv1.RevokeRolePermissionRequest, _ ...grpc.CallOption) (*iamv1.RevokeRolePermissionResponse, error) {
	m.callCount++
	m.lastRoleID = in.GetRoleId()
	m.lastPermID = in.GetPermissionId()
	return &iamv1.RevokeRolePermissionResponse{}, nil
}

func (m *gidMockRoleClient) ListUserRoles(_ context.Context, in *iamv1.ListUserRolesRequest, _ ...grpc.CallOption) (*iamv1.ListUserRolesResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ListUserRolesResponse{}, nil
}

func (m *gidMockRoleClient) ListEffectiveUserRoles(_ context.Context, in *iamv1.ListEffectiveUserRolesRequest, _ ...grpc.CallOption) (*iamv1.ListEffectiveUserRolesResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ListEffectiveUserRolesResponse{Roles: []*iamv1.Role{m.sampleRole("77777777-7777-4777-8777-777777777777")}}, nil
}

func (m *gidMockRoleClient) AssignUserRole(_ context.Context, in *iamv1.AssignUserRoleRequest, _ ...grpc.CallOption) (*iamv1.AssignUserRoleResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	m.lastRoleID = in.GetRoleId()
	return &iamv1.AssignUserRoleResponse{
		UserRole: &iamv1.UserRole{
			Id:        "88888888-8888-4888-8888-888888888888",
			UserId:    in.GetUserId(),
			RoleId:    in.GetRoleId(),
			Scope:     &iamv1.Scope{Type: iamv1.ScopeType_SCOPE_TYPE_GLOBAL},
			CreatedAt: timestamppb.New(m.now),
		},
	}, nil
}

type gidMockGroupClient struct {
	iamv1.GroupServiceClient
	now         time.Time
	lastUserID  string
	lastGroupID string
	callCount   int
}

func (m *gidMockGroupClient) resetRecorded() {
	m.lastUserID = ""
	m.lastGroupID = ""
	m.callCount = 0
}

func (m *gidMockGroupClient) ListUserGroups(_ context.Context, in *iamv1.ListUserGroupsRequest, _ ...grpc.CallOption) (*iamv1.ListUserGroupsResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ListUserGroupsResponse{
		Groups: []*iamv1.Group{
			{
				Id:        "66666666-6666-4666-8666-666666666666",
				Code:      "ops_team",
				Name:      "Ops Team",
				IsActive:  true,
				CreatedAt: timestamppb.New(m.now),
				UpdatedAt: timestamppb.New(m.now),
			},
		},
	}, nil
}

func (m *gidMockGroupClient) ListGroupRoles(context.Context, *iamv1.ListGroupRolesRequest, ...grpc.CallOption) (*iamv1.ListGroupRolesResponse, error) {
	return &iamv1.ListGroupRolesResponse{}, nil
}

func (m *gidMockGroupClient) AddGroupMember(_ context.Context, in *iamv1.AddGroupMemberRequest, _ ...grpc.CallOption) (*iamv1.AddGroupMemberResponse, error) {
	m.callCount++
	m.lastGroupID = in.GetGroupId()
	m.lastUserID = in.GetUserId()
	return &iamv1.AddGroupMemberResponse{
		Member: &iamv1.GroupMember{
			GroupId:   in.GetGroupId(),
			UserId:    in.GetUserId(),
			CreatedAt: timestamppb.New(m.now),
		},
	}, nil
}

func (m *gidMockGroupClient) RemoveGroupMember(_ context.Context, in *iamv1.RemoveGroupMemberRequest, _ ...grpc.CallOption) (*iamv1.RemoveGroupMemberResponse, error) {
	m.callCount++
	m.lastGroupID = in.GetGroupId()
	m.lastUserID = in.GetUserId()
	return &iamv1.RemoveGroupMemberResponse{}, nil
}

type gidMockSessionClient struct {
	iamv1.SessionServiceClient
	now        time.Time
	lastUserID string
	callCount  int
}

func (m *gidMockSessionClient) resetRecorded() {
	m.lastUserID = ""
	m.callCount = 0
}

func (m *gidMockSessionClient) ListUserSessions(_ context.Context, in *iamv1.ListUserSessionsRequest, _ ...grpc.CallOption) (*iamv1.ListUserSessionsResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.ListUserSessionsResponse{
		Items: []*iamv1.Session{
			{
				Id:                "55555555-5555-4555-8555-555555555555",
				UserId:            in.GetUserId(),
				AuthenticatedAt:   timestamppb.New(m.now),
				LastActiveAt:      timestamppb.New(m.now),
				IdleExpiresAt:     timestamppb.New(m.now.Add(time.Hour)),
				AbsoluteExpiresAt: timestamppb.New(m.now.Add(8 * time.Hour)),
				CreatedAt:         timestamppb.New(m.now),
				Generation:        1,
			},
		},
		Page: &iamv1.PageInfo{Total: 1, Limit: 20},
	}, nil
}

func (m *gidMockSessionClient) RevokeAllUserSessions(_ context.Context, in *iamv1.RevokeAllUserSessionsRequest, _ ...grpc.CallOption) (*iamv1.RevokeAllUserSessionsResponse, error) {
	m.callCount++
	m.lastUserID = in.GetUserId()
	return &iamv1.RevokeAllUserSessionsResponse{RevokedCount: 2}, nil
}

type gidMockPermClient struct {
	iamv1.PermissionServiceClient
	now        time.Time
	lastPermID string
	callCount  int
}

func (m *gidMockPermClient) resetRecorded() {
	m.lastPermID = ""
	m.callCount = 0
}

func (m *gidMockPermClient) samplePermission(id string) *iamv1.Permission {
	return &iamv1.Permission{
		Id:          id,
		Code:        "iam.users.read",
		Resource:    "users",
		Action:      "read",
		Description: "Read IAM user principals and identities",
		IsSensitive: false,
		CreatedAt:   timestamppb.New(m.now),
	}
}

func (m *gidMockPermClient) GetPermissionById(_ context.Context, in *iamv1.GetPermissionByIdRequest, _ ...grpc.CallOption) (*iamv1.GetPermissionByIdResponse, error) {
	m.callCount++
	m.lastPermID = in.GetId()
	return &iamv1.GetPermissionByIdResponse{Permission: m.samplePermission(in.GetId())}, nil
}

type gidTestHarness struct {
	adapter  *iamadapter.Adapter
	users    *gidMockUserClient
	roles    *gidMockRoleClient
	perms    *gidMockPermClient
	groups   *gidMockGroupClient
	sessions *gidMockSessionClient
	selfCtx  context.Context
	adminCtx context.Context
}

func newGIDTestHarness(t *testing.T, selfUserUUID, adminUserUUID string) *gidTestHarness {
	t.Helper()
	now := time.Now()
	store := newGIDMemStore()

	addSess := func(rawTok, sessID, userID string, roles []authz.RoleGrant, perms []authz.PermissionGrant) context.Context {
		tokHash := authsession.HashOpaqueToken(rawTok)
		store.sessions[tokHash] = &authsession.SessionValidationContext{
			Token: &authsession.SessionTokenRecord{
				SessionID:    sessID,
				UserID:       userID,
				TokenHashHex: tokHash,
				UserKind:     "employee",
				AuthEpoch:    1,
				IssuedAtUnix: now.Unix(),
			},
			Security: &authsession.UserSecurityRecord{
				UserID:        userID,
				Status:        authz.StatusActive,
				AuthEpoch:     1,
				PermRevision:  1,
				RecoveryEpoch: 1,
			},
			Snapshot: &authz.AccessSnapshot{
				PrincipalID:    userID,
				UserKind:       "employee",
				Status:         authz.StatusActive,
				AuthEpoch:      1,
				PermRevision:   1,
				RecoveryEpoch:  1,
				SyncedAtUnix:   now.Unix(),
				FreshUntilUnix: now.Add(30 * time.Second).Unix(),
				Roles:          roles,
				Permissions:    perms,
			},
		}
		return middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
			RawCookie:  rawTok,
			CookieHash: tokHash,
		})
	}

	selfCtx := addSess(
		"self-tok",
		"sess-self",
		selfUserUUID,
		[]authz.RoleGrant{{Code: "customer"}},
		[]authz.PermissionGrant{
			{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: selfUserUUID}},
			{Code: authz.PermSessionsReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: selfUserUUID}},
		},
	)
	adminCtx := addSess(
		"admin-tok",
		"sess-admin",
		adminUserUUID,
		[]authz.RoleGrant{{Code: "superadmin"}},
		[]authz.PermissionGrant{
			{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.roles.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.roles.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.groups.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.groups.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.sessions.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			{Code: "iam.sessions.revoke", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		},
	)

	usersMock := &gidMockUserClient{now: now}
	rolesMock := &gidMockRoleClient{now: now, rolePermissions: make(map[string][]*iamv1.Permission)}
	permsMock := &gidMockPermClient{now: now}
	groupsMock := &gidMockGroupClient{now: now}
	sessionsMock := &gidMockSessionClient{now: now}

	authMgr, err := middleware.NewAuthManager(
		store,
		middleware.NewSnapshotLoader(gidDummyRefresher{}, time.Second, 30*time.Second),
		middleware.SessionConfig{CookieSecure: true},
	)
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := iamadapter.NewAdapter(source_connections.IAMClientSet{
		Auth:       &gidMockAuthClient{},
		Users:      usersMock,
		Roles:      rolesMock,
		Permission: permsMock,
		Groups:     groupsMock,
		Sessions:   sessionsMock,
	}, authMgr, authzSvc, rv, iamadapter.Config{LoginRateMax: 5, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	return &gidTestHarness{
		adapter:  adp,
		users:    usersMock,
		roles:    rolesMock,
		perms:    permsMock,
		groups:   groupsMock,
		sessions: sessionsMock,
		selfCtx:  selfCtx,
		adminCtx: adminCtx,
	}
}

func (h *gidTestHarness) resetCalls() {
	h.users.resetRecorded()
	h.roles.resetRecorded()
	h.perms.resetRecorded()
	h.groups.resetRecorded()
	h.sessions.resetRecorded()
}

// TestParseTypedID_UserGIDAndRawUUID memverifikasi ParseTypedID menormalkan GID dan UUID mentah menjadi UUID lowercase kanonik serta menolak GID resource yang salah.
func TestParseTypedID_UserGIDAndRawUUID(t *testing.T) {
	t.Parallel()

	const wantUUID = "11111111-1111-4111-8111-111111111111"
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "gid_lowercase", raw: "gid://integrione/user/" + wantUUID, want: wantUUID},
		{name: "gid_uppercase_padded", raw: "  gid://integrione/user/" + strings.ToUpper(wantUUID) + "  ", want: wantUUID},
		{name: "raw_uuid_lowercase", raw: wantUUID, want: wantUUID},
		{name: "raw_uuid_uppercase_padded", raw: " " + strings.ToUpper(wantUUID) + " ", want: wantUUID},
		{name: "wrong_resource_gid", raw: "gid://integrione/role/" + wantUUID, wantErr: true},
		{name: "wrong_prefix_gid", raw: "gid://other/user/" + wantUUID, wantErr: true},
		{name: "empty_string", raw: "   ", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := iamadapter.ParseTypedID(tc.raw, iamadapter.ResourceUser)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseTypedID(%q) expected error, got %q", tc.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTypedID(%q) unexpected error: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("ParseTypedID(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestNestedUserResolvers_AcceptGIDAndRawUUIDWithSelfOrAdmin memverifikasi resolver nested (*ForUser & UserForNested) menerima GID maupun UUID mentah untuk pemilik akun (self) dan admin.
func TestNestedUserResolvers_AcceptGIDAndRawUUIDWithSelfOrAdmin(t *testing.T) {
	t.Parallel()

	const (
		selfUUID  = "11111111-1111-4111-8111-111111111111"
		adminUUID = "22222222-2222-4222-8222-222222222222"
	)

	inputs := []struct {
		name string
		val  string
	}{
		{name: "gid_lowercase", val: "gid://integrione/user/" + selfUUID},
		{name: "gid_uppercase", val: "gid://integrione/user/" + strings.ToUpper(selfUUID)},
		{name: "raw_uuid_lowercase", val: selfUUID},
		{name: "raw_uuid_uppercase_padded", val: "  " + strings.ToUpper(selfUUID) + "  "},
	}

	type nestedOp struct {
		name      string
		call      func(ctx context.Context, h *gidTestHarness, userRef string) error
		getTarget func(h *gidTestHarness) (string, int)
	}

	ops := []nestedOp{
		{
			name: "UserForNested",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				res, err := h.adapter.UserForNested(ctx, userRef)
				if err == nil && (res == nil || res.RawID != selfUUID) {
					t.Fatalf("UserForNested(%q) unexpected result: %+v", userRef, res)
				}
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.users.lastUserID, h.users.callCount },
		},
		{
			name: "UserIdentitiesForUser",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				_, err := h.adapter.UserIdentitiesForUser(ctx, userRef, nil, nil, nil, nil)
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.users.lastUserID, h.users.callCount },
		},
		{
			name: "UserRolesForUser",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				_, err := h.adapter.UserRolesForUser(ctx, userRef, nil, nil, nil, nil, nil)
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.roles.lastUserID, h.roles.callCount },
		},
		{
			name: "UserEffectiveRolesForUser",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				_, err := h.adapter.UserEffectiveRolesForUser(ctx, userRef, nil, nil, nil, nil)
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.roles.lastUserID, h.roles.callCount },
		},
		{
			name: "UserGroupsForUser",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				_, err := h.adapter.UserGroupsForUser(ctx, userRef, nil, nil, nil, nil)
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.groups.lastUserID, h.groups.callCount },
		},
		{
			name: "UserSessionsForUser",
			call: func(ctx context.Context, h *gidTestHarness, userRef string) error {
				_, err := h.adapter.UserSessionsForUser(ctx, userRef, nil, nil, nil, nil, nil)
				return err
			},
			getTarget: func(h *gidTestHarness) (string, int) { return h.sessions.lastUserID, h.sessions.callCount },
		},
	}

	h := newGIDTestHarness(t, selfUUID, adminUUID)

	for _, op := range ops {
		for _, in := range inputs {
			t.Run("Self/"+op.name+"/"+in.name, func(t *testing.T) {
				h.resetCalls()
				if err := op.call(h.selfCtx, h, in.val); err != nil {
					t.Fatalf("%s(%q) self call failed: %v", op.name, in.val, err)
				}
				gotUUID, calls := op.getTarget(h)
				if calls != 1 || gotUUID != selfUUID {
					t.Fatalf("%s(%q) forwarded user_id=%q calls=%d, want %q (1 call)", op.name, in.val, gotUUID, calls, selfUUID)
				}
			})

			t.Run("Admin/"+op.name+"/"+in.name, func(t *testing.T) {
				h.resetCalls()
				if err := op.call(h.adminCtx, h, in.val); err != nil {
					t.Fatalf("%s(%q) admin call failed: %v", op.name, in.val, err)
				}
				gotUUID, calls := op.getTarget(h)
				if calls != 1 || gotUUID != selfUUID {
					t.Fatalf("%s(%q) admin forwarded user_id=%q calls=%d, want %q (1 call)", op.name, in.val, gotUUID, calls, selfUUID)
				}
			})
		}

		t.Run("RejectsWrongResourceGID/"+op.name, func(t *testing.T) {
			h.resetCalls()
			err := op.call(h.adminCtx, h, "gid://integrione/role/"+selfUUID)
			if err == nil {
				t.Fatalf("%s accepted wrong resource GID unexpectedly", op.name)
			}
			_, calls := op.getTarget(h)
			if calls != 0 {
				t.Fatalf("%s invoked gRPC (%d calls) despite invalid resource GID", op.name, calls)
			}
		})

		t.Run("RejectsOtherUserForNonAdmin/"+op.name, func(t *testing.T) {
			h.resetCalls()
			if err := op.call(h.selfCtx, h, adminUUID); err == nil {
				t.Fatalf("%s allowed non-admin access to other user raw UUID", op.name)
			}
			if err := op.call(h.selfCtx, h, "gid://integrione/user/"+adminUUID); err == nil {
				t.Fatalf("%s allowed non-admin access to other user GID", op.name)
			}
			_, calls := op.getTarget(h)
			if calls != 0 {
				t.Fatalf("%s invoked gRPC (%d calls) despite authorization failure", op.name, calls)
			}
		})
	}
}

// TestTopLevelUserIDOperations_AcceptGIDAndRawUUID memverifikasi seluruh query dan mutation top-level dengan argumen user_id menerima GID maupun UUID mentah serta meneruskan UUID kanonik lowercase ke gRPC.
func TestTopLevelUserIDOperations_AcceptGIDAndRawUUID(t *testing.T) {
	t.Parallel()

	const (
		adminUUID    = "22222222-2222-4222-8222-222222222222"
		targetUUID   = "33333333-3333-4333-8333-333333333333"
		identityUUID = "99999999-9999-4999-8999-999999999999"
		roleUUID     = "77777777-7777-4777-8777-777777777777"
		groupUUID    = "66666666-6666-4666-8666-666666666666"
	)

	formats := []struct {
		name string
		val  string
	}{
		{name: "gid", val: "gid://integrione/user/" + targetUUID},
		{name: "raw_uuid_lowercase", val: targetUUID},
		{name: "raw_uuid_uppercase", val: strings.ToUpper(targetUUID)},
	}

	h := newGIDTestHarness(t, targetUUID, adminUUID)

	for _, fmtCase := range formats {
		fmtCase := fmtCase
		t.Run("Queries/"+fmtCase.name, func(t *testing.T) {
			h.resetCalls()
			if u, err := h.adapter.User(h.adminCtx, fmtCase.val); err != nil || u == nil || u.RawID != targetUUID {
				t.Fatalf("User(%q) failed: u=%+v err=%v", fmtCase.val, u, err)
			}
			if h.users.lastUserID != targetUUID {
				t.Fatalf("User(%q) forwarded %q, want %q", fmtCase.val, h.users.lastUserID, targetUUID)
			}

			h.resetCalls()
			if _, err := h.adapter.UserIdentities(h.adminCtx, fmtCase.val, nil, nil, nil, nil); err != nil {
				t.Fatalf("UserIdentities(%q) failed: %v", fmtCase.val, err)
			}
			if h.users.lastUserID != targetUUID {
				t.Fatalf("UserIdentities(%q) forwarded %q, want %q", fmtCase.val, h.users.lastUserID, targetUUID)
			}

			h.resetCalls()
			if _, err := h.adapter.UserRoles(h.adminCtx, fmtCase.val, nil, nil, nil, nil, nil); err != nil {
				t.Fatalf("UserRoles(%q) failed: %v", fmtCase.val, err)
			}
			if h.roles.lastUserID != targetUUID {
				t.Fatalf("UserRoles(%q) forwarded %q, want %q", fmtCase.val, h.roles.lastUserID, targetUUID)
			}

			h.resetCalls()
			if _, err := h.adapter.UserEffectiveRoles(h.adminCtx, fmtCase.val, nil, nil, nil, nil); err != nil {
				t.Fatalf("UserEffectiveRoles(%q) failed: %v", fmtCase.val, err)
			}
			if h.roles.lastUserID != targetUUID {
				t.Fatalf("UserEffectiveRoles(%q) forwarded %q, want %q", fmtCase.val, h.roles.lastUserID, targetUUID)
			}

			h.resetCalls()
			if _, err := h.adapter.UserGroups(h.adminCtx, fmtCase.val, nil, nil, nil, nil); err != nil {
				t.Fatalf("UserGroups(%q) failed: %v", fmtCase.val, err)
			}
			if h.groups.lastUserID != targetUUID {
				t.Fatalf("UserGroups(%q) forwarded %q, want %q", fmtCase.val, h.groups.lastUserID, targetUUID)
			}

			h.resetCalls()
			if _, err := h.adapter.UserSessions(h.adminCtx, fmtCase.val, nil, nil, nil, nil, nil); err != nil {
				t.Fatalf("UserSessions(%q) failed: %v", fmtCase.val, err)
			}
			if h.sessions.lastUserID != targetUUID {
				t.Fatalf("UserSessions(%q) forwarded %q, want %q", fmtCase.val, h.sessions.lastUserID, targetUUID)
			}
		})

		t.Run("Mutations/"+fmtCase.name, func(t *testing.T) {
			h.resetCalls()
			if p, err := h.adapter.ActivateUser(h.adminCtx, fmtCase.val); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("ActivateUser(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.SuspendUser(h.adminCtx, fmtCase.val); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("SuspendUser(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.DisableUser(h.adminCtx, fmtCase.val); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("DisableUser(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.SoftDeleteUser(h.adminCtx, fmtCase.val); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("SoftDeleteUser(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.SetPrimaryUserIdentity(h.adminCtx, fmtCase.val, identityUUID); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("SetPrimaryUserIdentity(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.RemoveUserIdentity(h.adminCtx, fmtCase.val, identityUUID); err != nil || !p.WasSuccessful || h.users.lastUserID != targetUUID {
				t.Fatalf("RemoveUserIdentity(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.users.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.AssignUserRole(h.adminCtx, model.AssignUserRoleInput{
				UserID:    fmtCase.val,
				RoleID:    roleUUID,
				ScopeType: model.ScopeTypeGlobal,
			}); err != nil || !p.WasSuccessful || h.roles.lastUserID != targetUUID {
				t.Fatalf("AssignUserRole(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.roles.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.AddGroupMember(h.adminCtx, groupUUID, fmtCase.val); err != nil || !p.WasSuccessful || h.groups.lastUserID != targetUUID {
				t.Fatalf("AddGroupMember(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.groups.lastUserID)
			}

			h.resetCalls()
			if p, err := h.adapter.RemoveGroupMember(h.adminCtx, groupUUID, fmtCase.val); err != nil || !p.WasSuccessful || h.groups.lastUserID != targetUUID {
				t.Fatalf("RemoveGroupMember(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.groups.lastUserID)
			}

			reason := model.SessionEndReasonAdminRevoked
			h.resetCalls()
			if p, err := h.adapter.RevokeAllUserSessions(h.adminCtx, fmtCase.val, &reason); err != nil || !p.WasSuccessful || h.sessions.lastUserID != targetUUID {
				t.Fatalf("RevokeAllUserSessions(%q) failed: payload=%+v err=%v lastUser=%q", fmtCase.val, p, err, h.sessions.lastUserID)
			}
		})
	}
}
