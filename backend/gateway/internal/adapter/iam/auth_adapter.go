package iam

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// BrowserSession menginisialisasi atau menyelaraskan sesi browser dan header X-CSRF-Token untuk halaman /login, /register, maupun dashboard.
func (a *Adapter) BrowserSession(ctx context.Context) (*model.BrowserSessionPayload, error) {
	if err := a.ensureAuthDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authManager.AuthenticateOptional(ctx)
	if err != nil {
		return nil, err
	}
	if sess != nil && sess.Token != nil {
		resp, csrfErr := callIAMSessions(ctx, a, &sess.Caller, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.GetSessionCsrfResponse, error) {
			return sessions.GetSessionCsrf(rpcCtx, &iamv1.GetSessionCsrfRequest{
				SessionId: sess.Token.SessionID,
			})
		})
		if csrfErr == nil && resp != nil && resp.GetCsrfToken() != "" {
			csrf := resp.GetCsrfToken()
			if eff := browserauth.EffectsFromContext(ctx); eff != nil {
				eff.SetHeader(browserauth.CSRFHeaderName, csrf)
			}
			return &model.BrowserSessionPayload{
				Authenticated: true,
				PrincipalID:   ptrStr(sess.Token.UserID),
				CsrfToken:     ptrStr(csrf),
			}, nil
		}
		var mappedErr error
		if csrfErr == nil {
			mappedErr = middleware.ErrUnavailable("session csrf lookup returned empty token", nil)
		} else {
			mappedErr = mapRPCError(csrfErr)
		}
		if gwErr, ok := middleware.AsGatewayError(mappedErr); ok && (gwErr.Code == middleware.CodeUnauthenticated || gwErr.Code == middleware.CodeNotFound) {
			a.authManager.ClearSessionCookie(ctx)
		} else {
			return nil, mappedErr
		}
	}

	if err := a.checkPreAuthRateLimit(ctx); err != nil {
		return nil, err
	}
	rawCSRF, preErr := a.authManager.IssuePreAuthSession(ctx)
	if preErr != nil {
		return nil, preErr
	}
	return &model.BrowserSessionPayload{
		Authenticated: false,
		PrincipalID:   nil,
		CsrfToken:     ptrStr(rawCSRF),
	}, nil
}

