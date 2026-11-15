# Keputusan Owner dan Prinsip Tetap

> **Dokumentasi Integrione · 12 November 2026.** Penetapan owner, desain, implementasi, verifikasi dan deployment adalah status berbeda. Kebijakan normatif mengikuti [AGENTS.md](../../AGENTS.md), [arsitektur keamanan](../../agents/arsitektur.md), dan [batas management](../../agents/boundary.md).

## Keputusan produk
**DECIDED:** Fokus retail ERP dan website perdagangan produk jadi untuk UMKM/brand. **13 management** dalam scope: IAM, Enterprise Foundation, Product Lifecycle Management, E-Commerce Commerce, Pricing & Revenue, Order Management, Inventory & Warehouse, Finance & Accounting, Procurement & Sourcing, Customer Service, CRM & Sales, Document & Records, Human Resource Management.

**Manufacturing Operations: PAUSED untuk pekerjaan baru** dan di luar daftar 13 management aktif. Keputusan menjaga kode/riwayat historis tidak berarti service Manufacturing tersedia di checkout terbaru; jejak historis tetap dapat diaudit lewat Git. Urutan pengembangan **F1 internal → F2 vendor dan supplier → F3 marketplace/webhook/analitik → F4 customer dan calon pembeli terakhir**. Fase berbeda boleh menjalankan desain atau PoC, tetapi runtime publik/customer belum boleh dibuka sebelum gate-nya selesai.

## Keputusan owner 13 November 2026 — hanya satu HTTP API

**DECIDED, implementasi lintas Gateway/IAM masih berlangsung:** satu-satunya
endpoint HTTP API ERP yang boleh dipublikasikan adalah **`/api/graphql`**.
Login, logout, bootstrap CSRF, pemeriksaan/rotasi sesi, perubahan password
dan semua operasi manajemen dilakukan sebagai Query/Mutation GraphQL melalui
path itu. Tidak ada path REST `/api/auth/*`, `/healthz` atau API bisnis
tambahan. Liveness memakai GraphQL `query { health }` pada path yang sama.

Caddy hanya meneruskan `/api/graphql` ke Gateway dan menolak `/api/*`
lain dengan 404. Halaman serta asset Next.js bukan API. gRPC antar-service
tetap internal, bukan public HTTP endpoint. Pengecualian terbatas hanya diizinkan
secara eksplisit oleh owner untuk endpoint **inbound webhook server-to-server Payment Gateway**
resmi (seperti `/api/webhooks/xendit`, dan kelak `/api/webhooks/midtrans`) saat fase
integrasi PG F3/F4 diaktifkan; integrasi marketplace Shopee/TikTok yang meminta jalur
callback sendiri tetap membutuhkan keputusan implementasi & kontrak partner terpisah.

Keputusan endpoint tidak melemahkan D01–D12: opaque session/cookie HttpOnly,
Origin, pre-auth CSRF, validasi CSRF atas mutasi, mTLS, per-scope authorization
dan fail-closed tetap wajib. Handler REST historis di Gateway belum seluruhnya
dimigrasikan; hapus hanya setelah pengganti GraphQL beserta tes E2E berjalan,
jangan menyamakan perubahan Caddy dengan cutover IAM selesai.

## Keputusan owner 13 November 2026 — IAM database development reset baseline v1

**DECIDED, belum IMPLEMENTED/VERIFIED:** ERP masih pra-produksi. Seluruh migrasi
database IAM harus dibangun ulang dari DBML target sebagai **satu-satunya versi
`000001`**, tidak boleh dibuat `000002` atau patch upgrade skema IAM lama.
File SQL baseline lama diganti dengan hasil ekspor DBML yang telah diperiksa
constraint PostgreSQL; Go migration runner/seed/test harness ikut dibersihkan
dari jalur kompatibilitas historis. Database development **hanya `iam`**
dapat dibuat ulang setelah verifikasi host, DSN, scope, dan penghentian koneksi;
database bisnis lain, PostgreSQL cluster/volume, serta database produksi tidak
boleh dihapus. Reset tidak dilakukan otomatis saat startup.

Kode IAM dan Gateway yang mengacu ke tabel/fitur tanpa dukungan DBML target
harus **dihapus** pada seluruh lapisan: repository, service, transport gRPC,
protobuf/generated, GraphQL SDL/resolver/model/generated, konfigurasi, test
dan dependencies. Pemangkasan dilakukan dengan analisis caller/consumer dan
regenerasi tool resmi, disertai full compile/test. Pertahankan fitur aktif
sesuai DBML: principal, RBAC–ABAC, password/TOTP/recovery, opaque session,
Redis security state, signed CSRF, access request, authorization outbox,
Google/Facebook customer OAuth. Ini bukan izin menghapus service bisnis domain
lain atau melemahkan D01–D12.

**Rencana pelaksanaan:** [indeks instruksi IAM/Gateway](../../agents/instructions/README.md).
Semua instruksi berstatus DRAFT sampai implementasi dan tes sungguhan dilakukan.

