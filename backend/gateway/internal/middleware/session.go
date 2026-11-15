package middleware

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// SessionStorePort mendefinisikan operasi Redis minimum yang dibutuhkan Gateway Runtime (peran ACL gw_runtime).
type SessionStorePort interface {
	ReadSessionContext(ctx context.Context, tokenHashHex string, now time.Time) (*authsession.SessionValidationContext, error)
	TouchActivity(ctx context.Context, tokenHashHex string, tok authsession.SessionTokenRecord, idleTimeout time.Duration, now time.Time) error
	SetPreAuth(ctx context.Context, rec authsession.PreAuthRecord, ttl time.Duration) error
	GetPreAuth(ctx context.Context, cookieHashHex string, now time.Time) (*authsession.PreAuthRecord, error)
	DeletePreAuth(ctx context.Context, cookieHashHex string) error
	CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) (int64, error)
	GetGate(ctx context.Context) (*authsession.SecurityGateRecord, error)
}

// RequestMetadata menyimpan atribut HTTP mentah yang diperlukan untuk validasi keamanan per-permintaan.
type RequestMetadata struct {
	RequestID  string
	ClientIP   string
	UserAgent  string
	RawCookie  string
	CookieHash string
	CSRFToken  string
}

// ResolvedSession merepresentasikan hasil autentikasi fail-closed beserta snapshot akses terkini.
type ResolvedSession struct {
	RawToken  string
	TokenHash string
	Token     *authsession.SessionTokenRecord
	Activity  *authsession.SessionActivityRecord
	Security  *authsession.UserSecurityRecord
	Snapshot  *authz.AccessSnapshot
	Caller    authcontext.CallerContext
}

type reqMetaContextKey struct{}
type resolvedSessionContextKey struct{}

// WithRequestMetadata menyematkan RequestMetadata ke dalam context permintaan.
func WithRequestMetadata(ctx context.Context, meta RequestMetadata) context.Context {
	return context.WithValue(ctx, reqMetaContextKey{}, meta)
}

// RequestMetadataFromContext membaca RequestMetadata dari context permintaan.
func RequestMetadataFromContext(ctx context.Context) (RequestMetadata, bool) {
	if ctx == nil {
		return RequestMetadata{}, false
	}
	meta, ok := ctx.Value(reqMetaContextKey{}).(RequestMetadata)
	return meta, ok
}

// WithResolvedSession menyimpan ResolvedSession ke dalam context permintaan agar resolver tidak membaca Redis berulang kali.
func WithResolvedSession(ctx context.Context, sess *ResolvedSession) context.Context {
	if sess == nil {
		return ctx
	}
	return context.WithValue(ctx, resolvedSessionContextKey{}, sess)
}

// ResolvedSessionFromContext mengambil ResolvedSession yang telah tervalidasi bila sudah ada di context.
func ResolvedSessionFromContext(ctx context.Context) (*ResolvedSession, bool) {
	if ctx == nil {
		return nil, false
	}
	sess, ok := ctx.Value(resolvedSessionContextKey{}).(*ResolvedSession)
	return sess, ok && sess != nil
}

// SessionConfig menyimpan parameter kebijakan sesi dan cookie milik AuthManager.
type SessionConfig struct {
	CookieName         string
	CookieSecure       bool
	PreAuthTTL         time.Duration
	SessionIdleTimeout time.Duration
	SessionMaxAge      time.Duration
}

// AuthManager menangani siklus verifikasi sesi Redis, CSRF mutasi, pre-auth handle, dan rate limit secara fail-closed.
type AuthManager struct {
	store  SessionStorePort
	loader *SnapshotLoader
	cfg    SessionConfig

	mu   sync.Mutex
	nowF func() time.Time
}

// NewAuthManager membentuk AuthManager baru dengan validasi port wajib.
func NewAuthManager(store SessionStorePort, loader *SnapshotLoader, cfg SessionConfig) (*AuthManager, error) {
	if store == nil {
		return nil, errors.New("session store is required")
	}
	if loader == nil {
		return nil, errors.New("snapshot loader is required")
	}
	if strings.TrimSpace(cfg.CookieName) == "" {
		cfg.CookieName = browserauth.DefaultSessionCookieName
	}
	if cfg.PreAuthTTL <= 0 {
		cfg.PreAuthTTL = 15 * time.Minute
	}
	if cfg.SessionIdleTimeout <= 0 {
		cfg.SessionIdleTimeout = 30 * time.Minute
	}
	if cfg.SessionMaxAge <= 0 {
		cfg.SessionMaxAge = 8 * time.Hour
	}
	return &AuthManager{
		store:  store,
		loader: loader,
		cfg:    cfg,
		nowF:   time.Now,
	}, nil
}

