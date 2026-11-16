package session

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

const opaqueTokenBytes = 32

// Service mengorkestrasi siklus hidup sesi stateful D01-D12, verifikasi CSRF, dan evaluasi otorisasi fail-closed.
type Service struct {
	sessions        domainrepo.SessionRepository
	users           domainrepo.UserRepository
	tx              domainrepo.Transactor
	securityRuntime domainrepo.SecurityRuntimeRepository
	clock           domainservice.Clock
	idGen           domainservice.IDGenerator
	tokens          domainservice.TokenGenerator
	csrfDeriver     domainservice.CSRFDeriver
	policy          domainservice.SessionPolicy
	evaluator       domainservice.AccessEvaluator
	redirect        domainservice.RedirectResolver
	outboxCoord     *redisstore.Coordinator
	projector       domainservice.RuntimeSecurityProjector
}

// NewService membentuk Service sesi dengan validasi ketergantungan wajib secara fail-closed.
func NewService(deps Dependencies) (*Service, error) {
	if deps.Sessions == nil {
		return nil, iamerrors.NewValidation("sessions", iamerrors.Code("iam.service.nil_session_repo"), "session repository is required")
	}
	if deps.Users == nil {
		return nil, iamerrors.NewValidation("users", iamerrors.Code("iam.service.nil_user_repo"), "user repository is required")
	}
	if deps.IDGen == nil {
		return nil, iamerrors.NewValidation("id_gen", iamerrors.Code("iam.service.nil_id_generator"), "id generator is required")
	}
	if deps.Tokens == nil {
		return nil, iamerrors.NewValidation("tokens", iamerrors.Code("iam.service.nil_token_generator"), "token generator is required")
	}

	clock := deps.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	policy := deps.SessionPolicy
	if policy.IdleTimeout() <= 0 || policy.AbsoluteTimeout() <= 0 {
		policy = domainservice.NewDefaultSessionPolicy()
	}

	return &Service{
		sessions:        deps.Sessions,
		users:           deps.Users,
		tx:              deps.Transactor,
		securityRuntime: deps.SecurityRuntime,
		clock:           clock,
		idGen:           deps.IDGen,
		tokens:          deps.Tokens,
		csrfDeriver:     deps.CSRFDeriver,
		policy:          policy,
		evaluator:       deps.AccessEvaluator,
		redirect:        deps.Redirect,
		outboxCoord:     deps.OutboxCoord,
		projector:       deps.Projector,
	}, nil
}

// Issue menerbitkan sesi stateful baru beserta token sesi buram dan token CSRF 256-bit.
func (s *Service) Issue(ctx context.Context, in IssueInput) (*IssuedSession, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", in.UserID); err != nil {
		return nil, err
	}

	ipPtr, err := parseOptionalIP(in.IPAddress)
	if err != nil {
		return nil, err
	}

	// Pastikan hanya pengguna berstatus ACTIVE yang dapat diterbitkan sesi baru.
	u, err := s.users.SelectByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if err := u.CanAuthenticate(); err != nil {
		return nil, err
	}

	sessionID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.session.id_generation_failed"), "failed to generate session id")
	}

	sessionPlaintext, sessionDigest, err := s.tokens.GenerateOpaqueToken(opaqueTokenBytes)
	if err != nil {
		return nil, err
	}
	csrfPlaintext, csrfDigest, csrfKeyVer, err := s.deriveCSRF(sessionID, sessionDigest, 1)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	sess, err := entity.NewSession(sessionID, in.UserID, sessionDigest, csrfDigest, in.MFAVerified, ipPtr, in.UserAgent, now)
	if err != nil {
		return nil, err
	}
	sess.SetSecurityMetadata(1, csrfKeyVer)
	if err := s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.sessions.Insert(txCtx, sess); err != nil {
			return err
		}
		return s.recordSessionIssuedInTx(txCtx, sess)
	}); err != nil {
		return nil, err
	}

	return &IssuedSession{
		Session:      sess,
		SessionToken: sessionPlaintext,
		CSRFToken:    csrfPlaintext,
	}, nil
}

// Authenticate memvalidasi token sesi buram, menegakkan timeout idle/absolut, dan menggeser batas idle sesi aktif.
func (s *Service) Authenticate(ctx context.Context, sessionToken string) (*entity.Session, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	digest := s.tokens.Digest(sessionToken)
	if len(digest) == 0 {
		return nil, iamerrors.WithField(iamerrors.ErrUnauthorized, "session_token", "session token is required")
	}

	sess, err := s.sessions.SelectByTokenDigest(ctx, digest)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := s.policy.ValidateAndTouch(sess, now); err != nil {
		s.recordNaturalExpiryIfApplicable(ctx, sess.ID, err, now)
		return nil, err
	}
	if err := s.ensureSessionOwnerActive(ctx, sess); err != nil {
		return nil, err
	}

	if err := s.sessions.UpdateActivity(ctx, sess.ID, sess.LastActiveAt, sess.IdleExpiresAt); err != nil {
		return nil, err
	}
	return sess, nil
}

