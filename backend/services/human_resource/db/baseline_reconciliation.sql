-- Perubahan skema untuk ledger lama yang checksum-nya dikenali.
-- Assignment cancellation
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
-- Worker principal uniqueness
CREATE INDEX ix_hr_worker_iam_principal ON hr_workers (iam_principal_reference_id)
  WHERE iam_principal_reference_id IS NOT NULL;
CREATE UNIQUE INDEX uq_hr_worker_current_iam_principal ON hr_workers (iam_principal_reference_id)
  WHERE iam_principal_reference_id IS NOT NULL
    AND status IN ('active', 'leave_of_absence', 'suspended', 'notice_period');
-- HRM runtime
CREATE TABLE "hr_worker_assignment_transfers" (
  "id" uuid PRIMARY KEY,
  "worker_id" uuid NOT NULL,
  "source_assignment_id" uuid NOT NULL,
  "target_position_id" uuid NOT NULL,
  "effective_date" date NOT NULL,
  "reason_code" varchar(80) NOT NULL,
  "idempotency_key" varchar(160) UNIQUE NOT NULL,
  "status" varchar(30) NOT NULL DEFAULT 'submitted',
  "requested_by_principal_id" uuid NOT NULL,
  "requested_at" timestamptz NOT NULL,
  "decided_by_principal_id" uuid,
  "decided_at" timestamptz,
  "applied_assignment_id" uuid UNIQUE,
  CONSTRAINT "ck_hr_worker_assignment_transfer_input" CHECK (btrim(reason_code) <> '' AND btrim(idempotency_key) <> ''),
  CONSTRAINT "ck_hr_worker_assignment_transfer_decision" CHECK ((status = 'submitted' AND decided_at IS NULL AND decided_by_principal_id IS NULL AND applied_assignment_id IS NULL) OR (status IN ('applied', 'rejected') AND decided_at IS NOT NULL AND decided_by_principal_id IS NOT NULL AND decided_by_principal_id <> requested_by_principal_id AND decided_at >= requested_at AND ((status = 'applied' AND applied_assignment_id IS NOT NULL AND applied_assignment_id <> source_assignment_id) OR (status = 'rejected' AND applied_assignment_id IS NULL))))
);

ALTER TABLE hr_offers ADD COLUMN "candidate_response_evidence_id" uuid UNIQUE;

ALTER TABLE hr_offers ADD COLUMN "created_by_principal_id" uuid;

ALTER TABLE hr_offers ADD COLUMN "created_at" timestamptz;

ALTER TABLE hr_offers ADD COLUMN "withdrawn_at" timestamptz;

ALTER TABLE hr_offers ADD COLUMN "withdrawn_by_principal_id" uuid;

ALTER TABLE hr_offers ADD COLUMN "withdrawal_reason_code" varchar(80);

ALTER TABLE hr_offers ADD CONSTRAINT "ck_hr_offer_creator_pair" CHECK ((created_by_principal_id IS NULL AND created_at IS NULL) OR (created_by_principal_id IS NOT NULL AND created_at IS NOT NULL));

ALTER TABLE hr_offers ADD CONSTRAINT "ck_hr_offer_independent_approval" CHECK (created_by_principal_id IS NULL OR approved_by_principal_id IS NULL OR (approved_by_principal_id <> created_by_principal_id AND approved_at IS NOT NULL AND approved_at >= created_at));

ALTER TABLE hr_offers ADD CONSTRAINT "ck_hr_offer_withdrawal" CHECK ((offer_status <> 'withdrawn' AND withdrawn_at IS NULL AND withdrawn_by_principal_id IS NULL AND withdrawal_reason_code IS NULL) OR (offer_status = 'withdrawn' AND withdrawn_at IS NOT NULL AND withdrawn_by_principal_id IS NOT NULL AND withdrawal_reason_code IS NOT NULL AND issued_at IS NOT NULL AND withdrawn_at >= issued_at AND candidate_response IS NULL AND responded_at IS NULL));

