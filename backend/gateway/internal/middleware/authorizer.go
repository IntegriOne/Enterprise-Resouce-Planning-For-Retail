package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// SensitiveSessionChecker melakukan pengecekan otoritatif ke layanan IAM untuk sesi yang berpotensi melewati jendela reauth 10 menit.
type SensitiveSessionChecker interface {
	VerifySensitiveSession(ctx context.Context, rawToken, requiredPermission string, requiredScope authz.ScopeGrant, caller authcontext.CallerContext) error
}

// Authorizer mengevaluasi RBAC berscope dan kewajiban re-autentikasi sensitif (D08) secara fail-closed di Gateway.
type Authorizer struct {
	authManager        *AuthManager
	sensitiveChecker   SensitiveSessionChecker
	sensitiveReauthAge time.Duration
	nowF               func() time.Time
}

// NewAuthorizer membentuk Authorizer berbasis AuthManager Gateway dan batas usia re-autentikasi sensitif.
func NewAuthorizer(authManager *AuthManager, checker SensitiveSessionChecker, sensitiveReauthAge time.Duration) *Authorizer {
	if sensitiveReauthAge <= 0 {
		sensitiveReauthAge = 10 * time.Minute
	}
	return &Authorizer{
		authManager:        authManager,
		sensitiveChecker:   checker,
		sensitiveReauthAge: sensitiveReauthAge,
		nowF:               time.Now,
	}
}

// RequirePermission memastikan pemanggil terautentikasi, snapshot segar <= 30s, memiliki izin berscope,
// serta memenuhi jendela re-autentikasi & MFA bila izin atau operasi bersifat sensitif.
func (a *Authorizer) RequirePermission(
	ctx context.Context,
	permCode string,
	scope authz.ScopeGrant,
	requireSensitiveReauth bool,
) (*ResolvedSession, error) {
	if a == nil || a.authManager == nil {
		return nil, ErrUnavailable("authorizer unavailable", nil)
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if sess == nil || sess.Token == nil || sess.Snapshot == nil {
		return nil, ErrUnauthenticated("authentication required", nil)
	}

	if scope.Type == "" {
		scope.Type = authz.ScopeGlobal
	}
	grant, ok := sess.Snapshot.HasScopedPermission(permCode, scope)
	if !ok {
		return nil, ErrForbidden("missing required permission: "+strings.TrimSpace(permCode), nil)
	}

	if grant.IsSensitive || requireSensitiveReauth {
		if err := a.enforceSensitiveReauth(ctx, sess, grant.Code, scope); err != nil {
			return nil, err
		}
	}
	return sess, nil
}

// RequireSelfPermission menegakkan otorisasi aksi mandiri (*self) terikat tegas pada cakupan self milik principal aktif.
func (a *Authorizer) RequireSelfPermission(
	ctx context.Context,
	permCode string,
	requireSensitiveReauth bool,
) (*ResolvedSession, error) {
	if a == nil || a.authManager == nil {
		return nil, ErrUnavailable("authorizer unavailable", nil)
	}
	cleanPerm := strings.TrimSpace(permCode)
	if cleanPerm == "" {
		return nil, ErrForbidden("missing required self permission", nil)
	}
	sess, err := a.authManager.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if sess.Token == nil || strings.TrimSpace(sess.Token.UserID) == "" {
		return nil, ErrUnauthenticated("session missing principal identity", nil)
	}
	return a.RequirePermission(ctx, cleanPerm, authz.ScopeGrant{Type: authz.ScopeSelf, ID: sess.Token.UserID}, requireSensitiveReauth)
}

// RequireSelfSensitive memastikan principal aktif memenuhi izin *self beserta re-autentikasi terkini (10 menit) & MFA bila aktif.
func (a *Authorizer) RequireSelfSensitive(ctx context.Context, permCode string) (*ResolvedSession, error) {
	return a.RequireSelfPermission(ctx, permCode, true)
}

func (a *Authorizer) enforceSensitiveReauth(
	ctx context.Context,
	sess *ResolvedSession,
	permCode string,
	scope authz.ScopeGrant,
) error {
	if sess == nil || sess.Token == nil || sess.Snapshot == nil {
		return ErrUnauthenticated("authentication required", nil)
	}
	if sess.Snapshot.MFAEnabled && !sess.Token.MFAAuthenticated {
		return a.verifySensitiveWithChecker(ctx, sess, permCode, scope, "mfa step-up authentication required for sensitive action")
	}

	now := time.Now()
	if a.nowF != nil {
		now = a.nowF()
	}
	if sess.Token.IssuedAtUnix <= 0 || now.Sub(time.Unix(sess.Token.IssuedAtUnix, 0)) > a.sensitiveReauthAge {
		return a.verifySensitiveWithChecker(ctx, sess, permCode, scope, "recent authentication within 10 minutes is required for sensitive action")
	}
	return nil
}

func (a *Authorizer) verifySensitiveWithChecker(
	ctx context.Context,
	sess *ResolvedSession,
	permCode string,
	scope authz.ScopeGrant,
	forbiddenMsg string,
) error {
	if a.sensitiveChecker == nil {
		return ErrForbidden(forbiddenMsg, nil)
	}
	err := a.sensitiveChecker.VerifySensitiveSession(ctx, sess.RawToken, permCode, scope, sess.Caller)
	if err == nil {
		return nil
	}
	// Kegagalan konektivitas atau galat internal IAM tidak boleh dimask menjadi penolakan otorisasi FORBIDDEN.
	if gwErr, ok := AsGatewayError(err); ok && (gwErr.Code == CodeUnavailable || gwErr.Code == CodeInternalError) {
		return err
	}
	return ErrForbidden(forbiddenMsg, err)
}
