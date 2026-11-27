//go:build e2e

package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHRMPayrollQueriesRequireSessionOverHTTPS(t *testing.T) {
	cfg := e2eRuntimeConfig{
		caCertPath: filepath.Join("..", "..", "..", "deploy", "certs", "ca.crt"),
		baseURL:    "https://erp.localhost:8443",
	}
	const id = `"10000000-0000-4000-8000-000000000001"`
	for _, tc := range []struct{ name, query, nullableField string }{
		{"own payslip PDF", `{ myPayslipPdf(payroll_result_id: ` + id + `) { content_base64 } }`, "myPayslipPdf"},
		{"periods", "{ payrollPeriods { total_count } }", ""},
		{"groups", "{ payrollGroups { total_count } }", ""},
		{"rule sets", "{ payrollRuleSets { total_count } }", ""},
		{"run lookup", `{ payrollRuns(idempotency_key: "fixture-key") { total_count } }`, ""},
		{"period detail", `{ payrollPeriod(id: ` + id + `) { id } }`, "payrollPeriod"},
		{"group detail", `{ payrollGroup(id: ` + id + `) { id } }`, "payrollGroup"},
		{"rule set detail", `{ payrollRuleSet(id: ` + id + `) { id } }`, "payrollRuleSet"},
		{"period by group", `{ payrollPeriods(payroll_group_id: ` + id + `) { total_count } }`, ""},
		{"result detail", `{ payrollResult(id: ` + id + `) { net_pay_amount } }`, "payrollResult"},
		{"results by run", `{ payrollResults(payroll_run_id: ` + id + `) { total_count } }`, ""},
		{"result line detail", `{ payrollResultLine(id: ` + id + `) { amount } }`, "payrollResultLine"},
		{"result lines by result", `{ payrollResultLines(payroll_result_id: ` + id + `) { total_count } }`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newHTTPSBrowserClient(t, cfg)
			response, gql := client.doGraphQL(tc.query, nil, nil)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP 200, got %d", response.StatusCode)
			}
			requireErrorCode(t, gql, "UNAUTHENTICATED")
			wantData := "null"
			if tc.nullableField != "" {
				wantData = `{"` + tc.nullableField + `":null}`
			}
			if string(gql.Data) != wantData {
				t.Fatalf("expected %s, got %s", wantData, gql.Data)
			}
		})
	}
}

// TestHRMOwnPayslipPDFOverHTTPS memeriksa sebuah versi DRM dan penolakan akses lintas pekerja.
func TestHRMOwnPayslipPDFOverHTTPS(t *testing.T) {
	identifier := os.Getenv("E2E_HRM_PAYSLIP_EMPLOYEE_IDENTIFIER")
	password := os.Getenv("E2E_HRM_PAYSLIP_EMPLOYEE_PASSWORD")
	ownID := os.Getenv("E2E_HRM_OWN_PAYSLIP_RESULT_ID")
	otherID := os.Getenv("E2E_HRM_OTHER_PAYSLIP_RESULT_ID")
	digestHex := os.Getenv("E2E_HRM_OWN_PAYSLIP_PDF_SHA256")
	if identifier == "" || password == "" || ownID == "" || otherID == "" || digestHex == "" {
		t.Skip("requires a linked employee, two published results and the expected DRM PDF digest")
	}
	expectedDigest, digestErr := hex.DecodeString(digestHex)
	if strings.EqualFold(ownID, otherID) || digestErr != nil || len(expectedDigest) != sha256.Size {
		t.Fatal("fixture requires distinct payroll results and a valid SHA-256 digest")
	}

	cfg := e2eRuntimeConfig{
		caCertPath: filepath.Join("..", "..", "..", "deploy", "certs", "ca.crt"),
		baseURL:    "https://erp.localhost:8443",
	}
	client := newAnonBrowser(t, cfg)
	status, _, _ := client.login(identifier, password)
	if status != "AUTHENTICATED" {
		t.Fatalf("expected authenticated employee session, got %s", status)
	}
	t.Cleanup(func() { _, _, _ = logoutBrowserBestEffort(client) })

	const query = `query($id: ID!) { myPayslipPdf(payroll_result_id: $id) { content_base64 } }`
	response, gql := client.doGraphQL(query, map[string]any{"id": ownID}, nil)
	if response.StatusCode != http.StatusOK || len(gql.Errors) != 0 {
		t.Fatalf("own PDF query failed: HTTP %d, GraphQL errors=%d", response.StatusCode, len(gql.Errors))
	}
	if !strings.Contains(response.Header.Get("Cache-Control"), "no-store") {
		t.Fatal("private PDF response is cacheable")
	}
	var data struct {
		PDF *struct {
			Content []byte `json:"content_base64"`
		} `json:"myPayslipPdf"`
	}
	if err := json.Unmarshal(gql.Data, &data); err != nil || data.PDF == nil {
		t.Fatal("own PDF response is missing or malformed")
	}
	digest := sha256.Sum256(data.PDF.Content)
	if !bytes.HasPrefix(data.PDF.Content, []byte("%PDF-")) || !bytes.Equal(digest[:], expectedDigest) {
		t.Fatal("response does not match the expected DRM PDF version")
	}

	response, gql = client.doGraphQL(query, map[string]any{"id": otherID}, nil)
	if response.StatusCode != http.StatusOK || string(gql.Data) != `{"myPayslipPdf":null}` {
		t.Fatal("foreign payslip query returned unexpected data")
	}
	requireErrorCode(t, gql, "NOT_FOUND")
}
