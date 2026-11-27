// Package redisstore mengimplementasikan proyektor keamanan Redis, koordinator outbox pending/effective, dan worker pemulihan gate IAM.
package redisstore

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

var _ domainservice.RuntimeSecurityProjector = (*Projector)(nil)

// Projector memproyeksikan state autentikasi, sesi, epoch keamanan, dan AccessSnapshot ke Redis menggunakan peran iam_writer.
type Projector struct {
	store *authsession.Store
	clock domainservice.Clock
}

// NewProjector membangun Projector Redis aman. Bila store nil, proyektor berjalan nonaktif untuk unit test murni.
func NewProjector(store *authsession.Store, clock domainservice.Clock) *Projector {
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	return &Projector{
		store: store,
		clock: clock,
	}
}

// IsEnabled melaporkan apakah proyektor terhubung ke Redis keamanan.
func (p *Projector) IsEnabled() bool {
	return p != nil && p.store != nil
}

// SyncGate memperbarui status gerbang keamanan IAM di Redis sesuai recovery_epoch PG.
func (p *Projector) SyncGate(ctx context.Context, open bool, recoveryEpoch int64, source string) (int64, error) {
	if !p.IsEnabled() {
		return recoveryEpoch, nil
	}
	epoch, err := p.store.SetGate(ctx, open, recoveryEpoch, p.effectiveClock().Now(), source)
	if err != nil {
		return 0, mapRedisError("sync gate", err)
	}
	return epoch, nil
}

// GetGate membaca status gerbang keamanan terkini dari Redis.
func (p *Projector) GetGate(ctx context.Context) (*authsession.SecurityGateRecord, error) {
	if !p.IsEnabled() {
		return nil, nil
	}
	gate, err := p.store.GetGate(ctx)
	if err != nil {
		return nil, mapRedisError("read gate", err)
	}
	return gate, nil
}

// ProjectSessionIssued menulis pasangan token sesi, jendela aktivitas, dan user:sec ke Redis secara atomik.
func (p *Projector) ProjectSessionIssued(ctx context.Context, user *entity.User, sess *entity.Session, recoveryEpoch int64) error {
	return p.projectSession(ctx, user, sess, recoveryEpoch)
}

// ProjectSessionUpdated memperbarui proyeksi sesi dan generasi terbaru ke Redis (misal setelah MFA step-up).
func (p *Projector) ProjectSessionUpdated(ctx context.Context, user *entity.User, sess *entity.Session, recoveryEpoch int64) error {
	return p.projectSession(ctx, user, sess, recoveryEpoch)
}

