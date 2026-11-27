# List operasi

`human_resource.competency.v1.CompetencyService` memiliki 36 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/competency/v1/competency.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateWorkerQualification` | Membuat entri `worker_qualifications`. | Mutation | `CreateWorkerQualificationRequest` → `WorkerQualificationResponse` | `human_resource.worker_qualifications.create` |
| `UpdateWorkerQualification` | Mengubah entri `worker_qualifications` sesuai field mask. | Mutation | `UpdateWorkerQualificationRequest` → `WorkerQualificationResponse` | `human_resource.worker_qualifications.update` |
| `DeleteWorkerQualification` | Menghapus entri `worker_qualifications` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.worker_qualifications.delete` |
| `GetWorkerQualificationByID` | Mengambil entri `worker_qualifications` berdasarkan ID. | Query | `IDRequest` → `WorkerQualificationResponse` | `human_resource.worker_qualifications.read` |
| `GetAllWorkerQualifications` | Mendaftar entri `worker_qualifications` dengan paginasi. | Query | `PageRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByWorkerID` | Menyaring daftar `worker_qualifications` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByQualificationType` | Menyaring daftar `worker_qualifications` berdasarkan `qualification_type`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByQualificationCode` | Menyaring daftar `worker_qualifications` berdasarkan `qualification_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByIssuerPartyReferenceID` | Menyaring daftar `worker_qualifications` berdasarkan `issuer_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByCredentialReferenceID` | Menyaring daftar `worker_qualifications` berdasarkan `credential_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByVerificationStatus` | Menyaring daftar `worker_qualifications` berdasarkan `verification_status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `FindWorkerQualificationByVerifiedByPrincipalID` | Menyaring daftar `worker_qualifications` berdasarkan `verified_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerQualificationsResponse` | `human_resource.worker_qualifications.read` |
| `CreateCompetency` | Membuat entri `competencies`. | Mutation | `CreateCompetencyRequest` → `CompetencyResponse` | `human_resource.competencies.create` |
| `UpdateCompetency` | Mengubah entri `competencies` sesuai field mask. | Mutation | `UpdateCompetencyRequest` → `CompetencyResponse` | `human_resource.competencies.update` |
| `DeleteCompetency` | Menghapus entri `competencies` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.competencies.delete` |
| `GetCompetencyByID` | Mengambil entri `competencies` berdasarkan ID. | Query | `IDRequest` → `CompetencyResponse` | `human_resource.competencies.read` |
| `GetAllCompetencies` | Mendaftar entri `competencies` dengan paginasi. | Query | `PageRequest` → `CompetenciesResponse` | `human_resource.competencies.read` |
| `FindCompetencyByCompetencyCode` | Menyaring daftar `competencies` berdasarkan `competency_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompetenciesResponse` | `human_resource.competencies.read` |
| `FindCompetencyByCategoryCode` | Menyaring daftar `competencies` berdasarkan `category_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompetenciesResponse` | `human_resource.competencies.read` |
| `CreateJobCompetency` | Membuat entri `job_competencies`. | Mutation | `CreateJobCompetencyRequest` → `JobCompetencyResponse` | `human_resource.job_competencies.create` |
| `UpdateJobCompetency` | Mengubah entri `job_competencies` sesuai field mask. | Mutation | `UpdateJobCompetencyRequest` → `JobCompetencyResponse` | `human_resource.job_competencies.update` |
| `DeleteJobCompetency` | Menghapus entri `job_competencies` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.job_competencies.delete` |
| `GetJobCompetencyByID` | Mengambil entri `job_competencies` berdasarkan ID. | Query | `IDRequest` → `JobCompetencyResponse` | `human_resource.job_competencies.read` |
| `GetAllJobCompetencies` | Mendaftar entri `job_competencies` dengan paginasi. | Query | `PageRequest` → `JobCompetenciesResponse` | `human_resource.job_competencies.read` |
| `FindJobCompetencyByJobID` | Menyaring daftar `job_competencies` berdasarkan `job_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobCompetenciesResponse` | `human_resource.job_competencies.read` |
| `FindJobCompetencyByCompetencyID` | Menyaring daftar `job_competencies` berdasarkan `competency_id`, dengan paginasi. | Query | `StringFinderRequest` → `JobCompetenciesResponse` | `human_resource.job_competencies.read` |
| `CreateWorkerCompetencyAssessment` | Membuat entri `worker_competency_assessments`. | Mutation | `CreateWorkerCompetencyAssessmentRequest` → `WorkerCompetencyAssessmentResponse` | `human_resource.worker_competency_assessments.create` |
| `UpdateWorkerCompetencyAssessment` | Mengubah entri `worker_competency_assessments` sesuai field mask. | Mutation | `UpdateWorkerCompetencyAssessmentRequest` → `WorkerCompetencyAssessmentResponse` | `human_resource.worker_competency_assessments.update` |
| `DeleteWorkerCompetencyAssessment` | Menghapus entri `worker_competency_assessments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.worker_competency_assessments.delete` |
| `GetWorkerCompetencyAssessmentByID` | Mengambil entri `worker_competency_assessments` berdasarkan ID. | Query | `IDRequest` → `WorkerCompetencyAssessmentResponse` | `human_resource.worker_competency_assessments.read` |
| `GetAllWorkerCompetencyAssessments` | Mendaftar entri `worker_competency_assessments` dengan paginasi. | Query | `PageRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |
| `FindWorkerCompetencyAssessmentByWorkerID` | Menyaring daftar `worker_competency_assessments` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |
| `FindWorkerCompetencyAssessmentByCompetencyID` | Menyaring daftar `worker_competency_assessments` berdasarkan `competency_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |
| `FindWorkerCompetencyAssessmentByAssessedByPrincipalID` | Menyaring daftar `worker_competency_assessments` berdasarkan `assessed_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |
| `FindWorkerCompetencyAssessmentByEvidenceDocumentVersionReferenceID` | Menyaring daftar `worker_competency_assessments` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |
| `FindWorkerCompetencyAssessmentByIdempotencyKey` | Menyaring daftar `worker_competency_assessments` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCompetencyAssessmentsResponse` | `human_resource.worker_competency_assessments.read` |

## Butuh GraphQL

24 kandidat Query dan 12 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
