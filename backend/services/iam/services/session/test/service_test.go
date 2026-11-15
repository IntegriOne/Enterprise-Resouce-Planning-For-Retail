package session_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
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

type mockSessionRepo struct {
	mu       sync.Mutex
	sessions map[string]*entity.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[string]*entity.Session)}
}

func cloneSession(s *entity.Session) *entity.Session {
	if s == nil {
		return nil
	}
	rehydrated, err := entity.RehydrateSession(
		s.ID, s.UserID, s.TokenDigest(), s.CSRFTokenDigest(),
		s.MFAVerified, s.MFAVerifiedAt, s.IPAddress, s.UserAgent,
		s.AuthenticatedAt, s.LastActiveAt, s.IdleExpiresAt, s.AbsoluteExpiresAt,
		s.EndedAt, s.EndReason, s.CreatedAt,
	)
	if err != nil {
		cp := *s
		return &cp
	}
	rehydrated.SetSecurityMetadata(s.Generation, s.CSRFKeyVersion)
	return rehydrated
}

func (m *mockSessionRepo) SelectByID(_ context.Context, id string) (*entity.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return nil, iamerrors.ErrSessionNotFound
	}
	return cloneSession(s), nil
}

func (m *mockSessionRepo) SelectByTokenDigest(_ context.Context, tokenDigest []byte) (*entity.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if bytes.Equal(s.TokenDigest(), tokenDigest) {
			return cloneSession(s), nil
		}
	}
	return nil, iamerrors.ErrSessionNotFound
}

func (m *mockSessionRepo) SelectList(_ context.Context, filter domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	norm := filter.Normalize()
	var items []*entity.Session
	for _, s := range m.sessions {
		if s.UserID != norm.UserID {
			continue
		}
		if norm.ActiveOnly && s.EndedAt != nil {
			continue
		}
		items = append(items, cloneSession(s))
	}
	return domainrepo.NewPageResult(items, int64(len(items)), norm.Page), nil
}

func (m *mockSessionRepo) Insert(_ context.Context, s *entity.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = cloneSession(s)
	return nil
}

func (m *mockSessionRepo) UpdateActivity(_ context.Context, sessionID string, lastActiveAt, idleExpiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	s.LastActiveAt = lastActiveAt
	s.IdleExpiresAt = idleExpiresAt
	return nil
}

func (m *mockSessionRepo) UpdateMFAVerified(_ context.Context, sessionID string, verifiedAt time.Time, rotatedCSRFDigest []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	if err := s.MarkMFAVerified(verifiedAt); err != nil {
		return err
	}
	s.BumpGeneration()
	return s.RotateCSRF(rotatedCSRFDigest)
}

func (m *mockSessionRepo) UpdateEnded(_ context.Context, sessionID string, reason entity.SessionEndReason, endedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return iamerrors.ErrSessionNotFound
	}
	if s.EndedAt != nil {
		return iamerrors.ErrSessionAlreadyEnded
	}
	return s.End(reason, endedAt)
}

func (m *mockSessionRepo) UpdateEndedByUserID(_ context.Context, userID string, reason entity.SessionEndReason, endedAt time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, s := range m.sessions {
		if s.UserID == userID && s.EndedAt == nil {
			_ = s.End(reason, endedAt)
			count++
		}
	}
	return count, nil
}

func (m *mockSessionRepo) DeleteEndedBefore(_ context.Context, cutoff time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for id, s := range m.sessions {
		if s.EndedAt != nil && !s.EndedAt.After(cutoff) {
			delete(m.sessions, id)
			count++
		}
	}
	return count, nil
}

type mockUserRepo struct {
	mu           sync.Mutex
	users        map[string]*entity.User
	snapshots    map[string]*entity.AccessSnapshot
	staleChecked bool
	clock        *mutableClock
}

