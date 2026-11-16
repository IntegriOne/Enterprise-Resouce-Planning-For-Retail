-- Kasus regresi terhadap temuan audit dan batas kepemilikan internal.
BEGIN;
CREATE FUNCTION pg_temp.u(n bigint) RETURNS uuid LANGUAGE sql IMMUTABLE AS $$
  SELECT lpad(n::text,32,'0')::uuid
$$;
CREATE FUNCTION pg_temp.expect_error(statement text, expected_state text, label text)
RETURNS void LANGUAGE plpgsql AS $$
BEGIN
  BEGIN
    EXECUTE statement;
  EXCEPTION WHEN OTHERS THEN
    IF SQLSTATE = expected_state THEN
      RAISE NOTICE 'PASS: %', label;
      RETURN;
    END IF;
    RAISE EXCEPTION 'Unexpected SQLSTATE % for %: %', SQLSTATE, label, SQLERRM;
  END;
  RAISE EXCEPTION 'Expected rejection was not raised: %', label;
END;
$$;

INSERT INTO hr_job_families (id,family_code,family_name,job_domain)
VALUES (pg_temp.u(1),'RETAIL','Operasi toko','retail');
INSERT INTO hr_jobs (id,job_family_id,job_code,job_title,valid_from)
VALUES (pg_temp.u(2),pg_temp.u(1),'CASHIER','Kasir','2026-01-01');
INSERT INTO hr_positions (id,position_code,position_title,job_id,organization_unit_reference_id,legal_entity_reference_id,valid_from)
VALUES (pg_temp.u(3),'CASHIER-01','Kasir toko',pg_temp.u(2),pg_temp.u(900),pg_temp.u(901),'2026-01-01');
INSERT INTO hr_workers (id,worker_number,person_party_reference_id,worker_category,primary_legal_entity_reference_id)
VALUES (pg_temp.u(4),'W-A',pg_temp.u(904),'employee',pg_temp.u(901)),
       (pg_temp.u(5),'W-B',pg_temp.u(905),'employee',pg_temp.u(901));
INSERT INTO hr_onboarding_plans (id,worker_id,plan_code,start_date)
VALUES (pg_temp.u(100),pg_temp.u(4),'FIRST-WEEK','2026-01-01');
SELECT pg_temp.expect_error($q$UPDATE hr_onboarding_plans SET status='in_progress' WHERE id=pg_temp.u(100)$q$,'23514','onboarding mulai tanpa audit actor dan waktu');
UPDATE hr_onboarding_plans SET status='in_progress', started_at=now(), started_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(100);
SELECT pg_temp.expect_error($q$UPDATE hr_onboarding_plans SET status='completed', completed_at=now() WHERE id=pg_temp.u(100)$q$,'23514','onboarding selesai tanpa audit actor');
INSERT INTO hr_onboarding_tasks (id,onboarding_plan_id,task_code,task_name,task_owner_type)
VALUES (pg_temp.u(101),pg_temp.u(100),'ORIENTATION','Orientation','worker');
SELECT pg_temp.expect_error($q$UPDATE hr_onboarding_tasks SET completed_at=now() WHERE id=pg_temp.u(101)$q$,'23514','tugas onboarding selesai tanpa actor dan bukti');
INSERT INTO hr_worker_assignments (id,worker_id,position_id,assignment_type,valid_from,valid_to)
VALUES (pg_temp.u(6),pg_temp.u(4),pg_temp.u(3),'primary','2026-01-01','2026-06-30');
-- Periode bersebelahan sah; penempatan paralel secondary juga sah.
INSERT INTO hr_worker_assignments (id,worker_id,position_id,assignment_type,valid_from)
VALUES (pg_temp.u(7),pg_temp.u(4),pg_temp.u(3),'primary','2026-07-01'),
       (pg_temp.u(8),pg_temp.u(4),pg_temp.u(3),'secondary','2026-01-01');
