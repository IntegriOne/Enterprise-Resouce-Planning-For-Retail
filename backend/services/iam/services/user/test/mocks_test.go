package user_test

import (
	"context"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *mutableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *mutableClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

type noopTx struct{}

func (noopTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockStore struct {
	mu          sync.Mutex
	users       map[string]*entity.User
	identities  map[string]*entity.UserIdentity
	credentials map[string]*entity.UserCredential
	totp        map[string]*entity.UserTOTPFactor
	recovery    map[string]*entity.UserRecoveryCode
	tokens      map[string]*entity.UserToken
	sessions    map[string]*entity.Session
	roles       map[string][]entity.Role
	roleDefs    map[string]*entity.Role
	userRoles   map[string][]*entity.UserRole
}

func newMockStore() *mockStore {
	return &mockStore{
		users:       make(map[string]*entity.User),
		identities:  make(map[string]*entity.UserIdentity),
		credentials: make(map[string]*entity.UserCredential),
		totp:        make(map[string]*entity.UserTOTPFactor),
		recovery:    make(map[string]*entity.UserRecoveryCode),
		tokens:      make(map[string]*entity.UserToken),
		sessions:    make(map[string]*entity.Session),
		roles:       make(map[string][]entity.Role),
		roleDefs:    make(map[string]*entity.Role),
		userRoles:   make(map[string][]*entity.UserRole),
	}
}

type mockEmailSender struct {
	mu   sync.Mutex
	sent []domainservice.EmailMessage
}

func (m *mockEmailSender) ProviderName() string { return "mock" }

func (m *mockEmailSender) Send(_ context.Context, msg domainservice.EmailMessage) (*domainservice.EmailSendResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, msg)
	return &domainservice.EmailSendResult{Provider: "mock", MessageID: "mock-msg-id"}, nil
}

func (m *mockEmailSender) Messages() []domainservice.EmailMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domainservice.EmailMessage, len(m.sent))
	copy(out, m.sent)
	return out
}

type testHarness struct {
	store       *mockStore
	clock       *mutableClock
	idGen       *crypto.UUIDGenerator
	tokGen      *crypto.OpaqueTokenGenerator
	totpProv    *totp.RFC6238Provider
	sessSvc     *session.Service
	userSvc     *user.Service
	emailSender *mockEmailSender
}

