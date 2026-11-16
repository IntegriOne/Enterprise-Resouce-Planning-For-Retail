package security_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/metadata"
)

func TestPermissionCatalogCoversExplicitOperations(t *testing.T) {
	t.Parallel()
	catalog := security.PermissionCatalog()
	if len(catalog) != 408 {
		t.Fatalf("operation mapping count = %d, want 408", len(catalog))
	}
	codes := make(map[string]struct{}, len(catalog))
	for operation, code := range catalog {
		if !strings.HasPrefix(operation, "hr_") || !strings.HasPrefix(code, "human_resource.") {
			t.Fatalf("invalid catalog mapping: %q -> %q", operation, code)
		}
		codes[code] = struct{}{}
	}
	if len(codes) != 324 {
		t.Fatalf("permission count = %d, want 324", len(codes))
	}
	for operation, want := range map[string]string{
		"hr_workers.list":                        "human_resource.workers.read",
		"hr_workers.create":                      "human_resource.workers.create",
		"hr_workers.update":                      "human_resource.workers.update",
		"hr_workers.delete":                      "human_resource.workers.delete",
		"hr_leave_entitlements.post_leave_entry": "human_resource.leave_entitlements.post",
		"hr_leave_adjustments.approve":           "human_resource.leave_adjustments.approve",
		"hr_payroll_runs.approve":                "human_resource.payroll_runs.approve",
		"hr_career_paths.approve":                "human_resource.career_paths.approve",
		"hr_performance_cycles.launch":           "human_resource.performance_cycles.launch",
		"hr_talent_assessments.finalize":         "human_resource.talent_assessments.finalize",
		"hr_payroll_result_lines.read":           "human_resource.payroll_result_lines.read",
		"hr_application_stage_history.list":      "human_resource.application_stage_history.read",
		"hr_engagement_survey_results.read":      "human_resource.engagement_survey_results.read",
	} {
		if got := catalog[operation]; got != want {
			t.Errorf("mapping %q = %q, want %q", operation, got, want)
		}
	}
	for _, resource := range []string{
		"hr_worker_lifecycle_events", "hr_application_stage_history", "hr_payroll_results",
		"hr_payroll_result_lines", "hr_leave_ledger_entries", "hr_engagement_survey_results",
	} {
		for _, action := range []string{"create", "update", "delete"} {
			if _, exists := catalog[resource+"."+action]; exists {
				t.Errorf("read-only resource %q has %q permission", resource, action)
			}
		}
	}
}

func TestSensitiveOperationsRequireIndependentStepUp(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		resource  string
		action    string
		sensitive bool
	}{
		{"hr_workers", "delete", true},
		{"hr_payroll_runs", "approve", true},
		{"hr_career_paths", "approve", true},
		{"hr_performance_cycles", "launch", true},
		{"hr_leave_entitlements", "post_leave_entry", true},
		{"hr_leave_adjustments", "approve", true},
		{"hr_leave_policies", "create", true},
		{"hr_leave_policies", "update", true},
		{"hr_talent_assessments", "finalize", true},
		{"hr_compensation_records", "read", true},
		{"hr_compensation_bands", "list", true},
		{"hr_worker_competency_assessments", "read", true},
		{"hr_worker_competency_assessments", "create", true},
		{"hr_payroll_profiles", "read", true},
		{"hr_employee_relations_cases", "list", true},
		{"hr_jobs", "read", false},
		{"hr_workers", "update", false},
	} {
		if got := security.IsSensitiveOperation(test.resource, test.action); got != test.sensitive {
			t.Errorf("sensitivity %q.%q = %t, want %t", test.resource, test.action, got, test.sensitive)
		}
	}
}

func TestCatalogPolicySeparatesReadDeleteAndApproval(t *testing.T) {
	t.Parallel()
	now := time.Now().Unix()
	reader := &sessionReader{state: &authsession.SessionValidationContext{
		Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
		Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", IssuedAtUnix: now},
		Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "principal", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
			SyncedAtUnix: now, FreshUntilUnix: now + 30,
			Permissions: []authz.PermissionGrant{
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.workers.delete", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.payroll_runs.approve", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.performance_cycles.launch", Scope: authz.ScopeGrant{Type: "global"}},
			},
		},
	}}
	policy, err := security.NewPolicy(reader, security.PermissionCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
	ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
		Authenticated: true, PrincipalID: "principal", SessionID: "session", AuthEpoch: 1, PermRevision: 2,
	})
	for _, action := range []string{"read", "list"} {
		if _, err := policy.Authorize(ctx, operation.Access{Resource: "hr_workers", Action: action}); err != nil {
			t.Errorf("worker %s was denied: %v", action, err)
		}
	}
	for _, access := range []operation.Access{
		{Resource: "hr_workers", Action: "create"},
		{Resource: "hr_workers", Action: "delete"},
		{Resource: "hr_payroll_runs", Action: "approve"},
		{Resource: "hr_performance_cycles", Action: "launch"},
	} {
		if _, err := policy.Authorize(ctx, access); err == nil {
			t.Errorf("operation %s.%s was allowed without permission or MFA", access.Resource, access.Action)
		}
	}
	reader.state.Token.MFAAuthenticated = true
	for _, access := range []operation.Access{
		{Resource: "hr_workers", Action: "delete"},
		{Resource: "hr_payroll_runs", Action: "approve"},
		{Resource: "hr_performance_cycles", Action: "launch"},
	} {
		if _, err := policy.Authorize(ctx, access); err != nil {
			t.Errorf("operation %s.%s was denied after MFA: %v", access.Resource, access.Action, err)
		}
	}
}
