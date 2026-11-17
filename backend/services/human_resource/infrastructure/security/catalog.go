package security

import "strings"

// PermissionCatalog mengikat resource aplikasi ke kode izin HRM yang disetujui.
// Daftar eksplisit ini mencegah tabel baru otomatis memperoleh akses bisnis.
func PermissionCatalog() map[string]string {
	resources := []string{
		"access_requests", "application_assessments", "application_stage_history", "applications",
		"benefit_enrollments", "benefit_plans", "candidates", "career_path_transitions", "career_paths",
		"collective_agreements", "compensation_bands", "compensation_records", "competencies",
		"development_actions", "development_plans", "emergency_contacts", "employee_relations_cases",
		"employee_relations_events", "employment_contracts", "engagement_survey_results", "engagement_surveys",
		"human_capital_metric_definitions", "human_capital_metric_snapshots", "interview_feedback",
		"interview_panelists", "interviews", "job_competencies", "job_families", "job_postings", "jobs",
		"learning_course_competencies", "learning_courses", "learning_enrollments", "leave_adjustments",
		"leave_entitlements", "leave_ledger_entries", "leave_policies", "offboarding_plans",
		"offboarding_tasks", "offers", "onboarding_plans", "onboarding_tasks", "pay_components",
		"payroll_adjustments", "payroll_groups", "payroll_periods", "payroll_profiles",
		"payroll_result_lines", "payroll_results", "payroll_rule_sets", "payroll_runs",
		"performance_cycles", "performance_goal_updates", "performance_goals",
		"performance_review_goal_scores", "performance_reviews", "policies", "policy_acknowledgements",
		"positions", "promotion_decisions", "promotion_recommendations", "requisitions",
		"service_request_events", "service_requests", "succession_candidates", "succession_plans",
		"talent_assessment_scores", "talent_assessments", "talent_calibration_decisions",
		"talent_calibration_sessions", "talent_dimensions", "talent_frameworks",
		"talent_pool_memberships", "talent_pools", "talent_review_cycles", "worker_assignments",
		"worker_career_aspirations", "worker_competency_assessments", "worker_documents",
		"worker_lifecycle_events", "worker_qualifications", "workers", "workforce_plan_lines",
		"workforce_plans",
	}
	readOnly := map[string]bool{
		"application_stage_history": true,
		"engagement_survey_results": true,
		"leave_ledger_entries":      true,
		"payroll_result_lines":      true,
		"payroll_results":           true,
		"worker_lifecycle_events":   true,
	}
	permissions := make(map[string]string, 84*5+5)
	for _, resource := range resources {
		operation := "hr_" + resource
		code := "human_resource." + resource + ".read"
		permissions[operation+".read"] = code
		permissions[operation+".list"] = code
		if readOnly[resource] {
			continue
		}
		for _, action := range []string{"create", "update", "delete"} {
			permissions[operation+"."+action] = "human_resource." + resource + "." + action
		}
	}
	permissions["hr_leave_entitlements.post_leave_entry"] = "human_resource.leave_entitlements.post"
	permissions["hr_leave_adjustments.approve"] = "human_resource.leave_adjustments.approve"
	permissions["hr_payroll_runs.approve"] = "human_resource.payroll_runs.approve"
	permissions["hr_career_paths.approve"] = "human_resource.career_paths.approve"
	permissions["hr_performance_cycles.launch"] = "human_resource.performance_cycles.launch"
	permissions["hr_onboarding_plans.start"] = "human_resource.onboarding_plans.start"
	permissions["hr_onboarding_plans.complete"] = "human_resource.onboarding_plans.complete"
	permissions["hr_onboarding_tasks.complete"] = "human_resource.onboarding_tasks.complete"
	permissions["hr_talent_assessments.finalize"] = "human_resource.talent_assessments.finalize"
	permissions["hr_applications.start_screening"] = "human_resource.applications.start_screening"
	permissions["hr_requisitions.submit"] = "human_resource.requisitions.submit"
	permissions["hr_requisitions.approve"] = "human_resource.requisitions.approve"
	permissions["hr_requisitions.open"] = "human_resource.requisitions.open"
	return permissions
}

// IsSensitiveOperation menetapkan batas minimal step-up yang tidak bergantung pada flag katalog IAM.
func IsSensitiveOperation(resource, action string) bool {
	if action == "delete" || action == "approve" || action == "open" || action == "post_leave_entry" || action == "finalize" || action == "launch" || action == "start" || action == "complete" || action == "start_screening" || action == "submit" {
		return true
	}
	if action == "read" || action == "list" {
		return strings.HasPrefix(resource, "hr_compensation_") || resource == "hr_employment_contracts" || resource == "hr_payroll_results" ||
			resource == "hr_payroll_result_lines" || resource == "hr_payroll_profiles" ||
			resource == "hr_payroll_adjustments" || resource == "hr_worker_competency_assessments" ||
			resource == "hr_worker_qualifications" ||
			resource == "hr_performance_reviews" ||
			resource == "hr_succession_plans" || resource == "hr_succession_candidates" ||
			strings.HasPrefix(resource, "hr_employee_relations_")
	}
	if resource == "hr_leave_policies" && (action == "create" || action == "update") {
		return true
	}
	if resource == "hr_positions" && (action == "create" || action == "update") {
		return true
	}
	return strings.HasPrefix(resource, "hr_compensation_") || strings.HasPrefix(resource, "hr_payroll_") ||
		resource == "hr_employment_contracts" ||
		resource == "hr_promotion_decisions" || resource == "hr_worker_competency_assessments" ||
		resource == "hr_worker_qualifications" ||
		resource == "hr_performance_reviews" || resource == "hr_succession_plans" ||
		resource == "hr_succession_candidates"
}
