# List operasi

`human_resource.compensation.v1.CompensationService` memiliki 25 RPC bisnis. [Kontrak proto](../../../backend/services/human_resource/contract/proto/human_resource/compensation/v1/compensation.proto) memuat definisi kolom dan tipe pesan.

## Penjelasan operasi

| Operasi | Penjelasan | GraphQL | Request → Response | RBAC |
| --- | --- | --- | --- | --- |
| `CreateCompensationBand` | Membuat entri `compensation_bands`. | Mutation | `CreateCompensationBandRequest` → `CompensationBandResponse` | `human_resource.compensation_bands.create` |
| `UpdateCompensationBand` | Mengubah entri `compensation_bands` sesuai field mask. | Mutation | `UpdateCompensationBandRequest` → `CompensationBandResponse` | `human_resource.compensation_bands.update` |
| `DeleteCompensationBand` | Menghapus entri `compensation_bands` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.compensation_bands.delete` |
| `GetCompensationBandByID` | Mengambil entri `compensation_bands` berdasarkan ID. | Query | `IDRequest` → `CompensationBandResponse` | `human_resource.compensation_bands.read` |
| `GetAllCompensationBands` | Mendaftar entri `compensation_bands` dengan paginasi. | Query | `PageRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `FindCompensationBandByBandCode` | Menyaring daftar `compensation_bands` berdasarkan `band_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `FindCompensationBandByJobLevelCode` | Menyaring daftar `compensation_bands` berdasarkan `job_level_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `FindCompensationBandByJurisdictionCode` | Menyaring daftar `compensation_bands` berdasarkan `jurisdiction_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `FindCompensationBandByCurrencyCode` | Menyaring daftar `compensation_bands` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `FindCompensationBandByApprovalReferenceID` | Menyaring daftar `compensation_bands` berdasarkan `approval_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationBandsResponse` | `human_resource.compensation_bands.read` |
| `ApproveCompensationRecord` | Mengesahkan proposal kompensasi melalui penyetuju independen. | Mutation | `IDRequest` → `CompensationRecordResponse` | `human_resource.compensation_records.approve` |
| `CreateCompensationRecord` | Membuat entri `compensation_records`. | Mutation | `CreateCompensationRecordRequest` → `CompensationRecordResponse` | `human_resource.compensation_records.create` |
| `UpdateCompensationRecord` | Mengubah entri `compensation_records` sesuai field mask. | Mutation | `UpdateCompensationRecordRequest` → `CompensationRecordResponse` | `human_resource.compensation_records.update` |
| `DeleteCompensationRecord` | Menghapus entri `compensation_records` berdasarkan ID dengan aturan penghapusan HRM. | Mutation | `IDRequest` → `EmptyResponse` | `human_resource.compensation_records.delete` |
| `GetCompensationRecordByID` | Mengambil entri `compensation_records` berdasarkan ID. | Query | `IDRequest` → `CompensationRecordResponse` | `human_resource.compensation_records.read` |
| `GetAllCompensationRecords` | Mendaftar entri `compensation_records` dengan paginasi. | Query | `PageRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByWorkerID` | Menyaring daftar `compensation_records` berdasarkan `worker_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByAssignmentID` | Menyaring daftar `compensation_records` berdasarkan `assignment_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByComponentID` | Menyaring daftar `compensation_records` berdasarkan `component_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByCompensationBandID` | Menyaring daftar `compensation_records` berdasarkan `compensation_band_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByCompensationType` | Menyaring daftar `compensation_records` berdasarkan `compensation_type`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByCurrencyCode` | Menyaring daftar `compensation_records` berdasarkan `currency_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByStatus` | Menyaring daftar `compensation_records` berdasarkan `status`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByReasonCode` | Menyaring daftar `compensation_records` berdasarkan `reason_code`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordByApprovedByPrincipalID` | Menyaring daftar `compensation_records` berdasarkan `approved_by_principal_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |
| `FindCompensationRecordBySourceChangeReferenceID` | Menyaring daftar `compensation_records` berdasarkan `source_change_reference_id`, dengan paginasi. | Query | `StringFinderRequest` → `CompensationRecordsResponse` | `human_resource.compensation_records.read` |

## Butuh GraphQL

19 kandidat Query dan 7 kandidat Mutation. Gunakan klasifikasi per RPC untuk desain schema. Status publikasi dan pola pemetaan ada pada [indeks handoff](1-inventaris-operasi-grpc-hrm.md).

## RBAC

Kolom RBAC mencatat kode permission yang diminta HRM untuk entry point RPC. Validasi sesi, step-up, scope dan aturan workflow tetap ditegakkan oleh HRM. Lihat [katalog permission](../../../backend/services/human_resource/infrastructure/security/catalog.go) dan [policy akses](../../../backend/services/human_resource/infrastructure/security/policy.go).

`ApproveCompensationRecord` membutuhkan penyetuju manusia dengan izin sensitif dan MFA terbaru. HRM memeriksa peristiwa pembuatan serta perubahan proposal di outbox. Pembuat, setiap editor, dan principal IAM pekerja penerima ditolak sebagai penyetuju. Proposal lama tanpa riwayat yang dapat diverifikasi tidak dapat disetujui lewat command ini. Persetujuan mencatat actor dan waktu server serta status `approved`. Aktivasi efektif dan pembayaran memiliki workflow tersendiri.

HRM menolak persetujuan jika periode proposal bertumpang tindih dengan catatan `approved` atau `active` untuk pekerja, penempatan, dan komponen yang sama. Tanggal awal dan akhir dianggap inklusif.
