# List operasi

`human_resource.accessrequest.v1.AccessRequestService` memiliki 15 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/accessrequest/v1/accessrequest.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateAccessRequest` | Membuat entri `access_requests`. | Mutation | `CreateAccessRequestRequest` → `AccessRequestResponse` | `human_resource.access_requests.create` |
| `UpdateAccessRequest` | Mengubah entri `access_requests` sesuai field mask. | Mutation | `UpdateAccessRequestRequest` → `AccessRequestResponse` | `human_resource.access_requests.update` |
| `DeleteAccessRequest` | Menghapus entri `access_requests` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.access_requests.delete` |
| `GetAccessRequestByID` | Mengambil entri `access_requests` berdasarkan ID. | Query | `IDRequest` → `AccessRequestResponse` | `human_resource.access_requests.read` |
| `GetAllAccessRequests` | Mendaftar entri `access_requests` dengan paginasi. | Query | `PageRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByWorkerID` | Menyaring daftar `access_requests` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByLifecycleEventID` | Menyaring daftar `access_requests` berdasarkan `lifecycle_event_id`, dengan paginasi. | Query | `Int64FinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByRequestedScopeType` | Menyaring daftar `access_requests` berdasarkan `requested_scope_type`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByRequestedScopeReferenceID` | Menyaring daftar `access_requests` berdasarkan `requested_scope_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByApprovedCatalogReferenceID` | Menyaring daftar `access_requests` berdasarkan `approved_catalog_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByIAMRequestReferenceID` | Menyaring daftar `access_requests` berdasarkan `iam_request_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByIAMDecisionEventReferenceID` | Menyaring daftar `access_requests` berdasarkan `iam_decision_event_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByStatus` | Menyaring daftar `access_requests` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByRequestedByPrincipalID` | Menyaring daftar `access_requests` berdasarkan `requested_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |
| `FindAccessRequestByIdempotencyKey` | Menyaring daftar `access_requests` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `AccessRequestsResponse` | `human_resource.access_requests.read` |

## Butuh GraphQL

12 kandidat Query dan 3 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).