func (p *Projector) projectSession(ctx context.Context, user *entity.User, sess *entity.Session, recoveryEpoch int64) error {
	if !p.IsEnabled() || user == nil || sess == nil {
		return nil
	}
	if sess.EndedAt != nil {
		return p.ProjectSessionRevoked(ctx, user, sess.TokenDigest(), recoveryEpoch)
	}

	now := p.effectiveClock().Now()
	tokDigestHex := hex.EncodeToString(sess.TokenDigest())
	csrfDigestHex := hex.EncodeToString(sess.CSRFTokenDigest())
	if tokDigestHex == "" || csrfDigestHex == "" {
		return iamerrors.NewValidation("session_digests", iamerrors.Code("iam.session.missing_digests"), "session digests are required for redis projection")
	}

	gen := sess.Generation
	if gen < 1 {
		gen = 1
	}
	csrfVer := sess.CSRFKeyVersion
	if csrfVer < 1 {
		csrfVer = 1
	}
	authEpoch := user.AuthEpoch
	if authEpoch < 1 {
		authEpoch = 1
	}
	permRev := user.PermRevision
	if permRev < 1 {
		permRev = 1
	}
	if recoveryEpoch < 1 {
		recoveryEpoch = 1
	}

	// Gunakan MFAVerifiedAt bila lebih baru dari AuthenticatedAt agar jendela re-autentikasi 10 menit di Redis selaras dengan IsRecentAuth.
	effectiveAuthAt := sess.AuthenticatedAt
	if sess.MFAVerified && sess.MFAVerifiedAt != nil && sess.MFAVerifiedAt.After(effectiveAuthAt) {
		effectiveAuthAt = *sess.MFAVerifiedAt
	}

	tokRec := authsession.SessionTokenRecord{
		SessionID:        sess.ID,
		UserID:           sess.UserID,
		TokenHashHex:     tokDigestHex,
		CSRFDigestHex:    csrfDigestHex,
		CSRFKeyVersion:   csrfVer,
		UserKind:         user.Kind.String(),
		Generation:       gen,
		AuthEpoch:        authEpoch,
		IssuedAtUnix:     effectiveAuthAt.Unix(),
		AbsExpiresAtUnix: sess.AbsoluteExpiresAt.Unix(),
		MFAAuthenticated: sess.MFAVerified,
		Revoked:          false,
	}
	actRec := authsession.SessionActivityRecord{
		SessionID:       sess.ID,
		UserID:          sess.UserID,
		Generation:      gen,
		AuthEpoch:       authEpoch,
		LastSeenUnix:    sess.LastActiveAt.Unix(),
		IdleExpiresUnix: sess.IdleExpiresAt.Unix(),
	}
	secRec := authsession.UserSecurityRecord{
		UserID:        user.ID,
		Status:        user.Status.String(),
		AuthEpoch:     authEpoch,
		PermRevision:  permRev,
		RecoveryEpoch: recoveryEpoch,
		UpdatedAtUnix: now.Unix(),
	}

	return mapRedisError("write session", p.store.WriteSession(ctx, tokRec, actRec, secRec, now))
}

// ProjectSessionRevoked menghapus satu token sesi dari Redis dan menjaga user:sec tetap sinkron.
func (p *Projector) ProjectSessionRevoked(ctx context.Context, user *entity.User, tokenDigest []byte, recoveryEpoch int64) error {
	if !p.IsEnabled() {
		return nil
	}
	now := p.effectiveClock().Now()
	var hashes []string
	if len(tokenDigest) > 0 {
		hashes = []string{hex.EncodeToString(tokenDigest)}
	}
	var secPtr *authsession.UserSecurityRecord
	if user != nil {
		rec := buildUserSecurityRecord(user, recoveryEpoch, now)
		secPtr = &rec
	}
	return mapRedisError("revoke session", p.store.RevokeSessionTokens(ctx, hashes, secPtr, false, now))
}

// ProjectUserSessionsRevoked menghapus seluruh sesi user di Redis sekaligus menaikkan auth_epoch pada user:sec.
func (p *Projector) ProjectUserSessionsRevoked(ctx context.Context, user *entity.User, tokenDigests [][]byte, recoveryEpoch int64) error {
	if !p.IsEnabled() {
		return nil
	}
	now := p.effectiveClock().Now()
	hashes := make([]string, 0, len(tokenDigests))
	for _, d := range tokenDigests {
		if len(d) > 0 {
			hashes = append(hashes, hex.EncodeToString(d))
		}
	}
	var secPtr *authsession.UserSecurityRecord
	if user != nil {
		rec := buildUserSecurityRecord(user, recoveryEpoch, now)
		secPtr = &rec
	}
	return mapRedisError("revoke user sessions", p.store.RevokeSessionTokens(ctx, hashes, secPtr, true, now))
}

// ProjectUserSecurityChanged memperbarui status user:sec secara monotonik dan menghapus AccessSnapshot lama bila diminta.
func (p *Projector) ProjectUserSecurityChanged(ctx context.Context, user *entity.User, recoveryEpoch int64, invalidateAuthz bool) error {
	if !p.IsEnabled() || user == nil {
		return nil
	}
	now := p.effectiveClock().Now()
	sec := buildUserSecurityRecord(user, recoveryEpoch, now)
	_, err := p.store.BumpUserSecurity(ctx, sec, invalidateAuthz, now)
	return mapRedisError("bump user security", err)
}

