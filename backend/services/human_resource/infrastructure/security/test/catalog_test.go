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
	if len(catalog) != 425 {
		t.Fatalf("operation mapping count = %d, want 425", len(catalog))
	}
	codes := make(map[string]struct{}, len(catalog))
	for operation, code := range catalog {
		if !strings.HasPrefix(operation, "hr_") || !strings.HasPrefix(code, "human_resource.") {
			t.Fatalf("invalid catalog mapping: %q -> %q", operation, code)
		}
		codes[code] = struct{}{}
	}
	if len(codes) != 341 {
		t.Fatalf("permission count = %d, want 341", len(codes))
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
		"hr_worker_assignments.approve":          "human_resource.worker_assignments.approve",
		"hr_positions.activate":                  "human_resource.positions.activate",
		"hr_workers.apply_primary_assignment":    "human_resource.workers.apply_primary_assignment",
		"hr_performance_cycles.launch":           "human_resource.performance_cycles.launch",
		"hr_onboarding_plans.start":              "human_resource.onboarding_plans.start",
		"hr_onboarding_plans.complete":           "human_resource.onboarding_plans.complete",
		"hr_onboarding_tasks.complete":           "human_resource.onboarding_tasks.complete",
		"hr_talent_assessments.finalize":         "human_resource.talent_assessments.finalize",
		"hr_applications.start_screening":        "human_resource.applications.start_screening",
		"hr_applications.advance":                "human_resource.applications.advance",
		"hr_applications.reject":                 "human_resource.applications.reject",
		"hr_requisitions.submit":                 "human_resource.requisitions.submit",
		"hr_requisitions.approve":                "human_resource.requisitions.approve",
		"hr_requisitions.reject":                 "human_resource.requisitions.reject",
		"hr_requisitions.open":                   "human_resource.requisitions.open",
		"hr_requisitions.close":                  "human_resource.requisitions.close",
		"hr_job_postings.publish":                "human_resource.job_postings.publish",
		"hr_job_postings.close":                  "human_resource.job_postings.close",
		"hr_job_postings.withdraw":               "human_resource.job_postings.withdraw",
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
		{"hr_workers", "apply_primary_assignment", true},
		{"hr_payroll_runs", "approve", true},
		{"hr_career_paths", "approve", true},
		{"hr_performance_cycles", "launch", true},
		{"hr_onboarding_plans", "start", true},
		{"hr_onboarding_plans", "complete", true},
		{"hr_requisitions", "submit", true},
		{"hr_requisitions", "approve", true},
		{"hr_requisitions", "reject", true},
		{"hr_requisitions", "open", true},
		{"hr_requisitions", "close", true},
		{"hr_job_postings", "publish", true},
		{"hr_job_postings", "close", true},
		{"hr_job_postings", "withdraw", true},
		{"hr_applications", "reject", true},
		{"hr_applications", "advance", true},
		{"hr_leave_entitlements", "post_leave_entry", true},
		{"hr_leave_adjustments", "approve", true},
		{"hr_leave_policies", "create", true},
		{"hr_leave_policies", "update", true},
		{"hr_positions", "create", true},
		{"hr_positions", "update", true},
		{"hr_positions", "activate", true},
		{"hr_talent_assessments", "finalize", true},
		{"hr_compensation_records", "read", true},
		{"hr_compensation_bands", "list", true},
		{"hr_worker_competency_assessments", "read", true},
		{"hr_worker_competency_assessments", "create", true},
		{"hr_worker_qualifications", "read", true},
		{"hr_worker_qualifications", "create", true},
		{"hr_employment_contracts", "read", true},
		{"hr_employment_contracts", "create", true},
		{"hr_performance_reviews", "read", true},
		{"hr_performance_reviews", "create", true},
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
		Token:    &authsession.SessionTokenRecord{UserID: "principal", SessionID: "session", UserKind: "employee", IssuedAtUnix: now},
		Security: &authsession.UserSecurityRecord{UserID: "principal", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "principal", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
			SyncedAtUnix: now, FreshUntilUnix: now + 30,
			Permissions: []authz.PermissionGrant{
				{Code: "human_resource.workers.read", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.workers.delete", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.payroll_runs.approve", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.performance_cycles.launch", Scope: authz.ScopeGrant{Type: "global"}},
				{Code: "human_resource.onboarding_plans.start", Scope: authz.ScopeGrant{Type: "global"}},
			},
		},
	}}
	policy, err := security.NewPolicy(reader, security.PermissionCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-ig-session-token-hash", strings.Repeat("a", 64)))
	ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
		Authenticated: true, PrincipalID: "principal", SessionID: "session", UserKind: "employee", AuthEpoch: 1, PermRevision: 2,
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
		{Resource: "hr_onboarding_plans", Action: "start"},
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
		{Resource: "hr_onboarding_plans", Action: "start"},
	} {
		if _, err := policy.Authorize(ctx, access); err != nil {
			t.Errorf("operation %s.%s was denied after MFA: %v", access.Resource, access.Action, err)
		}
	}
}
