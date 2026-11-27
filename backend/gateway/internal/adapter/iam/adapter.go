// Package iam mengimplementasikan adapter domain IAM, kontrol otorisasi RBAC/reauth per operasi, dan mapper model GraphQL <-> gRPC.
package iam

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
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

// PartyUserLinkRequest menyimpan parameter validasi dan penautan akun mitra ke PIC Enterprise Foundation.
type PartyUserLinkRequest struct {
	PartyID          string
	ContactID        *string
	RequiredRole     string
	FullName         string
	Email            string
	ContactRole      *string
	JobTitle         *string
	Phone            *string
	IsPrimaryContact *bool
}

// PartyPortalProvisioner memverifikasi target mitra Enterprise Foundation dan menetapkan FK `party_contacts.iam_user_id`.
type PartyPortalProvisioner interface {
	PreparePortalUserProvisioning(ctx context.Context, partyID string, contactID *string, requiredRole string) (normalizedPartyID string, normalizedContactID *string, err error)
	LinkPortalUser(ctx context.Context, req PartyUserLinkRequest, iamUserID string) (*model.PartyContact, error)
}

// WorkerPrincipalProvisioner memverifikasi dan menautkan `hr_workers.iam_principal_reference_id` pada layanan Human Resource.
type WorkerPrincipalProvisioner interface {
	PrepareWorkerPrincipalLink(ctx context.Context, workerID string) (normalizedWorkerID string, err error)
	CompleteWorkerPrincipalLink(ctx context.Context, workerID, iamUserID string) error
}

// Option mengonfigurasi kolaborator lintas domain opsional pada Adapter IAM.
type Option func(*Adapter)

// WithPartyPortalProvisioner memasang port penautan akun VENDOR/SUPPLIER terhadap Enterprise Foundation.
func WithPartyPortalProvisioner(provisioner PartyPortalProvisioner) Option {
	return func(a *Adapter) {
		a.partyProvisioner = provisioner
	}
}

// WithWorkerPrincipalProvisioner memasang port penautan akun EMPLOYEE terhadap Human Resource Management.
func WithWorkerPrincipalProvisioner(provisioner WorkerPrincipalProvisioner) Option {
	return func(a *Adapter) {
		a.workerProvisioner = provisioner
	}
}

// Adapter menyatukan panggilan gRPC IAM, manajemen sesi/CSRF, serta penegakan izin RBAC Gateway.
type Adapter struct {
	clients           source_connections.IAMClientSet
	authManager       *middleware.AuthManager
	authorizer        *middleware.Authorizer
	redirectValidator *browserauth.RedirectValidator
	partyProvisioner  PartyPortalProvisioner
	workerProvisioner WorkerPrincipalProvisioner
	cfg               Config
}

// NewAdapter membentuk Adapter domain IAM dengan pemeriksaan ketergantungan wajib fail-closed.
func NewAdapter(
	clients source_connections.IAMClientSet,
	authManager *middleware.AuthManager,
	authorizer *middleware.Authorizer,
	redirectValidator *browserauth.RedirectValidator,
	cfg Config,
	opts ...Option,
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
	adp := &Adapter{
		clients:           clients,
		authManager:       authManager,
		authorizer:        authorizer,
		redirectValidator: redirectValidator,
		cfg:               cfg,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(adp)
		}
	}
	return adp, nil
}

func (a *Adapter) ensureAuthDeps() error {
	if a == nil || a.authManager == nil {
		return middleware.ErrUnavailable("identity service unavailable", source_connections.ErrIAMUnavailable)
	}
	return nil
}

func (a *Adapter) ensureAuthorizerDeps() error {
	if a == nil || a.authManager == nil || a.authorizer == nil {
		return middleware.ErrUnavailable("identity service unavailable", source_connections.ErrIAMUnavailable)
	}
	return nil
}

func (a *Adapter) ensureAuthFlowDeps() error {
	if a == nil || a.authManager == nil || a.redirectValidator == nil {
		return middleware.ErrUnavailable("identity service unavailable", source_connections.ErrIAMUnavailable)
	}
	return nil
}

