
-- Dihasilkan oleh db/schema/generate.sh; jangan edit baseline secara manual.
-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL

CREATE TYPE "hr_worker_category_enum" AS ENUM (
  'employee',
  'fixed_term_employee',
  'contractor',
  'agency_worker',
  'apprentice',
  'intern',
  'volunteer',
  'contingent_worker'
);

CREATE TYPE "hr_worker_status_enum" AS ENUM (
  'preboarding',
  'active',
  'leave_of_absence',
  'suspended',
  'notice_period',
  'terminated',
  'deceased'
);

CREATE TYPE "hr_application_status_enum" AS ENUM (
  'received',
  'screening',
  'shortlisted',
  'interview',
  'assessment',
  'offer',
  'hired',
  'rejected',
  'withdrawn',
  'talent_pool'
);

CREATE TYPE "hr_payroll_run_status_enum" AS ENUM (
  'draft',
  'calculating',
  'calculated',
  'review',
  'approved',
  'posted',
  'paid',
  'cancelled',
  'failed'
);

CREATE TYPE "hr_compensation_status_enum" AS ENUM (
  'proposed',
  'approved',
  'active',
  'superseded',
  'cancelled'
);

CREATE TYPE "hr_performance_cycle_status_enum" AS ENUM (
  'draft',
  'open',
  'self_review',
  'manager_review',
  'calibration',
  'finalized',
  'archived'
);

CREATE TYPE "hr_learning_status_enum" AS ENUM (
  'planned',
  'enrolled',
  'in_progress',
  'completed',
  'failed',
  'withdrawn',
  'expired'
);

CREATE TYPE "hr_case_status_enum" AS ENUM (
  'open',
  'triage',
  'investigation',
  'action_required',
  'resolved',
  'closed',
  'withdrawn'
);

CREATE TYPE "hr_approval_decision_enum" AS ENUM (
  'approve',
  'reject',
  'return',
  'defer'
);

CREATE TABLE "hr_job_families" (
  "id" uuid PRIMARY KEY,
  "family_code" varchar(80) UNIQUE NOT NULL,
  "family_name" varchar(180) NOT NULL,
  "description" text,
  "job_domain" varchar(60) NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true
);

CREATE TABLE "hr_jobs" (
  "id" uuid PRIMARY KEY,
  "job_family_id" uuid NOT NULL,
  "job_code" varchar(100) UNIQUE NOT NULL,
  "job_title" varchar(180) NOT NULL,
  "job_level_code" varchar(60),
  "job_description" text,
  "responsibilities" jsonb NOT NULL DEFAULT ('[]'),
  "qualification_requirements" jsonb NOT NULL DEFAULT ('{}'),
  "valid_from" date NOT NULL,
  "valid_to" date,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_job_validity" CHECK (valid_to IS NULL OR valid_to >= valid_from)
);

CREATE TABLE "hr_positions" (
  "id" uuid PRIMARY KEY,
  "position_code" varchar(100) UNIQUE NOT NULL,
  "position_title" varchar(180) NOT NULL,
  "job_id" uuid NOT NULL,
  "organization_unit_reference_id" uuid NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "reports_to_position_id" uuid,
  "position_status" varchar(30) NOT NULL DEFAULT 'planned',
  "headcount_authorized" numeric(8,2) NOT NULL DEFAULT 1,
  "work_location_reference_id" uuid,
  "cost_center_reference_id" uuid,
  "valid_from" date NOT NULL,
  "valid_to" date,
  CONSTRAINT "ck_hr_position_not_self_manager" CHECK (reports_to_position_id IS NULL OR reports_to_position_id <> id),
  CONSTRAINT "ck_hr_position_values" CHECK (headcount_authorized > 0 AND (valid_to IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_workers" (
  "id" uuid PRIMARY KEY,
  "worker_number" varchar(100) UNIQUE NOT NULL,
  "person_party_reference_id" uuid NOT NULL,
  "worker_category" hr_worker_category_enum NOT NULL,
  "status" hr_worker_status_enum NOT NULL DEFAULT 'preboarding',
  "hire_date" date,
  "service_start_date" date,
  "termination_date" date,
  "termination_reason_code" varchar(80),
  "primary_legal_entity_reference_id" uuid NOT NULL,
  "primary_assignment_id" uuid,
  "iam_principal_reference_id" uuid,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  "created_by_principal_id" uuid,
  CONSTRAINT "ck_hr_worker_dates" CHECK (termination_date IS NULL OR hire_date IS NULL OR termination_date >= hire_date)
);

CREATE TABLE "hr_employment_contracts" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "contract_number" varchar(120) UNIQUE NOT NULL,
  "contract_type" varchar(60) NOT NULL,
  "contract_status" varchar(30) NOT NULL DEFAULT 'draft',
  "employing_legal_entity_reference_id" uuid NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "signed_at" timestamptz,
  "signed_document_reference_id" uuid,
  "signed_document_version_reference_id" uuid,
  "signed_content_digest" bytea,
  "probation_end_date" date,
  "notice_period_days" integer,
  "work_country_code" char(2) NOT NULL,
  "governing_jurisdiction_code" varchar(80) NOT NULL,
  "collective_agreement_id" uuid,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_contract_signed_evidence" CHECK (signed_at IS NULL OR (signed_document_reference_id IS NOT NULL AND signed_document_version_reference_id IS NOT NULL AND signed_content_digest IS NOT NULL)),
  CONSTRAINT "ck_hr_contract_effective_dates" CHECK (effective_to IS NULL OR effective_to >= effective_from),
  CONSTRAINT "ck_hr_contract_terms" CHECK ((probation_end_date IS NULL OR probation_end_date >= effective_from) AND (notice_period_days IS NULL OR notice_period_days >= 0))
);

CREATE TABLE "hr_worker_assignments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "position_id" uuid NOT NULL,
  "manager_worker_id" uuid,
  "assignment_type" varchar(40) NOT NULL,
  "employment_status_snapshot" varchar(40),
  "fte_percent" numeric(6,3) NOT NULL DEFAULT 100,
  "valid_from" date NOT NULL,
  "valid_to" date,
  "assignment_reason" varchar(80),
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_assignment_not_self_manager" CHECK (manager_worker_id IS NULL OR manager_worker_id <> worker_id),
  CONSTRAINT "ck_hr_assignment_values" CHECK (fte_percent > 0 AND fte_percent <= 100 AND (valid_to IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_worker_lifecycle_events" (
  "id" BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "event_type" varchar(60) NOT NULL,
  "occurred_at" timestamptz NOT NULL,
  "effective_date" date NOT NULL,
  "previous_status" hr_worker_status_enum,
  "new_status" hr_worker_status_enum,
  "source_reference" varchar(160),
  "actor_principal_id" uuid,
  "event_payload" jsonb NOT NULL DEFAULT ('{}'),
  "idempotency_key" varchar(160) UNIQUE NOT NULL
);

CREATE TABLE "hr_worker_documents" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "document_type" varchar(60) NOT NULL,
  "document_reference_id" uuid NOT NULL,
  "document_version_reference_id" uuid NOT NULL,
  "content_digest" bytea NOT NULL,
  "document_number_digest" bytea,
  "issuing_country_code" char(2),
  "issued_on" date,
  "expires_on" date,
  "verification_status" varchar(30) NOT NULL DEFAULT 'pending',
  "verified_by_principal_id" uuid,
  "verified_at" timestamptz,
  "retention_class" varchar(80) NOT NULL,
  "access_class" varchar(30) NOT NULL DEFAULT 'restricted',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_worker_document_dates" CHECK (expires_on IS NULL OR issued_on IS NULL OR expires_on >= issued_on)
);

CREATE TABLE "hr_emergency_contacts" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "contact_party_reference_id" uuid NOT NULL,
  "relationship_code" varchar(60) NOT NULL,
  "priority_order" integer NOT NULL,
  "authorization_confirmed_at" timestamptz,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_emergency_contact_priority" CHECK (priority_order > 0)
);

CREATE TABLE "hr_requisitions" (
  "id" uuid PRIMARY KEY,
  "requisition_number" varchar(120) UNIQUE NOT NULL,
  "workforce_plan_line_id" uuid,
  "position_id" uuid,
  "job_id" uuid NOT NULL,
  "organization_unit_reference_id" uuid NOT NULL,
  "hiring_manager_principal_id" uuid NOT NULL,
  "recruiter_principal_id" uuid,
  "requisition_type" varchar(40) NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "headcount" integer NOT NULL DEFAULT 1,
  "target_start_date" date,
  "budget_reference_id" uuid,
  "work_location_reference_id" uuid,
  "employment_category" hr_worker_category_enum NOT NULL,
  "minimum_qualifications" jsonb NOT NULL DEFAULT ('{}'),
  "desired_qualifications" jsonb NOT NULL DEFAULT ('{}'),
  "approved_at" timestamptz,
  "approved_by_principal_id" uuid,
  "opened_at" timestamptz,
  "closed_at" timestamptz,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_requisition_values" CHECK (headcount > 0 AND (closed_at IS NULL OR opened_at IS NULL OR closed_at >= opened_at))
);

CREATE TABLE "hr_candidates" (
  "id" uuid PRIMARY KEY,
  "candidate_number" varchar(120) UNIQUE NOT NULL,
  "person_party_reference_id" uuid,
  "preferred_name" varchar(180),
  "contact_reference_id" uuid,
  "source_type" varchar(60),
  "source_reference" varchar(160),
  "privacy_notice_version" varchar(60) NOT NULL,
  "notice_acknowledged_at" timestamptz,
  "data_retention_until" date,
  "candidate_status" varchar(30) NOT NULL DEFAULT 'active',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "created_by_principal_id" uuid
);

CREATE TABLE "hr_applications" (
  "id" uuid PRIMARY KEY,
  "candidate_id" uuid NOT NULL,
  "requisition_id" uuid NOT NULL,
  "application_number" varchar(120) UNIQUE NOT NULL,
  "status" hr_application_status_enum NOT NULL DEFAULT 'received',
  "applied_at" timestamptz NOT NULL DEFAULT (now()),
  "current_stage_code" varchar(80),
  "assigned_recruiter_principal_id" uuid,
  "screening_score" numeric(8,3),
  "scorecard_version" varchar(60),
  "disposition_reason_code" varchar(80),
  "decision_at" timestamptz,
  "decision_by_principal_id" uuid,
  "hired_worker_id" uuid,
  "source_posting_id" uuid,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_application_screening_score" CHECK (screening_score IS NULL OR screening_score BETWEEN 0 AND 100),
  CONSTRAINT "ck_hr_application_hired" CHECK (status <> 'hired' OR (hired_worker_id IS NOT NULL AND decision_at IS NOT NULL AND decision_by_principal_id IS NOT NULL))
);

CREATE TABLE "hr_application_stage_history" (
  "id" BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "application_id" uuid NOT NULL,
  "previous_status" hr_application_status_enum,
  "new_status" hr_application_status_enum NOT NULL,
  "stage_code" varchar(80),
  "changed_at" timestamptz NOT NULL DEFAULT (now()),
  "changed_by_principal_id" uuid,
  "reason_code" varchar(80)
);

CREATE TABLE "hr_interviews" (
  "id" uuid PRIMARY KEY,
  "application_id" uuid NOT NULL,
  "interview_round" integer NOT NULL,
  "interview_type" varchar(60) NOT NULL,
  "scheduled_start" timestamptz,
  "scheduled_end" timestamptz,
  "location_reference_id" uuid,
  "meeting_reference_id" uuid,
  "status" varchar(30) NOT NULL DEFAULT 'planned',
  "candidate_timezone" varchar(80),
  "created_by_principal_id" uuid,
  CONSTRAINT "ck_hr_interview_schedule" CHECK (interview_round > 0 AND (scheduled_end IS NULL OR scheduled_start IS NULL OR scheduled_end > scheduled_start))
);

CREATE TABLE "hr_interview_panelists" (
  "id" uuid PRIMARY KEY,
  "interview_id" uuid NOT NULL,
  "interviewer_principal_id" uuid NOT NULL,
  "interviewer_role" varchar(60),
  "scorecard_template_reference_id" uuid,
  "feedback_due_at" timestamptz
);

CREATE TABLE "hr_interview_feedback" (
  "id" uuid PRIMARY KEY,
  "interview_id" uuid NOT NULL,
  "interviewer_principal_id" uuid NOT NULL,
  "feedback_version" integer NOT NULL DEFAULT 1,
  "submitted_at" timestamptz NOT NULL DEFAULT (now()),
  "overall_recommendation" varchar(40) NOT NULL,
  "competency_scores" jsonb NOT NULL,
  "evidence_notes" text,
  "content_digest" bytea,
  "visibility_class" varchar(30) NOT NULL DEFAULT 'restricted',
  CONSTRAINT "ck_hr_interview_feedback_version" CHECK (feedback_version > 0)
);

CREATE TABLE "hr_offers" (
  "id" uuid PRIMARY KEY,
  "application_id" uuid NOT NULL,
  "offer_number" varchar(120) UNIQUE NOT NULL,
  "offer_version" integer NOT NULL DEFAULT 1,
  "offer_status" varchar(30) NOT NULL DEFAULT 'draft',
  "proposed_job_id" uuid NOT NULL,
  "proposed_legal_entity_reference_id" uuid NOT NULL,
  "proposed_start_date" date,
  "proposed_contract_type" varchar(60) NOT NULL,
  "proposed_compensation_document_version_reference_id" uuid,
  "currency_code" char(3),
  "offer_document_reference_id" uuid,
  "offer_document_version_reference_id" uuid,
  "offer_content_digest" bytea,
  "issued_at" timestamptz,
  "expires_at" timestamptz,
  "candidate_response" varchar(30),
  "responded_at" timestamptz,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_offer_issued_evidence" CHECK (issued_at IS NULL OR (offer_document_version_reference_id IS NOT NULL AND offer_content_digest IS NOT NULL AND approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL)),
  CONSTRAINT "ck_hr_offer_validity" CHECK (offer_version > 0 AND (expires_at IS NULL OR issued_at IS NULL OR expires_at > issued_at))
);

CREATE TABLE "hr_onboarding_plans" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "plan_code" varchar(100) NOT NULL,
  "plan_version" integer NOT NULL DEFAULT 1,
  "start_date" date NOT NULL,
  "target_completion_date" date,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "template_reference_id" uuid,
  "created_by_principal_id" uuid,
  "started_at" timestamptz,
  "started_by_principal_id" uuid,
  "completed_at" timestamptz,
  "completed_by_principal_id" uuid,
  CONSTRAINT "ck_hr_onboarding_plan_dates" CHECK (plan_version > 0 AND (target_completion_date IS NULL OR target_completion_date >= start_date)),
  CONSTRAINT "ck_hr_onboarding_plan_lifecycle" CHECK ((status = 'draft' AND started_at IS NULL AND started_by_principal_id IS NULL AND completed_at IS NULL AND completed_by_principal_id IS NULL) OR (status = 'in_progress' AND started_at IS NOT NULL AND started_by_principal_id IS NOT NULL AND completed_at IS NULL AND completed_by_principal_id IS NULL) OR (status = 'completed' AND started_at IS NOT NULL AND started_by_principal_id IS NOT NULL AND completed_at IS NOT NULL AND completed_by_principal_id IS NOT NULL AND completed_at >= started_at))
);

CREATE TABLE "hr_onboarding_tasks" (
  "id" uuid PRIMARY KEY,
  "onboarding_plan_id" uuid NOT NULL,
  "task_code" varchar(100) NOT NULL,
  "task_name" varchar(180) NOT NULL,
  "task_owner_type" varchar(30) NOT NULL,
  "assigned_principal_id" uuid,
  "due_at" timestamptz,
  "completed_at" timestamptz,
  "completed_by_principal_id" uuid,
  "completion_evidence_reference_id" uuid,
  "is_mandatory" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_onboarding_task_completion" CHECK ((completed_at IS NULL AND completed_by_principal_id IS NULL AND completion_evidence_reference_id IS NULL) OR (completed_at IS NOT NULL AND completed_by_principal_id IS NOT NULL AND completion_evidence_reference_id IS NOT NULL))
);

