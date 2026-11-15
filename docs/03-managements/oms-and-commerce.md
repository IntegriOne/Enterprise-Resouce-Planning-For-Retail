# Order Management dan E-Commerce Commerce

> **Dokumen management retail · Diperbarui 14 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md) dan [keputusan owner](../00-governance/decisions.md). Ini ringkasan, bukan instruksi melakukan perubahan lintas management atau seed IAM otomatis.

## Satu penulis untuk sales order & payment
- **OMS:** menerima/menolak sales order (`AWAITING_PAYMENT` → `CONFIRMED`), status order, cancel/return approval dan referensi fulfillment; order sales asli tetap di OMS.
- **Commerce:** channel, host/seller mapping, listing LIVE, buyer account/channel, *guest cart* (cookie HMAC) → klaim cart saat login (`commerceClaimGuestCart`), checkout, quote sealed, submission pending dan proyeksi hasil OMS/Finance; tidak memiliki order ledger kedua dan **tidak** menyimpan API key atau akuntansi Payment Gateway secara langsung.
- **Finance (`payment_gateway` / `online_collections`):** pemilik eksekusi Customer Payment Gateway (tahap awal Xendit, abstraksi siap Midtrans), pencatatan seluruh percobaan/status bayar (`PENDING`/`PAID`/`EXPIRED`/`FAILED`/`REFUNDED`), pemroses webhook terverifikasi, serta akuntansi clearing/MDR/refund.

## Flow F4 (customer terakhir)
Public website→list/catalog publish→guest cart HMAC→login di `/login` & `commerceClaimGuestCart`→qty/addresses/shipping→Pricing + Finance tax estimate→sealed quote/version/digest→buyer accepts→Commerce writes submission+outbox atomic→OMS accepts once by stable key (`AWAITING_PAYMENT`) + WMS Stock Hold TTL (sinkron durasi invoice Xendit)→Finance buat sesi bayar PG (Tahap 1 `InvoiceApi` / Tahap 2 `PaymentRequestApi` + `IdempotencyKey`) & tampilkan tautan/instruksi di Commerce→customer bayar→Webhook PG masuk ke Finance (validasi `x-callback-token` + durable inbox + `cash_accounts clearing`)→Outbox Finance ke OMS (`CONFIRMED`), WMS (alokasi kirim atau lepas hold bila `EXPIRED`), Commerce (proyeksi status lunas), dan DRM (arsip invoice customer sesuai `customer_party_id`/`principal_id`).

Jika buyer double click atau timeout setelah submit, lookup idempotency key stable pada OMS/Finance; jangan membuat dua order atau dua tagihan aktif yang bertabrakan. Jika harga/cart berubah, quote baru harus disetujui. `PENDING` tidak sama dengan ACCEPTED atau PAID. `commerce_payment_evidence_submissions` (jika dipakai pada kanal bukti manual) bukan bukti pelunasan otomatis AR Finance.

## Tahap A status checkpoint
Menurut audit 12 November 16:55, Commerce memiliki **empat operasi kanal GraphQL aktif di SDL terpisah**, dan E2E terbatas storefront PUBLIC lewat HTTP→gRPC mTLS→PostgreSQL telah dicatat; **Fase B–H dan integrasi IAM staf dengan data owner belum VERIFIED**. `gateway/graph/schema/ecommerce_commerce_channel.graphqls` aktif tidak berarti seluruh SDL Commerce aktif. Lihat [audit](../../agents/audit/ecommerce_commerce/3-2026-11-12-165500-e2e-public-dan-gerbang-iam.md).

## Marketplace
External order status hanya masukan idempoten ke OMS; provider tidak dapat memutuskan stock WMS/receipt Finance dari HTTP ACK. Satu shop account dan channel mapping harus divalidasi sebelum import.

[Commerce service](../../backend/services/ecommerce_commerce/README.md) · [OMS scaffold](../../backend/services/order_management/README.md) · [API](../04-engineering/api-contracts.md).
