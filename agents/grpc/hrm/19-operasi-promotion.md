# List operasi

`human_resource.promotion.v1.PromotionService` memiliki 33 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/promotion/v1/promotion.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreatePromotionRecommendation` | Membuat entri `promotion_recommendations`. | Mutation | `CreatePromotionRecommendationRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_recommendations.create` |
| `UpdatePromotionRecommendation` | Mengubah entri `promotion_recommendations` sesuai field mask. | Mutation | `UpdatePromotionRecommendationRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_recommendations.update` |
| `DeletePromotionRecommendation` | Menghapus entri `promotion_recommendations` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.promotion_recommendations.delete` |
| `GetPromotionRecommendationByID` | Mengambil entri `promotion_recommendations` berdasarkan ID. | Query | `IDRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_recommendations.read` |
| `GetAllPromotionRecommendations` | Mendaftar entri `promotion_recommendations` dengan paginasi. | Query | `PageRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByRecommendationNumber` | Menyaring daftar `promotion_recommendations` berdasarkan `recommendation_number`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByWorkerID` | Menyaring daftar `promotion_recommendations` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationBySourceAssignmentID` | Menyaring daftar `promotion_recommendations` berdasarkan `source_assignment_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByTargetPositionID` | Menyaring daftar `promotion_recommendations` berdasarkan `target_position_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByAssessmentID` | Menyaring daftar `promotion_recommendations` berdasarkan `assessment_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationBySupersedesRecommendationID` | Menyaring daftar `promotion_recommendations` berdasarkan `supersedes_recommendation_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByRationaleDocumentVersionReferenceID` | Menyaring daftar `promotion_recommendations` berdasarkan `rationale_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByRecommendedByPrincipalID` | Menyaring daftar `promotion_recommendations` berdasarkan `recommended_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByStatus` | Menyaring daftar `promotion_recommendations` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByApprovalDecisionID` | Menyaring daftar `promotion_recommendations` berdasarkan `approval_decision_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByAppliedAssignmentID` | Menyaring daftar `promotion_recommendations` berdasarkan `applied_assignment_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByAppliedLifecycleEventID` | Menyaring daftar `promotion_recommendations` berdasarkan `applied_lifecycle_event_id`, dengan paginasi. | Query | `Int64FinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `FindPromotionRecommendationByIdempotencyKey` | Menyaring daftar `promotion_recommendations` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionRecommendationsResponse` | `human_resource.promotion_recommendations.read` |
| `SubmitPromotionRecommendation` | Mengajukan entri `promotion_recommendations` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_recommendations.submit` |
| `ApprovePromotionRecommendation` | Menyetujui entri `promotion_recommendations` melalui transisi workflow HRM. | Mutation | `DecidePromotionRecommendationRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_decisions.approve` |
| `RejectPromotionRecommendation` | Menolak entri `promotion_recommendations` melalui transisi workflow HRM. | Mutation | `DecidePromotionRecommendationRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_decisions.reject` |
| `ReturnPromotionRecommendation` | Mengembalikan untuk revisi entri `promotion_recommendations` melalui transisi workflow HRM. | Mutation | `DecidePromotionRecommendationRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_decisions.return` |
| `ApplyPromotionRecommendation` | Menerapkan entri `promotion_recommendations` melalui transisi workflow HRM. | Mutation | `IDRequest` → `PromotionRecommendationResponse` | `human_resource.promotion_decisions.apply` |
| `CreatePromotionDecision` | Membuat entri `promotion_decisions`. | Mutation | `CreatePromotionDecisionRequest` → `PromotionDecisionResponse` | `human_resource.promotion_decisions.create` |
| `UpdatePromotionDecision` | Mengubah entri `promotion_decisions` sesuai field mask. | Mutation | `UpdatePromotionDecisionRequest` → `PromotionDecisionResponse` | `human_resource.promotion_decisions.update` |
| `DeletePromotionDecision` | Menghapus entri `promotion_decisions` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.promotion_decisions.delete` |
| `GetPromotionDecisionByID` | Mengambil entri `promotion_decisions` berdasarkan ID. | Query | `IDRequest` → `PromotionDecisionResponse` | `human_resource.promotion_decisions.read` |
| `GetAllPromotionDecisions` | Mendaftar entri `promotion_decisions` dengan paginasi. | Query | `PageRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |
| `FindPromotionDecisionByRecommendationID` | Menyaring daftar `promotion_decisions` berdasarkan `recommendation_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |
| `FindPromotionDecisionByStepNumber` | Menyaring daftar `promotion_decisions` berdasarkan `step_number`, dengan paginasi. | Query | `Int32FinderRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |
| `FindPromotionDecisionByDecidedByPrincipalID` | Menyaring daftar `promotion_decisions` berdasarkan `decided_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |
| `FindPromotionDecisionByRationaleDocumentVersionReferenceID` | Menyaring daftar `promotion_decisions` berdasarkan `rationale_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |
| `FindPromotionDecisionByIdempotencyKey` | Menyaring daftar `promotion_decisions` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `PromotionDecisionsResponse` | `human_resource.promotion_decisions.read` |

## Butuh GraphQL

22 kandidat Query dan 11 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

## Prasyarat masa kerja

`ApplyPromotionRecommendation` memeriksa ulang masa kerja saat penerapan. HRM menolak jika pekerja sudah memiliki tanggal terminasi atau tanggal efektif mendahului tanggal rekrut.

Penempatan sumber boleh sudah dijadwalkan berakhir tepat sehari sebelum tanggal efektif promosi. Jika penempatan sumber berakhir lebih awal, HRM menolak penerapan karena ada jeda.

Principal IAM pekerja yang menjadi subjek promosi tidak dapat menjalankan `ApplyPromotionRecommendation` untuk dirinya sendiri, termasuk saat rekomendasi sudah disetujui.

Sebelum mengubah penempatan, HRM memeriksa posisi tujuan masih aktif, berada pada badan hukum pekerja, serta mencakup tanggal efektif dan waktu penerapan.
