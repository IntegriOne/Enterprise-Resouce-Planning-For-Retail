package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/integrione/erp-retail/backend/gateway/graph/generated"
	"github.com/integrione/erp-retail/backend/gateway/graph/resolvers"
	hrmadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/human_resource"
	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

type testStore struct {
	mu       sync.Mutex
	sessions map[string]*authsession.SessionValidationContext
	preAuths map[string]authsession.PreAuthRecord
	rates    map[string]int64
}

func newTestStore() *testStore {
	return &testStore{
		sessions: make(map[string]*authsession.SessionValidationContext),
		preAuths: make(map[string]authsession.PreAuthRecord),
		rates:    make(map[string]int64),
	}
}

func (s *testStore) GetGate(ctx context.Context) (*authsession.SecurityGateRecord, error) {
	return &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1}, nil
}

func (s *testStore) ReadSessionContext(ctx context.Context, tokenHashHex string, now time.Time) (*authsession.SessionValidationContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sc, ok := s.sessions[tokenHashHex]; ok {
		return sc, nil
	}
	return nil, authsession.ErrSessionNotFound
}

func (s *testStore) TouchActivity(ctx context.Context, tokenHashHex string, tok authsession.SessionTokenRecord, idleTimeout time.Duration, now time.Time) error {
	return nil
}

func (s *testStore) SetPreAuth(ctx context.Context, rec authsession.PreAuthRecord, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preAuths[rec.TokenHashHex] = rec
	return nil
}

func (s *testStore) GetPreAuth(ctx context.Context, cookieHashHex string, now time.Time) (*authsession.PreAuthRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec, ok := s.preAuths[cookieHashHex]; ok {
		return &rec, nil
	}
	return nil, authsession.ErrSessionNotFound
}

func (s *testStore) DeletePreAuth(ctx context.Context, cookieHashHex string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.preAuths, cookieHashHex)
	return nil
}

func (s *testStore) CheckRateLimit(ctx context.Context, scope, subject string, limit int64, window time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := scope + ":" + subject
	s.rates[k]++
	if s.rates[k] > limit {
		return s.rates[k], authsession.ErrRateLimitExceeded
	}
	return s.rates[k], nil
}

type noopRefresher struct{}

func (noopRefresher) RefreshUserAccessSnapshot(context.Context, string, authcontext.CallerContext) (*authz.AccessSnapshot, error) {
	return nil, nil
}

type srvMockAuth struct {
	iamv1.AuthServiceClient
	loginToken string
	loginCSRF  string
}

func (m *srvMockAuth) AuthenticatePassword(ctx context.Context, in *iamv1.AuthenticatePasswordRequest, _ ...grpc.CallOption) (*iamv1.AuthenticatePasswordResponse, error) {
	return &iamv1.AuthenticatePasswordResponse{
		Status:       iamv1.AuthStatus_AUTH_STATUS_AUTHENTICATED,
		RedirectPath: "/identity",
		IssuedSession: &iamv1.IssuedSession{
			SessionToken: m.loginToken,
			CsrfToken:    m.loginCSRF,
			Session: &iamv1.Session{
				AbsoluteExpiresAt: timestamppb.New(time.Now().Add(8 * time.Hour)),
			},
		},
	}, nil
}

type srvMockUsers struct {
	iamv1.UserServiceClient
}

func (m *srvMockUsers) ListUsers(ctx context.Context, in *iamv1.ListUsersRequest, _ ...grpc.CallOption) (*iamv1.ListUsersResponse, error) {
	now := time.Now()
	return &iamv1.ListUsersResponse{
		Items: []*iamv1.User{
			{
				Id:          "u-1",
				Kind:        iamv1.UserKind_USER_KIND_EMPLOYEE,
				Status:      iamv1.UserStatus_USER_STATUS_ACTIVE,
				DisplayName: "Admin Ops",
				CreatedAt:   timestamppb.New(now),
				UpdatedAt:   timestamppb.New(now),
			},
		},
		Page: &iamv1.PageInfo{Total: 1, Limit: 20, Offset: 0, HasMore: false},
	}, nil
}

func buildTestGatewayServer(t testing.TB, enableIntrospection bool) (*Server, *testStore, string, string) {
	t.Helper()
	return buildTestGatewayServerWithLogger(t, nil, enableIntrospection)
}

func buildTestGatewayServerWithLogger(t testing.TB, logger *slog.Logger, enableIntrospection bool) (*Server, *testStore, string, string) {
	t.Helper()
	return buildTestGatewayServerWithHRM(t, logger, enableIntrospection, nil)
}

