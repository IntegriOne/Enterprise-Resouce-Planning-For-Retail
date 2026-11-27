# List operasi

`human_resource.personnel.v1.PersonnelService` memiliki 78 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/personnel/v1/personnel.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateWorker` | Membuat entri `workers`. | Mutation | `CreateWorkerRequest` → `WorkerResponse` | `human_resource.workers.create` |
| `ActivateWorker` | Mengaktifkan entri `workers` melalui transisi workflow HRM. | Mutation | `IDRequest` → `WorkerResponse` | `human_resource.workers.activate` |
| `LinkWorkerPrincipal` | Menautkan akun employee dengan pekerja aktif berdasarkan bukti provisioning yang terverifikasi melalui `createEmployeeUser` IAM. | Mutation | `LinkWorkerPrincipalRequest` → `WorkerResponse` | `iam.users.write` |
| `TerminateWorker` | Menerapkan terminasi berdasarkan rencana offboarding yang disetujui dan berjalan. | Mutation | `TerminateWorkerRequest` → `WorkerResponse` | `human_resource.workers.terminate` |
| `SubmitWorkerAssignmentTransfer` | Mengajukan transfer dari penempatan primer saat ini ke posisi lain dalam badan hukum yang sama. | Mutation | `SubmitWorkerAssignmentTransferRequest` → `SubmitWorkerAssignmentTransferResponse` | `human_resource.worker_assignments.create` |
| `ListWorkerAssignmentTransfers` | Membaca halaman transfer untuk satu pekerja. | Query | `ListWorkerAssignmentTransfersRequest` → `ListWorkerAssignmentTransfersResponse` | `human_resource.worker_assignment_transfers.read` |
| `GetWorkerAssignmentTransfer` | Mengambil status transfer menurut ID. | Query | `IDRequest` → `GetWorkerAssignmentTransferResponse` | `human_resource.worker_assignment_transfers.read` |
| `DecideWorkerAssignmentTransfer` | Menolak permintaan atau menyetujui dan menerapkan transfer penempatan primer secara atomik. | Mutation | `DecideWorkerAssignmentTransferRequest` → `DecideWorkerAssignmentTransferResponse` | `human_resource.worker_assignments.transfer` |
| `ApplyPrimaryAssignment` | Menerapkan penugasan utama pada entri `workers` melalui transisi workflow HRM. | Mutation | `ApplyPrimaryAssignmentRequest` → `WorkerResponse` | `human_resource.workers.apply_primary_assignment` |
| `UpdateWorker` | Mengubah entri `workers` sesuai field mask. | Mutation | `UpdateWorkerRequest` → `WorkerResponse` | `human_resource.workers.update` |
| `DeleteWorker` | Menghapus entri `workers` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.workers.delete` |
| `GetWorkerByID` | Mengambil entri `workers` berdasarkan ID. | Query | `IDRequest` → `WorkerResponse` | `human_resource.workers.read` · legal entity didukung |
| `GetAllWorkers` | Mendaftar entri `workers` dengan paginasi. | Query | `PageRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByWorkerNumber` | Menyaring daftar `workers` berdasarkan `worker_number`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByPersonPartyReferenceID` | Menyaring daftar `workers` berdasarkan `person_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByStatus` | Menyaring daftar `workers` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByTerminationReasonCode` | Menyaring daftar `workers` berdasarkan `termination_reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByPrimaryLegalEntityReferenceID` | Menyaring daftar `workers` berdasarkan `primary_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByPrimaryAssignmentID` | Menyaring daftar `workers` berdasarkan `primary_assignment_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByIAMPrincipalReferenceID` | Menyaring daftar `workers` berdasarkan `iam_principal_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `FindWorkerByCreatedByPrincipalID` | Menyaring daftar `workers` berdasarkan `created_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkersResponse` | `human_resource.workers.read` · legal entity didukung |
| `RecordSignedEmploymentContract` | Mencatat versi kontrak bertanda tangan setelah bukti DRM diverifikasi. | Mutation | `RecordSignedEmploymentContractRequest` → `EmploymentContractResponse` | `human_resource.employment_contracts.record_signature` |
| `CreateEmploymentContract` | Membuat entri `employment_contracts`. | Mutation | `CreateEmploymentContractRequest` → `EmploymentContractResponse` | `human_resource.employment_contracts.create` |
| `UpdateEmploymentContract` | Mengubah entri `employment_contracts` sesuai field mask. | Mutation | `UpdateEmploymentContractRequest` → `EmploymentContractResponse` | `human_resource.employment_contracts.update` |
| `DeleteEmploymentContract` | Menghapus entri `employment_contracts` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.employment_contracts.delete` |
| `GetEmploymentContractByID` | Mengambil entri `employment_contracts` berdasarkan ID. | Query | `IDRequest` → `EmploymentContractResponse` | `human_resource.employment_contracts.read` |
| `GetAllEmploymentContracts` | Mendaftar entri `employment_contracts` dengan paginasi. | Query | `PageRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByWorkerID` | Menyaring daftar `employment_contracts` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByContractNumber` | Menyaring daftar `employment_contracts` berdasarkan `contract_number`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByContractType` | Menyaring daftar `employment_contracts` berdasarkan `contract_type`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByContractStatus` | Menyaring daftar `employment_contracts` berdasarkan `contract_status`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByEmployingLegalEntityReferenceID` | Menyaring daftar `employment_contracts` berdasarkan `employing_legal_entity_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractBySignedDocumentReferenceID` | Menyaring daftar `employment_contracts` berdasarkan `signed_document_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractBySignedDocumentVersionReferenceID` | Menyaring daftar `employment_contracts` berdasarkan `signed_document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByWorkCountryCode` | Menyaring daftar `employment_contracts` berdasarkan `work_country_code`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByGoverningJurisdictionCode` | Menyaring daftar `employment_contracts` berdasarkan `governing_jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `FindEmploymentContractByCollectiveAgreementID` | Menyaring daftar `employment_contracts` berdasarkan `collective_agreement_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmploymentContractsResponse` | `human_resource.employment_contracts.read` |
| `CreateWorkerAssignment` | Membuat entri `worker_assignments`. | Mutation | `CreateWorkerAssignmentRequest` → `WorkerAssignmentResponse` | `human_resource.worker_assignments.create` |
| `ApproveWorkerAssignment` | Menyetujui entri `worker_assignments` melalui transisi workflow HRM. | Mutation | `IDRequest` → `WorkerAssignmentResponse` | `human_resource.worker_assignments.approve` |
| `UpdateWorkerAssignment` | Mengubah entri `worker_assignments` sesuai field mask. | Mutation | `UpdateWorkerAssignmentRequest` → `WorkerAssignmentResponse` | `human_resource.worker_assignments.update` |
| `DeleteWorkerAssignment` | Menghapus entri `worker_assignments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.worker_assignments.delete` |
| `GetWorkerAssignmentByID` | Mengambil entri `worker_assignments` berdasarkan ID. | Query | `IDRequest` → `WorkerAssignmentResponse` | `human_resource.worker_assignments.read` |
| `GetAllWorkerAssignments` | Mendaftar entri `worker_assignments` dengan paginasi. | Query | `PageRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByWorkerID` | Menyaring daftar `worker_assignments` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByPositionID` | Menyaring daftar `worker_assignments` berdasarkan `position_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByManagerWorkerID` | Menyaring daftar `worker_assignments` berdasarkan `manager_worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByAssignmentType` | Menyaring daftar `worker_assignments` berdasarkan `assignment_type`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByRequestedByPrincipalID` | Menyaring daftar `worker_assignments` berdasarkan `requested_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `FindWorkerAssignmentByApprovedByPrincipalID` | Menyaring daftar `worker_assignments` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerAssignmentsResponse` | `human_resource.worker_assignments.read` |
| `CreateWorkerLifecycleEvent` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreateWorkerLifecycleEventRequest` → `WorkerLifecycleEventResponse` | — |
| `UpdateWorkerLifecycleEvent` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdateWorkerLifecycleEventRequest` → `WorkerLifecycleEventResponse` | — |
| `DeleteWorkerLifecycleEvent` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IdentityIDRequest` → `EmptyResponse` | — |
| `GetWorkerLifecycleEventByID` | Mengambil entri `worker_lifecycle_events` berdasarkan ID. | Query | `IdentityIDRequest` → `WorkerLifecycleEventResponse` | `human_resource.worker_lifecycle_events.read` |
| `GetAllWorkerLifecycleEvents` | Mendaftar entri `worker_lifecycle_events` dengan paginasi. | Query | `PageRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByWorkerID` | Menyaring daftar `worker_lifecycle_events` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByEventType` | Menyaring daftar `worker_lifecycle_events` berdasarkan `event_type`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByPreviousStatus` | Menyaring daftar `worker_lifecycle_events` berdasarkan `previous_status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByNewStatus` | Menyaring daftar `worker_lifecycle_events` berdasarkan `new_status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByActorPrincipalID` | Menyaring daftar `worker_lifecycle_events` berdasarkan `actor_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `FindWorkerLifecycleEventByIdempotencyKey` | Menyaring daftar `worker_lifecycle_events` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerLifecycleEventsResponse` | `human_resource.worker_lifecycle_events.read` |
| `CreateWorkerDocument` | Membuat entri `worker_documents`. | Mutation | `CreateWorkerDocumentRequest` → `WorkerDocumentResponse` | `human_resource.worker_documents.create` |
| `UpdateWorkerDocument` | Mengubah entri `worker_documents` sesuai field mask. | Mutation | `UpdateWorkerDocumentRequest` → `WorkerDocumentResponse` | `human_resource.worker_documents.update` |
| `DeleteWorkerDocument` | Menghapus entri `worker_documents` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.worker_documents.delete` |
| `GetWorkerDocumentByID` | Mengambil entri `worker_documents` berdasarkan ID. | Query | `IDRequest` → `WorkerDocumentResponse` | `human_resource.worker_documents.read` |
| `GetAllWorkerDocuments` | Mendaftar entri `worker_documents` dengan paginasi. | Query | `PageRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByWorkerID` | Menyaring daftar `worker_documents` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByDocumentType` | Menyaring daftar `worker_documents` berdasarkan `document_type`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByDocumentReferenceID` | Menyaring daftar `worker_documents` berdasarkan `document_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByDocumentVersionReferenceID` | Menyaring daftar `worker_documents` berdasarkan `document_version_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByIssuingCountryCode` | Menyaring daftar `worker_documents` berdasarkan `issuing_country_code`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByVerificationStatus` | Menyaring daftar `worker_documents` berdasarkan `verification_status`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `FindWorkerDocumentByVerifiedByPrincipalID` | Menyaring daftar `worker_documents` berdasarkan `verified_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `WorkerDocumentsResponse` | `human_resource.worker_documents.read` |
| `CreateEmergencyContact` | Membuat entri `emergency_contacts`. | Mutation | `CreateEmergencyContactRequest` → `EmergencyContactResponse` | `human_resource.emergency_contacts.create` |
| `UpdateEmergencyContact` | Mengubah entri `emergency_contacts` sesuai field mask. | Mutation | `UpdateEmergencyContactRequest` → `EmergencyContactResponse` | `human_resource.emergency_contacts.update` |
| `DeleteEmergencyContact` | Menghapus entri `emergency_contacts` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.emergency_contacts.delete` |
| `GetEmergencyContactByID` | Mengambil entri `emergency_contacts` berdasarkan ID. | Query | `IDRequest` → `EmergencyContactResponse` | `human_resource.emergency_contacts.read` |
| `GetAllEmergencyContacts` | Mendaftar entri `emergency_contacts` dengan paginasi. | Query | `PageRequest` → `EmergencyContactsResponse` | `human_resource.emergency_contacts.read` |
| `FindEmergencyContactByWorkerID` | Menyaring daftar `emergency_contacts` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmergencyContactsResponse` | `human_resource.emergency_contacts.read` |
| `FindEmergencyContactByContactPartyReferenceID` | Menyaring daftar `emergency_contacts` berdasarkan `contact_party_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `EmergencyContactsResponse` | `human_resource.emergency_contacts.read` |
| `FindEmergencyContactByRelationshipCode` | Menyaring daftar `emergency_contacts` berdasarkan `relationship_code`, dengan paginasi. | Query | `StringFinderRequest` → `EmergencyContactsResponse` | `human_resource.emergency_contacts.read` |

## Butuh GraphQL

52 kandidat Query dan 23 kandidat Mutation dan 3 RPC nonaktif. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

## Prasyarat penempatan primer

`ApplyPrimaryAssignment` menolak pengganti yang tanggal mulainya sama dengan atau lebih awal dari tanggal akhir penempatan primer lama. Tanggal pada kedua rentang berlaku secara inklusif.

`SubmitWorkerAssignmentTransfer` menerima tanggal efektif dalam format `YYYY-MM-DD`. Pengulangan dengan principal, idempotency key dan isi yang sama mengembalikan permintaan yang tersimpan. Key dengan payload berbeda ditolak. Pengajuan menyimpan status `submitted` dan belum menerapkan transfer.

`DecideWorkerAssignmentTransfer` menerima keputusan `APPROVE` atau `REJECT`. Principal harus berbeda dari pengaju dan pekerja terkait. Kedua keputusan memerlukan MFA sesuai policy HRM. Persetujuan dapat diterapkan mulai tanggal efektif. Retry keputusan yang sama oleh principal yang sama mengembalikan hasil yang sudah tersimpan.

## Terminasi pekerja

RPC privat `TerminateWorker` menerima ID pekerja dan ID rencana offboarding. Tanggal efektif dan alasan berasal dari rencana yang disetujui dan berstatus `in_progress`. HRM memeriksa identitas, persetujuan independen, dan tanggal efektif. Command menolak terminasi diri sendiri.

Transaksi mengunci rencana lalu pekerja. Penempatan yang disetujui dan kontrak bertanda tangan yang masih berlangsung ditutup pada hari kerja terakhir secara inklusif. Status pekerja berubah menjadi `terminated`, referensi penempatan primer dikosongkan, dan riwayat lifecycle serta outbox dicatat dalam transaksi yang sama. Permintaan ulang setelah berhasil ditolak dengan `FailedPrecondition` tanpa perubahan tambahan. IAM tetap memutuskan pencabutan sesi dan akses. Mutation Gateway membutuhkan izin `human_resource.workers.terminate` dan MFA.

Kontrak bertanda tangan yang dijadwalkan mulai setelah hari kerja terakhir berubah menjadi `cancelled` dengan tanggal dan bukti tanda tangan tetap tersimpan. Penempatan disetujui yang dimulai setelah hari kerja terakhir diberi `cancelled_at` dengan jadwal dan metadata persetujuan tetap tersimpan. Gateway dapat menampilkan `cancelledAt` dari respons penempatan. Kapasitas dan penetapan penempatan primer mengabaikan penempatan yang sudah batal. Transfer tertunda tetap tersimpan dan petugas yang berwenang dapat mencatat penolakan melalui `DecideWorkerAssignmentTransfer`. Percobaan menyetujui transfer setelah terminasi ditolak. [Laporan gRPC](../../e2e/human_resource/41-personnel-mutation-terminate-worker.md)

Perubahan penempatan utama membuat usulan transfer yang masih menunjuk penempatan lama tidak dapat disetujui. Usulan tersebut tetap dapat ditolak secara eksplisit. [Laporan tes](../../e2e/human_resource/24-personnel-mutation-decide-worker-assignment-transfer.md)

Periode penempatan harus dimulai pada atau setelah tanggal mulai bekerja pekerja, jika tanggal tersebut sudah tercatat. Jika penempatan menyertakan manajer, tanggal mulai bekerja manajer juga menjadi batas awal.

`ApplyPrimaryAssignment` memeriksa ulang tanggal mulai bekerja terhadap awal penempatan yang sudah disetujui, termasuk data lama. Transfer dan promosi memeriksa tanggal mulai bekerja terhadap tanggal efektif penempatan pengganti.

Baseline HRM memastikan satu principal IAM paling banyak terhubung ke satu hubungan kerja berjalan, yang mencakup status aktif, cuti panjang, ditangguhkan, dan masa pemberitahuan. Referensi hubungan kerja yang telah berakhir tetap bisa disimpan untuk riwayat. RPC pengaitan principal tersedia dan menunggu verifier bukti provisioning pada startup produksi.

`CreateWorker` dan `UpdateWorker` menolak penulisan referensi principal melalui input CRUD. `GetUserById` di IAM memberikan ID, tipe, dan status akun. Pengaitan masih membutuhkan bukti terverifikasi bahwa akun employee tersebut cocok dengan `person_party_reference_id` pekerja, misalnya melalui keputusan provisioning yang mengikat akun dan worker. Izin `human_resource.workers.link_principal` memerlukan MFA. Provisioning grant tetap menjadi tanggung jawab IAM.

Pembuatan dan perubahan kontrak draf menolak pekerja dengan tanggal terminasi yang sudah tercatat, termasuk data lama yang statusnya masih aktif. Penghapusan draf yang belum bertanda tangan tetap tersedia sesuai aturan kontrak.

`RecordSignedEmploymentContract` menerima ID kontrak draf dan ID versi dokumen. Verifier DRM harus membuktikan isi kontrak serta tanda tangan para pihak dan mengembalikan ID dokumen, versi, waktu, serta digest SHA-256. MFA diperlukan. Startup produksi masih memerlukan verifier DRM.

`LinkWorkerPrincipal` menolak pengaitan jika principal sudah tercatat pada hubungan kerja berakhir milik party orang lain. Riwayat untuk party orang yang sama dapat dipertahankan saat kontrak kerja baru dibuat, dengan verifikasi provisioning tetap wajib.

## Penjelasan GetWorkerAssignmentTransfer

Mengambil tanggal efektif, status, alasan, waktu pengajuan dan waktu keputusan, serta referensi pekerja dan posisi untuk membuka kembali detail transfer.

## Butuh GraphQL Query

`workerAssignmentTransfer(id: ID!)` menerima UUID atau Global ID transfer. Hasil command transfer dan hasil query memakai identitas Node yang sama. Detail dapat dibaca ulang lewat `node(id: ...)`. UI menampilkan nama dan nomor pekerja, kode dan judul posisi, tanggal efektif, dan label status.

## RBAC

Employee memerlukan grant global `human_resource.worker_assignment_transfers.read`. Akses diperiksa di Gateway dan service Personnel. [Bukti HTTPS](../../e2e/gateway/156-hrm-query-detail-transfer-penugasan.md).

## Penjelasan ListWorkerAssignmentTransfers

Riwayat satu pekerja diurutkan dari pengajuan terbaru. ID transfer memecahkan seri bila waktu pengajuan sama.

## Butuh GraphQL Query

`workerAssignmentTransfers(worker_id: ID!, after: String, before: String, first: Int, last: Int)` menerima UUID atau Global ID pekerja. UI membaca `nodes`, `edges`, `pageInfo`, dan `total_count`. Tampilkan label posisi, tanggal efektif, dan status. Detail juga menyediakan `reason_code`, `requested_at`, dan `decided_at`. Gunakan label alasan untuk kode yang dikenal dan format waktu lokal. Transfer yang masih diajukan memiliki `decided_at: null`. Gunakan ID untuk navigasi detail.

## RBAC

Employee memerlukan grant global `human_resource.worker_assignment_transfers.read`. Gateway dan Personnel memeriksa izin sebelum mengembalikan daftar atau total.

## Penjelasan SubmitWorkerAssignmentTransfer dan DecideWorkerAssignmentTransfer

Pengajuan memindahkan penempatan primer ke posisi lain dalam badan hukum yang sama. UI memilih posisi berdasarkan kode dan judul, lalu menampilkan tanggal efektif dan alasan. Keputusan independen menolak permintaan atau menutup penempatan lama dan menerapkan penempatan baru dalam satu transaksi. Retry pengajuan menggunakan idempotency key yang sama. Retry keputusan yang sama mengembalikan assignment yang sudah dibuat.

## Butuh GraphQL Mutation

`submitWorkerAssignmentTransfer(input: SubmitWorkerAssignmentTransferInput!)` dan `decideWorkerAssignmentTransfer(id: ID!, decision: WorkerAssignmentTransferDecision!)` sudah tersedia. Setelah mutation berhasil, UI dapat membaca kembali status melalui `workerAssignmentTransfer` atau memperbarui daftar `workerAssignmentTransfers`. [Bukti alur dua aktor](../../e2e/gateway/114-hrm-mutation-transfer-assignment-mfa.md).

## RBAC

Pengajuan memerlukan grant global `human_resource.worker_assignments.create`. Keputusan memerlukan grant global `human_resource.worker_assignments.transfer`, MFA, serta principal yang berbeda dari pengaju dan pekerja terkait. HRM menggunakan informasi posisi dan badan hukum dari jalur Enterprise Foundation yang sudah ada.

## Referensi tanggal pekerja pada Enterprise Foundation

Saat tanggal masuk atau tanggal mulai kerja tersedia, HRM mencocokkan masing-masing tanggal dengan rentang pendirian dan pembubaran badan hukum aktif di EF. Pekerja preboarding tetap dapat dibuat tanpa tanggal sesuai kontrak. Pemeriksaan berlaku saat referensi pekerja diverifikasi, termasuk pembuatan, koreksi, dan aktivasi. [Adapter EF](../../../backend/services/human_resource/infrastructure/integration/enterprise_foundation/client.go)

Regresi HTTPS dengan superadmin dan grant GraphQL IAM menolak pembuatan pekerja sebelum tanggal pendirian EF, lalu menerima tanggal yang cocok dan menyelesaikan provisioning IAM. Runner memakai database disposable. [Runner](../../../backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py)