CREATE TABLE "hr_compensation_bands" (
  "id" uuid PRIMARY KEY,
  "band_code" varchar(80) NOT NULL,
  "band_version" integer NOT NULL,
  "job_level_code" varchar(60) NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "currency_code" char(3) NOT NULL,
  "pay_frequency" varchar(30) NOT NULL,
  "minimum_amount" numeric(20,6) NOT NULL,
  "midpoint_amount" numeric(20,6) NOT NULL,
  "maximum_amount" numeric(20,6) NOT NULL,
  "valid_from" date NOT NULL,
  "valid_to" date,
  "approval_reference_id" uuid,
  CONSTRAINT "ck_hr_compensation_band_range" CHECK (band_version > 0 AND minimum_amount >= 0 AND midpoint_amount >= minimum_amount AND maximum_amount >= midpoint_amount AND (valid_to IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_compensation_records" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "assignment_id" uuid NOT NULL,
  "component_id" uuid NOT NULL,
  "component_code_snapshot" varchar(80) NOT NULL,
  "compensation_band_id" uuid,
  "compensation_type" varchar(40) NOT NULL,
  "amount" numeric(20,6) NOT NULL,
  "currency_code" char(3) NOT NULL,
  "pay_frequency" varchar(30) NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "status" hr_compensation_status_enum NOT NULL DEFAULT 'proposed',
  "reason_code" varchar(80),
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "source_change_reference_id" uuid,
  "access_class" varchar(30) NOT NULL DEFAULT 'restricted',
  CONSTRAINT "ck_hr_compensation_record_values" CHECK (amount >= 0 AND (effective_to IS NULL OR effective_to >= effective_from)),
  CONSTRAINT "ck_hr_compensation_approval" CHECK (status NOT IN ('approved','active','superseded') OR (approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL))
);

CREATE TABLE "hr_benefit_plans" (
  "id" uuid PRIMARY KEY,
  "plan_code" varchar(100) NOT NULL,
  "plan_version" integer NOT NULL,
  "plan_name" varchar(180) NOT NULL,
  "benefit_type" varchar(60) NOT NULL,
  "provider_party_reference_id" uuid,
  "jurisdiction_code" varchar(80) NOT NULL,
  "eligibility_rule" jsonb NOT NULL DEFAULT ('{}'),
  "employee_contribution_rule_reference_id" uuid,
  "employer_contribution_rule_reference_id" uuid,
  "valid_from" date NOT NULL,
  "valid_to" date,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_benefit_plan_validity" CHECK (plan_version > 0 AND (valid_to IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_benefit_enrollments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "benefit_plan_id" uuid NOT NULL,
  "enrollment_status" varchar(30) NOT NULL DEFAULT 'pending',
  "effective_from" date NOT NULL,
  "effective_to" date,
  "enrollment_reference" varchar(160),
  "provider_member_reference" varchar(200),
  "employee_contribution_amount" numeric(20,6),
  "employer_contribution_amount" numeric(20,6),
  "currency_code" char(3),
  "elected_at" timestamptz,
  "verified_at" timestamptz,
  CONSTRAINT "ck_hr_benefit_enrollment_period" CHECK (effective_to IS NULL OR effective_to >= effective_from),
  CONSTRAINT "ck_hr_benefit_contributions" CHECK ((employee_contribution_amount IS NULL OR employee_contribution_amount >= 0) AND (employer_contribution_amount IS NULL OR employer_contribution_amount >= 0))
);

CREATE TABLE "hr_payroll_profiles" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "payroll_country_code" char(2) NOT NULL,
  "payroll_jurisdiction_code" varchar(80) NOT NULL,
  "payroll_group_id" uuid NOT NULL,
  "employing_legal_entity_reference_id" uuid NOT NULL,
  "currency_code" char(3) NOT NULL,
  "tax_profile_reference_id" uuid,
  "bank_account_token_reference_id" uuid,
  "tax_status_code" varchar(60),
  "withholding_election_reference_id" uuid,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_payroll_profile_period" CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE TABLE "hr_pay_components" (
  "id" uuid PRIMARY KEY,
  "component_code" varchar(80) NOT NULL,
  "component_version" integer NOT NULL,
  "component_name" varchar(180) NOT NULL,
  "component_type" varchar(30) NOT NULL,
  "calculation_method" varchar(40) NOT NULL,
  "tax_treatment_code" varchar(80),
  "jurisdiction_code" varchar(80) NOT NULL,
  "currency_code" char(3),
  "rule_set_id" uuid,
  "formula_reference_id" uuid,
  "taxable" boolean,
  "pensionable" boolean,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_component_type" CHECK (component_type IN ('earning','deduction','employer_contribution','tax','reimbursement')),
  CONSTRAINT "ck_hr_pay_component_period" CHECK (component_version > 0 AND (effective_to IS NULL OR effective_to >= effective_from))
);

CREATE TABLE "hr_payroll_periods" (
  "id" uuid PRIMARY KEY,
  "payroll_group_id" uuid NOT NULL,
  "employing_legal_entity_reference_id" uuid NOT NULL,
  "currency_code" char(3) NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "period_code" varchar(80) NOT NULL,
  "period_start" date NOT NULL,
  "period_end" date NOT NULL,
  "pay_date" date NOT NULL,
  "period_status" varchar(30) NOT NULL DEFAULT 'open',
  "calendar_reference_id" uuid,
  CONSTRAINT "ck_hr_payroll_period_dates" CHECK (period_end >= period_start)
);

CREATE TABLE "hr_payroll_runs" (
  "id" uuid PRIMARY KEY,
  "payroll_period_id" uuid NOT NULL,
  "payroll_group_id" uuid NOT NULL,
  "employing_legal_entity_reference_id" uuid NOT NULL,
  "currency_code" char(3) NOT NULL,
  "rule_set_id" uuid NOT NULL,
  "original_run_id" uuid,
  "settlement_direction" varchar(10) NOT NULL DEFAULT 'payable',
  "run_number" integer NOT NULL,
  "run_type" varchar(40) NOT NULL,
  "status" hr_payroll_run_status_enum NOT NULL DEFAULT 'draft',
  "jurisdiction_code" varchar(80) NOT NULL,
  "rule_engine_name" varchar(120) NOT NULL,
  "rule_engine_version" varchar(80) NOT NULL,
  "input_snapshot_reference_id" uuid NOT NULL,
  "input_digest" bytea NOT NULL,
  "output_digest" bytea,
  "calculated_at" timestamptz,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "finance_journal_reference_id" uuid,
  "payment_batch_reference_id" uuid,
  "finance_settlement_event_reference_id" uuid,
  "finance_settled_at" timestamptz,
  "started_by_principal_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_payroll_run_type" CHECK (run_type IN ('regular','off_cycle','bonus','correction','termination','reversal')),
  CONSTRAINT "ck_hr_payroll_settlement_direction" CHECK (settlement_direction IN ('payable','recovery')),
  CONSTRAINT "ck_hr_payroll_correction_origin" CHECK ((run_type IN ('correction','reversal')) = (original_run_id IS NOT NULL)),
  CONSTRAINT "ck_hr_payroll_not_self_correction" CHECK (original_run_id IS NULL OR original_run_id <> id),
  CONSTRAINT "ck_hr_payroll_settlement_evidence" CHECK (status <> 'paid' OR (payment_batch_reference_id IS NOT NULL AND finance_settlement_event_reference_id IS NOT NULL AND finance_settled_at IS NOT NULL)),
  CONSTRAINT "ck_hr_payroll_run_number" CHECK (run_number > 0),
  CONSTRAINT "ck_hr_payroll_run_calculation_state" CHECK ((status NOT IN ('calculated','review','approved','posted','paid') OR (calculated_at IS NOT NULL AND output_digest IS NOT NULL))),
  CONSTRAINT "ck_hr_payroll_run_approval_state" CHECK ((status NOT IN ('approved','posted','paid') OR (approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL))),
  CONSTRAINT "ck_hr_payroll_run_posting_state" CHECK ((status NOT IN ('posted','paid') OR finance_journal_reference_id IS NOT NULL))
);

CREATE TABLE "hr_payroll_results" (
  "id" uuid PRIMARY KEY,
  "payroll_run_id" uuid NOT NULL,
  "payroll_group_id" uuid NOT NULL,
  "worker_id" uuid NOT NULL,
  "payroll_profile_reference_id" uuid NOT NULL,
  "currency_code" char(3) NOT NULL,
  "gross_amount" numeric(20,6) NOT NULL,
  "reimbursement_amount" numeric(20,6) NOT NULL DEFAULT 0,
  "total_deductions" numeric(20,6) NOT NULL,
  "employer_cost_amount" numeric(20,6) NOT NULL,
  "net_pay_amount" numeric(20,6) NOT NULL,
  "result_status" varchar(30) NOT NULL DEFAULT 'calculated',
  "payment_reference_id" uuid,
  "payslip_document_reference_id" uuid,
  "calculation_digest" bytea NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_payroll_result_nonnegative" CHECK (reimbursement_amount >= 0 AND gross_amount >= 0 AND total_deductions >= 0 AND employer_cost_amount >= 0 AND net_pay_amount >= 0),
  CONSTRAINT "ck_hr_payroll_net_equation" CHECK (net_pay_amount = gross_amount + reimbursement_amount - total_deductions)
);

CREATE TABLE "hr_payroll_result_lines" (
  "id" uuid PRIMARY KEY,
  "payroll_result_id" uuid NOT NULL,
  "component_id" uuid NOT NULL,
  "component_code_snapshot" varchar(80) NOT NULL,
  "component_name_snapshot" varchar(180) NOT NULL,
  "component_type_snapshot" varchar(30) NOT NULL,
  "amount" numeric(20,6) NOT NULL,
  "quantity" numeric(20,8),
  "rate" numeric(20,8),
  "taxable_base" numeric(20,6),
  "calculation_reference" varchar(160),
  "line_sequence" integer NOT NULL,
  CONSTRAINT "ck_hr_result_line_type" CHECK (component_type_snapshot IN ('earning','deduction','employer_contribution','tax','reimbursement')),
  CONSTRAINT "ck_hr_payroll_result_line" CHECK (line_sequence > 0 AND (taxable_base IS NULL OR taxable_base >= 0))
);

CREATE TABLE "hr_payroll_adjustments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "payroll_period_id" uuid NOT NULL,
  "component_id" uuid NOT NULL,
  "adjustment_type" varchar(40) NOT NULL,
  "amount" numeric(20,6) NOT NULL,
  "currency_code" char(3) NOT NULL,
  "reason" text NOT NULL,
  "source_reference" varchar(160),
  "requested_by_principal_id" uuid NOT NULL,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "hr_leave_policies" (
  "id" uuid PRIMARY KEY,
  "policy_code" varchar(100) NOT NULL,
  "policy_version" integer NOT NULL,
  "policy_name" varchar(180) NOT NULL,
  "leave_type_code" varchar(80) NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "accrual_method" varchar(40) NOT NULL,
  "annual_entitlement" numeric(12,4),
  "carryover_limit" numeric(12,4),
  "max_balance" numeric(12,4),
  "minimum_service_days" integer,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "is_paid" boolean NOT NULL DEFAULT true,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_leave_policy_amounts" CHECK (policy_version > 0 AND (annual_entitlement IS NULL OR annual_entitlement >= 0) AND (carryover_limit IS NULL OR carryover_limit >= 0) AND (max_balance IS NULL OR max_balance >= 0)),
  CONSTRAINT "ck_hr_leave_policy_period" CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE TABLE "hr_leave_entitlements" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "leave_policy_id" uuid NOT NULL,
  "leave_type_code_snapshot" varchar(80) NOT NULL,
  "entitlement_period_start" date NOT NULL,
  "entitlement_period_end" date NOT NULL,
  "opening_balance" numeric(12,4) NOT NULL DEFAULT 0,
  "accrued_quantity" numeric(12,4) NOT NULL DEFAULT 0,
  "used_quantity_snapshot" numeric(12,4) NOT NULL DEFAULT 0,
  "adjusted_quantity" numeric(12,4) NOT NULL DEFAULT 0,
  "closing_balance_snapshot" numeric(12,4) NOT NULL DEFAULT 0,
  "uom_code" varchar(30) NOT NULL DEFAULT 'day',
  "revision" bigint NOT NULL DEFAULT 1,
  "calculated_at" timestamptz NOT NULL DEFAULT (now()),
  "source_workforce_snapshot_reference_id" uuid,
  CONSTRAINT "ck_hr_leave_balance_equation" CHECK (revision > 0 AND closing_balance_snapshot = opening_balance + accrued_quantity + adjusted_quantity - used_quantity_snapshot),
  CONSTRAINT "ck_hr_leave_entitlement_dates" CHECK (entitlement_period_end >= entitlement_period_start),
  CONSTRAINT "ck_hr_leave_entitlement_values" CHECK (opening_balance >= 0 AND accrued_quantity >= 0 AND used_quantity_snapshot >= 0 AND closing_balance_snapshot >= 0)
);

CREATE TABLE "hr_leave_adjustments" (
  "id" uuid PRIMARY KEY,
  "entitlement_id" uuid NOT NULL,
  "quantity_delta" numeric(12,4) NOT NULL,
  "reason_code" varchar(80) NOT NULL,
  "reason" text NOT NULL,
  "requested_by_principal_id" uuid NOT NULL,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_leave_adjustment_nonzero" CHECK (quantity_delta <> 0),
  CONSTRAINT "ck_hr_leave_adjustment_approval" CHECK ((approved_at IS NULL) = (approved_by_principal_id IS NULL))
);

CREATE TABLE "hr_performance_cycles" (
  "id" uuid PRIMARY KEY,
  "cycle_code" varchar(100) UNIQUE NOT NULL,
  "cycle_name" varchar(180) NOT NULL,
  "cycle_type" varchar(40) NOT NULL,
  "status" hr_performance_cycle_status_enum NOT NULL DEFAULT 'draft',
  "period_start" date NOT NULL,
  "period_end" date NOT NULL,
  "self_review_due_at" timestamptz,
  "manager_review_due_at" timestamptz,
  "calibration_due_at" timestamptz,
  "launched_at" timestamptz,
  "launched_by_principal_id" uuid,
  "configuration" jsonb NOT NULL DEFAULT ('{}'),
  CONSTRAINT "ck_hr_performance_cycle_dates" CHECK (period_end >= period_start)
);

CREATE TABLE "hr_performance_goals" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "performance_cycle_id" uuid NOT NULL,
  "parent_goal_id" uuid,
  "goal_code" varchar(100),
  "goal_title" varchar(240) NOT NULL,
  "goal_description" text,
  "goal_type" varchar(40) NOT NULL,
  "measure_definition" jsonb,
  "target_value" numeric(24,8),
  "actual_value" numeric(24,8),
  "unit_code" varchar(30),
  "weight_percent" numeric(6,3),
  "due_date" date,
  "goal_status" varchar(30) NOT NULL DEFAULT 'draft',
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_performance_goal_weight" CHECK (weight_percent IS NULL OR weight_percent BETWEEN 0 AND 100)
);

CREATE TABLE "hr_performance_reviews" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "performance_cycle_id" uuid NOT NULL,
  "reviewer_principal_id" uuid NOT NULL,
  "reviewer_role" varchar(40) NOT NULL,
  "review_version" integer NOT NULL DEFAULT 1,
  "overall_rating_code" varchar(60),
  "competency_assessment" jsonb NOT NULL DEFAULT ('{}'),
  "narrative" text,
  "development_plan_id" uuid,
  "submitted_at" timestamptz,
  "finalized_at" timestamptz,
  "visibility_class" varchar(30) NOT NULL DEFAULT 'restricted',
  "content_digest" bytea,
  CONSTRAINT "ck_hr_performance_review_version" CHECK (review_version > 0)
);

CREATE TABLE "hr_learning_courses" (
  "id" uuid PRIMARY KEY,
  "course_code" varchar(100) NOT NULL,
  "course_version" varchar(60) NOT NULL,
  "course_name" varchar(240) NOT NULL,
  "course_type" varchar(40) NOT NULL,
  "provider_party_reference_id" uuid,
  "delivery_mode" varchar(40) NOT NULL,
  "duration_hours" numeric(10,2),
  "credential_awarded" boolean NOT NULL DEFAULT false,
  "expiry_months" integer,
  "valid_from" date,
  "valid_to" date,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_course_values" CHECK ((duration_hours IS NULL OR duration_hours >= 0) AND (expiry_months IS NULL OR expiry_months > 0) AND (valid_to IS NULL OR valid_from IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_learning_enrollments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "course_id" uuid NOT NULL,
  "enrollment_status" hr_learning_status_enum NOT NULL DEFAULT 'planned',
  "assigned_by_principal_id" uuid,
  "assigned_at" timestamptz NOT NULL DEFAULT (now()),
  "due_at" timestamptz,
  "started_at" timestamptz,
  "completed_at" timestamptz,
  "score" numeric(8,3),
  "credential_reference_id" uuid,
  "funding_reference_id" uuid,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_learning_score" CHECK (score IS NULL OR score BETWEEN 0 AND 100)
);

CREATE TABLE "hr_worker_qualifications" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "qualification_type" varchar(60) NOT NULL,
  "qualification_code" varchar(120),
  "title" varchar(240) NOT NULL,
  "issuer_party_reference_id" uuid,
  "credential_reference_id" uuid,
  "obtained_on" date,
  "expires_on" date,
  "verification_status" varchar(30) NOT NULL DEFAULT 'unverified',
  "verified_by_principal_id" uuid,
  "verified_at" timestamptz,
  CONSTRAINT "ck_hr_qualification_dates" CHECK (expires_on IS NULL OR obtained_on IS NULL OR expires_on >= obtained_on)
);

CREATE TABLE "hr_succession_plans" (
  "id" uuid PRIMARY KEY,
  "position_id" uuid NOT NULL,
  "plan_code" varchar(100) NOT NULL,
  "plan_version" integer NOT NULL,
  "review_date" date NOT NULL,
  "confidentiality_class" varchar(30) NOT NULL DEFAULT 'highly_restricted',
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "succession_risk" varchar(30),
  CONSTRAINT "ck_hr_succession_plan_version" CHECK (plan_version > 0)
);

CREATE TABLE "hr_succession_candidates" (
  "id" uuid PRIMARY KEY,
  "succession_plan_id" uuid NOT NULL,
  "candidate_worker_id" uuid NOT NULL,
  "readiness_code" varchar(40) NOT NULL,
  "readiness_assessed_at" date NOT NULL,
  "readiness_rationale" text,
  "talent_assessment_id" uuid,
  "development_plan_id" uuid,
  "risk_of_loss" varchar(30)
);