func (m *AuthManager) now() time.Time {
	m.mu.Lock()
	f := m.nowF
	m.mu.Unlock()
	if f == nil {
		return time.Now()
	}
	return f()
}

// BuildCallerContext membentuk CallerContext dasar (baik terautentikasi maupun anonim ber-metadata permintaan).
func (m *AuthManager) BuildCallerContext(ctx context.Context) authcontext.CallerContext {
	meta, _ := RequestMetadataFromContext(ctx)
	caller := authcontext.CallerContext{
		RequestID: meta.RequestID,
		ClientIP:  meta.ClientIP,
		UserAgent: meta.UserAgent,
	}
	if sess, ok := ResolvedSessionFromContext(ctx); ok && sess != nil {
		return sess.Caller
	}
	return caller
}

// AuthenticateOptional memeriksa cookie sesi saat ini tanpa mewajibkan login (misal pada Query.browserSession).
// Mengembalikan (nil, nil) apabila cookie tidak ada atau sesi tidak ditemukan/berakhir; mengembalikan error bila Redis down atau gate tertutup.
func (m *AuthManager) AuthenticateOptional(ctx context.Context) (*ResolvedSession, error) {
	if sess, ok := ResolvedSessionFromContext(ctx); ok {
		return sess, nil
	}
	meta, _ := RequestMetadataFromContext(ctx)
	if meta.CookieHash == "" {
		return nil, nil
	}

	now := m.now()
	sessCtx, err := m.store.ReadSessionContext(ctx, meta.CookieHash, now)
	if err != nil {
		if isUnavailableErr(err) {
			return nil, ErrUnavailable("security session store unavailable", err)
		}
		return nil, nil
	}
	if err := m.store.TouchActivity(ctx, meta.CookieHash, *sessCtx.Token, m.cfg.SessionIdleTimeout, now); err != nil {
		if isUnavailableErr(err) {
			return nil, ErrUnavailable("security session store unavailable", err)
		}
		return nil, nil
	}

	caller := buildAuthenticatedCaller(meta, sessCtx.Token, sessCtx.Snapshot, sessCtx.Security)
	snap, err := m.loader.EnsureFresh(ctx, sessCtx, caller)
	if err != nil {
		if gwErr, ok := AsGatewayError(err); ok && gwErr.Code == CodeUnavailable {
			return nil, err
		}
		return nil, nil
	}

	caller.PermRevision = snap.PermRevision
	caller.AuthEpoch = snap.AuthEpoch
	return &ResolvedSession{
		RawToken:  meta.RawCookie,
		TokenHash: meta.CookieHash,
		Token:     sessCtx.Token,
		Activity:  sessCtx.Activity,
		Security:  sessCtx.Security,
		Snapshot:  snap,
		Caller:    caller,
	}, nil
}

// RequireAuthenticated memvalidasi sesi opaque, gate Redis, idle/abs expiry, TouchActivity Lua XX, dan kesegaran snapshot <= 30s.
func (m *AuthManager) RequireAuthenticated(ctx context.Context) (*ResolvedSession, error) {
	if sess, ok := ResolvedSessionFromContext(ctx); ok {
		return sess, nil
	}
	meta, _ := RequestMetadataFromContext(ctx)
	if meta.CookieHash == "" {
		return nil, ErrUnauthenticated("session cookie is required", nil)
	}

	now := m.now()
	sessCtx, err := m.store.ReadSessionContext(ctx, meta.CookieHash, now)
	if err != nil {
		if isUnavailableErr(err) {
			return nil, ErrUnavailable("security session store unavailable", err)
		}
		if errors.Is(err, authsession.ErrUserSecurityBlocked) {
			m.clearSessionCookie(ctx)
			return nil, ErrForbidden("user account is suspended or blocked", err)
		}
		// Pertahankan handle pra-autentikasi (pre-auth) yang sah agar permintaan login/verifikasi berikutnya tidak kehilangan pasangan CSRF.
		if errors.Is(err, authsession.ErrSessionNotFound) {
			preAuth, preErr := m.store.GetPreAuth(ctx, meta.CookieHash, now)
			if isUnavailableErr(preErr) {
				return nil, ErrUnavailable("security session store unavailable", preErr)
			}
			if preErr == nil && preAuth != nil {
				return nil, ErrUnauthenticated("session is missing, expired, or revoked", err)
			}
		}
		m.clearSessionCookie(ctx)
		return nil, ErrUnauthenticated("session is missing, expired, or revoked", err)
	}

	if err := m.store.TouchActivity(ctx, meta.CookieHash, *sessCtx.Token, m.cfg.SessionIdleTimeout, now); err != nil {
		if isUnavailableErr(err) {
			return nil, ErrUnavailable("security session store unavailable", err)
		}
		m.clearSessionCookie(ctx)
		return nil, ErrUnauthenticated("session is no longer active", err)
	}

	caller := buildAuthenticatedCaller(meta, sessCtx.Token, sessCtx.Snapshot, sessCtx.Security)
	snap, err := m.loader.EnsureFresh(ctx, sessCtx, caller)
	if err != nil {
		return nil, err
	}
	caller.PermRevision = snap.PermRevision
	caller.AuthEpoch = snap.AuthEpoch

	return &ResolvedSession{
		RawToken:  meta.RawCookie,
		TokenHash: meta.CookieHash,
		Token:     sessCtx.Token,
		Activity:  sessCtx.Activity,
		Security:  sessCtx.Security,
		Snapshot:  snap,
		Caller:    caller,
	}, nil
}

