WITH period AS (
  SELECT p.*
  FROM hr_payroll_periods p
  JOIN hr_payroll_groups g ON g.id = p.payroll_group_id
  WHERE p.id = ? AND p.period_status = 'open' AND g.is_active
),
profiles AS (
  SELECT p.id, p.worker_id, p.payroll_country_code, p.tax_status_code,
         p.tax_profile_reference_id, p.withholding_election_reference_id,
         p.effective_from, p.effective_to, w.worker_category, w.hire_date, w.termination_date
  FROM hr_payroll_profiles p
  JOIN period t ON (p.payroll_group_id, p.employing_legal_entity_reference_id, p.currency_code, p.payroll_jurisdiction_code)
                 = (t.payroll_group_id, t.employing_legal_entity_reference_id, t.currency_code, t.jurisdiction_code)
  JOIN hr_workers w ON w.id = p.worker_id
  WHERE p.is_active AND p.effective_from <= t.period_end AND (p.effective_to IS NULL OR p.effective_to >= t.period_start)
    AND w.status IN ('active', 'leave_of_absence', 'suspended', 'notice_period', 'terminated', 'deceased')
    AND w.hire_date <= t.period_end AND (w.termination_date IS NULL OR w.termination_date >= t.period_start)
),
compensation AS (
  SELECT c.id, c.worker_id, c.component_id, c.component_code_snapshot, c.compensation_type,
         c.amount::text AS amount, c.pay_frequency, c.effective_from, c.effective_to
  FROM hr_compensation_records c
  JOIN hr_worker_assignments a ON a.id = c.assignment_id AND a.worker_id = c.worker_id
  JOIN hr_positions o ON o.id = a.position_id
  JOIN period t ON t.employing_legal_entity_reference_id = o.legal_entity_reference_id AND t.currency_code = c.currency_code
  WHERE c.status IN ('active', 'superseded')
    AND c.approved_at <= ? AND c.effective_from <= t.period_end AND (c.effective_to IS NULL OR c.effective_to >= t.period_start)
    AND a.cancelled_at IS NULL AND a.valid_from <= t.period_end AND (a.valid_to IS NULL OR a.valid_to >= t.period_start)
    AND EXISTS (SELECT 1 FROM profiles p WHERE p.worker_id = c.worker_id)
),
adjustments AS (
  SELECT a.id, a.worker_id, a.component_id, a.adjustment_type, a.amount::text AS amount
  FROM hr_payroll_adjustments a
  JOIN period t ON t.id = a.payroll_period_id AND t.currency_code = a.currency_code
  WHERE a.approved_at <= ? AND a.approved_by_principal_id IS NOT NULL
    AND EXISTS (SELECT 1 FROM profiles p WHERE p.worker_id = a.worker_id)
),
snapshot AS (
  SELECT t.*,
    convert_to(jsonb_build_object(
      'schema_version', 1,
      'period_id', t.id, 'group_id', t.payroll_group_id,
      'employer_id', t.employing_legal_entity_reference_id,
      'currency_code', t.currency_code, 'jurisdiction_code', t.jurisdiction_code,
      'period_start', t.period_start, 'period_end', t.period_end, 'pay_date', t.pay_date,
      'profiles', (SELECT jsonb_agg(to_jsonb(p) ORDER BY p.id) FROM profiles p),
      'compensation', (SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY c.id), '[]'::jsonb) FROM compensation c),
      'adjustments', (SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY a.id), '[]'::jsonb) FROM adjustments a)
    )::text, 'UTF8') AS content
  FROM period t
  WHERE EXISTS (SELECT 1 FROM profiles)
)
INSERT INTO hr_payroll_input_snapshots
  (id, payroll_period_id, payroll_group_id, employing_legal_entity_reference_id, jurisdiction_code,
   currency_code, content, digest, created_by_principal_id, created_at)
SELECT ?, id, payroll_group_id, employing_legal_entity_reference_id, jurisdiction_code, currency_code,
       content, sha256(content), ?, ?
FROM snapshot
RETURNING id, payroll_period_id, payroll_group_id, employing_legal_entity_reference_id, jurisdiction_code,
          currency_code, digest, created_by_principal_id, created_at
