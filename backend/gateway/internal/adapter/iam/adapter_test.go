package iam

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type memSessionStore struct {
	sessions map[string]*authsession.SessionValidationContext
	preAuths map[string]authsession.PreAuthRecord
	rates    map[string]int64
}

func newMemStore() *memSessionStore {
	return &memSessionStore{
		sessions: make(map[string]*authsession.SessionValidationContext),
		preAuths: make(map[string]authsession.PreAuthRecord),
		rates:    make(map[string]int64),
	}
}

func (s *memSessionStore) GetGate(ctx context.Context) (*authsession.SecurityGateRecord, error) {
	return &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1}, nil
}

func (s *memSessionStore) ReadSessionContext(ctx context.Context, tokenHashHex string, now time.Time) (*authsession.SessionValidationContext, error) {
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

func (s *memSessionStore) TouchActivity(ctx context.Context, tokenHashHex string, tok authsession.SessionTokenRecord, idleTimeout time.Duration, now time.Time) error {
	return nil
}

func (s *memSessionStore) SetPreAuth(ctx context.Context, rec authsession.PreAuthRecord, ttl time.Duration) error {
	s.preAuths[rec.TokenHashHex] = rec
	return nil
}

func (s *memSessionStore) GetPreAuth(ctx context.Context, cookieHashHex string, now time.Time) (*authsession.PreAuthRecord, error) {
	if rec, ok := s.preAuths[cookieHashHex]; ok {
		return &rec, nil
	}
	return nil, authsession.ErrSessionNotFound
}

func (s *memSessionStore) DeletePreAuth(ctx context.Context, cookieHashHex string) error {
	delete(s.preAuths, cookieHashHex)
	return nil
}

func (s *memSessionStore) CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) (int64, error) {
	k := scope + "|" + subject
	s.rates[k]++
	if s.rates[k] > limit {
		return s.rates[k], authsession.ErrRateLimitExceeded
	}
	return s.rates[k], nil
}

type mockAuthClient struct {
	iamv1.AuthServiceClient
	authFn      func(context.Context, *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error)
	resetFn     func(context.Context, *iamv1.RequestPasswordResetRequest) (*iamv1.RequestPasswordResetResponse, error)
	changePwdFn func(context.Context, *iamv1.ChangePasswordRequest) (*iamv1.ChangePasswordResponse, error)
}

func (m *mockAuthClient) AuthenticatePassword(ctx context.Context, in *iamv1.AuthenticatePasswordRequest, _ ...grpc.CallOption) (*iamv1.AuthenticatePasswordResponse, error) {
	return m.authFn(ctx, in)
}

func (m *mockAuthClient) RequestPasswordReset(ctx context.Context, in *iamv1.RequestPasswordResetRequest, _ ...grpc.CallOption) (*iamv1.RequestPasswordResetResponse, error) {
	return m.resetFn(ctx, in)
}

func (m *mockAuthClient) ChangePassword(ctx context.Context, in *iamv1.ChangePasswordRequest, _ ...grpc.CallOption) (*iamv1.ChangePasswordResponse, error) {
	return m.changePwdFn(ctx, in)
}

type mockUserClient struct {
	iamv1.UserServiceClient
	lastCreateKind iamv1.UserKind
	createFn       func(context.Context, *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error)
	listFn         func(context.Context, *iamv1.ListUsersRequest) (*iamv1.ListUsersResponse, error)
	getByIDFn      func(context.Context, *iamv1.GetUserByIdRequest) (*iamv1.GetUserByIdResponse, error)
	updProfileFn   func(context.Context, *iamv1.UpdateUserProfileRequest) (*iamv1.UpdateUserProfileResponse, error)
	updEmailFn     func(context.Context, *iamv1.UpdateUserEmailRequest) (*iamv1.UpdateUserEmailResponse, error)
	updUsernameFn  func(context.Context, *iamv1.UpdateUserUsernameRequest) (*iamv1.UpdateUserUsernameResponse, error)
	delPendingFn   func(context.Context, *iamv1.DeletePendingUserEmailRequest) (*iamv1.DeletePendingUserEmailResponse, error)
	confirmEmailFn func(context.Context, *iamv1.ConfirmEmailVerificationByTokenRequest) (*iamv1.ConfirmEmailVerificationByTokenResponse, error)
	resendEmailFn  func(context.Context, *iamv1.ResendValidationEmailTokenRequest) (*iamv1.ResendValidationEmailTokenResponse, error)
}

