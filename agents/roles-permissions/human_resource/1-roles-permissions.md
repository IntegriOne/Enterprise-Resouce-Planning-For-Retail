# Katalog Roles & Permissions: Human Resource Management

Status: **keputusan katalog management** oleh developer HRM atas mandat owner 16 November 2026, dengan [katalog Enterprise Foundation](../enterprise_foundation/1-roles-permissions.md) sebagai pola. Ini menetapkan kode dan batas izin HRM; belum membuat permission atau grant di IAM. Seluruh 22 domain dan 915 RPC gRPC sudah terdaftar, tetapi aktivasi bisnis tetap memerlukan validator, scope dan E2E yang terbukti.

## Prinsip penamaan dan aktivasi

Permission memakai format `human_resource.<resource>.<action>` yang sama dengan Enterprise Foundation. Resource mengikuti nama tabel tanpa awalan `hr_`, bukan string bebas dari request. Contoh: `human_resource.workers.read`, `human_resource.leave_entitlements.post`, `human_resource.payroll_runs.approve`. Untuk operasi biasa, action adalah `read` atau `write`; tiga command sensitif memakai action tersendiri. Mapping contoh: `hr_workers.read`/`hr_workers.list` → `human_resource.workers.read`; `hr_workers.create`/`hr_workers.update`/`hr_workers.delete` → `human_resource.workers.write`.

GetByID, GetAll, FindBy dan count memakai `read` yang sama; filter scope diterapkan sebelum pagination dan count. Create, Update dan Delete memakai `write` yang sama, tetapi setiap operasi tetap diuji terhadap workflow, status final, kepemilikan objek, dan pemeriksaan ulang sebelum commit. `write` tidak otomatis mengizinkan command khusus. Tidak ada permission wildcard dan tidak ada pewarisan hak bisnis otomatis dari role Super Admin IAM.

## Domain dan kelompok resource

| Domain | Resource bisnis (`human_resource.<resource>.<action>`) | Target scope dan batas objek |
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

Inventaris ini dicocokkan dengan `TableName()` entity HRM pada 16 November 2026: 84 resource bisnis pada 22 domain. Setiap nama resource menghasilkan `human_resource.<resource>.read`. Sebanyak 78 resource juga menghasilkan `.write`. Enam resource berikut **hanya memiliki `.read`** karena merupakan riwayat, ledger, atau hasil yang dihasilkan workflow: `worker_lifecycle_events`, `application_stage_history`, `payroll_results`, `payroll_result_lines`, `leave_ledger_entries`, dan `engagement_survey_results`. Total dasar: 162 kode (84 read + 78 write), ditambah tiga command khusus di bawah = **165 kode**. Dua tabel integrasi (`hr_inbox_events`, `hr_outbox_events`) tidak menjadi permission CRUD pengguna. Kode adalah katalog untuk provisioning resmi IAM, bukan seed atau grant.

## Arti tindakan dan pengecualian

| Tindakan | Operasi aplikasi | Batas |
| --- | --- | --- |
| `read` | Get satu record, GetAll, FindBy dan count | Objek dan semua baris harus berada dalam scope actor; count tidak boleh membocorkan data di luar scope. |
| `write` | Create, Update, Delete pada resource yang memiliki `.write` | Field audit dari server; tidak dapat mengubah state final melalui field bebas. Delete selalu sensitif, recheck sebelum commit, tunduk retensi dan referensi. Operasi yang ditolak oleh Guard/workflow tetap ditolak walau permission ada. |

Contoh ekspansi konkret: `workers` menghasilkan `human_resource.workers.read` dan `human_resource.workers.write`. Ledger cuti hanya menghasilkan `human_resource.leave_ledger_entries.read`; penulisan melalui `PostLedgerEntry` dengan permission entitlement `.post`. Hasil payroll, riwayat lifecycle/aplikasi, dan hasil survei juga hanya dibaca pengguna; penulisannya adalah proses internal yang memerlukan workflow dan otorisasi tersendiri. Generic CRUD pada resource tersebut tidak dipublikasikan sebagai mutasi bisnis. `write` payroll tidak memberi hak approve/pay; `write` talent tidak memberi hak finalize. Approval payroll bukan izin pembayaran bank atau posting jurnal Finance.

## Command sensitif yang sudah tersedia di application service

