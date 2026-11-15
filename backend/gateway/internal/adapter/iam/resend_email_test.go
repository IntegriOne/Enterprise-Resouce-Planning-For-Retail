package iam

import (
	"context"
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

// TestResendValidationEmailToken_Case1AlreadyVerifiedAndCase2Neutral menguji propagasi error email sudah terverifikasi pada sesi login serta netralitas anti-enumerasi untuk pengguna anonim.
func TestResendValidationEmailToken_Case1AlreadyVerifiedAndCase2Neutral(t *testing.T) {
	now := time.Now()
	userUUID := "44444444-4444-4444-8444-444444444444"
	store := newMemStore()
	rawTok := "resend-verified-tok"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:    "sess-resend",
			UserID:       userUUID,
			TokenHashHex: tokHash,
			UserKind:     "employee",
			AuthEpoch:    1,
			IssuedAtUnix: now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{
			UserID:        userUUID,
			Status:        authz.StatusActive,
			AuthEpoch:     1,
			PermRevision:  1,
			RecoveryEpoch: 1,
		},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:   userUUID,
			UserKind:      "employee",
			Status:        authz.StatusActive,
			AuthEpoch:     1,
			PermRevision:  1,
			RecoveryEpoch: 1,
			Permissions: []authz.PermissionGrant{
				{Code: authz.PermUsersUpdateSelf, Scope: authz.ScopeGrant{Type: authz.ScopeSelf, ID: userUUID}},
			},
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		},
	}

	usersMock := &mockUserClient{
		resendEmailFn: func(_ context.Context, _ *iamv1.ResendValidationEmailTokenRequest) (*iamv1.ResendValidationEmailTokenResponse, error) {
			st, _ := status.New(codes.FailedPrecondition, "your email has been verified").WithDetails(&errdetails.ErrorInfo{
				Reason: "iam.identity.email_already_verified",
			})
			return nil, st.Err()
		},
	}

	authMgr, err := middleware.NewAuthManager(
		store,
		middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second),
		middleware.SessionConfig{CookieSecure: true},
	)
	if err != nil {
		t.Fatalf("NewAuthManager() error = %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(source_connections.IAMClientSet{
		Auth:     &mockAuthClient{},
		Users:    usersMock,
		Sessions: &mockSessionClient{},
	}, authMgr, authzSvc, rv, Config{LoginRateMax: 5, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	authCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
	})

	// Kasus (1) terautentikasi saat tidak ada pending email dan email sudah terverifikasi wajib mengembalikan CONFLICT + "your email has been verified".
	authPayload, err := adp.ResendValidationEmailToken(authCtx, &model.ResendValidationEmailInput{})
	if err != nil {
		t.Fatalf("authenticated ResendValidationEmailToken() unexpected transport error = %v", err)
	}
	if authPayload.WasSuccessful || len(authPayload.Errors) != 1 {
		t.Fatalf("authenticated payload expected WasSuccessful=false with 1 error, got %+v", authPayload)
	}
	if authPayload.Errors[0].Type != model.ErrorTypeEnumConflict {
		t.Fatalf("authenticated error type = %q, want %q", authPayload.Errors[0].Type, model.ErrorTypeEnumConflict)
	}
	if authPayload.Errors[0].Message != "your email has been verified" {
		t.Fatalf("authenticated error message = %q, want %q", authPayload.Errors[0].Message, "your email has been verified")
	}

	// Kasus (2) anonim harus tetap netral (IsResend=true, WasSuccessful=true) agar tidak membocorkan status email terverifikasi ke penyerang.
	anonEmail := "verified@integrione.local"
	anonPayload, err := adp.ResendValidationEmailToken(context.Background(), &model.ResendValidationEmailInput{
		Email: &anonEmail,
	})
	if err != nil {
		t.Fatalf("unauthenticated ResendValidationEmailToken() unexpected transport error = %v", err)
	}
	if !anonPayload.WasSuccessful || len(anonPayload.Errors) != 0 || anonPayload.IsResend == nil || !*anonPayload.IsResend {
		t.Fatalf("unauthenticated payload expected neutral IsResend=true WasSuccessful=true, got %+v", anonPayload)
	}
}
