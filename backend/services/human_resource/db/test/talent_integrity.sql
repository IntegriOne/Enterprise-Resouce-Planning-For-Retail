-- Talent internal dan promosi berbagi worker HRM tanpa mutasi jabatan otomatis.
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

INSERT INTO hr_job_families (id,family_code,family_name,job_domain) VALUES (pg_temp.u(1),'RETAIL','Retail','retail');
INSERT INTO hr_jobs (id,job_family_id,job_code,job_title,valid_from) VALUES
(pg_temp.u(2),pg_temp.u(1),'SUPERVISOR','Supervisor','2026-01-01'),
(pg_temp.u(3),pg_temp.u(1),'MANAGER','Manager','2026-01-01');
INSERT INTO hr_positions (id,position_code,position_title,job_id,organization_unit_reference_id,legal_entity_reference_id,valid_from) VALUES
(pg_temp.u(4),'SUP-01','Supervisor toko',pg_temp.u(2),pg_temp.u(900),pg_temp.u(901),'2026-01-01'),
(pg_temp.u(5),'MGR-01','Manager toko',pg_temp.u(3),pg_temp.u(900),pg_temp.u(901),'2026-01-01');
INSERT INTO hr_workers (id,worker_number,person_party_reference_id,worker_category,primary_legal_entity_reference_id) VALUES
(pg_temp.u(6),'ANDI',pg_temp.u(906),'employee',pg_temp.u(901)),
(pg_temp.u(7),'BUDI',pg_temp.u(907),'employee',pg_temp.u(901));
INSERT INTO hr_worker_assignments (id,worker_id,position_id,assignment_type,valid_from)
VALUES (pg_temp.u(8),pg_temp.u(6),pg_temp.u(4),'primary','2026-01-01');
UPDATE hr_workers SET primary_assignment_id=pg_temp.u(8) WHERE id=pg_temp.u(6);
INSERT INTO hr_talent_frameworks (id,framework_code,framework_version,framework_name,rubric_document_version_reference_id,rubric_digest,effective_from,approved_by_principal_id,approved_at) VALUES
(pg_temp.u(10),'RETAIL',1,'Talent retail',pg_temp.u(910),'\x01','2026-01-01',pg_temp.u(999),now()),
(pg_temp.u(11),'OTHER',1,'Rubrik lain',pg_temp.u(911),'\x02','2026-01-01',pg_temp.u(999),now());
INSERT INTO hr_talent_dimensions (id,framework_id,dimension_code,dimension_name,definition,rating_anchors,weight_percent) VALUES
(pg_temp.u(12),pg_temp.u(10),'POTENTIAL','Potensi','Kesiapan lingkup lebih luas','{"1":"dasar","2":"terbimbing","3":"mandiri","4":"memimpin","5":"mengembangkan"}',100),
(pg_temp.u(13),pg_temp.u(11),'POTENTIAL','Potensi rubrik lain','Rubrik berbeda','{}',100);
INSERT INTO hr_talent_review_cycles (id,cycle_code,framework_id,legal_entity_reference_id,period_start,period_end,owner_principal_id,status) VALUES
(pg_temp.u(14),'2026-H2',pg_temp.u(10),pg_temp.u(901),'2026-07-01','2026-12-31',pg_temp.u(999),'open'),
(pg_temp.u(15),'2026-OTHER',pg_temp.u(11),pg_temp.u(901),'2026-07-01','2026-12-31',pg_temp.u(999),'open');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_dimensions SET weight_percent=50 WHERE id=pg_temp.u(12)$q$,'23514','rubrik dimensi terpakai dibekukan');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_frameworks SET rubric_digest='\x04' WHERE id=pg_temp.u(10)$q$,'23514','manifest rubrik disahkan immutable');
INSERT INTO hr_talent_assessments (id,worker_id,talent_review_cycle_id,framework_id,assessment_version,assessed_by_principal_id,evidence_document_version_reference_id,evidence_digest) VALUES
(pg_temp.u(16),pg_temp.u(6),pg_temp.u(14),pg_temp.u(10),1,pg_temp.u(998),pg_temp.u(912),'\x03');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_assessments SET status='finalized',finalized_at=now(),finalized_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(16)$q$,'23514','finalisasi tanpa skor wajib');
SELECT pg_temp.expect_error($q$INSERT INTO hr_talent_assessment_scores (id,assessment_id,framework_id,dimension_id,score,evidence_document_version_reference_id,evidence_digest) VALUES (pg_temp.u(17),pg_temp.u(16),pg_temp.u(10),pg_temp.u(13),4,pg_temp.u(912),'\x03')$q$,'23503','dimensi dari rubrik lain ditolak');
INSERT INTO hr_talent_assessment_scores (id,assessment_id,framework_id,dimension_id,score,evidence_document_version_reference_id,evidence_digest)
VALUES (pg_temp.u(17),pg_temp.u(16),pg_temp.u(10),pg_temp.u(12),4,pg_temp.u(912),'\x03');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_assessment_scores SET score=6 WHERE id=pg_temp.u(17)$q$,'23514','skor talent di luar rubrik');
UPDATE hr_talent_assessments SET status='finalized',finalized_at=now(),finalized_by_principal_id=pg_temp.u(999) WHERE id=pg_temp.u(16);
SELECT pg_temp.expect_error($q$UPDATE hr_talent_assessment_scores SET score=5 WHERE id=pg_temp.u(17)$q$,'23514','skor talent final immutable');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_assessments SET evidence_digest='\x04' WHERE id=pg_temp.u(16)$q$,'23514','bukti asesmen final immutable');
INSERT INTO hr_talent_calibration_sessions (id,talent_review_cycle_id,session_code,chair_principal_id,scheduled_at,status)
VALUES (pg_temp.u(18),pg_temp.u(14),'PANEL',pg_temp.u(999),now(),'open');
SELECT pg_temp.expect_error($q$INSERT INTO hr_talent_calibration_decisions (id,session_id,talent_review_cycle_id,assessment_id,dimension_id,original_score,calibrated_score,rationale_document_version_reference_id,rationale_digest,decided_by_principal_id) VALUES (pg_temp.u(19),pg_temp.u(18),pg_temp.u(15),pg_temp.u(16),pg_temp.u(12),4,5,pg_temp.u(913),'\x04',pg_temp.u(999))$q$,'23503','kalibrasi lintas siklus');
SELECT pg_temp.expect_error($q$INSERT INTO hr_talent_calibration_decisions (id,session_id,talent_review_cycle_id,assessment_id,dimension_id,original_score,calibrated_score,rationale_document_version_reference_id,rationale_digest,decided_by_principal_id) VALUES (pg_temp.u(19),pg_temp.u(18),pg_temp.u(14),pg_temp.u(16),pg_temp.u(12),3,5,pg_temp.u(913),'\x04',pg_temp.u(999))$q$,'23503','skor awal kalibrasi palsu');
INSERT INTO hr_talent_calibration_decisions (id,session_id,talent_review_cycle_id,assessment_id,dimension_id,original_score,calibrated_score,rationale_document_version_reference_id,rationale_digest,decided_by_principal_id)
VALUES (pg_temp.u(19),pg_temp.u(18),pg_temp.u(14),pg_temp.u(16),pg_temp.u(12),4,5,pg_temp.u(913),'\x04',pg_temp.u(999));
SELECT pg_temp.expect_error($q$UPDATE hr_talent_calibration_decisions SET calibrated_score=3 WHERE id=pg_temp.u(19)$q$,'23514','keputusan kalibrasi append-only');
INSERT INTO hr_talent_pools (id,pool_code,pool_version,pool_name,legal_entity_reference_id,target_job_id,eligibility_document_version_reference_id,eligibility_digest,valid_from,owner_principal_id)
VALUES (pg_temp.u(20),'STORE-MANAGER',1,'Calon manager toko',pg_temp.u(901),pg_temp.u(3),pg_temp.u(914),'\x05','2026-01-01',pg_temp.u(999));
SELECT pg_temp.expect_error($q$INSERT INTO hr_talent_pool_memberships (id,talent_pool_id,worker_id,assessment_id,nominated_by_principal_id,valid_from) VALUES (pg_temp.u(21),pg_temp.u(20),pg_temp.u(7),pg_temp.u(16),pg_temp.u(998),'2026-11-01')$q$,'23503','pool memakai asesmen pekerja lain');
INSERT INTO hr_talent_pool_memberships (id,talent_pool_id,worker_id,assessment_id,nominated_by_principal_id,valid_from)
VALUES (pg_temp.u(21),pg_temp.u(20),pg_temp.u(6),pg_temp.u(16),pg_temp.u(998),'2026-11-01');
SELECT pg_temp.expect_error($q$UPDATE hr_talent_pool_memberships SET status='active' WHERE id=pg_temp.u(21)$q$,'23514','pool aktif tanpa approval');
UPDATE hr_talent_pool_memberships SET status='active',approved_by_principal_id=pg_temp.u(999),approved_at=now() WHERE id=pg_temp.u(21);
SELECT pg_temp.expect_error($q$INSERT INTO hr_talent_pool_memberships (id,talent_pool_id,worker_id,assessment_id,nominated_by_principal_id,valid_from,status,approved_by_principal_id,approved_at) VALUES (pg_temp.u(22),pg_temp.u(20),pg_temp.u(6),pg_temp.u(16),pg_temp.u(998),'2026-12-01','active',pg_temp.u(999),now())$q$,'23P01','keanggotaan pool aktif bertumpuk');
INSERT INTO hr_career_paths (id,path_code,path_version,path_name,job_family_id,valid_from,approved_by_principal_id,approved_at)
VALUES (pg_temp.u(23),'STORE',1,'Karier toko',pg_temp.u(1),'2026-01-01',pg_temp.u(999),now());
INSERT INTO hr_career_path_transitions (id,career_path_id,from_job_id,to_job_id,movement_type,requirements_document_version_reference_id,requirements_digest)
VALUES (pg_temp.u(24),pg_temp.u(23),pg_temp.u(2),pg_temp.u(3),'promotion',pg_temp.u(915),'\x06');
SELECT pg_temp.expect_error($q$UPDATE hr_career_path_transitions SET to_job_id=from_job_id WHERE id=pg_temp.u(24)$q$,'23514','jalur karier menunjuk jabatan yang sama');
SELECT pg_temp.expect_error($q$INSERT INTO hr_worker_career_aspirations (id,worker_id,target_job_id,career_path_transition_id,preference_rank,recorded_by_principal_id) VALUES (pg_temp.u(25),pg_temp.u(6),pg_temp.u(2),pg_temp.u(24),1,pg_temp.u(996))$q$,'23503','aspirasi dan tujuan jalur berbeda');
INSERT INTO hr_worker_career_aspirations (id,worker_id,target_job_id,career_path_transition_id,preference_rank,recorded_by_principal_id)
VALUES (pg_temp.u(25),pg_temp.u(6),pg_temp.u(3),pg_temp.u(24),1,pg_temp.u(996));