func (u *mockUserRepo) SelectByID(_ context.Context, id string) (*entity.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	usr, ok := u.users[id]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cp := *usr
	return &cp, nil
}
func (u *mockUserRepo) SelectList(_ context.Context, _ domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	return domainrepo.PageResult[*entity.User]{}, nil
}
func (u *mockUserRepo) SelectAccessSnapshotByUserID(_ context.Context, userID string, now time.Time) (*entity.AccessSnapshot, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	snap, ok := u.snapshots[userID]
	if !ok {
		return nil, iamerrors.ErrUserNotFound
	}
	cloned := snap.Clone()
	if u.staleChecked {
		cloned.CheckedAt = now.Add(-45 * time.Second)
	} else {
		cloned.CheckedAt = now
	}
	return &cloned, nil
}
func (u *mockUserRepo) Insert(_ context.Context, _ *entity.User) error { return nil }
func (u *mockUserRepo) InsertWithIdentities(_ context.Context, _ *entity.User, _ []*entity.UserIdentity) error {
	return nil
}
func (u *mockUserRepo) Update(_ context.Context, _ *entity.User) error                 { return nil }
func (u *mockUserRepo) UpdateLastLogin(_ context.Context, _ string, _ time.Time) error { return nil }
func (u *mockUserRepo) SoftDelete(_ context.Context, _ string, _ time.Time) error      { return nil }

func TestSessionService_IssueAuthenticateTimeoutsCSRFAndLogout(t *testing.T) {
	t.Parallel()

	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	tokGen := crypto.NewOpaqueTokenGenerator()
	sessRepo := newMockSessionRepo()
	userRepo := &mockUserRepo{
		users:     make(map[string]*entity.User),
		snapshots: make(map[string]*entity.AccessSnapshot),
		clock:     clock,
	}

	svc, err := session.NewService(session.Dependencies{
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
		t.Fatalf("NewService() error = %v", err)
	}

	userID, _ := idGen.NewUUID()
	u, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Alice", clock.Now())
	userRepo.users[userID] = u

	// Terbitkan sesi valid dengan IP valid.
	issued, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:      userID,
		MFAVerified: false,
		IPAddress:   "127.0.0.1",
		UserAgent:   "IntegrationTest/1.0",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if issued.SessionToken == "" || issued.CSRFToken == "" {
		t.Fatal("expected non-empty opaque session and CSRF tokens")
	}

	// Tolak IP address yang tidak valid.
	if _, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:    userID,
		IPAddress: "not-an-ip",
	}); !errors.Is(err, iamerrors.ErrInvalidIPAddress) {
		t.Fatalf("expected ErrInvalidIPAddress, got %v", err)
	}

	// Validasi CSRF yang salah vs benar.
	if err := svc.VerifyCSRFOnSession(issued.Session, "wrong-csrf"); !errors.Is(err, iamerrors.ErrCSRFMismatch) {
		t.Fatalf("expected ErrCSRFMismatch, got %v", err)
	}
	if err := svc.VerifyCSRFOnSession(issued.Session, issued.CSRFToken); err != nil {
		t.Fatalf("VerifyCSRFOnSession(valid) error = %v", err)
	}

	// Tandai MFA terverifikasi + rotasi token CSRF lama agar token lama tidak berlaku lagi.
	rotated, err := svc.MarkMFAVerified(context.Background(), issued.Session.ID)
	if err != nil {
		t.Fatalf("MarkMFAVerified() error = %v", err)
	}
	if rotated.CSRFToken == issued.CSRFToken {
		t.Fatal("expected CSRF token rotation upon MFA verification")
	}
	if _, err := svc.VerifyCSRF(context.Background(), issued.Session.ID, issued.CSRFToken); !errors.Is(err, iamerrors.ErrCSRFMismatch) {
		t.Fatalf("expected old CSRF token to fail after rotation, got %v", err)
	}
	if _, err := svc.VerifyCSRF(context.Background(), issued.Session.ID, rotated.CSRFToken); err != nil {
		t.Fatalf("VerifyCSRF(new) error = %v", err)
	}

	// Maju 20 menit -> sesi masih aktif dan idle timeout bergeser (+30m dari sekarang).
	clock.Advance(20 * time.Minute)
	touched, err := svc.Authenticate(context.Background(), issued.SessionToken)
	if err != nil {
		t.Fatalf("Authenticate(+20m) error = %v", err)
	}
	if !touched.MFAVerified {
		t.Fatal("expected touched session to retain MFAVerified")
	}

	// Maju 31 menit tanpa aktivitas -> melewati idle timeout 30 menit.
	clock.Advance(31 * time.Minute)
	if _, err := svc.Authenticate(context.Background(), issued.SessionToken); !errors.Is(err, iamerrors.ErrSessionIdleExpired) {
		t.Fatalf("expected ErrSessionIdleExpired after 31m idle, got %v", err)
	}

	// Terbitkan sesi baru dan uji absolute timeout 8 jam meski terus disentuh tiap 20 menit.
	issued2, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:      userID,
		MFAVerified: true,
	})
	if err != nil {
		t.Fatalf("Issue(2) error = %v", err)
	}
	for i := 0; i < 23; i++ {
		clock.Advance(20 * time.Minute)
		if _, err := svc.Authenticate(context.Background(), issued2.SessionToken); err != nil {
			t.Fatalf("expected session active before 8h at iteration %d, got %v", i, err)
		}
	}
	// Total 23*20m = 7h40m -> tambah 25m -> 8h05m melewati batas mutlak 8h.
	clock.Advance(25 * time.Minute)
	if _, err := svc.Authenticate(context.Background(), issued2.SessionToken); !errors.Is(err, iamerrors.ErrSessionAbsoluteExpired) {
		t.Fatalf("expected ErrSessionAbsoluteExpired past 8h, got %v", err)
	}

	// Terbitkan sesi 3 lalu logout eksplisit.
	issued3, err := svc.Issue(context.Background(), session.IssueInput{UserID: userID})
	if err != nil {
		t.Fatalf("Issue(3) error = %v", err)
	}
	if err := svc.Logout(context.Background(), issued3.SessionToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), issued3.SessionToken); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("expected ErrSessionRevoked after logout, got %v", err)
	}
}

