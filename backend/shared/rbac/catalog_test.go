package rbac

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadEmbeddedCatalogs(t *testing.T) {
	t.Parallel()

	catalogs, err := LoadEmbeddedCatalogs()
	if err != nil {
		t.Fatalf("LoadEmbeddedCatalogs() error = %v", err)
	}

	if len(catalogs.Manifests) != 4 {
		t.Fatalf("manifest count = %d, want 4", len(catalogs.Manifests))
	}
	if len(catalogs.Permissions) != 457 {
		t.Fatalf("permission count = %d, want 457 (15 iam + 10 enterprise_foundation + 409 human_resource + 23 finance_accounting)", len(catalogs.Permissions))
	}
	if len(catalogs.Roles) != 16 {
		t.Fatalf("role count = %d, want 16", len(catalogs.Roles))
	}
	if len(catalogs.PermissionCodesForDomain("finance_accounting")) != 23 {
		t.Fatalf("finance_accounting permissions count = %d, want 23", len(catalogs.PermissionCodesForDomain("finance_accounting")))
	}
	if len(catalogs.PermissionCodesForDomain("iam")) != 15 {
		t.Fatalf("iam permissions count = %d, want 15", len(catalogs.PermissionCodesForDomain("iam")))
	}
	if len(catalogs.PermissionCodesForDomain("enterprise_foundation")) != 10 {
		t.Fatalf("enterprise_foundation permissions count = %d, want 10", len(catalogs.PermissionCodesForDomain("enterprise_foundation")))
	}

	hrmPermissions := catalogs.PermissionCodesForDomain("human_resource")
	wantHRM := []string{
		"human_resource.access_requests.create",
		"human_resource.access_requests.read",
		"human_resource.candidates.create",
		"human_resource.candidates.delete",
		"human_resource.candidates.read",
		"human_resource.candidates.update",
		"human_resource.employment_contracts.create",
		"human_resource.employment_contracts.delete",
		"human_resource.employment_contracts.read",
		"human_resource.employment_contracts.update",
		"human_resource.job_families.create",
		"human_resource.job_families.read",
		"human_resource.jobs.create",
		"human_resource.jobs.read",
		"human_resource.leave_adjustments.approve",
		"human_resource.leave_adjustments.create",
		"human_resource.leave_adjustments.read",
		"human_resource.leave_adjustments.reject",
		"human_resource.leave_entitlements.close",
		"human_resource.leave_entitlements.post",
		"human_resource.leave_entitlements.read",
		"human_resource.leave_ledger_entries.read",
		"human_resource.leave_policies.read",
		"human_resource.offboarding_plans.approve",
		"human_resource.offboarding_plans.cancel",
		"human_resource.offboarding_plans.create",
		"human_resource.offboarding_plans.read",
		"human_resource.offboarding_plans.start",
		"human_resource.offboarding_tasks.create",
		"human_resource.offboarding_tasks.read",
		"human_resource.offboarding_tasks.waive",
		"human_resource.payroll_groups.read",
		"human_resource.payroll_periods.read",
		"human_resource.payroll_result_lines.read",
		"human_resource.payroll_results.read",
		"human_resource.payroll_rule_sets.read",
		"human_resource.payroll_runs.approve",
		"human_resource.payroll_runs.cancel",
		"human_resource.payroll_runs.create",
		"human_resource.payroll_runs.read",
		"human_resource.payroll_runs.request_posting",
		"human_resource.payroll_runs.submit_review",
		"human_resource.payslips.read_self",
		"human_resource.payslips.reissue",
		"human_resource.positions.activate",
		"human_resource.positions.create",
		"human_resource.positions.delete",
		"human_resource.positions.read",
		"human_resource.positions.update",
		"human_resource.requisitions.approve",
		"human_resource.requisitions.close",
		"human_resource.requisitions.create",
		"human_resource.requisitions.delete",
		"human_resource.requisitions.open",
		"human_resource.requisitions.read",
		"human_resource.requisitions.reject",
		"human_resource.requisitions.submit",
		"human_resource.requisitions.update",
		"human_resource.worker_assignment_transfers.read",
		"human_resource.worker_assignments.approve",
		"human_resource.worker_assignments.create",
		"human_resource.worker_assignments.read",
		"human_resource.worker_assignments.transfer",
		"human_resource.workers.activate",
		"human_resource.workers.apply_primary_assignment",
		"human_resource.workers.create",
		"human_resource.workers.read",
		"human_resource.workers.terminate",
		"human_resource.workers.update",
		"human_resource.workforce_plan_lines.create",
		"human_resource.workforce_plan_lines.delete",
		"human_resource.workforce_plan_lines.read",
		"human_resource.workforce_plan_lines.update",
		"human_resource.workforce_plans.cancel",
		"human_resource.workforce_plans.create",
		"human_resource.workforce_plans.delete",
		"human_resource.workforce_plans.read",
		"human_resource.workforce_plans.submit",
		"human_resource.workforce_plans.supersede",
		"human_resource.workforce_plans.update",
	}
	if len(hrmPermissions) != 409 {
		t.Fatalf("HRM permission count = %d, want 409", len(hrmPermissions))
	}
	for _, code := range wantHRM {
		if !slices.Contains(hrmPermissions, code) {
			t.Fatalf("missing expected HRM permission code %q", code)
		}
	}
	if !catalogs.PermissionByCode["human_resource.workers.activate"].IsSensitive {
		t.Fatal("worker activation permission must require sensitive authentication")
	}

	superadmin, ok := catalogs.RoleByCode["superadmin"]
	if !ok {
		t.Fatal("superadmin role missing from embedded catalogs")
	}
	if len(superadmin.PermissionCodes) != 25 || !superadmin.IsSystem || superadmin.DefaultRedirectPath != "/identity" || superadmin.RedirectPriority != 1 {
		t.Fatalf("unexpected superadmin role spec: %+v", superadmin)
	}

	iamOperator, ok := catalogs.RoleByCode["iam_operator"]
	if !ok || slices.Contains(iamOperator.PermissionCodes, "iam.roles.write") || len(iamOperator.PermissionCodes) != 14 {
		t.Fatalf("unexpected iam_operator role spec: %+v", iamOperator)
	}

	efAuto, ok := catalogs.RoleByCode["ef_automation"]
	if !ok || !efAuto.IsSystem || len(efAuto.PermissionCodes) != 10 {
		t.Fatalf("unexpected ef_automation role spec: %+v", efAuto)
	}
}

