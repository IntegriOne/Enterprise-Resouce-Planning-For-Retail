# Draf katalog akses HRM

Status: **usulan untuk review owner**, berdasarkan instruksi 16 November 2026. Belum aktif di IAM, belum ada seed/grant. Daftar domain service bukan bukti seluruh RPC sudah tersedia; transport sedang diselesaikan. Setiap permission baru hanya boleh diaktifkan setelah RPC, validator, scope dan E2E terkait terbukti.

## Prinsip penamaan dan aktivasi

Permission memakai `hrm.<resource>:<action>`; resource mengikuti nama tabel tanpa awalan `hr_`, bukan string bebas dari request. Contoh: `hrm.workers:read`, `hrm.leave_entitlements:post`, `hrm.payroll_runs:approve`. Mapping internal service `hr_workers.read` → `hrm.workers:read` disimpan pada konfigurasi server setelah katalog disetujui.

Setiap resource mendapatkan permission terpisah untuk `read`, `list`, `create`, `update`, `delete` yang memang didukung workflow. Finder memakai `list`; count mengikuti scope yang sama dengan list. Keberadaan method CRUD tidak mengizinkan bypass status final, approval, ledger immutable atau guard lifecycle. Tidak ada permission wildcard dan tidak ada pewarisan hak bisnis otomatis dari role superadmin IAM.

## Domain dan kelompok resource

| Domain | Resource bisnis | Batas akses yang perlu dibuktikan |
| --- | --- | --- |
| workforceplanning | job_families, jobs, positions, workforce_plans, workforce_plan_lines | Legal entity/unit untuk posisi dan rencana; master global khusus pengelola katalog. |
| personnel | workers, employment_contracts, worker_assignments, worker_lifecycle_events, worker_documents, emergency_contacts | Legal entity/unit/worker; PII sensitif; account IAM tidak dapat ditautkan lewat CRUD umum. |
| recruitment | requisitions, candidates, applications, application_stage_history, interviews, interview_panelists, interview_feedback, offers, job_postings, application_assessments | Requisition/unit dan keanggotaan panel; kandidat eksternal tidak otomatis menjadi employee. |
| onboarding | onboarding_plans, onboarding_tasks | Worker/unit dan penanggung jawab tugas. |
| compensation | compensation_bands, compensation_records | Legal entity/worker; seluruh mutasi sensitif. |
| benefits | benefit_plans, benefit_enrollments | Legal entity/worker; eligibilitas dan periode. |
| payroll | pay_components, payroll_rule_sets, payroll_groups, payroll_periods, payroll_profiles, payroll_runs, payroll_results, payroll_result_lines, payroll_adjustments | Legal entity/group/period; hasil final immutable; approval independen dari kalkulator. |
| leave | leave_policies, leave_entitlements, leave_adjustments, leave_ledger_entries | Worker/legal entity; saldo hanya melalui posting ledger dan approval yang sah. |
| competency | competencies, worker_qualifications, job_competencies, worker_competency_assessments | Master katalog terpisah dari asesmen worker; evidence tervalidasi. |
| performance | performance_cycles, performance_goals, performance_goal_updates, performance_reviews, performance_review_goal_scores | Worker/unit dan evaluator sah; review final tidak dapat diubah melalui CRUD. |
| learning | learning_courses, learning_course_competencies, learning_enrollments | Worker, peserta, penyelenggara dan kompetensi terkait. |
| succession | succession_plans, succession_candidates | Posisi dan panel talent; informasi potensi sensitif. |
| employeerelations | employee_relations_cases, employee_relations_events, collective_agreements | Anggota kasus berwenang; narasi dan identitas tidak tersedia bagi seluruh staf HR. |
| offboarding | offboarding_plans, offboarding_tasks | Worker/unit dan PIC; revoke akses tetap keputusan IAM. |
| career | development_plans, development_actions, career_paths, career_path_transitions, worker_career_aspirations | Worker/unit, coach dan rencana yang relevan. |
| policy | policies, policy_acknowledgements | Kebijakan yang berlaku; acknowledgement milik worker, tidak boleh diwakilkan tanpa mandat. |
| hrservice | service_requests, service_request_events | Pemohon, PIC dan klasifikasi kasus. |
| engagement | engagement_surveys, engagement_survey_results | Pemilik survei; hasil individual sensitif dan agregasi tidak boleh membocorkan identitas. |
| accessrequest | access_requests | Pemohon/PIC; keputusan/grant akhir hanya IAM. |
| analytics | human_capital_metric_definitions, human_capital_metric_snapshots | Dimensi legal entity/unit; metrik tidak memperluas hak data sumber. |
| talent | talent_frameworks, talent_dimensions, talent_review_cycles, talent_assessments, talent_assessment_scores, talent_calibration_sessions, talent_calibration_decisions, talent_pools, talent_pool_memberships | Panel/cycle/unit; asesmen final, kalibrasi dan evidence dipisahkan. |
| promotion | promotion_recommendations, promotion_decisions | Worker, target assignment dan approval independen; approval tidak langsung mengubah jabatan. |

