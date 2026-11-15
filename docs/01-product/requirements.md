# Kebutuhan Sistem (SRS), Fungsi (FRD) dan Teknis (TRD)

> **Integrione Retail ERP · 12 November 2026 · status: rancangan produk.** Sumber keputusan produk: [governance](../00-governance/decisions.md). Kebutuhan target bukan bukti implementasi.

## SRS: perilaku yang diwajibkan
| ID | Requirement | Error/negative |
|---|---|---|
| S-01 | IAM memverifikasi user/workload dan scope; resource owner menerapkan ABAC | Invalid session ≠ forbidden ≠ unavailable; deny tidak menulis DB/outbox |
| S-02 | Data master dan transaksi hanya ditulis management pemilik | Tidak ada cross-DB SQL/FK/import internal |
| S-03 | SKU/varian PLM mempunyai versi dan mapping tiap channel | Draft/expired tidak tampil publik |
| S-04 | PO→WMS receipt→Finance matching mempertahankan selisih | Tidak menyamakan ordered qty dengan received qty |
| S-05 | OMS menerima order dari Commerce/marketplace dengan key stabil | Timeout/replay tidak membuat dua sales order |
| S-06 | Pricing/tax/hold sources berversi dan memiliki expiry | Unavailable/expired menahan quote, bukan membuat nilai 0 palsu |
| S-07 | Webhook channel diverifikasi dan dipersist sebelum ACK | Bad signature/replay ditolak, duplicate inbox dedupe |
| S-08 | Analytics product view/order/refund berprovenance | Unknown denominator menjadi N/A; tidak double count |
| S-09 | Customer/guest/CRM consent terisolasi | PII lintas buyer/party ditolak |
| S-10 | Seluruh service mempunyai shutdown/readiness/audit dan recovery | Worker crash/event out of order aman |

## FRD: fungsi bisnis per owner
- **PLM:** definisi produk, variant, identitas SKU dan revisi; **Commerce:** channel, listing, cart/quote, submission.
- **Pricing:** price list/promosi dan keputusan harga; **OMS:** sales order accepted/rejected, cancel/return.
- **Procurement:** PO/ASN; **WMS:** receipt, stock/hold, pick/pack/dispatch; **Finance:** AP/AR/receipt, GL.
- **IAM:** role/grant/scopes dan session; **Foundation:** legal entity/party; **HRM:** worker; **DRM:** document version; **CRM/Customer Service:** lead/tiket.
Semua fungsi wajib memiliki input-output/error, permission, invariant serta test; nama RPC aktual diambil dari source, tidak direka.

## TRD: implementasi yang disepakati
Go workspace multi-module; GraphQL Gateway gqlgen; gRPC proto; PostgreSQL GORM per management; Redis security runtime IAM. Gunakan outbox/inbox transaksi lokal, idempotency, contract version, mTLS/workload identity serta secret vault. Schema/GraphQL aktif harus dicek ke registrasi nyata; tidak ada generator output edit manual.

## NFR dan kualitas
Security/privacy, konsistensi data, availability/fail-closed, latency dan queue lag terukur, accessibility (target perlu persetujuan), backup restore dan interoperability provider. Angka SLO/RTO/RPO belum ditentukan; catat secara terbuka hingga diukur dan disetujui.

## Traceability
Format `BRD-ID → PRD-ID → SRS-ID → domain/service/GraphQL/RPC → test case → SHA → hasil → UAT`. Lihat [tes](../04-engineering/testing.md) dan [project plan](../06-project/project-plan.md).