func buildTestGatewayServerWithHRM(t testing.TB, logger *slog.Logger, enableIntrospection bool, hrmConn *grpc.ClientConn) (*Server, *testStore, string, string) {
	t.Helper()
	store := newTestStore()
	now := time.Now()

	loginRawTok := "session-raw-secret-123"
	loginTokHash := authsession.HashOpaqueToken(loginRawTok)
	loginRawCSRF := "session-csrf-secret-456"
	loginCSRFHash := authsession.HashOpaqueToken(loginRawCSRF)

	store.sessions[loginTokHash] = &authsession.SessionValidationContext{
		Token: &authsession.SessionTokenRecord{
			SessionID:     "sess-1",
			UserID:        "user-admin",
			TokenHashHex:  loginTokHash,
			CSRFDigestHex: loginCSRFHash,
			UserKind:      "employee",
			AuthEpoch:     1,
			IssuedAtUnix:  now.Unix(),
		},
		Security: &authsession.UserSecurityRecord{UserID: "user-admin", Status: authz.StatusActive, AuthEpoch: 1, PermRevision: 1, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID:         "user-admin",
			UserKind:            "employee",
			Status:              authz.StatusActive,
			AuthEpoch:           1,
			PermRevision:        1,
			RecoveryEpoch:       1,
			CanAccessIdentity:   true,
			DefaultRedirectPath: "/identity",
			SyncedAtUnix:        now.Unix(),
			FreshUntilUnix:      now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: "iam.users.read", Scope: authz.ScopeGrant{Type: authz.ScopeGlobal}},
			},
		},
	}

	loader := middleware.NewSnapshotLoader(noopRefresher{}, time.Second, 30*time.Second)
	authMgr, err := middleware.NewAuthManager(store, loader, middleware.SessionConfig{
		CookieName:         config.SessionCookieName,
		CookieSecure:       true,
		PreAuthTTL:         15 * time.Minute,
		SessionIdleTimeout: 30 * time.Minute,
		SessionMaxAge:      8 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthManager failed: %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	ov, _ := browserauth.NewOriginValidator([]string{"https://erp.localhost:8443"})
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := iamadapter.NewAdapter(source_connections.IAMClientSet{
		Auth:  &srvMockAuth{loginToken: loginRawTok, loginCSRF: loginRawCSRF},
		Users: &srvMockUsers{},
	}, authMgr, authzSvc, rv, iamadapter.Config{LoginRateMax: 50, LoginRateWindow: time.Minute})
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	cfg := &config.Config{
		AppEnv:                    "production",
		ServiceName:               "integrione-gateway-test",
		HTTPHost:                  "127.0.0.1",
		HTTPPort:                  8080,
		MaxComplexity:             250,
		EnableIntrospection:       enableIntrospection,
		AllowedOrigins:            []string{"https://erp.localhost:8443"},
		CookieName:                config.SessionCookieName,
		CookieSecure:              true,
		InternalGatewayHMACSecret: "0123456789abcdef0123456789abcdef",
	}

	execSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			IAMAdapter: adp,
			HRMAdapter: hrmadapter.NewAdapter(authMgr, authzSvc, hrmConn),
		},
		Complexity: resolvers.NewComplexityRoot(cfg.MaxComplexity),
	})

	srv, err := New(Dependencies{
		Config:          cfg,
		Logger:          logger,
		Schema:          execSchema,
		AuthManager:     authMgr,
		OriginValidator: ov,
	})
	if err != nil {
		t.Fatalf("server.New failed: %v", err)
	}
	return srv, store, loginRawTok, loginRawCSRF
}