## Keputusan keamanan D01–D12
**Sumber normatif:** [bagian 21 agents/arsitektur.md](../../agents/arsitektur.md#21-keputusan-owner-yang-ditetapkan). Ringkasan yang tidak mengubah isi:
- Sesi browser target **opaque**, cookie `__Host-integrione_session`; IAM mengelola lifecycle dan Redis security state.
- **Idle 30 menit**, **absolut 8 jam**. Snapshot authorization freshness **maksimum 30 detik dari pemeriksaan sumber otoritatif**, bukan cookie/session TTL.
- Pencabutan dinyatakan **effective** hanya setelah state runtime terkonfirmasi; sampai itu **pending**. Sensitif write melakukan recheck sebelum commit; kegagalan dependency keamanan adalah **fail closed**.
- Realm dan legal entity adalah hal berbeda; RBAC role tidak menggantikan grant/scope/ABAC objek; Redis failover memerlukan recovery gate yang benar.

**JWT browser access 15 menit adalah detail legacy, bukan keputusan sesi browser terbaru.** Penggunaan bearer untuk API/workload tetap tergantung kontrak sendiri, tidak menjadi fallback ketika cookie opaque gagal.

## Keputusan arsitektur 14 November 2026 — Customer Payment Gateway, DRM & `/profile`

**DECIDED (Blueprint F3/F4 — kode & skema PG/DRM belum diaktifkan selama fokus tim masih di IAM & F1):**
1. **Pemilik Payment Gateway Customer:** Integrasi **Customer Payment Gateway Web Commerce** (tahap awal **Xendit** lewat `github.com/xendit/xendit-go/v7`, dengan abstraksi multi-PG siap **Midtrans**) ditempatkan di **`finance_accounting`** pada sub-domain baru (`payment_gateway` / `online_collections`), tetap dalam batas **13 management aktif** tanpa membuat microservice ke-14. Sub-domain ini dipisahkan tegas dari tabel AP manual (`ap_payments`) dan AR manual 2-orang SoD (`ar_receipts`).
2. **Pembayaran AP Supplier & Vendor (CMT) tetap Manual:** Seluruh pembayaran keluar ke Supplier bahan mentah dan Vendor CMT tetap diproses manual oleh Finance (tutup buku/periodik lewat bank di luar ERP + bukti DRM); fitur *payout/disbursement* PG (`PayoutApi`) tidak dipakai.
3. **Strategi Hybrid Xendit & Database Analitik Global:**
   - **Tahap 1:** `InvoiceApi.CreateInvoice` (*Hosted Payment Page* Xendit) untuk cakupan cepat VA, QRIS, eWallet, Retail Outlet, dan Kartu.
   - **Tahap 2:** `PaymentRequestApi.CreatePaymentRequest` (wajib `.IdempotencyKey(...)`) untuk tampilan VA/QRIS *native* di UI IntegriOne.
   - **Ledger & Analitik:** `finance_accounting` menyimpan seluruh histori percobaan & status pembayaran (`PENDING`, `PAID`, `EXPIRED`, `FAILED`, `CANCELLED`, `REFUNDED`, kanal, `failure_code`, fee/MDR) untuk analisis dan rekonsiliasi bank (`TransactionApi` + `FeeResponse` ke `cash_accounts` bertipe `clearing`), serta `RefundApi`. Validasi webhook Xendit wajib memeriksa `x-callback-token` secara *constant-time comparison* (SDK `v7` hanya menyediakan struct unmarshal).
   - **Alur Lintas Management:** `ecommerce_commerce` (guest cart HMAC → login `/login` → `commerceClaimGuestCart` → checkout & order submission) memanggil `finance_accounting` via Gateway/gRPC; `order_management` menjaga Sales Order tunggal (`AWAITING_PAYMENT` → `CONFIRMED`) dan `inventory_warehouse_management` memegang Stock Hold TTL sinkron durasi tagihan agar tidak *oversell*.
4. **Pengecualian Terbatas Jalur Webhook PG:** Seluruh UI/browser API tetap tunggal di `/api/graphql`; saat fase PG diaktifkan, Ingress/Caddy & Gateway diberi pengecualian terbatas hanya untuk endpoint *server-to-server inbound webhook* PG terverifikasi token/signature (seperti `/api/webhooks/xendit` & `/api/webhooks/midtrans`), tanpa CSRF/cookie sesi browser, dan persist durable inbox sebelum ACK.
5. **Keamanan DRM & Route `/profile`:** `document_records_management` (DRM) menyimpan dokumen lintas departemen termasuk slip gaji karyawan, invoice customer pasca-pembayaran, RFQ/PO/kontrak maklun CMT, dan arsip Finance dengan kewajiban **anti-IDOR & anti-*Privilege Escalation*** (RBAC + ABAC + scope kepemilikan `principal_id`/`party_id`/`legal_entity_id` *fail-closed*). Route **`/profile`** disediakan sebagai rute akun mandiri terisolasi untuk seluruh pengguna terautentikasi (Internal, Customer, Vendor CMT, dan Supplier).

## Keputusan arsitektur dan ownership
Satu fakta bisnis memiliki satu penulis: OMS untuk sales order; WMS untuk stok fisik; Finance untuk invoice/receipt/settlement PG/GL; Pricing untuk harga; PLM untuk item; Commerce untuk etalase, cart, quote dan submission. Transaksi lintas management memakai kontrak owner, outbox/inbox, key idempoten, rekonsiliasi, tanpa SQL/FK lintas database.

## Belum diputuskan
Model deployment banyak klien: tetap baseline **satu instalasi per perusahaan**; opsi shared multi-tenant belum disetujui. Pembayaran AP ke Supplier/Vendor tetap manual; implementasi detail rate limit/SLO, kontrak partner API Shopee/TikTok, pengaturan pajak transaksi PG, serta jadwal aktivasi kode F3/F4 (setelah IAM & F1–F2 selesai) membutuhkan validasi terpisah sebelum kode dijalankan di produksi.
