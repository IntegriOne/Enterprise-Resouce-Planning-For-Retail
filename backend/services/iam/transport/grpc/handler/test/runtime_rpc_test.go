package handler_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
	"github.com/integrione/erp-retail/backend/services/iam/services/permission"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
	iamgrpc "github.com/integrione/erp-retail/backend/services/iam/transport/grpc"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

func buildHandlerTestRegistry(t *testing.T) (*services.Registry, *memStore) {
	t.Helper()

	store := newMemStore()
	tx := &memTransactor{}
	clock := domainservice.SystemClock{}
	idGen := crypto.NewUUIDGenerator()
	hasher := crypto.NewDefaultArgon2idHasher()
	tokenGen := crypto.NewOpaqueTokenGenerator()
	cipher, err := crypto.NewEphemeralAES256Cipher()
	if err != nil {
		t.Fatalf("NewEphemeralAES256Cipher() error = %v", err)
	}
	csrfDeriver, err := crypto.NewHMACCSRFDeriver(1, map[int][]byte{
		1: []byte("0123456789abcdef0123456789abcdef"),
	}, false)
	if err != nil {
		t.Fatalf("NewHMACCSRFDeriver() error = %v", err)
	}
	totpProv := totp.NewDefaultProvider()
	redirect := domainservice.NewDefaultRedirectResolver()
	sessPolicy := domainservice.NewDefaultSessionPolicy()
	evaluator := domainservice.NewAccessEvaluator(sessPolicy)

	permSvc, err := permission.NewService(&memPermRepo{store}, clock, idGen)
	if err != nil {
		t.Fatalf("permission.NewService() error = %v", err)
	}
	roleSvc, err := role.NewService(role.Dependencies{
		Transactor:      tx,
		Roles:           &memRoleRepo{store},
		RolePermissions: &memRolePermRepo{store},
		Permissions:     &memPermRepo{store},
		Users:           &memUserRepo{store},
		UserRoles:       &memUserRoleRepo{store},
		Clock:           clock,
		IDGen:           idGen,
		Redirect:        redirect,
	})
	if err != nil {
		t.Fatalf("role.NewService() error = %v", err)
	}
	groupSvc, err := group.NewService(group.Dependencies{
		Transactor:   tx,
		Groups:       &memGroupRepo{store},
		GroupMembers: &memGroupMemberRepo{store},
		GroupRoles:   &memGroupRoleRepo{store},
		Roles:        &memRoleRepo{store},
		Users:        &memUserRepo{store},
		Clock:        clock,
		IDGen:        idGen,
	})
	if err != nil {
		t.Fatalf("group.NewService() error = %v", err)
	}
	sessSvc, err := session.NewService(session.Dependencies{
		Sessions:        &memSessionRepo{store},
		Users:           &memUserRepo{store},
		Clock:           clock,
		IDGen:           idGen,
		Tokens:          tokenGen,
		CSRFDeriver:     csrfDeriver,
		SessionPolicy:   sessPolicy,
		AccessEvaluator: evaluator,
		Redirect:        redirect,
	})
	if err != nil {
		t.Fatalf("session.NewService() error = %v", err)
	}
	userSvc, err := user.NewService(user.Dependencies{
		Transactor:     tx,
		Users:          &memUserRepo{store},
		Identities:     &memIdentityRepo{store},
		Credentials:    &memCredentialRepo{store},
		TOTPFactors:    &memTOTPRepo{store},
		RecoveryCodes:  &memRecoveryCodeRepo{store},
		Tokens:         &memTokenRepo{store},
		Roles:          &memRoleRepo{store},
		UserRoles:      &memUserRoleRepo{store},
		Sessions:       sessSvc,
		Clock:          clock,
		IDGen:          idGen,
		PasswordHasher: hasher,
		PasswordPolicy: domainservice.NewDefaultPasswordPolicy(),
		TokenGen:       tokenGen,
		SecretCipher:   cipher,
		TOTPProvider:   totpProv,
		Redirect:       redirect,
	})
	if err != nil {
		t.Fatalf("user.NewService() error = %v", err)
	}

	return &services.Registry{
		Permissions: permSvc,
		Roles:       roleSvc,
		Groups:      groupSvc,
		Sessions:    sessSvc,
		Users:       userSvc,
	}, store
}

