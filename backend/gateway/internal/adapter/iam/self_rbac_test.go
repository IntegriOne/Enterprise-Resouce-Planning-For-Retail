package iam

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type selfMockAuthClient struct {
	iamv1.AuthServiceClient
	mu                 sync.Mutex
	calls              int
	changePwdFn        func(context.Context, *iamv1.ChangePasswordRequest) (*iamv1.ChangePasswordResponse, error)
	verifySessionMfaFn func(context.Context, *iamv1.VerifySessionMfaRequest) (*iamv1.VerifySessionMfaResponse, error)
	enrollTotpFn       func(context.Context, *iamv1.EnrollTotpRequest) (*iamv1.EnrollTotpResponse, error)
	confirmTotpFn      func(context.Context, *iamv1.ConfirmTotpRequest) (*iamv1.ConfirmTotpResponse, error)
	disableTotpFn      func(context.Context, *iamv1.DisableTotpRequest) (*iamv1.DisableTotpResponse, error)
	regenCodesFn       func(context.Context, *iamv1.RegenerateRecoveryCodesRequest) (*iamv1.RegenerateRecoveryCodesResponse, error)
}

func (m *selfMockAuthClient) record() {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
}

func (m *selfMockAuthClient) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func (m *selfMockAuthClient) ChangePassword(ctx context.Context, in *iamv1.ChangePasswordRequest, _ ...grpc.CallOption) (*iamv1.ChangePasswordResponse, error) {
	m.record()
	if m.changePwdFn != nil {
		return m.changePwdFn(ctx, in)
	}
	return &iamv1.ChangePasswordResponse{Changed: true, SessionsRevoked: 1, PasswordChangedAt: timestamppb.Now()}, nil
}

func (m *selfMockAuthClient) VerifySessionMfa(ctx context.Context, in *iamv1.VerifySessionMfaRequest, _ ...grpc.CallOption) (*iamv1.VerifySessionMfaResponse, error) {
	m.record()
	if m.verifySessionMfaFn != nil {
		return m.verifySessionMfaFn(ctx, in)
	}
	return &iamv1.VerifySessionMfaResponse{
		Session:   &iamv1.Session{Id: in.GetSessionId(), UserId: in.GetUserId(), MfaVerified: true},
		CsrfToken: "csrf-rotated-mfa",
	}, nil
}

func (m *selfMockAuthClient) EnrollTotp(ctx context.Context, in *iamv1.EnrollTotpRequest, _ ...grpc.CallOption) (*iamv1.EnrollTotpResponse, error) {
	m.record()
	if m.enrollTotpFn != nil {
		return m.enrollTotpFn(ctx, in)
	}
	return &iamv1.EnrollTotpResponse{FactorId: "factor-1", OtpAuthUri: "otpauth://totp/IntegriOne?secret=BASE32"}, nil
}

func (m *selfMockAuthClient) ConfirmTotp(ctx context.Context, in *iamv1.ConfirmTotpRequest, _ ...grpc.CallOption) (*iamv1.ConfirmTotpResponse, error) {
	m.record()
	if m.confirmTotpFn != nil {
		return m.confirmTotpFn(ctx, in)
	}
	return &iamv1.ConfirmTotpResponse{
		Factor:        &iamv1.UserTotpFactor{Id: "factor-1", UserId: in.GetUserId(), IsEnabled: true},
		RecoveryCodes: []string{"REC-1", "REC-2"},
	}, nil
}

func (m *selfMockAuthClient) DisableTotp(ctx context.Context, in *iamv1.DisableTotpRequest, _ ...grpc.CallOption) (*iamv1.DisableTotpResponse, error) {
	m.record()
	if m.disableTotpFn != nil {
		return m.disableTotpFn(ctx, in)
	}
	return &iamv1.DisableTotpResponse{}, nil
}

func (m *selfMockAuthClient) RegenerateRecoveryCodes(ctx context.Context, in *iamv1.RegenerateRecoveryCodesRequest, _ ...grpc.CallOption) (*iamv1.RegenerateRecoveryCodesResponse, error) {
	m.record()
	if m.regenCodesFn != nil {
		return m.regenCodesFn(ctx, in)
	}
	return &iamv1.RegenerateRecoveryCodesResponse{RecoveryCodes: []string{"NEW-1", "NEW-2"}}, nil
}

type selfMockUserClient struct {
	iamv1.UserServiceClient
	mu    sync.Mutex
	calls int
	now   time.Time
}

func (m *selfMockUserClient) record() {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
}

func (m *selfMockUserClient) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func (m *selfMockUserClient) GetUserById(_ context.Context, in *iamv1.GetUserByIdRequest, _ ...grpc.CallOption) (*iamv1.GetUserByIdResponse, error) {
	m.record()
	return &iamv1.GetUserByIdResponse{
		User: &iamv1.User{
			Id:          in.GetUserId(),
			Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
			Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
			DisplayName: "User " + in.GetUserId(),
			CreatedAt:   timestamppb.New(m.now),
			UpdatedAt:   timestamppb.New(m.now),
		},
	}, nil
}

func (m *selfMockUserClient) UpdateUserProfile(_ context.Context, in *iamv1.UpdateUserProfileRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserProfileResponse, error) {
	m.record()
	return &iamv1.UpdateUserProfileResponse{
		User: &iamv1.User{
			Id:          in.GetUserId(),
			Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
			Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
			DisplayName: in.GetDisplayName(),
			AvatarUrl:   in.GetAvatarUrl(),
			CreatedAt:   timestamppb.New(m.now),
			UpdatedAt:   timestamppb.New(m.now),
		},
	}, nil
}

