# Customer Relationship & Sales – Web Commerce – Rujukan opportunity, quotation berbantuan dan bukti order

- **Role** — Developer Web Commerce bersama developer CRM
- **Priority** — P1
- **Status** — Usulan kontrak siap ditinjau pemilik
- **Repository dan baseline** — ERP Retail, `main` berbasis `3459732f459439513d9f0c9fc68c5b63959d5a43`, checkout 29 November 2026
- **Target management** — Web Commerce

## Task

Rujukan opportunity, quotation berbantuan dan bukti order. Hasil akhir yang dapat diverifikasi adalah sebagai berikut. Satu quotation untuk satu permintaan yang diulang, hubungan dapat ditelusuri dua arah, dan confirmed/cancelled order memengaruhi opportunity sekali per versi.

## Request

### 1. Files

| Aksi | File dan lokasi |
| --- | --- |
| READ | [backend/services/web_commerce/contract/proto/web_commerce/v1/quotation_order_aftersales.proto](../../../backend/services/web_commerce/contract/proto/web_commerce/v1/quotation_order_aftersales.proto) |
| READ | [backend/services/web_commerce/transport/grpc/handler/quotation_order_aftersales.go](../../../backend/services/web_commerce/transport/grpc/handler/quotation_order_aftersales.go) |
| READ | [backend/services/web_commerce/services/quotation/service.go](../../../backend/services/web_commerce/services/quotation/service.go) |
| READ | [backend/services/web_commerce/services/aftersales/service.go](../../../backend/services/web_commerce/services/aftersales/service.go) |
| READ | [docs/03-managements/oms-and-commerce.md](../../../docs/03-managements/oms-and-commerce.md) |
| CREATE jika diperlukan | `backend/services/web_commerce/infrastructure/integration/customer_relationship_sales/client.go` |

### 2. Functions / Methods / Operations

- file: Gunakan tabel Files sebagai titik masuk. Kontrak yang diberi label usulan perlu dibuat atau disepakati oleh pemilik.
- function: `RequestSalesQuotation, VerifySalesOpportunityContext, RequestAssistedQuotation, ApplyVerifiedSalesDocument, ApplySalesDocumentReversal`
- logic: Handler Commerce sudah mengurai crm_opportunity_reference_id dan service menyimpannya pada quote. Saat ID diberikan, verifikasi konteks CRM lebih dahulu. Pertahankan jalur B2B buyer yang sudah ada dan jalur normal tanpa reference. Untuk permintaan petugas sales sediakan kontrak privat atau workflow delegasi yang memeriksa mandat petugas, badan hukum, party/PIC, kebutuhan produk dan currency. Quote dibuat oleh Commerce dengan harga/terms owner. CRM menyimpan request key sebelum mengirim, lalu menerima nomor dan versi. Setelah OMS mengonfirmasi order, Commerce mengirim bukti OMS berisi ID sumber, versi, digest, order/line, qty, net dan waktu. CRM membekukan credit berdasarkan fakta ini. Jangan mensyaratkan link CRM sebelum Commerce bisa membuat dokumen.
- error_handling: UNAVAILABLE untuk verifier wajib saat assisted, FAILED_PRECONDITION untuk party/opportunity tidak cocok, replay identik mendapatkan ID quote yang sama. Jika hasil create timeout, lookup berdasarkan key sebelum mengulang. Quote-only tetap open.
- security: mTLS khusus CRM, buyer consent/delegation dan grant unit/party tervalidasi. Pertahankan otorisasi buyer pada jalur yang ada, termasuk B2C tanpa dummy party. Trust callback hanya setelah bukti owner tervalidasi.
- expected_result: Satu quotation untuk satu permintaan yang diulang, hubungan dapat ditelusuri dua arah, dan confirmed/cancelled order memengaruhi opportunity sekali per versi.
- testing: Assisted B2B sah. Buyer lain dan unit lain ditolak. B2C tanpa CRM tetap jalan. Timeout setelah create dan retry. Accepted quote tanpa konfirmasi OMS. Partial fulfill/cancel dengan event terbalik.

### 3. Integration Requirements

Pertukaran fakta antara Web Commerce dan CRM mengikuti identitas producer, legal entity, subject, source version/digest dan correlation ID. Pertahankan kepemilikan data serta transaksi terpisah. Untuk delivery event, ACK diterbitkan sesudah inbox commit. Pisahkan trace ID percobaan dari idempotency key bisnis.

## Action

1. Cocokkan kebutuhan di atas dengan jalur caller sebenarnya dan kontrak owner terkini. Pakai kembali method yang mencukupi.
2. Implementasikan producer dan consumer pada modul masing-masing, dengan adapter dan feature readiness yang terukur. Jangan mengubah basis data management lain.
3. Verifikasi kasus pada tabel dari entry point yang akan dipakai. Gunakan unit test pada layer services gRPC dan HTTPS dengan konfigurasi role/scope melalui GraphQL IAM untuk jalur manusia.
4. Kirim bukti kontrak, source SHA, ID uji tersamarkan, versi fakta, hasil retry dan status readiness ke developer CRM. Untuk dependency tertunda, catat alasan dan operasi yang tetap tersedia.

| Kasus | Harus diverifikasi |
| --- | --- |
| TC-01 | Assisted B2B sah |
| TC-02 | Buyer lain dan unit lain ditolak |
| TC-03 | B2C tanpa CRM tetap jalan |
| TC-04 | Timeout setelah create dan retry |
| TC-05 | Accepted quote tanpa konfirmasi OMS |
| TC-06 | Partial fulfill/cancel dengan event terbalik |

## Context

[Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Katalog CRM dan jenis operasi](../../../agents/grpc/customer_relationship_sales/1-katalog-operasi-relasi-pelanggan-dan-penjualan.md) · [Urutan wiring lintas management](../../../agents/instructions/customer_relationship_sales/2-leader-wiring-lintas-management-instructions.md).
