# Human Resource Management

Status: **entity tervalidasi, aturan domain, port dan repository GORM tersedia**. HRM dan Talent Management berada dalam satu service dengan master `hr_workers` yang sama. Recruitment/ATS termasuk HRM.

Skema, baseline SQL, validasi domain dan repository 86 tabel tersedia dalam modul Go mandiri. Lapisan aplikasi/transport masih placeholder; belum ada proses service, kontrak RPC, handler, otorisasi, atau deployment HRM. Struktur mengikuti lapisan service IAM pada checkout aktif, bukan path historis `human_resource_management`.

## Struktur lapisan

```text
human_resource/
├── cmd/                         # Titik masuk service
├── config/                      # Konfigurasi dan validasinya
├── contract/
│   ├── proto/human_resource/    # Kontrak milik HRM, menunggu implementasi
│   └── gen/                    # Keluaran generator resmi
├── db/
│   ├── schema/                 # DBML, invariant PostgreSQL dan generator
│   ├── migrations/             # Baseline tunggal 000001
│   └── test/                   # Tes integritas SQL yang sudah tersedia
├── domain/
│   ├── entity/                 # Entitas, value object dan state domain
│   ├── errors/                 # Kode dan pesan error domain
│   ├── repository/             # Interface persistensi milik HRM
│   └── service/                # Aturan domain dan port dependency
├── repository/
│   ├── pgstore/                # Transaksi dan persistensi PostgreSQL HRM
│   ├── inbox/                  # Persistensi penerimaan event
│   ├── outbox/                 # Persistensi pengiriman event
│   └── <domain>/               # Implementasi repository per domain
├── services/
│   └── <domain>/               # Use case dan koordinasi transaksi
├── infrastructure/
│   ├── integration/
│   │   ├── iam/                # Kontrak identitas/akses milik IAM
│   │   ├── enterprise_foundation/ # Referensi party, unit dan badan hukum
│   │   ├── finance/            # Kontrak jurnal dan settlement
│   │   └── document_records/   # Kontrak dokumen dan versi bukti
│   ├── inbox/                  # Consumer, deduplikasi dan retry
│   └── outbox/                 # Relay dan pengiriman event
├── transport/grpc/
│   ├── handler/                # Pemetaan RPC ke use case
│   └── interceptor/            # Konteks pemanggil dan guard transport
├── test/
│   ├── integration/            # Rencana tes transaksi/dependency
│   └── e2e/                    # Rencana tes alur HRM melalui Gateway
└── docs/                       # Desain dan batas implementasi
```

Placeholder `<domain>` di diagram berarti folder nyata yang tercantum pada tabel berikut. Unit test nantinya mengikuti package/lapisan terkait. Folder `test/` masih placeholder. Unit test entity berada di `domain/entity/test/`; aturan agregat diuji di `domain/service/<domain>/test/`.

## Domain yang disiapkan

Setiap nama di bawah mempunyai folder pada **`services/` dan `repository/`**. Pengelompokan ini adalah batas tanggung jawab implementasi dalam satu service; bukan microservice atau database terpisah. Interface/entitas berada pada `domain/`, mengikuti pola service yang ada.

| Folder domain | Tabel DBML yang menjadi tanggung jawabnya |
|---|---|
| `accessrequest` | `hr_access_requests` |
| `analytics` | `hr_human_capital_metric_definitions`, `hr_human_capital_metric_snapshots` |
| `benefits` | `hr_benefit_plans`, `hr_benefit_enrollments` |
| `career` | `hr_career_paths`, `hr_career_path_transitions`, `hr_worker_career_aspirations`, `hr_development_plans`, `hr_development_actions` |
| `compensation` | `hr_compensation_bands`, `hr_compensation_records` |
| `competency` | `hr_competencies`, `hr_job_competencies`, `hr_worker_competency_assessments`, `hr_worker_qualifications` |
| `employeerelations` | `hr_employee_relations_cases`, `hr_employee_relations_events`, `hr_collective_agreements` |
| `engagement` | `hr_engagement_surveys`, `hr_engagement_survey_results` |
| `hrservice` | `hr_service_requests`, `hr_service_request_events` |
| `learning` | `hr_learning_courses`, `hr_learning_enrollments`, `hr_learning_course_competencies` |
| `leave` | `hr_leave_policies`, `hr_leave_entitlements`, `hr_leave_adjustments`, `hr_leave_ledger_entries` |
| `offboarding` | `hr_offboarding_plans`, `hr_offboarding_tasks` |
| `onboarding` | `hr_onboarding_plans`, `hr_onboarding_tasks` |
| `payroll` | `hr_payroll_groups`, `hr_payroll_rule_sets`, `hr_payroll_profiles`, `hr_pay_components`, `hr_payroll_periods`, `hr_payroll_runs`, `hr_payroll_results`, `hr_payroll_result_lines`, `hr_payroll_adjustments` |
| `performance` | `hr_performance_cycles`, `hr_performance_goals`, `hr_performance_reviews`, `hr_performance_goal_updates`, `hr_performance_review_goal_scores` |
| `personnel` | `hr_workers`, `hr_employment_contracts`, `hr_worker_assignments`, `hr_worker_lifecycle_events`, `hr_worker_documents`, `hr_emergency_contacts` |
| `policy` | `hr_policies`, `hr_policy_acknowledgements` |
| `promotion` | `hr_promotion_recommendations`, `hr_promotion_decisions` |
| `recruitment` | `hr_requisitions`, `hr_candidates`, `hr_applications`, `hr_application_stage_history`, `hr_interviews`, `hr_interview_panelists`, `hr_interview_feedback`, `hr_offers`, `hr_job_postings`, `hr_application_assessments` |
| `succession` | `hr_succession_plans`, `hr_succession_candidates` |
| `talent` | `hr_talent_frameworks`, `hr_talent_dimensions`, `hr_talent_review_cycles`, `hr_talent_assessments`, `hr_talent_assessment_scores`, `hr_talent_calibration_sessions`, `hr_talent_calibration_decisions`, `hr_talent_pools`, `hr_talent_pool_memberships` |
| `workforceplanning` | `hr_job_families`, `hr_jobs`, `hr_positions`, `hr_workforce_plans`, `hr_workforce_plan_lines` |

