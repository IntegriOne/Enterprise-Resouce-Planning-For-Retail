package iam

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
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

func isUnavailableGatewayErr(err error) bool {
	gwErr, ok := middleware.AsGatewayError(err)
	return ok && gwErr.Code == middleware.CodeUnavailable
}

func newTestResilienceAdapter(t *testing.T, clients source_connections.IAMClientSet, checker middleware.SensitiveSessionChecker) (*Adapter, string) {
	t.Helper()
	now := time.Now()
	store := newMemStore()

	rawTok := "resilience-admin-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:        "sess-resilience",
			UserID:           "11111111-1111-4111-8111-111111111111",
			TokenHashHex:     tokHash,
			UserKind:         "employee",
			AuthEpoch:        1,
			IssuedAtUnix:     now.Unix(),
			MFAAuthenticated: true,
		},
		Security: &authsession.UserSecurityRecord{UserID: "11111111-1111-4111-8111-111111111111", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:         "11111111-1111-4111-8111-111111111111",
			UserKind:            "employee",
			Status:              authz.StatusActive,
			AuthEpoch:           1,
			PermRevision:        1,
			RecoveryEpoch:       1,
			CanAccessIdentity:   true,
			DefaultRedirectPath: "/identity",
			SyncedAtUnix:        now.Unix(),
			FreshUntilUnix:      now.Add(30 * time.Second).Unix(),
			Roles:               []authz.RoleGrant{{ID: "role-1", Code: "admin", Name: "Admin"}},
			Permissions: []authz.PermissionGrant{
				{Code: authz.PermUsersReadSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: "11111111-1111-4111-8111-111111111111"}},
				{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
				{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
				{Code: "iam.sessions.revoke", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	}

	authMgr, err := middleware.NewAuthManager(store, middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second), middleware.SessionConfig{CookieSecure: true})
	if err != nil {
		t.Fatalf("NewAuthManager: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, checker, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})
	adp, err := NewAdapter(clients, authMgr, authzSvc, rv, Config{LoginRateMax: 10, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}
	return adp, rawTok
}

func TestIAMAdapterResilience_UnavailableAndSanitizedErrors(t *testing.T) {
	t.Parallel()

	// 1. Registry terdegradasi (IAMTarget kosong) -> seluruh operasi IAM fail-closed dengan UNAVAILABLE tanpa panic.
	reg, err := source_connections.NewRegistry(source_connections.Config{
		IAMTarget:  "",
		IAMTimeout: time.Second,
		HMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	t.Cleanup(func() { _ = reg.Close() })

	bridge := source_connections.NewIAMRuntimeBridge(reg.IAM, time.Second)
	adp, rawTok := newTestResilienceAdapter(t, reg.IAM, bridge)
	authCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: authsession.HashOpaqueToken(rawTok),
		ClientIP:   "127.0.0.1",
	})

	if _, err := adp.Me(authCtx); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected Me to return UNAVAILABLE when IAM degraded, got %v", err)
	}
	if _, err := adp.BrowserSession(authCtx); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected authenticated BrowserSession to return UNAVAILABLE when IAM degraded, got %v", err)
	}
	if _, err := adp.Users(authCtx, nil, nil, nil, nil, nil, nil, nil, nil); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected Users to return UNAVAILABLE when IAM degraded, got %v", err)
	}

	loginRes, err := adp.UserLogin(context.Background(), model.UserLoginInput{Identifier: "a@integrione.local", Password: "secret"})
	if err != nil || loginRes.WasSuccessful || len(loginRes.Errors) != 1 || loginRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationError on UserLogin when IAM degraded, got payload=%+v err=%v", loginRes, err)
	}

	adp.workerProvisioner = &mockWorkerPrincipalProvisioner{}
	empRes, err := adp.CreateEmployeeUser(authCtx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: "55555555-5555-4555-8555-555555555555",
		FullName: "Tester",
		Email:    "tester@integrione.id",
		Username: "tester1",
		Password: "Password123!",
	})
	if err != nil || empRes.WasSuccessful || len(empRes.Errors) != 1 || empRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationError on CreateEmployeeUser when IAM degraded, got payload=%+v err=%v", empRes, err)
	}

	revRes, err := adp.RevokeSession(authCtx, EncodeGID(ResourceSession, "22222222-2222-4222-8222-222222222222"), nil)
	if err != nil || revRes.WasSuccessful || len(revRes.Errors) != 1 || revRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationError on RevokeSession when IAM degraded, got payload=%+v err=%v", revRes, err)
	}

	// 2. Stub mock mengembalikan gRPC Unavailable dengan rincian host/port internal -> pesan disanitasi ke klien.
	authMock := &mockAuthClient{
		authFn: func(context.Context, *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error) {
			return nil, status.Error(codes.Unavailable, "dial tcp iam:9001: connection refused")
		},
	}
	adpAuthFail, _ := newTestResilienceAdapter(t, source_connections.IAMClientSet{Auth: authMock}, nil)
	loginSanitized, err := adpAuthFail.UserLogin(context.Background(), model.UserLoginInput{Identifier: "a@integrione.local", Password: "secret"})
	if err != nil || loginSanitized.WasSuccessful || len(loginSanitized.Errors) != 1 || loginSanitized.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected sanitized UNAVAILABLE on UserLogin gRPC failure, got payload=%+v err=%v", loginSanitized, err)
	}
	if strings.Contains(strings.ToLower(loginSanitized.Errors[0].Message), "iam:9001") {
		t.Fatalf("expected iam:9001 to be stripped from UserLogin error message, got %q", loginSanitized.Errors[0].Message)
	}

	// 3. Error domain AlreadyExists tetap mempertahankan ExtractIAMErrorReason melalui callIAMUsers.
	usersMock := &mockUserClient{
		createFn: func(context.Context, *iamv1.CreateUserRequest) (*iamv1.CreateUserResponse, error) {
			st, _ := status.New(codes.AlreadyExists, "email exists").WithDetails(&errdetails.ErrorInfo{
				Reason:   "iam.identity.already_exists",
				Metadata: map[string]string{"field": "email"},
			})
			return nil, st.Err()
		},
	}
	adpDup, dupTok := newTestResilienceAdapter(t, source_connections.IAMClientSet{Users: usersMock}, nil)
	dupCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  dupTok,
		CookieHash: authsession.HashOpaqueToken(dupTok),
	})
	adpDup.clients.EmployeeProvisioning = mockEmployeeProvisioningClient{}
	adpDup.workerProvisioner = &mockWorkerPrincipalProvisioner{}
	dupRes, err := adpDup.CreateEmployeeUser(dupCtx, model.CreateEmployeeUserInput{
		ProvisioningRequestID: fixtureEmployeeProvisioningID, WorkerID: "55555555-5555-4555-8555-555555555555",
		FullName: "Dup",
		Email:    "dup@integrione.id",
		Username: "dup1",
		Password: "Password123!",
	})
	if err != nil || dupRes.WasSuccessful || len(dupRes.Errors) != 1 || dupRes.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("expected CONFLICT MutationError through callIAMUsers, got payload=%+v err=%v", dupRes, err)
	}
}