ALTER TABLE hr_offers ADD CONSTRAINT "ck_hr_offer_candidate_response_evidence" CHECK ((offer_status NOT IN ('accepted', 'declined') AND candidate_response IS NULL AND responded_at IS NULL AND candidate_response_evidence_id IS NULL) OR (offer_status IN ('accepted', 'declined') AND candidate_response IS NOT NULL AND candidate_response = offer_status AND responded_at IS NOT NULL AND candidate_response_evidence_id IS NOT NULL AND issued_at IS NOT NULL AND expires_at IS NOT NULL AND responded_at >= issued_at AND responded_at < expires_at));

ALTER TABLE hr_payroll_runs ADD COLUMN "posting_requested_by_principal_id" uuid;

ALTER TABLE hr_payroll_runs ADD COLUMN "posting_requested_at" timestamptz;

ALTER TABLE hr_payroll_runs ADD COLUMN "cancelled_at" timestamptz;

ALTER TABLE hr_payroll_runs ADD COLUMN "cancelled_by_principal_id" uuid;

ALTER TABLE hr_payroll_runs ADD COLUMN "cancellation_reason" varchar(300);

ALTER TABLE hr_payroll_runs ADD CONSTRAINT "ck_hr_payroll_cancel_evidence" CHECK ((status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by_principal_id IS NOT NULL AND coalesce(btrim(cancellation_reason), '') <> '') OR (status <> 'cancelled' AND cancelled_at IS NULL AND cancelled_by_principal_id IS NULL AND cancellation_reason IS NULL));

ALTER TABLE hr_payroll_runs ADD CONSTRAINT "ck_hr_payroll_cancel_independent" CHECK (status <> 'cancelled' OR (calculated_at IS NOT NULL AND cancelled_at >= calculated_at AND approved_at IS NULL AND approved_by_principal_id IS NULL AND cancelled_by_principal_id <> started_by_principal_id));

ALTER TABLE hr_payroll_runs ADD CONSTRAINT "ck_hr_payroll_posting_request" CHECK ((posting_requested_at IS NULL AND posting_requested_by_principal_id IS NULL) OR (posting_requested_at IS NOT NULL AND posting_requested_by_principal_id IS NOT NULL AND status IN ('approved','posted','paid') AND approved_at IS NOT NULL AND posting_requested_at >= approved_at));

ALTER TABLE hr_payroll_results ADD COLUMN "payslip_publication_reference_id" uuid;

ALTER TABLE hr_payroll_results ADD CONSTRAINT "ck_hr_payslip_publication_document" CHECK (payslip_publication_reference_id IS NULL OR payslip_document_reference_id IS NOT NULL);

ALTER TABLE hr_leave_entitlements ADD COLUMN "closed_at" timestamptz;

ALTER TABLE hr_leave_entitlements ADD COLUMN "closed_by_principal_id" uuid;

ALTER TABLE hr_leave_entitlements ADD CONSTRAINT "ck_hr_leave_entitlement_close" CHECK ((closed_at IS NULL AND closed_by_principal_id IS NULL) OR (closed_at IS NOT NULL AND closed_by_principal_id IS NOT NULL AND closing_balance_snapshot = 0 AND (closed_at AT TIME ZONE 'UTC')::date > entitlement_period_end));

ALTER TABLE hr_leave_adjustments ADD COLUMN "rejected_at" timestamptz;

ALTER TABLE hr_leave_adjustments ADD COLUMN "rejected_by_principal_id" uuid;

ALTER TABLE hr_leave_adjustments ADD COLUMN "rejection_reason_code" varchar(80);

ALTER TABLE hr_leave_adjustments ADD CONSTRAINT "ck_hr_leave_adjustment_rejection" CHECK ((rejected_at IS NULL AND rejected_by_principal_id IS NULL AND rejection_reason_code IS NULL) OR (rejected_at IS NOT NULL AND rejected_by_principal_id IS NOT NULL AND rejection_reason_code IS NOT NULL AND btrim(rejection_reason_code) <> '' AND approved_at IS NULL AND rejected_at >= created_at AND rejected_by_principal_id <> requested_by_principal_id));

ALTER TABLE hr_succession_plans ADD COLUMN "created_by_principal_id" uuid;

ALTER TABLE hr_succession_plans ADD COLUMN "created_at" timestamptz;

ALTER TABLE hr_succession_plans ADD COLUMN "approval_mandate_reference_id" uuid;