func doGraphQLPOST(t *testing.T, srv *Server, body map[string]any, headers map[string]string, cookies []*http.Cookie) *http.Response {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/graphql", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := srv.App().Test(req)
	if err != nil {
		t.Fatalf("app.Test POST failed: %v", err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, _ := io.ReadAll(resp.Body)
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
	return string(b)
}

func TestSingleEndpointRoutingAndSecurityGuards(t *testing.T) {
	srv, _, rawTok, rawCSRF := buildTestGatewayServer(t, false)

	// 1. Health GET /api/graphql?query={health} -> HTTP 200 ok.
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/graphql?query=%7Bhealth%7D", nil)
	respHealth, err := srv.App().Test(reqHealth)
	if err != nil || respHealth.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 on health query, got err=%v status=%d", err, respHealth.StatusCode)
	}
	if body := readBody(t, respHealth); !strings.Contains(body, `"health":"ok"`) {
		t.Fatalf("unexpected health body: %s", body)
	}

	// 2. Endpoint warisan non-GraphQL wajib 404.
	for _, p := range []string{"/api/auth/login", "/healthz", "/metrics", "/graphql"} {
		r := httptest.NewRequest(http.MethodGet, p, nil)
		res, _ := srv.App().Test(r)
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 on %s, got %d", p, res.StatusCode)
		}
		_ = res.Body.Close()
	}

	// 3. Mutasi via HTTP GET wajib ditolak oleh transport.GET / guard (406 atau FORBIDDEN).
	reqGetMut := httptest.NewRequest(http.MethodGet, "/api/graphql?query=mutation%7B_empty%7D", nil)
	resGetMut, _ := srv.App().Test(reqGetMut)
	bodyGetMut := readBody(t, resGetMut)
	if resGetMut.StatusCode != http.StatusNotAcceptable && !strings.Contains(bodyGetMut, "GET requests only allow query operations") && !strings.Contains(bodyGetMut, "FORBIDDEN") {
		t.Fatalf("expected GET mutation rejection, got status=%d body=%s", resGetMut.StatusCode, bodyGetMut)
	}

	// 4. Mutasi tanpa Origin atau dengan Origin asing ditolak.
	resNoOrigin := doGraphQLPOST(t, srv, map[string]any{"query": "mutation { _empty }"}, map[string]string{browserauth.CSRFHeaderName: rawCSRF}, nil)
	if b := readBody(t, resNoOrigin); !strings.Contains(b, "FORBIDDEN") {
		t.Fatalf("expected missing Origin rejection, got %s", b)
	}

	resForeignOrigin := doGraphQLPOST(t, srv, map[string]any{"query": "mutation { _empty }"}, map[string]string{
		"Origin":                   "https://evil.com",
		browserauth.CSRFHeaderName: rawCSRF,
	}, nil)
	if b := readBody(t, resForeignOrigin); !strings.Contains(b, "FORBIDDEN") {
		t.Fatalf("expected foreign Origin rejection, got %s", b)
	}

	// 5. Introspection nonaktif saat production.
	resIntrospect := doGraphQLPOST(t, srv, map[string]any{"query": "{ __schema { types { name } } }"}, nil, nil)
	if b := readBody(t, resIntrospect); !strings.Contains(b, "introspection") && !strings.Contains(b, "FORBIDDEN") {
		t.Fatalf("expected disabled introspection rejection, got %s", b)
	}

	// 6. Penolakan batas kompleksitas paginasi berlebih (>100 pada first/last Relay) serta ErrorPresenter extensions.type pada Query error.
	sessCookie := &http.Cookie{Name: config.SessionCookieName, Value: rawTok}
	resComplex := doGraphQLPOST(t, srv, map[string]any{
		"query": `{ users(first: 500) { edges { cursor node { id } } pageInfo { startCursor endCursor hasPreviousPage hasNextPage } total_count } }`,
	}, nil, []*http.Cookie{sessCookie})
	if b := readBody(t, resComplex); !strings.Contains(b, "COMPLEXITY_LIMIT_EXCEEDED") {
		t.Fatalf("expected COMPLEXITY_LIMIT_EXCEEDED, got %s", b)
	}

	resUnauthQuery := doGraphQLPOST(t, srv, map[string]any{
		"query": `{ users(first: 1) { total_count } }`,
	}, nil, nil)
	if b := readBody(t, resUnauthQuery); !strings.Contains(b, `"type":"AUTHENTICATION"`) || !strings.Contains(b, `"code":"UNAUTHENTICATED"`) {
		t.Fatalf("expected ErrorPresenter extensions.type=AUTHENTICATION and extensions.code=UNAUTHENTICATED, got %s", b)
	}

	// 7. Alur lengkap BrowserSession -> UserLogin (MutationResult payload) -> SSR Query cookie-only -> Query users snake_case.
	resBoot := doGraphQLPOST(t, srv, map[string]any{
		"query": `query { browserSession { authenticated principalId csrfToken } }`,
	}, nil, nil)
	preCookies := resBoot.Cookies()
	preCSRF := resBoot.Header.Get(browserauth.CSRFHeaderName)
	_ = readBody(t, resBoot)
	if len(preCookies) == 0 || preCSRF == "" {
		t.Fatalf("expected pre-auth cookie and X-CSRF-Token header from browserSession")
	}

	// Validasi mutasi gagal mengembalikan was_successful=false dan errors[{type,message}].
	resBadReset := doGraphQLPOST(t, srv, map[string]any{
		"query": `mutation($input: UserResetPasswordInput!) { userResetPassword(input: $input) { message errors { type message } was_successful } }`,
		"variables": map[string]any{
			"input": map[string]any{"email": "   "},
		},
	}, map[string]string{
		"Origin":                   "https://erp.localhost:8443",
		browserauth.CSRFHeaderName: preCSRF,
	}, preCookies)
	bodyBadReset := readBody(t, resBadReset)
	if !strings.Contains(bodyBadReset, `"was_successful":false`) || !strings.Contains(bodyBadReset, `"type":"ARGUMENT"`) || !strings.Contains(bodyBadReset, `"message":"email is required"`) {
		t.Fatalf("expected structured MutationResult ARGUMENT error on userResetPassword, got %s", bodyBadReset)
	}

	resLogin := doGraphQLPOST(t, srv, map[string]any{
		"query": `mutation($input: UserLoginInput!) { userLogin(input: $input) { status csrfToken redirectLocation errors { type message } was_successful } }`,
		"variables": map[string]any{
			"input": map[string]any{"identifier": "admin@integrione.local", "password": "secret"},
		},
	}, map[string]string{
		"Origin":                   "https://erp.localhost:8443",
		browserauth.CSRFHeaderName: preCSRF,
	}, preCookies)
	bodyLogin := readBody(t, resLogin)
	if !strings.Contains(bodyLogin, `"AUTHENTICATED"`) || !strings.Contains(bodyLogin, `"was_successful":true`) || resLogin.Header.Get(browserauth.LocationHeaderName) != "/identity" {
		t.Fatalf("expected promoted login session with Location /identity and was_successful=true, got %s (Location=%q)", bodyLogin, resLogin.Header.Get(browserauth.LocationHeaderName))
	}
	promotedCookies := resLogin.Cookies()

	// SSR Cookie-only POST query tanpa Origin/CSRF harus berhasil dan membaca Relay Connection serta GID/field snake_case.
	resSSR := doGraphQLPOST(t, srv, map[string]any{
		"query": `{ users(first: 10) { edges { cursor node { id _id kind status display_name } } nodes { id display_name } pageInfo { startCursor endCursor hasPreviousPage hasNextPage } total_count } }`,
	}, nil, promotedCookies)
	bodySSR := readBody(t, resSSR)
	if !strings.Contains(bodySSR, `"display_name":"Admin Ops"`) || !strings.Contains(bodySSR, `"hasNextPage":false`) || !strings.Contains(bodySSR, `"total_count":1`) {
		t.Fatalf("expected Relay snake_case users response on SSR cookie-only POST, got %s", bodySSR)
	}
}

