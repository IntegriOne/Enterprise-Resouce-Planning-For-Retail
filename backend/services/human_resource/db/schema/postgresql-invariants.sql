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