func newTestHarness(t interface{ Fatalf(string, ...any) }) *testHarness {
	clock := &mutableClock{now: time.Now()}
	store := newMockStore()
	idGen := crypto.NewUUIDGenerator()
	tokGen := crypto.NewOpaqueTokenGenerator()
	totpProv := totp.NewDefaultProvider()

	hasher, err := crypto.NewArgon2idHasher(crypto.Argon2Params{
		MemoryKiB:   8 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		t.Fatalf("NewArgon2idHasher() error = %v", err)
	}
	cipher, err := crypto.NewEphemeralAES256Cipher()
	if err != nil {
		t.Fatalf("NewEphemeralAES256Cipher() error = %v", err)
	}

	sessRepo := &mockSessionRepo{s: store}
	userRepo := &mockUserRepo{s: store}
	sessSvc, err := session.NewService(session.Dependencies{
		Sessions:        sessRepo,
		Users:           userRepo,
		Clock:           clock,
		IDGen:           idGen,
		Tokens:          tokGen,
		SessionPolicy:   domainservice.NewDefaultSessionPolicy(),
		AccessEvaluator: domainservice.NewAccessEvaluator(domainservice.NewDefaultSessionPolicy()),
		Redirect:        domainservice.NewDefaultRedirectResolver(),
	})
	if err != nil {
		t.Fatalf("session.NewService() error = %v", err)
	}

	userSvc, err := user.NewService(user.Dependencies{
		Transactor:     noopTx{},
		Users:          userRepo,
		Identities:     &mockIdentityRepo{s: store},
		Credentials:    &mockCredentialRepo{s: store},
		TOTPFactors:    &mockTOTPRepo{s: store},
		RecoveryCodes:  &mockRecoveryRepo{s: store},
		Tokens:         &mockTokenRepo{s: store},
		Roles:          &mockRoleRepo{s: store},
		UserRoles:      &mockUserRoleRepo{s: store},
		Sessions:       sessSvc,
		Clock:          clock,
		IDGen:          idGen,
		PasswordHasher: hasher,
		PasswordPolicy: domainservice.NewDefaultPasswordPolicy(),
		TokenGen:       tokGen,
		SecretCipher:   cipher,
		TOTPProvider:   totpProv,
		Redirect:       domainservice.NewDefaultRedirectResolver(),
		Config:         user.DefaultConfig(),
	})
	if err != nil {
		t.Fatalf("user.NewService() error = %v", err)
	}

	return &testHarness{
		store:    store,
		clock:    clock,
		idGen:    idGen,
		tokGen:   tokGen,
		totpProv: totpProv,
		sessSvc:  sessSvc,
		userSvc:  userSvc,
	}
}

type mockUserRepo struct{ s *mockStore }

func (r *mockUserRepo) SelectByID(_ context.Context, id string) (*entity.User, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[id]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}
func (r *mockUserRepo) SelectList(_ context.Context, f domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	norm := f.Normalize()
	var items []*entity.User
	for _, u := range r.s.users {
		if !norm.IncludeDeleted && u.IsDeleted() {
			continue
		}
		cp := *u
		items = append(items, &cp)
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}
func (r *mockUserRepo) SelectAccessSnapshotByUserID(_ context.Context, userID string, now time.Time) (*entity.AccessSnapshot, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[userID]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	mfa := false
	if f, ok := r.s.totp[userID]; ok && f.IsEnabled {
		mfa = true
	}
	return &entity.AccessSnapshot{
		UserID:     u.ID,
		UserKind:   u.Kind,
		UserStatus: u.Status,
		MFAEnabled: mfa,
		Roles:      r.s.roles[userID],
		CheckedAt:  now,
	}, nil
}
func (r *mockUserRepo) Insert(_ context.Context, u *entity.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *u
	r.s.users[u.ID] = &cp
	return nil
}
func (r *mockUserRepo) InsertWithIdentities(_ context.Context, u *entity.User, idents []*entity.UserIdentity) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, ident := range idents {
		for _, existing := range r.s.identities {
			if existing.Kind == ident.Kind && existing.Value == ident.Value {
				return iamerrors.ErrIdentityAlreadyExists
			}
		}
	}
	cp := *u
	r.s.users[u.ID] = &cp
	for _, ident := range idents {
		icp := *ident
		r.s.identities[ident.ID] = &icp
	}
	return nil
}
func (r *mockUserRepo) Update(_ context.Context, u *entity.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.users[u.ID]; !ok {
		return iamerrors.ErrUserNotFound
	}
	cp := *u
	r.s.users[u.ID] = &cp
	return nil
}
func (r *mockUserRepo) UpdateLastLogin(_ context.Context, userID string, loginAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[userID]
	if !ok {
		return iamerrors.ErrUserNotFound
	}
	u.RecordLogin(loginAt)
	return nil
}
func (r *mockUserRepo) SoftDelete(_ context.Context, id string, deletedAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[id]
	if !ok {
		return iamerrors.ErrUserNotFound
	}
	return u.SoftDelete(deletedAt)
}

type mockIdentityRepo struct{ s *mockStore }

func (r *mockIdentityRepo) SelectByID(_ context.Context, id string) (*entity.UserIdentity, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	ident, ok := r.s.identities[id]
	if !ok {
		return nil, iamerrors.ErrIdentityNotFound
	}
	cp := *ident
	return &cp, nil
}
func (r *mockIdentityRepo) SelectByKindAndValue(_ context.Context, kind entity.IdentityKind, value string) (*entity.UserIdentity, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, ident := range r.s.identities {
		if ident.Kind == kind && ident.Value == value {
			cp := *ident
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrIdentityNotFound
}
func (r *mockIdentityRepo) FindUserIdentityByEmailKind(ctx context.Context, email string) (*entity.UserIdentity, error) {
	norm, err := entity.ValidateEmail(email)
	if err != nil {
		return nil, err
	}
	return r.SelectByKindAndValue(ctx, entity.IdentityKindEmail, norm)
}
func (r *mockIdentityRepo) FindUserIdentityByUsernameKind(ctx context.Context, username string) (*entity.UserIdentity, error) {
	norm, err := entity.ValidateUsername(username)
	if err != nil {
		return nil, err
	}
	return r.SelectByKindAndValue(ctx, entity.IdentityKindUsername, norm)
}
func (r *mockIdentityRepo) SelectByEmailAndUsername(_ context.Context, email, username string) ([]*entity.UserIdentity, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []*entity.UserIdentity
	for _, ident := range r.s.identities {
		if email != "" && ident.Kind == entity.IdentityKindEmail && ident.Value == strings.ToLower(strings.TrimSpace(email)) {
			cp := *ident
			out = append(out, &cp)
		}
		if username != "" && ident.Kind == entity.IdentityKindUsername && ident.Value == strings.ToLower(strings.TrimSpace(username)) {
			cp := *ident
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (r *mockIdentityRepo) SelectByUserID(_ context.Context, userID string) ([]*entity.UserIdentity, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []*entity.UserIdentity
	for _, ident := range r.s.identities {
		if ident.UserID == userID {
			cp := *ident
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (r *mockIdentityRepo) CheckExistsByKindAndValue(_ context.Context, kind entity.IdentityKind, value string) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, ident := range r.s.identities {
		if ident.Kind == kind && ident.Value == value {
			return true, nil
		}
	}
	return false, nil
}
func (r *mockIdentityRepo) Insert(_ context.Context, ident *entity.UserIdentity) error {
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
func (r *mockIdentityRepo) Update(_ context.Context, ident *entity.UserIdentity) error {
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
func (r *mockIdentityRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.identities[id]; !ok {
		return iamerrors.ErrIdentityNotFound
	}
	delete(r.s.identities, id)
	return nil
}

type mockCredentialRepo struct{ s *mockStore }

func cloneCred(c *entity.UserCredential) *entity.UserCredential {
	cp, _ := entity.RehydrateUserCredential(
		c.UserID, c.PasswordHash(), c.MustChangePassword,
		c.PasswordChangedAt, c.FailedLoginAttempts, c.LockedUntil, c.CreatedAt, c.UpdatedAt,
	)
	return cp
}
func (r *mockCredentialRepo) SelectByUserID(_ context.Context, userID string) (*entity.UserCredential, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.credentials[userID]
	if !ok {
		return nil, iamerrors.ErrCredentialNotFound
	}
	return cloneCred(c), nil
}
func (r *mockCredentialRepo) Insert(_ context.Context, c *entity.UserCredential) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.credentials[c.UserID] = cloneCred(c)
	return nil
}
func (r *mockCredentialRepo) Update(_ context.Context, c *entity.UserCredential) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.credentials[c.UserID]; !ok {
		return iamerrors.ErrCredentialNotFound
	}
	r.s.credentials[c.UserID] = cloneCred(c)
	return nil
}
func (r *mockCredentialRepo) RecordFailedAttemptAtomic(_ context.Context, userID string, maxAttempts int, lockDuration time.Duration, now time.Time) (int, *time.Time, error) {
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
func (r *mockCredentialRepo) ResetFailedAttemptsAtomic(_ context.Context, userID string, now time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	c, ok := r.s.credentials[userID]
	if !ok {
		return nil
	}
	c.ResetFailedAttempts(now)
	return nil
}

type mockTOTPRepo struct{ s *mockStore }

func cloneTOTP(f *entity.UserTOTPFactor) *entity.UserTOTPFactor {
	cp, _ := entity.RehydrateUserTOTPFactor(
		f.ID, f.UserID, f.EncryptedSecret(), f.KeyVersion,
		f.IsEnabled, f.ConfirmedAt, f.LastUsedStep, f.CreatedAt, f.UpdatedAt,
	)
	return cp
}
func (r *mockTOTPRepo) SelectByUserID(_ context.Context, userID string) (*entity.UserTOTPFactor, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	f, ok := r.s.totp[userID]
	if !ok {
		return nil, iamerrors.ErrTOTPNotFound
	}
	return cloneTOTP(f), nil
}
func (r *mockTOTPRepo) Insert(_ context.Context, f *entity.UserTOTPFactor) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.totp[f.UserID] = cloneTOTP(f)
	return nil
}
func (r *mockTOTPRepo) Update(_ context.Context, f *entity.UserTOTPFactor) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.totp[f.UserID] = cloneTOTP(f)
	return nil
}
func (r *mockTOTPRepo) UpdateLastUsedStepIfGreater(_ context.Context, factorID string, step int64, updatedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, f := range r.s.totp {
		if f.ID == factorID {
			if f.LastUsedStep != nil && step <= *f.LastUsedStep {
				return false, nil
			}
			f.LastUsedStep = &step
			f.UpdatedAt = updatedAt
			return true, nil
		}
	}
	return false, iamerrors.ErrTOTPNotFound
}
func (r *mockTOTPRepo) DeleteByUserID(_ context.Context, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.totp[userID]; !ok {
		return iamerrors.ErrTOTPNotFound
	}
	delete(r.s.totp, userID)
	return nil
}

type mockRecoveryRepo struct{ s *mockStore }

func (r *mockRecoveryRepo) SelectUnusedByUserID(_ context.Context, userID string) ([]*entity.UserRecoveryCode, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []*entity.UserRecoveryCode
	for _, rc := range r.s.recovery {
		if rc.UserID == userID && !rc.IsUsed() {
			cp, _ := entity.RehydrateUserRecoveryCode(rc.ID, rc.UserID, rc.CodeHash(), rc.UsedAt, rc.CreatedAt)
			out = append(out, cp)
		}
	}
	return out, nil
}
func (r *mockRecoveryRepo) CountUnusedByUserID(_ context.Context, userID string) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var c int64
	for _, rc := range r.s.recovery {
		if rc.UserID == userID && !rc.IsUsed() {
			c++
		}
	}
	return c, nil
}
func (r *mockRecoveryRepo) InsertBatch(_ context.Context, codes []*entity.UserRecoveryCode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, rc := range codes {
		cp, _ := entity.RehydrateUserRecoveryCode(rc.ID, rc.UserID, rc.CodeHash(), rc.UsedAt, rc.CreatedAt)
		r.s.recovery[rc.ID] = cp
	}
	return nil
}
func (r *mockRecoveryRepo) UpdateUsedAtIfUnused(_ context.Context, id string, usedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	rc, ok := r.s.recovery[id]
	if !ok || rc.IsUsed() {
		return false, nil
	}
	_ = rc.Consume(usedAt)
	return true, nil
}
func (r *mockRecoveryRepo) DeleteByUserID(_ context.Context, userID string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for id, rc := range r.s.recovery {
		if rc.UserID == userID {
			delete(r.s.recovery, id)
		}
	}
	return nil
}

type mockTokenRepo struct{ s *mockStore }

func cloneToken(tok *entity.UserToken) *entity.UserToken {
	var identCopy *string
	if tok.IdentityID != nil {
		v := *tok.IdentityID
		identCopy = &v
	}
	cp, _ := entity.RehydrateUserToken(tok.ID, tok.UserID, identCopy, tok.Purpose, tok.Status, tok.TokenDigest(), tok.ExpiresAt, tok.UsedAt, tok.CreatedAt)
	return cp
}

func (r *mockTokenRepo) SelectByDigestAndPurpose(_ context.Context, digest []byte, purpose entity.TokenPurpose) (*entity.UserToken, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, tok := range r.s.tokens {
		if tok.Purpose == purpose && string(tok.TokenDigest()) == string(digest) {
			return cloneToken(tok), nil
		}
	}
	return nil, iamerrors.ErrTokenNotFound
}
func (r *mockTokenRepo) Insert(_ context.Context, tok *entity.UserToken) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if tok.Status == entity.TokenStatusActive {
		for _, existing := range r.s.tokens {
			if existing.UserID == tok.UserID && existing.Purpose == tok.Purpose && existing.Status == entity.TokenStatusActive {
				return iamerrors.NewValidation("token", iamerrors.Code("iam.token.duplicate_active"), "only one active token per purpose allowed")
			}
		}
	}
	r.s.tokens[tok.ID] = cloneToken(tok)
	return nil
}
func (r *mockTokenRepo) UpdateActiveStatusByUserAndPurpose(_ context.Context, userID string, purpose entity.TokenPurpose, nextStatus entity.TokenStatus) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var count int64
	for _, tok := range r.s.tokens {
		if tok.UserID == userID && tok.Purpose == purpose && tok.Status == entity.TokenStatusActive {
			tok.Status = nextStatus
			count++
		}
	}
	return count, nil
}
func (r *mockTokenRepo) UpdateUsedAtIfUnused(_ context.Context, id string, usedAt time.Time) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	tok, ok := r.s.tokens[id]
	if !ok || tok.Status != entity.TokenStatusActive || tok.UsedAt != nil {
		return false, nil
	}
	if err := tok.Consume(tok.Purpose, usedAt); err != nil {
		return false, err
	}
	return true, nil
}
func (r *mockTokenRepo) DeleteExpiredOrUsed(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

type mockRoleRepo struct{ s *mockStore }

func (r *mockRoleRepo) SelectByID(_ context.Context, id string) (*entity.Role, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if rl, ok := r.s.roleDefs[id]; ok {
		cp := *rl
		return &cp, nil
	}
	return nil, iamerrors.ErrRoleNotFound
}
func (r *mockRoleRepo) SelectByCode(_ context.Context, code string) (*entity.Role, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, rl := range r.s.roleDefs {
		if strings.EqualFold(rl.Code, code) {
			cp := *rl
			return &cp, nil
		}
	}
	return nil, iamerrors.ErrRoleNotFound
}
func (r *mockRoleRepo) SelectList(_ context.Context, _ domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	return domainrepo.PageResult[*entity.Role]{}, nil
}
func (r *mockRoleRepo) SelectEffectiveByUserID(_ context.Context, userID string, now time.Time) ([]entity.Role, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := append([]entity.Role(nil), r.s.roles[userID]...)
	seen := make(map[string]struct{}, len(out))
	for _, existing := range out {
		seen[existing.ID] = struct{}{}
	}
	for _, ur := range r.s.userRoles[userID] {
		if ur.ExpiresAt != nil && !ur.ExpiresAt.After(now) {
			continue
		}
		if rl, ok := r.s.roleDefs[ur.RoleID]; ok && rl.IsActive {
			if _, dup := seen[rl.ID]; !dup {
				seen[rl.ID] = struct{}{}
				out = append(out, *rl)
			}
		}
	}
	return out, nil
}
func (r *mockRoleRepo) Insert(_ context.Context, role *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *role
	r.s.roleDefs[role.ID] = &cp
	return nil
}
func (r *mockRoleRepo) Update(_ context.Context, role *entity.Role) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *role
	r.s.roleDefs[role.ID] = &cp
	return nil
}
func (r *mockRoleRepo) Delete(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	delete(r.s.roleDefs, id)
	return nil
}

type mockUserRoleRepo struct{ s *mockStore }

func (r *mockUserRoleRepo) SelectByUserID(_ context.Context, userID string, includeExpired bool, now time.Time) ([]*entity.UserRole, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []*entity.UserRole
	for _, ur := range r.s.userRoles[userID] {
		if !includeExpired && ur.ExpiresAt != nil && !ur.ExpiresAt.After(now) {
			continue
		}
		cp := *ur
		out = append(out, &cp)
	}
	return out, nil
}
func (r *mockUserRoleRepo) CheckExists(_ context.Context, userID, roleID string, scope entity.Scope) (bool, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, ur := range r.s.userRoles[userID] {
		if ur.RoleID == roleID && ur.Scope.Type == scope.Type {
			if ur.Scope.ID == nil && scope.ID == nil {
				return true, nil
			}
			if ur.Scope.ID != nil && scope.ID != nil && *ur.Scope.ID == *scope.ID {
				return true, nil
			}
		}
	}
	return false, nil
}
func (r *mockUserRoleRepo) Insert(_ context.Context, a *entity.UserRole) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	cp := *a
	r.s.userRoles[a.UserID] = append(r.s.userRoles[a.UserID], &cp)
	return nil
}
func (r *mockUserRoleRepo) DeleteByID(_ context.Context, id string) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for uid, list := range r.s.userRoles {
		for i, ur := range list {
			if ur.ID == id {
				r.s.userRoles[uid] = append(list[:i], list[i+1:]...)
				return nil
			}
		}
	}
	return iamerrors.ErrUserRoleNotFound
}

type mockSessionRepo struct{ s *mockStore }

func cloneSess(s *entity.Session) *entity.Session {
	cp, _ := entity.RehydrateSession(
		s.ID, s.UserID, s.TokenDigest(), s.CSRFTokenDigest(),
		s.MFAVerified, s.MFAVerifiedAt, s.IPAddress, s.UserAgent,
		s.AuthenticatedAt, s.LastActiveAt, s.IdleExpiresAt, s.AbsoluteExpiresAt,
		s.EndedAt, s.EndReason, s.CreatedAt,
	)
	return cp
}
func (r *mockSessionRepo) SelectByID(_ context.Context, id string) (*entity.Session, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	s, ok := r.s.sessions[id]
	if !ok {
		return nil, iamerrors.ErrSessionNotFound
	}
	return cloneSess(s), nil
}
func (r *mockSessionRepo) SelectByTokenDigest(_ context.Context, digest []byte) (*entity.Session, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, s := range r.s.sessions {
		if string(s.TokenDigest()) == string(digest) {
			return cloneSess(s), nil
		}
	}
	return nil, iamerrors.ErrSessionNotFound
}
func (r *mockSessionRepo) SelectList(_ context.Context, f domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	norm := f.Normalize()
	var items []*entity.Session
	for _, s := range r.s.sessions {
		if s.UserID == norm.UserID && (!norm.ActiveOnly || s.EndedAt == nil) {
			items = append(items, cloneSess(s))
		}
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}
func (r *mockSessionRepo) Insert(_ context.Context, s *entity.Session) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.sessions[s.ID] = cloneSess(s)
	return nil
}
func (r *mockSessionRepo) UpdateActivity(_ context.Context, id string, lastActiveAt, idleExpiresAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	s, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	s.LastActiveAt = lastActiveAt
	s.IdleExpiresAt = idleExpiresAt
	return nil
}
func (r *mockSessionRepo) UpdateMFAVerified(_ context.Context, id string, verifiedAt time.Time, rotatedCSRF []byte) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	s, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	_ = s.MarkMFAVerified(verifiedAt)
	return s.RotateCSRF(rotatedCSRF)
}
func (r *mockSessionRepo) UpdateEnded(_ context.Context, id string, reason entity.SessionEndReason, endedAt time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	s, ok := r.s.sessions[id]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	return s.End(reason, endedAt)
}
func (r *mockSessionRepo) UpdateEndedByUserID(_ context.Context, userID string, reason entity.SessionEndReason, endedAt time.Time) (int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var n int64
	for _, s := range r.s.sessions {
		if s.UserID == userID && s.EndedAt == nil {
			_ = s.End(reason, endedAt)
			n++
		}
	}
	return n, nil
}
func (r *mockSessionRepo) DeleteEndedBefore(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

var _ = netip.Addr{}
