# List operasi

`human_resource.succession.v1.SuccessionService` memiliki 19 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/succession/v1/succession.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `ApproveSuccessionPlan` | Mengesahkan rencana setelah seluruh nominasi lolos validasi ulang serta mandat panel dan konflik kepentingan diperiksa. | Mutation | `IDRequest` → `SuccessionPlanResponse` | `human_resource.succession_plans.approve` |
| `CreateSuccessionPlan` | Membuat entri `succession_plans`. | Mutation | `CreateSuccessionPlanRequest` → `SuccessionPlanResponse` | `human_resource.succession_plans.create` |
| `UpdateSuccessionPlan` | Mengubah entri `succession_plans` sesuai field mask. | Mutation | `UpdateSuccessionPlanRequest` → `SuccessionPlanResponse` | `human_resource.succession_plans.update` |
| `DeleteSuccessionPlan` | Menghapus entri `succession_plans` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.succession_plans.delete` |
| `GetSuccessionPlanByID` | Mengambil entri `succession_plans` berdasarkan ID. | Query | `IDRequest` → `SuccessionPlanResponse` | `human_resource.succession_plans.read` |
| `GetAllSuccessionPlans` | Mendaftar entri `succession_plans` dengan paginasi. | Query | `PageRequest` → `SuccessionPlansResponse` | `human_resource.succession_plans.read` |
| `FindSuccessionPlanByPositionID` | Menyaring daftar `succession_plans` berdasarkan `position_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionPlansResponse` | `human_resource.succession_plans.read` |
| `FindSuccessionPlanByPlanCode` | Menyaring daftar `succession_plans` berdasarkan `plan_code`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionPlansResponse` | `human_resource.succession_plans.read` |
| `FindSuccessionPlanByApprovedByPrincipalID` | Menyaring daftar `succession_plans` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionPlansResponse` | `human_resource.succession_plans.read` |
| `CreateSuccessionCandidate` | Membuat entri `succession_candidates`. | Mutation | `CreateSuccessionCandidateRequest` → `SuccessionCandidateResponse` | `human_resource.succession_candidates.create` |
| `UpdateSuccessionCandidate` | Mengubah entri `succession_candidates` sesuai field mask. | Mutation | `UpdateSuccessionCandidateRequest` → `SuccessionCandidateResponse` | `human_resource.succession_candidates.update` |
| `DeleteSuccessionCandidate` | Menghapus entri `succession_candidates` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.succession_candidates.delete` |
| `GetSuccessionCandidateByID` | Mengambil entri `succession_candidates` berdasarkan ID. | Query | `IDRequest` → `SuccessionCandidateResponse` | `human_resource.succession_candidates.read` |
| `GetAllSuccessionCandidates` | Mendaftar entri `succession_candidates` dengan paginasi. | Query | `PageRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |
| `FindSuccessionCandidateBySuccessionPlanID` | Menyaring daftar `succession_candidates` berdasarkan `succession_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |
| `FindSuccessionCandidateByCandidateWorkerID` | Menyaring daftar `succession_candidates` berdasarkan `candidate_worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |
| `FindSuccessionCandidateByReadinessCode` | Menyaring daftar `succession_candidates` berdasarkan `readiness_code`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |
| `FindSuccessionCandidateByTalentAssessmentID` | Menyaring daftar `succession_candidates` berdasarkan `talent_assessment_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |
| `FindSuccessionCandidateByDevelopmentPlanID` | Menyaring daftar `succession_candidates` berdasarkan `development_plan_id`, dengan paginasi. | Query | `StringFinderRequest` → `SuccessionCandidatesResponse` | `human_resource.succession_candidates.read` |

## Butuh GraphQL

12 kandidat Query dan 7 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

`ApproveSuccessionPlan` membutuhkan MFA, tanggal review yang sudah tercapai, minimal satu kandidat, dan verifier mandat panel. Runtime tanpa verifier mengembalikan `Unavailable`. Keputusan membekukan nominasi tanpa mengubah jabatan pekerja.
Verifier panel menerima ID dan versi rencana, posisi, principal penyetuju, tanggal review, serta daftar nominasi dengan ID pekerja, kode dan tanggal asesmen kesiapan, dan referensi asesmen talenta serta rencana pengembangan bila tersedia.

Respons rencana yang sudah disetujui menyertakan `approvalMandateReferenceId` untuk audit. Referensi ini berasal dari verifier. Gateway tidak mengirimkannya sebagai input approval.

Jika principal penyetuju terhubung ke salah satu pekerja yang dinominasikan, HRM mengembalikan `PermissionDenied` meskipun mandat panel dinyatakan sah.

Draf baru mengembalikan `createdByPrincipalId` dan `createdAt` yang ditetapkan HRM. Persetujuan menolak principal pembuat dan draf historis yang tidak memiliki bukti pembuat.