Inventaris ini dicocokkan dengan `TableName()` entity HRM pada 16 November 2026: 84 resource bisnis pada 22 domain. Dua tabel integrasi (`hr_inbox_events`, `hr_outbox_events`) tidak menjadi permission CRUD pengguna. Pemetaan ulang tetap wajib bila kontrak berubah; tabel ini bukan seed yang siap dieksekusi.

## Arti tindakan dan pengecualian

| Tindakan | Operasi aplikasi | Batas |
| --- | --- | --- |
| `read` | Get satu record berdasarkan key | Objek harus berada dalam scope actor. |
| `list` | GetAll, FindBy kolom dan count | Filter scope dijalankan sebelum pagination dan count; izin read tidak menyiratkan list. |
| `create` | Membuat record awal | Field audit dari server; tidak dapat langsung membuat state final. |
| `update` | Patch field bisnis yang diperbolehkan | Tidak mengubah owner/approval/status final melalui field bebas. |
| `delete` | Menghapus record yang masih boleh dihapus | Sensitif, recheck sebelum commit; retensi dan referensi tetap berlaku. |

Contoh ekspansi konkret: resource `workers` dengan read/list menjadi `hrm.workers:read` dan `hrm.workers:list`, bukan `hrm.workers:*`. Tidak semua resource diberi lima tindakan. Untuk `leave_ledger_entries`, hanya read/list diusulkan; penulisan melalui command `PostLedgerEntry` dengan permission entitlement `post`. Update/delete ledger tidak diusulkan. CRUD payroll tidak memberi hak approve/post/pay; CRUD talent tidak memberi hak finalize. Permission `approve` payroll bukan izin pembayaran bank atau posting jurnal Finance.

## Command sensitif yang sudah tersedia di application service

| Command | Kandidat permission | Ketentuan tambahan |
| --- | --- | --- |
| PostLedgerEntry | `hrm.leave_entitlements:post` | Lock entitlement, periode valid, source/digest, approval adjustment independen, ledger lengkap, outbox atomik. |
| ApprovePayrollRun | `hrm.payroll_runs:approve` | MFA; actor bukan kalkulator; mandat approval, legal entity, rekonsiliasi dan freeze hasil. |
| FinalizeTalentAssessment | `hrm.talent_assessments:finalize` | MFA; mandat panel, framework/cycle sama, seluruh skor/evidence lengkap. |

## Kandidat role

Role hanya kumpulan permission eksplisit, bukan wildcard. Pada matriks berikut, **R** = read/list dan **W** = create/update terhadap draft yang diperbolehkan workflow. Notasi ini hanya untuk review; materialisasi katalog harus mengekspansi setiap pasangan resource/action secara eksplisit. Delete tidak termasuk W.

