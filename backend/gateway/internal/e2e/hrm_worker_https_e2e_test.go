//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHRMWorkerLegalEntityGrantOverHTTPS(t *testing.T) {
	identifier := os.Getenv("E2E_HRM_EMPLOYEE_IDENTIFIER")
	password := os.Getenv("E2E_HRM_EMPLOYEE_PASSWORD")
	entityID := os.Getenv("E2E_HRM_LEGAL_ENTITY_ID")
	allowedID := os.Getenv("E2E_HRM_ALLOWED_WORKER_ID")
	deniedID := os.Getenv("E2E_HRM_DENIED_WORKER_ID")
	deniedEntityID := os.Getenv("E2E_HRM_DENIED_LEGAL_ENTITY_ID")
	if identifier == "" || password == "" || entityID == "" || allowedID == "" || deniedID == "" || deniedEntityID == "" {
		t.Skip("requires a scoped IAM employee and two workers in different legal entities")
	}
	if allowedID == deniedID || entityID == deniedEntityID {
		t.Fatal("fixture needs distinct allowed and denied workers")
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

	type worker struct {
		ID            string `json:"_id"`
		LegalEntityID string `json:"primary_legal_entity_reference_id"`
	}

	const detail = `query($id: ID!) { worker(id: $id) { _id primary_legal_entity_reference_id } }`
	response, gql := client.doGraphQL(detail, map[string]any{"id": allowedID}, nil)
	if response.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
		t.Fatalf("allowed worker query failed: HTTP %d, errors=%+v", response.StatusCode, gql.Errors)
	}
	var details struct {
		Worker *worker `json:"worker"`
	}
	if err := json.Unmarshal(gql.Data, &details); err != nil {
		t.Fatal(err)
	}
	if details.Worker == nil || details.Worker.ID != allowedID || details.Worker.LegalEntityID != entityID {
		t.Fatalf("allowed worker returned unexpected identity: %+v", details.Worker)
	}

	_, gql = client.doGraphQL(detail, map[string]any{"id": deniedID}, nil)
	requireErrorCode(t, gql, "NOT_FOUND")

	response, gql = client.doGraphQL("{ workers(first: 20) { nodes { _id primary_legal_entity_reference_id } total_count } }", nil, nil)
	if response.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
		t.Fatalf("scoped list failed: HTTP %d, errors=%+v", response.StatusCode, gql.Errors)
	}
	var list struct {
		Workers struct {
			Nodes []worker
			Total int `json:"total_count"`
		} `json:"workers"`
	}
	if err := json.Unmarshal(gql.Data, &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Workers.Nodes) == 0 || list.Workers.Total < len(list.Workers.Nodes) {
		t.Fatalf("unexpected scoped page size=%d total=%d", len(list.Workers.Nodes), list.Workers.Total)
	}
	for _, item := range list.Workers.Nodes {
		if item.ID == deniedID || item.LegalEntityID != entityID {
			t.Fatal("worker from outside the allowed legal entity appeared in the list")
		}
	}
	const filtered = `query($entity: ID!) {
		workers(legal_entity_id: $entity, first: 20) { nodes { _id primary_legal_entity_reference_id } total_count }
	}`
	for _, tc := range []struct {
		name, entity string
		allowed      bool
	}{
		{"within grant", entityID, true},
		{"outside grant", deniedEntityID, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response, gql := client.doGraphQL(filtered, map[string]any{"entity": tc.entity}, nil)
			if response.StatusCode != http.StatusOK || len(gql.Errors) > 0 {
				t.Fatalf("filtered list failed: HTTP %d, errors=%+v", response.StatusCode, gql.Errors)
			}
			var page struct {
				Workers struct {
					Nodes []worker
					Total int `json:"total_count"`
				} `json:"workers"`
			}
			if err := json.Unmarshal(gql.Data, &page); err != nil {
				t.Fatal(err)
			}
			items, total := page.Workers.Nodes, page.Workers.Total
			if !tc.allowed {
				if len(items) != 0 || total != 0 {
					t.Fatalf("out-of-scope filter returned %d items and total %d", len(items), total)
				}
				return
			}
			if len(items) == 0 || total < len(items) {
				t.Fatalf("unexpected filtered page size=%d total=%d", len(items), total)
			}
			for _, item := range items {
				if item.ID == deniedID || item.LegalEntityID != tc.entity {
					t.Fatal("filtered list included a worker from outside the granted legal entity")
				}
			}
		})
	}

}

