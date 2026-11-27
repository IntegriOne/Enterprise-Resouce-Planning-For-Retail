//go:build e2e

// Package e2e menguji independensi startup, failover Caddy antar replika Gateway,
// serta isolasi kegagalan IAM dan Redis tanpa membuat gateway-1 maupun gateway-2 mati.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func assertNoSensitiveIAMOrRedisDetails(t *testing.T, msg string) {
	t.Helper()
	lower := strings.ToLower(msg)
	for _, forbidden := range []string{
		"iam:9001",
		"redis:6379",
		"127.0.0.1",
		"dial tcp",
		"connection refused",
		"postgres",
		"sqlstate",
		"pq:",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("client error message leaked internal detail (%q): %q", forbidden, msg)
		}
	}
}

func assertDirectReplicaHealthOK(t *testing.T, replicaBaseURL string) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	body, _ := json.Marshal(map[string]string{"query": "query { health }"})
	req, err := http.NewRequest(http.MethodPost, replicaBaseURL+"/api/graphql", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build direct replica request %s: %v", replicaBaseURL, err)
	}
	req.Host = "erp.localhost:8443"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://erp.localhost:8443")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("direct replica POST %s/api/graphql failed: %v", replicaBaseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read direct replica response %s: %v", replicaBaseURL, err)
	}
	var gql graphQLResponse
	if err := json.Unmarshal(raw, &gql); err != nil {
		t.Fatalf("decode direct replica response %s (status=%d body=%s): %v", replicaBaseURL, resp.StatusCode, string(raw), err)
	}
	if resp.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
		t.Fatalf("direct replica %s health error: status=%d errs=%+v", replicaBaseURL, resp.StatusCode, gql.Errors)
	}
	var out struct {
		Health string `json:"health"`
	}
	if err := json.Unmarshal(gql.Data, &out); err != nil || out.Health != "ok" {
		t.Fatalf("unexpected direct replica %s health payload: data=%s err=%v", replicaBaseURL, string(gql.Data), err)
	}
}

func waitUntilHTTPSGraphQLConsecutiveOK(t *testing.T, client *browserClient, query string, required int, label string) {
	t.Helper()
	consecutive := 0
	var lastErrs []graphQLError
	var lastStatus int
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		resp, gql := client.doGraphQL(query, nil, nil)
		lastStatus = resp.StatusCode
		if resp.StatusCode == http.StatusOK && len(gql.Errors) == 0 {
			consecutive++
			if consecutive >= required {
				return
			}
			continue
		}
		consecutive = 0
		lastErrs = gql.Errors
		time.Sleep(350 * time.Millisecond)
	}
	t.Fatalf("HTTPS GraphQL failed to stabilize after %s (%d consecutive OK needed; lastStatus=%d errs=%+v)", label, required, lastStatus, lastErrs)
}

func assertContainerRunning(t *testing.T, containerName string) {
	t.Helper()
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Status}}", containerName).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect container %s failed: %v (%s)", containerName, err, string(out))
	}
	got := strings.TrimSpace(string(out))
	if got != "running" {
		t.Fatalf("container %s state = %q, want running", containerName, got)
	}
}

