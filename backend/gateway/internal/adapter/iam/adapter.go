// Package iam mengimplementasikan adapter domain IAM, kontrol otorisasi RBAC/reauth per operasi, dan mapper model GraphQL <-> gRPC.
package iam

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// DefaultAvatarBaseURL adalah awalan URL publik deterministik untuk avatar pengguna berformat AVIF.
const DefaultAvatarBaseURL = "https://erp.localhost:8443/media/avatars"

// Config menyimpan parameter kebijakan runtime untuk adapter IAM Gateway.
type Config struct {
	RPCTimeout      time.Duration
	LoginRateMax    int64
	LoginRateWindow time.Duration
	SessionMaxAge   time.Duration
	AvatarBaseURL   string
}

// Adapter menyatukan panggilan gRPC IAM, manajemen sesi/CSRF, serta penegakan izin RBAC Gateway.
type Adapter struct {
	clients           source_connections.IAMClientSet
	authManager       *middleware.AuthManager
	authorizer        *middleware.Authorizer
	redirectValidator *browserauth.RedirectValidator
	cfg               Config
}

// NewAdapter membentuk Adapter domain IAM dengan pemeriksaan ketergantungan wajib fail-closed.
func NewAdapter(
	clients source_connections.IAMClientSet,
	authManager *middleware.AuthManager,
	authorizer *middleware.Authorizer,
	redirectValidator *browserauth.RedirectValidator,
	cfg Config,
) (*Adapter, error) {
	if authManager == nil {
		return nil, errors.New("auth manager is required")
	}
	if authorizer == nil {
		return nil, errors.New("authorizer is required")
	}
	if redirectValidator == nil {
		return nil, errors.New("redirect validator is required")
	}
	if cfg.RPCTimeout <= 0 {
		cfg.RPCTimeout = 5 * time.Second
	}
	if cfg.LoginRateMax <= 0 {
		cfg.LoginRateMax = 20
	}
	if cfg.LoginRateWindow <= 0 {
		cfg.LoginRateWindow = 60 * time.Second
	}
	if cfg.SessionMaxAge <= 0 {
		cfg.SessionMaxAge = 8 * time.Hour
	}
	if strings.TrimSpace(cfg.AvatarBaseURL) == "" {
		cfg.AvatarBaseURL = DefaultAvatarBaseURL
	} else {
		cfg.AvatarBaseURL = strings.TrimRight(strings.TrimSpace(cfg.AvatarBaseURL), "/")
	}
	return &Adapter{
		clients:           clients,
		authManager:       authManager,
		authorizer:        authorizer,
		redirectValidator: redirectValidator,
		cfg:               cfg,
	}, nil
}

func (a *Adapter) rpcCtx(ctx context.Context, caller authcontext.CallerContext) (context.Context, context.CancelFunc) {
	return context.WithTimeout(authcontext.WithCaller(ctx, caller), a.cfg.RPCTimeout)
}

func (a *Adapter) requirePerm(ctx context.Context, permCode string, requireSensitive bool) (*middleware.ResolvedSession, error) {
	return a.authorizer.RequirePermission(ctx, permCode, authz.ScopeGrant{Type: authz.ScopeGlobal}, requireSensitive)
}

// requireSelfOrPerm memisahkan evaluasi otorisasi antara target akun milik sendiri (*self) dan akses administratif lintas pengguna.
func (a *Adapter) requireSelfOrPerm(
	ctx context.Context,
	targetUserID string,
	selfPermCode string,
	adminPermCode string,
) (*middleware.ResolvedSession, error) {
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	// Normalisasi target user ID agar pembandingan kepemilikan tetap akurat baik saat menerima GID maupun UUID mentah.
	if normalizedTarget, parseErr := ParseTypedID(targetUserID, ResourceUser); parseErr == nil &&
		sess.Token != nil &&
		strings.EqualFold(strings.TrimSpace(sess.Token.UserID), normalizedTarget) {
		selfSess, selfErr := a.authorizer.RequireSelfPermission(ctx, selfPermCode, false)
		if selfErr == nil {
			return selfSess, nil
		}
		// Pemegang izin administratif global tetap dapat membaca sumber dayanya sendiri meskipun tidak memegang varian *self.
		if adminSess, adminErr := a.requirePerm(ctx, adminPermCode, false); adminErr == nil {
			return adminSess, nil
		}
		return nil, selfErr
	}
	return a.requirePerm(ctx, adminPermCode, false)
}

