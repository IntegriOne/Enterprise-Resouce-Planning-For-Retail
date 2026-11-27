package handler

import (
	"context"
	"fmt"

	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

// SessionHandler menangani permintaan gRPC siklus hidup sesi opaque, CSRF, dan evaluasi otorisasi.
type SessionHandler struct {
	iamv1.UnimplementedSessionServiceServer
	svc       *session.Service
	errMapper *ErrorMapper
}

// NewSessionHandler membentuk handler gRPC untuk SessionService secara fail-closed.
func NewSessionHandler(svc *session.Service, errMapper *ErrorMapper) (*SessionHandler, error) {
	if svc == nil {
		return nil, fmt.Errorf("session service cannot be nil")
	}
	if errMapper == nil {
		errMapper = NewErrorMapper(nil)
	}
	return &SessionHandler{
		svc:       svc,
		errMapper: errMapper,
	}, nil
}

// IssueSession menerbitkan pasangan token sesi opaque dan CSRF baru untuk pengguna.
func (h *SessionHandler) IssueSession(ctx context.Context, req *iamv1.IssueSessionRequest) (*iamv1.IssueSessionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return nil, h.errMapper.ToStatus(iamerrors.ErrUnauthorized)
	}
	if caller.PrincipalID != req.GetUserId() {
		return nil, h.errMapper.ToStatus(iamerrors.ErrPermissionDenied)
	}
	if req.GetMfaVerified() {
		return nil, h.errMapper.ToStatus(iamerrors.NewValidation("mfa_verified", iamerrors.Code("iam.session.mfa_flag_requires_auth_flow"), "mfa verification must use the authentication flow"))
	}
	issued, err := h.svc.Issue(ctx, session.IssueInput{
		UserID:      req.GetUserId(),
		MFAVerified: false,
		IPAddress:   req.GetIpAddress(),
		UserAgent:   req.GetUserAgent(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if issued == nil {
		return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.empty_result"), "session issuance returned empty result"))
	}
	return &iamv1.IssueSessionResponse{IssuedSession: toProtoIssuedSession(issued)}, nil
}

// AuthenticateSession memvalidasi token sesi opaque dan menyegarkan sliding idle timeout.
func (h *SessionHandler) AuthenticateSession(ctx context.Context, req *iamv1.AuthenticateSessionRequest) (*iamv1.AuthenticateSessionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	sess, err := h.svc.Authenticate(ctx, req.GetSessionToken())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.AuthenticateSessionResponse{Session: toProtoSession(sess)}, nil
}

// VerifyCsrf memvalidasi token CSRF terhadap sesi aktif tertentu.
func (h *SessionHandler) VerifyCsrf(ctx context.Context, req *iamv1.VerifyCsrfRequest) (*iamv1.VerifyCsrfResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	sess, err := h.svc.VerifyCSRF(ctx, req.GetSessionId(), req.GetCsrfToken())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.VerifyCsrfResponse{Session: toProtoSession(sess)}, nil
}

// MarkSessionMfaVerified menandai sesi telah terverifikasi MFA dan merotasi token CSRF-nya.
func (h *SessionHandler) MarkSessionMfaVerified(ctx context.Context, req *iamv1.MarkSessionMfaVerifiedRequest) (*iamv1.MarkSessionMfaVerifiedResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindConflict, iamerrors.Code("iam.session.mfa_verification_required"), "mfa verification must use the authentication flow"))
}

// LogoutSession mengakhiri sesi aktif berdasarkan token sesi opaque.
func (h *SessionHandler) LogoutSession(ctx context.Context, req *iamv1.LogoutSessionRequest) (*iamv1.LogoutSessionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.Logout(ctx, req.GetSessionToken()); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.LogoutSessionResponse{}, nil
}

// RevokeSessionById mencabut satu sesi secara eksplisit berdasarkan UUID sesi.
func (h *SessionHandler) RevokeSessionById(ctx context.Context, req *iamv1.RevokeSessionByIdRequest) (*iamv1.RevokeSessionByIdResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	if err := h.svc.RevokeByID(ctx, req.GetSessionId(), fromProtoSessionEndReason(req.GetReason())); err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RevokeSessionByIdResponse{}, nil
}

// RevokeAllUserSessions mencabut seluruh sesi aktif milik pengguna.
func (h *SessionHandler) RevokeAllUserSessions(ctx context.Context, req *iamv1.RevokeAllUserSessionsRequest) (*iamv1.RevokeAllUserSessionsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	count, err := h.svc.RevokeAllUserSessions(ctx, req.GetUserId(), fromProtoSessionEndReason(req.GetReason()))
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.RevokeAllUserSessionsResponse{RevokedCount: count}, nil
}

// ListUserSessions mengambil riwayat atau sesi aktif pengguna secara terpaginasi.
func (h *SessionHandler) ListUserSessions(ctx context.Context, req *iamv1.ListUserSessionsRequest) (*iamv1.ListUserSessionsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.ListByUser(ctx, domainrepo.SessionFilter{
		UserID:     req.GetUserId(),
		ActiveOnly: req.GetActiveOnly(),
		Page:       fromProtoPageRequest(req.GetPage()),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ListUserSessionsResponse{
		Items: toProtoSessions(res.Items),
		Page:  toProtoPageInfo(res),
	}, nil
}

// CleanupEndedSessions membersihkan rekaman sesi yang sudah berakhir melewati retensi.
func (h *SessionHandler) CleanupEndedSessions(ctx context.Context, req *iamv1.CleanupEndedSessionsRequest) (*iamv1.CleanupEndedSessionsResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	retention, err := fromProtoDuration("retention", req.GetRetention())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	deleted, err := h.svc.CleanupEndedSessions(ctx, retention)
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.CleanupEndedSessionsResponse{DeletedCount: deleted}, nil
}

// AuthorizeSession mengevaluasi autentikasi sesi, CSRF opsional, izin, scope, dan step-up MFA.
func (h *SessionHandler) AuthorizeSession(ctx context.Context, req *iamv1.AuthorizeSessionRequest) (*iamv1.AuthorizeSessionResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	scope, err := fromProtoScopeOptional(req.GetRequiredScope())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	res, err := h.svc.Authorize(ctx, session.AuthorizeRequest{
		SessionToken:           req.GetSessionToken(),
		CSRFToken:              req.GetCsrfToken(),
		RequireCSRF:            req.GetRequireCsrf(),
		RequiredPermission:     req.GetRequiredPermission(),
		RequiredScope:          scope,
		RequireSensitiveReauth: req.GetRequireSensitiveReauth(),
	})
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if res == nil {
		return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.empty_result"), "session authorization returned empty result"))
	}
	return &iamv1.AuthorizeSessionResponse{
		User:     toProtoUser(res.User),
		Session:  toProtoSession(res.Session),
		Snapshot: toProtoAccessSnapshot(res.Snapshot),
	}, nil
}

// ResolvePostLoginRedirect menentukan rute tujuan pasca-login tervalidasi berdasarkan peran efektif pengguna.
func (h *SessionHandler) ResolvePostLoginRedirect(ctx context.Context, req *iamv1.ResolvePostLoginRedirectRequest) (*iamv1.ResolvePostLoginRedirectResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	path, err := h.svc.ResolvePostLoginRedirect(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	return &iamv1.ResolvePostLoginRedirectResponse{RedirectPath: path}, nil
}

// GetSessionCsrf mengekstrak ulang token CSRF deterministik untuk sesi aktif guna sinkronisasi multi-tab aman.
func (h *SessionHandler) GetSessionCsrf(ctx context.Context, req *iamv1.GetSessionCsrfRequest) (*iamv1.GetSessionCsrfResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return nil, h.errMapper.ToStatus(iamerrors.ErrUnauthorized)
	}
	if caller.SessionID == "" || caller.SessionID != req.GetSessionId() {
		return nil, h.errMapper.ToStatus(iamerrors.ErrPermissionDenied)
	}
	res, err := h.svc.GetSessionCSRF(ctx, req.GetSessionId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if res == nil {
		return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.empty_result"), "session csrf lookup returned empty result"))
	}
	return &iamv1.GetSessionCsrfResponse{
		Session:   toProtoSession(res.Session),
		CsrfToken: res.CSRFToken,
	}, nil
}

// RefreshUserAccessSnapshot mengevaluasi snapshot akses terbaru beserta izin dasbor /identity dan memproyeksikannya ke Redis.
func (h *SessionHandler) RefreshUserAccessSnapshot(ctx context.Context, req *iamv1.RefreshUserAccessSnapshotRequest) (*iamv1.RefreshUserAccessSnapshotResponse, error) {
	if req == nil {
		return nil, nilRequestError(h.errMapper)
	}
	res, err := h.svc.RefreshAccessSnapshot(ctx, req.GetUserId())
	if err != nil {
		return nil, h.errMapper.ToStatus(err)
	}
	if res == nil {
		return nil, h.errMapper.ToStatus(iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.empty_result"), "access snapshot refresh returned empty result"))
	}
	return &iamv1.RefreshUserAccessSnapshotResponse{
		User:                 toProtoUser(res.User),
		Snapshot:             toProtoAccessSnapshot(res.Snapshot),
		CanAccessIdentity:    res.CanAccessIdentity,
		DefaultRedirectPath:  res.DefaultRedirectPath,
		AllowedRedirectPaths: res.AllowedRedirectPaths,
	}, nil
}