func TestGraphQLDepthAliasLimitsJSONMaskingAndSecurityHeaders(t *testing.T) {
	srv, _, rawTok, _ := buildTestGatewayServer(t, false)
	sessCookie := &http.Cookie{Name: config.SessionCookieName, Value: rawTok}

	// 1. Header keamanan HTTP standar wajib selalu dikirim pada endpoint tunggal /api/graphql.
	reqHdr := httptest.NewRequest(http.MethodGet, "/api/graphql?query=%7Bhealth%7D", nil)
	respHdr, err := srv.App().Test(reqHdr)
	if err != nil {
		t.Fatalf("app.Test GET failed: %v", err)
	}
	_ = readBody(t, respHdr)
	expectedHeaders := map[string]string{
		"X-Frame-Options":              "DENY",
		"Content-Security-Policy":      "default-src 'none'; frame-ancestors 'none'",
		"Permissions-Policy":           "camera=(), microphone=(), geolocation=()",
		"Cross-Origin-Resource-Policy": "same-origin",
		"X-Content-Type-Options":       "nosniff",
		"Referrer-Policy":              "same-origin",
	}
	for k, want := range expectedHeaders {
		if got := respHdr.Header.Get(k); got != want {
			t.Fatalf("expected header %s=%q, got %q", k, want, got)
		}
	}

	// 2. JSON payload rusak wajib dimasking tanpa membocorkan detail parser Go encoding/json.
	reqBadJSON := httptest.NewRequest(http.MethodPost, "/api/graphql", strings.NewReader("{bad-json"))
	reqBadJSON.Header.Set("Content-Type", "application/json")
	respBadJSON, err := srv.App().Test(reqBadJSON)
	if err != nil {
		t.Fatalf("app.Test bad JSON failed: %v", err)
	}
	bodyBadJSON := readBody(t, respBadJSON)
	if !strings.Contains(bodyBadJSON, "invalid JSON request body") || strings.Contains(bodyBadJSON, "invalid character") {
		t.Fatalf("expected masked JSON decode error without parser details, got %s", bodyBadJSON)
	}

	// 3. Query GraphQL dengan >25 alias wajib ditolak oleh guard alias.
	resAliases := doGraphQLPOST(t, srv, map[string]any{
		"query": `{ a01: health a02: health a03: health a04: health a05: health a06: health a07: health a08: health a09: health a10: health a11: health a12: health a13: health a14: health a15: health a16: health a17: health a18: health a19: health a20: health a21: health a22: health a23: health a24: health a25: health a26: health }`,
	}, nil, nil)
	if b := readBody(t, resAliases); !strings.Contains(b, "graphql query exceeds maximum alias count") {
		t.Fatalf("expected alias limit error, got %s", b)
	}

	// 4. Query GraphQL dengan kedalaman >10 wajib ditolak oleh guard depth.
	deepQuery := `{ users(first: 1) { edges { node { groups(first: 1) { edges { node { roles(first: 1) { edges { node { permissions(first: 1) { total_count } } } } } } } } } } }`
	resDeep := doGraphQLPOST(t, srv, map[string]any{
		"query": deepQuery,
	}, nil, []*http.Cookie{sessCookie})
	if b := readBody(t, resDeep); !strings.Contains(b, "graphql query exceeds maximum depth") {
		t.Fatalf("expected max depth error, got %s", b)
	}
}