// VerifyCSRF memeriksa kecocokan token CSRF terhadap sesi aktif berdasarkan UUID sesi.
func (s *Service) VerifyCSRF(ctx context.Context, sessionID, csrfToken string) (*entity.Session, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return nil, err
	}

	sess, err := s.sessions.SelectByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	if err := sess.CheckStatus(now); err != nil {
		return nil, err
	}
	if err := s.ensureSessionOwnerActive(ctx, sess); err != nil {
		return nil, err
	}
	if err := s.VerifyCSRFOnSession(sess, csrfToken); err != nil {
		return nil, err
	}
	return sess, nil
}

// VerifyCSRFOnSession membandingkan token CSRF secara constant-time terhadap sesi yang sudah dimuat.
func (s *Service) VerifyCSRFOnSession(sess *entity.Session, csrfToken string) error {
	if s == nil || s.tokens == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.service_unavailable"), "session service is not initialized")
	}
	if sess == nil {
		return iamerrors.ErrSessionNotFound
	}
	digest := s.tokens.Digest(csrfToken)
	if len(digest) == 0 || !sess.MatchesCSRFDigest(digest) {
		return iamerrors.ErrCSRFMismatch
	}
	return nil
}

// MarkUserMFAVerified memastikan sesi milik pengguna sebelum mencatat step-up MFA.
func (s *Service) MarkUserMFAVerified(ctx context.Context, userID, sessionID string) (*CSRFTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return nil, err
	}
	sess, err := s.sessions.SelectByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		return nil, iamerrors.WithField(iamerrors.ErrPermissionDenied, "session_id", "session does not belong to user")
	}
	return s.MarkMFAVerified(ctx, sessionID)
}

// MarkMFAVerified mencatat verifikasi MFA pada sesi aktif sambil merotasi token CSRF secara atomik.
func (s *Service) MarkMFAVerified(ctx context.Context, sessionID string) (*CSRFTokenResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return nil, err
	}

	sess, err := s.sessions.SelectByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := sess.CheckStatus(now); err != nil {
		return nil, err
	}
	if err := s.ensureSessionOwnerActive(ctx, sess); err != nil {
		return nil, err
	}
	if err := sess.MarkMFAVerified(now); err != nil {
		return nil, err
	}
	sess.BumpGeneration()

	csrfPlaintext, csrfDigest, csrfKeyVer, err := s.deriveCSRF(sess.ID, sess.TokenDigest(), sess.Generation)
	if err != nil {
		return nil, err
	}
	if err := sess.RotateCSRF(csrfDigest); err != nil {
		return nil, err
	}
	sess.SetSecurityMetadata(sess.Generation, csrfKeyVer)
	if err := s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.sessions.UpdateMFAVerified(txCtx, sess.ID, now, csrfDigest); err != nil {
			return err
		}
		if s.securityRuntime != nil {
			if err := s.securityRuntime.UpdateSessionSecurityMetadata(txCtx, sess.ID, sess.Generation, sess.CSRFKeyVersion, csrfDigest); err != nil {
				return err
			}
		}
		return s.recordSessionUpdatedInTx(txCtx, sess)
	}); err != nil {
		return nil, err
	}

	return &CSRFTokenResult{
		Session:   sess,
		CSRFToken: csrfPlaintext,
	}, nil
}

// Logout mengakhiri sesi secara permanen berdasarkan token sesi buram yang aktif.
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	digest := s.tokens.Digest(sessionToken)
	if len(digest) == 0 {
		return iamerrors.WithField(iamerrors.ErrUnauthorized, "session_token", "session token is required")
	}

	sess, err := s.sessions.SelectByTokenDigest(ctx, digest)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if err := sess.End(entity.SessionEndLogout, now); err != nil {
		return err
	}
	return s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.sessions.UpdateEnded(txCtx, sess.ID, entity.SessionEndLogout, now); err != nil {
			return err
		}
		return s.recordSessionRevokedInTx(txCtx, sess)
	})
}

// RevokeByID mengakhiri satu sesi secara administratif berdasarkan UUID sesi dan alasan standar.
func (s *Service) RevokeByID(ctx context.Context, sessionID string, reason entity.SessionEndReason) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("session_id", sessionID); err != nil {
		return err
	}
	if !reason.IsValid() {
		return iamerrors.ErrInvalidSessionEndReason
	}

	sess, err := s.sessions.SelectByID(ctx, sessionID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if err := sess.End(reason, now); err != nil {
		return err
	}
	return s.runInTx(ctx, func(txCtx context.Context) error {
		if err := s.sessions.UpdateEnded(txCtx, sess.ID, reason, now); err != nil {
			return err
		}
		return s.recordSessionRevokedInTx(txCtx, sess)
	})
}