// RequireMutationCSRF menegakkan token X-CSRF-Token untuk seluruh GraphQL Mutation (baik sesi aktif maupun handle pra-autentikasi).
func (m *AuthManager) RequireMutationCSRF(ctx context.Context) (*ResolvedSession, error) {
	meta, _ := RequestMetadataFromContext(ctx)
	if strings.TrimSpace(meta.CSRFToken) == "" {
		return nil, ErrForbidden("missing X-CSRF-Token header", nil)
	}
	if meta.CookieHash == "" {
		return nil, ErrForbidden("missing session cookie for CSRF verification", nil)
	}

	now := m.now()
	sessCtx, err := m.store.ReadSessionContext(ctx, meta.CookieHash, now)
	if err == nil && sessCtx != nil && sessCtx.Token != nil {
		if !browserauth.VerifyCSRFTokenDigest(meta.CSRFToken, sessCtx.Token.CSRFDigestHex) {
			return nil, ErrForbidden("csrf token mismatch", nil)
		}
		if touchErr := m.store.TouchActivity(ctx, meta.CookieHash, *sessCtx.Token, m.cfg.SessionIdleTimeout, now); touchErr != nil {
			if isUnavailableErr(touchErr) {
				return nil, ErrUnavailable("security session store unavailable", touchErr)
			}
			return nil, ErrUnauthenticated("session is no longer active", touchErr)
		}
		caller := buildAuthenticatedCaller(meta, sessCtx.Token, sessCtx.Snapshot, sessCtx.Security)
		snap, snapErr := m.loader.EnsureFresh(ctx, sessCtx, caller)
		if snapErr != nil {
			return nil, snapErr
		}
		caller.PermRevision = snap.PermRevision
		caller.AuthEpoch = snap.AuthEpoch
		return &ResolvedSession{
			RawToken:  meta.RawCookie,
			TokenHash: meta.CookieHash,
			Token:     sessCtx.Token,
			Activity:  sessCtx.Activity,
			Security:  sessCtx.Security,
			Snapshot:  snap,
			Caller:    caller,
		}, nil
	}

	if err != nil && isUnavailableErr(err) {
		return nil, ErrUnavailable("security session store unavailable", err)
	}

	// Bila tidak cocok sesi aktif, periksa apakah cookie merupakan handle pra-autentikasi valid milik Gateway.
	if errors.Is(err, authsession.ErrSessionNotFound) {
		preAuth, preErr := m.store.GetPreAuth(ctx, meta.CookieHash, now)
		if preErr != nil {
			if isUnavailableErr(preErr) {
				return nil, ErrUnavailable("security session store unavailable", preErr)
			}
			m.ClearSessionCookie(ctx)
			return nil, ErrForbidden("invalid or expired pre-auth CSRF handle", preErr)
		}
		if !browserauth.VerifyCSRFTokenDigest(meta.CSRFToken, preAuth.CSRFDigestHex) {
			return nil, ErrForbidden("pre-auth csrf token mismatch", nil)
		}
		return nil, nil
	}

	if !isUnavailableErr(err) {
		m.ClearSessionCookie(ctx)
	}
	return nil, ErrForbidden("session is invalid or expired for mutation", err)
}