func TestServerPanicRecoveryAndErrorPresenter(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// 1. Uji middleware recover Fiber langsung di luar gqlgen RecoverFunc agar panic HTTP apa pun tidak menjatuhkan proses Gateway.
	panicApp := fiber.New()
	panicApp.Use(requestPanicRecoverMiddleware(logger))
	panicApp.Get("/panic-probe", func(c fiber.Ctx) error {
		panic("unexpected route panic secret=topsecret")
	})
	reqPanic := httptest.NewRequest(http.MethodGet, "/panic-probe", nil)
	respPanic, err := panicApp.Test(reqPanic)
	if err != nil {
		t.Fatalf("app.Test panic-probe failed: %v", err)
	}
	bodyPanic := readBody(t, respPanic)
	if respPanic.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected HTTP 500 after recovered panic, got %d (%s)", respPanic.StatusCode, bodyPanic)
	}
	if !strings.Contains(bodyPanic, `"code":"INTERNAL_ERROR"`) || strings.Contains(bodyPanic, "topsecret") {
		t.Fatalf("expected masked INTERNAL_ERROR JSON response without panic leak, got %s", bodyPanic)
	}

	// 2. Uji ErrorPresenter menangani error nil tanpa panic serta memisahkan level log ERROR/WARN dengan request_id.
	presenter := buildErrorPresenter(logger)
	ctx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RequestID: "req-pres-1"})
	if nilErrOut := presenter(ctx, nil); nilErrOut == nil || nilErrOut.Extensions["code"] != string(middleware.CodeInternalError) {
		t.Fatalf("expected masked INTERNAL_ERROR for nil presenter input, got %+v", nilErrOut)
	}

	_ = presenter(ctx, middleware.NewGatewayError(middleware.CodeInternalError, "internal server error", context.DeadlineExceeded))
	_ = presenter(ctx, middleware.ErrUnavailable("identity service unavailable", context.DeadlineExceeded))

	// 3. Pastikan singleErrorResponse meneruskan ctx ke ErrorPresenter sehingga request_id tetap tercatat di log.
	ctxGuard := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{RequestID: "req-guard-1"})
	respHandler := singleErrorResponse(ctxGuard, presenter, middleware.ErrUnavailable("authentication manager unavailable", nil))
	guardResp := respHandler(ctxGuard)
	if guardResp == nil || len(guardResp.Errors) != 1 || guardResp.Errors[0].Extensions["code"] != string(middleware.CodeUnavailable) {
		t.Fatalf("expected singleErrorResponse UNAVAILABLE, got %+v", guardResp)
	}

	logOut := logBuf.String()
	if !strings.Contains(logOut, `"level":"ERROR"`) || !strings.Contains(logOut, `"msg":"gateway internal error"`) {
		t.Fatalf("expected ERROR log for CodeInternalError, got %s", logOut)
	}
	if !strings.Contains(logOut, `"level":"WARN"`) || !strings.Contains(logOut, `"msg":"gateway upstream unavailable"`) {
		t.Fatalf("expected WARN log for CodeUnavailable, got %s", logOut)
	}
	if !strings.Contains(logOut, `"request_id":"req-pres-1"`) || !strings.Contains(logOut, `"request_id":"req-guard-1"`) {
		t.Fatalf("expected request_id correlation in ErrorPresenter log, got %s", logOut)
	}
}