SELECT pg_temp.expect_error($q$INSERT INTO hr_worker_assignments (id,worker_id,position_id,assignment_type,valid_from) VALUES (pg_temp.u(9),pg_temp.u(4),pg_temp.u(3),'primary','2026-06-30')$q$,'23P01','primary overlap termasuk batas tanggal');
SELECT pg_temp.expect_error($q$UPDATE hr_worker_assignments SET fte_percent=101 WHERE id=pg_temp.u(6)$q$,'23514','FTE di atas 100');
SELECT pg_temp.expect_error($q$UPDATE hr_workers SET primary_assignment_id=pg_temp.u(6) WHERE id=pg_temp.u(5)$q$,'23503','primary assignment pekerja lain');

INSERT INTO hr_payroll_groups (id,group_code,group_name,employing_legal_entity_reference_id,jurisdiction_code,currency_code,currency_scale,pay_frequency,timezone)
VALUES (pg_temp.u(10),'MONTHLY','Bulanan',pg_temp.u(901),'ID','IDR',0,'monthly','Asia/Jakarta'),
       (pg_temp.u(11),'MONTHLY','Perusahaan kedua',pg_temp.u(902),'ID','IDR',0,'monthly','Asia/Jakarta');
INSERT INTO hr_payroll_rule_sets (id,rule_code,rule_version,jurisdiction_code,effective_from,artifact_document_version_reference_id,artifact_digest,legal_basis_document_version_reference_id,approved_by_principal_id,approved_at)
VALUES (pg_temp.u(12),'ID-RULES','test-v1','ID','2026-01-01',pg_temp.u(910),'\x01',pg_temp.u(911),pg_temp.u(999),now());
INSERT INTO hr_pay_components (id,component_code,component_version,component_name,component_type,calculation_method,jurisdiction_code,effective_from)
VALUES (pg_temp.u(13),'BASE',1,'Gaji pokok','earning','fixed','ID','2026-01-01'),
       (pg_temp.u(14),'TAX',1,'Pajak','tax','statutory_engine','ID','2026-01-01'),
       (pg_temp.u(15),'EXPENSE',1,'Penggantian biaya','reimbursement','imported','ID','2026-01-01'),
       (pg_temp.u(16),'MEAL',1,'Uang makan','earning','fixed','ID','2026-01-01'),
       (pg_temp.u(17),'BASE',2,'Gaji pokok v2','earning','fixed','ID','2026-01-01');
INSERT INTO hr_compensation_records (id,worker_id,assignment_id,component_id,component_code_snapshot,compensation_type,amount,currency_code,pay_frequency,effective_from,status,approved_by_principal_id,approved_at)
VALUES (pg_temp.u(20),pg_temp.u(4),pg_temp.u(6),pg_temp.u(13),'BASE','base_salary',10000000,'IDR','monthly','2026-01-01','active',pg_temp.u(999),now()),
       (pg_temp.u(21),pg_temp.u(4),pg_temp.u(6),pg_temp.u(16),'MEAL','allowance',100000,'IDR','monthly','2026-01-01','active',pg_temp.u(999),now());
SELECT pg_temp.expect_error($q$UPDATE hr_compensation_records SET worker_id=pg_temp.u(5) WHERE id=pg_temp.u(20)$q$,'23503','kompensasi dan penempatan beda pekerja');
SELECT pg_temp.expect_error($q$INSERT INTO hr_compensation_records (id,worker_id,assignment_id,component_id,component_code_snapshot,compensation_type,amount,currency_code,pay_frequency,effective_from,status,approved_by_principal_id,approved_at) VALUES (pg_temp.u(22),pg_temp.u(4),pg_temp.u(6),pg_temp.u(17),'BASE','base_salary',12000000,'IDR','monthly','2026-02-01','active',pg_temp.u(999),now())$q$,'23P01','overlap kompensasi lintas versi komponen');
SELECT pg_temp.expect_error($q$UPDATE hr_compensation_records SET approved_by_principal_id=NULL WHERE id=pg_temp.u(20)$q$,'23514','kompensasi aktif tanpa penyetuju');