func (m *mockUserClient) CreateUser(ctx context.Context, in *iamv1.CreateUserRequest, _ ...grpc.CallOption) (*iamv1.CreateUserResponse, error) {
	m.lastCreateKind = in.GetKind()
	return m.createFn(ctx, in)
}

func (m *mockUserClient) ListUsers(ctx context.Context, in *iamv1.ListUsersRequest, _ ...grpc.CallOption) (*iamv1.ListUsersResponse, error) {
	return m.listFn(ctx, in)
}

func (m *mockUserClient) GetUserById(ctx context.Context, in *iamv1.GetUserByIdRequest, _ ...grpc.CallOption) (*iamv1.GetUserByIdResponse, error) {
	return m.getByIDFn(ctx, in)
}

func (m *mockUserClient) UpdateUserProfile(ctx context.Context, in *iamv1.UpdateUserProfileRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserProfileResponse, error) {
	return m.updProfileFn(ctx, in)
}

func (m *mockUserClient) UpdateUserEmail(ctx context.Context, in *iamv1.UpdateUserEmailRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserEmailResponse, error) {
	return m.updEmailFn(ctx, in)
}

func (m *mockUserClient) UpdateUserUsername(ctx context.Context, in *iamv1.UpdateUserUsernameRequest, _ ...grpc.CallOption) (*iamv1.UpdateUserUsernameResponse, error) {
	return m.updUsernameFn(ctx, in)
}

func (m *mockUserClient) DeletePendingUserEmail(ctx context.Context, in *iamv1.DeletePendingUserEmailRequest, _ ...grpc.CallOption) (*iamv1.DeletePendingUserEmailResponse, error) {
	return m.delPendingFn(ctx, in)
}

func (m *mockUserClient) ConfirmEmailVerificationByToken(ctx context.Context, in *iamv1.ConfirmEmailVerificationByTokenRequest, _ ...grpc.CallOption) (*iamv1.ConfirmEmailVerificationByTokenResponse, error) {
	return m.confirmEmailFn(ctx, in)
}

func (m *mockUserClient) ResendValidationEmailToken(ctx context.Context, in *iamv1.ResendValidationEmailTokenRequest, _ ...grpc.CallOption) (*iamv1.ResendValidationEmailTokenResponse, error) {
	return m.resendEmailFn(ctx, in)
}

type mockSessionClient struct {
	iamv1.SessionServiceClient
	getCSRFFn func(context.Context, *iamv1.GetSessionCsrfRequest) (*iamv1.GetSessionCsrfResponse, error)
}

func (m *mockSessionClient) GetSessionCsrf(ctx context.Context, in *iamv1.GetSessionCsrfRequest, _ ...grpc.CallOption) (*iamv1.GetSessionCsrfResponse, error) {
	return m.getCSRFFn(ctx, in)
}

type dummyRefresher struct{}

func (dummyRefresher) RefreshUserAccessSnapshot(context.Context, string, authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	return nil, nil
}