`hr_outbox_events` dan `hr_inbox_events` ditangani repository/infrastruktur khusus tersebut. Pemetaan mencakup seluruh **86 tabel** tanpa kepemilikan tabel ganda. Nama subdomain merupakan pedoman awal; nama package protobuf dan versi kontrak ditetapkan saat kontrak benar-benar dirancang.

Talent Management mencakup `competency`, `performance`, `talent`, `career`, `succession`, `promotion`, dan `learning`. Rekrutmen tetap pada `recruitment`; perubahan resmi hubungan kerja tetap pada `personnel`.

## Alur dan batas implementasi

- Handler gRPC memetakan kontrak ke `services`; service menggunakan entitas/aturan/interface `domain`; repository dan adapter menyediakan implementasi dependency. `domain` tidak bergantung pada transport atau implementasi infrastruktur.
- Operasi lintas subdomain HRM, misalnya penerapan promosi, harus mengubah assignment, lifecycle event, status rekomendasi dan outbox dalam satu transaksi. Approval rekomendasi belum berarti jabatan berubah.
- `personnel` mengelola tautan identitas pada `hr_workers.iam_principal_reference_id`. Binding akun dan lookup pekerja harus memakai kontrak IAM yang disepakati, dengan konteks legal entity dan hubungan kerja yang tepat. Worker ID bukan user ID; rehire dapat membentuk hubungan kerja baru. Kredensial, login, session, role dan grant tetap dimiliki IAM.
- `accessrequest` mengelola permintaan/proyeksi akses HRM. Status lokal tidak menjadi izin otomatis. Adapter integrasi berkomunikasi lewat kontrak owner; tidak memakai SQL, FK atau repository management lain.
- Master party/unit/lokasi tetap Foundation; jurnal dan pembayaran tetap Finance; dokumen berversi/retensi tetap Document Records. Folder integrasi tidak membuktikan kontrak dependency tersedia.
- Hak/saldo cuti ada di `leave`. Roster/shift/absensi operasional tidak diambil alih oleh scaffold ini.
- API publik tetap `/api/graphql` melalui Gateway; gRPC privat. Wiring Gateway, aktivasi modul dan katalog permission memerlukan implementasi serta verifikasi tersendiri.

## Urutan development

1. Lanjutkan config/DB bootstrap dan kontrak HRM sesuai kebijakan proyek; selaraskan kontrak identitas IAM dan referensi Foundation terlebih dahulu.
2. Kerjakan `personnel`, job/position pada `workforceplanning`, onboarding/offboarding dan permintaan akses dengan pengujian otorisasi serta scope.
3. Tambahkan recruitment, cuti, kompensasi/benefit/payroll dan domain Talent secara bertahap sesuai prioritas owner.
4. Implementasikan integrasi, transaksi outbox/inbox, uji concurrency dan E2E sebelum aktivasi Gateway.

[Desain skema](docs/schema-design.md) · [Talent Management](docs/talent-management.md) · [DBML](db/schema/human_resource_management.dbml) · [Boundary](../../../agents/boundary.md) · [Arsitektur](../../../agents/arsitektur.md)

[Kontrak repository, pagination dan verifikasi](docs/repository.md).

[Validasi domain, aturan agregat dan wiring writer](docs/domain.md).
