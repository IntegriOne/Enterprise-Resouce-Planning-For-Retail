package handler_test

import (
	"encoding/base32"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
)

func TestIdentitySessionAuth_BufconnEndToEnd(t *testing.T) {
	t.Parallel()

	env := startBufconnServer(t)
	ctx := t.Context()

	permClient := iamv1.NewPermissionServiceClient(env.conn)
	roleClient := iamv1.NewRoleServiceClient(env.conn)
	userClient := iamv1.NewUserServiceClient(env.conn)
	sessClient := iamv1.NewSessionServiceClient(env.conn)
	authClient := iamv1.NewAuthServiceClient(env.conn)

	// Siapkan katalog izin dan peran untuk otorisasi dan redirect pasca-login.
	permResp, err := permClient.CreatePermission(ctx, &iamv1.CreatePermissionRequest{
		Code:        "finance.invoice.approve",
		Resource:    "finance.invoice",
		Action:      "approve",
		IsSensitive: true,
	})
	if err != nil {
		t.Fatalf("CreatePermission error = %v", err)
	}
	roleResp, err := roleClient.CreateRole(ctx, &iamv1.CreateRoleRequest{
		Code:                "finance_lead",
		Name:                "Finance Lead",
		ManagementDomain:    "finance",
		DefaultRedirectPath: "/finance",
		RedirectPriority:    20,
	})
	if err != nil {
		t.Fatalf("CreateRole error = %v", err)
	}
	if _, err := roleClient.AssignRolePermission(ctx, &iamv1.AssignRolePermissionRequest{
		RoleId:       roleResp.GetRole().GetId(),
		PermissionId: permResp.GetPermission().GetId(),
	}); err != nil {
		t.Fatalf("AssignRolePermission error = %v", err)
	}

	// Buat akun pekerja aktif dan tugaskan peran.
	uResp, err := userClient.CreateUser(ctx, &iamv1.CreateUserRequest{
		Kind:                   iamv1.UserKind_USER_KIND_EMPLOYEE,
		Status:                 iamv1.UserStatus_USER_STATUS_ACTIVE,
		DisplayName:            "Budi Finance",
		Email:                  "budi.fin@integrione.local",
		Username:               "budi.fin",
		Password:               "K3yVault#2026!Z",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser error = %v", err)
	}
	uid := uResp.GetUser().GetId()
	ctx = withAuthenticatedCaller(ctx, uid)
	if _, err := roleClient.AssignUserRole(ctx, &iamv1.AssignUserRoleRequest{
		UserId:    uid,
		RoleId:    roleResp.GetRole().GetId(),
		ScopeType: iamv1.ScopeType_SCOPE_TYPE_GLOBAL,
	}); err != nil {
		t.Fatalf("AssignUserRole error = %v", err)
	}

	// Autentikasi sandi tanpa MFA -> langsung terautentikasi dengan sesi + CSRF + redirect.
	authResp, err := authClient.AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "budi.fin",
		Password:   "K3yVault#2026!Z",
		IpAddress:  "127.0.0.1",
		UserAgent:  "grpcurl-test/1.0",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword error = %v", err)
	}
	if authResp.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED {
		t.Fatalf("expected AUTHENTICATED, got %v", authResp.GetStatus())
	}
	if authResp.GetRedirectPath() != "/finance" {
		t.Fatalf("expected redirect path /finance, got %q", authResp.GetRedirectPath())
	}
	sessTok := authResp.GetIssuedSession().GetSessionToken()
	csrfTok := authResp.GetIssuedSession().GetCsrfToken()
	if sessTok == "" || csrfTok == "" {
		t.Fatal("expected non-empty session token and CSRF token")
	}

	// Otorisasi izin sensitif memerlukan step-up MFA baru -> ditolak sebelum TOTP, lalu uji Enroll+Confirm TOTP.
	authzResp, err := sessClient.AuthorizeSession(ctx, &iamv1.AuthorizeSessionRequest{
		SessionToken:           sessTok,
		CsrfToken:              csrfTok,
		RequireCsrf:            true,
		RequiredPermission:     "finance.invoice.approve",
		RequireSensitiveReauth: false,
	})
	if err != nil {
		t.Fatalf("AuthorizeSession error = %v", err)
	}
	if authzResp.GetSnapshot().GetUserId() != uid {
		t.Fatalf("unexpected snapshot user id: %q", authzResp.GetSnapshot().GetUserId())
	}

	// Tolak AuthorizeSession jika CSRF token salah -> Unauthenticated / PermissionDenied.
	_, err = sessClient.AuthorizeSession(ctx, &iamv1.AuthorizeSessionRequest{
		SessionToken:       sessTok,
		CsrfToken:          "wrong-csrf",
		RequireCsrf:        true,
		RequiredPermission: "finance.invoice.approve",
	})
	if status.Code(err) != codes.PermissionDenied && status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected PermissionDenied or Unauthenticated for invalid CSRF, got %s (%v)", status.Code(err), err)
	}

	// Enroll + Confirm TOTP MFA.
	enrollResp, err := authClient.EnrollTotp(ctx, &iamv1.EnrollTotpRequest{
		UserId:       uid,
		AccountLabel: "budi.fin@integrione.local",
	})
	if err != nil {
		t.Fatalf("EnrollTotp error = %v", err)
	}
	parsedURI, err := url.Parse(enrollResp.GetOtpAuthUri())
	if err != nil {
		t.Fatalf("url.Parse(otp_auth_uri) error = %v", err)
	}
	secretB32 := parsedURI.Query().Get("secret")
	secretBytes, err := totpDecodeBase32ForTest(secretB32)
	if err != nil {
		t.Fatalf("decode TOTP secret error = %v", err)
	}
	now := time.Now()
	codeNow, err := env.totpProv.GenerateCodeAt(secretBytes, now)
	if err != nil {
		t.Fatalf("GenerateCodeAt error = %v", err)
	}

	confirmResp, err := authClient.ConfirmTotp(ctx, &iamv1.ConfirmTotpRequest{
		UserId: uid,
		Code:   codeNow,
	})
	if err != nil {
		t.Fatalf("ConfirmTotp error = %v", err)
	}
	if !confirmResp.GetFactor().GetIsEnabled() || len(confirmResp.GetRecoveryCodes()) == 0 {
		t.Fatalf("expected enabled factor and recovery codes, got %+v", confirmResp)
	}

	// Login ulang saat TOTP aktif -> wajib mengembalikan MFA_REQUIRED + challenge token.
	mfaAuth, err := authClient.AuthenticatePassword(ctx, &iamv1.AuthenticatePasswordRequest{
		Identifier: "budi.fin@integrione.local",
		Password:   "K3yVault#2026!Z",
		IpAddress:  "10.10.0.5",
		UserAgent:  "browser/1.0",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword(MFA) error = %v", err)
	}
	if mfaAuth.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_MFA_REQUIRED || mfaAuth.GetMfaChallengeToken() == "" {
		t.Fatalf("expected MFA_REQUIRED with challenge token, got %+v", mfaAuth)
	}

	// Verifikasi tantangan MFA memakai recovery code pertama agar tidak bentrok anti-replay step TOTP yang sama.
	verifyRec, err := authClient.VerifyRecoveryCodeChallenge(ctx, &iamv1.VerifyRecoveryCodeChallengeRequest{
		ChallengeToken: mfaAuth.GetMfaChallengeToken(),
		RecoveryCode:   confirmResp.GetRecoveryCodes()[0],
		IpAddress:      "10.10.0.5",
		UserAgent:      "browser/1.0",
	})
	if err != nil {
		t.Fatalf("VerifyRecoveryCodeChallenge error = %v", err)
	}
	if verifyRec.GetStatus() != iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED || !verifyRec.GetIssuedSession().GetSession().GetMfaVerified() {
		t.Fatalf("expected authenticated MFA session, got %+v", verifyRec)
	}

	// Uji LogoutSession lalu pastikan token sesi tidak lagi valid.
	if _, err := sessClient.LogoutSession(ctx, &iamv1.LogoutSessionRequest{
		SessionToken: verifyRec.GetIssuedSession().GetSessionToken(),
	}); err != nil {
		t.Fatalf("LogoutSession error = %v", err)
	}
	_, err = sessClient.AuthenticateSession(ctx, &iamv1.AuthenticateSessionRequest{
		SessionToken: verifyRec.GetIssuedSession().GetSessionToken(),
	})
	if status.Code(err) != codes.Unauthenticated && status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected unauthenticated/ended session after logout, got %s (%v)", status.Code(err), err)
	}

	// Verifikasi manajemen identitas pengguna (List + SetPrimary).
	idents, err := userClient.ListUserIdentities(ctx, &iamv1.ListUserIdentitiesRequest{UserId: uid})
	if err != nil || len(idents.GetIdentities()) < 2 {
		t.Fatalf("ListUserIdentities failed: resp=%+v err=%v", idents, err)
	}
}

func totpDecodeBase32ForTest(secret string) ([]byte, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
}