func TestHRMWorkerFiltersRequireSessionOverHTTPS(t *testing.T) {
	cfg := e2eRuntimeConfig{
		caCertPath: filepath.Join("..", "..", "..", "deploy", "certs", "ca.crt"),
		baseURL:    "https://erp.localhost:8443",
	}
	for _, tc := range []struct{ name, query string }{
		{"worker number", `{ workers(worker_number: "fixture-worker") { total_count } }`},
		{"legal entity", `{ workers(legal_entity_id: "20000000-0000-4000-8000-000000000001") { total_count } }`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newHTTPSBrowserClient(t, cfg)
			response, gql := client.doGraphQL(tc.query, nil, nil)
			if response.StatusCode != http.StatusOK || string(gql.Data) != "null" {
				t.Fatalf("expected HTTP 200 and null data, got status %d and data %s", response.StatusCode, gql.Data)
			}
			requireErrorCode(t, gql, "UNAUTHENTICATED")
		})
	}
}

func TestHRMWorkerAndPositionGuardsOverHTTPS(t *testing.T) {
	cfg := e2eRuntimeConfig{
		caCertPath: filepath.Join("..", "..", "..", "deploy", "certs", "ca.crt"),
		baseURL:    "https://erp.localhost:8443",
	}
	const id = `"10000000-0000-4000-8000-000000000001"`
	for _, tc := range []struct {
		field, operation, code string
	}{
		{"employmentContract", `{ employmentContract(id:` + id + `) { id } }`, "UNAUTHENTICATED"},
		{"createEmploymentContract", `mutation { createEmploymentContract(input:{
			worker_id:` + id + `, contract_number:"FIXTURE", contract_type:"fixed_term",
			employing_legal_entity_reference_id:` + id + `, effective_from:"2026-12-01",
			work_country_code:"ID", governing_jurisdiction_code:"ID"
		}) { was_successful errors { type } employment_contract { id } } }`, "FORBIDDEN"},
		{"recordSignedEmploymentContract", `mutation { recordSignedEmploymentContract(id:` + id + `, document_version_id:` + id + `) { was_successful errors { type } employment_contract { id } } }`, "FORBIDDEN"},
		{"deleteEmploymentContract", `mutation { deleteEmploymentContract(id:` + id + `) { was_successful errors { type } deleted } }`, "FORBIDDEN"},
		{"employmentContracts", `{ employmentContracts(worker_id:` + id + `) { was_successful errors { type } deleted { total_count } } }`, "UNAUTHENTICATED"},
		{"workerAssignment", `{ workerAssignment(id:` + id + `) { id } }`, "UNAUTHENTICATED"},
		{"workerAssignments", `{ workerAssignments(worker_id:` + id + `) { total_count } }`, "UNAUTHENTICATED"},
		{"workerAssignments by position", `{ workerAssignments(position_id:` + id + `) { total_count } }`, "UNAUTHENTICATED"},
		{"approveWorkerAssignment", `mutation { approveWorkerAssignment(id:` + id + `) { was_successful errors { type } worker_assignment { id } } }`, "FORBIDDEN"},
		{"applyPrimaryAssignment", `mutation { applyPrimaryAssignment(id:` + id + `, assignment_id:` + id + `) { was_successful errors { type } worker { id } } }`, "FORBIDDEN"},
		{"decideWorkerAssignmentTransfer", `mutation { decideWorkerAssignmentTransfer(id:` + id + `, decision:REJECT) { was_successful errors { type } worker_assignment_transfer { id } } }`, "FORBIDDEN"},
		{"activatePosition", `mutation { activatePosition(id:` + id + `) { was_successful errors { type } position { id } } }`, "FORBIDDEN"},
		{"updatePosition", `mutation { updatePosition(id:` + id + `, input:{position_title:"Fixture"}) { was_successful errors { type } position { id } } }`, "FORBIDDEN"},
		{"deletePosition", `mutation { deletePosition(id:` + id + `) { was_successful errors { type } deleted } }`, "FORBIDDEN"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			client := newHTTPSBrowserClient(t, cfg)
			response, gql := client.doGraphQL(tc.operation, nil, nil)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("expected HTTP 200, got %d", response.StatusCode)
			}
			requireErrorCode(t, gql, tc.code)
			if tc.code == "FORBIDDEN" && !strings.Contains(gql.Errors[0].Message, "missing X-CSRF-Token") {
				t.Fatalf("expected CSRF rejection, got %+v", gql.Errors[0])
			}
		})
	}
}