// MyWorkspaceAccess memeriksa kelayakan akses dashboard /identity dan menentukan rute redirect default secara fail-closed.
func (a *Adapter) MyWorkspaceAccess(ctx context.Context) (*model.WorkspaceAccessPayload, error) {
	if err := a.ensureAuthFlowDeps(); err != nil {
		return nil, err
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	canAccess := sess.Snapshot.CanAccessWorkspaceIdentity()
	defaultLoc := a.redirectValidator.ResolveSafeLocation(sess.Snapshot.ResolveDefaultLocation())
	return &model.WorkspaceAccessPayload{
		CanAccessIdentity: canAccess,
		DefaultLocation:   defaultLoc,
	}, nil
}

const neutralPasswordResetMessage = "if your email is registered and verified, password reset instructions have been sent"

func ptrUserLoginStatus(v model.UserLoginStatus) *model.UserLoginStatus {
	return &v
}

func ptrVerifyEmailStatus(v model.VerifyEmailStatus) *model.VerifyEmailStatus {
	return &v
}

// UserLogin memverifikasi kredensial sandi via IAM gRPC, menerbitkan cookie __Host- sesi atau tantangan MFA, serta header Location aman.
func (a *Adapter) UserLogin(ctx context.Context, input model.UserLoginInput) (*model.UserLoginPayload, error) {
	fail := func(err error) (*model.UserLoginPayload, error) {
		return &model.UserLoginPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthFlowDeps(); err != nil {
		return fail(err)
	}
	if err := a.checkAuthRateLimit(ctx, input.Identifier); err != nil {
		return fail(err)
	}
	meta, _ := middleware.RequestMetadataFromContext(ctx)
	caller := a.authManager.BuildCallerContext(ctx)

	resp, err := callIAMAuth(ctx, a, &caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.AuthenticatePasswordResponse, error) {
		return auth.AuthenticatePassword(rpcCtx, &iamv1.AuthenticatePasswordRequest{
			Identifier: strings.TrimSpace(input.Identifier),
			Password:   input.Password,
			IpAddress:  meta.ClientIP,
			UserAgent:  meta.UserAgent,
		})
	})
	if err != nil {
		return fail(normalizeAuthError(err, "invalid login credentials"))
	}

	if resp.GetStatus() == iamv1.AuthStatus_AUTH_STATUS_MFA_REQUIRED {
		var expiresAt *string
		if ts := resp.GetMfaChallengeExpiresAt(); ts != nil && ts.IsValid() {
			formatted := ts.AsTime().Local().Format(time.RFC3339)
			expiresAt = &formatted
		}
		challenge := resp.GetMfaChallengeToken()
		return &model.UserLoginPayload{
			Status:        ptrUserLoginStatus(model.UserLoginStatusMfaRequired),
			MfaChallenge:  &challenge,
			MfaExpiresAt:  expiresAt,
			Errors:        noMutationErrors(),
			WasSuccessful: true,
		}, nil
	}

	issued := resp.GetIssuedSession()
	if issued == nil || issued.GetSessionToken() == "" {
		return fail(middleware.ErrUnavailable("login succeeded without issued session", nil))
	}
	redirectLoc := a.redirectValidator.ResolveSafeLocation(resp.GetRedirectPath())
	maxAge := resolveSessionCookieMaxAge(issued.GetSession(), a.cfg.SessionMaxAge)
	a.authManager.PromoteAuthenticatedCookie(ctx, issued.GetSessionToken(), issued.GetCsrfToken(), redirectLoc, maxAge)

	return &model.UserLoginPayload{
		Status:           ptrUserLoginStatus(model.UserLoginStatusAuthenticated),
		CsrfToken:        ptrStr(issued.GetCsrfToken()),
		RedirectLocation: ptrStr(redirectLoc),
		Errors:           noMutationErrors(),
		WasSuccessful:    true,
	}, nil
}

// ValidateUserMfa menuntaskan tantangan MFA TOTP atau recovery code lalu merotasi cookie menjadi sesi terautentikasi penuh.
func (a *Adapter) ValidateUserMfa(ctx context.Context, input model.ValidateUserMfaInput) (*model.ValidateUserMfaPayload, error) {
	fail := func(err error) (*model.ValidateUserMfaPayload, error) {
		return &model.ValidateUserMfaPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthFlowDeps(); err != nil {
		return fail(err)
	}
	if err := a.checkAuthRateLimit(ctx, ""); err != nil {
		return fail(err)
	}
	challenge := strings.TrimSpace(input.Challenge)
	if err := a.checkMFAChallengeRateLimit(ctx, challenge); err != nil {
		return fail(err)
	}
	totpCode := derefStr(input.TotpCode)
	recoveryCode := derefStr(input.RecoveryCode)
	if challenge == "" || (totpCode == "" && recoveryCode == "") {
		return fail(middleware.ErrInvalidArgument("challenge and either totpCode or recoveryCode are required", nil))
	}

	meta, _ := middleware.RequestMetadataFromContext(ctx)
	caller := a.authManager.BuildCallerContext(ctx)

	var (
		issued      *iamv1.IssuedSession
		rawRedirect string
	)
	if totpCode != "" {
		resp, err := callIAMAuth(ctx, a, &caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.VerifyMfaChallengeResponse, error) {
			return auth.VerifyMfaChallenge(rpcCtx, &iamv1.VerifyMfaChallengeRequest{
				ChallengeToken: challenge,
				Code:           totpCode,
				IpAddress:      meta.ClientIP,
				UserAgent:      meta.UserAgent,
			})
		})
		if err != nil {
			return fail(normalizeAuthError(err, "invalid mfa verification"))
		}
		issued = resp.GetIssuedSession()
		rawRedirect = resp.GetRedirectPath()
	} else {
		resp, err := callIAMAuth(ctx, a, &caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.VerifyRecoveryCodeChallengeResponse, error) {
			return auth.VerifyRecoveryCodeChallenge(rpcCtx, &iamv1.VerifyRecoveryCodeChallengeRequest{
				ChallengeToken: challenge,
				RecoveryCode:   recoveryCode,
				IpAddress:      meta.ClientIP,
				UserAgent:      meta.UserAgent,
			})
		})
		if err != nil {
			return fail(normalizeAuthError(err, "invalid mfa verification"))
		}
		issued = resp.GetIssuedSession()
		rawRedirect = resp.GetRedirectPath()
	}

	if issued == nil || issued.GetSessionToken() == "" {
		return fail(middleware.ErrUnavailable("mfa verification succeeded without issued session", nil))
	}
	redirectLoc := a.redirectValidator.ResolveSafeLocation(rawRedirect)
	maxAge := resolveSessionCookieMaxAge(issued.GetSession(), a.cfg.SessionMaxAge)
	a.authManager.PromoteAuthenticatedCookie(ctx, issued.GetSessionToken(), issued.GetCsrfToken(), redirectLoc, maxAge)

	return &model.ValidateUserMfaPayload{
		CsrfToken:        ptrStr(issued.GetCsrfToken()),
		RedirectLocation: ptrStr(redirectLoc),
		Errors:           noMutationErrors(),
		WasSuccessful:    true,
	}, nil
}