INSERT INTO hr_payroll_profiles (id,worker_id,payroll_country_code,payroll_jurisdiction_code,payroll_group_id,employing_legal_entity_reference_id,currency_code,effective_from)
VALUES (pg_temp.u(30),pg_temp.u(4),'ID','ID',pg_temp.u(10),pg_temp.u(901),'IDR','2026-01-01'),
       (pg_temp.u(31),pg_temp.u(5),'ID','ID',pg_temp.u(10),pg_temp.u(901),'IDR','2026-01-01'),
       (pg_temp.u(32),pg_temp.u(4),'ID','ID',pg_temp.u(11),pg_temp.u(902),'IDR','2026-01-01');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_profiles SET employing_legal_entity_reference_id=pg_temp.u(902) WHERE id=pg_temp.u(30)$q$,'23503','grup payroll beda employer');
SELECT pg_temp.expect_error($q$INSERT INTO hr_payroll_profiles (id,worker_id,payroll_country_code,payroll_jurisdiction_code,payroll_group_id,employing_legal_entity_reference_id,currency_code,effective_from) VALUES (pg_temp.u(33),pg_temp.u(4),'ID','ID',pg_temp.u(10),pg_temp.u(901),'IDR','2026-02-01')$q$,'23P01','profil payroll bertumpuk');
-- Pembayaran sebelum akhir periode diperbolehkan, misalnya tanggal 25.
INSERT INTO hr_payroll_periods (id,payroll_group_id,employing_legal_entity_reference_id,currency_code,jurisdiction_code,period_code,period_start,period_end,pay_date)
VALUES (pg_temp.u(40),pg_temp.u(10),pg_temp.u(901),'IDR','ID','2026-11','2026-11-01','2026-11-30','2026-11-25');
INSERT INTO hr_payroll_runs (id,payroll_period_id,payroll_group_id,employing_legal_entity_reference_id,currency_code,rule_set_id,run_number,run_type,jurisdiction_code,rule_engine_name,rule_engine_version,input_snapshot_reference_id,input_digest,started_by_principal_id,idempotency_key)
VALUES (pg_temp.u(41),pg_temp.u(40),pg_temp.u(10),pg_temp.u(901),'IDR',pg_temp.u(12),1,'regular','ID','test','v1',pg_temp.u(912),'\x02',pg_temp.u(999),'run-1');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET run_type='correction' WHERE id=pg_temp.u(41)$q$,'23514','koreksi tanpa run asal');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET payroll_group_id=pg_temp.u(11) WHERE id=pg_temp.u(41)$q$,'23503','run beda grup periode');
INSERT INTO hr_payroll_results (id,payroll_run_id,payroll_group_id,worker_id,payroll_profile_reference_id,currency_code,gross_amount,reimbursement_amount,total_deductions,employer_cost_amount,net_pay_amount,calculation_digest)
VALUES (pg_temp.u(42),pg_temp.u(41),pg_temp.u(10),pg_temp.u(4),pg_temp.u(30),'IDR',10000000,3000000,2000000,10000000,11000000,'\x03');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_results SET payroll_profile_reference_id=pg_temp.u(31) WHERE id=pg_temp.u(42)$q$,'23503','hasil payroll memakai profil pekerja lain');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_results SET payroll_profile_reference_id=pg_temp.u(32) WHERE id=pg_temp.u(42)$q$,'23503','hasil payroll memakai profil grup lain');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_results SET net_pay_amount=9500000 WHERE id=pg_temp.u(42)$q$,'23514','persamaan gross reimbursement deductions net');
UPDATE hr_payroll_runs SET status='calculated',calculated_at=now(),output_digest='\x04' WHERE id=pg_temp.u(41);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET status='approved',approved_at=now(),approved_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(41)$q$,'23514','approval tanpa rincian payroll');
INSERT INTO hr_payroll_result_lines (id,payroll_result_id,component_id,component_code_snapshot,component_name_snapshot,component_type_snapshot,amount,line_sequence)
VALUES (pg_temp.u(43),pg_temp.u(42),pg_temp.u(13),'BASE','Gaji pokok','earning',10000000,1),
       (pg_temp.u(44),pg_temp.u(42),pg_temp.u(14),'TAX','Pajak','tax',2000000,2),
       (pg_temp.u(45),pg_temp.u(42),pg_temp.u(15),'EXPENSE','Biaya','reimbursement',3000000,3);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_result_lines SET component_type_snapshot='earning' WHERE id=pg_temp.u(44)$q$,'23503','tipe snapshot baris tidak sesuai komponen');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET status='approved',approved_at=now() WHERE id=pg_temp.u(41)$q$,'23514','approval tanpa aktor');
