// Package security mengadaptasi proyeksi IAM menjadi keputusan akses service HRM.
package security

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"google.golang.org/grpc/metadata"
)

// SessionReader membaca proyeksi keamanan melalui adapter IAM bersama.
type SessionReader interface {
	ReadSessionContext(context.Context, string, time.Time) (*authsession.SessionValidationContext, error)
}

// Policy hanya memakai mapping permission yang diberikan composition root.
type Policy struct {
	sessions    SessionReader
	permissions map[string]string
}

// NewPolicy menyalin katalog agar perubahan map pemanggil tidak mengubah izin runtime.
func NewPolicy(sessions SessionReader, permissions map[string]string) (*Policy, error) {
	if err := operation.RequireDependency(sessions); err != nil {
		return nil, err
	}
	catalog := make(map[string]string, len(permissions))
	for operation, permission := range permissions {
		if strings.TrimSpace(operation) == "" || strings.TrimSpace(permission) == "" {
			return nil, hrerrors.ErrInvalidInput
		}
		catalog[operation] = permission
	}
	return &Policy{sessions: sessions, permissions: catalog}, nil
}

// Authorize memeriksa ulang sesi, revoke, revisi dan permission pada setiap operasi.
// Scope non-global ditolak sampai pemetaan atribut/relasi domain tersedia secara lengkap.
func (p *Policy) Authorize(ctx context.Context, access operation.Access) (operation.Actor, error) {
	if p == nil || p.sessions == nil {
		return operation.Actor{}, hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return operation.Actor{}, err
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return operation.Actor{}, hrerrors.ErrUnauthenticated
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return operation.Actor{}, hrerrors.ErrUnauthenticated
	}
	hashes := md.Get("x-ig-session-token-hash")
	if len(hashes) != 1 || len(hashes[0]) != 64 {
		return operation.Actor{}, hrerrors.ErrUnauthenticated
	}
	if _, err := hex.DecodeString(hashes[0]); err != nil {
		return operation.Actor{}, hrerrors.ErrUnauthenticated
	}
	now := time.Now()
	state, err := p.sessions.ReadSessionContext(ctx, hashes[0], now)
	if err != nil {
		if ctx.Err() != nil {
			return operation.Actor{}, ctx.Err()
		}
		for _, denied := range []error{
			authsession.ErrSessionNotFound, authsession.ErrSessionRevoked,
			authsession.ErrSessionExpired, authsession.ErrSessionNotActive,
			authsession.ErrUserSecurityBlocked,
		} {
			if errors.Is(err, denied) {
				return operation.Actor{}, hrerrors.ErrUnauthenticated
			}
		}
		return operation.Actor{}, fmt.Errorf("%w: security state could not be verified", hrerrors.ErrUnavailable)
	}
	if state == nil || state.Token == nil {
		return operation.Actor{}, hrerrors.ErrUnauthenticated
	}
	if state.Security == nil || state.Snapshot == nil {
		return operation.Actor{}, hrerrors.ErrUnavailable
	}
	if state.Gate == nil || !state.Gate.Open || !state.Snapshot.IsFresh(now) {
		return operation.Actor{}, hrerrors.ErrUnavailable
	}
	matchedSnapshot := state.Snapshot.PrincipalID == caller.PrincipalID &&
		state.Snapshot.AuthEpoch == caller.AuthEpoch && state.Snapshot.PermRevision == caller.PermRevision
	matchedRecovery := state.Gate.RecoveryEpoch == state.Security.RecoveryEpoch &&
		state.Gate.RecoveryEpoch == state.Snapshot.RecoveryEpoch
	if !matchedSnapshot || !matchedRecovery || state.Security.UserID != caller.PrincipalID {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	matchedIdentity := state.Token.UserID == caller.PrincipalID && state.Token.SessionID == caller.SessionID
	matchedRevision := state.Security.AuthEpoch == caller.AuthEpoch && state.Security.PermRevision == caller.PermRevision
	if !matchedIdentity || !matchedRevision {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	permission, exists := p.permissions[access.Resource+"."+access.Action]
	if !exists {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	grant, allowed := state.Snapshot.HasPermission(permission)
	if !allowed || (grant.IsSensitive && !state.Token.MFAAuthenticated) {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	return operation.Actor{
		PrincipalID:        caller.PrincipalID,
		SessionID:          caller.SessionID,
		Kind:               "human",
		AuthEpoch:          caller.AuthEpoch,
		PermissionRevision: caller.PermRevision,
	}, nil
}

// ReadScope hanya menyatakan seluruh baris setelah permission global terbukti.
func (p *Policy) ReadScope(ctx context.Context, access operation.Access) (operation.Scope, error) {
	if _, err := p.Authorize(ctx, access); err != nil {
		return operation.Scope{}, err
	}
	return operation.Scope{All: true}, nil
}