CREATE TABLE "hr_employee_relations_cases" (
  "id" uuid PRIMARY KEY,
  "case_number" varchar(120) UNIQUE NOT NULL,
  "case_type" varchar(60) NOT NULL,
  "status" hr_case_status_enum NOT NULL DEFAULT 'open',
  "subject_worker_id" uuid,
  "reporting_party_reference_id" uuid,
  "assigned_investigator_principal_id" uuid,
  "opened_at" timestamptz NOT NULL DEFAULT (now()),
  "due_at" timestamptz,
  "resolved_at" timestamptz,
  "outcome_code" varchar(80),
  "restricted_summary_reference_id" uuid NOT NULL,
  "confidentiality_class" varchar(30) NOT NULL DEFAULT 'highly_restricted',
  "legal_hold_reference_id" uuid,
  "idempotency_key" varchar(160) UNIQUE NOT NULL
);

CREATE TABLE "hr_employee_relations_events" (
  "id" BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "case_id" uuid NOT NULL,
  "event_type" varchar(60) NOT NULL,
  "occurred_at" timestamptz NOT NULL DEFAULT (now()),
  "actor_principal_id" uuid,
  "event_digest" bytea NOT NULL,
  "event_reference_id" uuid
);

CREATE TABLE "hr_human_capital_metric_definitions" (
  "id" uuid PRIMARY KEY,
  "metric_code" varchar(100) NOT NULL,
  "metric_version" varchar(60) NOT NULL,
  "metric_name" varchar(180) NOT NULL,
  "metric_domain" varchar(60) NOT NULL,
  "numerator_definition" text,
  "denominator_definition" text,
  "inclusion_exclusion_rules" jsonb,
  "unit_code" varchar(30),
  "standard_reference" varchar(120)
);

CREATE TABLE "hr_human_capital_metric_snapshots" (
  "id" uuid PRIMARY KEY,
  "metric_definition_id" uuid NOT NULL,
  "period_start" date NOT NULL,
  "period_end" date NOT NULL,
  "organization_unit_reference_id" uuid,
  "jurisdiction_code" varchar(80),
  "value_numeric" numeric(24,8),
  "value_text" varchar(500),
  "population_count" bigint,
  "data_quality_status" varchar(30) NOT NULL DEFAULT 'provisional',
  "snapshot_version" integer NOT NULL DEFAULT 1,
  "methodology_version" varchar(60) NOT NULL,
  "confidentiality_class" varchar(30) NOT NULL DEFAULT 'restricted',
  "calculated_at" timestamptz NOT NULL DEFAULT (now()),
  "source_digest" bytea NOT NULL,
  CONSTRAINT "ck_hr_hc_metric_snapshot_values" CHECK (snapshot_version > 0 AND period_end >= period_start AND (population_count IS NULL OR population_count >= 0))
);

CREATE TABLE "hr_workforce_plans" (
  "id" uuid PRIMARY KEY,
  "plan_code" varchar(100) NOT NULL,
  "plan_version" integer NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "period_start" date NOT NULL,
  "period_end" date NOT NULL,
  "scenario_code" varchar(40) NOT NULL DEFAULT 'baseline',
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "finance_budget_reference_id" uuid,
  "created_by_principal_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_workforce_plan_period" CHECK (plan_version > 0 AND period_end >= period_start),
  CONSTRAINT "ck_hr_workforce_plan_status" CHECK (status IN ('draft','review','approved','superseded','cancelled')),
  CONSTRAINT "ck_hr_workforce_plan_approval" CHECK (status NOT IN ('approved','superseded') OR (approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL))
);

CREATE TABLE "hr_workforce_plan_lines" (
  "id" uuid PRIMARY KEY,
  "workforce_plan_id" uuid NOT NULL,
  "line_number" integer NOT NULL,
  "organization_unit_reference_id" uuid NOT NULL,
  "job_id" uuid NOT NULL,
  "work_location_reference_id" uuid,
  "worker_category" hr_worker_category_enum NOT NULL,
  "current_fte" numeric(12,3) NOT NULL,
  "required_fte" numeric(12,3) NOT NULL,
  "estimated_cost" numeric(20,6),
  "currency_code" char(3),
  "needed_from" date NOT NULL,
  "needed_until" date,
  "action_code" varchar(40) NOT NULL,
  "rationale" text NOT NULL,
  CONSTRAINT "ck_hr_workforce_line_capacity" CHECK (line_number > 0 AND current_fte >= 0 AND required_fte >= 0),
  CONSTRAINT "ck_hr_workforce_line_cost" CHECK (estimated_cost IS NULL OR (estimated_cost >= 0 AND currency_code IS NOT NULL)),
  CONSTRAINT "ck_hr_workforce_line_dates" CHECK (needed_until IS NULL OR needed_until >= needed_from),
  CONSTRAINT "ck_hr_workforce_line_action" CHECK (action_code IN ('recruit','develop','redeploy','retain','reduce'))
);

CREATE TABLE "hr_competencies" (
  "id" uuid PRIMARY KEY,
  "competency_code" varchar(100) NOT NULL,
  "competency_version" integer NOT NULL,
  "competency_name" varchar(180) NOT NULL,
  "category_code" varchar(60) NOT NULL,
  "proficiency_scale" jsonb NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_competency_version" CHECK (competency_version > 0)
);

CREATE TABLE "hr_job_competencies" (
  "id" uuid PRIMARY KEY,
  "job_id" uuid NOT NULL,
  "competency_id" uuid NOT NULL,
  "required_level" smallint NOT NULL,
  "is_mandatory" boolean NOT NULL DEFAULT true,
  "effective_from" date NOT NULL,
  "effective_to" date,
  CONSTRAINT "ck_hr_job_competency_level" CHECK (required_level BETWEEN 1 AND 5),
  CONSTRAINT "ck_hr_job_competency_dates" CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE TABLE "hr_worker_competency_assessments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "competency_id" uuid NOT NULL,
  "assessed_level" smallint NOT NULL,
  "assessment_method" varchar(60) NOT NULL,
  "assessed_at" timestamptz NOT NULL,
  "assessed_by_principal_id" uuid NOT NULL,
  "expires_on" date,
  "evidence_document_version_reference_id" uuid,
  "evidence_digest" bytea,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_worker_competency_level" CHECK (assessed_level BETWEEN 1 AND 5),
  CONSTRAINT "ck_hr_competency_evidence" CHECK ((evidence_document_version_reference_id IS NULL) = (evidence_digest IS NULL))
);

CREATE TABLE "hr_job_postings" (
  "id" uuid PRIMARY KEY,
  "requisition_id" uuid NOT NULL,
  "channel_code" varchar(80) NOT NULL,
  "posting_version" integer NOT NULL,
  "external_posting_reference" varchar(200),
  "content_document_version_reference_id" uuid NOT NULL,
  "content_digest" bytea NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "published_at" timestamptz,
  "closes_at" timestamptz,
  "created_by_principal_id" uuid NOT NULL,
  CONSTRAINT "ck_hr_posting_version_dates" CHECK (posting_version > 0 AND (closes_at IS NULL OR published_at IS NULL OR closes_at > published_at)),
  CONSTRAINT "ck_hr_posting_status" CHECK (status IN ('draft','published','closed','withdrawn') AND (status <> 'published' OR published_at IS NOT NULL))
);

CREATE TABLE "hr_application_assessments" (
  "id" uuid PRIMARY KEY,
  "application_id" uuid NOT NULL,
  "assessment_code" varchar(100) NOT NULL,
  "attempt_number" integer NOT NULL DEFAULT 1,
  "rubric_document_version_reference_id" uuid NOT NULL,
  "rubric_digest" bytea NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'planned',
  "assessor_principal_id" uuid,
  "scheduled_at" timestamptz,
  "completed_at" timestamptz,
  "normalized_score" numeric(8,3),
  "recommendation_code" varchar(40),
  "evidence_document_version_reference_id" uuid,
  "evidence_digest" bytea,
  CONSTRAINT "ck_hr_assessment_score" CHECK (attempt_number > 0 AND (normalized_score IS NULL OR normalized_score BETWEEN 0 AND 100)),
  CONSTRAINT "ck_hr_assessment_status" CHECK (status IN ('planned','in_progress','completed','cancelled')),
  CONSTRAINT "ck_hr_assessment_completed" CHECK (status <> 'completed' OR (completed_at IS NOT NULL AND assessor_principal_id IS NOT NULL)),
  CONSTRAINT "ck_hr_assessment_evidence" CHECK ((evidence_document_version_reference_id IS NULL) = (evidence_digest IS NULL))
);

CREATE TABLE "hr_offboarding_plans" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "plan_number" varchar(120) UNIQUE NOT NULL,
  "reason_code" varchar(80) NOT NULL,
  "notice_date" date NOT NULL,
  "last_working_date" date NOT NULL,
  "termination_effective_date" date NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "exit_interview_document_version_reference_id" uuid,
  "final_payroll_run_id" uuid,
  "finance_clearance_reference_id" uuid,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "completed_at" timestamptz,
  "created_by_principal_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_offboarding_dates" CHECK (last_working_date >= notice_date AND termination_effective_date >= last_working_date),
  CONSTRAINT "ck_hr_offboarding_status" CHECK (status IN ('draft','approved','in_progress','completed','cancelled')),
  CONSTRAINT "ck_hr_offboarding_approval" CHECK (status NOT IN ('approved','in_progress','completed') OR (approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL)),
  CONSTRAINT "ck_hr_offboarding_completed" CHECK (status <> 'completed' OR completed_at IS NOT NULL)
);

CREATE TABLE "hr_offboarding_tasks" (
  "id" uuid PRIMARY KEY,
  "offboarding_plan_id" uuid NOT NULL,
  "task_code" varchar(100) NOT NULL,
  "task_name" varchar(180) NOT NULL,
  "owner_management" varchar(80) NOT NULL,
  "assigned_principal_id" uuid,
  "status" varchar(30) NOT NULL DEFAULT 'pending',
  "is_mandatory" boolean NOT NULL DEFAULT true,
  "due_at" timestamptz,
  "completed_at" timestamptz,
  "owner_confirmation_reference_id" uuid,
  "waiver_reason" text,
  "waived_by_principal_id" uuid,
  CONSTRAINT "ck_hr_offboarding_task_status" CHECK (status IN ('pending','in_progress','completed','waived')),
  CONSTRAINT "ck_hr_offboarding_task_evidence" CHECK (status <> 'completed' OR (completed_at IS NOT NULL AND owner_confirmation_reference_id IS NOT NULL)),
  CONSTRAINT "ck_hr_offboarding_task_waiver" CHECK (status <> 'waived' OR (waiver_reason IS NOT NULL AND waived_by_principal_id IS NOT NULL))
);

CREATE TABLE "hr_payroll_groups" (
  "id" uuid PRIMARY KEY,
  "group_code" varchar(80) NOT NULL,
  "group_name" varchar(180) NOT NULL,
  "employing_legal_entity_reference_id" uuid NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "currency_code" char(3) NOT NULL,
  "currency_scale" smallint NOT NULL,
  "pay_frequency" varchar(30) NOT NULL,
  "timezone" varchar(80) NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_payroll_currency_scale" CHECK (currency_scale BETWEEN 0 AND 6)
);

CREATE TABLE "hr_payroll_rule_sets" (
  "id" uuid PRIMARY KEY,
  "rule_code" varchar(100) NOT NULL,
  "rule_version" varchar(60) NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "artifact_document_version_reference_id" uuid NOT NULL,
  "artifact_digest" bytea NOT NULL,
  "legal_basis_document_version_reference_id" uuid NOT NULL,
  "approved_by_principal_id" uuid NOT NULL,
  "approved_at" timestamptz NOT NULL,
  CONSTRAINT "ck_hr_payroll_rule_dates" CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE TABLE "hr_leave_ledger_entries" (
  "id" uuid PRIMARY KEY,
  "entitlement_id" uuid NOT NULL,
  "entry_sequence" bigint NOT NULL,
  "entry_type" varchar(30) NOT NULL,
  "quantity_delta" numeric(12,4) NOT NULL,
  "effective_date" date NOT NULL,
  "adjustment_id" uuid,
  "reverses_entry_id" uuid UNIQUE,
  "source_management" varchar(80) NOT NULL,
  "source_event_id" varchar(160) NOT NULL,
  "source_line_key" varchar(100) NOT NULL,
  "source_digest" bytea NOT NULL,
  "reason_code" varchar(80) NOT NULL,
  "actor_principal_id" uuid NOT NULL,
  "recorded_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_leave_entry_values" CHECK (entry_sequence > 0 AND quantity_delta <> 0),
  CONSTRAINT "ck_hr_leave_entry_type" CHECK (entry_type IN ('opening','accrual','usage','adjustment','expiry','reversal')),
  CONSTRAINT "ck_hr_leave_entry_sign" CHECK ((entry_type IN ('opening','accrual') AND quantity_delta > 0) OR (entry_type IN ('usage','expiry') AND quantity_delta < 0) OR entry_type IN ('adjustment','reversal')),
  CONSTRAINT "ck_hr_leave_reversal_reference" CHECK ((entry_type = 'reversal') = (reverses_entry_id IS NOT NULL)),
  CONSTRAINT "ck_hr_leave_not_self_reversal" CHECK (reverses_entry_id IS NULL OR reverses_entry_id <> id),
  CONSTRAINT "ck_hr_leave_adjustment_reference" CHECK ((entry_type = 'adjustment') = (adjustment_id IS NOT NULL))
);

CREATE TABLE "hr_development_plans" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "plan_code" varchar(100) NOT NULL,
  "plan_version" integer NOT NULL,
  "target_job_id" uuid,
  "start_date" date NOT NULL,
  "target_date" date NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "mentor_worker_id" uuid,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_development_plan_dates" CHECK (plan_version > 0 AND target_date >= start_date),
  CONSTRAINT "ck_hr_development_mentor" CHECK (mentor_worker_id IS NULL OR mentor_worker_id <> worker_id),
  CONSTRAINT "ck_hr_development_plan_status" CHECK (status IN ('draft','active','completed','cancelled'))
);

CREATE TABLE "hr_development_actions" (
  "id" uuid PRIMARY KEY,
  "development_plan_id" uuid NOT NULL,
  "action_number" integer NOT NULL,
  "competency_id" uuid,
  "course_id" uuid,
  "action_type" varchar(40) NOT NULL,
  "objective" text NOT NULL,
  "target_level" smallint,
  "due_date" date NOT NULL,
  "completed_at" timestamptz,
  "evidence_document_version_reference_id" uuid,
  CONSTRAINT "ck_hr_development_action_values" CHECK (action_number > 0 AND (target_level IS NULL OR target_level BETWEEN 1 AND 5)),
  CONSTRAINT "ck_hr_development_action_type" CHECK (action_type IN ('course','coaching','mentoring','assignment','self_study'))
);

CREATE TABLE "hr_collective_agreements" (
  "id" uuid PRIMARY KEY,
  "agreement_code" varchar(100) NOT NULL,
  "agreement_version" integer NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "representative_party_reference_id" uuid NOT NULL,
  "jurisdiction_code" varchar(80) NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "signed_at" timestamptz NOT NULL,
  "document_version_reference_id" uuid NOT NULL,
  "content_digest" bytea NOT NULL,
  CONSTRAINT "ck_hr_collective_agreement_dates" CHECK (agreement_version > 0 AND (effective_to IS NULL OR effective_to >= effective_from))
);

CREATE TABLE "hr_policies" (
  "id" uuid PRIMARY KEY,
  "policy_code" varchar(100) NOT NULL,
  "policy_version" integer NOT NULL,
  "policy_name" varchar(180) NOT NULL,
  "category_code" varchar(60) NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "document_version_reference_id" uuid NOT NULL,
  "content_digest" bytea NOT NULL,
  "approved_by_principal_id" uuid NOT NULL,
  "approved_at" timestamptz NOT NULL,
  CONSTRAINT "ck_hr_policy_dates" CHECK (policy_version > 0 AND (effective_to IS NULL OR effective_to >= effective_from))
);

CREATE TABLE "hr_policy_acknowledgements" (
  "id" uuid PRIMARY KEY,
  "policy_id" uuid NOT NULL,
  "worker_id" uuid NOT NULL,
  "assigned_at" timestamptz NOT NULL DEFAULT (now()),
  "due_at" timestamptz,
  "acknowledged_at" timestamptz,
  "acknowledged_by_principal_id" uuid,
  "evidence_document_version_reference_id" uuid,
  CONSTRAINT "ck_hr_policy_ack_actor" CHECK ((acknowledged_at IS NULL) = (acknowledged_by_principal_id IS NULL)),
  CONSTRAINT "ck_hr_policy_ack_dates" CHECK (acknowledged_at IS NULL OR acknowledged_at >= assigned_at)
);

CREATE TABLE "hr_service_requests" (
  "id" uuid PRIMARY KEY,
  "request_number" varchar(120) UNIQUE NOT NULL,
  "worker_id" uuid NOT NULL,
  "request_type" varchar(60) NOT NULL,
  "requested_by_principal_id" uuid NOT NULL,
  "assigned_principal_id" uuid,
  "status" varchar(30) NOT NULL DEFAULT 'open',
  "restricted_details_document_version_reference_id" uuid NOT NULL,
  "opened_at" timestamptz NOT NULL DEFAULT (now()),
  "due_at" timestamptz,
  "resolved_at" timestamptz,
  "resolution_document_version_reference_id" uuid,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_service_request_status" CHECK (status IN ('open','in_progress','waiting','resolved','closed','cancelled')),
  CONSTRAINT "ck_hr_service_request_resolution" CHECK (status NOT IN ('resolved','closed') OR (resolved_at IS NOT NULL AND resolution_document_version_reference_id IS NOT NULL)),
  CONSTRAINT "ck_hr_service_request_dates" CHECK (resolved_at IS NULL OR resolved_at >= opened_at)
);

