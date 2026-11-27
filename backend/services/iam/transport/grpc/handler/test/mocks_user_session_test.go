package handler_test

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
)

// --- UserRepository ---

type memUserRepo struct{ s *memStore }

func (r *memUserRepo) SelectByID(_ context.Context, id string) (*entity.User, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	u, ok := r.s.users[id]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *memUserRepo) SelectList(_ context.Context, filter domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.User
	for _, u := range r.s.users {
		if !filter.IncludeDeleted && u.IsDeleted() {
			continue
		}
		if filter.Kind != nil && u.Kind != *filter.Kind {
			continue
		}
		if filter.Status != nil && u.Status != *filter.Status {
			continue
		}
		if filter.Query != "" && !strings.Contains(strings.ToLower(u.DisplayName), strings.ToLower(filter.Query)) {
			continue
		}
		cp := *u
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.User) int { return strings.Compare(a.ID, b.ID) })
	return paginateSlice(out, filter.Page), nil
}

func (r *memUserRepo) SelectAccessSnapshotByUserID(ctx context.Context, userID string, now time.Time) (*entity.AccessSnapshot, error) {
	r.s.mu.RLock()
	u, ok := r.s.users[userID]
	if !ok {
		r.s.mu.RUnlock()
		return nil, iamerrors.ErrUserNotFound
	}
	var mfaEnabled bool
	if f, ok := r.s.totpFactors[userID]; ok && f.IsEnabled {
		mfaEnabled = true
	}
	r.s.mu.RUnlock()

	roles, err := (&memRoleRepo{r.s}).SelectEffectiveByUserID(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var perms []entity.ScopedPermission
	addRolePerms := func(roleID string, scope entity.Scope) {
		for _, rp := range r.s.rolePermissions {
			if rp.RoleID == roleID {
				if p, ok := r.s.permissions[rp.PermissionID]; ok {
					perms = append(perms, entity.ScopedPermission{
						PermissionCode: p.Code,
						IsSensitive:    p.IsSensitive,
						Scope:          scope,
					})
				}
			}
		}
	}
	for _, ur := range r.s.userRoles {
		if ur.UserID == userID && ur.IsEffective(now) {
			if role, ok := r.s.roles[ur.RoleID]; ok && role.IsActive {
				addRolePerms(role.ID, ur.Scope)
			}
		}
	}
	userGroups := make(map[string]struct{})
	for _, gm := range r.s.groupMembers {
		if gm.UserID == userID {
			if g, ok := r.s.groups[gm.GroupID]; ok && g.IsActive {
				userGroups[g.ID] = struct{}{}
			}
		}
	}
	for _, gr := range r.s.groupRoles {
		if _, inGroup := userGroups[gr.GroupID]; inGroup && gr.IsEffective(now) {
			if role, ok := r.s.roles[gr.RoleID]; ok && role.IsActive {
				addRolePerms(role.ID, gr.Scope)
			}
		}
	}

	return &entity.AccessSnapshot{
		UserID:      u.ID,
		UserKind:    u.Kind,
		UserStatus:  u.Status,
		MFAEnabled:  mfaEnabled,
		Roles:       roles,
		Permissions: perms,
		CheckedAt:   now,
	}, nil
}

func (r *memUserRepo) Insert(_ context.Context, u *entity.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *u
	r.s.users[u.ID] = &cp
	return nil
}

func (r *memUserRepo) InsertWithIdentities(_ context.Context, u *entity.User, idents []*entity.UserIdentity) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, existing := range r.s.identities {
		for _, incoming := range idents {
			if existing.Kind == incoming.Kind && existing.Value == incoming.Value {
				if incoming.Kind == entity.IdentityKindEmail {
					return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "email", "email already registered")
				}
				return iamerrors.WithField(iamerrors.ErrIdentityAlreadyExists, "username", "username already registered")
			}
		}
	}
	cp := *u
	r.s.users[u.ID] = &cp
	for _, id := range idents {
		idCopy := *id
		r.s.identities[id.ID] = &idCopy
	}
	return nil
}

func (r *memUserRepo) Update(_ context.Context, u *entity.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.users[u.ID]; !ok {
		return iamerrors.ErrUserNotFound
	}
	cp := *u
	r.s.users[u.ID] = &cp
	return nil
}

func (r *memUserRepo) UpdateLastLogin(_ context.Context, userID string, loginAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[userID]
	if !ok {
		return iamerrors.ErrUserNotFound
	}
	u.RecordLogin(loginAt)
	return nil
}

func (r *memUserRepo) SoftDelete(_ context.Context, id string, deletedAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[id]
	if !ok {
		return iamerrors.ErrUserNotFound
	}
	return u.SoftDelete(deletedAt)
}

