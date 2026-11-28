# Customer Relationship & Sales – Finance Accounting – Bukti penjualan neto yang diakui dan commercial credit

- **Role** — Developer Finance Accounting bersama developer CRM
- **Priority** — P1
- **Status** — Usulan kontrak siap ditinjau pemilik
- **Repository dan baseline** — ERP Retail, `main` berbasis `3459732f459439513d9f0c9fc68c5b63959d5a43`, checkout 29 November 2026
- **Target management** — Finance Accounting

## Task

Bukti penjualan neto yang diakui dan commercial credit. Hasil akhir yang dapat diverifikasi adalah sebagai berikut. Jumlah per periode dapat direkonsiliasi terhadap penjualan neto dan credit posted Finance. Pembayaran, quote dan provider settlement tidak menambah achievement.

## Request

### 1. Files

| Aksi | File dan lokasi |
| --- | --- |
| READ | [backend/services/finance_accounting/contract/proto/finance_accounting/v1/reporting.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/reporting.proto) |
| READ | [backend/services/finance_accounting/internal/service/reporting/service.go](../../../backend/services/finance_accounting/internal/service/reporting/service.go) |
| READ | [docs/03-managements/oms-and-commerce.md](../../../docs/03-managements/oms-and-commerce.md) |
| CREATE jika diperlukan | `backend/services/finance_accounting/contract/proto/finance_accounting/v1/sales_performance_evidence.proto` |

### 2. Functions / Methods / Operations

- file: Gunakan tabel Files sebagai titik masuk. Kontrak yang diberi label usulan perlu dibuat atau disepakati oleh pemilik.
- function: `ListRecognizedSalesEvidence (usulan privat), ApplyRecognizedSalesEvidence (CRM)`
- logic: Reporting Finance menyediakan laporan GL. CRM memerlukan proyeksi per sumber penjualan dan per baris yang dapat diatribusikan. Berikan feed replayable atau RPC incremental terscope dengan cursor/watermark. Kembalikan ID posting/line, order/line asal, original sale untuk credit, legal entity, currency, amount net merchandise/service, accounting date, direction, source version dan digest. Sertakan koreksi/reversal/credit posted. Backfill memakai kontrak sama. CRM mencocokkan link credit beku sesuai tanggal dan menampilkan coverage/unassigned. Net di luar tax dan shipping sesuai definisi KPI.
- error_handling: Saat original posting belum ada, tahan credit untuk reconciliation. Cursor tidak boleh meloncat karena timestamp sama. Gunakan sequence/tie-breaker. Owner unavailable menghasilkan status stale/unavailable dan watermark, dengan nilai nol hanya jika terbukti.
- security: Workload CRM hanya berhak pada proyeksi scope legal entity yang disetujui. Abaikan data rekening, pembayaran individu, tax number, margin dan detail GL yang tak dibutuhkan. Pencabutan akses berlaku saat membaca ulang.
- expected_result: Jumlah per periode dapat direkonsiliasi terhadap penjualan neto dan credit posted Finance. Pembayaran, quote dan provider settlement tidak menambah achievement.
- testing: Posted revenue. Return dengan credit di periode berikutnya. Duplicate dan reversal. Sale belum teratribusi. Snapshot tua dan replay incremental. Scope entity berbeda.

### 3. Integration Requirements

Pertukaran fakta antara Finance Accounting dan CRM mengikuti identitas producer, legal entity, subject, source version/digest dan correlation ID. Pertahankan kepemilikan data serta transaksi terpisah. Untuk delivery event, ACK diterbitkan sesudah inbox commit. Pisahkan trace ID percobaan dari idempotency key bisnis.

## Action

1. Cocokkan kebutuhan di atas dengan jalur caller sebenarnya dan kontrak owner terkini. Pakai kembali method yang mencukupi.
2. Implementasikan producer dan consumer pada modul masing-masing, dengan adapter dan feature readiness yang terukur. Jangan mengubah basis data management lain.
3. Verifikasi kasus pada tabel dari entry point yang akan dipakai. Gunakan unit test pada layer services gRPC dan HTTPS dengan konfigurasi role/scope melalui GraphQL IAM untuk jalur manusia.
4. Kirim bukti kontrak, source SHA, ID uji tersamarkan, versi fakta, hasil retry dan status readiness ke developer CRM. Untuk dependency tertunda, catat alasan dan operasi yang tetap tersedia.

| Kasus | Harus diverifikasi |
| --- | --- |
| TC-01 | Posted revenue |
| TC-02 | Return dengan credit di periode berikutnya |
| TC-03 | Duplicate dan reversal |
| TC-04 | Sale belum teratribusi |
| TC-05 | Snapshot tua dan replay incremental |
| TC-06 | Scope entity berbeda |

## Context

[Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Katalog CRM dan jenis operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md) · [Urutan wiring lintas management](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md).