CREATE TABLE "hr_service_request_events" (
  "id" uuid PRIMARY KEY,
  "service_request_id" uuid NOT NULL,
  "event_type" varchar(60) NOT NULL,
  "decision" hr_approval_decision_enum,
  "actor_principal_id" uuid NOT NULL,
  "occurred_at" timestamptz NOT NULL DEFAULT (now()),
  "evidence_document_version_reference_id" uuid,
  "idempotency_key" varchar(160) UNIQUE NOT NULL
);

CREATE TABLE "hr_engagement_surveys" (
  "id" uuid PRIMARY KEY,
  "survey_code" varchar(100) NOT NULL,
  "survey_version" integer NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "topic_code" varchar(60) NOT NULL,
  "opens_at" timestamptz NOT NULL,
  "closes_at" timestamptz NOT NULL,
  "questionnaire_document_version_reference_id" uuid NOT NULL,
  "questionnaire_digest" bytea NOT NULL,
  "minimum_reporting_population" integer NOT NULL,
  CONSTRAINT "ck_hr_survey_values" CHECK (survey_version > 0 AND closes_at > opens_at AND minimum_reporting_population >= 5)
);

CREATE TABLE "hr_engagement_survey_results" (
  "id" uuid PRIMARY KEY,
  "survey_id" uuid NOT NULL,
  "organization_unit_reference_id" uuid,
  "question_code" varchar(100) NOT NULL,
  "result_version" integer NOT NULL,
  "respondent_count" integer NOT NULL,
  "score_value" numeric(12,6) NOT NULL,
  "scale_minimum" numeric(12,6) NOT NULL,
  "scale_maximum" numeric(12,6) NOT NULL,
  "source_digest" bytea NOT NULL,
  "calculated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_survey_result_population" CHECK (result_version > 0 AND respondent_count >= 5),
  CONSTRAINT "ck_hr_survey_result_scale" CHECK (scale_maximum > scale_minimum AND score_value BETWEEN scale_minimum AND scale_maximum)
);

CREATE TABLE "hr_access_requests" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "lifecycle_event_id" bigint NOT NULL,
  "request_kind" varchar(30) NOT NULL,
  "requested_scope_type" varchar(40) NOT NULL,
  "requested_scope_reference_id" uuid NOT NULL,
  "approved_catalog_reference_id" uuid,
  "iam_request_reference_id" uuid,
  "iam_decision_event_reference_id" uuid,
  "status" varchar(30) NOT NULL DEFAULT 'pending',
  "requested_by_principal_id" uuid NOT NULL,
  "requested_at" timestamptz NOT NULL DEFAULT (now()),
  "last_reconciled_at" timestamptz,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_access_request_kind" CHECK (request_kind IN ('provision','change','revoke')),
  CONSTRAINT "ck_hr_access_request_status" CHECK (status IN ('pending','submitted','approved','rejected','applied','failed','cancelled')),
  CONSTRAINT "ck_hr_access_request_receipt" CHECK (status NOT IN ('submitted','approved','rejected','applied') OR iam_request_reference_id IS NOT NULL),
  CONSTRAINT "ck_hr_access_request_decision" CHECK (status NOT IN ('approved','rejected','applied') OR iam_decision_event_reference_id IS NOT NULL),
  CONSTRAINT "ck_hr_access_catalog" CHECK (request_kind = 'revoke' OR approved_catalog_reference_id IS NOT NULL)
);

CREATE TABLE "hr_outbox_events" (
  "id" uuid PRIMARY KEY,
  "aggregate_type" varchar(80) NOT NULL,
  "aggregate_id" uuid NOT NULL,
  "aggregate_version" bigint NOT NULL,
  "event_type" varchar(100) NOT NULL,
  "schema_version" integer NOT NULL,
  "payload" jsonb NOT NULL,
  "payload_digest" bytea NOT NULL,
  "occurred_at" timestamptz NOT NULL DEFAULT (now()),
  "available_at" timestamptz NOT NULL DEFAULT (now()),
  "published_at" timestamptz,
  "attempt_count" integer NOT NULL DEFAULT 0,
  "last_error_code" varchar(80),
  CONSTRAINT "ck_hr_outbox_values" CHECK (aggregate_version > 0 AND schema_version > 0 AND attempt_count >= 0)
);

CREATE TABLE "hr_inbox_events" (
  "id" uuid PRIMARY KEY,
  "source_management" varchar(80) NOT NULL,
  "source_event_id" varchar(160) NOT NULL,
  "event_type" varchar(100) NOT NULL,
  "schema_version" integer NOT NULL,
  "source_aggregate_reference" varchar(160) NOT NULL,
  "source_aggregate_version" bigint NOT NULL,
  "payload_digest" bytea NOT NULL,
  "received_at" timestamptz NOT NULL DEFAULT (now()),
  "processed_at" timestamptz,
  "status" varchar(30) NOT NULL DEFAULT 'pending',
  "failure_code" varchar(80),
  CONSTRAINT "ck_hr_inbox_version" CHECK (schema_version > 0 AND source_aggregate_version > 0),
  CONSTRAINT "ck_hr_inbox_state" CHECK (status IN ('pending','processed','rejected') AND (status <> 'processed' OR processed_at IS NOT NULL))
);

CREATE TABLE "hr_talent_frameworks" (
  "id" uuid PRIMARY KEY,
  "framework_code" varchar(100) NOT NULL,
  "framework_version" integer NOT NULL,
  "framework_name" varchar(180) NOT NULL,
  "rubric_document_version_reference_id" uuid NOT NULL,
  "rubric_digest" bytea NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date,
  "approved_by_principal_id" uuid NOT NULL,
  "approved_at" timestamptz NOT NULL,
  CONSTRAINT "ck_hr_talent_framework_version" CHECK (framework_version > 0 AND (effective_to IS NULL OR effective_to >= effective_from))
);

CREATE TABLE "hr_talent_dimensions" (
  "id" uuid PRIMARY KEY,
  "framework_id" uuid NOT NULL,
  "dimension_code" varchar(80) NOT NULL,
  "dimension_name" varchar(180) NOT NULL,
  "definition" text NOT NULL,
  "rating_anchors" jsonb NOT NULL,
  "weight_percent" numeric(6,3) NOT NULL,
  "is_required" boolean NOT NULL DEFAULT true,
  CONSTRAINT "ck_hr_talent_dimension_weight" CHECK (weight_percent BETWEEN 0 AND 100)
);

CREATE TABLE "hr_talent_review_cycles" (
  "id" uuid PRIMARY KEY,
  "cycle_code" varchar(100) UNIQUE NOT NULL,
  "framework_id" uuid NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "organization_unit_reference_id" uuid,
  "period_start" date NOT NULL,
  "period_end" date NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "owner_principal_id" uuid NOT NULL,
  "closed_at" timestamptz,
  CONSTRAINT "ck_hr_talent_cycle_period" CHECK (period_end >= period_start),
  CONSTRAINT "ck_hr_talent_cycle_status" CHECK (status IN ('draft','open','calibration','closed','cancelled')),
  CONSTRAINT "ck_hr_talent_cycle_closed" CHECK (status <> 'closed' OR closed_at IS NOT NULL)
);

CREATE TABLE "hr_talent_assessments" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "talent_review_cycle_id" uuid NOT NULL,
  "framework_id" uuid NOT NULL,
  "assessment_version" integer NOT NULL,
  "performance_review_id" uuid,
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "assessed_by_principal_id" uuid NOT NULL,
  "assessed_at" timestamptz NOT NULL DEFAULT (now()),
  "evidence_document_version_reference_id" uuid NOT NULL,
  "evidence_digest" bytea NOT NULL,
  "segment_code" varchar(80),
  "finalized_by_principal_id" uuid,
  "finalized_at" timestamptz,
  CONSTRAINT "ck_hr_talent_assessment_version" CHECK (assessment_version > 0),
  CONSTRAINT "ck_hr_talent_assessment_status" CHECK (status IN ('draft','submitted','finalized','withdrawn')),
  CONSTRAINT "ck_hr_talent_assessment_finalized" CHECK (status <> 'finalized' OR (finalized_at IS NOT NULL AND finalized_by_principal_id IS NOT NULL)),
  CONSTRAINT "ck_hr_talent_assessment_dates" CHECK (finalized_at IS NULL OR finalized_at >= assessed_at)
);

CREATE TABLE "hr_talent_assessment_scores" (
  "id" uuid PRIMARY KEY,
  "assessment_id" uuid NOT NULL,
  "framework_id" uuid NOT NULL,
  "dimension_id" uuid NOT NULL,
  "score" smallint NOT NULL,
  "evidence_document_version_reference_id" uuid NOT NULL,
  "evidence_digest" bytea NOT NULL,
  CONSTRAINT "ck_hr_talent_score" CHECK (score BETWEEN 1 AND 5)
);

CREATE TABLE "hr_talent_calibration_sessions" (
  "id" uuid PRIMARY KEY,
  "talent_review_cycle_id" uuid NOT NULL,
  "session_code" varchar(100) NOT NULL,
  "chair_principal_id" uuid NOT NULL,
  "scheduled_at" timestamptz NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'planned',
  "closed_at" timestamptz,
  "minutes_document_version_reference_id" uuid,
  CONSTRAINT "ck_hr_calibration_status" CHECK (status IN ('planned','open','closed','cancelled')),
  CONSTRAINT "ck_hr_calibration_closed" CHECK (status <> 'closed' OR (closed_at IS NOT NULL AND minutes_document_version_reference_id IS NOT NULL))
);

CREATE TABLE "hr_talent_calibration_decisions" (
  "id" uuid PRIMARY KEY,
  "session_id" uuid NOT NULL,
  "talent_review_cycle_id" uuid NOT NULL,
  "assessment_id" uuid NOT NULL,
  "dimension_id" uuid NOT NULL,
  "original_score" smallint NOT NULL,
  "calibrated_score" smallint NOT NULL,
  "rationale_document_version_reference_id" uuid NOT NULL,
  "rationale_digest" bytea NOT NULL,
  "decided_by_principal_id" uuid NOT NULL,
  "decided_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "ck_hr_calibration_score" CHECK (calibrated_score BETWEEN 1 AND 5)
);

CREATE TABLE "hr_talent_pools" (
  "id" uuid PRIMARY KEY,
  "pool_code" varchar(100) NOT NULL,
  "pool_version" integer NOT NULL,
  "pool_name" varchar(180) NOT NULL,
  "legal_entity_reference_id" uuid NOT NULL,
  "target_job_id" uuid,
  "eligibility_document_version_reference_id" uuid NOT NULL,
  "eligibility_digest" bytea NOT NULL,
  "valid_from" date NOT NULL,
  "valid_to" date,
  "owner_principal_id" uuid NOT NULL,
  CONSTRAINT "ck_hr_talent_pool_dates" CHECK (pool_version > 0 AND (valid_to IS NULL OR valid_to >= valid_from))
);

CREATE TABLE "hr_talent_pool_memberships" (
  "id" uuid PRIMARY KEY,
  "talent_pool_id" uuid NOT NULL,
  "worker_id" uuid NOT NULL,
  "assessment_id" uuid NOT NULL,
  "nominated_by_principal_id" uuid NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'nominated',
  "valid_from" date NOT NULL,
  "valid_to" date,
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  "exit_reason_code" varchar(80),
  CONSTRAINT "ck_hr_pool_membership_dates" CHECK (valid_to IS NULL OR valid_to >= valid_from),
  CONSTRAINT "ck_hr_pool_membership_status" CHECK (status IN ('nominated','active','rejected','exited')),
  CONSTRAINT "ck_hr_pool_membership_approval" CHECK (status NOT IN ('active','exited') OR (approved_by_principal_id IS NOT NULL AND approved_at IS NOT NULL)),
  CONSTRAINT "ck_hr_pool_membership_exit" CHECK (status <> 'exited' OR (valid_to IS NOT NULL AND exit_reason_code IS NOT NULL))
);

CREATE TABLE "hr_career_paths" (
  "id" uuid PRIMARY KEY,
  "path_code" varchar(100) NOT NULL,
  "path_version" integer NOT NULL,
  "path_name" varchar(180) NOT NULL,
  "job_family_id" uuid NOT NULL,
  "valid_from" date NOT NULL,
  "valid_to" date,
  "status" varchar(20) NOT NULL DEFAULT 'draft',
  "created_by_principal_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "approved_by_principal_id" uuid,
  "approved_at" timestamptz,
  CONSTRAINT "ck_hr_career_path_dates" CHECK (path_version > 0 AND (valid_to IS NULL OR valid_to >= valid_from)),
  CONSTRAINT "ck_hr_career_path_status" CHECK (status IN ('draft','approved')),
  CONSTRAINT "ck_hr_career_path_approval" CHECK ((status = 'draft' AND approved_by_principal_id IS NULL AND approved_at IS NULL) OR (status = 'approved' AND approved_by_principal_id IS NOT NULL AND approved_at IS NOT NULL)),
  CONSTRAINT "ck_hr_career_path_separation" CHECK (approved_by_principal_id IS NULL OR approved_by_principal_id <> created_by_principal_id)
);

CREATE TABLE "hr_career_path_transitions" (
  "id" uuid PRIMARY KEY,
  "career_path_id" uuid NOT NULL,
  "from_job_id" uuid NOT NULL,
  "to_job_id" uuid NOT NULL,
  "movement_type" varchar(30) NOT NULL,
  "minimum_tenure_months" integer NOT NULL DEFAULT 0,
  "requirements_document_version_reference_id" uuid NOT NULL,
  "requirements_digest" bytea NOT NULL,
  CONSTRAINT "ck_hr_career_transition_jobs" CHECK (from_job_id <> to_job_id AND minimum_tenure_months >= 0),
  CONSTRAINT "ck_hr_career_transition_type" CHECK (movement_type IN ('promotion','lateral','specialization'))
);

CREATE TABLE "hr_worker_career_aspirations" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "target_job_id" uuid NOT NULL,
  "career_path_transition_id" uuid,
  "preference_rank" integer NOT NULL,
  "recorded_at" timestamptz NOT NULL DEFAULT (now()),
  "recorded_by_principal_id" uuid NOT NULL,
  "target_date" date,
  "mobility_preference" varchar(30) NOT NULL DEFAULT 'unspecified',
  "preferred_location_reference_id" uuid,
  "withdrawn_at" timestamptz,
  CONSTRAINT "ck_hr_career_preference_rank" CHECK (preference_rank > 0),
  CONSTRAINT "ck_hr_career_mobility" CHECK (mobility_preference IN ('unspecified','current_location','open_to_relocation','remote')),
  CONSTRAINT "ck_hr_career_aspiration_dates" CHECK (withdrawn_at IS NULL OR withdrawn_at >= recorded_at)
);

CREATE TABLE "hr_promotion_recommendations" (
  "id" uuid PRIMARY KEY,
  "recommendation_number" varchar(120) UNIQUE NOT NULL,
  "worker_id" uuid NOT NULL,
  "source_assignment_id" uuid NOT NULL,
  "target_position_id" uuid NOT NULL,
  "assessment_id" uuid NOT NULL,
  "supersedes_recommendation_id" uuid,
  "recommended_effective_date" date NOT NULL,
  "rationale_document_version_reference_id" uuid NOT NULL,
  "rationale_digest" bytea NOT NULL,
  "recommended_by_principal_id" uuid NOT NULL,
  "recommended_at" timestamptz NOT NULL DEFAULT (now()),
  "status" varchar(30) NOT NULL DEFAULT 'draft',
  "approval_decision_id" uuid,
  "applied_assignment_id" uuid UNIQUE,
  "applied_lifecycle_event_id" bigint UNIQUE,
  "applied_at" timestamptz,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_promotion_status" CHECK (status IN ('draft','submitted','approved','rejected','returned','cancelled','applied')),
  CONSTRAINT "ck_hr_promotion_approval_reference" CHECK ((status IN ('approved','applied')) = (approval_decision_id IS NOT NULL)),
  CONSTRAINT "ck_hr_promotion_application" CHECK ((status = 'applied') = (applied_assignment_id IS NOT NULL AND applied_lifecycle_event_id IS NOT NULL AND applied_at IS NOT NULL)),
  CONSTRAINT "ck_hr_promotion_no_partial_application" CHECK (status = 'applied' OR (applied_assignment_id IS NULL AND applied_lifecycle_event_id IS NULL AND applied_at IS NULL)),
  CONSTRAINT "ck_hr_promotion_new_assignment" CHECK (applied_assignment_id IS NULL OR applied_assignment_id <> source_assignment_id),
  CONSTRAINT "ck_hr_promotion_not_self_revision" CHECK (supersedes_recommendation_id IS NULL OR supersedes_recommendation_id <> id)
);

CREATE TABLE "hr_promotion_decisions" (
  "id" uuid PRIMARY KEY,
  "recommendation_id" uuid NOT NULL,
  "step_number" integer NOT NULL,
  "decision" hr_approval_decision_enum NOT NULL,
  "decided_by_principal_id" uuid NOT NULL,
  "decided_at" timestamptz NOT NULL DEFAULT (now()),
  "rationale_document_version_reference_id" uuid NOT NULL,
  "rationale_digest" bytea NOT NULL,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_promotion_decision_step" CHECK (step_number > 0)
);