func callIAM[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	stubAvailable func(source_connections.IAMClientSet) bool,
	call func(context.Context, source_connections.IAMClientSet) (*Resp, error),
) (*Resp, error) {
	if a == nil {
		return nil, middleware.ErrUnavailable("identity service unavailable", source_connections.ErrIAMUnavailable)
	}
	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}
	if caller != nil {
		callCtx = authcontext.WithCaller(callCtx, *caller)
	}
	ready := stubAvailable != nil && stubAvailable(a.clients)
	return source_connections.InvokeIAM(callCtx, a.clients, ready, a.cfg.RPCTimeout, call)
}

func callIAMAuth[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.AuthServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Auth != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Auth)
	})
}

func callIAMUsers[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.UserServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Users != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Users)
	})
}

func callIAMSessions[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.SessionServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Sessions != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Sessions)
	})
}

func callIAMRoles[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.RoleServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Roles != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Roles)
	})
}

func callIAMPermission[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.PermissionServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Permission != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Permission)
	})
}

func callIAMGroups[Resp any](
	ctx context.Context,
	a *Adapter,
	caller *authcontext.CallerContext,
	call func(context.Context, iamv1.GroupServiceClient) (*Resp, error),
) (*Resp, error) {
	return callIAM(ctx, a, caller, func(c source_connections.IAMClientSet) bool { return c.Groups != nil }, func(rpcCtx context.Context, c source_connections.IAMClientSet) (*Resp, error) {
		return call(rpcCtx, c.Groups)
	})
}

func (a *Adapter) requirePerm(ctx context.Context, permCode string, requireSensitive bool) (*middleware.ResolvedSession, error) {
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("authorizer unavailable", nil)
	}
	return a.authorizer.RequirePermission(ctx, permCode, authz.ScopeGrant{Type: authz.ScopeGlobal}, requireSensitive)
}

// requireSelfOrPerm memisahkan evaluasi otorisasi antara target akun milik sendiri (*self) dan akses administratif lintas pengguna.
func (a *Adapter) requireSelfOrPerm(
	ctx context.Context,
	targetUserID string,
	selfPermCode string,
	adminPermCode string,
) (*middleware.ResolvedSession, error) {
	if a == nil || a.authManager == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("authorizer unavailable", nil)
	}
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
	if a == nil || a.authManager == nil {
		return middleware.ErrUnavailable("rate limiter unavailable", nil)
	}
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
	if a == nil || a.authManager == nil {
		return middleware.ErrUnavailable("rate limiter unavailable", nil)
	}
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
	if a == nil || a.authManager == nil {
		return middleware.ErrUnavailable("rate limiter unavailable", nil)
	}
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

func (a *Adapter) checkUsernameRateLimit(ctx context.Context) error {
	meta, _ := middleware.RequestMetadataFromContext(ctx)
	ip := strings.TrimSpace(meta.ClientIP)
	if ip == "" {
		ip = "unknown"
	}
	return a.authManager.CheckRateLimit(ctx, "check_username_ip", ip, a.cfg.LoginRateMax, a.cfg.LoginRateWindow)
}

func unwrapGatewayCause(err error) error {
	if gwErr, ok := middleware.AsGatewayError(err); ok && gwErr.Err != nil {
		return gwErr.Err
	}
	return err
}

// normalizeAuthError menormalkan error autentikasi/MFA agar tidak membocorkan status akun manusia, lockout, atau replay,
// sambil menegakkan pesan kontrak eksplisit "system cannot login" saat akun layanan system dicoba untuk login.
func normalizeAuthError(err error, ambiguousMsg string) error {
	cause := unwrapGatewayCause(err)
	if source_connections.ExtractIAMErrorReason(err) == "iam.user.system_cannot_login" {
		return middleware.ErrForbidden("system cannot login", cause)
	}
	mapped := mapRPCError(err)
	if gwErr, ok := middleware.AsGatewayError(mapped); ok {
		switch gwErr.Code {
		case middleware.CodeInvalidArgument, middleware.CodeUnavailable, middleware.CodeInternalError:
			return mapped
		}
	}
	return middleware.ErrUnauthenticated(ambiguousMsg, cause)
}

func mapRPCError(err error) error {
	return source_connections.MapIAMGRPCError(err)
}
