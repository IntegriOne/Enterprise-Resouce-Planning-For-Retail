# Customer Relationship & Sales – Customer Service – Handoff dukungan sales dan status tiket milik Customer Service

- **Role** — Developer Customer Service bersama developer CRM
- **Priority** — Bersyarat pada kesiapan owner
- **Status** — Usulan kontrak siap ditinjau pemilik
- **Repository dan baseline** — ERP Retail, `main` berbasis `3459732f459439513d9f0c9fc68c5b63959d5a43`, checkout 29 November 2026
- **Target management** — Customer Service

## Task

Handoff dukungan sales dan status tiket milik Customer Service. Hasil akhir yang dapat diverifikasi adalah sebagai berikut. Nomor tiket berasal dari CS. Permintaan CRM yang diulang memulihkan nomor yang sama. Operator tidak dapat menandai refund selesai dari CRM.

## Request

### 1. Files

| Aksi | File dan lokasi |
| --- | --- |
| READ | [agents/context/customer_service_management/1-context.md](../../../agents/context/customer_service_management/1-context.md) |
| READ | [docs/03-managements/crm-and-customer-service.md](../../../docs/03-managements/crm-and-customer-service.md) |
| READ | [agents/boundary.md](../../../agents/boundary.md) |
| READ | [backend/services/web_commerce/services/aftersales/service.go](../../../backend/services/web_commerce/services/aftersales/service.go) |
| DESIGN | `Kontrak usulan privat Customer Service ketika modul aktif, sesuai path modul yang disepakati owner` |

### 2. Functions / Methods / Operations

- file: Gunakan tabel Files sebagai titik masuk. Kontrak yang diberi label usulan perlu dibuat atau disepakati oleh pemilik.
- function: `RequestRelationshipSupport (CRM), RequestSupportHandoff dan GetSupportHandoffStatus (usulan CS), ApplySupportHandoffResult (CRM)`
- logic: Terima ID request stabil, entity/unit, subject/contact ref, alasan, sumber dan evidence ref minimum. Satu request menghasilkan satu tiket CS. Kembalikan nomor display, status terizinkan, versi dan waktu. CRM menyimpan korelasi/proyeksi. CS mengelola first response aktual, SLA, assignment, reopen dan escalation. CS mengajukan return/refund ke Commerce/OMS/Finance sesuai pemilik, menunggu hasil resmi. Pada checkout belum ada implementasi server CS di backend/services. Serahkan sendiri sebagai dependency bersyarat.
- error_handling: ACK pending saat owner belum terhubung. Request key sama payload berbeda ditolak. Outcome unknown atau tindakan owner belum tuntas memblokir klaim resolved. Dukungan manual tercatat sebagai activity CRM selama kontrak belum aktif.
- security: Penerima memverifikasi delegasi dan kepemilikan order/subjek, membatasi detail kepada agent tiket. Izin melihat CRM tidak otomatis memberi hak melihat case atau uang.
- expected_result: Nomor tiket berasal dari CS. Permintaan CRM yang diulang memulihkan nomor yang sama. Operator tidak dapat menandai refund selesai dari CRM.
- testing: Handoff sah. Duplicate request. Wrong subject/order. CS offline. Reject dengan alasan dan status unknown. Replay status lama tidak memundurkan versi.

### 3. Integration Requirements

Pertukaran fakta antara Customer Service dan CRM mengikuti identitas producer, legal entity, subject, source version/digest dan correlation ID. Pertahankan kepemilikan data serta transaksi terpisah. Untuk delivery event, ACK diterbitkan sesudah inbox commit. Pisahkan trace ID percobaan dari idempotency key bisnis.

## Action

1. Cocokkan kebutuhan di atas dengan jalur caller sebenarnya dan kontrak owner terkini. Pakai kembali method yang mencukupi.
2. Implementasikan producer dan consumer pada modul masing-masing, dengan adapter dan feature readiness yang terukur. Jangan mengubah basis data management lain.
3. Verifikasi kasus pada tabel dari entry point yang akan dipakai. Gunakan unit test pada layer services gRPC dan HTTPS dengan konfigurasi role/scope melalui GraphQL IAM untuk jalur manusia.
4. Kirim bukti kontrak, source SHA, ID uji tersamarkan, versi fakta, hasil retry dan status readiness ke developer CRM. Untuk dependency tertunda, catat alasan dan operasi yang tetap tersedia.

| Kasus | Harus diverifikasi |
| --- | --- |
| TC-01 | Handoff sah |
| TC-02 | Duplicate request |
| TC-03 | Wrong subject/order |
| TC-04 | CS offline |
| TC-05 | Reject dengan alasan dan status unknown |
| TC-06 | Replay status lama tidak memundurkan versi |

## Context

[Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Katalog CRM dan jenis operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md) · [Urutan wiring lintas management](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md).