ALTER TABLE hr_succession_plans ADD CONSTRAINT "ck_hr_succession_creator_pair" CHECK ((created_by_principal_id IS NULL AND created_at IS NULL) OR (created_by_principal_id IS NOT NULL AND created_at IS NOT NULL));

ALTER TABLE hr_succession_plans ADD CONSTRAINT "ck_hr_succession_independent_approval" CHECK (approved_at IS NULL OR (created_by_principal_id IS NOT NULL AND created_at IS NOT NULL AND approved_by_principal_id <> created_by_principal_id AND approved_at >= created_at));

ALTER TABLE hr_succession_plans ADD CONSTRAINT "ck_hr_succession_approval_evidence" CHECK ((approved_at IS NULL AND approved_by_principal_id IS NULL AND approval_mandate_reference_id IS NULL) OR (approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL AND approval_mandate_reference_id IS NOT NULL));

ALTER TABLE hr_workforce_plans ADD COLUMN "superseded_by_plan_id" uuid REFERENCES hr_workforce_plans (id) ON DELETE RESTRICT;

ALTER TABLE hr_workforce_plans ADD COLUMN "supersession_reason_code" varchar(80);

ALTER TABLE hr_workforce_plans ADD CONSTRAINT "ck_hr_workforce_plan_supersession" CHECK ((status = 'superseded' AND superseded_by_plan_id IS NOT NULL AND superseded_by_plan_id <> id AND supersession_reason_code IS NOT NULL AND btrim(supersession_reason_code) <> '') OR (status <> 'superseded' AND superseded_by_plan_id IS NULL AND supersession_reason_code IS NULL));

ALTER TABLE hr_offboarding_plans ADD COLUMN "cancelled_at" timestamptz;

ALTER TABLE hr_offboarding_plans ADD COLUMN "cancelled_by_principal_id" uuid;

ALTER TABLE hr_offboarding_plans ADD COLUMN "cancellation_reason" varchar(300);

ALTER TABLE hr_offboarding_plans ADD CONSTRAINT "ck_hr_offboarding_cancellation" CHECK ((status = 'cancelled' AND approved_at IS NOT NULL AND approved_by_principal_id IS NOT NULL AND cancelled_at IS NOT NULL AND cancelled_at >= approved_at AND cancelled_by_principal_id IS NOT NULL AND cancellation_reason IS NOT NULL AND btrim(cancellation_reason) <> '') OR (status <> 'cancelled' AND cancelled_at IS NULL AND cancelled_by_principal_id IS NULL AND cancellation_reason IS NULL));

CREATE INDEX "ix_hr_worker_assignment_transfer_history" ON "hr_worker_assignment_transfers" ("worker_id", "status", "requested_at");

CREATE UNIQUE INDEX "uq_hr_payroll_result_payslip" ON "hr_payroll_results" ("payslip_document_reference_id");

CREATE UNIQUE INDEX uq_hr_offboarding_open_worker ON hr_offboarding_plans (worker_id)
  WHERE status IN ('draft', 'approved', 'in_progress');

CREATE UNIQUE INDEX uq_hr_offer_active_application ON hr_offers (application_id)
  WHERE offer_status IN ('issued', 'accepted');