CREATE TABLE "hr_learning_course_competencies" (
  "course_id" uuid NOT NULL,
  "competency_id" uuid NOT NULL,
  "target_level" smallint NOT NULL,
  CONSTRAINT "ck_hr_course_competency_level" CHECK (target_level BETWEEN 1 AND 5),
  PRIMARY KEY ("course_id", "competency_id")
);

CREATE TABLE "hr_performance_goal_updates" (
  "id" uuid PRIMARY KEY,
  "goal_id" uuid NOT NULL,
  "observed_at" timestamptz NOT NULL,
  "actual_value" numeric(24,8) NOT NULL,
  "evidence_document_version_reference_id" uuid NOT NULL,
  "evidence_digest" bytea NOT NULL,
  "recorded_by_principal_id" uuid NOT NULL,
  "recorded_at" timestamptz NOT NULL DEFAULT (now()),
  "supersedes_update_id" uuid UNIQUE,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  CONSTRAINT "ck_hr_goal_update_revision" CHECK (supersedes_update_id IS NULL OR supersedes_update_id <> id)
);

CREATE TABLE "hr_performance_review_goal_scores" (
  "id" uuid PRIMARY KEY,
  "performance_review_id" uuid NOT NULL,
  "goal_id" uuid NOT NULL,
  "worker_id" uuid NOT NULL,
  "performance_cycle_id" uuid NOT NULL,
  "rating" smallint NOT NULL,
  "weight_percent_snapshot" numeric(6,3) NOT NULL,
  "rubric_document_version_reference_id" uuid NOT NULL,
  "evidence_document_version_reference_id" uuid NOT NULL,
  CONSTRAINT "ck_hr_review_goal_score" CHECK (rating BETWEEN 1 AND 5 AND weight_percent_snapshot BETWEEN 0 AND 100)
);

CREATE INDEX "ix_hr_job_family_level" ON "hr_jobs" ("job_family_id", "job_level_code", "is_active");

CREATE INDEX "ix_hr_position_org_period" ON "hr_positions" ("organization_unit_reference_id", "valid_from");

CREATE INDEX "ix_hr_worker_person_status" ON "hr_workers" ("person_party_reference_id", "status");

CREATE INDEX "ix_hr_contract_worker_period" ON "hr_employment_contracts" ("worker_id", "effective_from");

CREATE UNIQUE INDEX "uq_hr_worker_assignment_period" ON "hr_worker_assignments" ("worker_id", "position_id", "assignment_type", "valid_from");

CREATE UNIQUE INDEX "uq_hr_assignment_worker_pair" ON "hr_worker_assignments" ("id", "worker_id");

CREATE UNIQUE INDEX ON "hr_worker_assignments" ("id", "worker_id", "position_id");

CREATE INDEX "ix_hr_assignment_position_period" ON "hr_worker_assignments" ("position_id", "valid_from", "valid_to");

CREATE INDEX "ix_hr_assignment_manager" ON "hr_worker_assignments" ("manager_worker_id", "valid_from");

CREATE INDEX "ix_hr_worker_lifecycle_history" ON "hr_worker_lifecycle_events" ("worker_id", "effective_date", "occurred_at");

CREATE UNIQUE INDEX ON "hr_worker_lifecycle_events" ("id", "worker_id");

CREATE INDEX "ix_hr_worker_document_expiry" ON "hr_worker_documents" ("worker_id", "document_type", "expires_on");

CREATE UNIQUE INDEX "uq_hr_emergency_contact_priority" ON "hr_emergency_contacts" ("worker_id", "priority_order");

CREATE INDEX "ix_hr_requisition_queue" ON "hr_requisitions" ("status", "recruiter_principal_id", "opened_at");

CREATE INDEX "ix_hr_candidate_person" ON "hr_candidates" ("person_party_reference_id", "candidate_status");

CREATE UNIQUE INDEX "uq_hr_candidate_requisition_application" ON "hr_applications" ("candidate_id", "requisition_id");

CREATE INDEX "ix_hr_application_requisition_status" ON "hr_applications" ("requisition_id", "status", "applied_at");

CREATE INDEX "ix_hr_application_stage_history" ON "hr_application_stage_history" ("application_id", "changed_at");

CREATE UNIQUE INDEX "uq_hr_interview_round" ON "hr_interviews" ("application_id", "interview_round", "interview_type");

CREATE UNIQUE INDEX "uq_hr_interview_panelist" ON "hr_interview_panelists" ("interview_id", "interviewer_principal_id");

CREATE UNIQUE INDEX "uq_hr_interview_panelist_identity_pair" ON "hr_interview_panelists" ("id", "interviewer_principal_id");

CREATE UNIQUE INDEX "uq_hr_interview_feedback_version" ON "hr_interview_feedback" ("interview_id", "interviewer_principal_id", "feedback_version");

CREATE UNIQUE INDEX "uq_hr_application_offer_version" ON "hr_offers" ("application_id", "offer_version");

CREATE UNIQUE INDEX "uq_hr_onboarding_plan_version" ON "hr_onboarding_plans" ("worker_id", "plan_code", "plan_version");

CREATE UNIQUE INDEX "uq_hr_onboarding_task" ON "hr_onboarding_tasks" ("onboarding_plan_id", "task_code");

CREATE UNIQUE INDEX "uq_hr_compensation_band_version" ON "hr_compensation_bands" ("band_code", "band_version", "jurisdiction_code");

CREATE UNIQUE INDEX "uq_hr_worker_compensation_period" ON "hr_compensation_records" ("worker_id", "assignment_id", "component_id", "effective_from");

CREATE UNIQUE INDEX "uq_hr_benefit_plan_version" ON "hr_benefit_plans" ("plan_code", "plan_version", "jurisdiction_code");

CREATE UNIQUE INDEX "uq_hr_worker_benefit_enrollment" ON "hr_benefit_enrollments" ("worker_id", "benefit_plan_id", "effective_from");

CREATE UNIQUE INDEX "uq_hr_payroll_profile_period" ON "hr_payroll_profiles" ("worker_id", "payroll_group_id", "effective_from");

CREATE UNIQUE INDEX "uq_hr_payroll_profile_worker_group" ON "hr_payroll_profiles" ("id", "worker_id", "payroll_group_id", "currency_code");

CREATE UNIQUE INDEX "uq_hr_pay_component_version" ON "hr_pay_components" ("component_code", "component_version", "jurisdiction_code");

CREATE UNIQUE INDEX ON "hr_pay_components" ("id", "component_code");

CREATE UNIQUE INDEX ON "hr_pay_components" ("id", "component_code", "component_type");

CREATE UNIQUE INDEX "uq_hr_payroll_period_code" ON "hr_payroll_periods" ("payroll_group_id", "period_code");

CREATE UNIQUE INDEX "uq_hr_payroll_period_scope" ON "hr_payroll_periods" ("id", "payroll_group_id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id");

CREATE UNIQUE INDEX "uq_hr_payroll_run_number" ON "hr_payroll_runs" ("payroll_period_id", "run_number");

CREATE UNIQUE INDEX "uq_hr_payroll_run_group" ON "hr_payroll_runs" ("id", "payroll_group_id", "currency_code");

CREATE UNIQUE INDEX "uq_hr_payroll_worker_result" ON "hr_payroll_results" ("payroll_run_id", "worker_id");

CREATE UNIQUE INDEX "uq_hr_payroll_result_line_sequence" ON "hr_payroll_result_lines" ("payroll_result_id", "line_sequence");

CREATE INDEX "ix_hr_payroll_adjustment_worker_period" ON "hr_payroll_adjustments" ("worker_id", "payroll_period_id", "component_id");

CREATE UNIQUE INDEX "uq_hr_leave_policy_version" ON "hr_leave_policies" ("policy_code", "policy_version", "jurisdiction_code");

CREATE UNIQUE INDEX ON "hr_leave_policies" ("id", "leave_type_code");

CREATE UNIQUE INDEX "uq_hr_leave_entitlement_period" ON "hr_leave_entitlements" ("worker_id", "leave_policy_id", "entitlement_period_start");

CREATE UNIQUE INDEX "uq_hr_leave_adjustment_account" ON "hr_leave_adjustments" ("id", "entitlement_id");

CREATE UNIQUE INDEX "uq_hr_worker_performance_goal" ON "hr_performance_goals" ("worker_id", "performance_cycle_id", "goal_code");

CREATE UNIQUE INDEX "uq_hr_goal_worker_cycle_pair" ON "hr_performance_goals" ("id", "worker_id", "performance_cycle_id");

CREATE UNIQUE INDEX "uq_hr_performance_review_version" ON "hr_performance_reviews" ("worker_id", "performance_cycle_id", "reviewer_principal_id", "reviewer_role", "review_version");

CREATE UNIQUE INDEX ON "hr_performance_reviews" ("id", "worker_id");

CREATE UNIQUE INDEX ON "hr_performance_reviews" ("id", "worker_id", "performance_cycle_id");

CREATE UNIQUE INDEX "uq_hr_learning_course_version" ON "hr_learning_courses" ("course_code", "course_version");

CREATE UNIQUE INDEX "uq_hr_worker_course_assignment" ON "hr_learning_enrollments" ("worker_id", "course_id", "assigned_at");

CREATE INDEX "ix_hr_worker_qualification_expiry" ON "hr_worker_qualifications" ("worker_id", "qualification_type", "expires_on");

CREATE UNIQUE INDEX "uq_hr_succession_plan_version" ON "hr_succession_plans" ("position_id", "plan_code", "plan_version");

CREATE UNIQUE INDEX "uq_hr_succession_candidate" ON "hr_succession_candidates" ("succession_plan_id", "candidate_worker_id");

CREATE INDEX "ix_hr_employee_case_queue" ON "hr_employee_relations_cases" ("status", "assigned_investigator_principal_id", "due_at");

CREATE INDEX "ix_hr_employee_case_event_history" ON "hr_employee_relations_events" ("case_id", "occurred_at");

CREATE UNIQUE INDEX "uq_hr_hc_metric_definition_version" ON "hr_human_capital_metric_definitions" ("metric_code", "metric_version");

CREATE UNIQUE INDEX "uq_hr_hc_metric_snapshot_scope" ON "hr_human_capital_metric_snapshots" ("metric_definition_id", "period_start", "period_end", (COALESCE(organization_unit_reference_id::text, '')), (COALESCE(jurisdiction_code, '')), "methodology_version", "snapshot_version");

CREATE UNIQUE INDEX ON "hr_workforce_plans" ("legal_entity_reference_id", "plan_code", "plan_version");

CREATE UNIQUE INDEX ON "hr_workforce_plan_lines" ("workforce_plan_id", "line_number");

CREATE UNIQUE INDEX ON "hr_competencies" ("competency_code", "competency_version");

CREATE UNIQUE INDEX ON "hr_job_competencies" ("job_id", "competency_id", "effective_from");

CREATE INDEX ON "hr_worker_competency_assessments" ("worker_id", "competency_id", "assessed_at");

CREATE UNIQUE INDEX ON "hr_job_postings" ("requisition_id", "channel_code", "posting_version");

CREATE UNIQUE INDEX ON "hr_job_postings" ("id", "requisition_id");

CREATE UNIQUE INDEX ON "hr_application_assessments" ("application_id", "assessment_code", "attempt_number");

CREATE INDEX ON "hr_offboarding_plans" ("worker_id", "termination_effective_date");

CREATE UNIQUE INDEX ON "hr_offboarding_tasks" ("offboarding_plan_id", "task_code");

CREATE UNIQUE INDEX ON "hr_payroll_groups" ("employing_legal_entity_reference_id", "group_code");

CREATE UNIQUE INDEX "uq_hr_payroll_group_scope" ON "hr_payroll_groups" ("id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id");

CREATE UNIQUE INDEX ON "hr_payroll_rule_sets" ("rule_code", "rule_version", "jurisdiction_code");

CREATE UNIQUE INDEX ON "hr_payroll_rule_sets" ("id", "jurisdiction_code");

CREATE UNIQUE INDEX ON "hr_leave_ledger_entries" ("entitlement_id", "entry_sequence");

CREATE UNIQUE INDEX ON "hr_leave_ledger_entries" ("id", "entitlement_id");

CREATE UNIQUE INDEX "uq_hr_leave_source_event_line" ON "hr_leave_ledger_entries" ("source_management", "source_event_id", "source_line_key");

CREATE UNIQUE INDEX ON "hr_leave_ledger_entries" ("adjustment_id");

CREATE UNIQUE INDEX ON "hr_development_plans" ("worker_id", "plan_code", "plan_version");

CREATE UNIQUE INDEX ON "hr_development_plans" ("id", "worker_id");

CREATE UNIQUE INDEX ON "hr_development_actions" ("development_plan_id", "action_number");

CREATE UNIQUE INDEX ON "hr_collective_agreements" ("legal_entity_reference_id", "agreement_code", "agreement_version");

CREATE UNIQUE INDEX ON "hr_policies" ("legal_entity_reference_id", "policy_code", "policy_version");

CREATE UNIQUE INDEX ON "hr_policy_acknowledgements" ("policy_id", "worker_id");

CREATE INDEX ON "hr_service_requests" ("status", "assigned_principal_id", "due_at");

CREATE INDEX ON "hr_service_request_events" ("service_request_id", "occurred_at");

CREATE UNIQUE INDEX ON "hr_engagement_surveys" ("legal_entity_reference_id", "survey_code", "survey_version");

CREATE UNIQUE INDEX "uq_hr_survey_result_scope" ON "hr_engagement_survey_results" ("survey_id", (COALESCE(organization_unit_reference_id::text, '')), "question_code", "result_version");

CREATE INDEX ON "hr_access_requests" ("status", "requested_at");

CREATE UNIQUE INDEX "uq_hr_outbox_aggregate_event" ON "hr_outbox_events" ("aggregate_type", "aggregate_id", "aggregate_version", "event_type");

CREATE INDEX "ix_hr_outbox_delivery" ON "hr_outbox_events" ("published_at", "available_at");

CREATE UNIQUE INDEX ON "hr_inbox_events" ("source_management", "source_event_id");

CREATE INDEX "ix_hr_inbox_processing" ON "hr_inbox_events" ("status", "received_at");

CREATE UNIQUE INDEX ON "hr_talent_frameworks" ("framework_code", "framework_version");

CREATE UNIQUE INDEX ON "hr_talent_dimensions" ("framework_id", "dimension_code");

CREATE UNIQUE INDEX ON "hr_talent_dimensions" ("id", "framework_id");

CREATE UNIQUE INDEX ON "hr_talent_review_cycles" ("id", "framework_id");

CREATE UNIQUE INDEX ON "hr_talent_assessments" ("worker_id", "talent_review_cycle_id", "assessment_version");

CREATE UNIQUE INDEX ON "hr_talent_assessments" ("id", "worker_id");

CREATE UNIQUE INDEX ON "hr_talent_assessments" ("id", "framework_id");

CREATE UNIQUE INDEX ON "hr_talent_assessments" ("id", "talent_review_cycle_id");

CREATE UNIQUE INDEX ON "hr_talent_assessment_scores" ("assessment_id", "dimension_id");

CREATE UNIQUE INDEX ON "hr_talent_assessment_scores" ("assessment_id", "dimension_id", "score");

CREATE UNIQUE INDEX ON "hr_talent_calibration_sessions" ("talent_review_cycle_id", "session_code");

CREATE UNIQUE INDEX ON "hr_talent_calibration_sessions" ("id", "talent_review_cycle_id");

CREATE UNIQUE INDEX ON "hr_talent_calibration_decisions" ("session_id", "assessment_id", "dimension_id");

CREATE UNIQUE INDEX ON "hr_talent_pools" ("legal_entity_reference_id", "pool_code", "pool_version");

CREATE UNIQUE INDEX ON "hr_talent_pool_memberships" ("talent_pool_id", "worker_id", "valid_from");

CREATE UNIQUE INDEX ON "hr_career_paths" ("path_code", "path_version");

CREATE UNIQUE INDEX ON "hr_career_path_transitions" ("career_path_id", "from_job_id", "to_job_id");

CREATE UNIQUE INDEX ON "hr_career_path_transitions" ("id", "to_job_id");

CREATE UNIQUE INDEX ON "hr_worker_career_aspirations" ("worker_id", "target_job_id", "recorded_at");

CREATE UNIQUE INDEX ON "hr_promotion_recommendations" ("id", "worker_id");

CREATE UNIQUE INDEX ON "hr_promotion_decisions" ("recommendation_id", "step_number");

CREATE UNIQUE INDEX ON "hr_promotion_decisions" ("id", "recommendation_id");

CREATE INDEX "ix_hr_goal_measurement_history" ON "hr_performance_goal_updates" ("goal_id", "observed_at");

CREATE UNIQUE INDEX ON "hr_performance_goal_updates" ("id", "goal_id");

CREATE UNIQUE INDEX ON "hr_performance_review_goal_scores" ("performance_review_id", "goal_id");

COMMENT ON COLUMN "hr_job_families"."job_domain" IS 'manufacturing, engineering, finance, sales, operations, professional, executive, other';

COMMENT ON TABLE "hr_positions" IS 'Position hierarchy cycles and temporal overlap are validated by HR workflow; position is an authorized seat, not an employee record.';

COMMENT ON COLUMN "hr_positions"."organization_unit_reference_id" IS 'Organization structure is Enterprise Foundation-owned';

COMMENT ON COLUMN "hr_positions"."cost_center_reference_id" IS 'Finance owns cost-center accounting';

COMMENT ON TABLE "hr_workers" IS 'One worker record per employment relationship; rehire can create a new relationship linked to the same person_party_reference_id. Do not delete employment history.';

