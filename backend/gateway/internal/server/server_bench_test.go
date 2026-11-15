package server

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

func BenchmarkVerifyCSRFTokenDigest(b *testing.B) {
	plain, digest, _ := browserauth.GenerateOpaquePair()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !browserauth.VerifyCSRFTokenDigest(plain, digest) {
			b.Fatal("csrf mismatch")
		}
	}
}

func BenchmarkOriginValidator(b *testing.B) {
	ov, _ := browserauth.NewOriginValidator([]string{"https://erp.localhost:8443"})
	hdr := http.Header{
		"Origin":         []string{"https://erp.localhost:8443"},
		"Sec-Fetch-Site": []string{"same-origin"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ov.ValidateRequest(hdr, true); err != nil {
			b.Fatal(err)
		}
	}
}

func newSilentBenchLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func BenchmarkAuthorizerSnapshotHit(b *testing.B) {
	srv, _, rawTok, _ := buildTestGatewayServerWithLogger(b, newSilentBenchLogger(), false)
	_ = srv

	store := newTestStore()
	tokHash := authsession.HashOpaqueToken(rawTok)
	now := browserauth.NewResponseEffects()
	_ = now

	loader := middleware.NewSnapshotLoader(noopRefresher{}, 0, authz.DefaultSnapshotMaxAge)
	mgr, _ := middleware.NewAuthManager(store, loader, middleware.SessionConfig{CookieSecure: true})
	authzSvc := middleware.NewAuthorizer(mgr, nil, 0)

	// Siapkan satu sesi hangat di store lokal untuk mengukur overhead in-process RBAC tanpa gRPC.
	ctx := middleware.WithResolvedSession(context.Background(), &middleware.ResolvedSession{
		RawToken:  rawTok,
		TokenHash: tokHash,
		Token:     &authsession.SessionTokenRecord{SessionID: "s1", UserID: "u1", IssuedAtUnix: 1 << 32},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "u1",
			Status:      authz.StatusActive,
			Permissions: []authz.PermissionGrant{
				{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := authzSvc.RequirePermission(ctx, "iam.users.read", authz.ScopeGrant{Type: authz.ScopeGlobal}, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPGraphQLHealth(b *testing.B) {
	srv, _, _, _ := buildTestGatewayServerWithLogger(b, newSilentBenchLogger(), false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/graphql?query=%7Bhealth%7D", nil)
		resp, err := srv.App().Test(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			b.Fatalf("health bench failed: err=%v status=%d", err, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
}

func BenchmarkHTTPGraphQLCookieQuery(b *testing.B) {
	srv, _, rawTok, _ := buildTestGatewayServerWithLogger(b, newSilentBenchLogger(), false)
	payload := []byte(`{"query":"{ users(first: 10) { edges { cursor node { id _id display_name } } pageInfo { endCursor hasNextPage } total_count } }"}`)
	cookie := &http.Cookie{Name: config.SessionCookieName, Value: rawTok}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/graphql", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		resp, err := srv.App().Test(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			b.Fatalf("cookie query bench failed: err=%v status=%d", err, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
}