func TestUserCreationVariantsAndRBACIsolation(t *testing.T) {
	now := time.Now()
	store := newMemStore()

	rawTok := "admin-session-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-admin",
			UserID:       "admin-1",
			TokenHashHex: tokHash,
			UserKind:     "employee",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: "admin-1", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:         "admin-1",
			UserKind:            "employee",
			Status:              authz.StatusActive,
			AuthEpoch:           1,
			PermRevision:        1,
			RecoveryEpoch:       1,
			CanAccessIdentity:   true,
			DefaultRedirectPath: "/identity",
			SyncedAtUnix:        now.Unix(),
			FreshUntilUnix:      now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
				{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	}

	usersMock := &mockUserClient{
		createFn: func(_ context.Context, in *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			return &iamv1.CreateUserResponse{
				User: &iamv1.User{
					Id:          "new-user-id",
					Kind:        in.GetKind(),
					Status:      in.GetStatus(),
					DisplayName: in.GetDisplayName(),
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
		listFn: func(_ context.Context, _ *iamv1.ListUsersRequest) (*iamv1.ListUsersResponse, error) {
			return &iamv1.ListUsersResponse{
				Page: &iamv1.PageInfo{Total: 0, Limit: 20, Offset: 0, HasMore: false},
			}, nil
		},
	}

	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(source_connections.IAMClientSet{
		Auth:     &mockAuthClient{},
		Users:    usersMock,
		Sessions: &mockSessionClient{},
	}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
	})

	// 1. Uji keempat varian pembuatan pengguna memetakan ke UserKind yang tepat dengan input ringkas.
	uCust, err := adp.CreateUser(ctx, model.CreateUserInput{FullName: "Buyer", Email: "b@test.id", Username: "buyer1", Password: "Password123!"})
	if err != nil || !uCust.WasSuccessful || len(uCust.Errors) != 0 || uCust.User == nil || uCust.User.Kind != model.UserKindCustomer || usersMock.lastCreateKind != iamv1.UserKind_USER_KIND_CUSTOMER {
		t.Fatalf("CreateUser failed: payload=%+v err=%v", uCust, err)
	}
	if uCust.User.ID != "gid://integrione/user/new-user-id" || uCust.User.RawID != "new-user-id" {
		t.Fatalf("unexpected GID/RawID on created user: id=%q raw=%q", uCust.User.ID, uCust.User.RawID)
	}

	uEmp, err := adp.CreateEmployeeUser(ctx, model.CreateEmployeeUserInput{FullName: "Staff", Email: "s@test.id", Username: "staff1", Password: "Password123!"})
	if err != nil || !uEmp.WasSuccessful || uEmp.User == nil || uEmp.User.Kind != model.UserKindEmployee || usersMock.lastCreateKind != iamv1.UserKind_USER_KIND_EMPLOYEE {
		t.Fatalf("CreateEmployeeUser failed: payload=%+v err=%v", uEmp, err)
	}

	uVend, err := adp.CreateVendorUser(ctx, model.CreateVendorUserInput{FullName: "Vendor", Email: "v@test.id", Username: "vendor1", Password: "Password123!"})
	if err != nil || !uVend.WasSuccessful || uVend.User == nil || uVend.User.Kind != model.UserKindVendor || usersMock.lastCreateKind != iamv1.UserKind_USER_KIND_VENDOR {
		t.Fatalf("CreateVendorUser failed: payload=%+v err=%v", uVend, err)
	}

	uSupp, err := adp.CreateSupplierUser(ctx, model.CreateSupplierUserInput{FullName: "Supplier", Email: "sup@test.id", Username: "supplier1", Password: "Password123!"})
	if err != nil || !uSupp.WasSuccessful || uSupp.User == nil || uSupp.User.Kind != model.UserKindSupplier || usersMock.lastCreateKind != iamv1.UserKind_USER_KIND_SUPPLIER {
		t.Fatalf("CreateSupplierUser failed: payload=%+v err=%v", uSupp, err)
	}

	// 2. Uji isolasi RBAC (TestAB): pemegang iam.users.* ditolak untuk operasi peran dengan MutationError bertipe AUTHORIZATION.
	roleRes, err := adp.CreateRole(ctx, model.CreateRoleInput{
		Code:                "finance.viewer",
		Name:                "Finance Viewer",
		ManagementDomain:    "finance",
		DefaultRedirectPath: "/finance",
	})
	if err != nil || roleRes.WasSuccessful || len(roleRes.Errors) != 1 || roleRes.Errors[0].Type != model.ErrorTypeEnumAuthorization || roleRes.Errors[0].Message == "" {
		t.Fatalf("expected AUTHORIZATION MutationError on CreateRole, got payload=%+v err=%v", roleRes, err)
	}
}

func TestUserResetPasswordAntiEnumerationAndLoginCookie(t *testing.T) {
	now := time.Now()
	store := newMemStore()

	authClient := &mockAuthClient{
		resetFn: func(_ context.Context, _ *iamv1.RequestPasswordResetRequest) (*iamv1.RequestPasswordResetResponse, error) {
			return nil, status.Error(codes.NotFound, "account not found")
		},
		authFn: func(_ context.Context, _ *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error) {
			return &iamv1.AuthenticatePasswordResponse{
				Status:       iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED,
				RedirectPath: "/identity",
				IssuedSession: &iamv1.IssuedSession{
					SessionToken: "promoted-session-tok",
					CsrfToken:    "promoted-csrf",
					Session: &iamv1.Session{
						AbsoluteExpiresAt: timestamppb.New(now.Add(8 * time.Hour)),
					},
				},
			}, nil
		},
	}

	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, _ := NewAdapter(source_connections.IAMClientSet{Auth: authClient}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})

	// Validasi email kosong mengembalikan ARGUMENT MutationError.
	badReset, err := adp.UserResetPassword(context.Background(), model.UserResetPasswordInput{Email: "   "})
	if err != nil || badReset.WasSuccessful || len(badReset.Errors) != 1 || badReset.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT MutationError on empty email reset, got payload=%+v err=%v", badReset, err)
	}

	// Anti-enumerasi reset sandi: NotFound tetap mengembalikan was_successful true dengan pesan netral.
	resReset, err := adp.UserResetPassword(context.Background(), model.UserResetPasswordInput{Email: "nobody@test.id"})
	if err != nil || !resReset.WasSuccessful || len(resReset.Errors) != 0 || resReset.Message == nil || *resReset.Message == "" {
		t.Fatalf("expected anti-enumeration success payload, got payload=%+v err=%v", resReset, err)
	}

	// Login berhasil menerbitkan ResponseEffects X-CSRF-Token & Location aman.
	eff := browserauth.NewResponseEffects()
	loginCtx := browserauth.WithResponseEffects(context.Background(), eff)
	res, err := adp.UserLogin(loginCtx, model.UserLoginInput{Identifier: "admin@integrione.local", Password: "secret"})
	if err != nil || !res.WasSuccessful || len(res.Errors) != 0 {
		t.Fatalf("UserLogin failed: payload=%+v err=%v", res, err)
	}
	if res.Status == nil || *res.Status != model.UserLoginStatusAuthenticated || res.RedirectLocation == nil || *res.RedirectLocation != "/identity" {
		t.Fatalf("unexpected login payload: %+v", res)
	}
}

