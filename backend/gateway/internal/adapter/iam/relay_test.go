package iam

import (
	"context"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type mockSecurityAuthClient struct {
	iamv1.AuthServiceClient
	authFn      func(context.Context, *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error)
	verifyMFAFn func(context.Context, *iamv1.VerifyMfaChallengeRequest) (*iamv1.VerifyMfaChallengeResponse, error)
}

func (m *mockSecurityAuthClient) AuthenticatePassword(ctx context.Context, in *iamv1.AuthenticatePasswordRequest, _ ...grpc.CallOption) (*iamv1.AuthenticatePasswordResponse, error) {
	return m.authFn(ctx, in)
}

func (m *mockSecurityAuthClient) VerifyMfaChallenge(ctx context.Context, in *iamv1.VerifyMfaChallengeRequest, _ ...grpc.CallOption) (*iamv1.VerifyMfaChallengeResponse, error) {
	return m.verifyMFAFn(ctx, in)
}

type mockSecurityRoleClient struct {
	iamv1.RoleServiceClient
	createCalled bool
	revokeCalled bool
}

func (m *mockSecurityRoleClient) CreateRole(_ context.Context, _ *iamv1.CreateRoleRequest, _ ...grpc.CallOption) (*iamv1.CreateRoleResponse, error) {
	m.createCalled = true
	return nil, status.Error(codes.Internal, "unexpected CreateRole RPC call")
}

func (m *mockSecurityRoleClient) RevokeUserRole(_ context.Context, _ *iamv1.RevokeUserRoleRequest, _ ...grpc.CallOption) (*iamv1.RevokeUserRoleResponse, error) {
	m.revokeCalled = true
	return nil, status.Error(codes.Internal, "unexpected RevokeUserRole RPC call")
}

type mockSecurityUserClient struct {
	iamv1.UserServiceClient
	softDeleteCalled bool
}

func (m *mockSecurityUserClient) SoftDeleteUser(_ context.Context, _ *iamv1.SoftDeleteUserRequest, _ ...grpc.CallOption) (*iamv1.SoftDeleteUserResponse, error) {
	m.softDeleteCalled = true
	return nil, status.Error(codes.Internal, "unexpected SoftDeleteUser RPC call")
}

func TestRelayBoundsDecodeCursorAndBackwardUpperBoundNoPanic(t *testing.T) {
	if got, err := DecodeCursor(EncodeCursor(maxPaginationOffset)); err != nil || got != maxPaginationOffset {
		t.Fatalf("expected DecodeCursor at maxPaginationOffset=%d to pass, got (%d, %v)", maxPaginationOffset, got, err)
	}
	if _, err := DecodeCursor(EncodeCursor(maxPaginationOffset + 1)); err == nil {
		t.Fatal("expected DecodeCursor above maxPaginationOffset to fail")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT above maxPaginationOffset, got %v", err)
	}

	// Pastikan paginasi mundur tidak panik ketika total menyusut sehingga upperBound < offset.
	beforeOffset := 5
	win := &RelayWindow{
		StartOffset:  2,
		BeforeOffset: &beforeOffset,
		Limit:        5,
		IsBackward:   true,
	}
	items, offset, pageInfo, total, err := FetchRelayPage(win, func(_, _ int32) ([]string, *iamv1.PageInfo, error) {
		return []string{"a", "b"}, &iamv1.PageInfo{Total: 1}, nil
	})
	if err != nil {
		t.Fatalf("FetchRelayPage() unexpected error = %v", err)
	}
	if len(items) != 0 || offset != 2 || total != 1 || pageInfo == nil {
		t.Fatalf("unexpected clamped backward page: items=%v offset=%d total=%d pageInfo=%+v", items, offset, total, pageInfo)
	}

	// Tolak offset di atas batas sebelum memanggil RPC backend.
	rpcInvoked := false
	_, _, _, _, boundErr := FetchRelayPage(&RelayWindow{StartOffset: maxPaginationOffset + 2000, Limit: 10}, func(_, _ int32) ([]string, *iamv1.PageInfo, error) {
		rpcInvoked = true
		return nil, nil, nil
	})
	if boundErr == nil || rpcInvoked {
		t.Fatalf("expected clampRelayBounds error before RPC call, got err=%v rpcInvoked=%v", boundErr, rpcInvoked)
	}
}

func TestUserLoginAndValidateMFAAmbiguousErrorsAndThrottle(t *testing.T) {
	store := newMemStore()
	var upstreamErr error
	authClient := &mockSecurityAuthClient{
		authFn: func(_ context.Context, _ *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error) {
			return nil, upstreamErr
		},
		verifyMFAFn: func(_ context.Context, _ *iamv1.VerifyMfaChallengeRequest) (*iamv1.VerifyMfaChallengeResponse, error) {
			return nil, upstreamErr
		},
	}

	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})
	adp, err := NewAdapter(source_connections.IAMClientSet{Auth: authClient}, authMgr, authzSvc, rv, Config{
		LoginRateMax:    20,
		LoginRateWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	ctx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{ClientIP: "203.0.113.10"})

	// Login wajib menutupi NOT_FOUND, PERMISSION_DENIED, dan RESOURCE_EXHAUSTED dengan pesan AUTHENTICATION yang identik.
	var firstMsg string
	for _, code := range []codes.Code{codes.NotFound, codes.PermissionDenied, codes.ResourceExhausted} {
		upstreamErr = status.Errorf(code, "internal detail for %s", code.String())
		res, callErr := adp.UserLogin(ctx, model.UserLoginInput{Identifier: "probe@integrione.local", Password: "WrongPassword!"})
		if callErr != nil || res.WasSuccessful || len(res.Errors) != 1 {
			t.Fatalf("UserLogin(%s) unexpected payload=%+v err=%v", code, res, callErr)
		}
		if res.Errors[0].Type != model.ErrorTypeEnumAuthentication {
			t.Fatalf("UserLogin(%s) expected AUTHENTICATION, got %s (%s)", code, res.Errors[0].Type, res.Errors[0].Message)
		}
		if firstMsg == "" {
			firstMsg = res.Errors[0].Message
		} else if res.Errors[0].Message != firstMsg {
			t.Fatalf("UserLogin(%s) leaked distinct message %q vs %q", code, res.Errors[0].Message, firstMsg)
		}
	}

	// Percobaan login memakai akun otomasi system (termasuk password kosong) dengan ErrorInfo eksplisit wajib mengembalikan AUTHORIZATION "system cannot login".
	sysSt, _ := status.New(codes.PermissionDenied, "system cannot login").WithDetails(&errdetails.ErrorInfo{
		Domain: "iam",
		Reason: "iam.user.system_cannot_login",
	})
	upstreamErr = sysSt.Err()
	sysLoginRes, sysLoginErr := adp.UserLogin(ctx, model.UserLoginInput{Identifier: "auto_system_iam", Password: ""})
	if sysLoginErr != nil || sysLoginRes.WasSuccessful || len(sysLoginRes.Errors) != 1 || sysLoginRes.Errors[0].Type != model.ErrorTypeEnumAuthorization || sysLoginRes.Errors[0].Message != "system cannot login" {
		t.Fatalf("expected AUTHORIZATION 'system cannot login' for auto_system_iam, got payload=%+v err=%v", sysLoginRes, sysLoginErr)
	}

	// ValidateUserMfa menutupi CONFLICT replay TOTP menjadi AUTHENTICATION dan memblokir setelah 5 percobaan per challenge.
	upstreamErr = status.Error(codes.AlreadyExists, "totp code already consumed")
	totp := "123456"
	for attempt := 1; attempt <= 5; attempt++ {
		res, callErr := adp.ValidateUserMfa(ctx, model.ValidateUserMfaInput{Challenge: "mfa-chal-1", TotpCode: &totp})
		if callErr != nil || res.WasSuccessful || len(res.Errors) != 1 || res.Errors[0].Type != model.ErrorTypeEnumAuthentication {
			t.Fatalf("ValidateUserMfa attempt %d expected AUTHENTICATION, got payload=%+v err=%v", attempt, res, callErr)
		}
	}
	resThrottled, callErr := adp.ValidateUserMfa(ctx, model.ValidateUserMfaInput{Challenge: "mfa-chal-1", TotpCode: &totp})
	if callErr != nil || resThrottled.WasSuccessful || len(resThrottled.Errors) != 1 || resThrottled.Errors[0].Type != model.ErrorTypeEnumThrottle {
		t.Fatalf("expected 6th ValidateUserMfa attempt to return THROTTLE, got payload=%+v err=%v", resThrottled, callErr)
	}
}

