//go:build e2e

package e2e

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestHRMLeaveGuardsOverHTTPS(t *testing.T) {
	cfg := e2eRuntimeConfig{
		caCertPath: filepath.Join("..", "..", "..", "deploy", "certs", "ca.crt"),
		baseURL:    "https://erp.localhost:8443",
	}
	const id = `"10000000-0000-4000-8000-000000000001"`
	for _, tc := range []struct{ field, operation, wantCode, wantData string }{
		{"leavePolicy", `{ leavePolicy(id: ` + id + `) { id } }`, "UNAUTHENTICATED", `{"leavePolicy":null}`},
		{"leavePolicies", "{ leavePolicies { total_count } }", "UNAUTHENTICATED", "null"},
		{"leaveEntitlement", `{ leaveEntitlement(id: ` + id + `) { id } }`, "UNAUTHENTICATED", `{"leaveEntitlement":null}`},
		{"leaveEntitlements", `{ leaveEntitlements(worker_id: ` + id + `) { total_count } }`, "UNAUTHENTICATED", "null"},
		{"leaveLedgerEntries", `{ leaveLedgerEntries(entitlement_id: ` + id + `) { total_count } }`, "UNAUTHENTICATED", "null"},
		{"leaveAdjustments", `{ leaveAdjustments(entitlement_id: ` + id + `) { total_count } }`, "UNAUTHENTICATED", "null"},
		{"createLeaveAdjustment", `mutation { createLeaveAdjustment(input: {
			entitlement_id:` + id + `, quantity_delta:"-0.50", reason_code:"FIXTURE", reason:"Fixture", idempotency_key:"fixture"
		}) { was_successful errors { type } leave_adjustment { id } } }`, "FORBIDDEN", ""},
		{"approveLeaveAdjustment", `mutation { approveLeaveAdjustment(id: ` + id + `) { was_successful errors { type } leave_adjustment { id } } }`, "FORBIDDEN", ""},
		{"rejectLeaveAdjustment", `mutation { rejectLeaveAdjustment(id: ` + id + `, reason_code:"FIXTURE") { was_successful errors { type } leave_adjustment { id } } }`, "FORBIDDEN", ""},
		{"closeLeaveEntitlement", `mutation { closeLeaveEntitlement(id: ` + id + `) { was_successful errors { type } leave_entitlement { id } } }`, "FORBIDDEN", ""},
		{"postLeaveAdjustment", `mutation { postLeaveAdjustment(input:{
			entitlement_id:` + id + `, adjustment_id:` + id + `, quantity_delta:"-0.50", effective_date:"2026-11-25", reason_code:"FIXTURE"
		}) { was_successful errors { type } leave_entitlement { id } } }`, "FORBIDDEN", ""},
	} {
		t.Run(tc.field, func(t *testing.T) {
			client := newHTTPSBrowserClient(t, cfg)
			response, gql := client.doGraphQL(tc.operation, nil, nil)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP 200, got %d", response.StatusCode)
			}
			requireErrorCode(t, gql, tc.wantCode)
			if tc.wantData != "" && string(gql.Data) != tc.wantData {
				t.Fatalf("expected data %s, got %s", tc.wantData, gql.Data)
			}
			if tc.wantCode == "FORBIDDEN" && !strings.Contains(gql.Errors[0].Message, "missing X-CSRF-Token") {
				t.Fatalf("expected CSRF rejection, got %+v", gql.Errors[0])
			}
		})
	}
}