// RevokeAllUserSessions mencabut seluruh sesi aktif milik seorang pengguna secara massal.
func (s *Service) RevokeAllUserSessions(ctx context.Context, userID string, reason entity.SessionEndReason) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return 0, err
	}
	if !reason.IsValid() {
		return 0, iamerrors.ErrInvalidSessionEndReason
	}
	var revoked int64
	err := s.runInTx(ctx, func(txCtx context.Context) error {
		count, err := s.revokeAllSessionsWithOutbox(txCtx, userID, reason)
		if err != nil {
			return err
		}
		revoked = count
		return nil
	})
	if err != nil {
		return 0, err
	}
	return revoked, nil
}

// ListByUser mengambil daftar sesi pengguna secara berhalaman.
func (s *Service) ListByUser(ctx context.Context, filter domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	if err := s.ensureReady(); err != nil {
		return domainrepo.PageResult[*entity.Session]{}, err
	}
	norm := filter.Normalize()
	if err := entity.ValidateUUID("user_id", norm.UserID); err != nil {
		return domainrepo.PageResult[*entity.Session]{}, err
	}
	return s.sessions.SelectList(ctx, norm)
}

// CleanupEndedSessions membersihkan histori sesi berakhir yang melewati periode retensi lokal OS.
func (s *Service) CleanupEndedSessions(ctx context.Context, retention time.Duration) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	if retention <= 0 {
		return 0, iamerrors.NewValidation("retention", iamerrors.Code("iam.session.invalid_retention"), "retention duration must be positive")
	}
	cutoff := s.clock.Now().Add(-retention)
	return s.sessions.DeleteEndedBefore(ctx, cutoff)
}

// Authorize mengevaluasi otorisasi menyeluruh (sesi, CSRF opsional, status user, kesegaran snapshot, scope, dan reauth sensitif) secara fail-closed.
func (s *Service) Authorize(ctx context.Context, in AuthorizeRequest) (*AuthorizeResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	sess, err := s.Authenticate(ctx, in.SessionToken)
	if err != nil {
		return nil, err
	}
	if in.RequireCSRF {
		if err := s.VerifyCSRFOnSession(sess, in.CSRFToken); err != nil {
			return nil, err
		}
	}

	u, err := s.users.SelectByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	snapshot, err := s.users.SelectAccessSnapshotByUserID(ctx, sess.UserID, now)
	if err != nil {
		return nil, err
	}

	if err := s.evaluator.Authorize(domainservice.AuthorizeInput{
		User:                   u,
		Session:                sess,
		Snapshot:               snapshot,
		RequiredPermission:     in.RequiredPermission,
		RequiredScope:          in.RequiredScope,
		RequireSensitiveReauth: in.RequireSensitiveReauth,
		Now:                    now,
	}); err != nil {
		return nil, err
	}

	return &AuthorizeResult{
		User:     u,
		Session:  sess,
		Snapshot: snapshot,
	}, nil
}

// ResolvePostLoginRedirect menentukan jalur tujuan dashboard internal pengguna berdasarkan peran efektif aktifnya.
func (s *Service) ResolvePostLoginRedirect(ctx context.Context, userID string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return "", err
	}

	u, err := s.users.SelectByID(ctx, userID)
	if err != nil {
		return "", err
	}
	snapshot, err := s.users.SelectAccessSnapshotByUserID(ctx, userID, s.clock.Now())
	if err != nil {
		return "", err
	}
	return s.redirect.ResolvePostLoginPath(u.Kind, snapshot.Roles)
}

func (s *Service) recordNaturalExpiryIfApplicable(ctx context.Context, sessionID string, evalErr error, now time.Time) {
	switch {
	case errors.Is(evalErr, iamerrors.ErrSessionIdleExpired):
		_ = s.sessions.UpdateEnded(ctx, sessionID, entity.SessionEndIdleTimeout, now)
	case errors.Is(evalErr, iamerrors.ErrSessionAbsoluteExpired):
		_ = s.sessions.UpdateEnded(ctx, sessionID, entity.SessionEndAbsoluteTimeout, now)
	}
}

func parseOptionalIP(raw string) (*netip.Addr, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := netip.ParseAddr(trimmed)
	if err != nil {
		return nil, iamerrors.WithField(iamerrors.ErrInvalidIPAddress, "ip_address", err.Error())
	}
	unmapped := parsed.Unmap()
	return &unmapped, nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.sessions == nil || s.users == nil || s.tokens == nil {
		return iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.session.service_unavailable"), "session service is not initialized")
	}
	return nil
}

func (s *Service) ensureSessionOwnerActive(ctx context.Context, sess *entity.Session) error {
	if sess == nil {
		return iamerrors.ErrUnauthorized
	}
	u, err := s.users.SelectByID(ctx, sess.UserID)
	if err != nil {
		if errors.Is(err, iamerrors.ErrUserNotFound) {
			return iamerrors.ErrUnauthorized
		}
		return err
	}
	if err := u.CanAuthenticate(); err != nil {
		return iamerrors.ErrUnauthorized
	}
	return nil
}
