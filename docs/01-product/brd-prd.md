# Kebutuhan Bisnis (BRD) dan Produk (PRD)

> **Integrione Retail ERP · 12 November 2026 · status: rancangan produk.** Sumber keputusan produk: [governance](../00-governance/decisions.md). Kebutuhan target bukan bukti implementasi.

## BRD: kebutuhan dan outcome
| BRD ID | Masalah / kebutuhan | Bukti keberhasilan |
|---|---|---|
| BR-01 | Pekerjaan order/stok/finance terpisah dan banyak input ulang | Workflow ERP internal dengan sumber data owner yang konsisten |
| BR-02 | Vendor/supplier memerlukan transparansi PO/dokumen tanpa melihat bisnis lain | Akses IAM scoped party, actor negatif dan expiry teruji |
| BR-03 | Stok dan order website/marketplace rentan beda atau ganda | OMS idempoten dan WMS hold dengan rekonsiliasi |
| BR-04 | Pemilik ingin analisis produk dan efektivitas kanal | Metrik SKU/channel dengan source, currency, watermark dan correction |
| BR-05 | Pelanggan butuh website aman dan nyaman | UAT mobile/desktop dan E2E checkout/after-sales sesudah fase internal selesai |

Tidak ada target pendapatan, SLA, volume pengguna atau tanggal produksi yang disetujui; jangan mencantumkan angka placeholder sebagai fakta.

## PRD: fitur prioritas
| PRD ID | Fase | User story | Acceptance |
|---|---|---|---|
| P-01 | F1 | Owner mengatur perusahaan dan staff | Scope IAM/EF valid, negatif tidak menulis data |
| P-02 | F1 | Staf mengelola SKU/varian | Unik/canonical PLM dan status listing terpisah |
| P-03 | F1 | Purchasing membuat PO; gudang menerima parsial | Procurement actual commitment, WMS stock valid, Finance mismatch jelas |
| P-04 | F1 | Owner memeriksa laporan penjualan/biaya | Semua angka dari Finance/OMS/WMS, bukan input manual palsu |
| P-05 | F2 | Supplier melihat PO sendiri dan membuat ASN | Wrong-party 403/NOT_FOUND aman; duplicate idempotent |
| P-06 | F2 | Vendor mengakses dokumen/kontraknya | Hak tidak melebar ke supplier lainnya |
| P-07 | F3 | Admin menghubungkan toko marketplace | Token aman; official permission; event verified/reconciled |
| P-08 | F3 | Owner melihat performa produk lintas kanal | Definition of metric dan missing data N/A |
| P-09 | F4 | Calon pembeli mencari dan memilih produk | Hanya listing LIVE, variannya jelas, tidak bocor stok internal |
| P-10 | F4 | Buyer checkout dan melacak order | Quote sealed; OMS satu order; Finance satu sumber uang; retur via owner |

## Produk yang sengaja tidak dibangun sekarang
Manufacturing baru, produk finansial/bank transfer otomatis, shared multitenant tanpa ADR, konektor marketplace tanpa izin API, janji sertifikasi kepatuhan, serta F4 public store sebelum auth/customer guard siap.

## Penutupan kebutuhan
Setiap requirement punya domain owner, status, acceptance/test ID, dependency, reviewer. Hanya owner boleh menaikkan kebutuhan ke kontrak komersial yang mengikat.