UPDATE hr_payroll_result_lines SET amount=1 WHERE id=pg_temp.u(43);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET status='approved',approved_at=now(),approved_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(41)$q$,'23514','approval total tidak sesuai baris');
UPDATE hr_payroll_result_lines SET amount=10000000 WHERE id=pg_temp.u(43);
UPDATE hr_payroll_runs SET status='approved',approved_at=now(),approved_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(41);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_result_lines SET amount=1 WHERE id=pg_temp.u(43)$q$,'23514','rincian payroll immutable setelah approval');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_results SET employer_cost_amount=1 WHERE id=pg_temp.u(42)$q$,'23514','hasil payroll immutable setelah approval');
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET status='draft' WHERE id=pg_temp.u(41)$q$,'23514','rollback status payroll approved');
UPDATE hr_payroll_results SET payslip_document_reference_id=pg_temp.u(913) WHERE id=pg_temp.u(42);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_results SET payslip_document_reference_id=pg_temp.u(914) WHERE id=pg_temp.u(42)$q$,'23514','overwrite bukti payslip');
UPDATE hr_payroll_runs SET status='posted',finance_journal_reference_id=pg_temp.u(915) WHERE id=pg_temp.u(41);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_runs SET status='paid' WHERE id=pg_temp.u(41)$q$,'23514','paid tanpa bukti Finance');
UPDATE hr_payroll_runs SET status='paid',payment_batch_reference_id=pg_temp.u(916),finance_settlement_event_reference_id=pg_temp.u(917),finance_settled_at=now() WHERE id=pg_temp.u(41);
SELECT pg_temp.expect_error($q$UPDATE hr_payroll_rule_sets SET artifact_digest='\x05' WHERE id=pg_temp.u(12)$q$,'23514','artefak aturan immutable');

INSERT INTO hr_leave_policies (id,policy_code,policy_version,policy_name,leave_type_code,jurisdiction_code,accrual_method,effective_from)
VALUES (pg_temp.u(50),'ANNUAL',1,'Cuti tahunan','annual','ID','annual_grant','2026-01-01');
INSERT INTO hr_leave_entitlements (id,worker_id,leave_policy_id,leave_type_code_snapshot,entitlement_period_start,entitlement_period_end,opening_balance,closing_balance_snapshot)
VALUES (pg_temp.u(51),pg_temp.u(4),pg_temp.u(50),'annual','2026-01-01','2026-12-31',12,12),
       (pg_temp.u(52),pg_temp.u(5),pg_temp.u(50),'annual','2026-01-01','2026-12-31',12,12);
INSERT INTO hr_leave_ledger_entries (id,entitlement_id,entry_sequence,entry_type,quantity_delta,effective_date,source_management,source_event_id,source_line_key,source_digest,reason_code,actor_principal_id)
VALUES (pg_temp.u(53),pg_temp.u(51),1,'opening',12,'2026-01-01','human_resource_management','grant-a','1','\x01','annual',pg_temp.u(999)),
       (pg_temp.u(54),pg_temp.u(51),2,'usage',-1,'2026-11-01','workforce','leave-1','1','\x02','approved',pg_temp.u(999));