ALTER TABLE "hr_worker_assignment_transfers" ADD FOREIGN KEY ("source_assignment_id", "worker_id") REFERENCES "hr_worker_assignments" ("id", "worker_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_assignment_transfers" ADD FOREIGN KEY ("target_position_id") REFERENCES "hr_positions" ("id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "hr_worker_assignment_transfers" ADD FOREIGN KEY ("applied_assignment_id", "worker_id", "target_position_id") REFERENCES "hr_worker_assignments" ("id", "worker_id", "position_id") ON DELETE RESTRICT DEFERRABLE INITIALLY IMMEDIATE;

CREATE OR REPLACE FUNCTION hr_guard_payroll_result() RETURNS trigger LANGUAGE plpgsql AS $$
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
    IF TG_TABLE_NAME = 'hr_payroll_results' AND TG_OP IN ('INSERT','UPDATE')
      AND to_jsonb(NEW)->>'payslip_document_reference_id' IS NOT NULL
      AND (TG_OP = 'INSERT' OR to_jsonb(NEW)->>'payslip_document_reference_id'
        IS DISTINCT FROM to_jsonb(OLD)->>'payslip_document_reference_id')
      AND (run_status IS NULL OR run_status NOT IN ('posted','paid')) THEN
      RAISE EXCEPTION 'Payroll must be posted before attaching a payslip' USING ERRCODE = '23514';
    END IF;
    IF run_status IN ('review','cancelled') THEN
      RAISE EXCEPTION 'Payroll results under review are immutable' USING ERRCODE = '23514';
    END IF;
    IF run_status IN ('approved','posted','paid') THEN
      -- Metadata bukti boleh dilengkapi; nilai perhitungan tetap immutable.
      IF TG_TABLE_NAME = 'hr_payroll_results' AND TG_OP = 'UPDATE'
        AND (to_jsonb(NEW) - ARRAY['result_status','payment_reference_id','payslip_document_reference_id','payslip_publication_reference_id'])
          = (to_jsonb(OLD) - ARRAY['result_status','payment_reference_id','payslip_document_reference_id','payslip_publication_reference_id'])
        AND (to_jsonb(OLD)->>'payment_reference_id' IS NULL
          OR to_jsonb(NEW)->>'payment_reference_id' IS NOT DISTINCT FROM to_jsonb(OLD)->>'payment_reference_id')
        AND (to_jsonb(OLD)->>'payslip_document_reference_id' IS NULL
          OR to_jsonb(NEW)->>'payslip_document_reference_id' IS NOT DISTINCT FROM to_jsonb(OLD)->>'payslip_document_reference_id')
        AND (to_jsonb(NEW)->>'payslip_publication_reference_id'
          IS NOT DISTINCT FROM to_jsonb(OLD)->>'payslip_publication_reference_id'
          OR to_jsonb(OLD)->>'payslip_document_reference_id' IS NULL) THEN
        CONTINUE;
      END IF;
      RAISE EXCEPTION 'Approved payroll result is immutable' USING ERRCODE = '23514';
    END IF;
  END LOOP;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION hr_guard_payroll_run() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
  editable_fields text[] := ARRAY['status','finance_journal_reference_id',
    'payment_batch_reference_id','finance_settlement_event_reference_id','finance_settled_at',
    'posting_requested_at','posting_requested_by_principal_id'];
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status <> 'draft' THEN
      RAISE EXCEPTION 'Payroll run must start as draft' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF OLD.status = 'cancelled' THEN
    IF TG_OP = 'DELETE' OR NEW IS DISTINCT FROM OLD THEN
      RAISE EXCEPTION 'Cancelled payroll run is immutable' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'UPDATE' AND NEW.status = 'cancelled' THEN
    IF OLD.status <> 'review' OR NEW.cancelled_at IS NULL OR NEW.cancelled_by_principal_id IS NULL
      OR NEW.cancelled_by_principal_id = OLD.started_by_principal_id
      OR NEW.cancelled_at < OLD.calculated_at OR coalesce(btrim(NEW.cancellation_reason), '') = ''
      OR (to_jsonb(NEW) - ARRAY['status','cancelled_at','cancelled_by_principal_id','cancellation_reason'])
        IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['status','cancelled_at','cancelled_by_principal_id','cancellation_reason']) THEN
      RAISE EXCEPTION 'Payroll cancellation requires an unchanged review snapshot and independent reviewer' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'UPDATE' AND (NEW.cancelled_at IS NOT NULL OR NEW.cancelled_by_principal_id IS NOT NULL OR NEW.cancellation_reason IS NOT NULL) THEN
    RAISE EXCEPTION 'Payroll cancellation evidence requires cancelled status' USING ERRCODE = '23514';
  END IF;
  IF OLD.status = 'review' THEN
    IF TG_OP = 'DELETE' THEN
      RAISE EXCEPTION 'Payroll run under review cannot be deleted' USING ERRCODE = '23514';
    END IF;
    IF (NEW.status = 'review' AND NEW IS DISTINCT FROM OLD)
      OR (NEW.status NOT IN ('review','approved'))
      OR (to_jsonb(NEW) - ARRAY['status','approved_at','approved_by_principal_id'])
        IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['status','approved_at','approved_by_principal_id']) THEN
      RAISE EXCEPTION 'Payroll snapshot under review cannot be overwritten' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF TG_OP = 'UPDATE' AND NEW.status = 'review' AND OLD.status <> 'review'
    AND (OLD.status <> 'calculated' OR (to_jsonb(NEW) - 'status') IS DISTINCT FROM (to_jsonb(OLD) - 'status')) THEN
    RAISE EXCEPTION 'Payroll review requires an unchanged calculated snapshot' USING ERRCODE = '23514';
  END IF;
  IF OLD.status IN ('approved','posted','paid') THEN
    IF TG_OP = 'DELETE' THEN
      RAISE EXCEPTION 'Approved payroll run cannot be deleted' USING ERRCODE = '23514';
    END IF;
    IF (to_jsonb(NEW) - editable_fields) IS DISTINCT FROM (to_jsonb(OLD) - editable_fields)
      OR NOT (NEW.status = OLD.status OR (OLD.status = 'approved' AND NEW.status = 'posted')
        OR (OLD.status = 'posted' AND NEW.status = 'paid'))
      OR ((OLD.status <> 'approved' OR NEW.status <> 'approved' OR OLD.posting_requested_at IS NOT NULL)
        AND (NEW.posting_requested_at IS DISTINCT FROM OLD.posting_requested_at
          OR NEW.posting_requested_by_principal_id IS DISTINCT FROM OLD.posting_requested_by_principal_id))
      OR (OLD.finance_journal_reference_id IS NOT NULL AND NEW.finance_journal_reference_id IS DISTINCT FROM OLD.finance_journal_reference_id)
      OR (OLD.payment_batch_reference_id IS NOT NULL AND NEW.payment_batch_reference_id IS DISTINCT FROM OLD.payment_batch_reference_id)
      OR (OLD.finance_settlement_event_reference_id IS NOT NULL AND NEW.finance_settlement_event_reference_id IS DISTINCT FROM OLD.finance_settlement_event_reference_id)
      OR (OLD.finance_settled_at IS NOT NULL AND NEW.finance_settled_at IS DISTINCT FROM OLD.finance_settled_at) THEN
      RAISE EXCEPTION 'Approved payroll snapshot or settlement evidence cannot be overwritten' USING ERRCODE = '23514';
    END IF;
  ELSIF TG_OP = 'UPDATE' THEN
    IF NEW.status IN ('posted','paid') OR NEW.posting_requested_at IS NOT NULL THEN
      RAISE EXCEPTION 'Payroll must be approved before posting' USING ERRCODE = '23514';
    END IF;
    IF NEW.status = 'approved' THEN
      IF OLD.status <> 'review' THEN
        RAISE EXCEPTION 'Payroll approval requires review' USING ERRCODE = '23514';
      END IF;
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

-- HRM payroll input snapshots
CREATE TABLE hr_payroll_input_snapshots (
  id uuid PRIMARY KEY,
  payroll_period_id uuid NOT NULL,
  payroll_group_id uuid NOT NULL,
  employing_legal_entity_reference_id uuid NOT NULL,
  jurisdiction_code varchar(80) NOT NULL,
  currency_code char(3) NOT NULL,
  content bytea NOT NULL,
  digest bytea NOT NULL,
  created_by_principal_id uuid NOT NULL,
  created_at timestamptz NOT NULL,
  CONSTRAINT ck_hr_payroll_input_snapshot_digest CHECK (octet_length(content) > 0 AND digest = sha256(content)),
  CONSTRAINT fk_hr_payroll_input_snapshot_period
    FOREIGN KEY (payroll_period_id, payroll_group_id, jurisdiction_code, currency_code, employing_legal_entity_reference_id)
    REFERENCES hr_payroll_periods (id, payroll_group_id, jurisdiction_code, currency_code, employing_legal_entity_reference_id)
    ON DELETE RESTRICT
);

CREATE TRIGGER tr_hr_payroll_input_snapshot_immutable BEFORE UPDATE OR DELETE ON hr_payroll_input_snapshots
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