func TestIAMAdapterResilience_NilAdapterAndInternalS2SIsolation(t *testing.T) {
	t.Parallel()

	// 1. Pemanggilan method pada receiver *Adapter nil harus mengembalikan UNAVAILABLE tanpa panic.
	var nilAdp *Adapter
	if _, err := nilAdp.Me(context.Background()); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE from nil Adapter.Me, got %v", err)
	}
	if _, err := nilAdp.BrowserSession(context.Background()); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE from nil Adapter.BrowserSession, got %v", err)
	}
	if _, err := nilAdp.Node(context.Background(), "gid://integrione/user/11111111-1111-4111-8111-111111111111"); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE from nil Adapter.Node, got %v", err)
	}
	loginRes, err := nilAdp.UserLogin(context.Background(), model.UserLoginInput{Identifier: "x", Password: "y"})
	if err != nil || loginRes == nil || loginRes.WasSuccessful || len(loginRes.Errors) != 1 || loginRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationError from nil Adapter.UserLogin, got res=%+v err=%v", loginRes, err)
	}
	revRes, err := nilAdp.RevokeSession(context.Background(), "gid://integrione/session/11111111-1111-4111-8111-111111111111", nil)
	if err != nil || revRes == nil || revRes.WasSuccessful || len(revRes.Errors) != 1 || revRes.Errors[0].Type != model.ErrorTypeEnumUnavailable {
		t.Fatalf("expected UNAVAILABLE MutationError from nil Adapter.RevokeSession, got res=%+v err=%v", revRes, err)
	}

	// 2. Kegagalan konteks internal S2S pada GetSessionCsrf memetakan ke UNAVAILABLE tanpa menghapus cookie sesi peramban.
	sessMock := &mockSessionClient{
		getCSRFFn: func(context.Context, *iamv1.GetSessionCsrfRequest) (*iamv1.GetSessionCsrfResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "invalid internal caller context")
		},
	}
	adpS2SFail, rawTok := newTestResilienceAdapter(t, source_connections.IAMClientSet{Sessions: sessMock}, nil)
	eff := browserauth.NewResponseEffects()
	authCtx := browserauth.WithResponseEffects(middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: authsession.HashOpaqueToken(rawTok),
	}), eff)
	if _, err := adpS2SFail.BrowserSession(authCtx); !isUnavailableGatewayErr(err) {
		t.Fatalf("expected UNAVAILABLE when GetSessionCsrf fails internal caller check, got %v", err)
	}
	rec := httptest.NewRecorder()
	eff.ApplyTo(rec)
	if cookies := rec.Result().Cookies(); len(cookies) > 0 {
		t.Fatalf("expected browser session cookie NOT to be cleared on S2S internal error, got %+v", cookies)
	}

	// 3. normalizeAuthError membungkus cause gRPC sekali saja tanpa double-wrapping GatewayError.
	stSys, _ := status.New(codes.PermissionDenied, "system cannot login").WithDetails(
		&errdetails.ErrorInfo{Reason: "iam.user.system_cannot_login"},
	)
	mapped := mapRPCError(stSys.Err())
	norm := normalizeAuthError(mapped, "invalid login credentials")
	gwNorm, ok := middleware.AsGatewayError(norm)
	if !ok || gwNorm.Code != middleware.CodeForbidden || gwNorm.Message != "system cannot login" {
		t.Fatalf("expected system cannot login FORBIDDEN, got %+v", gwNorm)
	}
	if _, nestedGW := middleware.AsGatewayError(gwNorm.Err); nestedGW {
		t.Fatalf("expected cause of normalized auth error NOT to double-wrap GatewayError, got %T", gwNorm.Err)
	}
}