UPDATE hr_leave_entitlements SET used_quantity_snapshot=1,closing_balance_snapshot=11,revision=2 WHERE id=pg_temp.u(51);
SELECT pg_temp.expect_error($q$UPDATE hr_leave_entitlements SET closing_balance_snapshot=10 WHERE id=pg_temp.u(51)$q$,'23514','persamaan saldo cuti');
SELECT pg_temp.expect_error($q$UPDATE hr_leave_ledger_entries SET quantity_delta=-2 WHERE id=pg_temp.u(54)$q$,'23514','ledger cuti immutable');
SELECT pg_temp.expect_error($q$INSERT INTO hr_leave_ledger_entries (id,entitlement_id,entry_sequence,entry_type,quantity_delta,effective_date,source_management,source_event_id,source_line_key,source_digest,reason_code,actor_principal_id) VALUES (pg_temp.u(55),pg_temp.u(51),3,'usage',-1,'2026-11-01','workforce','leave-1','1','\x02','approved',pg_temp.u(999))$q$,'23505','event cuti ganda');
SELECT pg_temp.expect_error($q$INSERT INTO hr_leave_ledger_entries (id,entitlement_id,entry_sequence,entry_type,quantity_delta,effective_date,reverses_entry_id,source_management,source_event_id,source_line_key,source_digest,reason_code,actor_principal_id) VALUES (pg_temp.u(56),pg_temp.u(52),1,'reversal',1,'2026-11-01',pg_temp.u(54),'workforce','reversal-1','1','\x02','reversed',pg_temp.u(999))$q$,'23503','reversal lintas entitlement');
SELECT pg_temp.expect_error($q$INSERT INTO hr_leave_ledger_entries (id,entitlement_id,entry_sequence,entry_type,quantity_delta,effective_date,source_management,source_event_id,source_line_key,source_digest,reason_code,actor_principal_id) VALUES (pg_temp.u(57),pg_temp.u(51),3,'usage',1,'2026-11-01','workforce','leave-2','1','\x02','approved',pg_temp.u(999))$q$,'23514','tanda penggunaan cuti positif');

INSERT INTO hr_human_capital_metric_definitions (id,metric_code,metric_version,metric_name,metric_domain)
VALUES (pg_temp.u(60),'HEADCOUNT','1','Jumlah pekerja','composition');
INSERT INTO hr_human_capital_metric_snapshots (id,metric_definition_id,period_start,period_end,methodology_version,source_digest)
VALUES (pg_temp.u(61),pg_temp.u(60),'2026-01-01','2026-12-31','1','\x01');
SELECT pg_temp.expect_error($q$INSERT INTO hr_human_capital_metric_snapshots (id,metric_definition_id,period_start,period_end,methodology_version,source_digest) VALUES (pg_temp.u(62),pg_temp.u(60),'2026-01-01','2026-12-31','1','\x02')$q$,'23505','metrik scope NULL tidak ganda');
INSERT INTO hr_human_capital_metric_snapshots (id,metric_definition_id,period_start,period_end,methodology_version,snapshot_version,source_digest)
VALUES (pg_temp.u(63),pg_temp.u(60),'2026-01-01','2026-12-31','1',2,'\x02');
SELECT pg_temp.expect_error($q$INSERT INTO hr_employee_relations_cases (id,case_number,case_type,subject_worker_id,restricted_summary_reference_id,idempotency_key) VALUES (pg_temp.u(64),'CASE-1','grievance',pg_temp.u(9999),pg_temp.u(930),'case-1')$q$,'23503','kasus hubungan karyawan tanpa worker sah');

INSERT INTO hr_worker_lifecycle_events (id,worker_id,event_type,occurred_at,effective_date,idempotency_key)
VALUES (1,pg_temp.u(4),'hire',now(),'2026-01-01','hire-a');
SELECT pg_temp.expect_error($q$INSERT INTO hr_access_requests (id,worker_id,lifecycle_event_id,request_kind,requested_scope_type,requested_scope_reference_id,requested_by_principal_id,idempotency_key) VALUES (pg_temp.u(70),pg_temp.u(5),1,'revoke','legal_entity',pg_temp.u(901),pg_temp.u(999),'revoke-b')$q$,'23503','access request memakai event pekerja lain');
SELECT pg_temp.expect_error($q$INSERT INTO hr_access_requests (id,worker_id,lifecycle_event_id,request_kind,requested_scope_type,requested_scope_reference_id,requested_by_principal_id,status,idempotency_key) VALUES (pg_temp.u(71),pg_temp.u(4),1,'revoke','legal_entity',pg_temp.u(901),pg_temp.u(999),'applied','revoke-a')$q$,'23514','akses applied tanpa keputusan IAM');

