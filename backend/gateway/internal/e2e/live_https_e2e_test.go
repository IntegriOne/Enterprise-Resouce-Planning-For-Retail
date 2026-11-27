//go:build e2e

// Package e2e menjalankan pengujian integrasi langsung end-to-end terhadap Caddy HTTPS,
// dua replika Gateway, layanan gRPC IAM, Redis TLS ber-ACL, dan PostgreSQL 18.
package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/integrione/erp-retail/backend/shared/authsession"
)

type e2eRuntimeConfig struct {
	repoRoot              string
	caCertPath            string
	baseURL               string
	superadminUsername    string
	superadminPassword    string
	autoSystemIAMUsername string
	redisAddr             string
	gwRedisPass           string
	iamRedisPass          string
	postgresService       string
}

type graphQLError struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

type browserClient struct {
	t          *testing.T
	httpClient *http.Client
	baseURL    string
	csrfToken  string
}

func loadE2ERuntimeConfig(t *testing.T) e2eRuntimeConfig {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", "..", ".."))

	iamEnv := parseEnvFile(t, filepath.Join(repoRoot, "backend", "services", "iam", ".env"))
	redisSecrets := parseEnvFile(t, filepath.Join(repoRoot, "backend", "deploy", "redis", ".dev-redis-secrets"))

	superUser := envOr(iamEnv, "E2E_SUPERADMIN_USERNAME", "IAM_SUPERADMIN_USERNAME", "superadmin")
	superPass := envOr(iamEnv, "E2E_SUPERADMIN_PASSWORD", "IAM_SUPERADMIN_PASSWORD", "")
	autoIAMUser := envOr(iamEnv, "E2E_AUTO_SYSTEM_IAM_USERNAME", "IAM_AUTO_SYSTEM_IAM_USERNAME", "auto_system_iam")
	gwRedisPass := envOr(redisSecrets, "E2E_GW_REDIS_PASSWORD", "GW_REDIS_PW", "")
	iamRedisPass := envOr(redisSecrets, "E2E_IAM_REDIS_PASSWORD", "IAM_REDIS_PW", "")

	if superPass == "" || gwRedisPass == "" || iamRedisPass == "" {
		t.Fatal("missing runtime credentials from local untracked .env / .dev-redis-secrets")
	}

	return e2eRuntimeConfig{
		repoRoot:              repoRoot,
		caCertPath:            filepath.Join(repoRoot, "backend", "deploy", "certs", "ca.crt"),
		baseURL:               "https://erp.localhost:8443",
		superadminUsername:    superUser,
		superadminPassword:    superPass,
		autoSystemIAMUsername: autoIAMUser,
		redisAddr:             "127.0.0.1:6379",
		gwRedisPass:           gwRedisPass,
		iamRedisPass:          iamRedisPass,
		postgresService:       "integrione-retail-postgres-1",
	}
}

func parseEnvFile(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open env file %s: %v", path, err)
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		values[strings.TrimSpace(k)] = v
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan env file %s: %v", path, err)
	}
	return values
}

func envOr(m map[string]string, envKey, fileKey, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		return v
	}
	if v := strings.TrimSpace(m[fileKey]); v != "" {
		return v
	}
	return fallback
}

