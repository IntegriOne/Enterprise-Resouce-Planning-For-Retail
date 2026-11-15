# Product Analytics — Definisi Metrik dan Provenance

> **Baseline 12 November 2026.** Kebijakan keamanan/izin harus cocok dengan [agents/arsitektur.md](../../agents/arsitektur.md); ownership ditentukan oleh [boundary management](../../agents/boundary.md). Status runtime perlu dibuktikan dari kode/tes terbaru.

## Tujuan
Membandingkan penjualan produk per SKU/varian/channel, bukan membuat transaksi baru. Analytics membaca proyeksi owner (OMS, WMS, Finance, Commerce) dan data provider bila **secara resmi tersedia**.

| Metrik | Source dan definisi yang harus ditampilkan |
|---|---|
| Product views/impressions | Event website consented atau statistik provider resmi; definisi unique/time window diperjelas |
| Add to cart | Commerce cart event terdedupe; bukan order |
| Checkout started | Checkout ID/version; bukan order accepted |
| Orders | OMS accepted order unique per source/channel external ID |
| Gross merchandise value (GMV) | Total order accepted pada basis tax/shipping/currency yang dinyatakan |
| Net sales | Sales setelah discount/returns/refunds menurut kebijakan dan timing, **tidak disamakan settlement atau revenue recognition** |
| Refund/return rate | Finance refund dan OMS return/cancel terpisah, denominator cohort didefinisikan |
| Gross margin | Net sales minus COGS dari Finance/WMS; NULL jika COGS belum tersedia |
| Stock availability | WMS snapshot dengan as_of/freshness, bukan angka riwayat sales |
| Sync lag | Provider occurred→received→processed, watermark per shop |
| Payment attempts & conversion | `finance_accounting` (`payment_gateway`): jumlah inisiasi bayar, `PAID`, `PENDING`, `EXPIRED`, `FAILED`, `CANCELLED` per kanal (VA/QRIS/eWallet/Kartu) dan provider |
| Payment failure & expiry breakdown | `finance_accounting`: klasifikasi `failure_code` Xendit/PG serta rasio checkout kedaluwarsa untuk optimasi funnel |
| PG fee/MDR & settlement lag | `finance_accounting` (`TransactionApi` / `FeeResponse` & `clearing` recon): biaya MDR + PPN fee serta selisih waktu transaksi→cair ke bank |

## Model fakta
`source_management, provider, shop/account, item_id, variant_id, mapping_version, event_id, aggregate_revision, occurred_at_utc, received_at_utc, currency, amount_scale, digest, consent_purpose`. Jangan mengandalkan SKU string sebagai satu-satunya identitas lintas platform. Event correction dan delete/erasure dikelola sesuai ownership serta retensi.

## Aturan agregasi
Dedupe berdasarkan key asal, event yang datang terlambat dicatat dan mengubah agregat dengan correction sesuai period. Snapshot irreversible vs recompute harus diputuskan; bila denominator/view data tidak ada tampilkan **N/A**, bukan 0. Dashboard menampilkan freshness, source, timezone, currency dan audit metric version. Jangan membagikan PII pelanggan ke external analytics tanpa dasar hukum dan izin platform.

## Minimum dashboard
Penjualan per channel, SKU terlaris, stok habis, return/refund ratio, funnel view→cart→accepted order, margin jika COGS valid, webhook health dan reconciliation drift. Chart/analytics baru boleh disebut VERIFIED setelah dibandingkan dengan database owner dan event fixtures.