func (m *selfMockUserClient) UpdateUserEmail(_ context.Context, in *iamv1.UpdateUserEmailRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserEmailResponse, error) {
	m.record()
	return &iamv1.UpdateUserEmailResponse{PendingEmail: in.GetEmail(), IsPending: true, ExpiresAt: timestamppb.New(m.now.Add(24 * time.Hour))}, nil
}

func (m *selfMockUserClient) UpdateUserUsername(_ context.Context, in *iamv1.UpdateUserUsernameRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserUsernameResponse, error) {
	m.record()
	return &iamv1.UpdateUserUsernameResponse{Username: in.GetUsername()}, nil
}

func (m *selfMockUserClient) DeletePendingUserEmail(_ context.Context, _ *iamv1.DeletePendingUserEmailRequest, _ ...grpc.CallOption) (*iamv1.DeletePendingUserEmailResponse, error) {
	m.record()
	return &iamv1.DeletePendingUserEmailResponse{Deleted: true}, nil
}

func (m *selfMockUserClient) ResendValidationEmailToken(_ context.Context, _ *iamv1.ResendValidationEmailTokenRequest, _ ...grpc.CallOption) (*iamv1.ResendValidationEmailTokenResponse, error) {
	m.record()
	return &iamv1.ResendValidationEmailTokenResponse{IsResend: true}, nil
}

func (m *selfMockUserClient) ListUserIdentities(_ context.Context, in *iamv1.ListUserIdentitiesRequest, _ ...grpc.CallOption) (*iamv1.ListUserIdentitiesResponse, error) {
	m.record()
	identID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	if in.GetUserId() == "22222222-2222-4222-8222-222222222222" {
		identID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	}
	return &iamv1.ListUserIdentitiesResponse{
		Identities: []*iamv1.UserIdentity{
			{
				Id:         identID,
				UserId:     in.GetUserId(),
				Kind:       iamv1.IdentityKind_IDENTITY_KIND_EMAIL,
				Value:      in.GetUserId() + "@integrione.id",
				IsVerified: true,
				IsPrimary:  true,
				CreatedAt:  timestamppb.New(m.now),
				UpdatedAt:  timestamppb.New(m.now),
			},
		},
	}, nil
}

type selfMockSessionClient struct {
	iamv1.SessionServiceClient
	mu             sync.Mutex
	calls          int
	revokeCalls    int
	sessionsByUser map[string][]*iamv1.Session
}