func newHTTPSBrowserClient(t *testing.T, cfg e2eRuntimeConfig) *browserClient {
	t.Helper()
	caPEM, err := os.ReadFile(cfg.caCertPath)
	if err != nil {
		t.Fatalf("read ca.crt: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("append ca.crt to cert pool failed")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    pool,
			ServerName: "erp.localhost",
			MinVersion: tls.VersionTLS12,
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.HasPrefix(addr, "erp.localhost:") {
				_, port, _ := net.SplitHostPort(addr)
				addr = net.JoinHostPort("127.0.0.1", port)
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	return &browserClient{
		t:       t,
		baseURL: cfg.baseURL,
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Jar:       jar,
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *browserClient) doGraphQL(query string, vars map[string]any, headers map[string]string) (*http.Response, graphQLResponse) {
	c.t.Helper()
	payload := map[string]any{"query": query}
	if vars != nil {
		payload["variables"] = vars
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatalf("marshal graphql payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/graphql", bytes.NewReader(raw))
	if err != nil {
		c.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if _, hasOrigin := headers["Origin"]; !hasOrigin {
		req.Header.Set("Origin", c.baseURL)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
	}
	if _, hasCSRF := headers["X-CSRF-Token"]; !hasCSRF && c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}
	for k, v := range headers {
		if v == "" {
			req.Header.Del(k)
		} else {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.t.Fatalf("http POST /api/graphql failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("read response body: %v", err)
	}

	if newCSRF := strings.TrimSpace(resp.Header.Get("X-CSRF-Token")); newCSRF != "" {
		c.csrfToken = newCSRF
	}

	var decoded graphQLResponse
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		c.t.Fatalf("unmarshal graphql response (status=%d, body=%s): %v", resp.StatusCode, string(bodyBytes), err)
	}
	return resp, decoded
}

func (c *browserClient) sessionCookie() *http.Cookie {
	c.t.Helper()
	u, _ := url.Parse(c.baseURL)
	for _, ck := range c.httpClient.Jar.Cookies(u) {
		if ck.Name == "__Host-integrione_session" {
			return ck
		}
	}
	return nil
}

func (c *browserClient) bootstrapSession() *http.Cookie {
	c.t.Helper()
	resp, gql := c.doGraphQL(`query { browserSession { authenticated principalId csrfToken } }`, nil, map[string]string{
		"X-CSRF-Token": "",
	})
	if resp.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
		c.t.Fatalf("bootstrap browserSession failed: status=%d errors=%+v", resp.StatusCode, gql.Errors)
	}
	var out struct {
		BrowserSession struct {
			Authenticated bool    `json:"authenticated"`
			PrincipalID   *string `json:"principalId"`
			CSRFToken     string  `json:"csrfToken"`
		} `json:"browserSession"`
	}
	if err := json.Unmarshal(gql.Data, &out); err != nil {
		c.t.Fatalf("decode browserSession: %v", err)
	}
	if out.BrowserSession.CSRFToken == "" {
		c.t.Fatal("expected non-empty csrfToken from browserSession")
	}
	c.csrfToken = out.BrowserSession.CSRFToken
	var sessionCk *http.Cookie
	for _, ck := range resp.Cookies() {
		if ck.Name == "__Host-integrione_session" {
			sessionCk = ck
			break
		}
	}
	if sessionCk == nil || sessionCk.Value == "" {
		c.t.Fatal("expected __Host-integrione_session Set-Cookie header to be sent")
	}
	return sessionCk
}

func (c *browserClient) login(identifier, password string) (string, string, string) {
	c.t.Helper()
	resp, gql := c.doGraphQL(`mutation Login($input: UserLoginInput!) {
		userLogin(input: $input) {
			status
			mfaChallenge
			csrfToken
			redirectLocation
			errors { type message }
			was_successful
		}
	}`, map[string]any{
		"input": map[string]any{
			"identifier": identifier,
			"password":   password,
		},
	}, nil)
	if resp.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
		c.t.Fatalf("userLogin failed for %s: status=%d errors=%+v", identifier, resp.StatusCode, gql.Errors)
	}
	var out struct {
		UserLogin struct {
			Status           string `json:"status"`
			MFAChallenge     string `json:"mfaChallenge"`
			CSRFToken        string `json:"csrfToken"`
			RedirectLocation string `json:"redirectLocation"`
			WasSuccessful    bool   `json:"was_successful"`
			Errors           []struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"userLogin"`
	}
	if err := json.Unmarshal(gql.Data, &out); err != nil {
		c.t.Fatalf("decode userLogin: %v", err)
	}
	if !out.UserLogin.WasSuccessful || len(out.UserLogin.Errors) > 0 {
		c.t.Fatalf("userLogin unsuccessful for %s: %+v", identifier, out.UserLogin)
	}
	if out.UserLogin.CSRFToken != "" {
		c.csrfToken = out.UserLogin.CSRFToken
	}
	return out.UserLogin.Status, out.UserLogin.MFAChallenge, out.UserLogin.RedirectLocation
}

func requireErrorCode(t *testing.T, gql graphQLResponse, wantCode string) {
	t.Helper()
	if len(gql.Errors) == 0 {
		t.Fatalf("expected GraphQL error %s, got success data=%s", wantCode, string(gql.Data))
	}
	code, _ := gql.Errors[0].Extensions["code"].(string)
	if code != wantCode {
		t.Fatalf("expected GraphQL error code %q, got %q (full error=%+v)", wantCode, code, gql.Errors[0])
	}
}

func requireMutationPayloadError(t *testing.T, gql graphQLResponse, opName string, wantTypes ...string) (string, string) {
	t.Helper()
	if len(gql.Errors) > 0 {
		t.Fatalf("%s returned top-level GraphQL errors=%+v, want MutationResult payload error", opName, gql.Errors)
	}
	var root map[string]map[string]any
	if err := json.Unmarshal(gql.Data, &root); err != nil {
		t.Fatalf("%s unmarshal data=%s: %v", opName, string(gql.Data), err)
	}
	payload := root[opName]
	if payload == nil {
		t.Fatalf("%s missing payload in data=%s", opName, string(gql.Data))
	}
	wasOK, ok := payload["was_successful"].(bool)
	if !ok || wasOK {
		t.Fatalf("%s expected was_successful=false, got %+v", opName, payload)
	}
	rawErrs, ok := payload["errors"].([]any)
	if !ok || len(rawErrs) == 0 {
		t.Fatalf("%s expected non-empty errors[], got %+v", opName, payload)
	}
	first, _ := rawErrs[0].(map[string]any)
	errType, _ := first["type"].(string)
	errMsg, _ := first["message"].(string)
	if strings.TrimSpace(errType) == "" || strings.TrimSpace(errMsg) == "" {
		t.Fatalf("%s expected non-empty type and message in MutationError, got %+v", opName, first)
	}
	if len(wantTypes) > 0 {
		matched := false
		for _, wt := range wantTypes {
			if errType == wt {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("%s expected MutationError.type in %v, got %q (msg=%q)", opName, wantTypes, errType, errMsg)
		}
	}
	return errType, errMsg
}

func psqlScalar(t *testing.T, cfg e2eRuntimeConfig, query string) string {
	t.Helper()
	cmd := exec.Command("docker", "exec", cfg.postgresService, "psql", "-U", "postgres", "-d", "iam", "-tAc", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("psql query %q failed: %v\noutput: %s", query, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func composeCmd(t *testing.T, cfg e2eRuntimeConfig, args ...string) string {
	t.Helper()
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Dir = cfg.repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker compose %v failed: %v\noutput: %s", args, err, string(out))
	}
	return string(out)
}

func newE2ERedisClient(t *testing.T, cfg e2eRuntimeConfig, username, password string) *redis.Client {
	t.Helper()
	client, err := authsession.NewTLSClient(authsession.RedisConfig{
		Addr:                  cfg.redisAddr,
		Username:              username,
		Password:              password,
		CACertPath:            cfg.caCertPath,
		ServerName:            "redis",
		TLSInsecureSkipVerify: false,
	})
	if err != nil {
		t.Fatalf("NewTLSClient(%s): %v", username, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := authsession.Ping(ctx, client); err != nil {
		_ = client.Close()
		t.Fatalf("redis ping (%s): %v", username, err)
	}
	return client
}

func generateTOTPCode(t *testing.T, otpAuthURI string, now time.Time) string {
	t.Helper()
	u, err := url.Parse(otpAuthURI)
	if err != nil {
		t.Fatalf("parse otpauth URI: %v", err)
	}
	secret := strings.ToUpper(strings.TrimSpace(u.Query().Get("secret")))
	if secret == "" {
		t.Fatalf("missing secret in otpauth URI: %s", otpAuthURI)
	}
	if mod := len(secret) % 8; mod != 0 {
		secret += strings.Repeat("=", 8-mod)
	}
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Fatalf("decode base32 secret: %v", err)
	}

	counter := uint64(now.Unix() / 30)
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	codeInt := (int(sum[offset]&0x7f) << 24) |
		(int(sum[offset+1]&0xff) << 16) |
		(int(sum[offset+2]&0xff) << 8) |
		int(sum[offset+3]&0xff)
	return fmt.Sprintf("%06d", codeInt%1000000)
}

func TestLiveHTTPSIAMGatewayE2E(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)

	stripGID := func(id string) string {
		if idx := strings.LastIndex(id, "/"); idx >= 0 {
			return id[idx+1:]
		}
		return id
	}

	// Siapkan cleanup data uji E2E di PostgreSQL agar environment dev tetap bersih.
	var createdUserIDs []string
	var createdRoleIDs []string
	var createdGroupIDs []string
	t.Cleanup(func() {
		for _, id := range createdUserIDs {
			_ = exec.Command("docker", "exec", cfg.postgresService, "psql", "-U", "postgres", "-d", "iam", "-c",
				fmt.Sprintf("DELETE FROM users WHERE id = '%s';", stripGID(id))).Run()
		}
		for _, id := range createdGroupIDs {
			_ = exec.Command("docker", "exec", cfg.postgresService, "psql", "-U", "postgres", "-d", "iam", "-c",
				fmt.Sprintf("DELETE FROM groups WHERE id = '%s';", stripGID(id))).Run()
		}
		for _, id := range createdRoleIDs {
			_ = exec.Command("docker", "exec", cfg.postgresService, "psql", "-U", "postgres", "-d", "iam", "-c",
				fmt.Sprintf("DELETE FROM roles WHERE id = '%s';", stripGID(id))).Run()
		}
	})

	anon := newHTTPSBrowserClient(t, cfg)

	t.Run("SingleEndpointAndFrontendRoutes", func(t *testing.T) {
		resp, err := anon.httpClient.Get(cfg.baseURL + "/api/graphql?query=%7Bhealth%7D")
		if err != nil {
			t.Fatalf("GET health: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"health":"ok"`) {
			t.Fatalf("unexpected /api/graphql health response: status=%d body=%s", resp.StatusCode, string(body))
		}

		// Pastikan rute REST lama atau endpoint tambahan tidak dipublikasikan.
		for _, path := range []string{"/api/auth/login", "/healthz", "/metrics"} {
			r, err := anon.httpClient.Get(cfg.baseURL + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			_ = r.Body.Close()
			if r.StatusCode != http.StatusNotFound {
				t.Fatalf("expected 404 on %s, got %d", path, r.StatusCode)
			}
		}

		// Pastikan GET mutation ditolak sebelum masuk resolver.
		mutGet, err := anon.httpClient.Get(cfg.baseURL + "/api/graphql?query=mutation+%7B+userLogout+%7B+logged_out+%7D+%7D")
		if err != nil {
			t.Fatalf("GET mutation: %v", err)
		}
		_ = mutGet.Body.Close()
		if mutGet.StatusCode != http.StatusNotAcceptable {
			t.Fatalf("expected 406 on GET mutation, got %d", mutGet.StatusCode)
		}

		// Pastikan batas complexity > 100 menolak query paginasi berbeban tinggi.
		_, compErr := anon.doGraphQL(`query { users(first: 500) { edges { node { id } } pageInfo { hasNextPage } total_count } }`, nil, nil)
		requireErrorCode(t, compErr, "COMPLEXITY_LIMIT_EXCEEDED")

		// Pastikan halaman Next.js /login & /verify-email render 200, /register mengembalikan 404 (registrasi ditangani via view di /login), dan /identity redirect 307 ke /login saat anonim.
		for _, page := range []string{"/login", "/verify-email"} {
			r, err := anon.httpClient.Get(cfg.baseURL + page)
			if err != nil {
				t.Fatalf("GET %s: %v", page, err)
			}
			_ = r.Body.Close()
			if r.StatusCode != http.StatusOK {
				t.Fatalf("expected 200 on %s, got %d", page, r.StatusCode)
			}
		}
		regResp, err := anon.httpClient.Get(cfg.baseURL + "/register")
		if err != nil {
			t.Fatalf("GET /register: %v", err)
		}
		_ = regResp.Body.Close()
		if regResp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 on removed /register route, got %d", regResp.StatusCode)
		}
		idResp, err := anon.httpClient.Get(cfg.baseURL + "/identity")
		if err != nil {
			t.Fatalf("GET /identity: %v", err)
		}
		_ = idResp.Body.Close()
		if idResp.StatusCode != http.StatusTemporaryRedirect || idResp.Header.Get("Location") != "/login" {
			t.Fatalf("expected 307 -> /login on unauthenticated /identity, got %d loc=%q", idResp.StatusCode, idResp.Header.Get("Location"))
		}
	})

	admin := newHTTPSBrowserClient(t, cfg)
	t.Run("BrowserBootstrapOriginCSRFAndSuperAdminLogin", func(t *testing.T) {
		ck := admin.bootstrapSession()
		if !ck.HttpOnly || !ck.Secure || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/" {
			t.Fatalf("unexpected Set-Cookie attributes: %+v", ck)
		}

		// Origin tidak tepercaya harus ditolak.
		_, badOrigin := admin.doGraphQL(`mutation { userLogout { logged_out } }`, nil, map[string]string{
			"Origin": "https://evil.example",
		})
		requireErrorCode(t, badOrigin, "FORBIDDEN")

		// Token CSRF kosong atau salah saat cookie ada harus ditolak.
		_, missingCSRF := admin.doGraphQL(`mutation { userLogout { logged_out } }`, nil, map[string]string{
			"X-CSRF-Token": "",
		})
		requireErrorCode(t, missingCSRF, "FORBIDDEN")
		if !strings.Contains(strings.ToLower(missingCSRF.Errors[0].Message), "csrf") {
			t.Fatalf("expected CSRF message for missing token, got %q", missingCSRF.Errors[0].Message)
		}

		_, badCSRF := admin.doGraphQL(`mutation { userLogout { logged_out } }`, nil, map[string]string{
			"X-CSRF-Token": "forged-csrf-token",
		})
		requireErrorCode(t, badCSRF, "FORBIDDEN")
		if !strings.Contains(strings.ToLower(badCSRF.Errors[0].Message), "csrf") {
			t.Fatalf("expected CSRF message for mismatched token, got %q", badCSRF.Errors[0].Message)
		}

		// Akun otomasi non-interaktif (auto_system_iam / auto_system_ef), baik dengan kata sandi kosong maupun terisi,
		// wajib ditolak pada endpoint login interaktif dengan pesan kontrak eksplisit "system cannot login".
		for _, autoCase := range []struct {
			identifier string
			password   string
		}{
			{identifier: cfg.autoSystemIAMUsername, password: ""},
			{identifier: strings.ToUpper(cfg.autoSystemIAMUsername), password: ""},
			{identifier: cfg.autoSystemIAMUsername, password: cfg.superadminPassword},
			{identifier: "auto_system_ef", password: ""},
		} {
			autoResp, autoGql := admin.doGraphQL(`mutation($input: UserLoginInput!) {
				userLogin(input: $input) { status errors { type message } was_successful }
			}`, map[string]any{
				"input": map[string]any{"identifier": autoCase.identifier, "password": autoCase.password},
			}, nil)
			if autoResp.StatusCode != http.StatusOK || len(autoGql.Errors) > 0 {
				t.Fatalf("expected GraphQL mutation error when %s attempts login: status=%d errs=%+v", autoCase.identifier, autoResp.StatusCode, autoGql.Errors)
			}
			var autoOut struct {
				UserLogin struct {
					WasSuccessful bool `json:"was_successful"`
					Errors        []struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					} `json:"errors"`
				} `json:"userLogin"`
			}
			_ = json.Unmarshal(autoGql.Data, &autoOut)
			if autoOut.UserLogin.WasSuccessful || len(autoOut.UserLogin.Errors) != 1 ||
				autoOut.UserLogin.Errors[0].Type != "AUTHORIZATION" ||
				autoOut.UserLogin.Errors[0].Message != "system cannot login" {
				t.Fatalf("expected AUTHORIZATION 'system cannot login' for (%q, passLen=%d), got %+v", autoCase.identifier, len(autoCase.password), autoOut.UserLogin)
			}
		}

		preLoginCookie := ck.Value
		preLoginCSRF := admin.csrfToken
		status, _, redirectLoc := admin.login(cfg.superadminUsername, cfg.superadminPassword)
		if status != "AUTHENTICATED" {
			t.Fatalf("expected AUTHENTICATED for human superadmin, got %s", status)
		}
		if redirectLoc == "" || !strings.HasPrefix(redirectLoc, "/") || strings.HasPrefix(redirectLoc, "//") {
			t.Fatalf("unexpected redirectLocation %q", redirectLoc)
		}
		postLoginCookie := admin.sessionCookie().Value
		if postLoginCookie == preLoginCookie || admin.csrfToken == preLoginCSRF {
			t.Fatal("expected session token and CSRF token rotation on login")
		}

		// Verifikasi permintaan SSR tanpa header Origin tetapi membawa Cookie+CSRF tetap sah,
		// sekaligus menguji konsistensi sesi di balik load balancer Caddy (gateway-1 & gateway-2).
		for i := 0; i < 6; i++ {
			resp, gql := admin.doGraphQL(`query { myWorkspaceAccess { canAccessIdentity defaultLocation } }`, nil, map[string]string{
				"Origin":         "",
				"Sec-Fetch-Site": "",
			})
			if resp.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
				t.Fatalf("SSR request %d failed: status=%d errors=%+v", i, resp.StatusCode, gql.Errors)
			}
			var ws struct {
				MyWorkspaceAccess struct {
					CanAccessIdentity bool   `json:"canAccessIdentity"`
					DefaultLocation   string `json:"defaultLocation"`
				} `json:"myWorkspaceAccess"`
			}
			_ = json.Unmarshal(gql.Data, &ws)
			if !ws.MyWorkspaceAccess.CanAccessIdentity || ws.MyWorkspaceAccess.DefaultLocation != "/identity" {
				t.Fatalf("unexpected workspace access payload: %+v", ws.MyWorkspaceAccess)
			}
		}
	})

	var customerID, employeeID, vendorID, supplierID string
	var customRoleID, customGroupID string
	custEmail := fmt.Sprintf("e2e-cust-%s@integrione.local", suffix)
	empEmail := fmt.Sprintf("e2e-emp-%s@integrione.local", suffix)
	empUsername := "e2e_emp_" + suffix
	vendEmail := fmt.Sprintf("e2e-vend-%s@integrione.local", suffix)
	supEmail := fmt.Sprintf("e2e-sup-%s@integrione.local", suffix)
	commonPass := "StrongPass#2026!"

	t.Run("UserVariantsSnakeCaseAndRBACLifecycle", func(t *testing.T) {
		createCase := func(op, inputType string, extraFields map[string]any, wantKind string) string {
			q := fmt.Sprintf(`mutation Create($input: %s!) {
				res: %s(input: $input) {
					user {
						id
						_id
						kind
						status
						display_name
						auth_epoch
						perm_revision
					}
					errors { type message }
					was_successful
				}
			}`, inputType, op)
			resp, gql := admin.doGraphQL(q, map[string]any{"input": extraFields}, nil)
			if resp.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
				t.Fatalf("%s failed: status=%d errors=%+v", op, resp.StatusCode, gql.Errors)
			}
			var out struct {
				Res struct {
					User struct {
						ID           string `json:"id"`
						RawID        string `json:"_id"`
						Kind         string `json:"kind"`
						Status       string `json:"status"`
						DisplayName  string `json:"display_name"`
						AuthEpoch    int    `json:"auth_epoch"`
						PermRevision int    `json:"perm_revision"`
					} `json:"user"`
					WasSuccessful bool `json:"was_successful"`
				} `json:"res"`
			}
			if err := json.Unmarshal(gql.Data, &out); err != nil {
				t.Fatalf("decode %s: %v", op, err)
			}
			if !out.Res.WasSuccessful || out.Res.User.ID == "" || !strings.HasPrefix(out.Res.User.ID, "gid://integrione/user/") || out.Res.User.RawID == "" || out.Res.User.Kind != wantKind || out.Res.User.Status != "PENDING" {
				t.Fatalf("%s unexpected payload: %+v", op, out.Res)
			}
			createdUserIDs = append(createdUserIDs, out.Res.User.RawID)
			// Aktifkan user dan verifikasi identitas agar dapat digunakan pada skenario login & lifecycle lanjutan.
			_, actGql := admin.doGraphQL(fmt.Sprintf(`mutation { activateUser(user_id: "%s") { user { id status } errors { type message } was_successful } }`, out.Res.User.ID), nil, nil)
			if len(actGql.Errors) > 0 {
				t.Fatalf("activateUser after %s failed: %+v", op, actGql.Errors)
			}
			_ = psqlScalar(t, cfg, fmt.Sprintf("UPDATE user_identities SET is_verified = true, verified_at = NOW() WHERE user_id = '%s'; SELECT 'ok';", out.Res.User.RawID))
			return out.Res.User.ID
		}

		customerID = createCase("createUser", "CreateUserInput", map[string]any{
			"full_name": "E2E Customer",
			"email":     custEmail,
			"username":  "e2e_cust_" + suffix,
			"password":  commonPass,
		}, "CUSTOMER")

		employeeID = createCase("createEmployeeUser", "CreateEmployeeUserInput", map[string]any{
			"full_name": "E2E Employee",
			"email":     empEmail,
			"username":  empUsername,
			"password":  commonPass,
		}, "EMPLOYEE")

		vendorPartyID := createFixturePartyForRole(t, admin, cfg, "VENDOR")
		supplierPartyID := createFixturePartyForRole(t, admin, cfg, "SUPPLIER")

		vendorID = createCase("createVendorUser", "CreateVendorUserInput", map[string]any{
			"party_id":  vendorPartyID,
			"full_name": "E2E Vendor",
			"email":     vendEmail,
			"username":  "e2e_vend_" + suffix,
			"password":  commonPass,
		}, "VENDOR")

		supplierID = createCase("createSupplierUser", "CreateSupplierUserInput", map[string]any{
			"party_id":  supplierPartyID,
			"full_name": "E2E Supplier",
			"email":     supEmail,
			"username":  "e2e_sup_" + suffix,
			"password":  commonPass,
		}, "SUPPLIER")

		// Verifikasi operasi baca pengguna, identitas, siklus status, dan RBAC role/group dalam format Relay.
		_, listGql := admin.doGraphQL(`query { users(first: 20) { edges { cursor node { id _id kind status display_name } } nodes { id } pageInfo { startCursor endCursor hasPreviousPage hasNextPage } total_count } }`, nil, nil)
		if len(listGql.Errors) > 0 {
			t.Fatalf("users query failed: %+v", listGql.Errors)
		}

		_, listIDGql := admin.doGraphQL(`query($user_id: ID!) { userIdentities(user_id: $user_id, first: 10) { edges { cursor node { id kind value is_verified is_primary } } nodes { id kind value is_verified is_primary } pageInfo { startCursor endCursor hasPreviousPage hasNextPage } total_count } }`, map[string]any{
			"user_id": employeeID,
		}, nil)
		if len(listIDGql.Errors) > 0 {
			t.Fatalf("userIdentities query failed: %+v", listIDGql.Errors)
		}

		for _, mut := range []string{
			fmt.Sprintf(`mutation { suspendUser(user_id: "%s") { user { id status } errors { type message } was_successful } }`, vendorID),
			fmt.Sprintf(`mutation { activateUser(user_id: "%s") { user { id status } errors { type message } was_successful } }`, vendorID),
			fmt.Sprintf(`mutation { disableUser(user_id: "%s") { user { id status } errors { type message } was_successful } }`, supplierID),
			fmt.Sprintf(`mutation { softDeleteUser(user_id: "%s") { deleted errors { type message } was_successful } }`, vendorID),
		} {
			_, g := admin.doGraphQL(mut, nil, nil)
			if len(g.Errors) > 0 {
				t.Fatalf("lifecycle mutation %s failed: %+v", mut, g.Errors)
			}
		}

		// Buat role & group khusus lalu berikan ke employee via group.
		_, roleGql := admin.doGraphQL(`mutation($input: CreateRoleInput!) { createRole(input: $input) { role { id code name management_domain default_redirect_path } errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{
				"code":                  "e2e_role_" + suffix,
				"name":                  "E2E Role",
				"management_domain":     "iam",
				"default_redirect_path": "/identity",
				"redirect_priority":     40,
			},
		}, nil)
		if len(roleGql.Errors) > 0 {
			t.Fatalf("createRole failed: %+v", roleGql.Errors)
		}
		var roleOut struct {
			CreateRole struct {
				Role struct {
					ID string `json:"id"`
				} `json:"role"`
				WasSuccessful bool `json:"was_successful"`
			} `json:"createRole"`
		}
		_ = json.Unmarshal(roleGql.Data, &roleOut)
		if !roleOut.CreateRole.WasSuccessful || roleOut.CreateRole.Role.ID == "" {
			t.Fatalf("createRole unsuccessful: %+v", roleOut.CreateRole)
		}
		customRoleID = roleOut.CreateRole.Role.ID
		createdRoleIDs = append(createdRoleIDs, customRoleID)

		_, permsGql := admin.doGraphQL(`query { permissions(first: 50) { nodes { id code is_sensitive } pageInfo { endCursor hasNextPage } total_count } }`, nil, nil)
		if len(permsGql.Errors) > 0 {
			t.Fatalf("permissions query failed: %+v", permsGql.Errors)
		}
		var permsOut struct {
			Permissions struct {
				Nodes []struct {
					ID   string `json:"id"`
					Code string `json:"code"`
				} `json:"nodes"`
			} `json:"permissions"`
		}
		_ = json.Unmarshal(permsGql.Data, &permsOut)
		if len(permsOut.Permissions.Nodes) == 0 {
			t.Fatal("expected seeded IAM permissions")
		}
		var usersReadPermID string
		for _, p := range permsOut.Permissions.Nodes {
			if p.Code == "iam.users.read" {
				usersReadPermID = p.ID
				break
			}
		}
		if usersReadPermID == "" {
			t.Fatal("missing iam.users.read permission")
		}

		_, rpGql := admin.doGraphQL(`mutation($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { role_permission_edge { role_id node { id code } } errors { type message } was_successful } }`, map[string]any{
			"r": customRoleID,
			"p": usersReadPermID,
		}, nil)
		if len(rpGql.Errors) > 0 {
			t.Fatalf("assignRolePermission failed: %+v", rpGql.Errors)
		}

		_, grpGql := admin.doGraphQL(`mutation($input: CreateGroupInput!) { createGroup(input: $input) { group { id code name } errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{
				"code": "e2e_grp_" + suffix,
				"name": "E2E Group",
			},
		}, nil)
		if len(grpGql.Errors) > 0 {
			t.Fatalf("createGroup failed: %+v", grpGql.Errors)
		}
		var grpOut struct {
			CreateGroup struct {
				Group struct {
					ID string `json:"id"`
				} `json:"group"`
				WasSuccessful bool `json:"was_successful"`
			} `json:"createGroup"`
		}
		_ = json.Unmarshal(grpGql.Data, &grpOut)
		if !grpOut.CreateGroup.WasSuccessful || grpOut.CreateGroup.Group.ID == "" {
			t.Fatalf("createGroup unsuccessful: %+v", grpOut.CreateGroup)
		}
		customGroupID = grpOut.CreateGroup.Group.ID
		createdGroupIDs = append(createdGroupIDs, customGroupID)

		_, gmGql := admin.doGraphQL(`mutation($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_member_edge { group_id node { id } } errors { type message } was_successful } }`, map[string]any{
			"g": customGroupID,
			"u": employeeID,
		}, nil)
		if len(gmGql.Errors) > 0 {
			t.Fatalf("addGroupMember failed: %+v", gmGql.Errors)
		}

		_, grGql := admin.doGraphQL(`mutation($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { group_role_edge { assignment_id group_id node { id } } errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{
				"group_id":   customGroupID,
				"role_id":    customRoleID,
				"scope_type": "GLOBAL",
			},
		}, nil)
		if len(grGql.Errors) > 0 {
			t.Fatalf("assignGroupRole failed: %+v", grGql.Errors)
		}
	})

	t.Run("NegativeAuthorizationTestABAndZeroDBMutation", func(t *testing.T) {
		custClient := newHTTPSBrowserClient(t, cfg)
		custClient.bootstrapSession()
		st, _, loc := custClient.login(custEmail, commonPass)
		if st != "AUTHENTICATED" || loc != "/" {
			t.Fatalf("expected customer AUTHENTICATED redirect '/', got status=%s loc=%s", st, loc)
		}

		usersBefore := psqlScalar(t, cfg, "SELECT count(*) FROM users;")
		rolesBefore := psqlScalar(t, cfg, "SELECT count(*) FROM roles;")
		groupsBefore := psqlScalar(t, cfg, "SELECT count(*) FROM groups;")

		// TestAB: Customer yang terautentikasi dilarang mengakses operasi IAM admin apa pun.
		_, usersForbGql := custClient.doGraphQL(`query { users(first: 10) { edges { node { id } } pageInfo { hasNextPage } total_count } }`, nil, nil)
		requireErrorCode(t, usersForbGql, "FORBIDDEN")

		forbiddenOps := []struct {
			name  string
			query string
			vars  map[string]any
		}{
			{"createEmployeeUser", `mutation($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { user { id } errors { type message } was_successful } }`, map[string]any{
				"input": map[string]any{"worker_id": "55555555-5555-4555-8555-555555555555", "provisioning_request_id": "66666666-6666-4666-8666-666666666666", "full_name": "Evil Emp", "email": "evil-emp@integrione.local", "username": "evil_emp_" + suffix, "password": commonPass},
			}},
			{"createRole", `mutation($input: CreateRoleInput!) { createRole(input: $input) { role { id } errors { type message } was_successful } }`, map[string]any{
				"input": map[string]any{"code": "evil_role", "name": "Evil", "management_domain": "iam", "default_redirect_path": "/identity"},
			}},
			{"createGroup", `mutation($input: CreateGroupInput!) { createGroup(input: $input) { group { id } errors { type message } was_successful } }`, map[string]any{
				"input": map[string]any{"code": "evil_grp", "name": "Evil"},
			}},
		}
		for _, op := range forbiddenOps {
			_, gql := custClient.doGraphQL(op.query, op.vars, nil)
			requireMutationPayloadError(t, gql, op.name, "AUTHORIZATION")
		}

		usersAfter := psqlScalar(t, cfg, "SELECT count(*) FROM users;")
		rolesAfter := psqlScalar(t, cfg, "SELECT count(*) FROM roles;")
		groupsAfter := psqlScalar(t, cfg, "SELECT count(*) FROM groups;")
		if usersBefore != usersAfter || rolesBefore != rolesAfter || groupsBefore != groupsAfter {
			t.Fatalf("rejected operations mutated postgres rows: users=%s->%s roles=%s->%s groups=%s->%s",
				usersBefore, usersAfter, rolesBefore, rolesAfter, groupsBefore, groupsAfter)
		}
		_ = customerID
	})

	t.Run("EmployeeSelfServiceTOTPMFAAndStaleSensitiveDenial", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		empClient := newHTTPSBrowserClient(t, cfg)
		empClient.bootstrapSession()
		st, _, _ := empClient.login(empUsername, commonPass)
		if st != "AUTHENTICATED" {
			t.Fatalf("expected employee login AUTHENTICATED, got %s", st)
		}

		// Cek updateMyProfile, changeUsername, Query me + nested identities, serta Relay node()
		_, profGql := empClient.doGraphQL(`mutation($input: UpdateMyProfileInput!) { updateMyProfile(input: $input) { user { id _id display_name } errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{"display_name": "E2E Employee Updated", "locale": "id-ID", "timezone": "Asia/Jakarta"},
		}, nil)
		if len(profGql.Errors) > 0 {
			t.Fatalf("updateMyProfile failed: %+v", profGql.Errors)
		}
		_, meGql := empClient.doGraphQL(`query($id: ID!) { me { id _id identities(first: 5) { total_count nodes { id value } } } node(id: $id) { id ... on User { _id } } }`, map[string]any{
			"id": employeeID,
		}, nil)
		if len(meGql.Errors) > 0 {
			t.Fatalf("me/node query failed: %+v", meGql.Errors)
		}

		// Daftar TOTP -> konfirmasi dengan kode OTP sungguhan
		_, enrollGql := empClient.doGraphQL(`mutation { enrollMyTotp { factor_id otp_auth_uri errors { type message } was_successful } }`, nil, nil)
		if len(enrollGql.Errors) > 0 {
			t.Fatalf("enrollMyTotp failed: %+v", enrollGql.Errors)
		}
		var enrollOut struct {
			EnrollMyTotp struct {
				FactorID      string `json:"factor_id"`
				OtpAuthURI    string `json:"otp_auth_uri"`
				WasSuccessful bool   `json:"was_successful"`
			} `json:"enrollMyTotp"`
		}
		_ = json.Unmarshal(enrollGql.Data, &enrollOut)
		// Pastikan kita berada jauh dari ujung jendela 30 detik agar ketiga langkah (-30s, 0s, +30s) stabil dalam satu eksekusi cepat.
		if rem := 30 - (time.Now().Unix() % 30); rem <= 5 {
			time.Sleep(time.Duration(rem)*time.Second + 300*time.Millisecond)
		}
		totpBaseTime := time.Now()
		confirmCode := generateTOTPCode(t, enrollOut.EnrollMyTotp.OtpAuthURI, totpBaseTime.Add(-30*time.Second))

		_, confirmGql := empClient.doGraphQL(`mutation($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{"code": confirmCode},
		}, nil)
		if len(confirmGql.Errors) > 0 {
			t.Fatalf("confirmMyTotp failed: %+v", confirmGql.Errors)
		}
		var confirmOut struct {
			ConfirmMyTotp struct {
				Enabled       bool     `json:"enabled"`
				RecoveryCodes []string `json:"recovery_codes"`
				WasSuccessful bool     `json:"was_successful"`
			} `json:"confirmMyTotp"`
		}
		_ = json.Unmarshal(confirmGql.Data, &confirmOut)
		if !confirmOut.ConfirmMyTotp.WasSuccessful || !confirmOut.ConfirmMyTotp.Enabled || len(confirmOut.ConfirmMyTotp.RecoveryCodes) == 0 {
			t.Fatalf("unexpected confirmMyTotp output: %+v", confirmOut.ConfirmMyTotp)
		}

		// Logout lalu login ulang untuk menguji tantangan MFA (replay TOTP ditolak CONFLICT, recovery code berhasil).
		_, outGql := empClient.doGraphQL(`mutation { userLogout { logged_out errors { type message } was_successful } }`, nil, nil)
		if len(outGql.Errors) > 0 {
			t.Fatalf("userLogout failed: %+v", outGql.Errors)
		}
		_, unauthProfile := empClient.doGraphQL(`query { me { id } }`, nil, nil)
		requireErrorCode(t, unauthProfile, "UNAUTHENTICATED")

		empClient.bootstrapSession()
		st2, challenge, _ := empClient.login(empUsername, commonPass)
		if st2 != "MFA_REQUIRED" || challenge == "" {
			t.Fatalf("expected MFA_REQUIRED with challenge, got status=%s challenge=%q", st2, challenge)
		}
		_, replayMfaGql := empClient.doGraphQL(`mutation($input: ValidateUserMfaInput!) {
			validateUserMfa(input: $input) {
				csrfToken
				redirectLocation
				errors { type message }
				was_successful
			}
		}`, map[string]any{
			"input": map[string]any{
				"challenge": challenge,
				"totpCode":  confirmCode,
			},
		}, nil)
		requireMutationPayloadError(t, replayMfaGql, "validateUserMfa", "AUTHENTICATION")

		_, mfaGql := empClient.doGraphQL(`mutation($input: ValidateUserMfaInput!) {
			validateUserMfa(input: $input) {
				csrfToken
				redirectLocation
				errors { type message }
				was_successful
			}
		}`, map[string]any{
			"input": map[string]any{
				"challenge":    challenge,
				"recoveryCode": confirmOut.ConfirmMyTotp.RecoveryCodes[0],
			},
		}, nil)
		if len(mfaGql.Errors) > 0 {
			t.Fatalf("validateUserMfa with recoveryCode failed: %+v", mfaGql.Errors)
		}

		// Uji verifikasi ulang MFA sesi & penolakan re-autentikasi >10 menit secara nyata di Redis IAM.
		_, verifyMfaGql := empClient.doGraphQL(`mutation($code: String!) { verifyMySessionMfa(totp_code: $code) { verified errors { type message } was_successful } }`, map[string]any{
			"code": generateTOTPCode(t, enrollOut.EnrollMyTotp.OtpAuthURI, totpBaseTime),
		}, nil)
		if len(verifyMfaGql.Errors) > 0 {
			t.Fatalf("verifyMySessionMfa failed: %+v", verifyMfaGql.Errors)
		}

		// Simulasikan sesi berusia >10 menit di token Redis untuk memverifikasi penolakan AUTHORIZATION step-up/reauth.
		iamRedis := newE2ERedisClient(t, cfg, "iam_writer", cfg.iamRedisPass)
		defer iamRedis.Close()
		store, err := authsession.NewStore(iamRedis, authsession.NewKeyspace("development"))
		if err != nil {
			t.Fatalf("NewStore: %v", err)
		}
		empTokHash := authsession.HashOpaqueToken(empClient.sessionCookie().Value)
		sctx, err := store.ReadSessionContext(context.Background(), empTokHash, time.Now())
		if err != nil || sctx == nil || sctx.Token == nil || sctx.Activity == nil || sctx.Security == nil {
			t.Fatalf("ReadSessionContext failed: %v", err)
		}
		staleTime := time.Now().Add(-15 * time.Minute).Unix()
		sctx.Token.IssuedAtUnix = staleTime
		sctx.Token.MFAAuthenticated = false
		if err := store.WriteSession(context.Background(), *sctx.Token, *sctx.Activity, *sctx.Security, time.Now()); err != nil {
			t.Fatalf("WriteSession stale failed: %v", err)
		}
		_ = psqlScalar(t, cfg, fmt.Sprintf("UPDATE sessions SET authenticated_at = NOW() - INTERVAL '15 minutes', mfa_verified = false, mfa_verified_at = NULL WHERE id = '%s';", sctx.Token.SessionID))
		_, staleSensitiveErr := empClient.doGraphQL(`mutation { disableMyTotp { disabled errors { type message } was_successful } }`, nil, nil)
		_, staleMsg := requireMutationPayloadError(t, staleSensitiveErr, "disableMyTotp", "AUTHORIZATION")
		if msg := strings.ToLower(staleMsg); !strings.Contains(msg, "step-up") && !strings.Contains(msg, "reauth") {
			t.Fatalf("expected step-up/reauth message on AUTHORIZATION sensitive rejection, got %q", staleMsg)
		}

		// Segarkan kembali MFA sesi via verifyMySessionMfa (+30s skew) lalu uji disableMyTotp berhasil.
		_, reauthGql := empClient.doGraphQL(`mutation($code: String!) { verifyMySessionMfa(totp_code: $code) { verified errors { type message } was_successful } }`, map[string]any{
			"code": generateTOTPCode(t, enrollOut.EnrollMyTotp.OtpAuthURI, totpBaseTime.Add(30*time.Second)),
		}, nil)
		if len(reauthGql.Errors) > 0 {
			t.Fatalf("re-verify verifyMySessionMfa failed: %+v", reauthGql.Errors)
		}
		_, disableTotpGql := empClient.doGraphQL(`mutation { disableMyTotp { disabled errors { type message } was_successful } }`, nil, nil)
		if len(disableTotpGql.Errors) > 0 {
			t.Fatalf("disableMyTotp after step-up failed: %+v", disableTotpGql.Errors)
		}

		// Uji pencabutan sesi oleh Super Admin langsung menggagalkan sesi employee.
		_, revGql := admin.doGraphQL(`mutation($u: ID!) { revokeAllUserSessions(user_id: $u, reason: ADMIN_REVOKED) { revoked_count errors { type message } was_successful } }`, map[string]any{
			"u": employeeID,
		}, nil)
		if len(revGql.Errors) > 0 {
			t.Fatalf("revokeAllUserSessions failed: %+v", revGql.Errors)
		}
		_, afterRevokeErr := empClient.doGraphQL(`query { me { id } }`, nil, nil)
		requireErrorCode(t, afterRevokeErr, "UNAUTHENTICATED")
	})

	t.Run("RegistrationVerificationPasswordResetAndRateLimit", func(t *testing.T) {
		resetGatewayRateLimits(t, cfg)
		pub := newHTTPSBrowserClient(t, cfg)
		pub.bootstrapSession()
		regEmail := fmt.Sprintf("e2e-reg-%s@integrione.local", suffix)

		_, regGql := pub.doGraphQL(`mutation($input: CreateUserInput!) {
			createUser(input: $input) { user { id _id kind status display_name } redirectLocation errors { type message } was_successful }
		}`, map[string]any{
			"input": map[string]any{
				"full_name": "Registered Customer",
				"email":     regEmail,
				"username":  strings.ToLower("e2e_reg_" + suffix),
				"password":  commonPass,
			},
		}, nil)
		if len(regGql.Errors) > 0 {
			t.Fatalf("createUser failed: %+v", regGql.Errors)
		}
		var regOut struct {
			CreateUser struct {
				User          any  `json:"user"`
				WasSuccessful bool `json:"was_successful"`
			} `json:"createUser"`
		}
		if err := json.Unmarshal(regGql.Data, &regOut); err != nil {
			t.Fatalf("unmarshal createUser data: %v", err)
		}
		if !regOut.CreateUser.WasSuccessful || regOut.CreateUser.User != nil {
			t.Fatalf("unexpected public createUser output: %+v", regOut.CreateUser)
		}
		regUID := psqlScalar(t, cfg, fmt.Sprintf("SELECT user_id FROM user_identities WHERE value = '%s' LIMIT 1;", regEmail))
		if regUID == "" {
			t.Fatalf("expected user_identities row for %s", regEmail)
		}
		createdUserIDs = append(createdUserIDs, regUID)

		_, resendGql := pub.doGraphQL(`mutation($input: ResendValidationEmailInput) { resendValidationEmailToken(input: $input) { is_resend errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{"email": regEmail},
		}, nil)
		if len(resendGql.Errors) > 0 {
			t.Fatalf("resendValidationEmailToken failed: %+v", resendGql.Errors)
		}

		_, badVer := pub.doGraphQL(`mutation { verifyEmail(input: { token: "invalid-verification-token" }) { status verified errors { type message } was_successful } }`, nil, nil)
		if len(badVer.Errors) > 0 {
			t.Fatalf("verifyEmail invalid token unexpected error: %+v", badVer.Errors)
		}
		var verOut struct {
			VerifyEmail struct {
				Status        string `json:"status"`
				Verified      bool   `json:"verified"`
				WasSuccessful bool   `json:"was_successful"`
			} `json:"verifyEmail"`
		}
		_ = json.Unmarshal(badVer.Data, &verOut)
		if !verOut.VerifyEmail.WasSuccessful || verOut.VerifyEmail.Verified || verOut.VerifyEmail.Status != "INVALID" {
			t.Fatalf("expected was_successful=true, verified=false, status=INVALID, got %+v", verOut.VerifyEmail)
		}

		// Anti-enumeration reset kata sandi dan penolakan token reset tidak valid
		for _, targetEmail := range []string{regEmail, "nonexistent-" + suffix + "@integrione.local"} {
			_, rstReq := pub.doGraphQL(`mutation($input: UserResetPasswordInput!) { userResetPassword(input: $input) { message errors { type message } was_successful } }`, map[string]any{
				"input": map[string]any{"email": targetEmail},
			}, nil)
			if len(rstReq.Errors) > 0 {
				t.Fatalf("userResetPassword(%s) failed: %+v", targetEmail, rstReq.Errors)
			}
		}
		_, badReset := pub.doGraphQL(`mutation($input: IAMResetPasswordInput!) { resetPassword(input: $input) { reset errors { type message } was_successful } }`, map[string]any{
			"input": map[string]any{"token": "invalid-reset-token", "newPassword": "AnotherStrongPass#2026!"},
		}, nil)
		requireMutationPayloadError(t, badReset, "resetPassword")

		// Rate limit probe terhadap email dummy non-sistem agar tidak mengunci human superadmin.
		gwRLRedis := newE2ERedisClient(t, cfg, "gw_runtime", cfg.gwRedisPass)
		defer func() {
			ctxClean, cancelClean := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelClean()
			iter := gwRLRedis.Scan(ctxClean, 0, "ig:development:gw:rl:*", 100).Iterator()
			for iter.Next(ctxClean) {
				_ = gwRLRedis.Del(ctxClean, iter.Val()).Err()
			}
			_ = gwRLRedis.Close()
		}()

		probeEmail := fmt.Sprintf("rl-probe-%s@integrione.local", suffix)
		var rateLimited bool
		for i := 0; i < 25; i++ {
			_, rlGql := pub.doGraphQL(`mutation($input: UserLoginInput!) { userLogin(input: $input) { status errors { type message } was_successful } }`, map[string]any{
				"input": map[string]any{"identifier": probeEmail, "password": "WrongPassword#123"},
			}, nil)
			var rlOut struct {
				UserLogin struct {
					WasSuccessful bool `json:"was_successful"`
					Errors        []struct {
						Type string `json:"type"`
					} `json:"errors"`
				} `json:"userLogin"`
			}
			if len(rlGql.Data) > 0 && json.Unmarshal(rlGql.Data, &rlOut) == nil && !rlOut.UserLogin.WasSuccessful && len(rlOut.UserLogin.Errors) > 0 && rlOut.UserLogin.Errors[0].Type == "THROTTLE" {
				rateLimited = true
				break
			}
		}
		if !rateLimited {
			t.Fatal("expected THROTTLE error after rapid failed login attempts")
		}
	})

	t.Run("RedisACLEnforcementAndFailClosedRecovery", func(t *testing.T) {
		gwRedis := newE2ERedisClient(t, cfg, "gw_runtime", cfg.gwRedisPass)
		defer gwRedis.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// ACL gw_runtime dilarang menulis kunci internal IAM maupun mengubah security gate atau menjalankan FLUSHDB.
		if err := gwRedis.Set(ctx, "ig:development:iam:tok:forbidden-probe", "1", time.Minute).Err(); err == nil {
			t.Fatal("expected gw_runtime write to ig:development:iam:tok:* to be rejected by Redis ACL")
		}
		if err := gwRedis.Set(ctx, "ig:development:iam:ctrl:gate", "0", time.Minute).Err(); err == nil {
			t.Fatal("expected gw_runtime write to ig:development:iam:ctrl:gate to be rejected by Redis ACL")
		}
		if err := gwRedis.Do(ctx, "FLUSHDB").Err(); err == nil {
			t.Fatal("expected FLUSHDB to be disabled/rejected")
		}

		// Uji fail-closed saat IAM gRPC berhenti lalu pemulihan otomatis saat IAM menyala kembali.
		t.Cleanup(func() {
			_ = exec.Command("docker", "compose", "-f", filepath.Join(cfg.repoRoot, "docker-compose.yml"), "up", "-d", "--wait", "--no-build", "redis", "iam").Run()
		})
		composeCmd(t, cfg, "stop", "iam")
		_, downGql := admin.doGraphQL(`query { users(first: 5) { edges { node { id } } pageInfo { hasNextPage } total_count } }`, nil, nil)
		if len(downGql.Errors) == 0 {
			t.Fatal("expected error when IAM gRPC upstream is stopped")
		}
		downCode, _ := downGql.Errors[0].Extensions["code"].(string)
		if downCode != "UNAVAILABLE" && downCode != "INTERNAL" {
			t.Fatalf("expected UNAVAILABLE/INTERNAL when IAM down, got %q", downCode)
		}

		composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "iam")
		waitForGraphQLRecovery(t, admin, "IAM restart")

		// Uji fail-closed saat Redis berhenti lalu pemulihan otomatis Gateway & IAM setelah Redis menyala kembali.
		composeCmd(t, cfg, "stop", "redis")
		_, redisDownGql := admin.doGraphQL(`query { users(first: 5) { edges { node { id } } pageInfo { hasNextPage } total_count } }`, nil, nil)
		if len(redisDownGql.Errors) == 0 {
			t.Fatal("expected fail-closed error when Redis is stopped")
		}
		composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "redis")
		waitForGraphQLRecovery(t, admin, "Redis restart")
	})
}