func TestGIDCodecAndNodeRBAC(t *testing.T) {
	userUUID := "11111111-1111-4111-8111-111111111111"
	otherUUID := "22222222-2222-4222-8222-222222222222"

	gid := EncodeGID(ResourceUser, userUUID)
	if gid != "gid://integrione/user/"+userUUID {
		t.Fatalf("unexpected GID format: %q", gid)
	}

	res, parsed, err := DecodeNodeGID(gid)
	if err != nil || res != ResourceUser || parsed != userUUID {
		t.Fatalf("DecodeNodeGID failed: res=%q id=%q err=%v", res, parsed, err)
	}

	if got, err := ParseTypedID(gid, ResourceUser); err != nil || got != userUUID {
		t.Fatalf("ParseTypedID(GID) failed: got=%q err=%v", got, err)
	}
	if got, err := ParseTypedID(userUUID, ResourceUser); err != nil || got != userUUID {
		t.Fatalf("ParseTypedID(UUID) failed: got=%q err=%v", got, err)
	}
	if _, err := ParseTypedID(EncodeGID(ResourceRole, userUUID), ResourceUser); err == nil {
		t.Fatal("expected mismatch resource error from ParseTypedID")
	}
	if _, _, err := DecodeNodeGID("gid://integrione/group-role/" + userUUID); err == nil {
		t.Fatal("expected non-Node resource rejection from DecodeNodeGID")
	}
	if _, _, err := DecodeNodeGID("gid://integrione/user/not-a-uuid"); err == nil {
		t.Fatal("expected invalid UUID rejection from DecodeNodeGID")
	}

	now := time.Now()
	store := newMemStore()
	rawTok := "self-user-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-self",
			UserID:       userUUID,
			TokenHashHex: tokHash,
			UserKind:     "customer",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: userUUID, Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    userUUID,
			UserKind:       "customer",
			Status:         authz.StatusActive,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userUUID}},
			},
		},
	}

	usersMock := &mockUserClient{
		getByIDFn: func(_ context.Context, in *iamv1.GetUserByIdRequest) (*iamv1.GetUserByIdResponse, error) {
			return &iamv1.GetUserByIdResponse{
				User: &iamv1.User{
					Id:          in.GetUserId(),
					Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
					Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
					DisplayName: "Customer Self",
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
	}

	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/"})
	adp, _ := NewAdapter(source_connections.IAMClientSet{Users: usersMock}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})

	ctx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})

	// Pemilik akun dapat mengambil Node miliknya sendiri tanpa izin iam.users.read.
	nodeVal, err := adp.Node(ctx, gid)
	if err != nil {
		t.Fatalf("Node self lookup failed: %v", err)
	}
	uNode, ok := nodeVal.(*model.User)
	if !ok || uNode.ID != gid || uNode.RawID != userUUID {
		t.Fatalf("unexpected Node result: %+v", nodeVal)
	}

	// Mengakses Node pengguna lain tanpa izin iam.users.read harus ditolak FORBIDDEN.
	if _, err := adp.Node(ctx, EncodeGID(ResourceUser, otherUUID)); err == nil {
		t.Fatal("expected FORBIDDEN when reading another user via Node without permission")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeForbidden {
		t.Fatalf("expected FORBIDDEN, got %v", err)
	}
}

