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
	permissions["hr_compensation_records.approve"] = "human_resource.compensation_records.approve"
	permissions["hr_learning_enrollments.withdraw"] = "human_resource.learning_enrollments.withdraw"
	permissions["hr_learning_enrollments.start"] = "human_resource.learning_enrollments.start"
	permissions["hr_learning_enrollments.complete"] = "human_resource.learning_enrollments.complete"
	permissions["hr_benefit_enrollments.end"] = "human_resource.benefit_enrollments.end"
	permissions["hr_benefit_enrollments.approve"] = "human_resource.benefit_enrollments.approve"
	permissions["hr_service_requests.resolve"] = "human_resource.service_requests.resolve"
	permissions["hr_service_requests.cancel"] = "human_resource.service_requests.cancel"
	permissions["hr_service_requests.close"] = "human_resource.service_requests.close"
	permissions["hr_service_requests.assign"] = "human_resource.service_requests.assign"
	permissions["hr_leave_entitlements.close"] = "human_resource.leave_entitlements.close"
	permissions["hr_leave_entitlements.post_leave_entry"] = "human_resource.leave_entitlements.post"
	permissions["hr_leave_adjustments.reject"] = "human_resource.leave_adjustments.reject"
	permissions["hr_leave_adjustments.approve"] = "human_resource.leave_adjustments.approve"
	permissions["hr_payroll_runs.cancel"] = "human_resource.payroll_runs.cancel"
	permissions["hr_payroll_runs.calculate"] = "human_resource.payroll_runs.calculate"
	permissions["hr_payroll_runs.request_posting"] = "human_resource.payroll_runs.request_posting"
	permissions["hr_payroll_runs.submit_review"] = "human_resource.payroll_runs.submit_review"
	permissions["hr_payroll_runs.approve"] = "human_resource.payroll_runs.approve"
	permissions["hr_development_actions.complete"] = "human_resource.development_actions.complete"
	permissions["hr_development_plans.complete"] = "human_resource.development_plans.complete"
	permissions["hr_development_plans.approve"] = "human_resource.development_plans.approve"
	permissions["hr_succession_plans.approve"] = "human_resource.succession_plans.approve"
	permissions["hr_career_paths.approve"] = "human_resource.career_paths.approve"
	permissions["hr_worker_assignment_transfers.read"] = "human_resource.worker_assignment_transfers.read"
	permissions["hr_worker_assignments.transfer"] = "human_resource.worker_assignments.transfer"
	permissions["hr_worker_assignments.approve"] = "human_resource.worker_assignments.approve"
	permissions["hr_workforce_plans.supersede"] = "human_resource.workforce_plans.supersede"
	permissions["hr_workforce_plans.approve"] = "human_resource.workforce_plans.approve"
	permissions["hr_workforce_plans.cancel"] = "human_resource.workforce_plans.cancel"
	permissions["hr_workforce_plans.submit"] = "human_resource.workforce_plans.submit"
	permissions["hr_positions.activate"] = "human_resource.positions.activate"
	permissions["hr_payslips.read_self"] = "human_resource.payslips.read_self"
	permissions["hr_payslips.reissue"] = "human_resource.payslips.reissue"
	permissions["hr_workers.activate"] = "human_resource.workers.activate"
	permissions["hr_employment_contracts.record_signature"] = "human_resource.employment_contracts.record_signature"
	// Penautan principal hanya dilayani sebagai bagian dari provisioning akun IAM.
	permissions["hr_workers.link_principal"] = "iam.users.write"
	permissions["hr_workers.terminate"] = "human_resource.workers.terminate"
	permissions["hr_workers.apply_primary_assignment"] = "human_resource.workers.apply_primary_assignment"
	permissions["hr_performance_cycles.launch"] = "human_resource.performance_cycles.launch"
	permissions["hr_offboarding_plans.complete"] = "human_resource.offboarding_plans.complete"
	permissions["hr_offboarding_tasks.complete"] = "human_resource.offboarding_tasks.complete"
	permissions["hr_offboarding_tasks.waive"] = "human_resource.offboarding_tasks.waive"
	permissions["hr_offboarding_plans.start"] = "human_resource.offboarding_plans.start"
	permissions["hr_offboarding_plans.cancel"] = "human_resource.offboarding_plans.cancel"
	permissions["hr_offboarding_plans.approve"] = "human_resource.offboarding_plans.approve"
	permissions["hr_onboarding_plans.start"] = "human_resource.onboarding_plans.start"
	permissions["hr_onboarding_plans.complete"] = "human_resource.onboarding_plans.complete"
	permissions["hr_onboarding_tasks.complete"] = "human_resource.onboarding_tasks.complete"
	permissions["hr_talent_assessments.finalize"] = "human_resource.talent_assessments.finalize"
	permissions["hr_offers.accept"] = "human_resource.offers.accept"
	permissions["hr_offers.decline"] = "human_resource.offers.decline"
	permissions["hr_offers.issue"] = "human_resource.offers.issue"
	permissions["hr_offers.approve"] = "human_resource.offers.approve"
	permissions["hr_offers.withdraw"] = "human_resource.offers.withdraw"
	permissions["hr_interviews.complete"] = "human_resource.interviews.complete"
	permissions["hr_interviews.start"] = "human_resource.interviews.start"
	permissions["hr_interviews.cancel"] = "human_resource.interviews.cancel"
	permissions["hr_applications.start_screening"] = "human_resource.applications.start_screening"
	permissions["hr_applications.advance"] = "human_resource.applications.advance"
	permissions["hr_applications.reject"] = "human_resource.applications.reject"
	permissions["hr_requisitions.submit"] = "human_resource.requisitions.submit"
	permissions["hr_requisitions.approve"] = "human_resource.requisitions.approve"
	permissions["hr_requisitions.reject"] = "human_resource.requisitions.reject"
	permissions["hr_requisitions.open"] = "human_resource.requisitions.open"
	permissions["hr_requisitions.close"] = "human_resource.requisitions.close"
	permissions["hr_job_postings.publish"] = "human_resource.job_postings.publish"
	permissions["hr_job_postings.close"] = "human_resource.job_postings.close"
	permissions["hr_job_postings.withdraw"] = "human_resource.job_postings.withdraw"
	permissions["hr_application_assessments.complete"] = "human_resource.application_assessments.complete"
	permissions["hr_talent_review_cycles.launch"] = "human_resource.talent_review_cycles.launch"
	permissions["hr_talent_review_cycles.start_calibration"] = "human_resource.talent_review_cycles.start_calibration"
	permissions["hr_talent_review_cycles.close"] = "human_resource.talent_review_cycles.close"
	permissions["hr_talent_assessments.submit"] = "human_resource.talent_assessments.submit"
	permissions["hr_talent_assessments.withdraw"] = "human_resource.talent_assessments.withdraw"
	permissions["hr_talent_calibration_sessions.open"] = "human_resource.talent_calibration_sessions.open"
	permissions["hr_talent_calibration_sessions.approve"] = "human_resource.talent_calibration_sessions.approve"
	permissions["hr_talent_pool_memberships.approve"] = "human_resource.talent_pool_memberships.approve"
	permissions["hr_talent_pool_memberships.reject"] = "human_resource.talent_pool_memberships.reject"
	permissions["hr_talent_pool_memberships.exit"] = "human_resource.talent_pool_memberships.exit"
	permissions["hr_performance_goals.submit"] = "human_resource.performance_goals.submit"
	permissions["hr_performance_goals.approve"] = "human_resource.performance_goals.approve"
	permissions["hr_performance_goals.reject"] = "human_resource.performance_goals.reject"
	permissions["hr_performance_reviews.submit"] = "human_resource.performance_reviews.submit"
	permissions["hr_performance_reviews.finalize"] = "human_resource.performance_reviews.finalize"
	permissions["hr_performance_cycles.close"] = "human_resource.performance_cycles.close"
	permissions["hr_promotion_recommendations.submit"] = "human_resource.promotion_recommendations.submit"
	permissions["hr_promotion_recommendations.approve"] = "human_resource.promotion_decisions.approve"
	permissions["hr_promotion_recommendations.reject"] = "human_resource.promotion_decisions.reject"
	permissions["hr_promotion_recommendations.return"] = "human_resource.promotion_decisions.return"
	permissions["hr_promotion_recommendations.apply"] = "human_resource.promotion_decisions.apply"
	permissions["hr_policy_acknowledgements.acknowledge"] = "human_resource.policy_acknowledgements.acknowledge"
	return permissions
}