// --- UserIdentityRepository ---

type memIdentityRepo struct{ s *memStore }

func (r *memIdentityRepo) SelectByID(_ context.Context, id string) (*entity.UserIdentity, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	ident, ok := r.s.identities[id]
	if !ok {
		return nil, iamerrors.ErrIdentityNotFound
	}
	cp := *ident
	return &cp, nil
}

func (r *memIdentityRepo) SelectByKindAndValue(_ context.Context, kind entity.IdentityKind, value string) (*entity.UserIdentity, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, ident := range r.s.identities {
		if ident.Kind == kind && ident.Value == value {
			cp := *ident
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrIdentityNotFound
}

func (r *memIdentityRepo) FindUserIdentityByEmailKind(ctx context.Context, email string) (*entity.UserIdentity, error) {
	norm, err := entity.ValidateEmail(email)
	if err != nil {
		return nil, err
	}
	return r.SelectByKindAndValue(ctx, entity.IdentityKindEmail, norm)
}

func (r *memIdentityRepo) FindUserIdentityByUsernameKind(ctx context.Context, username string) (*entity.UserIdentity, error) {
	norm, err := entity.ValidateUsername(username)
	if err != nil {
		return nil, err
	}
	return r.SelectByKindAndValue(ctx, entity.IdentityKindUsername, norm)
}

func (r *memIdentityRepo) SelectByEmailAndUsername(_ context.Context, email, username string) ([]*entity.UserIdentity, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.UserIdentity
	for _, ident := range r.s.identities {
		if (email != "" && ident.Kind == entity.IdentityKindEmail && ident.Value == email) ||
			(username != "" && ident.Kind == entity.IdentityKindUsername && ident.Value == username) {
			cp := *ident
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memIdentityRepo) SelectByUserID(_ context.Context, userID string) ([]*entity.UserIdentity, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.UserIdentity
	for _, ident := range r.s.identities {
		if ident.UserID == userID {
			cp := *ident
			out = append(out, &cp)
		}
	}
	slices.SortFunc(out, func(a, b *entity.UserIdentity) int { return strings.Compare(a.ID, b.ID) })
	return out, nil
}

func (r *memIdentityRepo) CheckExistsByKindAndValue(_ context.Context, kind entity.IdentityKind, value string) (bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, ident := range r.s.identities {
		if ident.Kind == kind && ident.Value == value {
			return true, nil
		}
	}
	return false, nil
}

func (r *memIdentityRepo) Insert(_ context.Context, ident *entity.UserIdentity) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, existing := range r.s.identities {
		if existing.Kind == ident.Kind && existing.Value == ident.Value {
			return iamerrors.ErrIdentityAlreadyExists
		}
		if existing.UserID == ident.UserID && existing.Kind == entity.IdentityKindUsername && ident.Kind == entity.IdentityKindUsername {
			return iamerrors.ErrIdentityLimitExceeded
		}
		if ident.Kind == entity.IdentityKindEmail && existing.UserID == ident.UserID && existing.Kind == entity.IdentityKindEmail && existing.IsVerified == ident.IsVerified {
			return iamerrors.ErrIdentityLimitExceeded
		}
	}
	cp := *ident
	r.s.identities[ident.ID] = &cp
	return nil
}

func (r *memIdentityRepo) Update(_ context.Context, ident *entity.UserIdentity) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.identities[ident.ID]; !ok {
		return iamerrors.ErrIdentityNotFound
	}
	for _, existing := range r.s.identities {
		if existing.ID == ident.ID {
			continue
		}
		if existing.Kind == ident.Kind && existing.Value == ident.Value {
			return iamerrors.ErrIdentityAlreadyExists
		}
		if existing.UserID == ident.UserID && existing.Kind == entity.IdentityKindUsername && ident.Kind == entity.IdentityKindUsername {
			return iamerrors.ErrIdentityLimitExceeded
		}
		if ident.Kind == entity.IdentityKindEmail && existing.UserID == ident.UserID && existing.Kind == entity.IdentityKindEmail && existing.IsVerified == ident.IsVerified {
			return iamerrors.ErrIdentityLimitExceeded
		}
	}
	cp := *ident
	r.s.identities[ident.ID] = &cp
	return nil
}

func (r *memIdentityRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.identities[id]; !ok {
		return iamerrors.ErrIdentityNotFound
	}
	delete(r.s.identities, id)
	return nil
}

// --- UserCredentialRepository, TOTPFactor, RecoveryCode, Token, Session ---

type memCredentialRepo struct{ s *memStore }

func (r *memCredentialRepo) SelectByUserID(_ context.Context, userID string) (*entity.UserCredential, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.credentials[userID]
	if !ok {
		return nil, iamerrors.ErrCredentialNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *memCredentialRepo) Insert(_ context.Context, c *entity.UserCredential) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *c
	r.s.credentials[c.UserID] = &cp
	return nil
}

func (r *memCredentialRepo) Update(_ context.Context, c *entity.UserCredential) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *c
	r.s.credentials[c.UserID] = &cp
	return nil
}

func (r *memCredentialRepo) DeleteByUserID(_ context.Context, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.credentials, userID)
	return nil
}