// VerifyEmail memvalidasi token verifikasi email sekali pakai baik untuk aktivasi registrasi maupun perubahan email.
func (a *Adapter) VerifyEmail(ctx context.Context, input model.VerifyEmailInput) (*model.VerifyEmailPayload, error) {
	fail := func(err error) (*model.VerifyEmailPayload, error) {
		return &model.VerifyEmailPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthDeps(); err != nil {
		return fail(err)
	}
	if err := a.checkAuthRateLimit(ctx, ""); err != nil {
		return fail(err)
	}
	caller := a.authManager.BuildCallerContext(ctx)
	resp, err := callIAMUsers(ctx, a, &caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ConfirmEmailVerificationByTokenResponse, error) {
		return users.ConfirmEmailVerificationByToken(rpcCtx, &iamv1.ConfirmEmailVerificationByTokenRequest{
			RawToken: strings.TrimSpace(input.Token),
		})
	})
	if err != nil {
		return fail(err)
	}
	return &model.VerifyEmailPayload{
		Status:        ptrVerifyEmailStatus(fromProtoVerifyEmailStatus(resp.GetStatus())),
		Verified:      ptrBool(resp.GetVerified()),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// ResendValidationEmailToken mengirim ulang token verifikasi email dan menandai token aktif sebelumnya menjadi expired.
func (a *Adapter) ResendValidationEmailToken(ctx context.Context, input *model.ResendValidationEmailInput) (*model.ResendValidationEmailPayload, error) {
	fail := func(err error) (*model.ResendValidationEmailPayload, error) {
		return &model.ResendValidationEmailPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthorizerDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authManager.AuthenticateOptional(ctx)
	if err != nil {
		return fail(err)
	}

	var email string
	if input != nil {
		email = derefStr(input.Email)
	}

	// Case 1: Pengguna terautentikasi mengirim ulang verifikasi untuk perubahan email pending miliknya.
	if sess != nil && sess.Token != nil {
		if _, pErr := a.authorizer.RequireSelfPermission(ctx, authz.PermUsersUpdateSelf, false); pErr != nil {
			return fail(pErr)
		}
		if rateErr := a.checkAuthRateLimit(ctx, sess.Token.UserID); rateErr != nil {
			return fail(rateErr)
		}
		resp, rpcErr := callIAMUsers(ctx, a, &sess.Caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ResendValidationEmailTokenResponse, error) {
			return users.ResendValidationEmailToken(rpcCtx, &iamv1.ResendValidationEmailTokenRequest{
				UserId: sess.Token.UserID,
				Email:  email,
			})
		})
		if rpcErr != nil {
			return fail(rpcErr)
		}
		return &model.ResendValidationEmailPayload{
			IsResend:      ptrBool(resp.GetIsResend()),
			Errors:        noMutationErrors(),
			WasSuccessful: true,
		}, nil
	}

	// Case 2: Pendaftar anonim berstatus PENDING mengirim ulang verifikasi memakai alamat email.
	if email == "" {
		return fail(middleware.ErrInvalidArgument("email is required", nil))
	}
	if rateErr := a.checkAuthRateLimit(ctx, email); rateErr != nil {
		return fail(rateErr)
	}
	caller := a.authManager.BuildCallerContext(ctx)
	resp, rpcErr := callIAMUsers(ctx, a, &caller, func(rpcCtx context.Context, users iamv1.UserServiceClient) (*iamv1.ResendValidationEmailTokenResponse, error) {
		return users.ResendValidationEmailToken(rpcCtx, &iamv1.ResendValidationEmailTokenRequest{
			Email: email,
		})
	})
	if rpcErr == nil {
		return &model.ResendValidationEmailPayload{
			IsResend:      ptrBool(resp.GetIsResend()),
			Errors:        noMutationErrors(),
			WasSuccessful: true,
		}, nil
	}

	mapped := mapRPCError(rpcErr)
	if gwErr, ok := middleware.AsGatewayError(mapped); ok {
		// Mencegah enumerasi status akun pada pemanggil anonim.
		switch gwErr.Code {
		case middleware.CodeInvalidArgument, middleware.CodeUnavailable, middleware.CodeTooManyRequests:
			return fail(mapped)
		case middleware.CodeNotFound, middleware.CodeForbidden, middleware.CodeUnprocessableEntity, middleware.CodeConflict:
			return &model.ResendValidationEmailPayload{
				IsResend:      ptrBool(true),
				Errors:        noMutationErrors(),
				WasSuccessful: true,
			}, nil
		}
	}
	return fail(mapped)
}

// UserResetPassword memicu penerbitan token reset sandi berbasis email terverifikasi tanpa membocorkan eksistensi akun.
func (a *Adapter) UserResetPassword(ctx context.Context, input model.UserResetPasswordInput) (*model.UserResetPasswordPayload, error) {
	fail := func(err error) (*model.UserResetPasswordPayload, error) {
		return &model.UserResetPasswordPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthDeps(); err != nil {
		return fail(err)
	}
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return fail(middleware.ErrInvalidArgument("email is required", nil))
	}
	if err := a.checkAuthRateLimit(ctx, email); err != nil {
		return fail(err)
	}

	caller := a.authManager.BuildCallerContext(ctx)
	_, err := callIAMAuth(ctx, a, &caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.RequestPasswordResetResponse, error) {
		return auth.RequestPasswordReset(rpcCtx, &iamv1.RequestPasswordResetRequest{
			Email: email,
		})
	})
	if err == nil {
		return &model.UserResetPasswordPayload{
			Message:       ptrStr(neutralPasswordResetMessage),
			Errors:        noMutationErrors(),
			WasSuccessful: true,
		}, nil
	}

	mapped := mapRPCError(err)
	if gwErr, ok := middleware.AsGatewayError(mapped); ok {
		// Mencegah enumerasi akun: apabila pengguna tidak ditemukan atau diblokir statusnya, respons tetap sukses generik.
		switch gwErr.Code {
		case middleware.CodeNotFound, middleware.CodeForbidden, middleware.CodeUnprocessableEntity, middleware.CodeConflict:
			return &model.UserResetPasswordPayload{
				Message:       ptrStr(neutralPasswordResetMessage),
				Errors:        noMutationErrors(),
				WasSuccessful: true,
			}, nil
		case middleware.CodeInvalidArgument, middleware.CodeUnavailable, middleware.CodeTooManyRequests:
			return fail(mapped)
		}
	}
	return fail(mapped)
}

// ResetPassword menuntaskan pengaturan ulang sandi dengan token pemulihan sekali pakai.
func (a *Adapter) ResetPassword(ctx context.Context, input model.IAMResetPasswordInput) (*model.ResetPasswordPayload, error) {
	fail := func(err error) (*model.ResetPasswordPayload, error) {
		return &model.ResetPasswordPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthDeps(); err != nil {
		return fail(err)
	}
	if err := a.checkAuthRateLimit(ctx, ""); err != nil {
		return fail(err)
	}
	caller := a.authManager.BuildCallerContext(ctx)
	_, err := callIAMAuth(ctx, a, &caller, func(rpcCtx context.Context, auth iamv1.AuthServiceClient) (*iamv1.ConfirmPasswordResetResponse, error) {
		return auth.ConfirmPasswordReset(rpcCtx, &iamv1.ConfirmPasswordResetRequest{
			RawToken:    strings.TrimSpace(input.Token),
			NewPassword: input.NewPassword,
		})
	})
	if err != nil {
		return fail(err)
	}
	return &model.ResetPasswordPayload{
		Reset:         ptrBool(true),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

// UserLogout mengakhiri sesi aktif di IAM & Redis serta menghapus cookie __Host- dari peramban.
func (a *Adapter) UserLogout(ctx context.Context) (*model.UserLogoutPayload, error) {
	fail := func(err error) (*model.UserLogoutPayload, error) {
		return &model.UserLogoutPayload{
			Errors:        ToMutationErrors(err),
			WasSuccessful: false,
		}, nil
	}
	if err := a.ensureAuthDeps(); err != nil {
		return fail(err)
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return fail(err)
	}
	_, rpcErr := callIAMSessions(ctx, a, &sess.Caller, func(rpcCtx context.Context, sessions iamv1.SessionServiceClient) (*iamv1.LogoutSessionResponse, error) {
		return sessions.LogoutSession(rpcCtx, &iamv1.LogoutSessionRequest{
			SessionToken: sess.RawToken,
		})
	})
	if rpcErr != nil {
		mapped := mapRPCError(rpcErr)
		if gwErr, ok := middleware.AsGatewayError(mapped); !ok || (gwErr.Code != middleware.CodeUnauthenticated && gwErr.Code != middleware.CodeNotFound) {
			return fail(mapped)
		}
	}
	a.authManager.ClearSessionCookie(ctx)
	if eff := browserauth.EffectsFromContext(ctx); eff != nil {
		eff.SetHeader(browserauth.CSRFHeaderName, "")
	}
	return &model.UserLogoutPayload{
		LoggedOut:     ptrBool(true),
		Errors:        noMutationErrors(),
		WasSuccessful: true,
	}, nil
}

func resolveSessionCookieMaxAge(s *iamv1.Session, fallback time.Duration) time.Duration {
	if s != nil && s.GetAbsoluteExpiresAt() != nil && s.GetAbsoluteExpiresAt().IsValid() {
		rem := time.Until(s.GetAbsoluteExpiresAt().AsTime().Local())
		if rem > 0 {
			return rem
		}
	}
	return fallback
}