// IsSensitiveOperation menetapkan batas minimal step-up yang tidak bergantung pada flag katalog IAM.
func IsSensitiveOperation(resource, action string) bool {
	if resource == "hr_payslips" && action == "reissue" {
		return true
	}
	if action == "accept" || action == "decline" || action == "issue" || action == "transfer" || action == "resolve" || action == "assign" || action == "supersede" || action == "cancel" || action == "end" || action == "waive" || action == "delete" || action == "approve" || action == "activate" || action == "terminate" || action == "record_signature" || action == "link_principal" || action == "apply_primary_assignment" || action == "open" || action == "close" || action == "publish" || action == "withdraw" || action == "reject" || action == "advance" || action == "post_leave_entry" || action == "finalize" || action == "launch" || action == "start" || action == "complete" || action == "start_screening" || action == "submit" || action == "start_calibration" || action == "exit" || action == "return" || action == "apply" {
		return true
	}
	if action == "read" || action == "list" {
		return strings.HasPrefix(resource, "hr_compensation_") || resource == "hr_employment_contracts" || resource == "hr_payroll_results" ||
			resource == "hr_payroll_result_lines" || resource == "hr_payroll_profiles" ||
			resource == "hr_payroll_adjustments" || resource == "hr_worker_competency_assessments" ||
			resource == "hr_worker_qualifications" ||
			resource == "hr_performance_reviews" || resource == "hr_promotion_decisions" ||
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