func (r *memCredentialRepo) RecordFailedAttemptAtomic(_ context.Context, userID string, maxAttempts int, lockDuration time.Duration, now time.Time) (int, *time.Time, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.credentials[userID]
	if !ok {
		return 0, nil, iamerrors.ErrCredentialNotFound
	}
	if c.LockedUntil != nil && !now.Before(*c.LockedUntil) {
		c.FailedLoginAttempts = 0
		c.LockedUntil = nil
	}
	c.RecordFailedAttempt(maxAttempts, lockDuration, now)
	var untilCopy *time.Time
	if c.LockedUntil != nil {
		u := *c.LockedUntil
		untilCopy = &u
	}
	return c.FailedLoginAttempts, untilCopy, nil
}

func (r *memCredentialRepo) ResetFailedAttemptsAtomic(_ context.Context, userID string, now time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.credentials[userID]
	if !ok {
		return nil
	}
	c.ResetFailedAttempts(now)
	return nil
}

type memTOTPRepo struct{ s *memStore }

func (r *memTOTPRepo) SelectByUserID(_ context.Context, userID string) (*entity.UserTOTPFactor, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	f, ok := r.s.totpFactors[userID]
	if !ok {
		return nil, iamerrors.ErrTOTPNotFound
	}
	cp := *f
	return &cp, nil
}

func (r *memTOTPRepo) Insert(_ context.Context, f *entity.UserTOTPFactor) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *f
	r.s.totpFactors[f.UserID] = &cp
	return nil
}

func (r *memTOTPRepo) Update(_ context.Context, f *entity.UserTOTPFactor) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *f
	r.s.totpFactors[f.UserID] = &cp
	return nil
}

