# List operasi

`human_resource.onboarding.v1.OnboardingService` memiliki 23 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/onboarding/v1/onboarding.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateOnboardingPlan` | Membuat entri `onboarding_plans`. | Mutation | `CreateOnboardingPlanRequest` → `OnboardingPlanResponse` | `human_resource.onboarding_plans.create` |
| `UpdateOnboardingPlan` | Mengubah entri `onboarding_plans` sesuai field mask. | Mutation | `UpdateOnboardingPlanRequest` → `OnboardingPlanResponse` | `human_resource.onboarding_plans.update` |
| `StartOnboardingPlan` | Memulai rencana setelah semua tugas wajib memiliki penanggung jawab yang divalidasi terhadap principal IAM pekerja. | Mutation | `IDRequest` → `OnboardingPlanResponse` | `human_resource.onboarding_plans.start` |
| `CompleteOnboardingPlan` | Memfinalisasi rencana sesudah tugas wajib selesai dengan bukti dan waktu penyelesaian yang valid. | Mutation | `IDRequest` → `OnboardingPlanResponse` | `human_resource.onboarding_plans.complete` |
| `DeleteOnboardingPlan` | Menghapus entri `onboarding_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.onboarding_plans.delete` |
| `GetOnboardingPlanByID` | Mengambil entri `onboarding_plans` berdasarkan ID. | Query | `IDRequest` → `OnboardingPlanResponse` | `human_resource.onboarding_plans.read` |
| `GetAllOnboardingPlans` | Mendaftar entri `onboarding_plans` dengan paginasi. | Query | `PageRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `FindOnboardingPlanByWorkerID` | Menyaring daftar `onboarding_plans` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `FindOnboardingPlanByPlanCode` | Menyaring daftar `onboarding_plans` berdasarkan `plan_code`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `FindOnboardingPlanByStatus` | Menyaring daftar `onboarding_plans` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `FindOnboardingPlanByTemplateReferenceID` | Menyaring daftar `onboarding_plans` berdasarkan `template_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `FindOnboardingPlanByCreatedByPrincipalID` | Menyaring daftar `onboarding_plans` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingPlansResponse` | `human_resource.onboarding_plans.read` |
| `CreateOnboardingTask` | Membuat entri `onboarding_tasks`. | Mutation | `CreateOnboardingTaskRequest` → `OnboardingTaskResponse` | `human_resource.onboarding_tasks.create` |
| `UpdateOnboardingTask` | Mengubah entri `onboarding_tasks` sesuai field mask. | Mutation | `UpdateOnboardingTaskRequest` → `OnboardingTaskResponse` | `human_resource.onboarding_tasks.update` |
| `CompleteOnboardingTask` | Menyelesaikan entri `onboarding_tasks` melalui transisi workflow HRM. | Mutation | `CompleteOnboardingTaskRequest` → `OnboardingTaskResponse` | `human_resource.onboarding_tasks.complete` |
| `DeleteOnboardingTask` | Menghapus entri `onboarding_tasks` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.onboarding_tasks.delete` |
| `GetOnboardingTaskByID` | Mengambil entri `onboarding_tasks` berdasarkan ID. | Query | `IDRequest` → `OnboardingTaskResponse` | `human_resource.onboarding_tasks.read` |
| `GetAllOnboardingTasks` | Mendaftar entri `onboarding_tasks` dengan paginasi. | Query | `PageRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |
| `FindOnboardingTaskByOnboardingPlanID` | Menyaring daftar `onboarding_tasks` berdasarkan `onboarding_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |
| `FindOnboardingTaskByTaskCode` | Menyaring daftar `onboarding_tasks` berdasarkan `task_code`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |
| `FindOnboardingTaskByTaskOwnerType` | Menyaring daftar `onboarding_tasks` berdasarkan `task_owner_type`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |
| `FindOnboardingTaskByAssignedPrincipalID` | Menyaring daftar `onboarding_tasks` berdasarkan `assigned_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |
| `FindOnboardingTaskByCompletionEvidenceReferenceID` | Menyaring daftar `onboarding_tasks` berdasarkan `completion_evidence_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `OnboardingTasksResponse` | `human_resource.onboarding_tasks.read` |

## Butuh GraphQL

14 kandidat Query dan 9 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

Registry menyediakan `WithOnboardingEvidenceVerifier` untuk bukti pada `CompleteOnboardingTask`. Startup standar belum memasang adapter.
