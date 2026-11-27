# List operasi

`human_resource.benefits.v1.BenefitsService` memiliki 22 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/benefits/v1/benefits.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `EndBenefitEnrollment` | Mengakhiri entri `benefit_enrollments` melalui transisi workflow HRM. | Mutation | `EndBenefitEnrollmentRequest` → `BenefitEnrollmentResponse` | `human_resource.benefit_enrollments.end` |
| `ApproveBenefitEnrollment` | Menyetujui entri `benefit_enrollments` melalui transisi workflow HRM. | Mutation | `IDRequest` → `BenefitEnrollmentResponse` | `human_resource.benefit_enrollments.approve` |
| `CreateBenefitPlan` | Membuat entri `benefit_plans`. | Mutation | `CreateBenefitPlanRequest` → `BenefitPlanResponse` | `human_resource.benefit_plans.create` |
| `UpdateBenefitPlan` | Mengubah entri `benefit_plans` sesuai field mask. | Mutation | `UpdateBenefitPlanRequest` → `BenefitPlanResponse` | `human_resource.benefit_plans.update` |
| `DeleteBenefitPlan` | Menghapus entri `benefit_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.benefit_plans.delete` |
| `GetBenefitPlanByID` | Mengambil entri `benefit_plans` berdasarkan ID. | Query | `IDRequest` → `BenefitPlanResponse` | `human_resource.benefit_plans.read` |
| `GetAllBenefitPlans` | Mendaftar entri `benefit_plans` dengan paginasi. | Query | `PageRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByPlanCode` | Menyaring daftar `benefit_plans` berdasarkan `plan_code`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByBenefitType` | Menyaring daftar `benefit_plans` berdasarkan `benefit_type`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByProviderPartyReferenceID` | Menyaring daftar `benefit_plans` berdasarkan `provider_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByJurisdictionCode` | Menyaring daftar `benefit_plans` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByEmployeeContributionRuleReferenceID` | Menyaring daftar `benefit_plans` berdasarkan `employee_contribution_rule_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `FindBenefitPlanByEmployerContributionRuleReferenceID` | Menyaring daftar `benefit_plans` berdasarkan `employer_contribution_rule_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitPlansResponse` | `human_resource.benefit_plans.read` |
| `CreateBenefitEnrollment` | Membuat entri `benefit_enrollments`. | Mutation | `CreateBenefitEnrollmentRequest` → `BenefitEnrollmentResponse` | `human_resource.benefit_enrollments.create` |
| `UpdateBenefitEnrollment` | Mengubah entri `benefit_enrollments` sesuai field mask. | Mutation | `UpdateBenefitEnrollmentRequest` → `BenefitEnrollmentResponse` | `human_resource.benefit_enrollments.update` |
| `DeleteBenefitEnrollment` | Menghapus entri `benefit_enrollments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.benefit_enrollments.delete` |
| `GetBenefitEnrollmentByID` | Mengambil entri `benefit_enrollments` berdasarkan ID. | Query | `IDRequest` → `BenefitEnrollmentResponse` | `human_resource.benefit_enrollments.read` |
| `GetAllBenefitEnrollments` | Mendaftar entri `benefit_enrollments` dengan paginasi. | Query | `PageRequest` → `BenefitEnrollmentsResponse` | `human_resource.benefit_enrollments.read` |
| `FindBenefitEnrollmentByWorkerID` | Menyaring daftar `benefit_enrollments` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitEnrollmentsResponse` | `human_resource.benefit_enrollments.read` |
| `FindBenefitEnrollmentByBenefitPlanID` | Menyaring daftar `benefit_enrollments` berdasarkan `benefit_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitEnrollmentsResponse` | `human_resource.benefit_enrollments.read` |
| `FindBenefitEnrollmentByEnrollmentStatus` | Menyaring daftar `benefit_enrollments` berdasarkan `enrollment_status`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitEnrollmentsResponse` | `human_resource.benefit_enrollments.read` |
| `FindBenefitEnrollmentByCurrencyCode` | Menyaring daftar `benefit_enrollments` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `BenefitEnrollmentsResponse` | `human_resource.benefit_enrollments.read` |

## Butuh GraphQL

14 kandidat Query dan 8 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