func TestRelayPaginationForwardBackwardAndInvalidCursor(t *testing.T) {
	items := []string{"u0", "u1", "u2", "u3", "u4"}

	// Forward pagination: first=2, setelah cursor offset 1 -> ["u2", "u3"], hasPrev=true, hasNext=true.
	c1 := EncodeCursor(1)
	first2 := 2
	sliced, startOffset, pageInfo, total, err := SliceRelayWindow(items, &c1, nil, &first2, nil)
	if err != nil {
		t.Fatalf("forward SliceRelayWindow failed: %v", err)
	}
	if total != 5 || startOffset != 2 || len(sliced) != 2 || sliced[0] != "u2" || sliced[1] != "u3" {
		t.Fatalf("unexpected forward slice: total=%d start=%d items=%v", total, startOffset, sliced)
	}
	if !pageInfo.HasPreviousPage || !pageInfo.HasNextPage || *pageInfo.StartCursor != EncodeCursor(2) || *pageInfo.EndCursor != EncodeCursor(3) {
		t.Fatalf("unexpected forward pageInfo: %+v", pageInfo)
	}

	// Backward pagination: last=2 sebelum cursor offset 4 -> ["u2", "u3"], hasPrev=true, hasNext=true.
	c4 := EncodeCursor(4)
	last2 := 2
	bSliced, bStart, bInfo, bTotal, err := SliceRelayWindow(items, nil, &c4, nil, &last2)
	if err != nil {
		t.Fatalf("backward SliceRelayWindow failed: %v", err)
	}
	if bTotal != 5 || bStart != 2 || len(bSliced) != 2 || bSliced[0] != "u2" || bSliced[1] != "u3" {
		t.Fatalf("unexpected backward slice: total=%d start=%d items=%v", bTotal, bStart, bSliced)
	}
	if !bInfo.HasPreviousPage || !bInfo.HasNextPage {
		t.Fatalf("unexpected backward pageInfo: %+v", bInfo)
	}

	// Kursor tidak valid harus mengembalikan INVALID_ARGUMENT.
	badCursor := "not-valid-base64!!!"
	if _, err := ParseRelayPage(&badCursor, nil, &first2, nil); err == nil {
		t.Fatal("expected error for invalid cursor")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for bad cursor, got %v", err)
	}
}