| Command | Kandidat permission | Ketentuan tambahan |
| --- | --- | --- |
| PostLedgerEntry | `human_resource.leave_entitlements.post` | Sensitif: MFA, lock entitlement, periode valid, source/digest, approval adjustment independen, ledger lengkap, outbox atomik. |
| ApprovePayrollRun | `human_resource.payroll_runs.approve` | Sensitif: MFA, actor bukan kalkulator, mandat approval, legal entity, rekonsiliasi dan freeze hasil. |
| FinalizeTalentAssessment | `human_resource.talent_assessments.finalize` | Sensitif: MFA, mandat panel, framework/cycle sama, seluruh skor/evidence lengkap. |

## Kandidat role

Role template hanya kumpulan permission eksplisit, bukan wildcard. Pada matriks berikut, **R** = permission `.read` dan **W** = permission `.write`; Delete termasuk `.write` hanya bila workflow dan retensi mengizinkannya, serta selalu membutuhkan step-up. Materialisasi IAM harus mengekspansi setiap pasangan resource/action secara eksplisit.

| Template role | Resource dan tindakan yang ditetapkan | Scope / pemisahan tugas |
| --- | --- | --- |
| `hrm_catalog_admin` | R/W job_families, jobs, competencies, learning_courses, learning_course_competencies | Katalog perusahaan; tidak otomatis membaca PII/gaji atau asesmen. |
| `hrm_workforce_planner` | R/W positions, workforce_plans, workforce_plan_lines; R jobs/job_families | Legal entity dan unit yang ditugaskan. |
| `hrm_personnel_operator` | R/W workers, employment_contracts, worker_assignments, worker_documents, emergency_contacts; R worker_lifecycle_events | Legal entity/unit; IAM principal link dan lifecycle sensitif memerlukan command resmi. |
| `hrm_recruiter` | R/W requisitions, candidates, applications, interviews, job_postings, offers, application_assessments; R application_stage_history | Requisition/unit; tidak dapat menerima offer atas nama pemberi persetujuan. |
| `hrm_interviewer` | R interviews/applications; W interview_feedback | Hanya interview yang panelnya memuat worker actor; data kandidat minimum. |
| `hrm_lifecycle_coordinator` | R/W onboarding_plans/tasks dan offboarding_plans/tasks | Worker/unit dan PIC tugas; tidak memiliki hak grant/revoke IAM. |
| `hrm_compensation_operator` | R/W compensation_bands/records, benefit_plans/enrollments | Legal entity/worker; MFA; perubahan gaji terpisah dari approval payroll. |
| `hrm_payroll_preparer` | R/W pay_components, payroll_rule_sets, payroll_groups, payroll_periods, payroll_profiles, payroll_runs, payroll_adjustments; R payroll_results/result_lines | Legal entity/group; hanya draft; tidak memiliki approve atau hak menulis hasil kalkulasi langsung. |
| `hrm_payroll_approver` | R payroll_runs/results/result_lines/adjustments/periods; `human_resource.payroll_runs.approve` | Mandat approval terverifikasi; bukan actor kalkulator; MFA dan recheck. |
| `hrm_leave_operator` | R/W leave_policies, leave_entitlements, leave_adjustments; R leave_ledger_entries; `human_resource.leave_entitlements.post` | Worker/legal entity; posting adjustment membutuhkan approval independen dan sumber sah. |
| `hrm_performance_reviewer` | R performance_cycles/goals/goal_updates; R/W performance_reviews/review_goal_scores | Hanya worker/cycle yang ditugaskan; state final tetap dilindungi. |
| `hrm_talent_assessor` | R talent_frameworks/dimensions/review_cycles; R/W talent_assessments/assessment_scores | Mandat panel/cycle; tanpa finalize atau keputusan kalibrasi final. |
| `hrm_talent_panel` | R talent_assessments/assessment_scores; R/W talent_calibration_sessions/decisions, talent_pools/memberships, succession_plans/candidates; `human_resource.talent_assessments.finalize` | Panel/cycle/unit; MFA; tidak berhak menerapkan promosi resmi. |
| `hrm_career_coach` | R/W development_plans/actions, worker_career_aspirations; R career_paths/transitions; R/W promotion_recommendations | Worker yang ditugaskan; tidak mendapat permission `human_resource.promotion_decisions.write` atau `human_resource.worker_assignments.write`. |
| `hrm_learning_operator` | R/W learning_enrollments, worker_qualifications; R learning_courses/competencies | Peserta/unit dan bukti kelulusan; tidak otomatis memberi level kompetensi. |
| `hrm_case_officer` | R/W employee_relations_cases/events, service_requests/events; R collective_agreements | Anggota kasus/PIC dan klasifikasi; bukan akses semua kasus perusahaan. |
| `hrm_policy_operator` | R/W policies; R policy_acknowledgements | Kebijakan berlaku; tidak membuat acknowledgement atas nama pekerja lain. |
| `hrm_engagement_analyst` | R/W engagement_surveys; R engagement_survey_results, human_capital_metric_definitions/snapshots | Hanya agregasi yang memenuhi aturan privasi; akses individual terpisah. |
| `hrm_access_request_operator` | R/W access_requests | Permohonan sesuai scope; keputusan dan grant tetap IAM. |
| `hrm_auditor` | R resource yang disebut secara eksplisit dalam mandat audit | Batas waktu dan purpose; tidak otomatis seluruh HRM/PII. |

