// Package security mengadaptasi proyeksi IAM menjadi keputusan akses service HRM.
package security

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
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

const sensitiveReauthWindow = 10 * time.Minute

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
func (p *Policy) Authorize(ctx context.Context, access operation.Access) (operation.Actor, error) {
	actor, _, err := p.authorize(ctx, access)
	return actor, err
}

func (p *Policy) authorize(ctx context.Context, access operation.Access) (operation.Actor, *authsession.SessionValidationContext, error) {
	if p == nil || p.sessions == nil {
		return operation.Actor{}, nil, hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return operation.Actor{}, nil, err
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
	}
	hashes := md.Get("x-ig-session-token-hash")
	if len(hashes) != 1 || len(hashes[0]) != 64 {
		return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
	}
	if _, err := hex.DecodeString(hashes[0]); err != nil {
		return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
	}
	now := time.Now()
	state, err := p.sessions.ReadSessionContext(ctx, hashes[0], now)
	if err != nil {
		if ctx.Err() != nil {
			return operation.Actor{}, nil, ctx.Err()
		}
		for _, denied := range []error{
			authsession.ErrSessionNotFound, authsession.ErrSessionRevoked,
			authsession.ErrSessionExpired, authsession.ErrSessionNotActive,
			authsession.ErrUserSecurityBlocked,
		} {
			if errors.Is(err, denied) {
				return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
			}
		}
		return operation.Actor{}, nil, fmt.Errorf("%w: security state could not be verified", hrerrors.ErrUnavailable)
	}
	if state == nil || state.Token == nil {
		return operation.Actor{}, nil, hrerrors.ErrUnauthenticated
	}
	if state.Security == nil || state.Snapshot == nil {
		return operation.Actor{}, nil, hrerrors.ErrUnavailable
	}
	if state.Gate == nil || !state.Gate.Open || !state.Snapshot.IsFresh(now) {
		return operation.Actor{}, nil, hrerrors.ErrUnavailable
	}
	if caller.UserKind != "employee" || state.Token.UserKind != caller.UserKind ||
		state.Snapshot.UserKind != caller.UserKind {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	matchedSnapshot := state.Snapshot.PrincipalID == caller.PrincipalID &&
		state.Snapshot.AuthEpoch == caller.AuthEpoch && state.Snapshot.PermRevision == caller.PermRevision
	matchedRecovery := state.Gate.RecoveryEpoch == state.Security.RecoveryEpoch &&
		state.Gate.RecoveryEpoch == state.Snapshot.RecoveryEpoch
	if !matchedSnapshot || !matchedRecovery || state.Security.UserID != caller.PrincipalID {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	matchedIdentity := state.Token.UserID == caller.PrincipalID && state.Token.SessionID == caller.SessionID
	matchedRevision := state.Security.AuthEpoch == caller.AuthEpoch && state.Security.PermRevision == caller.PermRevision
	if !matchedIdentity || !matchedRevision {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	permission, exists := p.permissions[access.Resource+"."+access.Action]
	if !exists {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	sensitive := permissionSensitive(state.Snapshot, permission)
	grant, allowed := state.Snapshot.HasPermission(permission)
	payslipSelfRead := access.Resource == "hr_payslips" && access.Action == "read_self"
	requestSelfCancel := access.Resource == "hr_service_requests" && access.Action == "cancel"
	if payslipSelfRead {
		allowed = false
	}
	if !allowed && (payslipSelfRead || requestSelfCancel) {
		for _, candidate := range state.Snapshot.Permissions {
			if strings.EqualFold(candidate.Code, permission) && candidate.Scope.Type == authz.ScopeSelf &&
				candidate.Scope.Matches(authz.ScopeGrant{Type: authz.ScopeSelf, ID: caller.PrincipalID}) &&
				grantUsable(candidate, state.Token, access, now, sensitive) {
				grant, allowed = candidate, true
				break
			}
		}
	}
	if column := legalEntityReadColumn(access); !allowed && column != "" {
		grant, allowed = scopedResourceGrant(state.Snapshot, permission, "legal_entity", column, access, state.Token, now, sensitive)
	}
	if access.Resource == "hr_requisitions" && access.Action == "approve" {
		// Otorisasi awal mendahului pembacaan. Command memeriksa unit record di dalam transaksi.
		grant, allowed = scopedResourceGrant(state.Snapshot, permission, "organization_unit",
			"organization_unit_reference_id", access, state.Token, now, sensitive)
	}
	if access.Resource == "hr_payslips" && access.Action == "reissue" {
		grant, allowed = scopedResourceGrant(state.Snapshot, permission, "legal_entity",
			"employing_legal_entity_reference_id", access, state.Token, now, sensitive)
	}
	if !allowed {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	if !grantUsable(grant, state.Token, access, now, sensitive) {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	return operation.Actor{
		PrincipalID:        caller.PrincipalID,
		SessionID:          caller.SessionID,
		Kind:               "human",
		AuthEpoch:          caller.AuthEpoch,
		PermissionRevision: caller.PermRevision,
	}, state, nil
}

// ReadScope menghasilkan filter SQL sebelum pagination dan count untuk resource dengan scope badan hukum.
func (p *Policy) ReadScope(ctx context.Context, access operation.Access) (operation.Scope, error) {
	_, state, err := p.authorize(ctx, access)
	if err != nil {
		return operation.Scope{}, err
	}
	permission := p.permissions[access.Resource+"."+access.Action]
	sensitive := permissionSensitive(state.Snapshot, permission)
	now := time.Now()
	if grant, ok := state.Snapshot.HasPermission(permission); ok && grantUsable(grant, state.Token, access, now, sensitive) {
		return operation.Scope{All: true}, nil
	}
	column := legalEntityReadColumn(access)
	if column == "" {
		return operation.Scope{}, hrerrors.ErrForbidden
	}
	ids := make([]any, 0)
	seen := make(map[string]bool)
	for _, grant := range state.Snapshot.Permissions {
		id := strings.ToLower(grant.Scope.ID)
		if !strings.EqualFold(grant.Code, permission) || grant.Scope.Type != "legal_entity" ||
			entity.ValidateUUID("scope_id", id) != nil || !grantUsable(grant, state.Token, access, now, sensitive) || seen[id] {
			continue
		}
		if len(ids) >= 1000 {
			return operation.Scope{}, hrerrors.ErrForbidden
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return operation.Scope{}, hrerrors.ErrForbidden
	}
	return operation.Scope{Filters: []domainrepo.Filter{{Column: column, Values: ids}}}, nil
}

func legalEntityReadColumn(access operation.Access) string {
	if access.Action != "read" && access.Action != "list" {
		return ""
	}
	switch access.Resource {
	case "hr_workers":
		return "primary_legal_entity_reference_id"
	case "hr_workforce_plans":
		return "legal_entity_reference_id"
	case "hr_workforce_plan_lines":
		return "plan_legal_entity_reference_id"
	default:
		return ""
	}
}

func scopedResourceGrant(snapshot *authz.AccessSnapshot, permission, scopeType, column string, access operation.Access, token *authsession.SessionTokenRecord, now time.Time, sensitive bool) (authz.PermissionGrant, bool) {
	if snapshot == nil || !strings.EqualFold(snapshot.Status, authz.StatusActive) {
		return authz.PermissionGrant{}, false
	}
	id, present := access.Attributes[column]
	if !present && len(access.Attributes) != 0 {
		return authz.PermissionGrant{}, false
	}
	if present {
		if entity.ValidateUUID(column, id) != nil {
			return authz.PermissionGrant{}, false
		}
	}
	for _, grant := range snapshot.Permissions {
		if !strings.EqualFold(grant.Code, permission) || grant.Scope.Type != scopeType ||
			entity.ValidateUUID("scope_id", grant.Scope.ID) != nil ||
			(present && !strings.EqualFold(grant.Scope.ID, id)) || !grantUsable(grant, token, access, now, sensitive) {
			continue
		}
		return grant, true
	}
	return authz.PermissionGrant{}, false
}

func permissionSensitive(snapshot *authz.AccessSnapshot, permission string) bool {
	for _, grant := range snapshot.Permissions {
		if strings.EqualFold(grant.Code, permission) && grant.IsSensitive {
			return true
		}
	}
	return false
}

func grantUsable(grant authz.PermissionGrant, token *authsession.SessionTokenRecord, access operation.Access, now time.Time, sensitive bool) bool {
	if !sensitive && !grant.IsSensitive && !IsSensitiveOperation(access.Resource, access.Action) {
		return true
	}
	if token == nil || !token.MFAAuthenticated || token.IssuedAtUnix <= 0 {
		return false
	}
	age := now.Sub(time.Unix(token.IssuedAtUnix, 0))
	return age >= 0 && age <= sensitiveReauthWindow
}