func (m *selfMockSessionClient) ListUserSessions(_ context.Context, in *iamv1.ListUserSessionsRequest, _ ...grpc.CallOption) (*iamv1.ListUserSessionsResponse, error) {
	m.mu.Lock()
	m.calls++
	all := m.sessionsByUser[in.GetUserId()]
	m.mu.Unlock()

	limit := int(in.GetPage().GetLimit())
	offset := int(in.GetPage().GetOffset())
	if limit <= 0 {
		limit = len(all)
	}
	if offset >= len(all) {
		return &iamv1.ListUserSessionsResponse{
			Items: nil,
			Page:  &iamv1.PageInfo{Total: int64(len(all)), Limit: int32(limit), Offset: int32(offset), HasMore: false},
		}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return &iamv1.ListUserSessionsResponse{
		Items: all[offset:end],
		Page: &iamv1.PageInfo{
			Total:   int64(len(all)),
			Limit:   int32(limit),
			Offset:  int32(offset),
			HasMore: end < len(all),
		},
	}, nil
}

func (m *selfMockSessionClient) RevokeSessionById(_ context.Context, _ *iamv1.RevokeSessionByIdRequest, _ ...grpc.CallOption) (*iamv1.RevokeSessionByIdResponse, error) {
	m.mu.Lock()
	m.calls++
	m.revokeCalls++
	m.mu.Unlock()
	return &iamv1.RevokeSessionByIdResponse{}, nil
}

type selfMockRoleClient struct {
	iamv1.RoleServiceClient
	mu              sync.Mutex
	calls           int
	roles           map[string]*iamv1.Role
	rolePermissions map[string][]*iamv1.Permission
	now             time.Time
}

func (m *selfMockRoleClient) record() {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
}

func (m *selfMockRoleClient) GetRoleById(_ context.Context, in *iamv1.GetRoleByIdRequest, _ ...grpc.CallOption) (*iamv1.GetRoleByIdResponse, error) {
	m.record()
	if r, ok := m.roles[in.GetId()]; ok {
		return &iamv1.GetRoleByIdResponse{Role: r}, nil
	}
	return &iamv1.GetRoleByIdResponse{Role: &iamv1.Role{Id: in.GetId(), Code: "custom.role", Name: "Custom Role", IsActive: true, CreatedAt: timestamppb.New(m.now), UpdatedAt: timestamppb.New(m.now)}}, nil
}

func (m *selfMockRoleClient) ListRolePermissions(_ context.Context, in *iamv1.ListRolePermissionsRequest, _ ...grpc.CallOption) (*iamv1.ListRolePermissionsResponse, error) {
	m.record()
	return &iamv1.ListRolePermissionsResponse{Permissions: m.rolePermissions[in.GetRoleId()]}, nil
}

func (m *selfMockRoleClient) ListUserRoles(_ context.Context, _ *iamv1.ListUserRolesRequest, _ ...grpc.CallOption) (*iamv1.ListUserRolesResponse, error) {
	m.record()
	return &iamv1.ListUserRolesResponse{}, nil
}

func (m *selfMockRoleClient) ListEffectiveUserRoles(_ context.Context, _ *iamv1.ListEffectiveUserRolesRequest, _ ...grpc.CallOption) (*iamv1.ListEffectiveUserRolesResponse, error) {
	m.record()
	return &iamv1.ListEffectiveUserRolesResponse{}, nil
}

func (m *selfMockRoleClient) AssignUserRole(_ context.Context, in *iamv1.AssignUserRoleRequest, _ ...grpc.CallOption) (*iamv1.AssignUserRoleResponse, error) {
	m.record()
	return &iamv1.AssignUserRoleResponse{
		UserRole: &iamv1.UserRole{
			Id:        "ur-1",
			UserId:    in.GetUserId(),
			RoleId:    in.GetRoleId(),
			Scope:     &iamv1.Scope{Type: in.GetScopeType(), Id: in.ScopeId},
			CreatedAt: timestamppb.New(m.now),
		},
	}, nil
}

func (m *selfMockRoleClient) AssignRolePermission(_ context.Context, in *iamv1.AssignRolePermissionRequest, _ ...grpc.CallOption) (*iamv1.AssignRolePermissionResponse, error) {
	m.record()
	return &iamv1.AssignRolePermissionResponse{
		RolePermission: &iamv1.RolePermission{
			RoleId:       in.GetRoleId(),
			PermissionId: in.GetPermissionId(),
			CreatedAt:    timestamppb.New(m.now),
		},
	}, nil
}

type selfMockGroupClient struct {
	iamv1.GroupServiceClient
	mu    sync.Mutex
	calls int
	now   time.Time
}

func (m *selfMockGroupClient) record() {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
}

func (m *selfMockGroupClient) ListUserGroups(_ context.Context, _ *iamv1.ListUserGroupsRequest, _ ...grpc.CallOption) (*iamv1.ListUserGroupsResponse, error) {
	m.record()
	return &iamv1.ListUserGroupsResponse{}, nil
}

func (m *selfMockGroupClient) ListGroupRoles(_ context.Context, _ *iamv1.ListGroupRolesRequest, _ ...grpc.CallOption) (*iamv1.ListGroupRolesResponse, error) {
	m.record()
	return &iamv1.ListGroupRolesResponse{}, nil
}

func (m *selfMockGroupClient) AssignGroupRole(_ context.Context, in *iamv1.AssignGroupRoleRequest, _ ...grpc.CallOption) (*iamv1.AssignGroupRoleResponse, error) {
	m.record()
	return &iamv1.AssignGroupRoleResponse{
		GroupRole: &iamv1.GroupRole{
			Id:        "gr-1",
			GroupId:   in.GetGroupId(),
			RoleId:    in.GetRoleId(),
			Scope:     &iamv1.Scope{Type: in.GetScopeType(), Id: in.ScopeId},
			CreatedAt: timestamppb.New(m.now),
		},
	}, nil
}

type selfMockPermClient struct {
	iamv1.PermissionServiceClient
	perms map[string]*iamv1.Permission
}

func (m *selfMockPermClient) GetPermissionById(_ context.Context, in *iamv1.GetPermissionByIdRequest, _ ...grpc.CallOption) (*iamv1.GetPermissionByIdResponse, error) {
	return &iamv1.GetPermissionByIdResponse{Permission: m.perms[in.GetId()]}, nil
}

type selfRBACHarness struct {
	now      time.Time
	store    *memSessionStore
	auth     *selfMockAuthClient
	users    *selfMockUserClient
	sessions *selfMockSessionClient
	roles    *selfMockRoleClient
	groups   *selfMockGroupClient
	perms    *selfMockPermClient
	adapter  *Adapter
}

func newSelfRBACHarness(t *testing.T) *selfRBACHarness {
	t.Helper()
	now := time.Now()
	store := newMemStore()
	authMock := &selfMockAuthClient{}
	userMock := &selfMockUserClient{now: now}
	sessMock := &selfMockSessionClient{sessionsByUser: make(map[string][]*iamv1.Session)}
	roleMock := &selfMockRoleClient{
		roles:           make(map[string]*iamv1.Role),
		rolePermissions: make(map[string][]*iamv1.Permission),
		now:             now,
	}
	groupMock := &selfMockGroupClient{now: now}
	permMock := &selfMockPermClient{perms: make(map[string]*iamv1.Permission)}

	authMgr, err := middleware.NewAuthManager(
		store,
		middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second),
		middleware.SessionConfig{CookieSecure: true},
	)
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(source_connections.IAMClientSet{
		Auth:       authMock,
		Users:      userMock,
		Sessions:   sessMock,
		Roles:      roleMock,
		Groups:     groupMock,
		Permission: permMock,
	}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	return &selfRBACHarness{
		now:      now,
		store:    store,
		auth:     authMock,
		users:    userMock,
		sessions: sessMock,
		roles:    roleMock,
		groups:   groupMock,
		perms:    permMock,
		adapter:  adp,
	}
}

func (h *selfRBACHarness) registerSession(
	rawToken, sessionID, userID, userKind string,
	issuedAt time.Time,
	mfaEnabled, mfaAuth bool,
	roles []authz.RoleGrant,
	perms []authz.PermissionGrant,
) context.Context {
	tokHash := authsession.HashOpaqueToken(rawToken)
	h.store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:        sessionID,
			UserID:           userID,
			TokenHashHex:     tokHash,
			UserKind:         userKind,
			AuthEpoch:        1,
			IssuedAtUnix:     issuedAt.Unix(),
			MFAAuthenticated: mfaAuth,
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
			UserKind:       userKind,
			Status:         authz.StatusActive,
			MFAEnabled:     mfaEnabled,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			Roles:          roles,
			Permissions:    perms,
			SyncedAtUnix:   h.now.Unix(),
			FreshUntilUnix: h.now.Add(30 * time.Second).Unix(),
		},
	}
	return middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawToken,
		CookieHash: tokHash,
	})
}