COMMENT ON COLUMN "hr_workers"."person_party_reference_id" IS 'Natural-person identity and shared name/contact data are Enterprise Foundation-owned';

COMMENT ON COLUMN "hr_workers"."iam_principal_reference_id" IS 'Optional identity link; IAM credential and roles are not HR-owned';

COMMENT ON TABLE "hr_employment_contracts" IS 'Contract content is legal-document data with role-restricted access, immutable signed snapshots, retention and jurisdictional rules.';

COMMENT ON COLUMN "hr_employment_contracts"."contract_type" IS 'permanent, fixed_term, hourly, contractor, apprenticeship, collective_agreement, other';

COMMENT ON TABLE "hr_worker_assignments" IS 'Penempatan paralel diperbolehkan; satu penempatan maksimal 100% FTE. Service memeriksa kapasitas posisi dan jumlah FTE lintas penempatan. Baseline mencegah periode primary yang bertumpuk. Roster/jam aktual tetap di batas Workforce.';

COMMENT ON COLUMN "hr_worker_assignments"."assignment_type" IS 'primary, secondary, acting, temporary, project, concurrent';

COMMENT ON TABLE "hr_worker_lifecycle_events" IS 'Append-only employment lifecycle history; corrections are compensating events with actor/reason, not destructive edits.';

COMMENT ON COLUMN "hr_worker_lifecycle_events"."event_type" IS 'hire, transfer, promotion, leave, return, suspension, termination, rehire, legal_entity_change';

COMMENT ON COLUMN "hr_worker_documents"."document_type" IS 'contract, qualification, license, policy_acknowledgement, work_authorization, tax_form, other';

COMMENT ON COLUMN "hr_worker_documents"."document_reference_id" IS 'File belongs in protected document/content storage; do not store document bytes here';

COMMENT ON COLUMN "hr_worker_documents"."document_version_reference_id" IS 'Versi immutable milik DRM';

COMMENT ON COLUMN "hr_worker_documents"."document_number_digest" IS 'Use keyed digest for searchable identifiers where required; avoid plaintext national ID';

COMMENT ON COLUMN "hr_requisitions"."requisition_type" IS 'new_headcount, replacement, temporary, confidential, evergreen';

COMMENT ON COLUMN "hr_requisitions"."budget_reference_id" IS 'Finance owns authorized budget';

COMMENT ON TABLE "hr_candidates" IS 'Candidate records have explicit purpose, notice, retention, correction/deletion workflow and restricted access under applicable law.';

COMMENT ON COLUMN "hr_candidates"."contact_reference_id" IS 'Candidate contact details use Foundation party/contact with limited access';

COMMENT ON TABLE "hr_applications" IS 'Assessment scoring requires transparent job-related criteria, human review and jurisdiction-specific anti-discrimination safeguards; no automatic adverse decision is assumed.';

COMMENT ON COLUMN "hr_offers"."proposed_compensation_document_version_reference_id" IS 'Rincian usulan kompensasi di DRM, belum menjadi kompensasi aktif pekerja';

COMMENT ON COLUMN "hr_onboarding_tasks"."task_owner_type" IS 'worker, manager, hr, it, security, facilities';

COMMENT ON TABLE "hr_compensation_records" IS 'Komponen membedakan tunjangan sejenis pada penempatan yang sama. Baseline melarang overlap kompensasi approved/active/superseded untuk penempatan dan kode komponen yang sama. Service memvalidasi mata uang, periode penempatan dan band sebelum approval.';

COMMENT ON COLUMN "hr_compensation_records"."compensation_type" IS 'base_salary, hourly_rate, allowance, incentive_target, commission_target, equity_reference';

COMMENT ON COLUMN "hr_benefit_plans"."benefit_type" IS 'health, retirement, insurance, leave, allowance, meal, transport, other';

COMMENT ON COLUMN "hr_payroll_profiles"."tax_profile_reference_id" IS 'Protected tax provider/vault reference; actual tax ID and calculation rules must be jurisdiction-restricted';

COMMENT ON COLUMN "hr_payroll_profiles"."bank_account_token_reference_id" IS 'Protected payment instrument reference owned by Finance/payment provider; never store full bank/card credentials here';

COMMENT ON COLUMN "hr_pay_components"."component_type" IS 'earning, deduction, employer_contribution, tax, reimbursement';

COMMENT ON COLUMN "hr_pay_components"."calculation_method" IS 'fixed, hourly, percentage, formula, imported, statutory_engine';

COMMENT ON COLUMN "hr_pay_components"."formula_reference_id" IS 'Versioned rule engine; legal rates and formulas are jurisdiction-specific and not hard-coded in schema';

COMMENT ON TABLE "hr_payroll_runs" IS 'Calculation outputs are immutable snapshots by run/version. Approved payroll posts through Finance; payment execution and bank reconciliation belong to Finance.';

COMMENT ON COLUMN "hr_payroll_runs"."settlement_direction" IS 'payable atau recovery; jumlah hasil selalu magnitudo positif';

COMMENT ON COLUMN "hr_payroll_runs"."run_type" IS 'regular, off_cycle, bonus, correction, termination, reversal';

COMMENT ON TABLE "hr_payroll_results" IS 'Net = gross + reimbursement - deductions (termasuk pajak pekerja). Recovery/reversal memakai magnitudo positif dan settlement_direction pada run. Baseline merekonsiliasi jumlah baris saat approval dan membekukan hasil setelah approval.';

COMMENT ON COLUMN "hr_payroll_results"."employer_cost_amount" IS 'Gross ditambah kontribusi pemberi kerja; reimbursement tidak masuk biaya remunerasi ini';

COMMENT ON TABLE "hr_leave_policies" IS 'Policy and entitlement parameters are versioned by jurisdiction and collective agreement; actual absence request/calendar impact is Workforce Management-owned.';

COMMENT ON COLUMN "hr_leave_policies"."accrual_method" IS 'annual_grant, per_pay_period, hours_worked, tenure_tier, statutory_engine';

COMMENT ON TABLE "hr_leave_entitlements" IS 'Ledger HRM adalah sumber saldo. Opening/accrued/used/adjusted adalah proyeksi ledger dalam transaksi yang sama dengan lock entitlement dan kenaikan revision. Perintah pemakaian dari owner absensi memerlukan persetujuan HRM; kegagalan dependency tidak boleh menghasilkan approval. Ledger menyimpan grant awal; opening_balance harus sama dengan jumlah entry opening.';

COMMENT ON COLUMN "hr_performance_cycles"."cycle_type" IS 'annual, semiannual, probation, project, continuous_feedback';

COMMENT ON COLUMN "hr_performance_goals"."goal_type" IS 'business, development, safety, quality, customer, capability';

COMMENT ON TABLE "hr_performance_reviews" IS 'Evaluation is confidential employment data; access, retention, worker response/correction and appeal process depend on law and policy.';

COMMENT ON COLUMN "hr_performance_reviews"."reviewer_role" IS 'self, manager, peer, project_lead, calibration_panel';

COMMENT ON COLUMN "hr_learning_courses"."course_type" IS 'compliance, safety, technical, leadership, professional, onboarding, qualification';

COMMENT ON COLUMN "hr_worker_qualifications"."qualification_type" IS 'degree, license, certification, trade_skill, safety_authorization, language';

COMMENT ON TABLE "hr_succession_candidates" IS 'Succession and talent-potential data are highly restricted and should not be used as unreviewed automated employment decisions.';

COMMENT ON COLUMN "hr_succession_candidates"."readiness_code" IS 'ready_now, 1_year, 2_3_years, development_needed';

COMMENT ON COLUMN "hr_employee_relations_cases"."case_type" IS 'grievance, conduct, investigation, accommodation, employee_support, ethics_referral';

COMMENT ON COLUMN "hr_employee_relations_cases"."restricted_summary_reference_id" IS 'Sensitive narratives/evidence stored in access-controlled case vault, not generic notes/logs';

COMMENT ON TABLE "hr_employee_relations_events" IS 'Append-only audit metadata; no sensitive narrative in event payload.';

COMMENT ON COLUMN "hr_human_capital_metric_definitions"."metric_domain" IS 'composition, diversity, cost, productivity, wellbeing, compliance, recruitment, mobility, turnover, skills, development';

COMMENT ON TABLE "hr_workforce_plan_lines" IS 'Rencana kebutuhan orang/FTE, bukan roster. Service memeriksa unit/job/lokasi serta tanggal dalam plan. Referensi budget tidak membuktikan persetujuan Finance.';

COMMENT ON COLUMN "hr_workforce_plan_lines"."action_code" IS 'recruit, develop, redeploy, retain, reduce';

COMMENT ON COLUMN "hr_competencies"."proficiency_scale" IS 'Rubrik versi ini mendefinisikan level 1 sampai 5 secara eksplisit';

COMMENT ON TABLE "hr_worker_competency_assessments" IS 'Hasil asesmen adalah bukti kompetensi, bukan izin IAM atau pengganti sertifikat. Level antarversi rubrik tidak otomatis dapat dibandingkan.';

COMMENT ON TABLE "hr_offboarding_plans" IS 'Service memeriksa seluruh tugas wajib, pekerja dalam payroll final, bukti clearance dan keputusan IAM. Rencana selesai tidak otomatis menandai pembayaran atau revoke; masing-masing harus dibuktikan owner.';

COMMENT ON TABLE "hr_payroll_groups" IS 'Identitas legal entity/jurisdiction/currency bersifat tetap setelah direferensikan. Perubahan employer atau mata uang membuat grup baru. Service memvalidasi timezone dan currency melalui kontrak/reference yang disahkan.';

COMMENT ON COLUMN "hr_payroll_groups"."currency_scale" IS 'Presisi pembulatan yang disahkan; nilai payroll disimpan enam desimal';

COMMENT ON TABLE "hr_payroll_rule_sets" IS 'Manifest immutable aturan payroll/THR/pajak/BPJS, versi engine dan parameter. Tidak ada tarif hukum bawaan. Service memeriksa keberlakuan, artefak/digest serta persetujuan domain sebelum menjalankan aturan; formula bukan SQL bebas.';

COMMENT ON TABLE "hr_leave_ledger_entries" IS 'Append-only; satu pembalikan penuh per entry, jumlah kebalikan tepat divalidasi service dengan lock. Masa ledger mengikuti entitlement. Opening/accrual/usage mengisi bucket sejenis; adjustment/expiry/reversal mengisi adjusted_quantity. UOM diwarisi entitlement; satu sumber tidak boleh diposting dua kali. Event out-of-order ditahan sebelum acknowledgement.';

COMMENT ON COLUMN "hr_development_actions"."action_type" IS 'course, coaching, mentoring, assignment, self_study';

COMMENT ON TABLE "hr_collective_agreements" IS 'Perjanjian kerja bersama berversi; pihak perwakilan mengacu Foundation. Tidak menyimpan daftar keanggotaan serikat pekerja.';

COMMENT ON TABLE "hr_policies" IS 'Kebijakan HR umum. Perhitungan hak cuti tetap memakai hr_leave_policies; tarif payroll memakai hr_payroll_rule_sets. Pengakuan telah membaca bukan persetujuan pemrosesan data otomatis.';

COMMENT ON COLUMN "hr_policies"."category_code" IS 'conduct, inclusion, wellbeing, remote_work, grievance, other';

COMMENT ON TABLE "hr_service_requests" IS 'Tiket layanan HR internal; bukan CRM/customer service. Penyelesaian tiket tidak menggantikan transaksi koreksi pada owner data. Grievance/investigasi sensitif memakai employee_relations_cases.';

COMMENT ON COLUMN "hr_service_requests"."request_type" IS 'employment_letter, data_correction, benefit_query, payroll_query, policy_query, other';

COMMENT ON TABLE "hr_engagement_surveys" IS 'Hanya definisi dan hasil agregat, tanpa jawaban per orang, worker_id, diagnosis atau identitas responden. Ambang minimum produk 5 dapat dinaikkan; bukan jaminan anonimitas atau klaim kewajiban hukum. Service menolak cohort kecil/irisan yang dapat mengidentifikasi ulang.';

COMMENT ON COLUMN "hr_engagement_surveys"."topic_code" IS 'engagement, wellbeing, inclusion, exit_experience';

COMMENT ON TABLE "hr_access_requests" IS 'Status adalah proyeksi hasil IAM, tidak pernah dipakai menghitung ALLOW. Revoke hanya akses hubungan kerja terkait; identitas pelanggan pekerja yang sama tidak otomatis dihapus. Actor dan scope harus terverifikasi server.';

COMMENT ON COLUMN "hr_access_requests"."request_kind" IS 'provision, change, revoke';

COMMENT ON COLUMN "hr_access_requests"."approved_catalog_reference_id" IS 'Katalog owner yang disetujui; bukan role/grant efektif';

COMMENT ON TABLE "hr_outbox_events" IS 'Ditulis atomik bersama perubahan bisnis. Pengiriman at-least-once; published bukan bukti domain penerima selesai. Payload memakai allowlist ID/status/versi minimum, tanpa slip gaji/narasi kasus/secret. Worker memakai lock SKIP LOCKED dan penerima melakukan deduplikasi.';

COMMENT ON TABLE "hr_inbox_events" IS 'Receipt processed dan efek bisnis harus satu transaksi. Sumber terautentikasi; event ID sama dengan digest berbeda ditolak. Ordering/replay dan payload retrieval memerlukan transport durable; tabel ini bukan implementasi consumer.';

COMMENT ON TABLE "hr_talent_frameworks" IS 'Rubrik disahkan dan berversi; arti level 1–5 serta pemetaan band/9-box harus eksplisit. Bukan label kepribadian atau keputusan kerja otomatis.';

COMMENT ON COLUMN "hr_talent_dimensions"."rating_anchors" IS 'Definisi bukti teramati untuk seluruh level 1–5; divalidasi service';

COMMENT ON TABLE "hr_talent_assessments" IS 'Potensi/readiness berbeda dari hasil performance dan kompetensi. Snapshot final immutable; revisi memakai assessment_version baru. Kandidat eksternal tetap di ATS sampai menjadi worker.';

COMMENT ON COLUMN "hr_talent_assessments"."segment_code" IS 'Kode hasil interpretasi rubrik berversi, bukan grant atau keputusan promosi';

COMMENT ON TABLE "hr_talent_calibration_decisions" IS 'Keputusan panel terpisah dari skor awal; bukti/skor awal tidak ditimpa. Service memeriksa assessment final, sesi terbuka dan mandat panel. Koreksi melalui sesi baru.';

COMMENT ON TABLE "hr_talent_pool_memberships" IS 'Keanggotaan talenta internal berbeda dari candidate talent_pool ATS dan group IAM; tidak memberi janji promosi.';

COMMENT ON TABLE "hr_career_path_transitions" IS 'Jalur bisa bercabang; kebutuhan skill memakai hr_job_competencies pada job tujuan. Mobilitas lateral dapat membentuk siklus yang sah; transisi tidak mengeksekusi perpindahan.';

COMMENT ON TABLE "hr_worker_career_aspirations" IS 'Aspirasi dinyatakan pekerja, bukan penilaian manajer. Jangan menyimpulkan kesediaan relokasi dari data keluarga atau demografi.';

COMMENT ON TABLE "hr_promotion_recommendations" IS 'Rekomendasi bukan perubahan jabatan. Service HRM menerapkan keputusan dalam transaksi: tutup assignment lama, buat assignment baru, catat event promotion dan outbox, kemudian tandai applied. IAM memproses perubahan akses secara terpisah. Draft dapat diedit; revisi setelah submitted memakai rekomendasi baru.';

COMMENT ON TABLE "hr_promotion_decisions" IS 'Riwayat keputusan append-only. Penyetuju berbeda dari pengusul; service memverifikasi mandat, seluruh tahap wajib dan larangan persetujuan diri pekerja melalui IAM.';

COMMENT ON TABLE "hr_learning_course_competencies" IS 'Satu kursus dapat mengembangkan banyak kompetensi. Selesai kursus tidak otomatis memberi skor kompetensi tanpa asesmen.';

COMMENT ON TABLE "hr_performance_goal_updates" IS 'Riwayat hasil KPI append-only; koreksi menunjuk pengukuran pada goal sama. actual_value pada goal adalah proyeksi terakhir, diperbarui service secara atomik.';

COMMENT ON TABLE "hr_performance_review_goal_scores" IS 'Goal dan review harus milik pekerja serta siklus yang sama. Service memverifikasi total bobot/rubrik dan membekukan rincian saat review difinalisasi.';