INSERT INTO hr_competencies (id,competency_code,competency_version,competency_name,category_code,proficiency_scale)
VALUES (pg_temp.u(80),'POS',1,'Pengoperasian kasir','technical','{"1":"pemula","5":"pelatih"}');
INSERT INTO hr_job_competencies (id,job_id,competency_id,required_level,effective_from)
VALUES (pg_temp.u(81),pg_temp.u(2),pg_temp.u(80),3,'2026-01-01');
SELECT pg_temp.expect_error($q$UPDATE hr_job_competencies SET required_level=6 WHERE id=pg_temp.u(81)$q$,'23514','level kompetensi di luar rubrik');
INSERT INTO hr_development_plans (id,worker_id,plan_code,plan_version,start_date,target_date)
VALUES (pg_temp.u(82),pg_temp.u(4),'CASHIER',1,'2026-01-01','2026-12-31');
INSERT INTO hr_succession_plans (id,position_id,plan_code,plan_version,review_date)
VALUES (pg_temp.u(83),pg_temp.u(3),'CASHIER',1,'2026-11-01');
SELECT pg_temp.expect_error($q$INSERT INTO hr_succession_candidates (id,succession_plan_id,candidate_worker_id,readiness_code,readiness_assessed_at,development_plan_id) VALUES (pg_temp.u(84),pg_temp.u(83),pg_temp.u(5),'development_needed','2026-11-01',pg_temp.u(82))$q$,'23503','suksesi memakai rencana pengembangan orang lain');
INSERT INTO hr_offboarding_plans (id,worker_id,plan_number,reason_code,notice_date,last_working_date,termination_effective_date,created_by_principal_id)
VALUES (pg_temp.u(85),pg_temp.u(4),'EXIT-1','resignation','2026-10-01','2026-11-30','2026-11-30',pg_temp.u(999));
INSERT INTO hr_offboarding_tasks (id,offboarding_plan_id,task_code,task_name,owner_management)
VALUES (pg_temp.u(86),pg_temp.u(85),'REVOKE','Cabut akses kerja','iam_management');
SELECT pg_temp.expect_error($q$UPDATE hr_offboarding_tasks SET status='completed',completed_at=now() WHERE id=pg_temp.u(86)$q$,'23514','offboarding selesai tanpa receipt owner');
SELECT pg_temp.expect_error($q$UPDATE hr_offboarding_tasks SET status='waived' WHERE id=pg_temp.u(86)$q$,'23514','waiver offboarding tanpa alasan dan aktor');
INSERT INTO hr_inbox_events (id,source_management,source_event_id,event_type,schema_version,source_aggregate_reference,source_aggregate_version,payload_digest)
VALUES (pg_temp.u(87),'finance_accounting','settled-1','payroll.settled',1,'batch-1',1,'\x01');
SELECT pg_temp.expect_error($q$INSERT INTO hr_inbox_events (id,source_management,source_event_id,event_type,schema_version,source_aggregate_reference,source_aggregate_version,payload_digest) VALUES (pg_temp.u(88),'finance_accounting','settled-1','payroll.settled',1,'batch-1',1,'\x01')$q$,'23505','deduplikasi event Finance');
INSERT INTO hr_leave_policies (id,policy_code,policy_version,policy_name,leave_type_code,jurisdiction_code,accrual_method,effective_from)
VALUES (pg_temp.u(89),'ANNUAL',2,'Cuti tahunan revisi','annual','ID','annual_grant','2026-07-01');
SELECT pg_temp.expect_error($q$INSERT INTO hr_leave_entitlements (id,worker_id,leave_policy_id,leave_type_code_snapshot,entitlement_period_start,entitlement_period_end) VALUES (pg_temp.u(90),pg_temp.u(4),pg_temp.u(89),'annual','2026-07-01','2026-12-31')$q$,'23P01','saldo cuti ganda lintas versi kebijakan');

ROLLBACK;
