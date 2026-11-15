package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type fakeSessionStore struct {
	mu            sync.Mutex
	gateOpen      bool
	gateErr       error
	readErr       error
	touchErr      error
	preAuthErr    error
	sessions      map[string]*authsession.SessionValidationContext
	preAuths      map[string]authsession.PreAuthRecord
	rateCounts    map[string]int64
	touchCalled   atomic.Int64
	deletedPreKey string
}

func newFakeStore() *fakeSessionStore {
	return &fakeSessionStore{
		gateOpen:   true,
		sessions:   make(map[string]*authsession.SessionValidationContext),
		preAuths:   make(map[string]authsession.PreAuthRecord),
		rateCounts: make(map[string]int64),
	}
}

func (f *fakeSessionStore) GetGate(ctx context.Context) (*authsession.SecurityGateRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.gateErr != nil {
		return nil, f.gateErr
	}
	return &authsession.SecurityGateRecord{Open: f.gateOpen, RecoveryEpoch: 1}, nil
}

func (f *fakeSessionStore) ReadSessionContext(ctx context.Context, tokenHashHex string, now time.Time) (*authsession.SessionValidationContext, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readErr != nil {
		return nil, f.readErr
	}
	if !f.gateOpen {
		return nil, authsession.ErrSecurityGateClosed
	}
	sc, ok := f.sessions[tokenHashHex]
	if !ok {
		return nil, authsession.ErrSessionNotFound
	}
	return sc, nil
}

func (f *fakeSessionStore) TouchActivity(ctx context.Context, tokenHashHex string, tok authsession.SessionTokenRecord, idleTimeout time.Duration, now time.Time) error {
	f.touchCalled.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.touchErr
}

func (f *fakeSessionStore) SetPreAuth(ctx context.Context, rec authsession.PreAuthRecord, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.preAuthErr != nil {
		return f.preAuthErr
	}
	f.preAuths[rec.TokenHashHex] = rec
	return nil
}

func (f *fakeSessionStore) GetPreAuth(ctx context.Context, cookieHashHex string, now time.Time) (*authsession.PreAuthRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.preAuthErr != nil {
		return nil, f.preAuthErr
	}
	rec, ok := f.preAuths[cookieHashHex]
	if !ok {
		return nil, authsession.ErrSessionNotFound
	}
	return &rec, nil
}

func (f *fakeSessionStore) DeletePreAuth(ctx context.Context, cookieHashHex string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedPreKey = cookieHashHex
	delete(f.preAuths, cookieHashHex)
	return nil
}

func (f *fakeSessionStore) CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := scope + ":" + subject
	f.rateCounts[key]++
	if f.rateCounts[key] > limit {
		return f.rateCounts[key], authsession.ErrRateLimitExceeded
	}
	return f.rateCounts[key], nil
}

type fakeRefresher struct {
	calls atomic.Int32
	delay time.Duration
	snap  *authz.AccessSnapshot
	err   error
}