// waitForGraphQLQueryRecovery memastikan kueri target berhasil beruntun melintasi seluruh replika Gateway pasca-pemulihan upstream.
func waitForGraphQLQueryRecovery(t *testing.T, client *browserClient, query, label string) {
	t.Helper()
	const requiredConsecutiveSuccess = 4
	consecutive := 0
	var lastErrs []graphQLError
	for i := 0; i < 24; i++ {
		_, gql := client.doGraphQL(query, nil, nil)
		if len(gql.Errors) == 0 {
			consecutive++
			if consecutive >= requiredConsecutiveSuccess {
				return
			}
			continue
		}
		consecutive = 0
		lastErrs = gql.Errors
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("expected automatic recovery after %s (%d consecutive successes), got errs=%+v", label, requiredConsecutiveSuccess, lastErrs)
}

// waitForGraphQLRecovery memastikan beberapa permintaan berurutan berhasil (melintasi seluruh replika Gateway di belakang Caddy) pasca-pemulihan upstream.
func waitForGraphQLRecovery(t *testing.T, client *browserClient, label string) {
	t.Helper()
	waitForGraphQLQueryRecovery(t, client, `query { users(first: 5) { edges { node { id } } pageInfo { hasNextPage } total_count } }`, "Gateway+IAM "+label)
}
