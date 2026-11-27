# List operasi

`human_resource.leave.v1.LeaveService` memiliki 44 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/leave/v1/leave.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateLeavePolicy` | Membuat entri `leave_policies`. | Mutation | `CreateLeavePolicyRequest` → `LeavePolicyResponse` | `human_resource.leave_policies.create` |
| `UpdateLeavePolicy` | Mengubah entri `leave_policies` sesuai field mask. | Mutation | `UpdateLeavePolicyRequest` → `LeavePolicyResponse` | `human_resource.leave_policies.update` |
| `DeleteLeavePolicy` | Menghapus entri `leave_policies` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.leave_policies.delete` |
| `GetLeavePolicyByID` | Mengambil entri `leave_policies` berdasarkan ID. | Query | `IDRequest` → `LeavePolicyResponse` | `human_resource.leave_policies.read` |
| `GetAllLeavePolicies` | Mendaftar entri `leave_policies` dengan paginasi. | Query | `PageRequest` → `LeavePoliciesResponse` | `human_resource.leave_policies.read` |
| `FindLeavePolicyByPolicyCode` | Menyaring daftar `leave_policies` berdasarkan `policy_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeavePoliciesResponse` | `human_resource.leave_policies.read` |
| `FindLeavePolicyByLeaveTypeCode` | Menyaring daftar `leave_policies` berdasarkan `leave_type_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeavePoliciesResponse` | `human_resource.leave_policies.read` |
| `FindLeavePolicyByJurisdictionCode` | Menyaring daftar `leave_policies` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeavePoliciesResponse` | `human_resource.leave_policies.read` |
| `CreateLeaveEntitlement` | Membuat entri `leave_entitlements`. | Mutation | `CreateLeaveEntitlementRequest` → `LeaveEntitlementResponse` | `human_resource.leave_entitlements.create` |
| `UpdateLeaveEntitlement` | Mengubah entri `leave_entitlements` sesuai field mask. | Mutation | `UpdateLeaveEntitlementRequest` → `LeaveEntitlementResponse` | `human_resource.leave_entitlements.update` |
| `DeleteLeaveEntitlement` | Menghapus entri `leave_entitlements` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.leave_entitlements.delete` |
| `GetLeaveEntitlementByID` | Mengambil entri `leave_entitlements` berdasarkan ID. | Query | `IDRequest` → `LeaveEntitlementResponse` | `human_resource.leave_entitlements.read` |
| `GetAllLeaveEntitlements` | Mendaftar entri `leave_entitlements` dengan paginasi. | Query | `PageRequest` → `LeaveEntitlementsResponse` | `human_resource.leave_entitlements.read` |
| `FindLeaveEntitlementByWorkerID` | Menyaring daftar `leave_entitlements` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveEntitlementsResponse` | `human_resource.leave_entitlements.read` |
| `FindLeaveEntitlementByLeavePolicyID` | Menyaring daftar `leave_entitlements` berdasarkan `leave_policy_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveEntitlementsResponse` | `human_resource.leave_entitlements.read` |
| `FindLeaveEntitlementByUOMCode` | Menyaring daftar `leave_entitlements` berdasarkan `uom_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveEntitlementsResponse` | `human_resource.leave_entitlements.read` |
| `FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID` | Menyaring daftar `leave_entitlements` berdasarkan `source_workforce_snapshot_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveEntitlementsResponse` | `human_resource.leave_entitlements.read` |
| `CreateLeaveAdjustment` | Membuat entri `leave_adjustments`. | Mutation | `CreateLeaveAdjustmentRequest` → `LeaveAdjustmentResponse` | `human_resource.leave_adjustments.create` |
| `UpdateLeaveAdjustment` | Mengubah entri `leave_adjustments` sesuai field mask. | Mutation | `UpdateLeaveAdjustmentRequest` → `LeaveAdjustmentResponse` | `human_resource.leave_adjustments.update` |
| `DeleteLeaveAdjustment` | Menghapus entri `leave_adjustments` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.leave_adjustments.delete` |
| `GetLeaveAdjustmentByID` | Mengambil entri `leave_adjustments` berdasarkan ID. | Query | `IDRequest` → `LeaveAdjustmentResponse` | `human_resource.leave_adjustments.read` |
| `GetAllLeaveAdjustments` | Mendaftar entri `leave_adjustments` dengan paginasi. | Query | `PageRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `FindLeaveAdjustmentByEntitlementID` | Menyaring daftar `leave_adjustments` berdasarkan `entitlement_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `FindLeaveAdjustmentByReasonCode` | Menyaring daftar `leave_adjustments` berdasarkan `reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `FindLeaveAdjustmentByRequestedByPrincipalID` | Menyaring daftar `leave_adjustments` berdasarkan `requested_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `FindLeaveAdjustmentByApprovedByPrincipalID` | Menyaring daftar `leave_adjustments` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `FindLeaveAdjustmentByIdempotencyKey` | Menyaring daftar `leave_adjustments` berdasarkan `idempotency_key`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveAdjustmentsResponse` | `human_resource.leave_adjustments.read` |
| `RejectLeaveAdjustment` | Menolak entri `leave_adjustments` melalui transisi workflow HRM. | Mutation | `RejectLeaveAdjustmentRequest` → `LeaveAdjustmentResponse` | `human_resource.leave_adjustments.reject` |
| `ApproveLeaveAdjustment` | Menyetujui entri `leave_adjustments` melalui transisi workflow HRM. | Mutation | `IDRequest` → `LeaveAdjustmentResponse` | `human_resource.leave_adjustments.approve` |
| `CreateLeaveLedgerEntry` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `CreateLeaveLedgerEntryRequest` → `LeaveLedgerEntryResponse` | — |
| `UpdateLeaveLedgerEntry` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `UpdateLeaveLedgerEntryRequest` → `LeaveLedgerEntryResponse` | — |
| `DeleteLeaveLedgerEntry` | RPC kompatibilitas untuk resource baca-saja. | Nonaktif | `IDRequest` → `EmptyResponse` | — |
| `GetLeaveLedgerEntryByID` | Mengambil entri `leave_ledger_entries` berdasarkan ID. | Query | `IDRequest` → `LeaveLedgerEntryResponse` | `human_resource.leave_ledger_entries.read` |
| `GetAllLeaveLedgerEntries` | Mendaftar entri `leave_ledger_entries` dengan paginasi. | Query | `PageRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByEntitlementID` | Menyaring daftar `leave_ledger_entries` berdasarkan `entitlement_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByEntryType` | Menyaring daftar `leave_ledger_entries` berdasarkan `entry_type`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByAdjustmentID` | Menyaring daftar `leave_ledger_entries` berdasarkan `adjustment_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByReversesEntryID` | Menyaring daftar `leave_ledger_entries` berdasarkan `reverses_entry_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryBySourceManagement` | Menyaring daftar `leave_ledger_entries` berdasarkan `source_management`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryBySourceEventID` | Menyaring daftar `leave_ledger_entries` berdasarkan `source_event_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByReasonCode` | Menyaring daftar `leave_ledger_entries` berdasarkan `reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `FindLeaveLedgerEntryByActorPrincipalID` | Menyaring daftar `leave_ledger_entries` berdasarkan `actor_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `LeaveLedgerEntriesResponse` | `human_resource.leave_ledger_entries.read` |
| `CloseLeaveEntitlement` | Menutup entri `leave_entitlements` melalui transisi workflow HRM. | Mutation | `IDRequest` → `LeaveEntitlementResponse` | `human_resource.leave_entitlements.close` |
| `PostLedgerEntry` | Membukukan entri ledger pada hak cuti dan memvalidasi saldo serta referensi penyesuaian terkait. | Mutation | `PostLedgerEntryRequest` → `LeaveEntitlementResponse` | `human_resource.leave_entitlements.post` |

## Butuh GraphQL

28 kandidat Query dan 13 kandidat Mutation dan 3 RPC nonaktif. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

`CloseLeaveEntitlement` mensyaratkan akhir periode sudah lewat, saldo nol, dan seluruh koreksi yang disetujui sudah tercatat dalam ledger. Koreksi draf masih dapat ditolak setelah penutupan.

`PostLedgerEntry` jenis `expiry` juga menolak kedaluwarsa selama penyesuaian yang disetujui belum dibukukan.

`ApproveLeaveAdjustment` memperhitungkan koreksi searah yang sudah disetujui dan belum dibukukan. Koreksi pengurang tidak boleh bersama-sama melampaui saldo, dan koreksi penambah tidak boleh bersama-sama melewati batas maksimum.

Pembalikan saldo awal melalui `PostLedgerEntry` juga menjaga saldo yang dialokasikan untuk koreksi pengurang yang sudah disetujui dan belum dibukukan.

Kasus persetujuan serentak dicakup oleh [tes integrasi koreksi cuti](../../e2e/human_resource/38-leave-mutation-approve-leave-adjustment.md).