func (r *fakeRefresher) RefreshUserAccessSnapshot(ctx context.Context, userID string, caller authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	r.calls.Add(1)
	if r.delay > 0 {
		select {
		case <-time.After(r.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return r.snap, r.err
}

type fakeSensitiveChecker struct {
	calls atomic.Int32
	err   error
}

func (c *fakeSensitiveChecker) VerifySensitiveSession(ctx context.Context, rawToken, requiredPermission string, requiredScope authz.ScopeGrant, caller authcontext.CallerContext) error {
	c.calls.Add(1)
	return c.err
}

func buildActiveSnapshot(userID string, now time.Time, permRev, authEpoch int64, perms ...authz.PermissionGrant) *authz.AccessSnapshot {
	return &authz.AccessSnapshot{
		PrincipalID:    userID,
		UserKind:       "employee",
		Status:         authz.StatusActive,
		AuthEpoch:      authEpoch,
		PermRevision:   permRev,
		RecoveryEpoch:  1,
		SyncedAtUnix:   now.Unix(),
		FreshUntilUnix: now.Add(30 * time.Second).Unix(),
		Permissions:    perms,
	}
}

func TestSnapshotLoaderSingleflightAndWithoutCancel(t *testing.T) {
	now := time.Now()
	ref := &fakeRefresher{
		delay: 40 * time.Millisecond,
		snap: buildActiveSnapshot("u1", now.Add(20*time.Millisecond), 3, 2, authz.PermissionGrant{
			Code:  "iam.users.read",
			Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
		}),
	}
	loader := NewSnapshotLoader(ref, 2*time.Second, 30*time.Second)

	staleSess := &authsession.SessionValidationContext{
		Token:    &authsession.SessionTokenRecord{SessionID: "s1", UserID: "u1", AuthEpoch: 2},
		Security: &authsession.UserSecurityRecord{UserID: "u1", Status: authz.StatusActive, AuthEpoch: 2, PermRevision: 3},
		Snapshot: buildActiveSnapshot("u1", now.Add(-45*time.Second), 3, 2),
	}

	var wg sync.WaitGroup
	errs := make(chan error, 12)

	// Panggil pertama dengan context yang segera dibatalkan untuk memastikan waiter lain tetap berhasil lewat WithoutCancel.
	c1, cancel1 := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		cancel1()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = loader.EnsureFresh(c1, staleSess, authcontext.CallerContext{})
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snap, err := loader.EnsureFresh(context.Background(), staleSess, authcontext.CallerContext{})
			if err != nil {
				errs <- err
				return
			}
			if snap == nil || snap.PrincipalID != "u1" {
				errs <- errors.New("unexpected snapshot result")
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("unexpected concurrent EnsureFresh error: %v", err)
	}
	if got := ref.calls.Load(); got != 1 {
		t.Fatalf("expected 1 coalesced singleflight IAM call, got %d", got)
	}
}

func TestAuthManagerFailClosedAndCSRF(t *testing.T) {
	now := time.Now()
	store := newFakeStore()
	loader := NewSnapshotLoader(&fakeRefresher{}, 2*time.Second, 30*time.Second)
	mgr, err := NewAuthManager(store, loader, SessionConfig{
		CookieName:         browserauth.DefaultSessionCookieName,
		CookieSecure:       true,
		PreAuthTTL:         15 * time.Minute,
		SessionIdleTimeout: 30 * time.Minute,
		SessionMaxAge:      8 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}

	// 1. Pre-auth session gagal fail-closed ketika gate tertutup.
	store.gateOpen = false
	effClosed := browserauth.NewResponseEffects()
	if _, err := mgr.IssuePreAuthSession(browserauth.WithResponseEffects(context.Background(), effClosed)); err == nil {
		t.Fatal("expected closed security gate to reject IssuePreAuthSession")
	}
	store.gateOpen = true

	// 2. Active session + CSRF check.
	rawTok := "raw-session-token"
	tokHash := authsession.HashOpaqueToken(rawTok)
	csrfPlain, csrfHash, _ := browserauth.GenerateOpaquePair()
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:     "s1",
			UserID:        "u1",
			TokenHashHex:  tokHash,
			CSRFDigestHex: csrfHash,
			AuthEpoch:     1,
			IssuedAtUnix:  now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: "u1", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1},
		Snapshot: buildActiveSnapshot("u1", now, 1, 1, authz.PermissionGrant{
			Code:  "iam.users.write",
			Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
		}),
	}

	ctxOK := WithRequestMetadata(context.Background(), RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
		CSRFToken:  csrfPlain,
	})
	resolved, err := mgr.RequireMutationCSRF(ctxOK)
	if err != nil || resolved == nil || resolved.Token.UserID != "u1" {
		t.Fatalf("expected RequireMutationCSRF to pass, got err=%v", err)
	}

	ctxBadCSRF := WithRequestMetadata(context.Background(), RequestMetadata{
		RawCookie:  rawTok,
		CookieHash: tokHash,
		CSRFToken:  "wrong-csrf",
	})
	if _, err := mgr.RequireMutationCSRF(ctxBadCSRF); err == nil {
		t.Fatal("expected bad CSRF token to fail")
	}

	// 3. Simulasi Redis down -> UNAVAILABLE + jangan anggap anonim.
	store.readErr = authsession.ErrRedisUnavailable
	if _, err := mgr.AuthenticateOptional(ctxOK); err == nil {
		t.Fatal("expected AuthenticateOptional to return UNAVAILABLE when Redis down")
	}

	// 4. RequireAuthenticated dengan pre-auth cookie sah mengembalikan UNAUTHENTICATED tanpa menghapus cookie pre-auth.
	store.readErr = nil
	preTok := "preauth-raw"
	preHash := authsession.HashOpaqueToken(preTok)
	preCSRFPlain, preCSRFHash, _ := browserauth.GenerateOpaquePair()
	store.preAuths[preHash] = authsession.PreAuthRecord{
		TokenHashHex:  preHash,
		CSRFDigestHex: preCSRFHash,
		CreatedAtUnix: now.Unix(),
		ExpiresAtUnix: now.Add(15 * time.Minute).Unix(),
	}
	effPre := browserauth.NewResponseEffects()
	ctxPre := browserauth.WithResponseEffects(WithRequestMetadata(context.Background(), RequestMetadata{
		RawCookie:  preTok,
		CookieHash: preHash,
		CSRFToken:  preCSRFPlain,
	}), effPre)
	if _, err := mgr.RequireAuthenticated(ctxPre); err == nil {
		t.Fatal("expected RequireAuthenticated on pre-auth handle to return UNAUTHENTICATED")
	} else if gwErr, ok := AsGatewayError(err); !ok || gwErr.Code != CodeUnauthenticated {
		t.Fatalf("expected UNAUTHENTICATED, got %v", err)
	}
	recPre := httptest.NewRecorder()
	effPre.ApplyTo(recPre)
	if gotCookies := recPre.Result().Cookies(); len(gotCookies) != 0 {
		t.Fatalf("expected pre-auth cookie NOT to be cleared by RequireAuthenticated, got cookies=%+v", gotCookies)
	}
	if _, err := mgr.RequireMutationCSRF(ctxPre); err != nil {
		t.Fatalf("expected subsequent RequireMutationCSRF with preserved pre-auth cookie to succeed, got %v", err)
	}
}

func TestAuthorizerRBACAndSensitiveReauth(t *testing.T) {
	now := time.Now()
	store := newFakeStore()
	loader := NewSnapshotLoader(&fakeRefresher{}, 2*time.Second, 30*time.Second)
	mgr, _ := NewAuthManager(store, loader, SessionConfig{CookieSecure: true})

	rawTok := "tok-authz"
	tokHash := authsession.HashOpaqueToken(rawTok)
	store.sessions[tokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:        "s-admin",
			UserID:           "admin-1",
			TokenHashHex:     tokHash,
			AuthEpoch:        1,
			IssuedAtUnix:     now.Add(-25 * time.Minute).Unix(),
			MFAAuthenticated: true,
		},
		Security: &authsession.UserSecurityRecord{UserID: "admin-1", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:    "admin-1",
			UserKind:       "employee",
			Status:         authz.StatusActive,
			MFAEnabled:     true,
			AuthEpoch:      1,
			PermRevision:   1,
			RecoveryEpoch:  1,
			SyncedAtUnix:   now.Unix(),
			FreshUntilUnix: now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
				{Code: "iam.users.write", IsSensitive: true, Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	}

	ctx := WithRequestMetadata(context.Background(), RequestMetadata{RawCookie: rawTok, CookieHash: tokHash})

	// Read tanpa sensitive reauth -> sukses.
	checker := &fakeSensitiveChecker{err: errors.New("reauth expired")}
	authzSvc := NewAuthorizer(mgr, checker, 10*time.Minute)
	if _, err := authzSvc.RequirePermission(ctx, "iam.users.read", authz.ScopeGrant{Type: authz.ScopeGlobal}, false); err != nil {
		t.Fatalf("expected non-sensitive read to succeed, got %v", err)
	}

	// Write dengan auth lama -> jatuh ke IAM checker, ditolak FORBIDDEN.
	if _, err := authzSvc.RequirePermission(ctx, "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true); err == nil {
		t.Fatal("expected expired reauth window to reject sensitive write")
	} else if gwErr, ok := AsGatewayError(err); !ok || gwErr.Code != CodeForbidden {
		t.Fatalf("expected FORBIDDEN error, got %v", err)
	}
	if checker.calls.Load() != 1 {
		t.Fatalf("expected fallback check on IAM sensitive session verifier, got %d", checker.calls.Load())
	}

	// Jika IAM menyatakan MFAVerifiedAt masih segar (<=10m), harus lolos.
	checker.err = nil
	if _, err := authzSvc.RequirePermission(ctx, "iam.users.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true); err != nil {
		t.Fatalf("expected IAM authoritative reauth verification to pass, got %v", err)
	}

	// Izin tidak dimiliki -> langsung FORBIDDEN tanpa menyentuh IAM sensitive check.
	if _, err := authzSvc.RequirePermission(ctx, "iam.roles.write", authz.ScopeGrant{Type: authz.ScopeGlobal}, true); err == nil {
		t.Fatal("expected missing permission to fail with FORBIDDEN")
	}
}
