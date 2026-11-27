# Pemetaan fungsi HR dan desain skema Retail ERP

Tanggal riset dan verifikasi: **16 November 2026**. Sumber utama: [DBML HRM](../db/schema/human_resource_management.dbml). Path aktif checkout adalah `human_resource`, bukan `human_resouce` atau path historis `human_resource_management`.

## Riset dan arti divisi HR

HR tidak hanya menangani recruitment. [CIPD Profession Map](https://www.cipd.org/en/the-people-profession/the-profession-map/explore-the-profession-map/specialist-knowledge/) memetakan workforce planning, employee experience, employee relations, equality/diversity/inclusion, learning and development, organisation development/design, people analytics, resourcing, reward, talent management dan wellbeing. Ini adalah disiplin/fungsi; pembagian menjadi departemen bergantung kebutuhan perusahaan.

[Workforce planning CIPD](https://www.cipd.org/en/knowledge/factsheets/workforce-planning-factsheet/) menjelaskan hubungan kebutuhan jumlah dan keterampilan pekerja dengan rencana bisnis. Implikasi desain lokal: rencana kebutuhan toko, job, lokasi, FTE, biaya dan tindakan memiliki baris terstruktur, tetapi struktur unit organisasi tetap memakai Foundation. Rencana kebutuhan orang berbeda dari roster harian.

[Pengantar recruitment CIPD](https://www.cipd.org/en/knowledge/factsheets/recruitment-factsheet/) menghubungkan definisi peran, menarik pelamar, seleksi dan pengangkatan. Implikasi desain lokal: lengkapi requisition/application/interview/offer dengan publikasi lowongan dan asesmen yang mempunyai rubrik berversi. Penilaian tidak otomatis menjadi keputusan penolakan kandidat.

Pemetaan tabel berikut merupakan **keputusan desain untuk repositori ini**, bukan struktur tabel yang diwajibkan CIPD, standar sertifikasi, atau keputusan mengaktifkan fitur sekaligus.

## Fungsi bisnis dan tabel

**Keputusan owner 16 November 2026:** Talent Management menjadi domain internal HRM; recruitment/ATS tetap di HRM. Rincian 17 tabel lanjutan dan alur promosi: [Talent Management](talent-management.md).

| Kelompok fungsi | Tabel utama | Perubahan |
|---|---|---|
| Perencanaan kebutuhan dan organisasi | `hr_job_families`, `hr_jobs`, `hr_positions`, `hr_workforce_plans`, `hr_workforce_plan_lines` | Dua tabel rencana/headcount baru; tidak menduplikasi unit organisasi Foundation. |
| Administrasi personalia | `hr_workers`, `hr_employment_contracts`, `hr_worker_assignments`, `hr_worker_lifecycle_events`, `hr_worker_documents`, `hr_emergency_contacts` | Bukti dokumen/kontrak berversi, FTE maksimal 100 per penempatan, larangan overlap primary. |
| Recruitment / talent acquisition | `hr_requisitions`, `hr_candidates`, `hr_applications`, `hr_application_stage_history`, `hr_interviews`, `hr_interview_panelists`, `hr_interview_feedback`, `hr_offers`, `hr_job_postings`, `hr_application_assessments` | Posting dan asesmen baru; sumber lamaran harus posting requisition yang sama; hired memerlukan keputusan dan worker. |
| Onboarding dan offboarding | `hr_onboarding_plans`, `hr_onboarding_tasks`, `hr_offboarding_plans`, `hr_offboarding_tasks` | Offboarding beserta tugas, waiver, konfirmasi owner dan payroll final. |
| Compensation & benefits | `hr_compensation_bands`, `hr_compensation_records`, `hr_benefit_plans`, `hr_benefit_enrollments` | Kompensasi per assignment dan komponen; beberapa tunjangan sejenis dapat dicatat terpisah. |
| Payroll | `hr_payroll_groups`, `hr_payroll_rule_sets`, `hr_payroll_profiles`, `hr_pay_components`, `hr_payroll_periods`, `hr_payroll_runs`, `hr_payroll_results`, `hr_payroll_result_lines`, `hr_payroll_adjustments` | Grup employer/jurisdiction/currency dan manifest aturan immutable baru; total, approval, koreksi, settlement dan snapshot diperketat. |
| Hak cuti | `hr_leave_policies`, `hr_leave_entitlements`, `hr_leave_adjustments`, `hr_leave_ledger_entries` | Ledger sumber saldo; grant, accrual, usage, adjustment, expiry, reversal dengan deduplikasi sumber. |
| Performance & talent management | `hr_performance_cycles`, `hr_performance_goals`, `hr_performance_reviews`, `hr_succession_plans`, `hr_succession_candidates`, `hr_development_plans`, `hr_development_actions` | Rencana pengembangan dan tindakan menggantikan JSON rencana yang sulit ditelusuri. |
| Learning & development dan kompetensi | `hr_learning_courses`, `hr_learning_enrollments`, `hr_worker_qualifications`, `hr_competencies`, `hr_job_competencies`, `hr_worker_competency_assessments` | Katalog rubrik, kebutuhan kompetensi jabatan, asesmen pekerja. |
| Hubungan karyawan/industrial, kebijakan dan layanan HR | `hr_employee_relations_cases`, `hr_employee_relations_events`, `hr_collective_agreements`, `hr_policies`, `hr_policy_acknowledgements`, `hr_service_requests`, `hr_service_request_events` | PKB, kebijakan/acknowledgement, tiket internal dan sejarahnya; FK subject worker diperbaiki. |
| Employee experience, wellbeing, inclusion dan people analytics | `hr_engagement_surveys`, `hr_engagement_survey_results`, `hr_human_capital_metric_definitions`, `hr_human_capital_metric_snapshots` | Survei/hasil agregat tanpa identitas responden; metrik punya versi snapshot dan scope NULL tidak bisa diduplikasi. |

Kelompok teknis terpisah: `hr_access_requests`, `hr_outbox_events`, `hr_inbox_events`. Ketiganya mencatat permintaan/receipt dan pengiriman, **bukan** divisi HR, sistem grant lokal, atau bukti integrasi sudah berjalan.

Perluasan pertama: 45 → 69 tabel. Penyatuan Talent Management menambah 17 tabel sehingga total sekarang **86 tabel**, 9 enum, **130 FK internal**, **150 CHECK DBML** dan 12 TableGroup (11 bisnis + 1 integrasi). Baseline menambah lima exclusion constraint, 14 trigger dan delapan function. Matriks di atas diperlengkap oleh tabel pada dokumen Talent Management.

## Mengapa cakupan ini proporsional

Skema adalah peta kemampuan jangka panjang. Menjalankan seluruh kemampuan sekaligus pada awal F1 akan berlebihan. Penambahan diarahkan pada fakta yang mempunyai identitas, periode, relasi, atau riwayat sendiri, bukan membuat tabel baru untuk setiap nama divisi. Wellbeing/inclusion memakai kebijakan, benefit, kasus berakses terbatas dan hasil agregat; tidak dibuat rekam medis atau master demografi sensitif baru.

Urutan implementasi yang disarankan:

1. **F1 inti:** worker, posisi/assignment, kontrak/dokumen, onboarding/offboarding, integrasi IAM dan owner master.
2. **Operasional HR:** rekrutmen, kompensasi/benefit, saldo cuti dan payroll setelah kontrak Finance/DRM serta parameter aturan disahkan.
3. **Pengembangan organisasi:** workforce planning, kompetensi, performance, learning, suksesi, engagement dan analytics sesuai kebutuhan nyata. Banyak toko kecil cukup memakai sebagian fungsi.

Tidak ada tabel `hr_divisions`: divisi/unit perusahaan tetap `organization_units` Foundation. Tidak dibuat kredensial/role/permission, bank account, jurnal, pembayaran kedua, master party, roster atau clock-in. General Affairs, pengadaan aset, pembelian dan operasi K3 bukan otomatis dipindahkan ke HR karena sering berada pada tim organisasi yang sama.

## Integritas yang sekarang ditegakkan

**DBML/constraint biasa:**

- Kompensasi hanya menunjuk assignment milik worker yang sama. Grain komponen memungkinkan base salary dan beberapa allowance berbeda. Snapshot kode terikat ke master komponen.
- Grup payroll menentukan employer, jurisdiction dan currency. Composite FK mengikat profil, periode, run serta hasil ke grup yang sama dan profil pekerja yang benar.
- Net payroll = gross + reimbursement − deductions. Deductions mencakup pajak pekerja; biaya remunerasi employer = gross + kontribusi employer, tanpa reimbursement. Angka disimpan enam desimal; pembulatan final sesuai currency masih tanggung jawab engine.
- Koreksi/reversal wajib menunjuk run asal dalam grup/mata uang sama. Jumlah hasil merupakan magnitudo nonnegatif; `settlement_direction=payable|recovery` menentukan arah kewajiban. Koreksi yang mencampur arah harus dipisahkan menjadi run terpisah.
- `paid` memerlukan journal, payment batch, event settlement dan waktu settlement Finance. UUID saja bukan bukti keaslian: service wajib memverifikasi event/kontrak Finance.
- Saldo cuti mengikuti persamaan bucket; ledger mempunyai tanda jumlah menurut entry_type, source event/line unik, FK reversal/adjustment ke entitlement yang sama dan satu reversal penuh per entry.
- Metrik menggunakan expression index COALESCE untuk scope kosong, methodology dan snapshot_version. Koreksi adalah snapshot versi baru.
- Rencana pengembangan dalam performance/succession dan event lifecycle dalam access request harus milik pekerja yang sama.

**Tambahan PostgreSQL dalam [postgresql-invariants.sql](../db/schema/postgresql-invariants.sql):**

- Exclusion constraint mencegah overlap primary assignment, kompensasi accepted untuk kode komponen sama, profil payroll per grup, dan entitlement jenis cuti sama, termasuk lintas versi kebijakan. Tanggal efektif inklusif. Pergantian kebijakan di tengah tahun memerlukan pemisahan periode dan pencatatan pemindahan saldo, bukan dua entitlement bertumpuk.
- Ledger cuti dan manifest aturan payroll tidak dapat di-update/delete.
- Approval payroll merekonsiliasi total terhadap semua result lines. Hasil dan baris mengambil lock run saat perubahan; setelah approved, nilai perhitungan dibekukan. Metadata payslip/payment dapat dilengkapi, tetapi bukti yang sudah diisi tidak dapat diganti. Status approved tidak dapat dikembalikan ke draft; koreksi melalui run baru.

[Dokumentasi constraint PostgreSQL](https://www.postgresql.org/docs/18/ddl-constraints.html) membedakan CHECK satu baris dari aturan antarbaris. Karena itu larangan overlap memakai exclusion, sedangkan approval/snapshot memakai trigger. Sintaks dan kemampuan representasi skema mengikuti [dokumentasi DBML](https://dbml.dbdiagram.io/docs/).

## Hal yang tetap wajib pada implementasi service

Keberadaan tabel/trigger tidak menyelesaikan aturan berikut:

- IAM/ABAC, scope legal entity/unit, independensi penyetuju, authentikasi sumber event, recheck sebelum commit dan katalog role yang disetujui owner. `hr_access_requests.status` tidak pernah menjadi dasar ALLOW.
- Posting ledger cuti dan pembaruan proyeksi dalam transaksi dengan row lock; cek periode, saldo cukup, aturan accrual, UOM, reversal tepat lawan jumlah asal, serta kesesuaian jumlah adjustment yang disetujui. CHECK persamaan bucket **belum** membuktikan kesamaan semua bucket dengan SUM ledger. Service harus menolak pemakaian saat owner absensi atau bukti sumber tidak tersedia.
- Validitas pekerja, assignment, jabatan, grup dan kontrak pada tanggal efektif; batas total FTE/kapasitas posisi; siklus hierarki; legal entity PKB sama dengan kontrak; penggunaan profil dan aturan payroll yang berlaku pada periode run; original run koreksi telah disahkan dan tidak membentuk siklus.
- Currency scale/rounding, kurva pajak, THR, BPJS, prorata, lembur terverifikasi dari owner waktu kerja, formula, tanggal payroll, konsumsi adjustment tepat sekali, dan kasus pajak lokal. Tidak ada tarif atau klaim kepatuhan Indonesia yang di-hardcode.
- Penyelesaian semua task wajib offboarding, validitas waiver, konfirmasi IAM dan Finance. `final_payroll_run_id` memerlukan pemeriksaan bahwa pekerja terkait benar-benar termasuk dalam run.
- Rubrik kompetensi yang lengkap dan konsisten antarversi; kelayakan assessment, proses banding, kebijakan retensi dan akses dokumen/kasus.
- Survei hanya menyimpan hasil agregat; ambang minimal lima pada DB adalah batas desain produk, bukan angka hukum atau jaminan anonim. Service harus menerapkan ambang survei yang dapat lebih tinggi, pembatasan irisan/filter serta otorisasi hasil.
- Worker outbox/inbox durable, retry, event ordering, pemeriksaan digest saat idempotency key berulang dan pemulihan crash. Tabel tidak berarti relay/consumer sudah berjalan.

Tidak ada Go service, kontrak HRM, endpoint atau worker baru dalam perubahan ini. Tidak ada database bisnis yang dimigrasikan. Desain harus melewati service tests, concurrency tests dan E2E GraphQL dengan IAM/owner nyata sebelum fitur disebut siap.

## Generasi dan verifikasi

Baseline pra-rilis [000001_hrm.up.sql](../db/migrations/000001_hrm.up.sql) dan [000001_hrm.down.sql](../db/migrations/000001_hrm.down.sql) mencakup provenance dan aturan pemisahan tugas. Database yang menyimpan checksum baseline lama memerlukan persiapan ulang sesuai prosedur development. Pengajuan lama tanpa provenance yang terbukti tidak dapat disetujui.

`generate.sh` sekarang memvalidasi DBML dengan CLI `@dbml/cli@12.0.0` tanpa menulis migrasi historis. Argumen path opsional mengekspor skema lengkap saat ini untuk inspeksi; output tidak boleh berada dalam folder migrasi. Skrip ini tidak menghasilkan migrasi delta.

```bash
bash backend/services/human_resource/db/schema/generate.sh
bash backend/services/human_resource/db/test/verify_schema.sh
```

Verifier menggunakan service PostgreSQL resmi Docker Compose, membuat nama unik `hrm_schema_*_test`, menjalankan up → kasus integritas → down → up, lalu menghapus hanya database yang berhasil dibuat verifier. Tidak mencetak secret, tidak menyentuh database management lain, tidak membuat service PostgreSQL host.

Hasil aktual: PostgreSQL Docker **18.6** (bukan mengasumsikan 18.7 dari dokumentasi); **70 skenario negatif lulus** (38 integritas HRM + 32 Talent Management) dan fixture positif lulus. Termasuk primary bersebelahan/secondary paralel, kompensasi komponen berbeda, payroll bayar sebelum akhir bulan, reimbursement yang membuat net melebihi gross, approval/posted/paid valid, serta versi metrik baru. Ini pengujian integritas SQL, bukan E2E GraphQL atau benchmark. `docker compose config --quiet` lulus. Tidak ada image HRM/runtime Go pada folder ini untuk dibuild; deployment service bersama tidak diubah.

## Catatan skill Database Engineer

- `usp`: `cognitive_schema_mapper`.
- `source`: `analysis`.
- `ontology`: fungsi HR → kelompok bisnis → tabel/relasi pada matriks di atas.
- `semanticGaps`: publikasi/asesmen recruitment, perencanaan tenaga kerja, kompetensi, offboarding, ledger cuti, pengembangan, kebijakan/layanan, PKB dan engagement yang sebelumnya belum direpresentasikan penuh.
- `recommendations`: perluasan pertama 24 tabel, dilanjutkan 17 tabel Talent Management atas keputusan owner dan penguatan integritas; implementasi bertahap mengikuti dependency, tanpa pemindahan ownership.
- `evidence`: sumber CIPD, boundary/principals, DBML sebelum/sesudah, export SQL, dan tes PostgreSQL yang dapat diulang.
- `confidence`: tinggi untuk cakupan struktur dan constraint yang diuji; menengah untuk prioritas fitur yang belum divalidasi dengan proses HR perusahaan; belum dinilai untuk aturan pajak/hukum, workload/performa dan kesiapan runtime.
