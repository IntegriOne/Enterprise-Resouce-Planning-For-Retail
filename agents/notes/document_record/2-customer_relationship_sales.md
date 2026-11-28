# Customer Relationship & Sales – Document Record – Keputusan akses dan scan untuk lampiran CRM

- **Role** — Developer Document Record bersama developer CRM
- **Priority** — P1
- **Status** — Usulan kontrak siap ditinjau pemilik
- **Repository dan baseline** — ERP Retail, `main` berbasis `3459732f459439513d9f0c9fc68c5b63959d5a43`, checkout 29 November 2026
- **Target management** — Document Record

## Task

Keputusan akses dan scan untuk lampiran CRM. Hasil akhir yang dapat diverifikasi adalah sebagai berikut. File tersedia hanya untuk aktor yang boleh mengakses versi clean tersebut. Pencabutan hak antara attach dan send menghentikan penggunaan.

## Request

### 1. Files

| Aksi | File dan lokasi |
| --- | --- |
| READ | [backend/services/document_record/contract/proto/document_record/v1/configuration_document.proto](../../../backend/services/document_record/contract/proto/document_record/v1/configuration_document.proto) |
| READ | [backend/services/document_record/services/document/service.go](../../../backend/services/document_record/services/document/service.go) |
| READ | [backend/services/document_record/contract/proto/document_record/v1/ingestion_publication.proto](../../../backend/services/document_record/contract/proto/document_record/v1/ingestion_publication.proto) |
| MODIFY bila kontrak existing belum cukup | `backend/services/document_record/contract/proto/document_record/v1/ingestion_publication.proto` |

### 2. Functions / Methods / Operations

- file: Gunakan tabel Files sebagai titik masuk. Kontrak yang diberi label usulan perlu dibuat atau disepakati oleh pemilik.
- function: `VerifyDocumentVersion, StreamDocument, pemeriksaan kelayakan lampiran untuk actor dan purpose`
- logic: VerifyDocumentVersion menerima exact version, digest, owner dan source. Service yang diperiksa mengecek status available dan binding tersebut. Respons ini belum menyatakan hasil scan maupun izin recipient untuk percakapan. Temukan jalur terotorisasi dan bukti scan di DRM, susun keduanya untuk read/send. Bila belum tersedia, tambahkan respons/RPC workload terbatas (usulan AuthorizeRelationshipAttachment) yang memeriksa principal/workload, action, audience, owner, source, version, digest, clean, revoked/hold, dan tujuan. Data file serta stream tetap di DRM. Metadata terkunci sampai syarat terpenuhi.
- error_handling: Mismatch/revoked/blocked → tolak. Pending scan atau DRM timeout → status pending/unavailable, tanpa URL atau bytes. Cek ulang saat view/send dan cocokkan revocation saat pengiriman.
- security: mTLS, delegasi user sah, grant dokumen pada recipient dan scope percakapan. Jangan menerima actor/hasil scan kiriman browser sebagai bukti. Hold memengaruhi retensi dan penghapusan sesuai policy owner.
- expected_result: File tersedia hanya untuk aktor yang boleh mengakses versi clean tersebut. Pencabutan hak antara attach dan send menghentikan penggunaan.
- testing: Versi benar. Digest salah. User unit lain dan peserta salah. Pending/blocked. Revoked setelah attach. Retensi/hold saat anonimisasi.

### 3. Integration Requirements

Pertukaran fakta antara Document Record dan CRM mengikuti identitas producer, legal entity, subject, source version/digest dan correlation ID. Pertahankan kepemilikan data serta transaksi terpisah. Untuk delivery event, ACK diterbitkan sesudah inbox commit. Pisahkan trace ID percobaan dari idempotency key bisnis.

## Action

1. Cocokkan kebutuhan di atas dengan jalur caller sebenarnya dan kontrak owner terkini. Pakai kembali method yang mencukupi.
2. Implementasikan producer dan consumer pada modul masing-masing, dengan adapter dan feature readiness yang terukur. Jangan mengubah basis data management lain.
3. Verifikasi kasus pada tabel dari entry point yang akan dipakai. Gunakan unit test pada layer services gRPC dan HTTPS dengan konfigurasi role/scope melalui GraphQL IAM untuk jalur manusia.
4. Kirim bukti kontrak, source SHA, ID uji tersamarkan, versi fakta, hasil retry dan status readiness ke developer CRM. Untuk dependency tertunda, catat alasan dan operasi yang tetap tersedia.

| Kasus | Harus diverifikasi |
| --- | --- |
| TC-01 | Versi benar |
| TC-02 | Digest salah |
| TC-03 | User unit lain dan peserta salah |
| TC-04 | Pending/blocked |
| TC-05 | Revoked setelah attach |
| TC-06 | Retensi/hold saat anonimisasi |

## Context

[Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Katalog CRM dan jenis operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md) · [Urutan wiring lintas management](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md).