// IssuePreAuthSession menghasilkan pasangan cookie __Host- pra-login dan CSRF token lalu menyimpannya ke Redis Gateway secara fail-closed.
func (m *AuthManager) IssuePreAuthSession(ctx context.Context) (string, error) {
	gate, gErr := m.store.GetGate(ctx)
	if gErr != nil {
		return "", ErrUnavailable("security gate verification failed", gErr)
	}
	if gate == nil || !gate.Open {
		return "", ErrUnavailable("security gate is closed", authsession.ErrSecurityGateClosed)
	}

	rawCookie, cookieHash, err := browserauth.GenerateOpaquePair()
	if err != nil {
		return "", ErrUnavailable("failed to generate pre-auth handle", err)
	}
	rawCSRF, csrfHash, err := browserauth.GenerateOpaquePair()
	if err != nil {
		return "", ErrUnavailable("failed to generate csrf token", err)
	}

	now := m.now()
	rec := authsession.PreAuthRecord{
		TokenHashHex:  cookieHash,
		CSRFDigestHex: csrfHash,
		CreatedAtUnix: now.Unix(),
		ExpiresAtUnix: now.Add(m.cfg.PreAuthTTL).Unix(),
	}
	if err := m.store.SetPreAuth(ctx, rec, m.cfg.PreAuthTTL); err != nil {
		return "", ErrUnavailable("failed to store pre-auth session", err)
	}

	if eff := browserauth.EffectsFromContext(ctx); eff != nil {
		eff.AddCookie(browserauth.BuildSessionCookie(m.cfg.CookieName, rawCookie, m.cfg.PreAuthTTL, m.cfg.CookieSecure))
		eff.SetHeader(browserauth.CSRFHeaderName, rawCSRF)
	}
	return rawCSRF, nil
}

// PromoteAuthenticatedCookie menetapkan cookie sesi buram hasil login/MFA, header CSRF/Location, serta menghapus handle pre-auth lama.
func (m *AuthManager) PromoteAuthenticatedCookie(ctx context.Context, rawSessionToken, csrfToken, redirectLocation string, maxAge time.Duration) {
	if maxAge <= 0 {
		maxAge = m.cfg.SessionMaxAge
	}
	meta, _ := RequestMetadataFromContext(ctx)
	if meta.CookieHash != "" && meta.RawCookie != rawSessionToken {
		_ = m.store.DeletePreAuth(ctx, meta.CookieHash)
	}
	if eff := browserauth.EffectsFromContext(ctx); eff != nil {
		eff.AddCookie(browserauth.BuildSessionCookie(m.cfg.CookieName, rawSessionToken, maxAge, m.cfg.CookieSecure))
		if csrfToken != "" {
			eff.SetHeader(browserauth.CSRFHeaderName, csrfToken)
		}
		if redirectLocation != "" {
			eff.SetHeader(browserauth.LocationHeaderName, redirectLocation)
		}
	}
}

// ClearSessionCookie menghapus cookie __Host- dari peramban (misal saat logout atau sesi dicabut).
func (m *AuthManager) ClearSessionCookie(ctx context.Context) {
	m.clearSessionCookie(ctx)
}

func (m *AuthManager) clearSessionCookie(ctx context.Context) {
	if eff := browserauth.EffectsFromContext(ctx); eff != nil {
		eff.AddCookie(browserauth.BuildClearSessionCookie(m.cfg.CookieName, m.cfg.CookieSecure))
	}
}

// CheckRateLimit menegakkan pembatas laju Redis atomik per scope dan subjek dengan perilaku fail-closed.
func (m *AuthManager) CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) error {
	if strings.TrimSpace(subject) == "" {
		subject = "unknown"
	}
	_, err := m.store.CheckRateLimit(ctx, scope, subject, limit, window)
	if err == nil {
		return nil
	}
	if errors.Is(err, authsession.ErrRateLimitExceeded) {
		return ErrTooManyRequests("rate limit exceeded, please retry later", err)
	}
	return ErrUnavailable("rate limiter unavailable", err)
}

func buildAuthenticatedCaller(
	meta RequestMetadata,
	tok *authsession.SessionTokenRecord,
	snap *authz.AccessSnapshot,
	sec *authsession.UserSecurityRecord,
) authcontext.CallerContext {
	var permRev int64
	if snap != nil && snap.PermRevision > permRev {
		permRev = snap.PermRevision
	}
	if sec != nil && sec.PermRevision > permRev {
		permRev = sec.PermRevision
	}
	return authcontext.CallerContext{
		PrincipalID:   tok.UserID,
		SessionID:     tok.SessionID,
		UserKind:      tok.UserKind,
		AuthEpoch:     tok.AuthEpoch,
		PermRevision:  permRev,
		RequestID:     meta.RequestID,
		ClientIP:      meta.ClientIP,
		UserAgent:     meta.UserAgent,
		Authenticated: true,
	}
}

func isUnavailableErr(err error) bool {
	return errors.Is(err, authsession.ErrRedisUnavailable) || errors.Is(err, authsession.ErrSecurityGateClosed)
}