func TestEmailVerificationResendAndSelfMutations(t *testing.T) {
	now := time.Now()
	userUUID := "33333333-3333-4333-8333-333333333333"
	store := newMemStore()
	rawTok := "self-mut-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-mut",
			UserID:       userUUID,
			TokenHashHex: tokHash,
			UserKind:     "customer",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: userUUID, Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    userUUID,
			UserKind:       "customer",
			Status:         authz.StatusActive,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userUUID}},
				{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userUUID}},
				{Code: authz.PermUsersDeleteSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userUUID}},
			},
		},
	}

	var capturedProfileReq *iamv1.UpdateUserProfileRequest
	usersMock := &mockUserClient{
		getByIDFn: func(_ context.Context, in *iamv1.GetUserByIdRequest) (*iamv1.GetUserByIdResponse, error) {
			return &iamv1.GetUserByIdResponse{
				User: &iamv1.User{
					Id:          in.GetUserId(),
					Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
					Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
					DisplayName: "Budi Santoso",
					Locale:      "id-ID",
					Timezone:    "Asia/Jakarta",
					AvatarUrl:   "https://erp.localhost:8443/media/avatars/existing.avif",
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
		updProfileFn: func(_ context.Context, in *iamv1.UpdateUserProfileRequest) (*iamv1.UpdateUserProfileResponse, error) {
			capturedProfileReq = in
			avatar := in.GetAvatarUrl()
			if in.GetPreserveAvatarUrl() {
				avatar = "https://erp.localhost:8443/media/avatars/existing.avif"
			}
			return &iamv1.UpdateUserProfileResponse{
				User: &iamv1.User{
					Id:          in.GetUserId(),
					Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
					Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
					DisplayName: in.GetDisplayName(),
					AvatarUrl:   avatar,
					Locale:      in.GetLocale(),
					Timezone:    in.GetTimezone(),
					CreatedAt:   timestamppb.New(now),
					UpdatedAt:   timestamppb.New(now),
				},
			}, nil
		},
		updEmailFn: func(_ context.Context, in *iamv1.UpdateUserEmailRequest) (*iamv1.UpdateUserEmailResponse, error) {
			return &iamv1.UpdateUserEmailResponse{
				PendingEmail: in.GetEmail(),
				IsPending:    true,
				ExpiresAt:    timestamppb.New(now.Add(24 * time.Hour)),
			}, nil
		},
		delPendingFn: func(_ context.Context, _ *iamv1.DeletePendingUserEmailRequest) (*iamv1.DeletePendingUserEmailResponse, error) {
			return &iamv1.DeletePendingUserEmailResponse{Deleted: true}, nil
		},
		updUsernameFn: func(_ context.Context, in *iamv1.UpdateUserUsernameRequest) (*iamv1.UpdateUserUsernameResponse, error) {
			switch in.GetUsername() {
			case "sancyty":
				st, _ := status.New(codes.FailedPrecondition, "username unchanged").WithDetails(&errdetails.ErrorInfo{Reason: "iam.identity.same_username"})
				return nil, st.Err()
			case "taken_user":
				st, _ := status.New(codes.AlreadyExists, "username already registered").WithDetails(&errdetails.ErrorInfo{Reason: "iam.identity.already_exists"})
				return nil, st.Err()
			}
			return &iamv1.UpdateUserUsernameResponse{Username: in.GetUsername()}, nil
		},
		confirmEmailFn: func(_ context.Context, in *iamv1.ConfirmEmailVerificationByTokenRequest) (*iamv1.ConfirmEmailVerificationByTokenResponse, error) {
			if in.GetRawToken() == "expired-tok" {
				return &iamv1.ConfirmEmailVerificationByTokenResponse{
					Verified: false,
					Status:   iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_EXPIRED,
				}, nil
			}
			return &iamv1.ConfirmEmailVerificationByTokenResponse{
				Verified: true,
				Status:   iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_VERIFIED,
			}, nil
		},
		resendEmailFn: func(_ context.Context, in *iamv1.ResendValidationEmailTokenRequest) (*iamv1.ResendValidationEmailTokenResponse, error) {
			if in.GetEmail() == "unknown@test.id" {
				return nil, status.Error(codes.NotFound, "pending email not found")
			}
			return &iamv1.ResendValidationEmailTokenResponse{IsResend: true}, nil
		},
	}

	authMock := &mockAuthClient{
		changePwdFn: func(_ context.Context, _ *iamv1.ChangePasswordRequest) (*iamv1.ChangePasswordResponse, error) {
			return &iamv1.ChangePasswordResponse{
				Changed:           true,
				SessionsRevoked:   2,
				PasswordChangedAt: timestamppb.New(now),
			}, nil
		},
	}

	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/"})
	adp, _ := NewAdapter(source_connections.IAMClientSet{Auth: authMock, Users: usersMock}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})

	authCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})

	// 1. VerifyEmail memetakan VERIFIED dan EXPIRED secara deterministik.
	vOK, err := adp.VerifyEmail(context.Background(), model.VerifyEmailInput{Token: "good-tok"})
	if err != nil || !vOK.WasSuccessful || vOK.Verified == nil || !*vOK.Verified || vOK.Status == nil || *vOK.Status != model.VerifyEmailStatusVerified {
		t.Fatalf("VerifyEmail verified failed: payload=%+v err=%v", vOK, err)
	}
	vExp, err := adp.VerifyEmail(context.Background(), model.VerifyEmailInput{Token: "expired-tok"})
	if err != nil || !vExp.WasSuccessful || vExp.Verified == nil || *vExp.Verified || vExp.Status == nil || *vExp.Status != model.VerifyEmailStatusExpired {
		t.Fatalf("VerifyEmail expired failed: payload=%+v err=%v", vExp, err)
	}

	// 2. ResendValidationEmailToken Kasus 1 (login) dan Kasus 2 (anonim anti-enumerasi).
	rAuth, err := adp.ResendValidationEmailToken(authCtx, nil)
	if err != nil || !rAuth.WasSuccessful || rAuth.IsResend == nil || !*rAuth.IsResend {
		t.Fatalf("ResendValidationEmailToken Case 1 failed: payload=%+v err=%v", rAuth, err)
	}
	unk := "unknown@test.id"
	rAnon, err := adp.ResendValidationEmailToken(context.Background(), &model.ResendValidationEmailInput{Email: &unk})
	if err != nil || !rAnon.WasSuccessful || rAnon.IsResend == nil || !*rAnon.IsResend {
		t.Fatalf("ResendValidationEmailToken Case 2 anti-enumeration failed: payload=%+v err=%v", rAnon, err)
	}
	rNoEmail, err := adp.ResendValidationEmailToken(context.Background(), nil)
	if err != nil || rNoEmail.WasSuccessful || len(rNoEmail.Errors) != 1 || rNoEmail.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT MutationError when calling ResendValidationEmailToken unauthenticated without email, got payload=%+v err=%v", rNoEmail, err)
	}

	// 3. UpdateMyProfile menjaga avatar_url yang sudah ada (PreserveAvatarUrl = true).
	prof, err := adp.UpdateMyProfile(authCtx, model.UpdateMyProfileInput{DisplayName: "Budi Baru"})
	if err != nil || !prof.WasSuccessful || prof.User == nil || capturedProfileReq == nil || !capturedProfileReq.GetPreserveAvatarUrl() {
		t.Fatalf("UpdateMyProfile did not set PreserveAvatarUrl: payload=%+v req=%+v err=%v", prof, capturedProfileReq, err)
	}
	if prof.User.AvatarURL != "https://erp.localhost:8443/media/avatars/existing.avif" {
		t.Fatalf("UpdateMyProfile unexpectedly altered avatar_url: %q", prof.User.AvatarURL)
	}

	// 4. UploadAvatarUser menerima container AVIF valid dan menolak payload non-AVIF / XSS.
	avifBytes := []byte{0, 0, 0, 16, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f', 0, 0, 0, 0}
	dataURI := "data:image/avif;base64," + base64.StdEncoding.EncodeToString(avifBytes)
	upPayload, err := adp.UploadAvatarUser(authCtx, model.UserAvatarUploadInput{Image: dataURI})
	if err != nil || !upPayload.WasSuccessful || upPayload.AvatarURL == nil || !strings.HasPrefix(*upPayload.AvatarURL, DefaultAvatarBaseURL+"/"+userUUID+"/") || !strings.HasSuffix(*upPayload.AvatarURL, ".avif") {
		t.Fatalf("UploadAvatarUser valid AVIF failed: payload=%+v err=%v", upPayload, err)
	}
	badPNG := base64.StdEncoding.EncodeToString([]byte("\x89PNG\r\n\x1a\n<script>alert(1)</script>"))
	badUp, err := adp.UploadAvatarUser(authCtx, model.UserAvatarUploadInput{Image: badPNG})
	if err != nil || badUp.WasSuccessful || len(badUp.Errors) != 1 || badUp.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected non-AVIF ARGUMENT MutationError in UploadAvatarUser, got payload=%+v err=%v", badUp, err)
	}

	// 5. UpdateUserEmail & DeletePendingUserEmail.
	updEmail, err := adp.UpdateUserEmail(authCtx, model.UserEmailInput{Email: "baru@integrione.id", CurrentPassword: "OldPassword123!"})
	if err != nil || !updEmail.WasSuccessful || updEmail.IsPending == nil || !*updEmail.IsPending || updEmail.PendingEmail == nil || *updEmail.PendingEmail != "baru@integrione.id" {
		t.Fatalf("UpdateUserEmail failed: payload=%+v err=%v", updEmail, err)
	}
	delRes, err := adp.DeletePendingUserEmail(authCtx)
	if err != nil || !delRes.WasSuccessful || delRes.Deleted == nil || !*delRes.Deleted {
		t.Fatalf("DeletePendingUserEmail failed: payload=%+v err=%v", delRes, err)
	}

	// 6. ChangeUsername: sukses, input kosong (ARGUMENT), username sama & sudah dipakai (CONFLICT).
	chgOK, err := adp.ChangeUsername(authCtx, model.UsernameInput{Username: "kibo"})
	if err != nil || !chgOK.WasSuccessful || len(chgOK.Errors) != 0 || chgOK.Username == nil || *chgOK.Username != "kibo" {
		t.Fatalf("ChangeUsername success failed: payload=%+v err=%v", chgOK, err)
	}
	chgEmpty, err := adp.ChangeUsername(authCtx, model.UsernameInput{Username: "   "})
	if err != nil || chgEmpty.WasSuccessful || len(chgEmpty.Errors) != 1 || chgEmpty.Errors[0].Type != model.ErrorTypeEnumArgument {
		t.Fatalf("expected ARGUMENT on empty ChangeUsername, got payload=%+v err=%v", chgEmpty, err)
	}
	chgSame, err := adp.ChangeUsername(authCtx, model.UsernameInput{Username: "sancyty"})
	if err != nil || chgSame.WasSuccessful || len(chgSame.Errors) != 1 || chgSame.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT on same ChangeUsername, got payload=%+v err=%v", chgSame, err)
	}
	chgTaken, err := adp.ChangeUsername(authCtx, model.UsernameInput{Username: "taken_user"})
	if err != nil || chgTaken.WasSuccessful || len(chgTaken.Errors) != 1 || chgTaken.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT on taken ChangeUsername, got payload=%+v err=%v", chgTaken, err)
	}

	// 7. ChangeUserPassword mengembalikan ChangeUserPasswordPayload dan membersihkan cookie sesi.
	eff := browserauth.NewResponseEffects()
	pwdCtx := browserauth.WithResponseEffects(authCtx, eff)
	pwdRes, err := adp.ChangeUserPassword(pwdCtx, model.ChangeUserPasswordInput{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword123!",
	})
	rec := httptest.NewRecorder()
	eff.ApplyTo(rec)
	clearedCookies := rec.Result().Cookies()
	if err != nil || !pwdRes.WasSuccessful || pwdRes.Changed == nil || !*pwdRes.Changed || pwdRes.SessionsRevoked == nil || *pwdRes.SessionsRevoked != 2 || len(clearedCookies) == 0 || clearedCookies[0].MaxAge >= 0 {
		t.Fatalf("ChangeUserPassword failed: payload=%+v cookies=%+v err=%v", pwdRes, clearedCookies, err)
	}
}