| Kandidat role | Usulan resource dan tindakan | Scope yang diusulkan / pemisahan tugas |
| --- | --- | --- |
| `hrm_catalog_admin` | R/W job_families, jobs, competencies, learning_courses, learning_course_competencies | Katalog perusahaan; tidak otomatis membaca PII/gaji atau asesmen. |
| `hrm_workforce_planner` | R/W positions, workforce_plans, workforce_plan_lines; R jobs/job_families | Legal entity dan unit yang ditugaskan. |
| `hrm_personnel_operator` | R/W workers, employment_contracts, worker_assignments, worker_documents, emergency_contacts; R worker_lifecycle_events | Legal entity/unit; IAM principal link dan lifecycle sensitif memerlukan command resmi. |
| `hrm_recruiter` | R/W requisitions, candidates, applications, interviews, job_postings, offers, application_assessments; R application_stage_history | Requisition/unit; tidak dapat menerima offer atas nama pemberi persetujuan. |
| `hrm_interviewer` | R interviews/applications; W interview_feedback | Hanya interview yang panelnya memuat worker actor; data kandidat minimum. |
| `hrm_lifecycle_coordinator` | R/W onboarding_plans/tasks dan offboarding_plans/tasks | Worker/unit dan PIC tugas; tidak memiliki hak grant/revoke IAM. |
| `hrm_compensation_operator` | R/W compensation_bands/records, benefit_plans/enrollments | Legal entity/worker; MFA; perubahan gaji terpisah dari approval payroll. |
| `hrm_payroll_preparer` | R/W pay_components, payroll_rule_sets, payroll_groups, payroll_periods, payroll_profiles, payroll_runs, payroll_results, payroll_result_lines, payroll_adjustments | Legal entity/group; hanya draft; tidak memiliki approve. |
| `hrm_payroll_approver` | R payroll_runs/results/result_lines/adjustments/periods; payroll_runs:approve | Mandat approval terverifikasi; bukan actor kalkulator; MFA dan recheck. |
| `hrm_leave_operator` | R/W leave_policies, leave_entitlements, leave_adjustments; R leave_ledger_entries; leave_entitlements:post | Worker/legal entity; posting adjustment membutuhkan approval independen dan sumber sah. |
| `hrm_performance_reviewer` | R performance_cycles/goals/goal_updates; R/W performance_reviews/review_goal_scores | Hanya worker/cycle yang ditugaskan; state final tetap dilindungi. |
| `hrm_talent_assessor` | R talent_frameworks/dimensions/review_cycles; R/W talent_assessments/assessment_scores | Mandat panel/cycle; tanpa finalize atau keputusan kalibrasi final. |
| `hrm_talent_panel` | R talent_assessments/assessment_scores; R/W talent_calibration_sessions/decisions, talent_pools/memberships, succession_plans/candidates; talent_assessments:finalize | Panel/cycle/unit; MFA; tidak berhak menerapkan promosi resmi. |
| `hrm_career_coach` | R/W development_plans/actions, worker_career_aspirations; R career_paths/transitions; R/W promotion_recommendations | Worker yang ditugaskan; tidak mendapat promotion_decisions atau worker_assignments:update. |
| `hrm_learning_operator` | R/W learning_enrollments, worker_qualifications; R learning_courses/competencies | Peserta/unit dan bukti kelulusan; tidak otomatis memberi level kompetensi. |
| `hrm_case_officer` | R/W employee_relations_cases/events, service_requests/events; R collective_agreements | Anggota kasus/PIC dan klasifikasi; bukan akses semua kasus perusahaan. |
| `hrm_policy_operator` | R/W policies; R policy_acknowledgements | Kebijakan berlaku; tidak membuat acknowledgement atas nama pekerja lain. |
| `hrm_engagement_analyst` | R/W engagement_surveys; R engagement_survey_results, human_capital_metric_definitions/snapshots | Hanya agregasi yang memenuhi aturan privasi; akses individual terpisah. |
| `hrm_access_request_operator` | R/W access_requests | Permohonan sesuai scope; keputusan dan grant tetap IAM. |
| `hrm_auditor` | R resource yang disebut secara eksplisit dalam mandat audit | Batas waktu dan purpose; tidak otomatis seluruh HRM/PII. |

Role di atas bukan keputusan final struktur jabatan perusahaan. Administrasi katalog talent, jalur karier, definisi metrik dan collective agreement perlu mandat pemilik resource; tidak otomatis diberikan kepada operator data. Delete harus mendapat daftar resource dan retensi tersendiri sebelum grant. Employee self-service memerlukan pemetaan principal→worker yang tervalidasi dan tidak boleh memakai global scope.

## Keputusan yang diminta saat review

1. Setujui atau koreksi format permission dan kelompok resource; setujui tindakan hanya bila proses bisnisnya memang dipakai.
2. Tentukan role yang dipakai perusahaan, legal entity/unit/panel/case scope, serta pemilik persetujuan grant.
3. Tetapkan mandat approval dan pemisahan tugas untuk payroll, cuti, talent dan promosi, termasuk daftar operasi yang wajib MFA.
4. Tetapkan kebutuhan self-service, akses PII, retensi, dan siapa yang boleh delete per resource.

Persetujuan dokumen tidak membuat grant IAM. Aktivasi memerlukan implementasi scope yang disetujui, validator, pengujian dan proses grant IAM terpisah.

## Pengujian sebelum aktivasi

Minimal dua identitas: berizin dan terautentikasi tanpa permission. Uji scope lain, UUID objek tebakan, finder/count, deny, expired/revoked session, perubahan revision, Redis/gate unavailable, MFA, SoD dan rollback transaksi. Untuk mutasi child, adapter integrity wajib mengunci parent serta memeriksa status final dan referensi owner dalam transaksi yang sama.

Implementasi policy awal hanya menerima grant global eksplisit yang sudah disetujui; scope parsial dan resource tanpa mapping ditolak. Ini pembatasan sementara, bukan persetujuan memberikan global grant kepada operator. Tidak ada role atau grant baru dibuat oleh dokumen ini.