func TestValidateManifestAndCatalogSetSecurityInvariants(t *testing.T) {
	t.Parallel()

	baseIAM := CatalogManifest{
		SchemaVersion:    1,
		ManagementDomain: "iam",
		Permissions: []PermissionSpec{
			{
				Code:        "iam.users.read",
				Resource:    "iam.users",
				Action:      "read",
				Description: "Read IAM users",
			},
			{
				Code:        "iam.users.readself",
				Resource:    "iam.users",
				Action:      "readself",
				Description: "Read own IAM user profile",
			},
		},
		Roles: []RoleSpec{
			{
				Code:                "superadmin",
				Name:                "Super Administrator",
				Description:         "Platform governance role",
				ManagementDomain:    "iam",
				DefaultRedirectPath: "/identity",
				RedirectPriority:    1,
				IsSystem:            true,
				PermissionCodes:     []string{"iam.users.read"},
			},
		},
	}

	tests := []struct {
		name      string
		manifests []CatalogManifest
		wantErr   string
	}{
		{
			name: "valid business template role with is_system false",
			manifests: []CatalogManifest{
				baseIAM,
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Read warehouse bins",
						},
					},
					Roles: []RoleSpec{
						{
							Code:                "warehouse_supervisor",
							Name:                "Warehouse Supervisor",
							Description:         "Supervises warehouse bins",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "/warehouse",
							RedirectPriority:    30,
							IsSystem:            false,
							PermissionCodes:     []string{"warehouse.bins.read"},
						},
					},
				},
			},
		},
		{
			name: "rejects permission violating domain namespace lock",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "finance.invoices.read",
							Resource:    "finance.invoices",
							Action:      "read",
							Description: "Cross-domain permission",
						},
					},
				},
			},
			wantErr: "violates namespace lock",
		},
		{
			name: "rejects permission code mismatch against resource and action",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.write",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Mismatched action",
						},
					},
				},
			},
			wantErr: "does not match resource",
		},
		{
			name: "rejects empty permission description",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "   ",
						},
					},
				},
			},
			wantErr: "description must not be empty",
		},
		{
			name: "rejects superadmin inheritance of new business domain in iam manifest",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "iam",
					Permissions:      baseIAM.Permissions,
					Roles: []RoleSpec{
						{
							Code:                "superadmin",
							Name:                "Super Administrator",
							Description:         "Platform governance role",
							ManagementDomain:    "iam",
							DefaultRedirectPath: "/identity",
							RedirectPriority:    1,
							IsSystem:            true,
							PermissionCodes:     []string{"iam.users.read", "warehouse.bins.read"},
						},
					},
				},
			},
			wantErr: "cannot auto-inherit non-baseline domain permission",
		},
		{
			name: "rejects non-iam manifest defining system non-automation role",
			manifests: []CatalogManifest{
				baseIAM,
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Read bins",
						},
					},
					Roles: []RoleSpec{
						{
							Code:                "warehouse_supervisor",
							Name:                "Warehouse Supervisor",
							Description:         "Should not be locked system role",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "/warehouse",
							RedirectPriority:    30,
							IsSystem:            true,
							PermissionCodes:     []string{"warehouse.bins.read"},
						},
					},
				},
			},
			wantErr: "cannot set is_system=true",
		},
		{
			name: "rejects foreign manifest creating reserved core role",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Read bins",
						},
					},
					Roles: []RoleSpec{
						{
							Code:                "superadmin",
							Name:                "Hijacked Admin",
							Description:         "Forbidden override",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "/warehouse",
							RedirectPriority:    1,
							IsSystem:            true,
							PermissionCodes:     []string{"warehouse.bins.read"},
						},
					},
				},
			},
			wantErr: "reserved core role",
		},
		{
			name: "rejects non-iam role referencing iam permission",
			manifests: []CatalogManifest{
				baseIAM,
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Read bins",
						},
					},
					Roles: []RoleSpec{
						{
							Code:                "warehouse_manager",
							Name:                "Warehouse Manager",
							Description:         "Tries to grab iam permission",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "/warehouse",
							RedirectPriority:    30,
							IsSystem:            false,
							PermissionCodes:     []string{"warehouse.bins.read", "iam.users.read"},
						},
					},
				},
			},
			wantErr: "cannot reference foreign domain permission",
		},
		{
			name: "rejects open redirect path",
			manifests: []CatalogManifest{
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions: []PermissionSpec{
						{
							Code:        "warehouse.bins.read",
							Resource:    "warehouse.bins",
							Action:      "read",
							Description: "Read bins",
						},
					},
					Roles: []RoleSpec{
						{
							Code:                "warehouse_staff",
							Name:                "Warehouse Staff",
							Description:         "Staff role",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "//evil.example/warehouse",
							RedirectPriority:    30,
							IsSystem:            false,
							PermissionCodes:     []string{"warehouse.bins.read"},
						},
					},
				},
			},
			wantErr: "internal relative route",
		},
		{
			name: "rejects role referencing nonexistent permission",
			manifests: []CatalogManifest{
				baseIAM,
				{
					SchemaVersion:    1,
					ManagementDomain: "warehouse",
					Permissions:      []PermissionSpec{},
					Roles: []RoleSpec{
						{
							Code:                "warehouse_staff",
							Name:                "Warehouse Staff",
							Description:         "Staff role",
							ManagementDomain:    "warehouse",
							DefaultRedirectPath: "/warehouse",
							RedirectPriority:    30,
							IsSystem:            false,
							PermissionCodes:     []string{"warehouse.bins.read"},
						},
					},
				},
			},
			wantErr: "unknown permission code",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateCatalogSet(tc.manifests)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("ValidateCatalogSet() unexpected error: %v", err)
			}
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("ValidateCatalogSet() error = %v, want substring %q", err, tc.wantErr)
				}
			}
		})
	}
}

func TestLoadCatalogsFromFSStrictDecodingAndFilenameMatch(t *testing.T) {
	t.Parallel()

	t.Run("rejects unknown json field", func(t *testing.T) {
		t.Parallel()
		memFS := fstest.MapFS{
			"catalogs/warehouse.json": {
				Data: []byte(`{"schema_version":1,"management_domain":"warehouse","unexpected":true,"permissions":[],"roles":[]}`),
			},
		}
		if _, err := LoadCatalogsFromFS(memFS, "catalogs"); err == nil || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("LoadCatalogsFromFS() error = %v, want unknown field error", err)
		}
	})

	t.Run("rejects management_domain mismatch with filename", func(t *testing.T) {
		t.Parallel()
		memFS := fstest.MapFS{
			"catalogs/warehouse.json": {
				Data: []byte(`{"schema_version":1,"management_domain":"finance","permissions":[],"roles":[]}`),
			},
		}
		if _, err := LoadCatalogsFromFS(memFS, "catalogs"); err == nil || !strings.Contains(err.Error(), "must match filename stem") {
			t.Fatalf("LoadCatalogsFromFS() error = %v, want filename stem mismatch error", err)
		}
	})
}
