# Finance & Accounting — AP, AR, Customer PG, GL dan Pajak

> **Dokumen management retail · Diperbarui 14 November 2026.** Kepemilikan bisnis normatif mengikuti [agents/boundary.md](../../agents/boundary.md) dan [keputusan owner](../00-governance/decisions.md). Ini ringkasan arsitektur, bukan instruksi mengubah database/kode Finance sebelum gate fasenya dibuka.

## Ownership
Finance memiliki chart of accounts, fiscal periods, AP vendor/supplier invoices (`ap_payments`), AR customer invoices (`ar_invoices`, `ar_receipts`), **sub-domain Customer Payment Gateway (`payment_gateway` / `online_collections`)**, cash/bank receipt verification & clearing accounts, GL/journals, cost/valuation, taxation dan laporan keuangan. Procurement PO, WMS receipt dan OMS order adalah **fakta sumber** yang harus divalidasi untuk keputusan akuntansi.

## Internal (AP Manual Supplier & Vendor CMT + Rekonsiliasi Bank)
- **AP Supplier & Vendor (CMT) tetap 100% manual:** Pembayaran tagihan bahan mentah maupun jasa maklun Vendor CMT dilakukan manusia melalui internet banking di luar ERP (tutup buku/periode bulanan) dengan alur Maker–Checker–Approver–Verifier di `ap_payments` serta unggah bukti ke DRM. API *payout/disbursement* PG (`PayoutApi`) **tidak digunakan**.
- AP invoice tidak boleh dianggap matched hanya karena PO ada; butuh PO lines + receiving actual qty/quality/WMS source, supplier/vendor party identity Foundation, dan tax/vendor invoice. Saat sources unavailable, kembalikan pending/unavailable/blocked, bukan posting otomatis.
- Rekonsiliasi mutasi rekening koran bank operasional tetap diinput dan diverifikasi teraudit tanpa H2H bank.

## Customer Payment Gateway (Web Commerce — F3/F4 Blueprint)
- **Pemilik Eksekusi & Ledger:** `finance_accounting` memiliki modul Customer PG (tahap awal **Xendit** lewat `github.com/xendit/xendit-go/v7`, dibungkus abstraksi multi-PG internal agar siap menampung **Midtrans**). Sub-domain ini **dipisahkan tabel dan aturannya** dari `ap_payments` dan `ar_receipts` SoD 2-orang manual agar integrasi webhook otomatis tidak melemahkan kontrol kas manual.
- **Database Percobaan & Analitik Global:** Setiap inisiasi bayar, status (`PENDING`, `PAID`, `EXPIRED`, `FAILED`, `CANCELLED`, `REFUNDED`), kanal pembayaran (VA, QRIS, eWallet, Retail Outlet, Kartu), `failure_code`, serta biaya PG dicatat untuk kebutuhan analisis conversion rate dan audit.
- **Strategi Xendit Bertahap:** Tahap 1 memakai `InvoiceApi.CreateInvoice` (*Hosted Payment Page*); Tahap 2 siap memakai `PaymentRequestApi.CreatePaymentRequest` (`.IdempotencyKey(...)`) untuk tampilan *native* UI IntegriOne; refund memakai `RefundApi`; rekonsiliasi settlement memakai `TransactionApi` (`SettlementStatus` + `FeeResponse` untuk `XenditFee` & PPN fee).
- **Akuntansi Clearing & Webhook:** Webhook masuk (jalur pengecualian terbatas `/api/webhooks/xendit`) **wajib** divalidasi header `x-callback-token` secara *constant-time*, disimpan di durable inbox sebelum ACK, memindah saldo masuk ke **`cash_accounts` bertipe `clearing`** (`Xendit Clearing`), melunasi tagihan AR + memicu arsip invoice customer di DRM, serta mengirim event outbox terverifikasi ke OMS (`CONFIRMED`), WMS (konfirmasi alokasi / release hold bila `EXPIRED`), dan Commerce. Saat settlement cair dari Xendit ke bank perusahaan, Finance menjurnal pencairan bersih bank + beban MDR/PPN dari akun clearing.

## Penjualan multikanal (Marketplace)
External order/settlement yang masuk via webhook Shopee/TikTok bukan bukti dana tersedia atau transaksi bank berhasil. Finance menerima source facts via adapter/contract dan memutuskan settlement/refund, posting GL, biaya marketplace, dan pencocokan payout. Commerce/OMS hanya menyimpan status/keterangan owner.

## Keamanan
Posting, approval, reversal, refund manual, rights change dan delete berisiko memerlukan IAM permission/scope, SoD dan recheck sebelum commit. Correction harus auditable, tidak melalui update SQL management lain. Uang/kuantitas memakai numeric/decimal dengan rounding dan currency scale eksplisit. Secret API key dan webhook verification token PG disimpan di secret manager/vault terpisah per environment.

## Batas roadmap
Arsitektur Customer Payment Gateway (Xendit/Midtrans) di atas telah **DIPUTUSKAN (14 November 2026)** sebagai blueprint resmi fase **F3/F4** (terpisah dari transfer/AP bank internal yang tetap manual). Skema DBML (`finance_accounting.dbml`), proto (`receipt.proto`), dan kode Go Finance **belum** dimodifikasi saat ini karena prioritas aktif masih menyelesaikan IAM/Gateway dan fase internal terlebih dahulu.

[Finance README](../../backend/services/finance_accounting/README.md) · [boundary](../../agents/boundary.md#finance_accounting).
