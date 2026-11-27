-- Invariant PostgreSQL yang tidak dapat dinyatakan penuh dalam DBML.
-- Tanggal efektif inklusif; periode berikutnya dimulai sehari setelah akhir.
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE UNIQUE INDEX uq_hr_offboarding_open_worker ON hr_offboarding_plans (worker_id)
  WHERE status IN ('draft', 'approved', 'in_progress');
CREATE INDEX ix_hr_worker_iam_principal ON hr_workers (iam_principal_reference_id)
  WHERE iam_principal_reference_id IS NOT NULL;
CREATE UNIQUE INDEX uq_hr_worker_current_iam_principal ON hr_workers (iam_principal_reference_id)
  WHERE iam_principal_reference_id IS NOT NULL AND status IN ('active', 'leave_of_absence', 'suspended', 'notice_period');

-- Satu penawaran terbit atau diterima per lamaran, terlepas dari revisinya.
CREATE UNIQUE INDEX uq_hr_offer_active_application ON hr_offers (application_id)
  WHERE offer_status IN ('issued', 'accepted');

CREATE UNIQUE INDEX uq_hr_worker_assignment_period
  ON hr_worker_assignments (worker_id, position_id, assignment_type, valid_from)
  WHERE cancelled_at IS NULL;

ALTER TABLE hr_worker_assignments ADD CONSTRAINT ex_hr_primary_assignment_period
  EXCLUDE USING gist (worker_id WITH =, daterange(valid_from, valid_to, '[]') WITH &&)
  WHERE (assignment_type = 'primary' AND cancelled_at IS NULL);
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
CREATE TRIGGER tr_hr_payroll_result_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_payroll_results
  FOR EACH ROW EXECUTE FUNCTION hr_guard_payroll_result();
CREATE TRIGGER tr_hr_payroll_line_guard BEFORE INSERT OR UPDATE OR DELETE ON hr_payroll_result_lines
  FOR EACH ROW EXECUTE FUNCTION hr_guard_payroll_result();

CREATE FUNCTION hr_guard_payroll_run() RETURNS trigger LANGUAGE plpgsql AS $$
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

CREATE TRIGGER tr_hr_payroll_input_snapshot_immutable BEFORE UPDATE OR DELETE ON hr_payroll_input_snapshots
  FOR EACH ROW EXECUTE FUNCTION hr_reject_immutable_change();