func TestToMutationErrorsReasonMappings(t *testing.T) {
	conflictSt, _ := status.New(codes.AlreadyExists, "identity limit exceeded").WithDetails(&errdetails.ErrorInfo{Reason: "iam.identity.limit_exceeded"})
	unprocSt, _ := status.New(codes.FailedPrecondition, "cannot remove last identity").WithDetails(&errdetails.ErrorInfo{Reason: "iam.identity.cannot_remove_last"})
	sysImmutableSt, _ := status.New(codes.FailedPrecondition, "system role is immutable").WithDetails(&errdetails.ErrorInfo{Reason: "iam.role.system_immutable"})

	conflictErrs := ToMutationErrors(mapRPCError(conflictSt.Err()))
	if len(conflictErrs) != 1 || conflictErrs[0].Type != model.ErrorTypeEnumConflict || conflictErrs[0].Message == "" {
		t.Fatalf("unexpected conflict mutation errors: %+v", conflictErrs)
	}

	unprocErrs := ToMutationErrors(mapRPCError(unprocSt.Err()))
	if len(unprocErrs) != 1 || unprocErrs[0].Type != model.ErrorTypeEnumUnprocessableEntity || unprocErrs[0].Message == "" {
		t.Fatalf("unexpected unprocessable mutation errors: %+v", unprocErrs)
	}

	sysImmutableErrs := ToMutationErrors(mapRPCError(sysImmutableSt.Err()))
	if len(sysImmutableErrs) != 1 || sysImmutableErrs[0].Type != model.ErrorTypeEnumUnprocessableEntity || sysImmutableErrs[0].Message == "" {
		t.Fatalf("unexpected system immutable mutation errors: %+v", sysImmutableErrs)
	}
}
