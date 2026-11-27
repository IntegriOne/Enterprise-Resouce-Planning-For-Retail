# List operasi

`human_resource.hrservice.v1.HRService` memiliki 27 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/hrservice/v1/hrservice.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CancelServiceRequest` | Pengusul membatalkan tiket terbuka yang belum ditugaskan. | Mutation | `IDRequest` → `ServiceRequestResponse` | `human_resource.service_requests.cancel` |
| `ResolveServiceRequest` | Menyelesaikan entri `service_requests` melalui transisi workflow HRM. | Mutation | `ResolveServiceRequestRequest` → `ServiceRequestResponse` | `human_resource.service_requests.resolve` |
| `CloseServiceRequest` | Menutup entri `service_requests` melalui transisi workflow HRM. | Mutation | `IDRequest` → `ServiceRequestResponse` | `human_resource.service_requests.close` |
| `AssignServiceRequest` | Menugaskan entri `service_requests` melalui transisi workflow HRM. | Mutation | `IDRequest` → `ServiceRequestResponse` | `human_resource.service_requests.assign` |
| `CreateServiceRequest` | Membuat entri `service_requests`. | Mutation | `CreateServiceRequestRequest` → `ServiceRequestResponse` | `human_resource.service_requests.create` |
| `UpdateServiceRequest` | Mengubah entri `service_requests` sesuai field mask. | Mutation | `UpdateServiceRequestRequest` → `ServiceRequestResponse` | `human_resource.service_requests.update` |
| `DeleteServiceRequest` | Menghapus entri `service_requests` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.service_requests.delete` |
| `GetServiceRequestByID` | Mengambil entri `service_requests` berdasarkan ID. | Query | `IDRequest` → `ServiceRequestResponse` | `human_resource.service_requests.read` |
| `GetAllServiceRequests` | Mendaftar entri `service_requests` dengan paginasi. | Query | `PageRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByRequestNumber` | Menyaring daftar `service_requests` berdasarkan `request_number`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByWorkerID` | Menyaring daftar `service_requests` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByRequestType` | Menyaring daftar `service_requests` berdasarkan `request_type`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByRequestedByPrincipalID` | Menyaring daftar `service_requests` berdasarkan `requested_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByAssignedPrincipalID` | Menyaring daftar `service_requests` berdasarkan `assigned_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByStatus` | Menyaring daftar `service_requests` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByRestrictedDetailsDocumentVersionReferenceID` | Menyaring daftar `service_requests` berdasarkan `restricted_details_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByResolutionDocumentVersionReferenceID` | Menyaring daftar `service_requests` berdasarkan `resolution_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `FindServiceRequestByIdempotencyKey` | Menyaring daftar `service_requests` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestsResponse` | `human_resource.service_requests.read` |
| `CreateServiceRequestEvent` | Membuat entri `service_request_events`. | Mutation | `CreateServiceRequestEventRequest` → `ServiceRequestEventResponse` | `human_resource.service_request_events.create` |
| `UpdateServiceRequestEvent` | Mengubah entri `service_request_events` sesuai field mask. | Mutation | `UpdateServiceRequestEventRequest` → `ServiceRequestEventResponse` | `human_resource.service_request_events.update` |
| `DeleteServiceRequestEvent` | Menghapus entri `service_request_events` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.service_request_events.delete` |
| `GetServiceRequestEventByID` | Mengambil entri `service_request_events` berdasarkan ID. | Query | `IDRequest` → `ServiceRequestEventResponse` | `human_resource.service_request_events.read` |
| `GetAllServiceRequestEvents` | Mendaftar entri `service_request_events` dengan paginasi. | Query | `PageRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |
| `FindServiceRequestEventByServiceRequestID` | Menyaring daftar `service_request_events` berdasarkan `service_request_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |
| `FindServiceRequestEventByEventType` | Menyaring daftar `service_request_events` berdasarkan `event_type`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |
| `FindServiceRequestEventByActorPrincipalID` | Menyaring daftar `service_request_events` berdasarkan `actor_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |
| `FindServiceRequestEventByEvidenceDocumentVersionReferenceID` | Menyaring daftar `service_request_events` berdasarkan `evidence_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |
| `FindServiceRequestEventByIdempotencyKey` | Menyaring daftar `service_request_events` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `ServiceRequestEventsResponse` | `human_resource.service_request_events.read` |

## Butuh GraphQL

18 kandidat Query dan 10 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

Registry menyediakan `WithHRServiceResolutionVerifier` untuk bukti pada `ResolveServiceRequest`. Startup standar belum memasang adapter.

`CloseServiceRequest` menolak penutupan jika waktu resolusi yang tersimpan masih berada di masa depan menurut jam server.

`ResolveServiceRequest` memeriksa waktu pembukaan tiket sebelum menghubungi DRM. Jika waktu pembukaan belum tercapai, HRM menolak resolusi tanpa meminta verifikasi dokumen.

[Hasil tes integrasi CancelServiceRequest](../../e2e/human_resource/48-hrservice-mutation-cancel-service-request.md)

Policy HRM menerima grant pembatalan dengan scope `self` yang cocok dengan principal sesi, dengan MFA terbaru sesuai aturan operasi sensitif. Grant global juga dapat melewati pemeriksaan izin, dan command selalu mencocokkan principal dengan pengusul tiket. Provisioning grant tetap mengikuti keputusan pemilik katalog IAM.

`CreateServiceRequest` mengisi principal pengusul dan waktu pembukaan dari konteks server. `worker_id` menunjuk pekerja subjek tiket dan HRM memeriksa bahwa record pekerja tersedia. Resolver mandiri yang kelak dibuat perlu menetapkan pekerja dari identitas sesi yang telah diverifikasi. Operasi staf yang menerima `worker_id` membutuhkan izin pembuatan yang sesuai.