Role di atas adalah **template katalog HRM yang diputuskan**, bukan assignment jabatan atau grant aktif. Administrasi katalog talent, jalur karier, definisi metrik dan collective agreement perlu mandat pemilik resource; tidak otomatis diberikan kepada operator data. Delete harus mendapat daftar resource dan retensi tersendiri sebelum grant. Employee self-service memerlukan pemetaan principal→worker yang tervalidasi dan tidak boleh memakai global scope.

## Scope dan operasi sensitif

| Scope IAM yang dipilih | Penggunaan HRM | Pemeriksaan objek wajib |
| --- | --- | --- |
| `global` | Hanya pengelola katalog perusahaan dan auditor dengan mandat eksplisit; payroll/PII global tidak diberikan kepada operator biasa. | Tetap cek status objek, workflow, SoD dan retensi. |
| `legal_entity` | Worker/assignment, payroll, kompensasi, benefit, posisi dan planning dalam badan hukum yang ditugaskan. | Turunkan legal entity dari objek HRM atau referensi Foundation yang tervalidasi; jangan percaya ID yang dikirim klien. |
| `organization_unit` | Recruiter, manager, reviewer, lifecycle coordinator dan talent panel pada unit terkait. | Hubungan worker/posisi/requisition ke unit diverifikasi pada waktu request. |
| `self` | Bacaan karyawan sendiri pada worker, kontrak, leave entitlement, plan pribadi dan acknowledgement yang memang diekspos. | Principal→worker aktif harus dibuktikan lewat link IAM–HRM dan status hubungan kerja. |

Keanggotaan panel interview/talent, PIC tugas dan officer kasus adalah **kondisi objek tambahan**, bukan jenis scope IAM baru. `facility`, `vendor_party`, dan `supplier_party` tidak diberikan untuk data HRM pada baseline ini. Scope `global` mewarisi target yang lebih sempit sesuai pola Enterprise Foundation, tetapi tetap tunduk pada kondisi objek dan larangan SoD.

Operasi yang wajib step-up: seluruh Delete; perubahan kompensasi dan payroll; akses baca gaji/kompensasi dan kasus hubungan industrial yang sensitif; pengaitan atau penghapusan link IAM–worker; posting cuti; approval payroll; finalisasi asesmen talent; keputusan promosi. Kode `.write` tidak dengan sendirinya membuat status final dapat diubah; command dan workflow harus memeriksa state objek serta recheck izin menjelang commit. Applicant/candidate dan dokumen pribadi dibatasi menurut purpose dan retensi.

Gateway menolak request tanpa autentikasi, permission, scope dan step-up sebelum RPC. Service HRM tetap memverifikasi proof internal dan memeriksa permission, scope objek, SoD serta invariant sebagai pemilik data menurut `agents/arsitektur.md` D01–D12 dan `agents/boundary.md`; keputusan katalog ini tidak menghapus pemeriksaan service. Penolakan HMAC internal dari downstream harus diperlakukan Gateway sebagai kegagalan dependency, bukan logout pengguna, mengikuti pola Enterprise Foundation.

Status runtime saat keputusan ini dicatat: adapter HRM masih hanya menerima grant `global`, peta permission production belum diaktifkan, dan validator integrity belum tersedia untuk seluruh mutasi. Grant dengan `legal_entity`, `organization_unit`, atau `self` **harus ditolak** sampai filter list, get, finder/count dan mutasi diuji E2E. Tidak ada role, permission atau grant IAM yang dibuat oleh dokumen ini; provisioning tetap melalui alur resmi IAM.

## Pengujian sebelum aktivasi

Minimal dua identitas: berizin dan terautentikasi tanpa permission. Uji scope lain, UUID objek tebakan, finder/count, deny, expired/revoked session, perubahan revision, Redis/gate unavailable, MFA, SoD dan rollback transaksi. Untuk mutasi child, adapter integrity wajib mengunci parent serta memeriksa status final dan referensi owner dalam transaksi yang sama.

Implementasi policy awal hanya menerima grant global eksplisit; scope parsial dan resource tanpa mapping ditolak. Ini pembatasan sementara, bukan persetujuan memberikan global grant kepada operator.
