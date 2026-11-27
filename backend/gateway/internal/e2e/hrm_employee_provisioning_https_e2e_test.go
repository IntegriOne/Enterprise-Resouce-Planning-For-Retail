//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
)

// TestHRMEmployeeProvisioningOverHTTPS membutuhkan pengajuan HRM dan sesi MFA dari stack uji.
func TestHRMEmployeeProvisioningOverHTTPS(t *testing.T) {
	baseURL := os.Getenv("E2E_HRM_PROVISIONING_URL")
	workerID := os.Getenv("E2E_HRM_PROVISIONING_WORKER")
	requestID := os.Getenv("E2E_HRM_PROVISIONING_REQUEST")
	session := os.Getenv("E2E_HRM_PROVISIONING_SESSION")
	if baseURL == "" || workerID == "" || requestID == "" || session == "" {
		t.Skip("requires disposable HRM provisioning fixtures and an authenticated operator")
	}
	cfg := loadE2ERuntimeConfig(t)
	cfg.baseURL = baseURL
	endpoint, err := url.Parse(baseURL)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
		t.Fatal("invalid HTTPS fixture URL")
	}
	client := newHTTPSBrowserClient(t, cfg)
	client.httpClient.Jar.SetCookies(endpoint, []*http.Cookie{
		{Name: "__Host-integrione_session", Value: session, Path: "/", Secure: true, HttpOnly: true},
	})
	response, gql := client.doGraphQL(`query { browserSession { authenticated csrfToken } }`, nil, nil)
	var browser struct {
		Session struct {
			Authenticated bool   `json:"authenticated"`
			CSRFToken     string `json:"csrfToken"`
		} `json:"browserSession"`
	}
	if response.StatusCode != http.StatusOK || len(gql.Errors) > 0 ||
		json.Unmarshal(gql.Data, &browser) != nil || !browser.Session.Authenticated || browser.Session.CSRFToken == "" {
		t.Fatal("expected an authenticated fixture session with a CSRF token")
	}
	client.csrfToken = browser.Session.CSRFToken

	tag := os.Getenv("E2E_HRM_PROVISIONING_USERNAME")
	password := os.Getenv("E2E_HRM_PROVISIONING_PASSWORD")
	if tag == "" && password == "" {
		tag, password = uniqueTag("e2eemp"), "StrongPass#2026!"
	} else if tag == "" || password == "" {
		t.Fatal("fixture username and password must be provided together")
	}
	query := fmt.Sprintf(`mutation ($input: CreateEmployeeUserInput!) {
		createEmployeeUser(input: $input) { %s }
	}`, gqlUserPayloadFields)
	vars := map[string]any{"input": map[string]any{
		"worker_id": workerID, "provisioning_request_id": requestID,
		"full_name": "Employee Fixture", "email": tag + "@integrione.test",
		"username": tag, "password": password,
	}}
	readDetail := func() map[string]any {
		t.Helper()
		return requireOpMap(t, doGQLOK(t, client, `query($id: ID!) {
			employeeProvisioningRequest(id: $id) { worker_id status }
		}`, map[string]any{"id": requestID}), "employeeProvisioningRequest")
	}
	if before := readDetail(); before["worker_id"] != workerID || before["status"] != "pending" {
		t.Fatalf("unexpected provisioning fixture: %+v", before)
	}

	create := func(t *testing.T, browser *browserClient) map[string]any {
		t.Helper()
		return requireMutationChildOK(t, doGQLOK(t, browser, query, vars), "createEmployeeUser", "user")
	}
	var accounts [2]map[string]any
	if !t.Run("concurrent requests", func(t *testing.T) {
		for i := range accounts {
			t.Run(fmt.Sprintf("browser %d", i+1), func(t *testing.T) {
				t.Parallel()
				other := newHTTPSBrowserClient(t, cfg)
				other.httpClient.Jar.SetCookies(endpoint, []*http.Cookie{
					{Name: "__Host-integrione_session", Value: session, Path: "/", Secure: true, HttpOnly: true},
				})
				other.csrfToken = browser.Session.CSRFToken
				accounts[i] = create(t, other)
			})
		}
	}) {
		t.FailNow()
	}
	account := accounts[0]
	if account["kind"] != "EMPLOYEE" || account["status"] != "PENDING" {
		t.Fatalf("unexpected provisioned account: %+v", account)
	}
	id := requireNonEmptyString(t, account, "id")
	if accounts[1]["id"] != id {
		t.Fatal("concurrent provisioning returned a different account")
	}
	if replay := create(t, client); replay["id"] != id {
		t.Fatal("repeated provisioning returned a different account")
	}
	if after := readDetail(); after["worker_id"] != workerID || after["status"] != "linked" {
		t.Fatalf("unexpected completed provisioning: %+v", after)
	}
}