func TestRBACSystemRoleCreationForbiddenAndIDORAndSelfDeleteGuards(t *testing.T) {
	now := time.Now()
	store := newMemStore()

	callerUUID := "11111111-1111-4111-8111-111111111111"
	otherUUID := "22222222-2222-4222-8222-222222222222"
	rawTok := "non-superadmin-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)

	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-operator",
			UserID:       callerUUID,
			TokenHashHex: tokHash,
			UserKind:     "employee",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: callerUUID, Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:         callerUUID,
			UserKind:            "employee",
			Status:              authz.StatusActive,
			MFAEnabled:          false,
			AuthEpoch:           1,
			PermRevision:        1,
			RecoveryEpoch:       1,
			CanAccessIdentity:   true,
			DefaultRedirectPath: "/identity",
			SyncedAtUnix:        now.Unix(),
			FreshUntilUnix:      now.Add(30 * time.Second).Unix(),
			Roles: []authz.RoleGrant{
				{Code: "iam_operator"},
			},
			Permissions: []authz.PermissionGrant{
				{Code: "iam.roles.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
				{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	}

	roleClient := &mockSecurityRoleClient{}
	userClient := &mockSecurityUserClient{}
	authMgr, _ := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(source_connections.IAMClientSet{
		Roles: roleClient,
		Users: userClient,
	}, authMgr, authzSvc, rv, Config{LoginRateMax: 10, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	authCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
		ClientIP:   "203.0.113.20",
	})

	isSys := true
	rolePayload, err := adp.CreateRole(authCtx, model.CreateRoleInput{
		Code:                "sys_escalate",
		Name:                "System Escalate",
		ManagementDomain:    "iam",
		DefaultRedirectPath: "/identity",
		IsSystem:            &isSys,
	})
	if err != nil || rolePayload.WasSuccessful || len(rolePayload.Errors) != 1 || rolePayload.Errors[0].Type != model.ErrorTypeEnumAuthorization || roleClient.createCalled {
		t.Fatalf("expected AUTHORIZATION error blocking is_system=true CreateRole, got payload=%+v err=%v rpc=%v", rolePayload, err, roleClient.createCalled)
	}

	delPayload, err := adp.SoftDeleteUser(authCtx, EncodeGID(ResourceUser, callerUUID))
	if err != nil || delPayload.WasSuccessful || len(delPayload.Errors) != 1 || delPayload.Errors[0].Type != model.ErrorTypeEnumAuthorization || userClient.softDeleteCalled {
		t.Fatalf("expected AUTHORIZATION error blocking self-SoftDeleteUser, got payload=%+v err=%v rpc=%v", delPayload, err, userClient.softDeleteCalled)
	}

	revPayload, err := adp.RevokeUserRole(authCtx, EncodeGID(ResourceUserRole, otherUUID))
	if err != nil || revPayload.WasSuccessful || len(revPayload.Errors) != 1 || revPayload.Errors[0].Type != model.ErrorTypeEnumAuthorization || roleClient.revokeCalled {
		t.Fatalf("expected AUTHORIZATION error blocking non-superadmin RevokeUserRole IDOR, got payload=%+v err=%v rpc=%v", revPayload, err, roleClient.revokeCalled)
	}
}
