# Finance Accounting – web_commerce – Pembayaran checkout dan refund

- **Role** — Agent backend web_commerce, koordinasi dengan owner Finance
- **Priority** — P1, setelah jalur internal dan mitra
- **Status** — SIAP DIKERJAKAN
- **Repository dan baseline** — ERP Retail, `main` berbasis `bba5f46cfec19129dce58b8a70d813e19f4ce3a1`, checkout 29 November 2026
- **Target management** — web_commerce

## Task

Client Commerce selalu mengembalikan false pada IsPrepaidContractReady. Kalkulasi pajak, pembuatan sesi pembayaran dan refund mengembalikan dependency unavailable. Kontrak Finance yang diperiksa belum memuat lifecycle dana pelanggan ini. Split shipment hanya mengonsumsi bagian didanai, refund dibatasi saldo aktual dan pendapatan dapat ditelusuri ke pemenuhan. Finance memberi status terverifikasi ke Commerce. Owner Finance mempertahankan kepemilikan jurnal dan uang. [Instruksi Finance](../../../agents/instructions/finance_accounting/9-leader-remediasi-logika-keuangan-dan-integrasi-instructions.md) · [Bukti audit](../../../agents/audit/finance_accounting/10-2026-11-29-010010-audit-logika-keuangan-dan-integrasi.md)

## Request

### 1. Files

| Aksi | File aktual atau tujuan |
| --- | --- |
| MODIFY | `backend/services/web_commerce/infrastructure/integration/finance/client.go:49` |
| MODIFY | `backend/services/web_commerce/services/checkout/service.go` |
| READ / Finance | `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto` |
| READ / Finance | `backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto` |
| READ / Finance | `backend/services/finance_accounting/infrastructure/inbox/processor.go` |
| READ / Finance | `backend/services/finance_accounting/infrastructure/outbox/publisher.go` |

### 2. Functions / Methods / Operations

- file: Gunakan titik masuk pada tabel di atas. Unit test berada pada layer service yang menangani operasi.
- function: CalculateSalesTax, RequestPaymentSession, RequestCustomerRefund, ReserveFulfillmentFunding, ReleaseFulfillmentFunding
- logic: Bangun kontrak privat Finance untuk quote pajak, payment session, status dana, reserve/release/consume per shipment, refund dan PSP settlement. Setelah bukti penerimaan verified, catat uang muka pelanggan. Gunakan bukti fulfillment sesuai kebijakan untuk revenue/COGS. Jurnal biaya provider dan net bank direkonsiliasi terpisah. Hormati prioritas internal → mitra → pelanggan.
- error_handling: Retry memakai sumber dan key yang sama. Satu reservasi tidak boleh sekaligus refund dan dipakai dispatch. Timeout menunggu status terverifikasi. Snapshot cap refund setelah alokasi, refund, return dan reversal.
- security: Workload Commerce/OMS/WMS sesuai peran, validasi owner order dan legal entity. Bedakan jalur collection/refund dari adapter payout pemasok. Status browser tidak menjadi bukti penerimaan.
- expected_result: Split shipment hanya mengonsumsi bagian didanai, refund dibatasi saldo aktual dan pendapatan dapat ditelusuri ke pemenuhan. Finance memberi status terverifikasi ke Commerce.
- testing: Satu sesi pada retry checkout · Reservasi dua shipment tidak melebihi dana · Refund bersamaan dengan dispatch · Event provider berulang dan terbalik urutan · Timeout lalu status terverifikasi

### 3. Integration Requirements

Commerce meminta quote pajak, sesi pembayaran dan refund. Finance mengembalikan referensi serta status ekonomi per order/fulfillment. Aktifkan IsPrepaidContractReady hanya bila dependensi dan implementasi nyata sehat. Cocokkan nominal, mata uang, scope dan customer order di kedua sisi. Pakai gRPC privat beridentitas workload dan inbox/outbox idempoten. Pertahankan ID, versi dan digest untuk menelusuri respons yang sah saat recipient pulih dari timeout.

## Action

1. Cocokkan versi kontrak dan adapter aktual kedua owner, termasuk data lama.
2. Perbarui service dan adapter milik web_commerce bersama penerbit kontrak Finance.
3. Uji service dan panggil jalur pengguna melalui HTTPS GraphQL. Kelola grant uji melalui GraphQL IAM.
4. Laporkan SHA, bukti kasus di bawah, alur retry dan hasil rekonsiliasi ke owner Finance.

| Kasus | Skenario yang harus dibuktikan |
| --- | --- |
| TC-01 | Satu sesi pada retry checkout |
| TC-02 | Reservasi dua shipment tidak melebihi dana |
| TC-03 | Refund bersamaan dengan dispatch |
| TC-04 | Event provider berulang dan terbalik urutan |
| TC-05 | Timeout lalu status terverifikasi |

## Context

[AGENTS.md](../../../AGENTS.md) · [agents/arsitektur.md](../../../agents/arsitektur.md) · [agents/boundary.md](../../../agents/boundary.md) · [agents/users.md](../../../agents/users.md). [Matriks API nyata dan usulan kontrak](../../../agents/grpc/finance_accounting/7-matriks-operasi-publik-privat-dan-usulan-keuangan.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