INSERT INTO hr_promotion_recommendations (id,recommendation_number,worker_id,source_assignment_id,target_position_id,assessment_id,recommended_effective_date,rationale_document_version_reference_id,rationale_digest,recommended_by_principal_id,idempotency_key)
VALUES (pg_temp.u(30),'PROMOTE-ANDI',pg_temp.u(6),pg_temp.u(8),pg_temp.u(5),pg_temp.u(16),'2027-01-01',pg_temp.u(916),'\x07',pg_temp.u(998),'promote-andi');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET worker_id=pg_temp.u(7) WHERE id=pg_temp.u(30)$q$,'23503','rekomendasi dengan assignment pekerja lain');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='approved' WHERE id=pg_temp.u(30)$q$,'23514','draft langsung approved');
UPDATE hr_promotion_recommendations SET status='submitted' WHERE id=pg_temp.u(30);
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET target_position_id=pg_temp.u(4) WHERE id=pg_temp.u(30)$q$,'23514','usulan berubah sesudah submit');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='approved',approval_decision_id=pg_temp.u(31) WHERE id=pg_temp.u(30)$q$,'23514','approved tanpa keputusan nyata');
SELECT pg_temp.expect_error($q$INSERT INTO hr_promotion_decisions (id,recommendation_id,step_number,decision,decided_by_principal_id,rationale_document_version_reference_id,rationale_digest,idempotency_key) VALUES (pg_temp.u(31),pg_temp.u(30),1,'approve',pg_temp.u(998),pg_temp.u(917),'\x08','decision-self')$q$,'23514','pengusul menyetujui usulan sendiri');
INSERT INTO hr_promotion_decisions (id,recommendation_id,step_number,decision,decided_by_principal_id,rationale_document_version_reference_id,rationale_digest,idempotency_key)
VALUES (pg_temp.u(31),pg_temp.u(30),1,'defer',pg_temp.u(999),pg_temp.u(917),'\x08','decision-1');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='approved',approval_decision_id=pg_temp.u(31) WHERE id=pg_temp.u(30)$q$,'23514','keputusan defer dianggap approve');
INSERT INTO hr_promotion_decisions (id,recommendation_id,step_number,decision,decided_by_principal_id,rationale_document_version_reference_id,rationale_digest,idempotency_key)
VALUES (pg_temp.u(32),pg_temp.u(30),2,'approve',pg_temp.u(999),pg_temp.u(918),'\x09','decision-2');
UPDATE hr_promotion_recommendations SET status='approved',approval_decision_id=pg_temp.u(32) WHERE id=pg_temp.u(30);
DO $$ BEGIN
  IF (SELECT primary_assignment_id FROM hr_workers WHERE id=pg_temp.u(6)) <> pg_temp.u(8)
    OR (SELECT count(*) FROM hr_worker_assignments WHERE worker_id=pg_temp.u(6)) <> 1 THEN
    RAISE EXCEPTION 'Talent evaluation or approval changed official employment';
  END IF;
  RAISE NOTICE 'PASS POSITIVE: talent dan approval tidak mengubah penempatan resmi';