ALTER TABLE "hr_jobs" ADD FOREIGN KEY ("job_family_id") REFERENCES "hr_job_families" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_positions" ADD FOREIGN KEY ("job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_positions" ADD FOREIGN KEY ("reports_to_position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_workers" ADD FOREIGN KEY ("primary_assignment_id", "id") REFERENCES "hr_worker_assignments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_employment_contracts" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_assignments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_assignments" ADD FOREIGN KEY ("position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_assignments" ADD FOREIGN KEY ("manager_worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_lifecycle_events" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_documents" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_emergency_contacts" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_requisitions" ADD FOREIGN KEY ("position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_requisitions" ADD FOREIGN KEY ("job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_applications" ADD FOREIGN KEY ("candidate_id") REFERENCES "hr_candidates" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_applications" ADD FOREIGN KEY ("requisition_id") REFERENCES "hr_requisitions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_applications" ADD FOREIGN KEY ("hired_worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_application_stage_history" ADD FOREIGN KEY ("application_id") REFERENCES "hr_applications" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_interviews" ADD FOREIGN KEY ("application_id") REFERENCES "hr_applications" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_interview_panelists" ADD FOREIGN KEY ("interview_id") REFERENCES "hr_interviews" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_interview_feedback" ADD FOREIGN KEY ("interview_id", "interviewer_principal_id") REFERENCES "hr_interview_panelists" ("interview_id", "interviewer_principal_id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_offers" ADD FOREIGN KEY ("application_id") REFERENCES "hr_applications" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_offers" ADD FOREIGN KEY ("proposed_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_onboarding_plans" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_onboarding_tasks" ADD FOREIGN KEY ("onboarding_plan_id") REFERENCES "hr_onboarding_plans" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_compensation_records" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_compensation_records" ADD FOREIGN KEY ("assignment_id", "worker_id") REFERENCES "hr_worker_assignments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_compensation_records" ADD FOREIGN KEY ("compensation_band_id") REFERENCES "hr_compensation_bands" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_benefit_enrollments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_benefit_enrollments" ADD FOREIGN KEY ("benefit_plan_id") REFERENCES "hr_benefit_plans" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_profiles" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_runs" ADD FOREIGN KEY ("payroll_period_id", "payroll_group_id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") REFERENCES "hr_payroll_periods" ("id", "payroll_group_id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_results" ADD FOREIGN KEY ("payroll_run_id", "payroll_group_id", "currency_code") REFERENCES "hr_payroll_runs" ("id", "payroll_group_id", "currency_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_results" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_results" ADD FOREIGN KEY ("payroll_profile_reference_id", "worker_id", "payroll_group_id", "currency_code") REFERENCES "hr_payroll_profiles" ("id", "worker_id", "payroll_group_id", "currency_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_result_lines" ADD FOREIGN KEY ("payroll_result_id") REFERENCES "hr_payroll_results" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_result_lines" ADD FOREIGN KEY ("component_id", "component_code_snapshot", "component_type_snapshot") REFERENCES "hr_pay_components" ("id", "component_code", "component_type") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_adjustments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_adjustments" ADD FOREIGN KEY ("payroll_period_id") REFERENCES "hr_payroll_periods" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_adjustments" ADD FOREIGN KEY ("component_id") REFERENCES "hr_pay_components" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_entitlements" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_entitlements" ADD FOREIGN KEY ("leave_policy_id", "leave_type_code_snapshot") REFERENCES "hr_leave_policies" ("id", "leave_type_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_adjustments" ADD FOREIGN KEY ("entitlement_id") REFERENCES "hr_leave_entitlements" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_goals" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_goals" ADD FOREIGN KEY ("performance_cycle_id") REFERENCES "hr_performance_cycles" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_goals" ADD FOREIGN KEY ("parent_goal_id", "worker_id", "performance_cycle_id") REFERENCES "hr_performance_goals" ("id", "worker_id", "performance_cycle_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_reviews" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_reviews" ADD FOREIGN KEY ("performance_cycle_id") REFERENCES "hr_performance_cycles" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_learning_enrollments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_learning_enrollments" ADD FOREIGN KEY ("course_id") REFERENCES "hr_learning_courses" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_qualifications" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_succession_plans" ADD FOREIGN KEY ("position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_succession_candidates" ADD FOREIGN KEY ("succession_plan_id") REFERENCES "hr_succession_plans" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_succession_candidates" ADD FOREIGN KEY ("candidate_worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_employee_relations_events" ADD FOREIGN KEY ("case_id") REFERENCES "hr_employee_relations_cases" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_human_capital_metric_snapshots" ADD FOREIGN KEY ("metric_definition_id") REFERENCES "hr_human_capital_metric_definitions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_workforce_plan_lines" ADD FOREIGN KEY ("workforce_plan_id") REFERENCES "hr_workforce_plans" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_workforce_plan_lines" ADD FOREIGN KEY ("job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_requisitions" ADD FOREIGN KEY ("workforce_plan_line_id") REFERENCES "hr_workforce_plan_lines" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_job_competencies" ADD FOREIGN KEY ("job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_job_competencies" ADD FOREIGN KEY ("competency_id") REFERENCES "hr_competencies" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_competency_assessments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_competency_assessments" ADD FOREIGN KEY ("competency_id") REFERENCES "hr_competencies" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_job_postings" ADD FOREIGN KEY ("requisition_id") REFERENCES "hr_requisitions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_applications" ADD FOREIGN KEY ("source_posting_id", "requisition_id") REFERENCES "hr_job_postings" ("id", "requisition_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_application_assessments" ADD FOREIGN KEY ("application_id") REFERENCES "hr_applications" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_offboarding_plans" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_offboarding_plans" ADD FOREIGN KEY ("final_payroll_run_id") REFERENCES "hr_payroll_runs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_offboarding_tasks" ADD FOREIGN KEY ("offboarding_plan_id") REFERENCES "hr_offboarding_plans" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_compensation_records" ADD FOREIGN KEY ("component_id", "component_code_snapshot") REFERENCES "hr_pay_components" ("id", "component_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_pay_components" ADD FOREIGN KEY ("rule_set_id", "jurisdiction_code") REFERENCES "hr_payroll_rule_sets" ("id", "jurisdiction_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_profiles" ADD FOREIGN KEY ("payroll_group_id", "payroll_jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") REFERENCES "hr_payroll_groups" ("id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_periods" ADD FOREIGN KEY ("payroll_group_id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") REFERENCES "hr_payroll_groups" ("id", "jurisdiction_code", "currency_code", "employing_legal_entity_reference_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_runs" ADD FOREIGN KEY ("rule_set_id", "jurisdiction_code") REFERENCES "hr_payroll_rule_sets" ("id", "jurisdiction_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_payroll_runs" ADD FOREIGN KEY ("original_run_id", "payroll_group_id", "currency_code") REFERENCES "hr_payroll_runs" ("id", "payroll_group_id", "currency_code") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_ledger_entries" ADD FOREIGN KEY ("entitlement_id") REFERENCES "hr_leave_entitlements" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_ledger_entries" ADD FOREIGN KEY ("reverses_entry_id", "entitlement_id") REFERENCES "hr_leave_ledger_entries" ("id", "entitlement_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_leave_ledger_entries" ADD FOREIGN KEY ("adjustment_id", "entitlement_id") REFERENCES "hr_leave_adjustments" ("id", "entitlement_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_plans" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_plans" ADD FOREIGN KEY ("mentor_worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_plans" ADD FOREIGN KEY ("target_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_actions" ADD FOREIGN KEY ("development_plan_id") REFERENCES "hr_development_plans" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_actions" ADD FOREIGN KEY ("competency_id") REFERENCES "hr_competencies" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_development_actions" ADD FOREIGN KEY ("course_id") REFERENCES "hr_learning_courses" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_reviews" ADD FOREIGN KEY ("development_plan_id", "worker_id") REFERENCES "hr_development_plans" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_succession_candidates" ADD FOREIGN KEY ("development_plan_id", "candidate_worker_id") REFERENCES "hr_development_plans" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_employment_contracts" ADD FOREIGN KEY ("collective_agreement_id") REFERENCES "hr_collective_agreements" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_employee_relations_cases" ADD FOREIGN KEY ("subject_worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_policy_acknowledgements" ADD FOREIGN KEY ("policy_id") REFERENCES "hr_policies" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_policy_acknowledgements" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_service_requests" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_service_request_events" ADD FOREIGN KEY ("service_request_id") REFERENCES "hr_service_requests" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_engagement_survey_results" ADD FOREIGN KEY ("survey_id") REFERENCES "hr_engagement_surveys" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_access_requests" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_access_requests" ADD FOREIGN KEY ("lifecycle_event_id", "worker_id") REFERENCES "hr_worker_lifecycle_events" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_dimensions" ADD FOREIGN KEY ("framework_id") REFERENCES "hr_talent_frameworks" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_review_cycles" ADD FOREIGN KEY ("framework_id") REFERENCES "hr_talent_frameworks" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_assessments" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_assessments" ADD FOREIGN KEY ("talent_review_cycle_id", "framework_id") REFERENCES "hr_talent_review_cycles" ("id", "framework_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_assessments" ADD FOREIGN KEY ("performance_review_id", "worker_id") REFERENCES "hr_performance_reviews" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_assessment_scores" ADD FOREIGN KEY ("assessment_id", "framework_id") REFERENCES "hr_talent_assessments" ("id", "framework_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_assessment_scores" ADD FOREIGN KEY ("dimension_id", "framework_id") REFERENCES "hr_talent_dimensions" ("id", "framework_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_calibration_sessions" ADD FOREIGN KEY ("talent_review_cycle_id") REFERENCES "hr_talent_review_cycles" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_calibration_decisions" ADD FOREIGN KEY ("session_id", "talent_review_cycle_id") REFERENCES "hr_talent_calibration_sessions" ("id", "talent_review_cycle_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_calibration_decisions" ADD FOREIGN KEY ("assessment_id", "talent_review_cycle_id") REFERENCES "hr_talent_assessments" ("id", "talent_review_cycle_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_calibration_decisions" ADD FOREIGN KEY ("assessment_id", "dimension_id", "original_score") REFERENCES "hr_talent_assessment_scores" ("assessment_id", "dimension_id", "score") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_pools" ADD FOREIGN KEY ("target_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_pool_memberships" ADD FOREIGN KEY ("talent_pool_id") REFERENCES "hr_talent_pools" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_talent_pool_memberships" ADD FOREIGN KEY ("assessment_id", "worker_id") REFERENCES "hr_talent_assessments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_career_paths" ADD FOREIGN KEY ("job_family_id") REFERENCES "hr_job_families" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_career_path_transitions" ADD FOREIGN KEY ("career_path_id") REFERENCES "hr_career_paths" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_career_path_transitions" ADD FOREIGN KEY ("from_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_career_path_transitions" ADD FOREIGN KEY ("to_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_career_aspirations" ADD FOREIGN KEY ("worker_id") REFERENCES "hr_workers" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_career_aspirations" ADD FOREIGN KEY ("target_job_id") REFERENCES "hr_jobs" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_career_aspirations" ADD FOREIGN KEY ("career_path_transition_id", "target_job_id") REFERENCES "hr_career_path_transitions" ("id", "to_job_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("source_assignment_id", "worker_id") REFERENCES "hr_worker_assignments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("target_position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("assessment_id", "worker_id") REFERENCES "hr_talent_assessments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("supersedes_recommendation_id", "worker_id") REFERENCES "hr_promotion_recommendations" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("approval_decision_id", "id") REFERENCES "hr_promotion_decisions" ("id", "recommendation_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("applied_assignment_id", "worker_id", "target_position_id") REFERENCES "hr_worker_assignments" ("id", "worker_id", "position_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_recommendations" ADD FOREIGN KEY ("applied_lifecycle_event_id", "worker_id") REFERENCES "hr_worker_lifecycle_events" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_promotion_decisions" ADD FOREIGN KEY ("recommendation_id") REFERENCES "hr_promotion_recommendations" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_learning_course_competencies" ADD FOREIGN KEY ("course_id") REFERENCES "hr_learning_courses" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_learning_course_competencies" ADD FOREIGN KEY ("competency_id") REFERENCES "hr_competencies" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_goal_updates" ADD FOREIGN KEY ("goal_id") REFERENCES "hr_performance_goals" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_goal_updates" ADD FOREIGN KEY ("supersedes_update_id", "goal_id") REFERENCES "hr_performance_goal_updates" ("id", "goal_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_review_goal_scores" ADD FOREIGN KEY ("performance_review_id", "worker_id", "performance_cycle_id") REFERENCES "hr_performance_reviews" ("id", "worker_id", "performance_cycle_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_performance_review_goal_scores" ADD FOREIGN KEY ("goal_id", "worker_id", "performance_cycle_id") REFERENCES "hr_performance_goals" ("id", "worker_id", "performance_cycle_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_succession_candidates" ADD FOREIGN KEY ("talent_assessment_id", "candidate_worker_id") REFERENCES "hr_talent_assessments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

-- Invariant PostgreSQL yang tidak dapat dinyatakan penuh dalam DBML.
-- Tanggal efektif inklusif; periode berikutnya dimulai sehari setelah akhir.
CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE hr_worker_assignments ADD CONSTRAINT ex_hr_primary_assignment_period
  EXCLUDE USING gist (worker_id WITH =, daterange(valid_from, valid_to, '[]') WITH &&)
  WHERE (assignment_type = 'primary');
ALTER TABLE hr_compensation_records ADD CONSTRAINT ex_hr_compensation_period
  EXCLUDE USING gist (worker_id WITH =, assignment_id WITH =,
    component_code_snapshot WITH =, daterange(effective_from, effective_to, '[]') WITH &&)
  WHERE (status IN ('approved', 'active', 'superseded'));
ALTER TABLE hr_payroll_profiles ADD CONSTRAINT ex_hr_payroll_profile_period
  EXCLUDE USING gist (worker_id WITH =, payroll_group_id WITH =,
    daterange(effective_from, effective_to, '[]') WITH &&);
ALTER TABLE hr_leave_entitlements ADD CONSTRAINT ex_hr_leave_entitlement_period
  EXCLUDE USING gist (worker_id WITH =, leave_type_code_snapshot WITH =,
    daterange(entitlement_period_start, entitlement_period_end, '[]') WITH &&);

-- Bukti dan ledger dikoreksi dengan versi/entry baru, bukan overwrite.
CREATE FUNCTION hr_reject_immutable_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'Immutable HR record cannot be updated or deleted' USING ERRCODE = '23514';
END;
$$;
CREATE TRIGGER tr_hr_leave_ledger_immutable BEFORE UPDATE OR DELETE ON hr_leave_ledger_entries
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
CREATE TRIGGER tr_hr_payroll_rule_immutable BEFORE UPDATE OR DELETE ON hr_payroll_rule_sets
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();

-- Pengubahan hasil mengambil lock run yang juga dipakai proses approval.
CREATE FUNCTION hr_guard_payroll_result() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  row_data jsonb;
  run_id uuid;
  run_status hr_payroll_run_status_enum;
BEGIN
  FOR row_data IN SELECT value FROM jsonb_array_elements(
    CASE TG_OP WHEN 'INSERT' THEN jsonb_build_array(to_jsonb(NEW))
      WHEN 'DELETE' THEN jsonb_build_array(to_jsonb(OLD))
      ELSE jsonb_build_array(to_jsonb(OLD), to_jsonb(NEW)) END)
  LOOP
    IF TG_TABLE_NAME = 'hr_payroll_results' THEN
      run_id := (row_data->>'payroll_run_id')::uuid;
    ELSE
      SELECT payroll_run_id INTO run_id FROM hr_payroll_results
        WHERE id = (row_data->>'payroll_result_id')::uuid;
    END IF;
    SELECT status INTO run_status FROM hr_payroll_runs WHERE id = run_id FOR UPDATE;
    IF run_status IN ('approved','posted','paid') THEN
      -- Metadata bukti boleh dilengkapi; nilai perhitungan tetap immutable.
      IF TG_TABLE_NAME = 'hr_payroll_results' AND TG_OP = 'UPDATE'
        AND (to_jsonb(NEW) - ARRAY['result_status','payment_reference_id','payslip_document_reference_id'])
          = (to_jsonb(OLD) - ARRAY['result_status','payment_reference_id','payslip_document_reference_id'])
        AND (to_jsonb(OLD)->>'payment_reference_id' IS NULL
          OR to_jsonb(NEW)->>'payment_reference_id' IS NOT DISTINCT FROM to_jsonb(OLD)->>'payment_reference_id')
        AND (to_jsonb(OLD)->>'payslip_document_reference_id' IS NULL
          OR to_jsonb(NEW)->>'payslip_document_reference_id' IS NOT DISTINCT FROM to_jsonb(OLD)->>'payslip_document_reference_id') THEN
        CONTINUE;
      END IF;
      RAISE EXCEPTION 'Approved payroll result is immutable' USING ERRCODE = '23514';
    END IF;
  END LOOP;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_payroll_result_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_payroll_results
  FOR EACH ROW EXECUTE FUNCTION hr_guard_payroll_result();
CREATE TRIGGER tr_hr_payroll_line_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_payroll_result_lines
  FOR EACH ROW EXECUTE FUNCTION hr_guard_payroll_result();

CREATE FUNCTION hr_guard_payroll_run() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  editable_fields text[] := ARRAY['status','finance_journal_reference_id',
    'payment_batch_reference_id','finance_settlement_event_reference_id','finance_settled_at'];
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'draft' THEN
      RAISE EXCEPTION 'Payroll run must start as draft' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF OLD.status IN ('approved','posted','paid') THEN
    IF TG_OP = 'DELETE' THEN
      RAISE EXCEPTION 'Approved payroll run cannot be deleted' USING ERRCODE = '23514';
    END IF;
    IF (to_jsonb(NEW) - editable_fields) IS DISTINCT FROM (to_jsonb(OLD) - editable_fields)
      OR NOT (NEW.status = OLD.status OR (OLD.status = 'approved' AND NEW.status = 'posted')
        OR (OLD.status = 'posted' AND NEW.status = 'paid'))
      OR (OLD.finance_journal_reference_id IS NOT NULL AND NEW.finance_journal_reference_id IS DISTINCT FROM OLD.finance_journal_reference_id)
      OR (OLD.payment_batch_reference_id IS NOT NULL AND NEW.payment_batch_reference_id IS DISTINCT FROM OLD.payment_batch_reference_id)
      OR (OLD.finance_settlement_event_reference_id IS NOT NULL AND NEW.finance_settlement_event_reference_id IS DISTINCT FROM OLD.finance_settlement_event_reference_id)
      OR (OLD.finance_settled_at IS NOT NULL AND NEW.finance_settled_at IS DISTINCT FROM OLD.finance_settled_at) THEN
      RAISE EXCEPTION 'Approved payroll snapshot or settlement evidence cannot be overwritten' USING ERRCODE = '23514';
    END IF;
  ELSIF TG_OP = 'UPDATE' THEN
    IF NEW.status IN ('posted','paid') THEN
      RAISE EXCEPTION 'Payroll must be approved before posting' USING ERRCODE = '23514';
    END IF;
    IF NEW.status = 'approved' THEN
      IF NOT EXISTS (SELECT 1 FROM hr_payroll_results WHERE payroll_run_id = NEW.id) THEN
        RAISE EXCEPTION 'Payroll approval requires results' USING ERRCODE = '23514';
      END IF;
      IF EXISTS (
        SELECT 1 FROM hr_payroll_results r
        LEFT JOIN LATERAL (
          SELECT count(*) AS line_count,
            coalesce(sum(amount) FILTER (WHERE component_type_snapshot = 'earning'),0) AS gross,
            coalesce(sum(amount) FILTER (WHERE component_type_snapshot IN ('deduction','tax')),0) AS deductions,
            coalesce(sum(amount) FILTER (WHERE component_type_snapshot = 'reimbursement'),0) AS reimbursement,
            coalesce(sum(amount) FILTER (WHERE component_type_snapshot = 'employer_contribution'),0) AS employer
          FROM hr_payroll_result_lines WHERE payroll_result_id = r.id
        ) totals ON true
        WHERE r.payroll_run_id = NEW.id AND (totals.line_count = 0
          OR r.gross_amount <> totals.gross OR r.total_deductions <> totals.deductions
          OR r.reimbursement_amount <> totals.reimbursement
          OR r.employer_cost_amount <> totals.gross + totals.employer)
      ) THEN
        RAISE EXCEPTION 'Payroll totals do not match result lines' USING ERRCODE = '23514';
      END IF;
    END IF;
  END IF;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_payroll_run_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_payroll_runs
  FOR EACH ROW EXECUTE FUNCTION hr_guard_payroll_run();

-- Snapshot penilaian final dan bukti riwayat tidak diubah diam-diam.
CREATE TRIGGER tr_hr_talent_calibration_immutable BEFORE UPDATE OR DELETE ON hr_talent_calibration_decisions
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
CREATE TRIGGER tr_hr_goal_update_immutable BEFORE UPDATE OR DELETE ON hr_performance_goal_updates
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
CREATE TRIGGER tr_hr_promotion_decision_immutable BEFORE UPDATE OR DELETE ON hr_promotion_decisions
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();

CREATE FUNCTION hr_guard_talent_assessment() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'draft' THEN
      RAISE EXCEPTION 'Talent assessment must start as draft' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF OLD.status = 'finalized' THEN
    RAISE EXCEPTION 'Finalized talent assessment is immutable' USING ERRCODE = '23514';
  END IF;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  IF NEW.status = 'finalized' THEN
    IF NOT EXISTS (SELECT 1 FROM hr_talent_assessment_scores WHERE assessment_id = NEW.id)
      OR EXISTS (SELECT 1 FROM hr_talent_dimensions d WHERE d.framework_id = NEW.framework_id AND d.is_required
        AND NOT EXISTS (SELECT 1 FROM hr_talent_assessment_scores s WHERE s.assessment_id = NEW.id AND s.dimension_id = d.id)) THEN
      RAISE EXCEPTION 'Required talent dimension scores are missing' USING ERRCODE = '23514';
    END IF;
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_talent_assessment_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_talent_assessments
  FOR EACH ROW EXECUTE FUNCTION hr_guard_talent_assessment();

CREATE FUNCTION hr_guard_talent_score() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  row_data jsonb;
  parent_status text;
BEGIN
  FOR row_data IN SELECT value FROM jsonb_array_elements(
    CASE TG_OP WHEN 'INSERT' THEN jsonb_build_array(to_jsonb(NEW))
      WHEN 'DELETE' THEN jsonb_build_array(to_jsonb(OLD))
      ELSE jsonb_build_array(to_jsonb(OLD), to_jsonb(NEW)) END)
  LOOP
    SELECT status INTO parent_status FROM hr_talent_assessments
      WHERE id = (row_data->>'assessment_id')::uuid FOR UPDATE;
    IF parent_status = 'finalized' THEN
      RAISE EXCEPTION 'Finalized talent scores are immutable' USING ERRCODE = '23514';
    END IF;
  END LOOP;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_talent_score_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_talent_assessment_scores
  FOR EACH ROW EXECUTE FUNCTION hr_guard_talent_score();

CREATE FUNCTION hr_guard_promotion_decision() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  recommendation hr_promotion_recommendations%ROWTYPE;
BEGIN
  SELECT * INTO recommendation FROM hr_promotion_recommendations WHERE id = NEW.recommendation_id FOR UPDATE;
  IF NOT FOUND OR recommendation.status <> 'submitted' THEN
    RAISE EXCEPTION 'Promotion decision requires a submitted recommendation' USING ERRCODE = '23514';
  END IF;
  IF NEW.decided_by_principal_id = recommendation.recommended_by_principal_id THEN
    RAISE EXCEPTION 'Promotion requester cannot approve own recommendation' USING ERRCODE = '23514';
  END IF;
  IF EXISTS (SELECT 1 FROM hr_promotion_decisions WHERE recommendation_id = NEW.recommendation_id AND step_number >= NEW.step_number) THEN
    RAISE EXCEPTION 'Promotion decision steps must increase' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_promotion_decision_guard BEFORE INSERT ON hr_promotion_decisions
  FOR EACH ROW EXECUTE FUNCTION hr_guard_promotion_decision();

CREATE FUNCTION hr_guard_promotion_recommendation() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  editable_fields text[] := ARRAY['status','approval_decision_id','applied_assignment_id','applied_lifecycle_event_id','applied_at'];
  latest_decision hr_promotion_decisions%ROWTYPE;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'draft' THEN
      RAISE EXCEPTION 'Promotion recommendation must start as draft' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.status <> 'draft' THEN
      RAISE EXCEPTION 'Submitted promotion history cannot be deleted' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF OLD.status <> 'draft' AND (to_jsonb(NEW) - editable_fields) IS DISTINCT FROM (to_jsonb(OLD) - editable_fields) THEN
    RAISE EXCEPTION 'Submitted promotion proposal is immutable; create a revision' USING ERRCODE = '23514';
  END IF;
  IF NOT (NEW.status = OLD.status
    OR (OLD.status = 'draft' AND NEW.status IN ('submitted','cancelled'))
    OR (OLD.status = 'submitted' AND NEW.status IN ('approved','rejected','returned','cancelled'))
    OR (OLD.status = 'approved' AND NEW.status = 'applied')) THEN
    RAISE EXCEPTION 'Invalid promotion status transition' USING ERRCODE = '23514';
  END IF;
  IF OLD.status IN ('approved','applied','rejected','returned','cancelled') THEN
    IF OLD.status = NEW.status AND NEW IS DISTINCT FROM OLD THEN
      RAISE EXCEPTION 'Final promotion evidence cannot be overwritten' USING ERRCODE = '23514';
    END IF;
    IF NEW.approval_decision_id IS DISTINCT FROM OLD.approval_decision_id THEN
      RAISE EXCEPTION 'Promotion approval cannot be replaced' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF NEW.status = 'submitted' THEN
    IF NOT EXISTS (SELECT 1 FROM hr_talent_assessments WHERE id = NEW.assessment_id AND status = 'finalized') THEN
      RAISE EXCEPTION 'Promotion requires a finalized talent assessment' USING ERRCODE = '23514';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM hr_worker_assignments WHERE id = NEW.source_assignment_id
      AND assignment_type = 'primary' AND position_id <> NEW.target_position_id) THEN
      RAISE EXCEPTION 'Promotion requires a primary source and a different target position' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF OLD.status = 'submitted' AND NEW.status IN ('approved','rejected','returned') THEN
    SELECT * INTO latest_decision FROM hr_promotion_decisions
      WHERE recommendation_id = NEW.id ORDER BY step_number DESC LIMIT 1;
    IF NOT FOUND
      OR (NEW.status = 'approved' AND (latest_decision.decision <> 'approve' OR NEW.approval_decision_id IS DISTINCT FROM latest_decision.id))
      OR (NEW.status = 'rejected' AND latest_decision.decision <> 'reject')
      OR (NEW.status = 'returned' AND latest_decision.decision <> 'return') THEN
      RAISE EXCEPTION 'Promotion status must match the latest decision' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF NEW.status = 'applied' THEN
    IF NOT EXISTS (SELECT 1 FROM hr_worker_assignments a JOIN hr_workers w ON w.id = a.worker_id
      WHERE a.id = NEW.applied_assignment_id AND a.assignment_type = 'primary'
        AND a.valid_from = NEW.recommended_effective_date AND w.primary_assignment_id = a.id)
      OR NOT EXISTS (SELECT 1 FROM hr_worker_assignments WHERE id = NEW.source_assignment_id
        AND valid_to = NEW.recommended_effective_date - 1)
      OR NOT EXISTS (SELECT 1 FROM hr_worker_lifecycle_events WHERE id = NEW.applied_lifecycle_event_id
        AND event_type = 'promotion' AND effective_date = NEW.recommended_effective_date AND source_reference = NEW.id::text) THEN
      RAISE EXCEPTION 'Promotion application requires matching assignment and lifecycle evidence' USING ERRCODE = '23514';
    END IF;
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_promotion_recommendation_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_promotion_recommendations
  FOR EACH ROW EXECUTE FUNCTION hr_guard_promotion_recommendation();

CREATE TRIGGER tr_hr_talent_framework_immutable BEFORE UPDATE OR DELETE ON hr_talent_frameworks
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
CREATE FUNCTION hr_guard_talent_dimension() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  row_data jsonb;
BEGIN
  FOR row_data IN SELECT value FROM jsonb_array_elements(
    CASE TG_OP WHEN 'INSERT' THEN jsonb_build_array(to_jsonb(NEW))
      WHEN 'DELETE' THEN jsonb_build_array(to_jsonb(OLD))
      ELSE jsonb_build_array(to_jsonb(OLD), to_jsonb(NEW)) END)
  LOOP
    PERFORM 1 FROM hr_talent_frameworks WHERE id = (row_data->>'framework_id')::uuid FOR UPDATE;
    IF EXISTS (SELECT 1 FROM hr_talent_review_cycles WHERE framework_id = (row_data->>'framework_id')::uuid) THEN
      RAISE EXCEPTION 'Talent dimensions are frozen once a cycle references the framework' USING ERRCODE = '23514';
    END IF;
  END LOOP;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER tr_hr_talent_dimension_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_talent_dimensions
  FOR EACH ROW EXECUTE FUNCTION hr_guard_talent_dimension();

ALTER TABLE hr_talent_pool_memberships ADD CONSTRAINT ex_hr_talent_pool_membership
  EXCLUDE USING gist (talent_pool_id WITH =, worker_id WITH =, daterange(valid_from, valid_to, '[]') WITH &&)
  WHERE (status IN ('active','exited'));

-- Simpan identitas pembuat dan pengusul tanpa mengubah baseline terpasang.
ALTER TABLE hr_requisitions
  ADD COLUMN created_by_principal_id uuid,
  ADD COLUMN submitted_at timestamptz,
  ADD COLUMN submitted_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_requisition_submission_pair
    CHECK ((submitted_at IS NULL) = (submitted_by_principal_id IS NULL)),
  ADD CONSTRAINT ck_hr_requisition_approval_separation
    CHECK ((approved_by_principal_id IS NULL OR created_by_principal_id IS NULL OR
      approved_by_principal_id <> created_by_principal_id) AND
      (approved_by_principal_id IS NULL OR submitted_by_principal_id IS NULL OR
      approved_by_principal_id <> submitted_by_principal_id));

-- Simpan alasan, actor, dan waktu penarikan posting secara atomik.
ALTER TABLE hr_job_postings
  ADD COLUMN withdrawn_at timestamptz,
  ADD COLUMN withdrawal_reason_code varchar(80),
  ADD COLUMN withdrawn_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_posting_withdrawn
    CHECK (status <> 'withdrawn' OR
      (published_at IS NOT NULL AND withdrawn_at IS NOT NULL AND
        withdrawn_at > published_at AND closes_at IS NULL AND
        withdrawal_reason_code IS NOT NULL AND btrim(withdrawal_reason_code) <> '' AND
        withdrawn_by_principal_id IS NOT NULL)),
  ADD CONSTRAINT ck_hr_posting_withdrawn_only
    CHECK (status = 'withdrawn' OR
      (withdrawn_at IS NULL AND withdrawal_reason_code IS NULL AND withdrawn_by_principal_id IS NULL));

-- Wajibkan alasan dan actor keputusan untuk lamaran yang ditolak.
ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_rejected
    CHECK (status <> 'rejected' OR
      (current_stage_code = 'rejected' AND disposition_reason_code IS NOT NULL AND
        btrim(disposition_reason_code) <> '' AND decision_at IS NOT NULL AND
        decision_at >= applied_at AND decision_by_principal_id IS NOT NULL AND
        hired_worker_id IS NULL));

-- Pastikan shortlist manual memiliki recruiter dan tidak meniru skor terverifikasi.
ALTER TABLE hr_applications DROP CONSTRAINT ck_hr_application_rejected;
ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_rejected
    CHECK (status <> 'rejected' OR
      (current_stage_code IS NOT NULL AND current_stage_code = 'rejected' AND
        disposition_reason_code IS NOT NULL AND btrim(disposition_reason_code) <> '' AND
        decision_at IS NOT NULL AND decision_at >= applied_at AND
        decision_by_principal_id IS NOT NULL AND hired_worker_id IS NULL));

ALTER TABLE hr_applications
  ADD CONSTRAINT ck_hr_application_shortlisted
    CHECK (status <> 'shortlisted' OR
      (current_stage_code IS NOT NULL AND current_stage_code = 'shortlisted' AND assigned_recruiter_principal_id IS NOT NULL AND
        screening_score IS NULL AND scorecard_version IS NULL AND disposition_reason_code IS NULL AND
        decision_at IS NULL AND decision_by_principal_id IS NULL AND hired_worker_id IS NULL));

-- Simpan keputusan penolakan requisition yang telah diajukan.
ALTER TABLE hr_requisitions
  ADD COLUMN rejected_at timestamptz,
  ADD COLUMN rejection_reason_code varchar(80),
  ADD COLUMN rejected_by_principal_id uuid,
  ADD CONSTRAINT ck_hr_requisition_rejected
    CHECK (status <> 'rejected' OR
      (created_by_principal_id IS NOT NULL AND submitted_at IS NOT NULL AND
       submitted_by_principal_id IS NOT NULL AND rejected_at IS NOT NULL AND
       rejected_at >= submitted_at AND rejection_reason_code IS NOT NULL AND
       btrim(rejection_reason_code) <> '' AND rejected_by_principal_id IS NOT NULL AND
       rejected_by_principal_id <> created_by_principal_id AND
       rejected_by_principal_id <> submitted_by_principal_id AND
       budget_reference_id IS NULL AND approved_at IS NULL AND
       approved_by_principal_id IS NULL AND opened_at IS NULL AND closed_at IS NULL)),
  ADD CONSTRAINT ck_hr_requisition_rejected_only
    CHECK (status = 'rejected' OR
      (rejected_at IS NULL AND rejection_reason_code IS NULL AND rejected_by_principal_id IS NULL));

-- Simpan pengaju draf penempatan untuk pemisahan tugas saat persetujuan.
ALTER TABLE hr_worker_assignments
  ADD COLUMN requested_by_principal_id uuid,
  ADD COLUMN requested_at timestamptz,
  ADD CONSTRAINT ck_hr_assignment_request_pair
    CHECK ((requested_by_principal_id IS NULL) = (requested_at IS NULL)),
  ADD CONSTRAINT ck_hr_assignment_approval_pair
    CHECK ((approved_by_principal_id IS NULL) = (approved_at IS NULL)),
  ADD CONSTRAINT ck_hr_assignment_approval_provenance
    CHECK (approved_at IS NULL OR
      (requested_by_principal_id IS NOT NULL AND requested_at IS NOT NULL AND
       approved_by_principal_id <> requested_by_principal_id AND approved_at >= requested_at));

CREATE INDEX ix_hr_assignment_requester
  ON hr_worker_assignments (requested_by_principal_id, requested_at);
-- HRM assignment cancellation
ALTER TABLE hr_worker_assignments
  ADD COLUMN cancelled_at timestamptz,
  ADD CONSTRAINT ck_hr_assignment_cancellation
    CHECK (cancelled_at IS NULL OR (approved_at IS NOT NULL AND cancelled_at >= approved_at));

DROP INDEX uq_hr_worker_assignment_period;
CREATE UNIQUE INDEX uq_hr_worker_assignment_period
  ON hr_worker_assignments (worker_id, position_id, assignment_type, valid_from)
  WHERE cancelled_at IS NULL;

ALTER TABLE hr_worker_assignments DROP CONSTRAINT ex_hr_primary_assignment_period;
ALTER TABLE hr_worker_assignments ADD CONSTRAINT ex_hr_primary_assignment_period
  EXCLUDE USING gist (worker_id WITH =, daterange(valid_from, valid_to, '[]') WITH &&)
  WHERE (assignment_type = 'primary' AND cancelled_at IS NULL);
