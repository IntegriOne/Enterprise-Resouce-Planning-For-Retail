# List operasi

`human_resource.workforceplanning.v1.WorkforcePlanningService` memiliki 56 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/workforceplanning/v1/workforceplanning.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateJobFamily` | Membuat entri `job_families`. | Mutation | `CreateJobFamilyRequest` → `JobFamilyResponse` | `human_resource.job_families.create` |
| `UpdateJobFamily` | Mengubah entri `job_families` sesuai field mask. | Mutation | `UpdateJobFamilyRequest` → `JobFamilyResponse` | `human_resource.job_families.update` |
| `DeleteJobFamily` | Menghapus entri `job_families` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.job_families.delete` |
| `GetJobFamilyByID` | Mengambil entri `job_families` berdasarkan ID. | Query | `IDRequest` → `JobFamilyResponse` | `human_resource.job_families.read` |
| `GetAllJobFamilies` | Mendaftar entri `job_families` dengan paginasi. | Query | `PageRequest` → `JobFamiliesResponse` | `human_resource.job_families.read` |
| `FindJobFamilyByFamilyCode` | Menyaring daftar `job_families` berdasarkan `family_code`, dengan paginasi. | Query | `StringFinderRequest` → `JobFamiliesResponse` | `human_resource.job_families.read` |
| `CreateJob` | Membuat entri `jobs`. | Mutation | `CreateJobRequest` → `JobResponse` | `human_resource.jobs.create` |
| `UpdateJob` | Mengubah entri `jobs` sesuai field mask. | Mutation | `UpdateJobRequest` → `JobResponse` | `human_resource.jobs.update` |
| `DeleteJob` | Menghapus entri `jobs` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.jobs.delete` |
| `GetJobsByIds` | Membaca sekumpulan jabatan dari ID yang dirujuk UI. | Query | `GetJobsByIdsRequest` → `JobsResponse` | `human_resource.jobs.read` |
| `GetJobByID` | Mengambil entri `jobs` berdasarkan ID. | Query | `IDRequest` → `JobResponse` | `human_resource.jobs.read` |
| `GetAllJobs` | Mendaftar entri `jobs` dengan paginasi. | Query | `PageRequest` → `JobsResponse` | `human_resource.jobs.read` |
| `FindJobByJobFamilyID` | Menyaring daftar `jobs` berdasarkan `job_family_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobsResponse` | `human_resource.jobs.read` |
| `FindJobByJobCode` | Menyaring daftar `jobs` berdasarkan `job_code`, dengan paginasi. | Query | `StringFinderRequest` → `JobsResponse` | `human_resource.jobs.read` |
| `FindJobByJobLevelCode` | Menyaring daftar `jobs` berdasarkan `job_level_code`, dengan paginasi. | Query | `StringFinderRequest` → `JobsResponse` | `human_resource.jobs.read` |
| `CreatePosition` | Membuat entri `positions`. | Mutation | `CreatePositionRequest` → `PositionResponse` | `human_resource.positions.create` |
| `ActivatePosition` | Mengaktifkan entri `positions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PositionResponse` | `human_resource.positions.activate` |
| `UpdatePosition` | Mengubah entri `positions` sesuai field mask. | Mutation | `UpdatePositionRequest` → `PositionResponse` | `human_resource.positions.update` |
| `DeletePosition` | Menghapus entri `positions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.positions.delete` |
| `GetPositionByID` | Mengambil entri `positions` berdasarkan ID. | Query | `IDRequest` → `PositionResponse` | `human_resource.positions.read` |
| `GetAllPositions` | Mendaftar entri `positions` dengan paginasi. | Query | `PageRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByPositionCode` | Menyaring daftar `positions` berdasarkan `position_code`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByJobID` | Menyaring daftar `positions` berdasarkan `job_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByOrganizationUnitReferenceID` | Menyaring daftar `positions` berdasarkan `organization_unit_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByLegalEntityReferenceID` | Menyaring daftar `positions` berdasarkan `legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByReportsToPositionID` | Menyaring daftar `positions` berdasarkan `reports_to_position_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByPositionStatus` | Menyaring daftar `positions` berdasarkan `position_status`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByWorkLocationReferenceID` | Menyaring daftar `positions` berdasarkan `work_location_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `FindPositionByCostCenterReferenceID` | Menyaring daftar `positions` berdasarkan `cost_center_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PositionsResponse` | `human_resource.positions.read` |
| `CancelWorkforcePlan` | Membatalkan entri `workforce_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.cancel` |
| `SupersedeWorkforcePlan` | Menggantikan entri `workforce_plans` melalui transisi workflow HRM. | Mutation | `SupersedeWorkforcePlanRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.supersede` |
| `ApproveWorkforcePlan` | Menyetujui entri `workforce_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.approve` |
| `SubmitWorkforcePlan` | Mengajukan entri `workforce_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.submit` |
| `CreateWorkforcePlan` | Membuat entri `workforce_plans`. | Mutation | `CreateWorkforcePlanRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.create` |
| `UpdateWorkforcePlan` | Mengubah entri `workforce_plans` sesuai field mask. | Mutation | `UpdateWorkforcePlanRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.update` |
| `DeleteWorkforcePlan` | Menghapus entri `workforce_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.workforce_plans.delete` |
| `GetWorkforcePlanByID` | Mengambil entri `workforce_plans` berdasarkan ID. | Query | `IDRequest` → `WorkforcePlanResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `GetAllWorkforcePlans` | Mendaftar entri `workforce_plans` dengan paginasi. | Query | `PageRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByPlanCode` | Menyaring daftar `workforce_plans` berdasarkan `plan_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByLegalEntityReferenceID` | Menyaring daftar `workforce_plans` berdasarkan `legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByScenarioCode` | Menyaring daftar `workforce_plans` berdasarkan `scenario_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByStatus` | Menyaring daftar `workforce_plans` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByFinanceBudgetReferenceID` | Menyaring daftar `workforce_plans` berdasarkan `finance_budget_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByCreatedByPrincipalID` | Menyaring daftar `workforce_plans` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `FindWorkforcePlanByApprovedByPrincipalID` | Menyaring daftar `workforce_plans` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlansResponse` | `human_resource.workforce_plans.read` · legal entity didukung |
| `CreateWorkforcePlanLine` | Membuat entri `workforce_plan_lines`. | Mutation | `CreateWorkforcePlanLineRequest` → `WorkforcePlanLineResponse` | `human_resource.workforce_plan_lines.create` |
| `UpdateWorkforcePlanLine` | Mengubah entri `workforce_plan_lines` sesuai field mask. | Mutation | `UpdateWorkforcePlanLineRequest` → `WorkforcePlanLineResponse` | `human_resource.workforce_plan_lines.update` |
| `DeleteWorkforcePlanLine` | Menghapus entri `workforce_plan_lines` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.workforce_plan_lines.delete` |
| `GetWorkforcePlanLineByID` | Mengambil entri `workforce_plan_lines` berdasarkan ID. | Query | `IDRequest` → `WorkforcePlanLineResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `GetAllWorkforcePlanLines` | Mendaftar entri `workforce_plan_lines` dengan paginasi. | Query | `PageRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByWorkforcePlanID` | Menyaring daftar `workforce_plan_lines` berdasarkan `workforce_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByLineNumber` | Menyaring daftar `workforce_plan_lines` berdasarkan `line_number`, dengan paginasi. | Query | `Int32FinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByOrganizationUnitReferenceID` | Menyaring daftar `workforce_plan_lines` berdasarkan `organization_unit_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByJobID` | Menyaring daftar `workforce_plan_lines` berdasarkan `job_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByWorkLocationReferenceID` | Menyaring daftar `workforce_plan_lines` berdasarkan `work_location_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByCurrencyCode` | Menyaring daftar `workforce_plan_lines` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |
| `FindWorkforcePlanLineByActionCode` | Menyaring daftar `workforce_plan_lines` berdasarkan `action_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkforcePlanLinesResponse` | `human_resource.workforce_plan_lines.read` · legal entity didukung |

## Butuh GraphQL

36 kandidat Query dan 20 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

Registry menyediakan `WithWorkforceBudgetVerifier` untuk bukti anggaran pada `ApproveWorkforcePlan`. Startup standar belum memasang adapter. Snapshot verifikasi menyertakan lokasi kerja bila tersedia, kategori pekerja dan kode tindakan per baris. [Kebutuhan Finance](../../notes/finance_accounting/2-human_resource-workforce-budget.md).

## Wiring Gateway

`workforcePlanLine` memakai `GetWorkforcePlanLineByID`. `workforcePlanLines` memakai `GetAllWorkforcePlanLines` tanpa filter atau `FindWorkforcePlanLineByWorkforcePlanID` saat ID induk dikirim. Keduanya memerlukan izin baca baris rencana. Ringkasan GraphQL menyajikan kebutuhan tenaga kerja dan jadwal, dengan presisi desimal FTE terjaga. [Matriks tes](../../e2e/gateway/17-hrm-query-baris-rencana-tenaga-kerja.md).

## Penjelasan GetJobsByIds

GraphQL `jobsByIds(ids: [ID!]!)` menerima hingga 100 UUID atau GID jabatan. Hasil unik mengikuti urutan kemunculan pertama. ID yang tidak ditemukan tidak menghasilkan node. Query memerlukan employee dengan grant global `human_resource.jobs.read`.

### Mutation katalog jabatan

`createJobFamily` memerlukan grant global `human_resource.job_families.create`. `createJob` memerlukan grant global `human_resource.jobs.create` dan menerima UUID atau GID bertipe job-family. UI dapat memakai kode dan nama untuk label pilihan. [Bukti HTTPS](../../e2e/gateway/177-hrm-mutation-katalog-jabatan-untuk-requisition.md).

Uji HTTPS tambahan membuat jabatan dan referensi Enterprise Foundation melalui GraphQL. Posisi dengan unit atau lokasi lintas badan hukum ditolak. [Hasil uji](../../e2e/gateway/179-hrm-mutation-posisi-referensi-enterprise-foundation.md).

### Koreksi lokasi posisi terencana

`updatePosition` mempertahankan lokasi jika field `work_location_reference_id` absen. Kirim `null` untuk mengosongkan lokasi atau ID lokasi Enterprise Foundation yang sesuai badan hukum untuk menetapkan lokasi. UI memakai nama dan kode lokasi sebagai label. Koreksi memerlukan grant global `human_resource.positions.update` dan MFA. Gateway memeriksa respons terhadap perubahan lokasi yang diminta.