func TestRuntimeRPCs_RegisterVerifyCSRFAndInternalProof(t *testing.T) {
	t.Parallel()

	secret := "0123456789abcdef0123456789abcdef"
	lis := bufconn.Listen(1024 * 1024)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg, _ := buildHandlerTestRegistry(t)
	srv := iamgrpc.NewServer(&config.Config{
		AppEnv:                    "production",
		ServiceName:               "integrione-iam-test",
		InternalGatewayHMACSecret: secret,
	}, logger)
	if err := handler.RegisterAll(srv.GRPCServer(), reg, logger); err != nil {
		t.Fatalf("RegisterAll error = %v", err)
	}
	go func() { _ = srv.GRPCServer().Serve(lis) }()
	t.Cleanup(func() { srv.GRPCServer().GracefulStop() })

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	userClient := iamv1.NewUserServiceClient(conn)
	sessClient := iamv1.NewSessionServiceClient(conn)
	authClient := iamv1.NewAuthServiceClient(conn)

	// 1. Panggilan tanpa proof internal di mode production wajib ditolak fail-closed.
	if _, err := userClient.CreateUser(t.Context(), &iamv1.CreateUserRequest{
		Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
		Status:      iamv1.UserStatus_USER_STATUS_PENDING,
		DisplayName: "Nobody",
		Email:       "nobody@integrione.local",
		Username:    "nobody01",
		Password:    "CustPass#2026!Z",
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied without gateway proof, got %v", err)
	}

	// 2. Panggilan dengan proof HMAC sah dari Gateway diterima.
	signedCtx := authcontext.AppendToOutgoingGRPCContext(t.Context(), authcontext.CallerContext{
		RequestID: "req-test-1",
		ClientIP:  "127.0.0.1",
		UserAgent: "test-agent",
	}, []byte(secret), time.Now())

	createResp, err := userClient.CreateUser(signedCtx, &iamv1.CreateUserRequest{
		Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
		Status:      iamv1.UserStatus_USER_STATUS_PENDING,
		Email:       "buyer@integrione.local",
		Username:    "buyer01",
		Password:    "CustPass#2026!Z",
		DisplayName: "Buyer One",
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if createResp.GetUser().GetStatus() != iamv1.UserStatus_USER_STATUS_PENDING {
		t.Fatalf("unexpected CreateUser response: %+v", createResp)
	}
	firstTokResp, err := userClient.ResendValidationEmailToken(signedCtx, &iamv1.ResendValidationEmailTokenRequest{
		Email: "buyer@integrione.local",
	})
	if err != nil || firstTokResp.GetVerificationToken() == "" {
		t.Fatalf("ResendValidationEmailToken() error = (%+v, %v)", firstTokResp, err)
	}

	verResp, err := userClient.ConfirmEmailVerificationByToken(signedCtx, &iamv1.ConfirmEmailVerificationByTokenRequest{
		RawToken: firstTokResp.GetVerificationToken(),
	})
	if err != nil || !verResp.GetVerified() || verResp.GetStatus() != iamv1.VerifyEmailStatus_VERIFY_EMAIL_STATUS_VERIFIED {
		t.Fatalf("ConfirmEmailVerificationByToken error = (%+v, %v)", verResp, err)
	}

	// Login memakai username dan cek RefreshUserAccessSnapshot + GetSessionCsrf.
	authResp, err := authClient.AuthenticatePassword(signedCtx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "buyer01",
		Password:   "CustPass#2026!Z",
		IpAddress:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword error = %v", err)
	}
	sid := authResp.GetIssuedSession().GetSession().GetId()
	csrfFromLogin := authResp.GetIssuedSession().GetCsrfToken()

	csrfResp, err := sessClient.GetSessionCsrf(signedCtx, &iamv1.GetSessionCsrfRequest{SessionId: sid})
	if err != nil || csrfResp.GetCsrfToken() != csrfFromLogin {
		t.Fatalf("GetSessionCsrf mismatch: got (%q, %v), want %q", csrfResp.GetCsrfToken(), err, csrfFromLogin)
	}

	snapResp, err := sessClient.RefreshUserAccessSnapshot(signedCtx, &iamv1.RefreshUserAccessSnapshotRequest{
		UserId: createResp.GetUser().GetId(),
	})
	if err != nil || snapResp.GetCanAccessIdentity() || snapResp.GetDefaultRedirectPath() != "/" {
		t.Fatalf("unexpected customer RefreshUserAccessSnapshot: (%+v, %v)", snapResp, err)
	}

	// Reset password hanya menerima email terverifikasi (username ditolak InvalidArgument).
	if _, err := authClient.RequestPasswordReset(signedCtx, &iamv1.RequestPasswordResetRequest{
		Identifier: "buyer01",
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for username password reset, got %v", err)
	}
	resetResp, err := authClient.RequestPasswordReset(signedCtx, &iamv1.RequestPasswordResetRequest{
		Email: "buyer@integrione.local",
	})
	if err != nil || resetResp.GetIssuedToken().GetPlaintextToken() == "" {
		t.Fatalf("RequestPasswordReset(email) error = (%+v, %v)", resetResp, err)
	}

	// Uji kirim ulang token pada akun pending kedua.
	reg2Resp, err := userClient.CreateUser(signedCtx, &iamv1.CreateUserRequest{
		Kind:        iamv1.UserKind_USER_KIND_CUSTOMER,
		Status:      iamv1.UserStatus_USER_STATUS_PENDING,
		Email:       "buyer2@integrione.local",
		Username:    "buyer02",
		Password:    "CustPass#2026!Z",
		DisplayName: "Buyer Two",
	})
	if err != nil || reg2Resp.GetUser().GetStatus() != iamv1.UserStatus_USER_STATUS_PENDING {
		t.Fatalf("CreateUser(2) error = (%+v, %v)", reg2Resp, err)
	}
	resend1, err := userClient.ResendValidationEmailToken(signedCtx, &iamv1.ResendValidationEmailTokenRequest{
		Email: "buyer2@integrione.local",
	})
	if err != nil || !resend1.GetIsResend() || resend1.GetVerificationToken() == "" {
		t.Fatalf("ResendValidationEmailToken(1) unexpected: (%+v, %v)", resend1, err)
	}
	resendResp, err := userClient.ResendValidationEmailToken(signedCtx, &iamv1.ResendValidationEmailTokenRequest{
		Email: "buyer2@integrione.local",
	})
	if err != nil || !resendResp.GetIsResend() || resendResp.GetVerificationToken() == "" || resendResp.GetVerificationToken() == resend1.GetVerificationToken() {
		t.Fatalf("ResendValidationEmailToken(2) unexpected: (%+v, %v)", resendResp, err)
	}

	_ = bytes.Repeat([]byte{0}, 1)
}
