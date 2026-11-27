# List operasi

`human_resource.career.v1.CareerService` memiliki 56 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/career/v1/career.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CompleteDevelopmentPlan` | Menyelesaikan entri `development_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `DevelopmentPlanResponse` | `human_resource.development_plans.complete` |
| `ApproveDevelopmentPlan` | Menyetujui entri `development_plans` melalui transisi workflow HRM. | Mutation | `IDRequest` → `DevelopmentPlanResponse` | `human_resource.development_plans.approve` |
| `CreateDevelopmentPlan` | Membuat entri `development_plans`. | Mutation | `CreateDevelopmentPlanRequest` → `DevelopmentPlanResponse` | `human_resource.development_plans.create` |
| `UpdateDevelopmentPlan` | Mengubah entri `development_plans` sesuai field mask. | Mutation | `UpdateDevelopmentPlanRequest` → `DevelopmentPlanResponse` | `human_resource.development_plans.update` |
| `DeleteDevelopmentPlan` | Menghapus entri `development_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.development_plans.delete` |
| `GetDevelopmentPlanByID` | Mengambil entri `development_plans` berdasarkan ID. | Query | `IDRequest` → `DevelopmentPlanResponse` | `human_resource.development_plans.read` |
| `GetAllDevelopmentPlans` | Mendaftar entri `development_plans` dengan paginasi. | Query | `PageRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByWorkerID` | Menyaring daftar `development_plans` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByPlanCode` | Menyaring daftar `development_plans` berdasarkan `plan_code`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByTargetJobID` | Menyaring daftar `development_plans` berdasarkan `target_job_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByStatus` | Menyaring daftar `development_plans` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByMentorWorkerID` | Menyaring daftar `development_plans` berdasarkan `mentor_worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `FindDevelopmentPlanByApprovedByPrincipalID` | Menyaring daftar `development_plans` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentPlansResponse` | `human_resource.development_plans.read` |
| `CompleteDevelopmentAction` | Menyelesaikan entri `development_actions` melalui transisi workflow HRM. | Mutation | `IDRequest` → `DevelopmentActionResponse` | `human_resource.development_actions.complete` |
| `CreateDevelopmentAction` | Membuat entri `development_actions`. | Mutation | `CreateDevelopmentActionRequest` → `DevelopmentActionResponse` | `human_resource.development_actions.create` |
| `UpdateDevelopmentAction` | Mengubah entri `development_actions` sesuai field mask. | Mutation | `UpdateDevelopmentActionRequest` → `DevelopmentActionResponse` | `human_resource.development_actions.update` |
| `DeleteDevelopmentAction` | Menghapus entri `development_actions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.development_actions.delete` |
| `GetDevelopmentActionByID` | Mengambil entri `development_actions` berdasarkan ID. | Query | `IDRequest` → `DevelopmentActionResponse` | `human_resource.development_actions.read` |
| `GetAllDevelopmentActions` | Mendaftar entri `development_actions` dengan paginasi. | Query | `PageRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByDevelopmentPlanID` | Menyaring daftar `development_actions` berdasarkan `development_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByActionNumber` | Menyaring daftar `development_actions` berdasarkan `action_number`, dengan paginasi. | Query | `Int32FinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByCompetencyID` | Menyaring daftar `development_actions` berdasarkan `competency_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByCourseID` | Menyaring daftar `development_actions` berdasarkan `course_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByActionType` | Menyaring daftar `development_actions` berdasarkan `action_type`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `FindDevelopmentActionByEvidenceDocumentVersionReferenceID` | Menyaring daftar `development_actions` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `DevelopmentActionsResponse` | `human_resource.development_actions.read` |
| `CreateCareerPath` | Membuat entri `career_paths`. | Mutation | `CreateCareerPathRequest` → `CareerPathResponse` | `human_resource.career_paths.create` |
| `UpdateCareerPath` | Mengubah entri `career_paths` sesuai field mask. | Mutation | `UpdateCareerPathRequest` → `CareerPathResponse` | `human_resource.career_paths.update` |
| `ApproveCareerPath` | Menyetujui entri `career_paths` melalui transisi workflow HRM. | Mutation | `IDRequest` → `CareerPathResponse` | `human_resource.career_paths.approve` |
| `DeleteCareerPath` | Menghapus entri `career_paths` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.career_paths.delete` |
| `GetCareerPathByID` | Mengambil entri `career_paths` berdasarkan ID. | Query | `IDRequest` → `CareerPathResponse` | `human_resource.career_paths.read` |
| `GetAllCareerPaths` | Mendaftar entri `career_paths` dengan paginasi. | Query | `PageRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `FindCareerPathByPathCode` | Menyaring daftar `career_paths` berdasarkan `path_code`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `FindCareerPathByJobFamilyID` | Menyaring daftar `career_paths` berdasarkan `job_family_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `FindCareerPathByStatus` | Menyaring daftar `career_paths` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `FindCareerPathByCreatedByPrincipalID` | Menyaring daftar `career_paths` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `FindCareerPathByApprovedByPrincipalID` | Menyaring daftar `career_paths` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathsResponse` | `human_resource.career_paths.read` |
| `CreateCareerPathTransition` | Membuat entri `career_path_transitions`. | Mutation | `CreateCareerPathTransitionRequest` → `CareerPathTransitionResponse` | `human_resource.career_path_transitions.create` |
| `UpdateCareerPathTransition` | Mengubah entri `career_path_transitions` sesuai field mask. | Mutation | `UpdateCareerPathTransitionRequest` → `CareerPathTransitionResponse` | `human_resource.career_path_transitions.update` |
| `DeleteCareerPathTransition` | Menghapus entri `career_path_transitions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.career_path_transitions.delete` |
| `GetCareerPathTransitionByID` | Mengambil entri `career_path_transitions` berdasarkan ID. | Query | `IDRequest` → `CareerPathTransitionResponse` | `human_resource.career_path_transitions.read` |
| `GetAllCareerPathTransitions` | Mendaftar entri `career_path_transitions` dengan paginasi. | Query | `PageRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `FindCareerPathTransitionByCareerPathID` | Menyaring daftar `career_path_transitions` berdasarkan `career_path_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `FindCareerPathTransitionByFromJobID` | Menyaring daftar `career_path_transitions` berdasarkan `from_job_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `FindCareerPathTransitionByToJobID` | Menyaring daftar `career_path_transitions` berdasarkan `to_job_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `FindCareerPathTransitionByMovementType` | Menyaring daftar `career_path_transitions` berdasarkan `movement_type`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `FindCareerPathTransitionByRequirementsDocumentVersionReferenceID` | Menyaring daftar `career_path_transitions` berdasarkan `requirements_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CareerPathTransitionsResponse` | `human_resource.career_path_transitions.read` |
| `CreateWorkerCareerAspiration` | Membuat entri `worker_career_aspirations`. | Mutation | `CreateWorkerCareerAspirationRequest` → `WorkerCareerAspirationResponse` | `human_resource.worker_career_aspirations.create` |
| `UpdateWorkerCareerAspiration` | Mengubah entri `worker_career_aspirations` sesuai field mask. | Mutation | `UpdateWorkerCareerAspirationRequest` → `WorkerCareerAspirationResponse` | `human_resource.worker_career_aspirations.update` |
| `DeleteWorkerCareerAspiration` | Menghapus entri `worker_career_aspirations` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.worker_career_aspirations.delete` |
| `GetWorkerCareerAspirationByID` | Mengambil entri `worker_career_aspirations` berdasarkan ID. | Query | `IDRequest` → `WorkerCareerAspirationResponse` | `human_resource.worker_career_aspirations.read` |
| `GetAllWorkerCareerAspirations` | Mendaftar entri `worker_career_aspirations` dengan paginasi. | Query | `PageRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |
| `FindWorkerCareerAspirationByWorkerID` | Menyaring daftar `worker_career_aspirations` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |
| `FindWorkerCareerAspirationByTargetJobID` | Menyaring daftar `worker_career_aspirations` berdasarkan `target_job_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |
| `FindWorkerCareerAspirationByCareerPathTransitionID` | Menyaring daftar `worker_career_aspirations` berdasarkan `career_path_transition_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |
| `FindWorkerCareerAspirationByRecordedByPrincipalID` | Menyaring daftar `worker_career_aspirations` berdasarkan `recorded_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |
| `FindWorkerCareerAspirationByPreferredLocationReferenceID` | Menyaring daftar `worker_career_aspirations` berdasarkan `preferred_location_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerCareerAspirationsResponse` | `human_resource.worker_career_aspirations.read` |

## Butuh GraphQL

37 kandidat Query dan 19 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

`CompleteDevelopmentPlan` memeriksa waktu penyelesaian seluruh tindakan. Tindakan yang waktunya mendahului persetujuan atau tanggal mulai rencana, atau masih di masa depan menurut jam server, menghalangi penyelesaian rencana.

`ApproveDevelopmentPlan` menolak draf yang sudah memiliki tindakan bertanda selesai. Persetujuan membuka workflow mentor, lalu mentor mencatat penyelesaian tindakan melalui command tersendiri.
