# List operasi

`human_resource.performance.v1.PerformanceService` memiliki 60 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/performance/v1/performance.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreatePerformanceCycle` | Membuat entri `performance_cycles`. | Mutation | `CreatePerformanceCycleRequest` → `PerformanceCycleResponse` | `human_resource.performance_cycles.create` |
| `LaunchPerformanceCycle` | Meluncurkan entri `performance_cycles` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceCycleResponse` | `human_resource.performance_cycles.launch` |
| `ClosePerformanceCycle` | Menutup entri `performance_cycles` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceCycleResponse` | `human_resource.performance_cycles.close` |
| `UpdatePerformanceCycle` | Mengubah entri `performance_cycles` sesuai field mask. | Mutation | `UpdatePerformanceCycleRequest` → `PerformanceCycleResponse` | `human_resource.performance_cycles.update` |
| `DeletePerformanceCycle` | Menghapus entri `performance_cycles` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.performance_cycles.delete` |
| `GetPerformanceCycleByID` | Mengambil entri `performance_cycles` berdasarkan ID. | Query | `IDRequest` → `PerformanceCycleResponse` | `human_resource.performance_cycles.read` |
| `GetAllPerformanceCycles` | Mendaftar entri `performance_cycles` dengan paginasi. | Query | `PageRequest` → `PerformanceCyclesResponse` | `human_resource.performance_cycles.read` |
| `FindPerformanceCycleByCycleCode` | Menyaring daftar `performance_cycles` berdasarkan `cycle_code`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceCyclesResponse` | `human_resource.performance_cycles.read` |
| `FindPerformanceCycleByCycleType` | Menyaring daftar `performance_cycles` berdasarkan `cycle_type`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceCyclesResponse` | `human_resource.performance_cycles.read` |
| `FindPerformanceCycleByStatus` | Menyaring daftar `performance_cycles` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceCyclesResponse` | `human_resource.performance_cycles.read` |
| `FindPerformanceCycleByLaunchedByPrincipalID` | Menyaring daftar `performance_cycles` berdasarkan `launched_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceCyclesResponse` | `human_resource.performance_cycles.read` |
| `CreatePerformanceGoal` | Membuat entri `performance_goals`. | Mutation | `CreatePerformanceGoalRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.create` |
| `UpdatePerformanceGoal` | Mengubah entri `performance_goals` sesuai field mask. | Mutation | `UpdatePerformanceGoalRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.update` |
| `DeletePerformanceGoal` | Menghapus entri `performance_goals` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.performance_goals.delete` |
| `GetPerformanceGoalByID` | Mengambil entri `performance_goals` berdasarkan ID. | Query | `IDRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.read` |
| `GetAllPerformanceGoals` | Mendaftar entri `performance_goals` dengan paginasi. | Query | `PageRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByWorkerID` | Menyaring daftar `performance_goals` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByPerformanceCycleID` | Menyaring daftar `performance_goals` berdasarkan `performance_cycle_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByParentGoalID` | Menyaring daftar `performance_goals` berdasarkan `parent_goal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByGoalCode` | Menyaring daftar `performance_goals` berdasarkan `goal_code`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByGoalType` | Menyaring daftar `performance_goals` berdasarkan `goal_type`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByUnitCode` | Menyaring daftar `performance_goals` berdasarkan `unit_code`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByGoalStatus` | Menyaring daftar `performance_goals` berdasarkan `goal_status`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `FindPerformanceGoalByApprovedByPrincipalID` | Menyaring daftar `performance_goals` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalsResponse` | `human_resource.performance_goals.read` |
| `SubmitPerformanceGoal` | Mengajukan entri `performance_goals` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.submit` |
| `ApprovePerformanceGoal` | Menyetujui entri `performance_goals` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.approve` |
| `RejectPerformanceGoal` | Menolak entri `performance_goals` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceGoalResponse` | `human_resource.performance_goals.reject` |
| `CreatePerformanceReview` | Membuat entri `performance_reviews`. | Mutation | `CreatePerformanceReviewRequest` → `PerformanceReviewResponse` | `human_resource.performance_reviews.create` |
| `UpdatePerformanceReview` | Mengubah entri `performance_reviews` sesuai field mask. | Mutation | `UpdatePerformanceReviewRequest` → `PerformanceReviewResponse` | `human_resource.performance_reviews.update` |
| `DeletePerformanceReview` | Menghapus entri `performance_reviews` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.performance_reviews.delete` |
| `GetPerformanceReviewByID` | Mengambil entri `performance_reviews` berdasarkan ID. | Query | `IDRequest` → `PerformanceReviewResponse` | `human_resource.performance_reviews.read` |
| `GetAllPerformanceReviews` | Mendaftar entri `performance_reviews` dengan paginasi. | Query | `PageRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `FindPerformanceReviewByWorkerID` | Menyaring daftar `performance_reviews` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `FindPerformanceReviewByPerformanceCycleID` | Menyaring daftar `performance_reviews` berdasarkan `performance_cycle_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `FindPerformanceReviewByReviewerPrincipalID` | Menyaring daftar `performance_reviews` berdasarkan `reviewer_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `FindPerformanceReviewByOverallRatingCode` | Menyaring daftar `performance_reviews` berdasarkan `overall_rating_code`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `FindPerformanceReviewByDevelopmentPlanID` | Menyaring daftar `performance_reviews` berdasarkan `development_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewsResponse` | `human_resource.performance_reviews.read` |
| `SubmitPerformanceReview` | Mengajukan entri `performance_reviews` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceReviewResponse` | `human_resource.performance_reviews.submit` |
| `FinalizePerformanceReview` | Memfinalisasi entri `performance_reviews` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PerformanceReviewResponse` | `human_resource.performance_reviews.finalize` |
| `CreatePerformanceGoalUpdate` | Membuat entri `performance_goal_updates`. | Mutation | `CreatePerformanceGoalUpdateRequest` → `PerformanceGoalUpdateResponse` | `human_resource.performance_goal_updates.create` |
| `UpdatePerformanceGoalUpdate` | Mengubah entri `performance_goal_updates` sesuai field mask. | Mutation | `UpdatePerformanceGoalUpdateRequest` → `PerformanceGoalUpdateResponse` | `human_resource.performance_goal_updates.update` |
| `DeletePerformanceGoalUpdate` | Menghapus entri `performance_goal_updates` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.performance_goal_updates.delete` |
| `GetPerformanceGoalUpdateByID` | Mengambil entri `performance_goal_updates` berdasarkan ID. | Query | `IDRequest` → `PerformanceGoalUpdateResponse` | `human_resource.performance_goal_updates.read` |
| `GetAllPerformanceGoalUpdates` | Mendaftar entri `performance_goal_updates` dengan paginasi. | Query | `PageRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `FindPerformanceGoalUpdateByGoalID` | Menyaring daftar `performance_goal_updates` berdasarkan `goal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID` | Menyaring daftar `performance_goal_updates` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `FindPerformanceGoalUpdateByRecordedByPrincipalID` | Menyaring daftar `performance_goal_updates` berdasarkan `recorded_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `FindPerformanceGoalUpdateBySupersedesUpdateID` | Menyaring daftar `performance_goal_updates` berdasarkan `supersedes_update_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `FindPerformanceGoalUpdateByIdempotencyKey` | Menyaring daftar `performance_goal_updates` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceGoalUpdatesResponse` | `human_resource.performance_goal_updates.read` |
| `CreatePerformanceReviewGoalScore` | Membuat entri `performance_review_goal_scores`. | Mutation | `CreatePerformanceReviewGoalScoreRequest` → `PerformanceReviewGoalScoreResponse` | `human_resource.performance_review_goal_scores.create` |
| `UpdatePerformanceReviewGoalScore` | Mengubah entri `performance_review_goal_scores` sesuai field mask. | Mutation | `UpdatePerformanceReviewGoalScoreRequest` → `PerformanceReviewGoalScoreResponse` | `human_resource.performance_review_goal_scores.update` |
| `DeletePerformanceReviewGoalScore` | Menghapus entri `performance_review_goal_scores` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.performance_review_goal_scores.delete` |
| `GetPerformanceReviewGoalScoreByID` | Mengambil entri `performance_review_goal_scores` berdasarkan ID. | Query | `IDRequest` → `PerformanceReviewGoalScoreResponse` | `human_resource.performance_review_goal_scores.read` |
| `GetAllPerformanceReviewGoalScores` | Mendaftar entri `performance_review_goal_scores` dengan paginasi. | Query | `PageRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByPerformanceReviewID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `performance_review_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByGoalID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `goal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByWorkerID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByPerformanceCycleID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `performance_cycle_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `rubric_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |
| `FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID` | Menyaring daftar `performance_review_goal_scores` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PerformanceReviewGoalScoresResponse` | `human_resource.performance_review_goal_scores.read` |

## Butuh GraphQL

38 kandidat Query dan 22 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
