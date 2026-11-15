# Roadmap dan Product Backlog

> **Integrione Retail ERP · 12 November 2026 · status: rancangan produk.** Sumber keputusan produk: [governance](../00-governance/decisions.md). Kebutuhan target bukan bukti implementasi.

## Keputusan urutan
- **F1 INTERNAL:** IAM browser pilot/permissions, Foundation legal entity/party, PLM SKU/variants, Procurement PO, WMS stock/receipt, Finance invoices/report; endpoint privat dengan E2E positif-negatif.
- **F2 VENDOR/SUPPLIER:** validasi representative dan grant scoped party/kontrak, portal procurement dan dokumen; tidak otomatis menambah auth dari status bisnis.
- **F3 MARKETPLACE + ANALYTICS:** official Shopee/TikTok Shop API, channel OAuth & token vault, webhook signed, durable inbox, idempotency OMS, outbox sync WMS, reconciler, SKU/channel metrics.
- **F4 BUYER/PROSPECT TERAKHIR:** website brand, kategori/listing LIVE, guest account/cart, Pricing quote, OMS order, Finance payment status, tracking, CRM consent dan Customer Service.

## Backlog prioritas (awal, perlu sinkronisasi ke issue aktif)
| ID | Domain | Deliverable dan bukti | Gate |
|---|---|---|---|
| B-01 | IAM/Gateway | Opaque cookie + CSRF cutover, 30m/8h, deny/fail closed; HTTP E2E | F1 |
| B-02 | PLM/Foundation | Legal entity/party dan SKU variants untuk stok/order | F1 |
| B-03 | Procurement/WMS/Finance | PO, receipt parsial, saldo/three-way matching | F1 |
| B-04 | IAM/Procurement/DRM | Vendor/supplier access dan dokumen scoped | F2 |
| B-05 | Integration/OMS/WMS | Webhook signed, dedupe, mapping external order, reconciliation | F3 |
| B-06 | Analytics/Finance/OMS | Metrics views/cart/orders/returns/net sales dengan watermark | F3 |
| B-07 | Commerce/Pricing/OMS | Storefront lengkap, quote & submission idempoten | F4 |
| B-08 | CRM/Customer Service | Lead, consent, ticket dan return flow | F4 |

## Aturan tanggal, biaya, penerimaan
Ini **roadmap urutan**, bukan tanggal pasti. Estimasi O/M/P (PERT), biaya, staffing, provider approvals, channel pertama, jam support dan SLA memerlukan data. Semua backlog hanya dianggap VERIFIED bila test sesuai scope, bukan saat schema atau mock selesai. Tidak ada promosi otomatis dari merge commit ke production.