func (r *memTOTPRepo) UpdateLastUsedStepIfGreater(_ context.Context, factorID string, step int64, updatedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, f := range r.s.totpFactors {
		if f.ID == factorID {
			if f.LastUsedStep != nil && step <= *f.LastUsedStep {
				return false, nil
			}
			if err := f.MarkStepUsed(step, updatedAt); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, iamerrors.ErrTOTPNotFound
}

func (r *memTOTPRepo) DeleteByUserID(_ context.Context, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.totpFactors, userID)
	return nil
}

type memRecoveryCodeRepo struct{ s *memStore }

func (r *memRecoveryCodeRepo) SelectUnusedByUserID(_ context.Context, userID string) ([]*entity.UserRecoveryCode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var out []*entity.UserRecoveryCode
	for _, rc := range r.s.recoveryCodes {
		if rc.UserID == userID && !rc.IsUsed() {
			cp := *rc
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memRecoveryCodeRepo) CountUnusedByUserID(_ context.Context, userID string) (int64, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var n int64
	for _, rc := range r.s.recoveryCodes {
		if rc.UserID == userID && !rc.IsUsed() {
			n++
		}
	}
	return n, nil
}

func (r *memRecoveryCodeRepo) InsertBatch(_ context.Context, codes []*entity.UserRecoveryCode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, rc := range codes {
		cp := *rc
		r.s.recoveryCodes[rc.ID] = &cp
	}
	return nil
}

func (r *memRecoveryCodeRepo) UpdateUsedAtIfUnused(_ context.Context, id string, usedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	rc, ok := r.s.recoveryCodes[id]
	if !ok || rc.IsUsed() {
		return false, nil
	}
	_ = rc.Consume(usedAt)
	return true, nil
}

func (r *memRecoveryCodeRepo) DeleteByUserID(_ context.Context, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for id, rc := range r.s.recoveryCodes {
		if rc.UserID == userID {
			delete(r.s.recoveryCodes, id)
		}
	}
	return nil
}

type memTokenRepo struct{ s *memStore }

func (r *memTokenRepo) SelectByDigestAndPurpose(_ context.Context, digest []byte, purpose entity.TokenPurpose) (*entity.UserToken, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, tok := range r.s.tokens {
		if tok.Purpose == purpose && bytes.Equal(tok.TokenDigest(), digest) {
			cp := *tok
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrTokenNotFound
}

func (r *memTokenRepo) Insert(_ context.Context, tok *entity.UserToken) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if tok.Status == entity.TokenStatusActive {
		for _, existing := range r.s.tokens {
			if existing.UserID == tok.UserID && existing.Purpose == tok.Purpose && existing.Status == entity.TokenStatusActive {
				return iamerrors.NewValidation("token", iamerrors.Code("iam.token.duplicate_active"), "only one active token per purpose allowed")
			}
		}
	}
	cp := *tok
	if tok.IdentityID != nil {
		id := *tok.IdentityID
		cp.IdentityID = &id
	}
	r.s.tokens[tok.ID] = &cp
	return nil
}

func (r *memTokenRepo) UpdateActiveStatusByUserAndPurpose(_ context.Context, userID string, purpose entity.TokenPurpose, nextStatus entity.TokenStatus) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var n int64
	for _, tok := range r.s.tokens {
		if tok.UserID == userID && tok.Purpose == purpose && tok.Status == entity.TokenStatusActive {
			tok.Status = nextStatus
			n++
		}
	}
	return n, nil
}

func (r *memTokenRepo) UpdateUsedAtIfUnused(_ context.Context, id string, usedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	tok, ok := r.s.tokens[id]
	if !ok || tok.Status != entity.TokenStatusActive || tok.UsedAt != nil {
		return false, nil
	}
	ts := usedAt
	tok.Status = entity.TokenStatusConsumed
	tok.UsedAt = &ts
	return true, nil
}

func (r *memTokenRepo) DeleteExpiredOrUsed(_ context.Context, cutoff time.Time) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var n int64
	for id, tok := range r.s.tokens {
		if tok.UsedAt != nil || !tok.ExpiresAt.After(cutoff) {
			delete(r.s.tokens, id)
			n++
		}
	}
	return n, nil
}

type memSessionRepo struct{ s *memStore }

func (r *memSessionRepo) SelectByID(_ context.Context, id string) (*entity.Session, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	sess, ok := r.s.sessions[id]
	if !ok {
		return nil, iamerrors.ErrSessionNotFound
	}
	cp := *sess
	return &cp, nil
}

func (r *memSessionRepo) SelectByTokenDigest(_ context.Context, digest []byte) (*entity.Session, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, sess := range r.s.sessions {
		if bytes.Equal(sess.TokenDigest(), digest) {
			cp := *sess
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrSessionNotFound
}

func (r *memSessionRepo) SelectList(_ context.Context, filter domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	now := time.Now()
	var out []*entity.Session
	for _, sess := range r.s.sessions {
		if filter.UserID != "" && sess.UserID != filter.UserID {
			continue
		}
		if filter.ActiveOnly && !sess.IsActive(now) {
			continue
		}
		cp := *sess
		out = append(out, &cp)
	}
	slices.SortFunc(out, func(a, b *entity.Session) int { return strings.Compare(a.ID, b.ID) })
	return paginateSlice(out, filter.Page), nil
}

func (r *memSessionRepo) Insert(_ context.Context, sess *entity.Session) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *sess
	r.s.sessions[sess.ID] = &cp
	return nil
}

func (r *memSessionRepo) UpdateActivity(_ context.Context, id string, lastActiveAt, idleExpiresAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	sess, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	sess.LastActiveAt = lastActiveAt
	sess.IdleExpiresAt = idleExpiresAt
	return nil
}

func (r *memSessionRepo) UpdateMFAVerified(_ context.Context, id string, verifiedAt time.Time, rotatedCSRFDigest []byte) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	sess, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	if err := sess.MarkMFAVerified(verifiedAt); err != nil {
		return err
	}
	return sess.RotateCSRF(rotatedCSRFDigest)
}

func (r *memSessionRepo) UpdateEnded(_ context.Context, id string, reason entity.SessionEndReason, endedAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	sess, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	return sess.End(reason, endedAt)
}

func (r *memSessionRepo) UpdateEndedByUserID(_ context.Context, userID string, reason entity.SessionEndReason, endedAt time.Time) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var n int64
	for _, sess := range r.s.sessions {
		if sess.UserID == userID && sess.EndedAt == nil {
			if err := sess.End(reason, endedAt); err == nil {
				n++
			}
		}
	}
	return n, nil
}

func (r *memSessionRepo) DeleteEndedBefore(_ context.Context, cutoff time.Time) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var n int64
	for id, sess := range r.s.sessions {
		if sess.EndedAt != nil && !sess.EndedAt.After(cutoff) {
			delete(r.s.sessions, id)
			n++
		}
	}
	return n, nil
}