func waitContainerHealthyBestEffort(ctx context.Context, containerName string) {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}:{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}", containerName).CombinedOutput()
		if err == nil {
			state := strings.TrimSpace(string(out))
			if state == "running:healthy" || state == "running:none" {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func waitUntilAnonBrowserBootstrapped(t *testing.T, client *browserClient, label string) {
	t.Helper()
	var lastErrs []graphQLError
	var lastStatus int
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, gql := client.doGraphQL(`query { browserSession { authenticated csrfToken } }`, nil, map[string]string{
			"X-CSRF-Token": "",
		})
		lastStatus = resp.StatusCode
		lastErrs = gql.Errors
		if resp.StatusCode == http.StatusOK && len(gql.Errors) == 0 {
			_ = client.bootstrapSession()
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("anon browserSession failed to recover after %s (lastStatus=%d errs=%+v)", label, lastStatus, lastErrs)
}

func waitUntilIAMLoginReachable(t *testing.T, cfg e2eRuntimeConfig, label string) {
	t.Helper()
	probeClient := newHTTPSBrowserClient(t, cfg)
	waitUntilAnonBrowserBootstrapped(t, probeClient, label+" bootstrap")
	deadline := time.Now().Add(15 * time.Second)
	var lastType, lastMsg string
	for time.Now().Before(deadline) {
		resp, gql := probeClient.doGraphQL(`mutation {
			userLogin(input: { identifier: "iam-recovery-probe@integrione.local", password: "ProbePassword#123" }) {
				status was_successful errors { type message }
			}
		}`, nil, nil)
		if resp.StatusCode == http.StatusOK && len(gql.Errors) == 0 {
			var out struct {
				UserLogin struct {
					Errors []struct {
						Type    string `json:"type"`
						Message string `json:"message"`
					} `json:"errors"`
				} `json:"userLogin"`
			}
			_ = json.Unmarshal(gql.Data, &out)
			if len(out.UserLogin.Errors) > 0 {
				lastType = out.UserLogin.Errors[0].Type
				lastMsg = out.UserLogin.Errors[0].Message
				if lastType != "UNAVAILABLE" {
					return
				}
			}
		}
		time.Sleep(350 * time.Millisecond)
	}
	t.Fatalf("IAM login failed to auto-recover after %s (lastType=%q msg=%q)", label, lastType, lastMsg)
}

func ensureFullStackHealthy(cfg e2eRuntimeConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	_ = exec.CommandContext(ctx, "docker", "compose", "-f", filepath.Join(cfg.repoRoot, "docker-compose.yml"),
		"up", "-d", "--wait", "--no-build", "postgres", "redis", "iam", "enterprise-foundation", "gateway-1", "gateway-2", "caddy",
	).Run()
	for _, c := range []string{
		"integrione-retail-redis-1",
		"integrione-retail-iam-1",
		"integrione-retail-gateway-1-1",
		"integrione-retail-gateway-2-1",
		"integrione-retail-caddy-1",
	} {
		waitContainerHealthyBestEffort(ctx, c)
	}
}

func TestGateway_BootsWithoutIAMAndRedisThenAutoRecovers(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	t.Cleanup(func() {
		ensureFullStackHealthy(cfg)
	})

	// 1. Hentikan IAM dan Redis sepenuhnya sebelum kedua replika Gateway dibuat ulang.
	composeCmd(t, cfg, "stop", "iam", "redis")

	// Tanpa --no-deps: membuktikan docker compose tidak lagi memblokir gateway-1/gateway-2 pada iam, redis, maupun antar replika.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "--force-recreate", "gateway-1", "gateway-2")
	assertContainerHealthy(t, "integrione-retail-gateway-1-1")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")

	// 2. Masing-masing replika merespons health "ok" secara independen, termasuk melalui Caddy HTTPS.
	assertDirectReplicaHealthOK(t, "http://127.0.0.1:5000")
	assertDirectReplicaHealthOK(t, "http://127.0.0.1:5001")

	client := newHTTPSBrowserClient(t, cfg)
	waitUntilHTTPSGraphQLConsecutiveOK(t, client, `query { health }`, 4, "HTTPS health while IAM+Redis stopped")

	// 3. Query & mutasi terproteksi fail-closed bersih dengan UNAVAILABLE pada payload error /api/graphql (HTTP 200), tanpa INTERNAL_ERROR atau kebocoran internal.
	for i := 0; i < 4; i++ {
		respBS, gqlBS := client.doGraphQL(`query { browserSession { authenticated csrfToken } }`, nil, map[string]string{
			"X-CSRF-Token": "",
		})
		if respBS.StatusCode != http.StatusOK {
			t.Fatalf("browserSession HTTP status = %d while Redis+IAM stopped, want 200 GraphQL error response", respBS.StatusCode)
		}
		requireErrorCode(t, gqlBS, "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, gqlBS.Errors[0].Message)

		respLogin, gqlLogin := client.doGraphQL(`mutation($in: UserLoginInput!) {
			userLogin(input: $in) { status was_successful errors { type message } }
		}`, map[string]any{
			"in": map[string]any{"identifier": cfg.superadminUsername, "password": cfg.superadminPassword},
		}, nil)
		if respLogin.StatusCode != http.StatusOK {
			t.Fatalf("userLogin HTTP status = %d while Redis stopped (fail-closed CSRF), want 200 GraphQL error response", respLogin.StatusCode)
		}
		requireErrorCode(t, gqlLogin, "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, gqlLogin.Errors[0].Message)
	}

	// 4. Nyalakan Redis saja (IAM masih mati): browserSession anonim langsung pulih otomatis tanpa restart Gateway, sementara userLogin fail-closed UNAVAILABLE.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "redis")
	waitContainerHealthyBestEffort(context.Background(), "integrione-retail-redis-1")

	waitUntilHTTPSGraphQLConsecutiveOK(t, client, `query { health }`, 2, "post-Redis start")
	waitUntilAnonBrowserBootstrapped(t, client, "post-Redis recovery")

	for i := 0; i < 4; i++ {
		_, errMsg := doMutationFailLabeled(t, "LOGIN_REDIS_UP_IAM_DOWN", client, `mutation($in: UserLoginInput!) {
			userLogin(input: $in) { status was_successful errors { type message } }
		}`, map[string]any{
			"in": map[string]any{"identifier": cfg.superadminUsername, "password": cfg.superadminPassword},
		}, "userLogin", "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, errMsg)
	}

	// 5. Nyalakan IAM: Gateway pulih otomatis tanpa restart replika maupun Caddy.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "iam")
	waitContainerHealthyBestEffort(context.Background(), "integrione-retail-iam-1")

	waitUntilIAMLoginReachable(t, cfg, "post-IAM boot start")
	resetGatewayRateLimits(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)
	waitForGraphQLRecovery(t, admin, "post-IAM-and-Redis-boot-recovery")
}

func TestGateway_ReplicaPeerAndCaddyFailoverIndependence(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	resetGatewayRateLimits(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	t.Cleanup(func() {
		ensureFullStackHealthy(cfg)
	})

	// 1. Hentikan gateway-1, lalu boot ulang gateway-2 + Caddy TANPA --no-deps untuk membuktikan independensi penuh dari gateway-1.
	composeCmd(t, cfg, "stop", "gateway-1")
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "--force-recreate", "gateway-2")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")
	assertDirectReplicaHealthOK(t, "http://127.0.0.1:5001")

	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "--force-recreate", "caddy")
	assertContainerRunning(t, "integrione-retail-caddy-1")

	// Caddy HTTPS tetap melayani health + operasi terautentikasi penuh lewat gateway-2 saja.
	waitUntilHTTPSGraphQLConsecutiveOK(t, admin, `query { health me { id } users(first: 2) { total_count } }`, 6, "Caddy failover to gateway-2 alone")

	// 2. Pulihkan gateway-1 lalu hentikan gateway-2: gateway-1 + Caddy melayani penuh secara independen.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "gateway-1")
	assertContainerHealthy(t, "integrione-retail-gateway-1-1")
	assertDirectReplicaHealthOK(t, "http://127.0.0.1:5000")

	composeCmd(t, cfg, "stop", "gateway-2")
	waitUntilHTTPSGraphQLConsecutiveOK(t, admin, `query { health me { id } users(first: 2) { total_count } }`, 6, "Caddy failover to gateway-1 alone")

	// 3. Kembalikan kedua replika ke kondisi sehat.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "gateway-2")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")
	waitForGraphQLRecovery(t, admin, "both-replicas-restored")
}

func TestGateway_RuntimeIAMOutageFaultIsolationAndAutoRecovery(t *testing.T) {
	cfg := loadE2ERuntimeConfig(t)
	resetGatewayRateLimits(t, cfg)
	seedEFPermissionCatalog(t, cfg)
	admin := newSuperAdminBrowser(t, cfg)

	// Siapkan akun uji sebelum memutus IAM (t.Cleanup milik fixture akan berjalan sesudah pemulihan container).
	_, efReader := provisionEFUserWithPerms(t, admin, cfg, "dec_ef_read", []string{
		"enterprise_foundation.addresses.read",
	}, "GLOBAL", nil)
	staleStaff, staleClient := provisionEFUserWithPerms(t, admin, cfg, "dec_iam_stale", []string{
		"iam.roles.write",
	}, "GLOBAL", nil)

	// Pastikan container dipulihkan LEBIH DULU pada LIFO cleanup sebelum pembersihan fixture PostgreSQL.
	t.Cleanup(func() {
		ensureFullStackHealthy(cfg)
	})

	// Segarkan snapshot EF dan snapshot staleClient di Redis (TTL <=30s) serta mundurkan timestamp sesi staleClient >10m sebelum IAM berhenti.
	_ = doGQLOK(t, efReader, `query { allAddresses(first: 1) { total_count } }`, nil)
	_ = doGQLOK(t, staleClient, `query { myWorkspaceAccess { canAccessIdentity defaultLocation } }`, nil)
	backdateSessionForStaleSensitiveTest(t, cfg, staleClient, staleStaff.ID)

	anonClient := newAnonBrowser(t, cfg)
	_ = anonClient.bootstrapSession()

	// Hentikan container IAM saat runtime -> kedua replika Gateway dan layanan EF tetap sehat.
	composeCmd(t, cfg, "stop", "iam")
	assertContainerHealthy(t, "integrione-retail-gateway-1-1")
	assertContainerHealthy(t, "integrione-retail-gateway-2-1")
	assertContainerHealthy(t, "integrione-retail-redis-1")

	// 6 request beruntun mencakup fase kegagalan awal serta fase fast-fail circuit breaker (threshold=5).
	for i := 0; i < 6; i++ {
		// a. Gateway health & query Enterprise Foundation (dengan snapshot segar Redis) tetap berjalan!
		efRoot := doGQLOK(t, efReader, `query { health allAddresses(first: 1) { total_count } }`, nil)
		if efRoot["health"] != "ok" {
			t.Fatalf("Gateway health != ok during IAM outage: %+v", efRoot)
		}

		// b. Query ke IAM fail-closed ketat dengan UNAVAILABLE (tidak pernah INTERNAL) dan tanpa kebocoran rincian internal.
		usersErr := doGQLErrorLabeled(t, "IAM_OUTAGE_USERS", admin, `query { users(first: 2) { total_count } }`, nil, "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, usersErr.Errors[0].Message)

		meErr := doGQLErrorLabeled(t, "IAM_OUTAGE_ME", admin, `query { me { id } }`, nil, "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, meErr.Errors[0].Message)

		// c. Mutasi login mengembalikan UNAVAILABLE bersih pada MutationResult.
		_, loginMsg := doMutationFailLabeled(t, "IAM_OUTAGE_LOGIN", anonClient, `mutation($in: UserLoginInput!) {
			userLogin(input: $in) { status was_successful errors { type message } }
		}`, map[string]any{
			"in": map[string]any{"identifier": cfg.superadminUsername, "password": cfg.superadminPassword},
		}, "userLogin", "UNAVAILABLE")
		assertNoSensitiveIAMOrRedisDetails(t, loginMsg)

		// d. Sensitive reauth check yang memanggil IAM saat IAM mati mengembalikan UNAVAILABLE (tidak dimask menjadi AUTHORIZATION).
		roleErrType, roleMsg := doMutationFailLabeled(t, "IAM_OUTAGE_SENSITIVE_REAUTH", staleClient, fmt.Sprintf(`mutation {
			createRole(input: {
				code: "outage_role_%d",
				name: "Outage Role",
				management_domain: "iam",
				default_redirect_path: "/identity"
			}) { %s }
		}`, i, gqlRolePayloadFields), nil, "createRole", "UNAVAILABLE")
		if roleErrType != "UNAVAILABLE" {
			t.Fatalf("expected UNAVAILABLE from stale sensitive reauth during IAM outage, got %q", roleErrType)
		}
		assertNoSensitiveIAMOrRedisDetails(t, roleMsg)

		// e. Cookie sesi browser admin tidak terhapus akibat outage IAM.
		if ck := admin.sessionCookie(); ck == nil || ck.Value == "" {
			t.Fatal("expected admin session cookie to remain intact during IAM outage")
		}
	}

	// Nyalakan ulang IAM -> kedua replika Gateway pulih otomatis tanpa restart kontainer Gateway.
	composeCmd(t, cfg, "up", "-d", "--wait", "--no-build", "iam")
	waitContainerHealthyBestEffort(context.Background(), "integrione-retail-iam-1")

	waitUntilIAMLoginReachable(t, cfg, "post-runtime-IAM-outage")
	resetGatewayRateLimits(t, cfg)
	waitForGraphQLRecovery(t, admin, "post-runtime-IAM-outage")
	meRecovered := doGQLOK(t, admin, `query { health me { id } }`, nil)
	_ = requireOpMap(t, meRecovered, "me")
}