func fullSelfPerms(userID string) []authz.PermissionGrant {
	scope := authz.ScopeGrant{Type: authz.ScopeSelf, ID: userID}
	return []authz.PermissionGrant{
		{Code: authz.PermUsersReadSelf, Scope: scope},
		{Code: authz.PermUsersUpdateSelf, Scope: scope},
		{Code: authz.PermUsersDeleteSelf, Scope: scope},
		{Code: authz.PermSessionsReadSelf, Scope: scope},
		{Code: authz.PermSessionsDeleteSelf, IsSensitive: true, Scope: scope},
	}
}

func TestSelfServiceOperations_AuthorizationAndSensitiveReauth(t *testing.T) {
	t.Parallel()

	const selfID = "11111111-1111-4111-8111-111111111111"
	h := newSelfRBACHarness(t)
	h.sessions.sessionsByUser[selfID] = []*iamv1.Session{{Id: "sess-own", UserId: selfID, CreatedAt: timestamppb.New(h.now)}}

	noSelfCtx := h.registerSession("tok-no-self", "sess-no-self", selfID, "customer", h.now, false, false, nil, nil)
	fullFreshCtx := h.registerSession("tok-full-fresh", "sess-own", selfID, "customer", h.now, false, false, nil, fullSelfPerms(selfID))
	staleUpdateCtx := h.registerSession("tok-stale-upd", "sess-stale", selfID, "customer", h.now.Add(-20*time.Minute), false, false, nil, fullSelfPerms(selfID))
	unverifiedMfaCtx := h.registerSession("tok-mfa-unverified", "sess-unmfa", selfID, "customer", h.now, true, false, nil, fullSelfPerms(selfID))

	avifPayload := "data:image/avif;base64," + base64.StdEncoding.EncodeToString([]byte{0, 0, 0, 16, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f', 0, 0, 0, 0})

	// 1. Operasi query mandiri (*self): Me & MySessions.
	if u, err := h.adapter.Me(fullFreshCtx); err != nil || u == nil || u.RawID != selfID {
		t.Fatalf("Me with %s failed: user=%+v err=%v", authz.PermUsersReadSelf, u, err)
	}
	if _, err := h.adapter.Me(noSelfCtx); err == nil {
		t.Fatal("Me must reject caller lacking iam.users.readself")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected FORBIDDEN on Me without self perm, got %v", err)
	}

	if conn, err := h.adapter.MySessions(fullFreshCtx, nil, nil, nil, nil, nil); err != nil || conn == nil || len(conn.Edges) != 1 {
		t.Fatalf("MySessions with %s failed: conn=%+v err=%v", authz.PermSessionsReadSelf, conn, err)
	}
	if _, err := h.adapter.MySessions(noSelfCtx, nil, nil, nil, nil, nil); err == nil {
		t.Fatal("MySessions must reject caller lacking iam.sessions.readself")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected FORBIDDEN on MySessions without self perm, got %v", err)
	}

	// 2. Mutasi *self non-sensitif vs mutasi *self wajib reauth sensitif.
	assertMutDenied := func(name string, errs []*model.MutationError, wasSuccessful bool) {
		t.Helper()
		if wasSuccessful || len(errs) == 0 || errs[0].Type != model.ErrorTypeEnumAuthorization {
			t.Fatalf("%s expected AUTHORIZATION failure, got wasSuccessful=%v errs=%+v", name, wasSuccessful, errs)
		}
	}

	// Non-sensitif: harus gagal tanpa *self, dan tetap sukses pada sesi lama jika memegang *self.
	pProfDeny, _ := h.adapter.UpdateMyProfile(noSelfCtx, model.UpdateMyProfileInput{DisplayName: "Deni"})
	assertMutDenied("UpdateMyProfile(noSelf)", pProfDeny.Errors, pProfDeny.WasSuccessful)
	if pProfOK, err := h.adapter.UpdateMyProfile(staleUpdateCtx, model.UpdateMyProfileInput{DisplayName: "Deni"}); err != nil || !pProfOK.WasSuccessful {
		t.Fatalf("UpdateMyProfile(staleUpdateCtx) failed: payload=%+v err=%v", pProfOK, err)
	}

	pAvaDeny, _ := h.adapter.UploadAvatarUser(noSelfCtx, model.UserAvatarUploadInput{Image: avifPayload})
	assertMutDenied("UploadAvatarUser(noSelf)", pAvaDeny.Errors, pAvaDeny.WasSuccessful)
	if pAvaOK, err := h.adapter.UploadAvatarUser(staleUpdateCtx, model.UserAvatarUploadInput{Image: avifPayload}); err != nil || !pAvaOK.WasSuccessful {
		t.Fatalf("UploadAvatarUser(staleUpdateCtx) failed: payload=%+v err=%v", pAvaOK, err)
	}

	pUsrDeny, _ := h.adapter.ChangeUsername(noSelfCtx, model.UsernameInput{Username: "kibo_baru"})
	assertMutDenied("ChangeUsername(noSelf)", pUsrDeny.Errors, pUsrDeny.WasSuccessful)
	if pUsrOK, err := h.adapter.ChangeUsername(staleUpdateCtx, model.UsernameInput{Username: "kibo_baru"}); err != nil || !pUsrOK.WasSuccessful {
		t.Fatalf("ChangeUsername(staleUpdateCtx) failed: payload=%+v err=%v", pUsrOK, err)
	}

	pMfaStepDeny, _ := h.adapter.VerifyMySessionMfa(noSelfCtx, "123456")
	assertMutDenied("VerifyMySessionMfa(noSelf)", pMfaStepDeny.Errors, pMfaStepDeny.WasSuccessful)
	if pMfaStepOK, err := h.adapter.VerifyMySessionMfa(unverifiedMfaCtx, "123456"); err != nil || !pMfaStepOK.WasSuccessful {
		t.Fatalf("VerifyMySessionMfa(unverifiedMfaCtx) failed: payload=%+v err=%v", pMfaStepOK, err)
	}

	pResendDeny, _ := h.adapter.ResendValidationEmailToken(noSelfCtx, nil)
	assertMutDenied("ResendValidationEmailToken(noSelf)", pResendDeny.Errors, pResendDeny.WasSuccessful)
	if pResendOK, err := h.adapter.ResendValidationEmailToken(staleUpdateCtx, nil); err != nil || !pResendOK.WasSuccessful {
		t.Fatalf("ResendValidationEmailToken(staleUpdateCtx) failed: payload=%+v err=%v", pResendOK, err)
	}

	pDelEmailDeny, _ := h.adapter.DeletePendingUserEmail(noSelfCtx)
	assertMutDenied("DeletePendingUserEmail(noSelf)", pDelEmailDeny.Errors, pDelEmailDeny.WasSuccessful)
	if pDelEmailOK, err := h.adapter.DeletePendingUserEmail(staleUpdateCtx); err != nil || !pDelEmailOK.WasSuccessful {
		t.Fatalf("DeletePendingUserEmail(staleUpdateCtx) failed: payload=%+v err=%v", pDelEmailOK, err)
	}

	// Mutasi *self sensitif: harus gagal bila tanpa *self, sesi kedaluwarsa >10m, maupun MFA belum step-up.
	type sensitiveCheck struct {
		name string
		call func(context.Context) ([]*model.MutationError, bool, error)
	}
	sensitiveCases := []sensitiveCheck{
		{
			name: "UpdateUserEmail",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.UpdateUserEmail(ctx, model.UserEmailInput{Email: "baru@integrione.id", CurrentPassword: "Password123!"})
				return p.Errors, p.WasSuccessful, err
			},
		},
		{
			name: "ChangeUserPassword",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.ChangeUserPassword(ctx, model.ChangeUserPasswordInput{CurrentPassword: "Password123!", NewPassword: "NewPassword123!"})
				return p.Errors, p.WasSuccessful, err
			},
		},
		{
			name: "EnrollMyTotp",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.EnrollMyTotp(ctx)
				return p.Errors, p.WasSuccessful, err
			},
		},
		{
			name: "ConfirmMyTotp",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.ConfirmMyTotp(ctx, model.ConfirmTotpInput{Code: "123456"})
				return p.Errors, p.WasSuccessful, err
			},
		},
		{
			name: "DisableMyTotp",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.DisableMyTotp(ctx)
				return p.Errors, p.WasSuccessful, err
			},
		},
		{
			name: "RegenerateMyRecoveryCodes",
			call: func(ctx context.Context) ([]*model.MutationError, bool, error) {
				p, err := h.adapter.RegenerateMyRecoveryCodes(ctx)
				return p.Errors, p.WasSuccessful, err
			},
		},
	}

	for _, sc := range sensitiveCases {
		sc := sc
		t.Run("Sensitive/"+sc.name, func(t *testing.T) {
			errs, ok, err := sc.call(noSelfCtx)
			if err != nil {
				t.Fatalf("%s(noSelfCtx) returned err=%v", sc.name, err)
			}
			assertMutDenied(sc.name+"(noSelf)", errs, ok)

			errs, ok, err = sc.call(staleUpdateCtx)
			if err != nil {
				t.Fatalf("%s(staleUpdateCtx) returned err=%v", sc.name, err)
			}
			assertMutDenied(sc.name+"(stale)", errs, ok)

			errs, ok, err = sc.call(unverifiedMfaCtx)
			if err != nil {
				t.Fatalf("%s(unverifiedMfaCtx) returned err=%v", sc.name, err)
			}
			assertMutDenied(sc.name+"(unverifiedMFA)", errs, ok)

			errs, ok, err = sc.call(fullFreshCtx)
			if err != nil || !ok || len(errs) != 0 {
				t.Fatalf("%s(fullFreshCtx) expected success, got ok=%v errs=%+v err=%v", sc.name, ok, errs, err)
			}
		})
	}
}