func (a *Adapter) checkAuthRateLimit(ctx context.Context, identifier string) error {
	meta, _ := middleware.RequestMetadataFromContext(ctx)
	ip := strings.TrimSpace(meta.ClientIP)
	if ip == "" {
		ip = "unknown"
	}
	if err := a.authManager.CheckRateLimit(ctx, "auth_ip", ip, a.cfg.LoginRateMax, a.cfg.LoginRateWindow); err != nil {
		return err
	}
	cleanID := strings.ToLower(strings.TrimSpace(identifier))
	if cleanID != "" {
		// Gunakan hash identifier lintas-IP agar password spraying terdistribusi tetap terbatasi tanpa menyimpan PII di Redis.
		if err := a.authManager.CheckRateLimit(ctx, "auth_subject", authsession.HashOpaqueToken(cleanID), a.cfg.LoginRateMax, a.cfg.LoginRateWindow); err != nil {
			return err
		}
	}
	return nil
}

// checkPreAuthRateLimit membatasi pembuatan sesi pra-autentikasi per IP untuk mencegah kelelahan memori Redis.
func (a *Adapter) checkPreAuthRateLimit(ctx context.Context) error {
	meta, _ := middleware.RequestMetadataFromContext(ctx)
	ip := strings.TrimSpace(meta.ClientIP)
	if ip == "" {
		ip = "unknown"
	}
	limit := a.cfg.LoginRateMax * 3
	if limit < 60 {
		limit = 60
	}
	return a.authManager.CheckRateLimit(ctx, "preauth_ip", ip, limit, a.cfg.LoginRateWindow)
}

// checkMFAChallengeRateLimit membatasi percobaan verifikasi per challenge MFA ter-hash untuk mencegah brute-force TOTP paralel.
func (a *Adapter) checkMFAChallengeRateLimit(ctx context.Context, challenge string) error {
	cleanChallenge := strings.TrimSpace(challenge)
	if cleanChallenge == "" {
		return nil
	}
	return a.authManager.CheckRateLimit(ctx, "mfa_challenge", authsession.HashOpaqueToken(cleanChallenge), 5, 5*time.Minute)
}

const maxPaginationOffset = 1_000_000_000

// clampRelayBounds memvalidasi batas limit dan offset agar aman dikonversi ke int32 tanpa integer overflow.
func clampRelayBounds(limit, offset int) (int32, int32, error) {
	if limit < 0 || offset < 0 || limit > maxPaginationOffset+1000 || offset > maxPaginationOffset+1000 {
		return 0, 0, middleware.ErrInvalidArgument("pagination bounds out of supported range", nil)
	}
	return int32(limit), int32(offset), nil
}

// normalizeAuthError menormalkan error autentikasi/MFA agar tidak membocorkan status akun, lockout, kedaluwarsa, atau replay ke penyerang.
func normalizeAuthError(err error, ambiguousMsg string) error {
	mapped := mapRPCError(err)
	if gwErr, ok := middleware.AsGatewayError(mapped); ok {
		switch gwErr.Code {
		case middleware.CodeInvalidArgument, middleware.CodeUnavailable, middleware.CodeInternalError:
			return mapped
		}
	}
	return middleware.ErrUnauthenticated(ambiguousMsg, nil)
}

func mapRPCError(err error) error {
	return source_connections.MapIAMGRPCError(err)
}