// ProjectAccessSnapshot menulis AccessSnapshot terkini (TTL 35s, freshness 30s) ke Redis untuk dikonsumsi Gateway.
func (p *Projector) ProjectAccessSnapshot(
	ctx context.Context,
	snapshot *entity.AccessSnapshot,
	canAccessIdentity bool,
	defaultRedirectPath string,
	allowedRedirectPaths []string,
) error {
	if !p.IsEnabled() || snapshot == nil {
		return nil
	}

	now := p.effectiveClock().Now()
	roles := make([]authz.RoleGrant, 0, len(snapshot.Roles))
	for _, r := range snapshot.Roles {
		roles = append(roles, authz.RoleGrant{
			ID:                  r.ID,
			Code:                r.Code,
			Name:                r.Name,
			Priority:            int32(r.RedirectPriority),
			DefaultRedirectPath: r.DefaultRedirectPath,
		})
	}
	perms := make([]authz.PermissionGrant, 0, len(snapshot.Permissions))
	for _, pr := range snapshot.Permissions {
		scopeID := ""
		if pr.Scope.ID != nil {
			scopeID = *pr.Scope.ID
		}
		perms = append(perms, authz.PermissionGrant{
			Code:        pr.PermissionCode,
			IsSensitive: pr.IsSensitive,
			Scope: authz.ScopeGrant{
				Type: pr.Scope.Type.String(),
				ID:   scopeID,
			},
		})
	}

	authEpoch := snapshot.AuthEpoch
	if authEpoch < 1 {
		authEpoch = 1
	}
	permRev := snapshot.PermRevision
	if permRev < 1 {
		permRev = 1
	}
	recEpoch := snapshot.RecoveryEpoch
	if recEpoch < 1 {
		recEpoch = 1
	}

	out := &authz.AccessSnapshot{
		PrincipalID:          snapshot.UserID,
		UserKind:             snapshot.UserKind.String(),
		Status:               snapshot.UserStatus.String(),
		MFAEnabled:           snapshot.MFAEnabled,
		AuthEpoch:            authEpoch,
		PermRevision:         permRev,
		RecoveryEpoch:        recEpoch,
		Roles:                roles,
		Permissions:          perms,
		CanAccessIdentity:    canAccessIdentity,
		DefaultRedirectPath:  strings.TrimSpace(defaultRedirectPath),
		AllowedRedirectPaths: append([]string(nil), allowedRedirectPaths...),
		SyncedAtUnix:         now.Unix(),
		FreshUntilUnix:       now.Add(authz.DefaultSnapshotMaxAge).Unix(),
	}
	return mapRedisError("project access snapshot", p.store.ProjectAccessSnapshot(ctx, out, authz.DefaultSnapshotMaxAge+5*time.Second, now))
}

func (p *Projector) effectiveClock() domainservice.Clock {
	if p != nil && p.clock != nil {
		return p.clock
	}
	return domainservice.SystemClock{}
}

func mapRedisError(op string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*iamerrors.DomainError](err); ok {
		return err
	}
	return iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.ErrUnavailable.Code, fmt.Sprintf("iam redis %s failed", op))
}

func buildUserSecurityRecord(user *entity.User, recoveryEpoch int64, now time.Time) authsession.UserSecurityRecord {
	authEpoch := user.AuthEpoch
	if authEpoch < 1 {
		authEpoch = 1
	}
	permRev := user.PermRevision
	if permRev < 1 {
		permRev = 1
	}
	if recoveryEpoch < 1 {
		recoveryEpoch = 1
	}
	status := user.Status.String()
	if user.IsDeleted() {
		status = entity.UserStatusDisabled.String()
	}
	return authsession.UserSecurityRecord{
		UserID:        user.ID,
		Status:        status,
		AuthEpoch:     authEpoch,
		PermRevision:  permRev,
		RecoveryEpoch: recoveryEpoch,
		UpdatedAtUnix: now.Unix(),
	}
}