func TestSelfServiceNestedAndNode_AntiIDORAndZeroOracle(t *testing.T) {
	t.Parallel()

	const (
		selfID  = "11111111-1111-4111-8111-111111111111"
		otherID = "22222222-2222-4222-8222-222222222222"
		adminID = "99999999-9999-4999-8999-999999999999"
	)

	const (
		sessPage1ID   = "11111111-1111-4111-8111-bbbbbbbbbbb1"
		sessPage2ID   = "11111111-1111-4111-8111-bbbbbbbbbbb2"
		sessOtherID   = "22222222-2222-4222-8222-bbbbbbbbbbb1"
		sessMissingID = "99999999-9999-4999-8999-bbbbbbbbbbb9"
		identSelfID   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		identOtherID  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
		identMissID   = "99999999-9999-4999-8999-aaaaaaaaaaaa"
	)

	h := newSelfRBACHarness(t)
	h.sessions.sessionsByUser[selfID] = []*iamv1.Session{
		{Id: sessPage1ID, UserId: selfID, CreatedAt: timestamppb.New(h.now)},
		{Id: sessPage2ID, UserId: selfID, CreatedAt: timestamppb.New(h.now)},
	}
	h.sessions.sessionsByUser[otherID] = []*iamv1.Session{
		{Id: sessOtherID, UserId: otherID, CreatedAt: timestamppb.New(h.now)},
	}

	selfCtx := h.registerSession("tok-nested-self", sessPage1ID, selfID, "vendor", h.now, false, false, nil, []authz.PermissionGrant{
		{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: selfID}},
		{Code: authz.PermSessionsReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: selfID}},
	})
	otherIDORSessCtx := h.registerSession("tok-nested-no-perm", "sess-no-perm", selfID, "vendor", h.now, false, false, nil, nil)
	adminCtx := h.registerSession("tok-nested-admin", "sess-admin", adminID, "employee", h.now, false, false, []authz.RoleGrant{{Code: "iam_readonly"}}, []authz.PermissionGrant{
		{Code: authz.PermUsersRead, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermRolesRead, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermGroupsRead, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermSessionsRead, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
	})

	nestedCases := []struct {
		name string
		call func(ctx context.Context, userID string) error
	}{
		{"UserForNested", func(ctx context.Context, uid string) error { _, err := h.adapter.UserForNested(ctx, uid); return err }},
		{"UserIdentitiesForUser", func(ctx context.Context, uid string) error {
			_, err := h.adapter.UserIdentitiesForUser(ctx, uid, nil, nil, nil, nil)
			return err
		}},
		{"UserRolesForUser", func(ctx context.Context, uid string) error {
			_, err := h.adapter.UserRolesForUser(ctx, uid, nil, nil, nil, nil, nil)
			return err
		}},
		{"UserEffectiveRolesForUser", func(ctx context.Context, uid string) error {
			_, err := h.adapter.UserEffectiveRolesForUser(ctx, uid, nil, nil, nil, nil)
			return err
		}},
		{"UserGroupsForUser", func(ctx context.Context, uid string) error {
			_, err := h.adapter.UserGroupsForUser(ctx, uid, nil, nil, nil, nil)
			return err
		}},
		{"UserSessionsForUser", func(ctx context.Context, uid string) error {
			_, err := h.adapter.UserSessionsForUser(ctx, uid, nil, nil, nil, nil, nil)
			return err
		}},
	}

	for _, nc := range nestedCases {
		nc := nc
		t.Run(nc.name, func(t *testing.T) {
			if err := nc.call(selfCtx, selfID); err != nil {
				t.Fatalf("%s own failed: %v", nc.name, err)
			}
			if err := nc.call(otherIDORSessCtx, selfID); err == nil {
				t.Fatalf("%s own without *self should fail", nc.name)
			} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
				t.Fatalf("%s own without *self expected FORBIDDEN, got %v", nc.name, err)
			}
			// IDOR lintas pengguna dengan hanya izin *self wajib ditolak FORBIDDEN.
			if err := nc.call(selfCtx, otherID); err == nil {
				t.Fatalf("%s cross-user IDOR must fail for *self holder", nc.name)
			} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
				t.Fatalf("%s cross-user IDOR expected FORBIDDEN, got %v", nc.name, err)
			}
			// Pemegang izin admin global tetap dapat mengakses data lintas pengguna.
			if err := nc.call(adminCtx, otherID); err != nil {
				t.Fatalf("%s admin cross-user failed: %v", nc.name, err)
			}
		})
	}

	// Uji Node(ResourceUser / ResourceUserIdentity / ResourceSession) anti-IDOR dan bebas oracle keberadaan.
	nodeVal, err := h.adapter.Node(selfCtx, EncodeGID(ResourceUser, selfID))
	if err != nil || nodeVal == nil {
		t.Fatalf("Node own user failed: %v", err)
	}
	if _, err := h.adapter.Node(selfCtx, EncodeGID(ResourceUser, otherID)); err == nil {
		t.Fatal("Node other user must be denied FORBIDDEN")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected FORBIDDEN on foreign Node user, got %v", err)
	}

	ownIdentNode, err := h.adapter.Node(selfCtx, EncodeGID(ResourceUserIdentity, identSelfID))
	if err != nil || ownIdentNode == nil {
		t.Fatalf("Node own identity failed: %v", err)
	}
	_, foreignIdentErr := h.adapter.Node(selfCtx, EncodeGID(ResourceUserIdentity, identOtherID))
	_, nonexistentIdentErr := h.adapter.Node(selfCtx, EncodeGID(ResourceUserIdentity, identMissID))
	if foreignIdentErr == nil || nonexistentIdentErr == nil || foreignIdentErr.Error() != nonexistentIdentErr.Error() {
		t.Fatalf("expected identical FORBIDDEN anti-oracle errors for foreign vs nonexistent identity node, got %v vs %v", foreignIdentErr, nonexistentIdentErr)
	}

	ownSessNode, err := h.adapter.Node(selfCtx, EncodeGID(ResourceSession, sessPage2ID))
	if err != nil || ownSessNode == nil {
		t.Fatalf("Node own session failed: %v", err)
	}
	_, foreignSessErr := h.adapter.Node(selfCtx, EncodeGID(ResourceSession, sessOtherID))
	_, nonexistentSessErr := h.adapter.Node(selfCtx, EncodeGID(ResourceSession, sessMissingID))
	if foreignSessErr == nil || nonexistentSessErr == nil || foreignSessErr.Error() != nonexistentSessErr.Error() {
		t.Fatalf("expected identical FORBIDDEN anti-oracle errors for foreign vs nonexistent session node, got %v vs %v", foreignSessErr, nonexistentSessErr)
	}
}