END $$;
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='applied' WHERE id=pg_temp.u(30)$q$,'23514','applied tanpa bukti penempatan dan event');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_decisions SET decision='reject' WHERE id=pg_temp.u(32)$q$,'23514','keputusan promosi immutable');
-- Simulasi transaksi domain HRM menerapkan promosi yang sudah disetujui.
UPDATE hr_worker_assignments SET valid_to='2026-12-31' WHERE id=pg_temp.u(8);
INSERT INTO hr_worker_assignments (id,worker_id,position_id,assignment_type,valid_from,assignment_reason)
VALUES (pg_temp.u(33),pg_temp.u(6),pg_temp.u(5),'primary','2027-01-01','promotion');
UPDATE hr_workers SET primary_assignment_id=pg_temp.u(33) WHERE id=pg_temp.u(6);
INSERT INTO hr_worker_lifecycle_events (id,worker_id,event_type,occurred_at,effective_date,source_reference,idempotency_key)
VALUES (1,pg_temp.u(6),'promotion',now(),'2027-01-01',pg_temp.u(30)::text,'promoted-andi');
UPDATE hr_worker_lifecycle_events SET event_type='hire' WHERE id=1;
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='applied',applied_assignment_id=pg_temp.u(33),applied_lifecycle_event_id=1,applied_at=now() WHERE id=pg_temp.u(30)$q$,'23514','event hire tidak membuktikan promosi');
UPDATE hr_worker_lifecycle_events SET event_type='promotion',effective_date='2027-01-02' WHERE id=1;
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='applied',applied_assignment_id=pg_temp.u(33),applied_lifecycle_event_id=1,applied_at=now() WHERE id=pg_temp.u(30)$q$,'23514','tanggal event tidak sesuai keputusan promosi');
UPDATE hr_worker_lifecycle_events SET effective_date='2027-01-01' WHERE id=1;
INSERT INTO hr_worker_lifecycle_events (id,worker_id,event_type,occurred_at,effective_date,source_reference,idempotency_key)
VALUES (2,pg_temp.u(7),'promotion',now(),'2027-01-01',pg_temp.u(30)::text,'promoted-wrong-worker');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='applied',applied_assignment_id=pg_temp.u(33),applied_lifecycle_event_id=2,applied_at=now() WHERE id=pg_temp.u(30)$q$,'23503','event promosi milik pekerja lain');
UPDATE hr_promotion_recommendations SET status='applied',applied_assignment_id=pg_temp.u(33),applied_lifecycle_event_id=1,applied_at=now() WHERE id=pg_temp.u(30);
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET status='submitted' WHERE id=pg_temp.u(30)$q$,'23514','status applied dikembalikan');
SELECT pg_temp.expect_error($q$UPDATE hr_promotion_recommendations SET applied_at=now()+interval '1 day' WHERE id=pg_temp.u(30)$q$,'23514','bukti penerapan ditimpa');