func TestSessionService_AuthorizeFailClosedAndStaleSnapshot(t *testing.T) {
	t.Parallel()

	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	tokGen := crypto.NewOpaqueTokenGenerator()
	sessRepo := newMockSessionRepo()
	userRepo := &mockUserRepo{
		users:     make(map[string]*entity.User),
		snapshots: make(map[string]*entity.AccessSnapshot),
		clock:     clock,
	}

	svc, err := session.NewService(session.Dependencies{
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
		t.Fatalf("NewService() error = %v", err)
	}

	userID, _ := idGen.NewUUID()
	u, _ := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Security Officer", clock.Now())
	userRepo.users[userID] = u
	userRepo.snapshots[userID] = &entity.AccessSnapshot{
		UserID:     userID,
		UserKind:   entity.UserKindEmployee,
		UserStatus: entity.UserStatusActive,
		MFAEnabled: true,
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "iam.roles:write", IsSensitive: true, Scope: entity.GlobalScope()},
		},
	}

	issued, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:      userID,
		MFAVerified: true,
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Akses dengan CSRF valid dan MFA baru -> diizinkan.
	res, err := svc.Authorize(context.Background(), session.AuthorizeRequest{
		SessionToken:           issued.SessionToken,
		CSRFToken:              issued.CSRFToken,
		RequireCSRF:            true,
		RequiredPermission:     "iam.roles:write",
		RequiredScope:          entity.GlobalScope(),
		RequireSensitiveReauth: true,
	})
	if err != nil || res.User.ID != userID {
		t.Fatalf("Authorize(valid) = (%+v, %v)", res, err)
	}

	// CSRF salah pada mutasi -> ditolak.
	if _, err := svc.Authorize(context.Background(), session.AuthorizeRequest{
		SessionToken:       issued.SessionToken,
		CSRFToken:          "invalid",
		RequireCSRF:        true,
		RequiredPermission: "iam.roles:write",
		RequiredScope:      entity.GlobalScope(),
	}); !errors.Is(err, iamerrors.ErrCSRFMismatch) {
		t.Fatalf("expected ErrCSRFMismatch, got %v", err)
	}

	// Snapshot basi (>30 detik) -> ditolak fail-closed ErrSnapshotStale.
	userRepo.staleChecked = true
	if _, err := svc.Authorize(context.Background(), session.AuthorizeRequest{
		SessionToken:       issued.SessionToken,
		RequiredPermission: "iam.roles:write",
		RequiredScope:      entity.GlobalScope(),
	}); !errors.Is(err, iamerrors.ErrSnapshotStale) {
		t.Fatalf("expected ErrSnapshotStale, got %v", err)
	}
	userRepo.staleChecked = false

	// Lewat jendela sensitif 10 menit -> reauth wajib ditolak.
	clock.Advance(11 * time.Minute)
	if _, err := svc.Authorize(context.Background(), session.AuthorizeRequest{
		SessionToken:           issued.SessionToken,
		RequiredPermission:     "iam.roles:write",
		RequiredScope:          entity.GlobalScope(),
		RequireSensitiveReauth: true,
	}); !errors.Is(err, iamerrors.ErrSensitiveReauthRequired) {
		t.Fatalf("expected ErrSensitiveReauthRequired past 10m, got %v", err)
	}
}
