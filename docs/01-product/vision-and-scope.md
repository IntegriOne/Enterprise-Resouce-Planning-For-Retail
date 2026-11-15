# Visi, Scope, dan Fase Produk

> **Integrione Retail ERP · 12 November 2026 · status: rancangan produk.** Sumber keputusan produk: [governance](../00-governance/decisions.md). Kebutuhan target bukan bukti implementasi.

## Visi
Menyediakan **ERP internal + website perdagangan produk jadi** bagi UMKM/pengusaha/brand, dengan sinkronisasi stok, supplier, order dan finance, kemudian marketplace resmi dan analitik SKU. Nilai bisnisnya satu operasi yang terhubung, bukan antarmuka untuk 32 microservice.

## 13 management prioritas
| Management | Peran utama |
|---|---|
| IAM | Principal internal/vendor/supplier/customer/workload, login, grant dan audit |
| Enterprise Foundation | Badan hukum, unit, lokasi, supplier/vendor/customer party |
| Product Lifecycle Management | Master item, SKU, varian/revisi dan kelayakan produk |
| E-Commerce Commerce | Channel/etalase, konten listing, buyer, cart, checkout dan submission |
| Pricing & Revenue | Harga, kupon, diskon, kebijakan dan expiry quote harga |
| OMS | Sales order lifecycle, cancel/return business decisions |
| Inventory & Warehouse | Stock fisik, receiving, holds, pick/pack/dispatch |
| Finance & Accounting | AP/AR, receipt manual, GL, pajak, biaya dan laporan |
| Procurement & Sourcing | Supplier PO, sourcing, ASN, komitmen pembelian |
| Customer Service | Tiket dan penanganan pelanggan |
| CRM & Sales | Lead, prospek, kampanye opt-in dan relationship |
| Document & Records | Bukti/dokumen berversi, retensi dan tautan arsip |
| HRM | Pegawai dan hubungan kerja, bukan IAM roles |

**Integrasi marketplace** (Shopee, TikTok Shop dan lainnya) adalah kapabilitas lintas-owner yang membutuhkan pemilik adapter Integration (tidak otomatis menambah ke 13 management). Analytics menggunakan fakta owner/proyeksi, bukan ledger kedua.

## Phase gate
| Urutan | Siapa | Definisi hasil |
|---|---|---|
| F1 Internal | Owner, admin, purchasing, WMS, Finance dan HR | Setup legal entity, SKU/varian, supplier master, PO/receipt, saldo stok, Finance, IAM scope. |
| F2 Vendor dan supplier | PIC manusia terverifikasi, party/kontrak masing-masing | Undangan, grant terbatas, PO/ASN/kontrak/bukti yang boleh diakses. |
| F3 Marketplace & analytics | Staf channel + workload connector | Official API/channel approval, webhook signed, inbox/outbox/reconciliation, metrik per SKU/channel. |
| F4 Customer & calon pembeli **terakhir** | Public, guest, logged-in buyer | Website, katalog LIVE, keranjang, quote, pembayaran sesuai policy, OMS order, tracking, tiket/retur. |

Manufacturing **PAUSED**; jangan menambah domain produksi atau menghapus implementasi lama. Cakupan retail khusus seperti expiry/lot produk impor ditangani PLM (persyaratan produk) dan WMS (lot/expiry stok) dalam batas kontrak masing-masing. Kebutuhan inspeksi/QMS di luar 13 management aktif dicatat sebagai dependency dan memerlukan keputusan owner, bukan alasan otomatis menambah management atau mengaktifkan Manufacturing.

## Keputusan yang belum tersedia
Jenis produk dan kebutuhan klien pilot, metode pembayaran/kurir, hak integrasi channel resmi, biaya/penagihan, UX brand, pilihan deployment dedicated vs shared, kapasitas/SLO dan review hukum. Status semua **OPEN**, bukan asumsi.