INSERT INTO hr_performance_cycles (id,cycle_code,cycle_name,cycle_type,period_start,period_end)
VALUES (pg_temp.u(40),'2026-H2','Semester kedua','semiannual','2026-07-01','2026-12-31');
INSERT INTO hr_performance_goals (id,worker_id,performance_cycle_id,goal_code,goal_title,goal_type)
VALUES (pg_temp.u(41),pg_temp.u(6),pg_temp.u(40),'SALES','Pencapaian toko','business'),
(pg_temp.u(42),pg_temp.u(7),pg_temp.u(40),'SALES','Pencapaian toko','business');
INSERT INTO hr_performance_reviews (id,worker_id,performance_cycle_id,reviewer_principal_id,reviewer_role)
VALUES (pg_temp.u(43),pg_temp.u(6),pg_temp.u(40),pg_temp.u(998),'manager');
SELECT pg_temp.expect_error($q$INSERT INTO hr_performance_review_goal_scores (id,performance_review_id,goal_id,worker_id,performance_cycle_id,rating,weight_percent_snapshot,rubric_document_version_reference_id,evidence_document_version_reference_id) VALUES (pg_temp.u(44),pg_temp.u(43),pg_temp.u(42),pg_temp.u(6),pg_temp.u(40),4,100,pg_temp.u(919),pg_temp.u(920))$q$,'23503','review menilai goal pekerja lain');
INSERT INTO hr_performance_goal_updates (id,goal_id,observed_at,actual_value,evidence_document_version_reference_id,evidence_digest,recorded_by_principal_id,idempotency_key)
VALUES (pg_temp.u(45),pg_temp.u(41),now(),90,pg_temp.u(920),'\x10',pg_temp.u(998),'kpi-1');
SELECT pg_temp.expect_error($q$UPDATE hr_performance_goal_updates SET actual_value=100 WHERE id=pg_temp.u(45)$q$,'23514','hasil KPI historis immutable');
SELECT pg_temp.expect_error($q$INSERT INTO hr_performance_goal_updates (id,goal_id,observed_at,actual_value,evidence_document_version_reference_id,evidence_digest,recorded_by_principal_id,supersedes_update_id,idempotency_key) VALUES (pg_temp.u(46),pg_temp.u(42),now(),100,pg_temp.u(920),'\x10',pg_temp.u(998),pg_temp.u(45),'kpi-2')$q$,'23503','koreksi pengukuran lintas goal');
INSERT INTO hr_competencies (id,competency_code,competency_version,competency_name,category_code,proficiency_scale)
VALUES (pg_temp.u(50),'LEADERSHIP',1,'Kepemimpinan','leadership','{}'),(pg_temp.u(51),'FINANCE',1,'Pengelolaan keuangan','technical','{}');
INSERT INTO hr_learning_courses (id,course_code,course_version,course_name,course_type,delivery_mode)
VALUES (pg_temp.u(52),'STORE-MGMT','1','Manajemen toko','leadership','classroom');
INSERT INTO hr_learning_course_competencies (course_id,competency_id,target_level)
VALUES (pg_temp.u(52),pg_temp.u(50),4),(pg_temp.u(52),pg_temp.u(51),3);
SELECT pg_temp.expect_error($q$INSERT INTO hr_learning_course_competencies (course_id,competency_id,target_level) VALUES (pg_temp.u(52),pg_temp.u(50),5)$q$,'23505','mapping kursus kompetensi ganda');
ROLLBACK;