func TestRevokeSessionAndRoleDelegationGuards(t *testing.T) {
	t.Parallel()

	const (
		selfID      = "11111111-1111-4111-8111-111111111111"
		otherID     = "22222222-2222-4222-8222-222222222222"
		operatorID  = "33333333-3333-4333-8333-333333333333"
		roleID      = "44444444-4444-4444-8444-444444444444"
		groupID     = "55555555-5555-4555-8555-555555555555"
		permReadID  = "66666666-6666-4666-8666-666666666666"
		permDelID   = "77777777-7777-4777-8777-777777777777"
		sessCurr    = "11111111-1111-4111-8111-ccccccccccc1"
		sessSecond  = "11111111-1111-4111-8111-ccccccccccc2"
		sessForeign = "22222222-2222-4222-8222-ccccccccccc1"
		sessMissing = "99999999-9999-4999-8999-ccccccccccc9"
	)

	h := newSelfRBACHarness(t)
	h.sessions.sessionsByUser[selfID] = []*iamv1.Session{
		{Id: sessCurr, UserId: selfID, CreatedAt: timestamppb.New(h.now)},
		{Id: sessSecond, UserId: selfID, CreatedAt: timestamppb.New(h.now)},
	}
	h.sessions.sessionsByUser[otherID] = []*iamv1.Session{
		{Id: sessForeign, UserId: otherID, CreatedAt: timestamppb.New(h.now)},
	}
	h.perms.perms[permReadID] = &iamv1.Permission{Id: permReadID, Code: authz.PermUsersReadSelf}
	h.perms.perms[permDelID] = &iamv1.Permission{Id: permDelID, Code: authz.PermUsersDeleteSelf}
	h.roles.roles[roleID] = &iamv1.Role{Id: roleID, Code: "custom.self", Name: "Custom Self", IsActive: true}

	selfFreshCtx := h.registerSession("tok-rev-fresh", sessCurr, selfID, "supplier", h.now, false, false, nil, fullSelfPerms(selfID))
	selfStaleCtx := h.registerSession("tok-rev-stale", sessCurr, selfID, "supplier", h.now.Add(-20*time.Minute), false, false, nil, fullSelfPerms(selfID))
	noDeleteSelfCtx := h.registerSession("tok-rev-nodelete", sessCurr, selfID, "supplier", h.now, false, false, nil, []authz.PermissionGrant{
		{Code: authz.PermSessionsReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: selfID}},
	})
	adminCtx := h.registerSession("tok-rev-admin", "sess-admin", "admin-id", "employee", h.now, false, false, []authz.RoleGrant{{Code: "superadmin"}}, []authz.PermissionGrant{
		{Code: authz.PermSessionsRevoke, IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermRolesWrite, IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermGroupsWrite, IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
	})

	// 1. RevokeSession sesi sekunder milik sendiri dengan iam.sessions.deleteself + reauth segar berhasil.
	if p, err := h.adapter.RevokeSession(selfFreshCtx, EncodeGID(ResourceSession, sessSecond), nil); err != nil || !p.WasSuccessful || p.Revoked == nil || !*p.Revoked {
		t.Fatalf("RevokeSession own secondary session failed: payload=%+v err=%v", p, err)
	}
	// Sesi kedaluwarsa >10m dan tanpa deleteself ditolak.
	if p, _ := h.adapter.RevokeSession(selfStaleCtx, EncodeGID(ResourceSession, sessCurr), nil); p.WasSuccessful || len(p.Errors) == 0 || p.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected AUTHORIZATION on stale RevokeSession, got %+v", p)
	}
	if p, _ := h.adapter.RevokeSession(noDeleteSelfCtx, EncodeGID(ResourceSession, sessCurr), nil); p.WasSuccessful || len(p.Errors) == 0 || p.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected AUTHORIZATION without deleteself, got %+v", p)
	}

	// Sesi milik orang lain dan sesi tidak ada harus sama-sama jatuh ke iam.sessions.revoke -> zero oracle.
	beforeRevokeCalls := h.sessions.revokeCalls
	pForeign, _ := h.adapter.RevokeSession(selfFreshCtx, EncodeGID(ResourceSession, sessForeign), nil)
	pMissing, _ := h.adapter.RevokeSession(selfFreshCtx, EncodeGID(ResourceSession, sessMissing), nil)
	if h.sessions.revokeCalls != beforeRevokeCalls {
		t.Fatalf("RevokeSessionById gRPC should not be invoked for unauthorized caller")
	}
	if pForeign.WasSuccessful || pMissing.WasSuccessful || len(pForeign.Errors) == 0 || len(pMissing.Errors) == 0 ||
		pForeign.Errors[0].Message != pMissing.Errors[0].Message || pForeign.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("expected identical AUTHORIZATION anti-oracle response for foreign vs missing session, got %+v vs %+v", pForeign, pMissing)
	}

	// Admin dengan iam.sessions.revoke dapat mencabut sesi lintas pengguna.
	if pAdm, err := h.adapter.RevokeSession(adminCtx, EncodeGID(ResourceSession, sessForeign), nil); err != nil || !pAdm.WasSuccessful {
		t.Fatalf("admin RevokeSession foreign failed: payload=%+v err=%v", pAdm, err)
	}

	// 2. Pembatasan cakupan SELF pada AssignUserRole & AssignGroupRole.
	mismatchScope := otherID
	if pBadUserScope, err := h.adapter.AssignUserRole(adminCtx, model.AssignUserRoleInput{
		UserID:    selfID,
		RoleID:    roleID,
		ScopeType: model.ScopeTypeSelf,
		ScopeID:   &mismatchScope,
	}); err != nil || pBadUserScope.WasSuccessful || len(pBadUserScope.Errors) == 0 || pBadUserScope.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT on mismatched self scope_id in AssignUserRole, got payload=%+v err=%v", pBadUserScope, err)
	}
	matchScope := selfID
	if pGoodUserScope, err := h.adapter.AssignUserRole(adminCtx, model.AssignUserRoleInput{
		UserID:    selfID,
		RoleID:    roleID,
		ScopeType: model.ScopeTypeSelf,
		ScopeID:   &matchScope,
	}); err != nil || !pGoodUserScope.WasSuccessful {
		t.Fatalf("expected valid self scope AssignUserRole to succeed, got payload=%+v err=%v", pGoodUserScope, err)
	}
	if pBadGroupScope, err := h.adapter.AssignGroupRole(adminCtx, model.AssignGroupRoleInput{
		GroupID:   groupID,
		RoleID:    roleID,
		ScopeType: model.ScopeTypeSelf,
		ScopeID:   &matchScope,
	}); err != nil || pBadGroupScope.WasSuccessful || len(pBadGroupScope.Errors) == 0 || pBadGroupScope.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT on AssignGroupRole with ScopeTypeSelf, got payload=%+v err=%v", pBadGroupScope, err)
	}

	// 3. ensureCanDelegatePermissionCodes dan eskalasi vertikal untuk pemegang *self saja vs operator IAM.
	if pSelfEscalate, err := h.adapter.AssignRolePermission(selfFreshCtx, roleID, permReadID); err != nil || pSelfEscalate.WasSuccessful || len(pSelfEscalate.Errors) == 0 || pSelfEscalate.Errors[0].Type != model.ErrorTypeEnumAuthorization {
		t.Fatalf("self-only user must be blocked from AssignRolePermission (vertical escalation), got payload=%+v err=%v", pSelfEscalate, err)
	}

	opCtx := h.registerSession("tok-op", "sess-op", operatorID, "employee", h.now, false, false, []authz.RoleGrant{{Code: "iam_operator"}}, []authz.PermissionGrant{
		{Code: authz.PermRolesWrite, IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
		{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: operatorID}},
	})
	if pOpOK, err := h.adapter.AssignRolePermission(opCtx, roleID, permReadID); err != nil || !pOpOK.WasSuccessful {
		t.Fatalf("IAM operator holding %s(ScopeSelf) expected to delegate %s, got payload=%+v err=%v", authz.PermUsersReadSelf, authz.PermUsersReadSelf, pOpOK, err)
	}
	if pOpDeny, err := h.adapter.AssignRolePermission(opCtx, roleID, permDelID); err != nil || pOpDeny.WasSuccessful || len(pOpDeny.Errors) == 0 || !strings.Contains(pOpDeny.Errors[0].Message, "delegate") {
		t.Fatalf("IAM operator lacking %s must be denied delegating it, got payload=%+v err=%v", authz.PermUsersDeleteSelf, pOpDeny, err)
	}

	// Verifikasi ensureCanDelegatePermissionCodes juga menolak jika target *self ber-scope user lain.
	wrongSelfOpSess := &middleware.ResolvedSession{
		Token: &authsession.SessionTokenRecord{UserID: operatorID},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: operatorID,
			Permissions: []authz.PermissionGrant{
				{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: otherID}},
			},
		},
	}
	if err := ensureCanDelegatePermissionCodes(wrongSelfOpSess, []string{authz.PermUsersReadSelf}); err == nil {
		t.Fatal("ensureCanDelegatePermissionCodes must reject *self grant scoped to another user")
	}
}
