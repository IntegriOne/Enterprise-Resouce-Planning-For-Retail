package handler

import (
	"context"
	"fmt"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

// AuthHandler menangani permintaan gRPC autentikasi sandi, MFA TOTP, kode pemulihan, dan manajemen kata sandi.
type AuthHandler struct {
	iamv1.UnimplementedAuthServiceServer
	svc       *user.Service
	errMapper *ErrorMapper
}

// NewAuthHandler membentuk handler gRPC untuk AuthService secara fail-closed.
func NewAuthHandler(svc *user.Service, errMapper *ErrorMapper) (*AuthHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("auth user service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &AuthHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// AuthenticatePassword memverifikasi email/username beserta kata sandi dan menerbitkan sesi atau tantangan MFA.
func (h *AuthHandler) AuthenticatePassword(ctx context.Context, req *iamv1.AuthenticatePasswordRequest) (*iamv1.AuthenticatePasswordResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{
		Identifier: req.GetIdentifier(),
		Password:   req.GetPassword(),
		IPAddress:  req.GetIpAddress(),
		UserAgent:  req.GetUserAgent(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AuthenticatePasswordResponse{
		Status:                toProtoAuthStatus(res.Status),
		User:                  toProtoUser(res.User),
		IssuedSession:         toProtoIssuedSession(res.IssuedSession),
		RedirectPath:          res.RedirectPath,
		MustChangePassword:    res.MustChangePassword,
		MfaChallengeToken:     res.MFAChallengeToken,
		MfaChallengeExpiresAt: toProtoTimestampPtr(res.MFAChallengeExpiresAt),
	}, nil
}

// VerifyMfaChallenge menyelesaikan tantangan MFA login dengan kode TOTP 6 digit.
func (h *AuthHandler) VerifyMfaChallenge(ctx context.Context, req *iamv1.VerifyMfaChallengeRequest) (*iamv1.VerifyMfaChallengeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.VerifyMFAChallenge(ctx, user.VerifyMFAChallengeInput{
		ChallengeToken: req.GetChallengeToken(),
		Code:           req.GetCode(),
		IPAddress:      req.GetIpAddress(),
		UserAgent:      req.GetUserAgent(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifyMfaChallengeResponse{
		Status:             toProtoAuthStatus(res.Status),
		User:               toProtoUser(res.User),
		IssuedSession:      toProtoIssuedSession(res.IssuedSession),
		RedirectPath:       res.RedirectPath,
		MustChangePassword: res.MustChangePassword,
	}, nil
}

// VerifyRecoveryCodeChallenge menyelesaikan tantangan MFA login menggunakan kode pemulihan sekali pakai.
func (h *AuthHandler) VerifyRecoveryCodeChallenge(ctx context.Context, req *iamv1.VerifyRecoveryCodeChallengeRequest) (*iamv1.VerifyRecoveryCodeChallengeResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.VerifyRecoveryCodeChallenge(ctx, user.VerifyRecoveryCodeInput{
		ChallengeToken: req.GetChallengeToken(),
		RecoveryCode:   req.GetRecoveryCode(),
		IPAddress:      req.GetIpAddress(),
		UserAgent:      req.GetUserAgent(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifyRecoveryCodeChallengeResponse{
		Status:             toProtoAuthStatus(res.Status),
		User:               toProtoUser(res.User),
		IssuedSession:      toProtoIssuedSession(res.IssuedSession),
		RedirectPath:       res.RedirectPath,
		MustChangePassword: res.MustChangePassword,
	}, nil
}

// VerifySessionMfa memvalidasi TOTP untuk step-up MFA sesi aktif sekaligus merotasi token CSRF.
func (h *AuthHandler) VerifySessionMfa(ctx context.Context, req *iamv1.VerifySessionMfaRequest) (*iamv1.VerifySessionMfaResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfSessionCaller(ctx, req.GetUserId(), req.GetSessionId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.VerifySessionMFA(ctx, req.GetUserId(), req.GetSessionId(), req.GetCode())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifySessionMfaResponse{
		Session:   toProtoSession(res.Session),
		CsrfToken: res.CSRFToken,
	}, nil
}

// ChangePassword memperbarui kata sandi pengguna terautentikasi lalu mencabut sesi lamanya.
func (h *AuthHandler) ChangePassword(ctx context.Context, req *iamv1.ChangePasswordRequest) (*iamv1.ChangePasswordResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.ChangePassword(ctx, user.ChangePasswordInput{
		UserID:          req.GetUserId(),
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ChangePasswordResponse{
		Changed:           res.Changed,
		SessionsRevoked:   res.SessionsRevoked,
		PasswordChangedAt: toProtoTimestamp(res.PasswordChangedAt),
	}, nil
}

// RequestPasswordReset menerbitkan token reset kata sandi untuk identitas email terdaftar.
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *iamv1.RequestPasswordResetRequest) (*iamv1.RequestPasswordResetResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	email := req.GetEmail()
	if email == "" {
		email = req.GetIdentifier()
	}
	res, err := h.svc.RequestPasswordReset(ctx, email)
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RequestPasswordResetResponse{IssuedToken: toProtoIssuedToken(res)}, nil
}

// ConfirmPasswordReset mengganti kata sandi pengguna dengan token reset sekali pakai.
func (h *AuthHandler) ConfirmPasswordReset(ctx context.Context, req *iamv1.ConfirmPasswordResetRequest) (*iamv1.ConfirmPasswordResetResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.ConfirmPasswordReset(ctx, user.ConfirmPasswordResetInput{
		RawToken:    req.GetRawToken(),
		NewPassword: req.GetNewPassword(),
	}); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ConfirmPasswordResetResponse{}, nil
}

// EnrollTotp memulai pendaftaran TOTP Google Authenticator terenkripsi bagi pengguna.
func (h *AuthHandler) EnrollTotp(ctx context.Context, req *iamv1.EnrollTotpRequest) (*iamv1.EnrollTotpResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.EnrollTOTP(ctx, req.GetUserId(), req.GetAccountLabel())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.EnrollTotpResponse{
		FactorId:   res.FactorID,
		OtpAuthUri: res.OTPAuthURI,
	}, nil
}

// ConfirmTotp memverifikasi kode TOTP perdana untuk mengaktifkan faktor MFA dan menerbitkan recovery codes.
func (h *AuthHandler) ConfirmTotp(ctx context.Context, req *iamv1.ConfirmTotpRequest) (*iamv1.ConfirmTotpResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.ConfirmTOTP(ctx, req.GetUserId(), req.GetCode())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ConfirmTotpResponse{
		Factor:        toProtoTOTPFactor(res.Factor),
		RecoveryCodes: res.RecoveryCodes,
	}, nil
}

// DisableTotp menonaktifkan TOTP pengguna dan menghapus kode pemulihannya.
func (h *AuthHandler) DisableTotp(ctx context.Context, req *iamv1.DisableTotpRequest) (*iamv1.DisableTotpResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if err := h.svc.DisableTOTP(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.DisableTotpResponse{}, nil
}

// RegenerateRecoveryCodes membuat ulang sekumpulan kode pemulihan sekali pakai bagi pengguna.
func (h *AuthHandler) RegenerateRecoveryCodes(ctx context.Context, req *iamv1.RegenerateRecoveryCodesRequest) (*iamv1.RegenerateRecoveryCodesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	codes, err := h.svc.RegenerateRecoveryCodes(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RegenerateRecoveryCodesResponse{RecoveryCodes: codes}, nil
}

// CountUnusedRecoveryCodes menghitung sisa kode pemulihan 2FA yang masih aktif.
func (h *AuthHandler) CountUnusedRecoveryCodes(ctx context.Context, req *iamv1.CountUnusedRecoveryCodesRequest) (*iamv1.CountUnusedRecoveryCodesResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := enforceSelfCaller(ctx, req.GetUserId()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	count, err := h.svc.CountUnusedRecoveryCodes(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CountUnusedRecoveryCodesResponse{Count: count}, nil
}
