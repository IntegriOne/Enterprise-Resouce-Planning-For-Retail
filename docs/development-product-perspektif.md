# Perspektif Produk pada Fase Development (Product Perspective Guide)

**Status:** Pedoman Wajib Lintas Agent (Frontend, UI/UX, IAM Management, Finance, Commerce, dan Gateway)  
**Diperbarui:** 14 November 2026 *(revisi: arsitektur `/profile`, keamanan DRM anti-IDOR, dan blueprint Customer Payment Gateway)*  
**Rujukan terkait:** [AGENTS.md](../AGENTS.md), [agents/users.md](../agents/users.md), [agents/boundary.md](../agents/boundary.md), [docs/00-governance/decisions.md](00-governance/decisions.md)

---

## 1. Koreksi Perspektif Dasar

Selama fase pengembangan, seluruh agent **dilarang** memosisikan **IntegriOne** sebagai perusahaan vendor software SaaS *multitenant* yang memasarkan "Platform Enterprise ERP", dan **dilarang** memosisikan `/login` sebagai portal khusus staf/admin internal.

Perspektif resmi yang wajib dipakai seluruh agent:

1. **Codebase Single-Tenant per Perusahaan:** Kita membangun codebase aplikasi terpadu (**Web Commerce + Sistem Operasional Retail ERP**) dengan model **satu instalasi per perusahaan (*single-tenant*)** untuk dipakai langsung oleh brand/perusahaan ritel klien, bukan website pemasaran SaaS untuk berjualan langganan ERP.
2. **Posisi Brand "IntegriOne" Saat Development:** Selama fase development, aplikasi diposisikan seolah-olah perusahaan ritel **IntegriOne** sendiri yang mengoperasikan website toko dan sistem operasionalnya. Seluruh tampilan publik dan halaman masuk merepresentasikan brand **IntegriOne** kepada pelanggan maupun internal.
3. **Satu Pintu Masuk Umum (`/login`):** Sesuai arsitektur satu tabel `users` (`customer`, `employee`, `vendor`, `supplier`) di IAM ([agents/users.md](../agents/users.md)), hanya ada **satu halaman `/login`** untuk tiga kelompok pengguna:
   - **Calon Pembeli / Customer Web Commerce**
   - **Internal (Employee / Staf / Admin)**
   - **Eksternal Mitra (Vendor CMT / Supplier Bahan Mentah)**

---

## 2. Aturan Wajib Frontend & UI/UX untuk Halaman Umum (`/login`)

1. **Gunakan Redaksi Umum yang Inklusif:**
   - Saat membuka `/login`, sistem belum mengetahui apakah pengunjung adalah calon pembeli, staf internal, atau mitra vendor/supplier.
   - **Dilarang** menaruh istilah teknis backoffice/eksklusif internal di `/login`, seperti *"Enterprise Resource Planning"*, *"Portal IAM"*, *"Hubungi Administrator IAM"*, atau diagram modul internal (*Gudang, Keuangan, Pengadaan, Omnichannel*).
   - **Wajib** memakai bahasa umum yang ramah pelanggan sekaligus profesional, misalnya: *"Masuk ke akun IntegriOne"*, *"Selamat datang kembali"*, *"Email atau username"*, *"Kata sandi"*, dan *"Lupa kata sandi?"*.
2. **Dilarang Memisahkan Tab/Role Sebelum Login:**
   - Tidak boleh ada tab/pilihan tipe pengguna (*Internal / Vendor / Supplier / Customer*) sebelum login.
   - Semua pengguna masuk lewat form yang sama; pemisahan ruang kerja dilakukan **setelah autentikasi berhasil (`Post-Authentication`)** sesuai hak akses dari server.
3. **Akomodasi Seluruh Aktor pada `/login`:**
   - Form utama: **Email atau username** + **Kata sandi** + **Lupa kata sandi**.
   - Dukungan alur login sosial (Google/Facebook) dan pendaftaran akun umum bagi calon pembeli Web Commerce, sambil tetap melayani staf internal dan vendor/supplier melalui form kredensial yang sama.

### Ringkasan Perbandingan

| Aspek | Perspektif Salah (Dilarang) | Perspektif Benar (Wajib) |
|---|---|---|
| **Identitas Aplikasi** | Vendor SaaS multitenant yang menjual software ERP | Website & sistem operasional milik perusahaan ritel (saat dev diwakili brand **IntegriOne**) |
| **Target `/login`** | Khusus admin atau staf internal ERP | Umum untuk **Calon Pembeli Web Commerce**, **Employee Internal**, **Vendor (CMT)**, dan **Supplier** |
| **Copywriting** | *"Portal IAM"*, *"ERP Platform"*, *"Hubungi Admin IAM"* | Bahasa umum yang ramah: *"Masuk ke akun Anda"*, *"Lupa kata sandi?"*, *"Daftar akun"* |
| **Penentuan Akses** | Memilih peran di layar `/login` | Diarahkan otomatis **setelah** autentikasi berhasil berdasarkan otorisasi IAM |

---

## 3. Arsitektur Dashboard & Sub-Menu Berdasarkan Kondisi Kode Backend Saat Ini

> **Fokus Tahap Saat Ini:**  
> Seluruh pengembangan tim saat ini memprioritaskan **IAM Management** dan **Dashboard IAM (`/identity`)**.

**Prinsip Utama:** **1 Backend Management ≠ 1 Route Dashboard Frontend.**  
Service yang berfungsi sebagai fondasi struktur perusahaan, penyimpanan bukti dokumen, atau sub-proses yang melekat pada operasional lain **tidak diberi route dashboard mandiri**, melainkan masuk sebagai **sub-menu** di dalam dashboard operasional terkait (sambil tetap memanggil GraphQL service pemilik data yang sah di `/api/graphql` sesuai [agents/boundary.md](../agents/boundary.md)).

Berdasarkan pemeriksaan langsung pada `backend/services/`, `backend/gateway/graph/resolvers/resolver.go`, `backend/gateway/graph/schema/`, dan `backend/gateway/configs/gateway.yaml`:
- Terdapat **13 service management** di backend.
- Baru **8 service** yang memiliki implementasi domain, schema GraphQL, serta adapter di `Resolver` Gateway (`iam_management`, `enterprise_foundation`, `human_resource_management`, `finance_accounting`, `procurement_sourcing`, `inventory_warehouse_management`, `product_lifecycle_management`, dan `ecommerce_commerce`).
- **5 service lainnya** berstatus tertunda (`document_records_management` baru punya DBML/schema dengan `service.go` kosong tanpa adapter Gateway; `pricing_revenue_management`, `order_management`, `crm_sales_management`, dan `customer_service_management` masih template TODO scaffold dengan `enabled: false`).

---

### A. Peta Route Dashboard Internal yang Siap Backend (7 Route Utama)

| No. | Route Dashboard (`Location`) | Pemilik Utama | Sub-Menu / Management yang Diserap | Backend Service yang Dipanggil (`/api/graphql`) | Status & Prioritas |
|---:|---|---|---|---|---|
| 1 | **`/identity`** | **IAM Management** | Principals/User, Roles, Permissions, Groups, Access Requests, Access Reviews, Delegations, API Keys, Security Policies, MFA & Active Sessions, Security Audit Logs. | `iam_management` (`enabled: true`) | **PRIORITAS UTAMA SAAT INI** |
| 2 | **`/human-resource`** | **Human Resource Management (HRM)** | Worker, Posisi, Kontrak, Rekrutmen, Penggajian (Payroll/BPJS), Cuti, Kinerja, Hubungan Industrial, **+ Sub-Menu Struktur Perusahaan (`enterprise_foundation`)**: Badan Hukum (`legal_entity`), Divisi/Departemen (`organization_unit`), serta Lokasi & Alamat Perusahaan (`location`, `address`). *(Kelak menyerap lampiran berkas pekerja dari `document_records_management`)*. | `human_resource_management` + `enterprise_foundation` (`enabled: true`) | Fase F1 (Setelah IAM) |
| 3 | **`/procurement`** | **Procurement & Sourcing** | Kategori Beli, Permintaan Beli (PR), RFx/Sourcing, Bidding, Kualifikasi Rekanan, Kontrak Jasa/Beli, Purchase Order (PO), Pemantauan ASN, **+ Sub-Menu Master Data Vendor (CMT) & Supplier (`enterprise_foundation`)**: Input & pemeliharaan data rekanan (`party`, `party_role_assignment`, NPWP/pajak rekanan, PIC eksternal). *(Kelak menyerap arsip kontrak/PO dari `document_records_management`)*. | `procurement_sourcing` + `enterprise_foundation` (`enabled: true`) | Fase F1 (Setelah IAM) |
| 4 | **`/warehouse`** | **Inventory & Warehouse (WMS)** | Pengaturan Fasilitas & Bin (`facility`), Inbound/Receiving (dari PO Supplier maupun barang jadi hasil Vendor CMT), Kontrol Saldo/Lot/Serial, Transfer/Alokasi Material, Stock Opname (`cycle_count`), Outbound (`picking`, `packing`, `shipping_execution`), serta Jadwal Kerja Gudang (`workforce`). *(Kelak menyerap dokumen surat jalan/terima dari `document_records_management`)*. | `inventory_warehouse_management` (`enabled: true`) | Fase F1 |
| 5 | **`/finance`** | **Finance (Operasional Keuangan)** | Accounts Payable (AP & 3-way matching PO/Receipt, termasuk tagihan pembelian bahan Supplier maupun jasa maklun Vendor CMT), Accounts Receivable (AR & verifikasi pembayaran), Cash & Bank, Kasbon, Budgeting, Revenue Recognition. *(Kelak menyerap bukti tagihan/bayar dari `document_records_management`)*. | `finance_accounting` (`enabled: true`) | Fase F1 *(satu backend dengan Accounting)* |
| 6 | **`/accounting`** | **Accounting (Pembukuan & Pajak)** | Bagan Akun (CoA), General Ledger & Jurnal, Fixed Assets & Depresiasi, Perpajakan (`taxation`), Konsolidasi, Laporan Keuangan (`financial_reporting`). *(Kelak menyerap arsip pembukuan/pajak dari `document_records_management`)*. | `finance_accounting` (`enabled: true`) | Fase F1 *(satu backend dengan Finance)* |
| 7 | **`/product`** | **Product Lifecycle (PLM)** | Master Item & SKU (bahan mentah & produk jadi), Varian, Struktur Produk / BOM (`bom_product_structure`), Spesifikasi & Risiko Desain, Engineering Change/Waiver, Rilis Produk & Dokumen Regulasi/TKDN. *(Kelak menyerap dokumen desain/TKDN dari `document_records_management` dan pengaturan harga dasar dari `pricing_revenue_management`)*. | `product_lifecycle_management` (`enabled: true`) | Fase F1 |

---

### B. Portal Eksternal Mitra (`/supplier` vs `/vendor`) & Web Commerce

Model rantai pasok bisnis perusahaan membedakan peran **Supplier** dan **Vendor (CMT)** secara jelas:

1. **Supplier (`/supplier`) — Pemasok Bahan Mentah:**
   - Pihak eksternal tempat perusahaan membeli **bahan mentah / material** (seperti kain, benang, aksesori, atau kemasan).
   - Alur di portal `/supplier`: melihat kualifikasi pemasok, penawaran/bidding, PO pembelian bahan mentah, dan mengirimkan konfirmasi/ASN pengiriman material ke gudang perusahaan.
2. **Vendor (`/vendor`) — Perusahaan CMT (*Cut, Make, Trim* / Maklun Pengolahan):**
   - Pihak eksternal penerima **bahan mentah yang sudah dibeli perusahaan dari Supplier** untuk diolah/dijahit/diproduksi menjadi **produk jadi**.
   - Alur di portal `/vendor`: menerima perintah kerja / kontrak maklun CMT, memantau penerimaan alokasi bahan mentah dari gudang perusahaan, melaporkan progres pengerjaan, serta mencatat pengiriman hasil produk jadi kembali ke gudang perusahaan.

> **Catatan Kesiapan Backend untuk `/supplier` & `/vendor`:**  
> Di `enterprise_foundation.dbml`, saat ini baru tersedia kode peran umum `vendor` (belum memisahkan secara eksplisit kode peran `supplier_material` vs `vendor_cmt`), dan di `procurement_sourcing` + `inventory_warehouse_management` alur subkontrak/penyerahan bahan mentah ke CMT serta penerimaan barang jadinya masih perlu dirinci kontraknya pada fase F2 (sambil tetap mematuhi aturan bahwa `Manufacturing Operations` berstatus `PAUSED`). Hal ini **akan didiskusikan lebih lanjut saat memasuki tahap F2**.

| No. | Route (`Location`) | Peran / Aktor | Cakupan Fungsional | Backend Service Pendukung | Fase |
|---:|---|---|---|---|---|
| 8 | **`/supplier`** | **Supplier (Pemasok Bahan Mentah)** | Kualifikasi pemasok material, bidding/RFx bahan baku, melihat PO bahan mentah, dan mengirim ASN ke gudang. | `procurement_sourcing` + `enterprise_foundation` | **Fase F2** |
| 9 | **`/vendor`** | **Vendor (Perusahaan CMT / Pengolah Produk Jadi)** | Kontrak jasa pengolahan/maklun CMT, serah-terima bahan mentah yang dipasok perusahaan untuk diolah menjadi produk jadi, dan pengiriman hasil jadi ke gudang. | `procurement_sourcing` + `enterprise_foundation` *(+ keterkaitan WMS/PLM saat didetailkan di F2)* | **Fase F2** |
| 10 | **`/`**, **`/account`**, dan **`/commerce`** *(operasional toko & pesanan)* | **Customer Web Commerce** (`/` & `/account`) dan **Tim Operasional Toko Online** (`/commerce`) | Pelanggan (`/` & `/account`): etalase produk jadi, *guest cart* yang otomatis tergabung ke akun saat login, checkout terintegrasi Payment Gateway, pelacakan pesanan, serta unduh invoice setelah pembayaran lunas. Operasional internal (`/commerce`): kelola channel & listing toko online, promo/harga, serta pemrosesan pesanan Web & Marketplace. | `ecommerce_commerce` + `finance_accounting` *(PG customer)* + `order_management` + `document_records_management` *(adapter `ecommerce_commerce` ada, `enabled: false` di `gateway.yaml` default)* | **Fase F3–F4** |
| 11 | **`/profile`** | **Seluruh Pengguna Terautentikasi** (Internal Employee, Customer Web Commerce, Vendor CMT, dan Supplier) | Ruang manajemen akun mandiri universal: ubah profil akun dasar, kata sandi, aktivasi/rotasi MFA/TOTP, daftar sesi aktif (`iam_management`), serta ringkasan identitas kontekstual sesuai peran (tanpa mencampur/membocorkan data internal pegawai ke akun eksternal/customer). | `iam_management` *(inti akun/keamanan)* + bacaan terotorisasi sesuai peran | **Fase F1–F4** *(mengikuti domain aktif)* |

---

### C. Konsepan Awal untuk Management Tanpa Route Mandiri, DRM, & Modul yang Ditunda

> **Catatan Penting:**  
> Seluruh poin untuk modul yang berstatus **ditunda** di bawah ini merupakan **konsepan awal perencanaan frontend** untuk menjaga kesatuan arsitektur, dan **akan didiskusikan lebih lanjut saat sudah saatnya** memasuki fase pengembangan masing-masing. Dilarang membuat implementasi UI atau API palsu sebelum backend-nya selesai.

| No. | Management Backend | Konsepan Awal Routing & Penempatan UI | Kondisi Backend Saat Ini | Catatan Tindak Lanjut |
|---:|---|---|---|---|
| 1 | **`enterprise_foundation`** | **TIDAK ADA route `/enterprise` atau `/organization`.** Struktur perusahaan (`legal_entity`, `organization_unit`, `location`) masuk sub-menu **`/human-resource`**. Master data rekanan (`party`, vendor CMT, supplier bahan mentah) masuk sub-menu **`/procurement`**. | Aktif di backend & Gateway. | Langsung diterapkan saat membangun `/human-resource` dan `/procurement`. |
| 2 | **`document_records_management` (DRM)** | **TIDAK ADA route dashboard `/documents` sendiri.** Menjadi **sub-layanan / sub-menu di berbagai management** untuk menyimpan berkas bernilai audit & retensi: **slip gaji karyawan** & berkas kontrak di `/human-resource` (serta `/profile` internal pegawai); **invoice customer** setelah pembayaran produk terkonfirmasi di `/account` & `/finance`; **RFQ**, penawaran, kontrak maklun CMT, PO, dan dokumen procurement di `/procurement`, `/supplier`, & `/vendor`; serta faktur pajak, bukti transfer AP/AR, spesifikasi desain/TKDN, dan surat jalan di `/accounting`, `/product`, dan `/warehouse`. | Punya DBML & `.graphqls` (13 KB), namun `service.go` masih kosong dan belum ada adapter Gateway. | **Ditunda — akan didiskusikan lebih lanjut saat sudah saatnya.** Wajib menegakkan **anti-IDOR & RBAC–ABAC ketat** sebelum dibuka. |
| 3 | **`pricing_revenue_management`** | **Konsepan awal: tidak butuh route `/pricing` sendiri.** Direncanakan masuk sebagai sub-menu pengaturan harga, diskon, dan promo di **`/product`** dan/atau **`/commerce`**. | Template TODO `.graphqls` (139 B), tanpa DBML, `enabled: false`. | **Ditunda — akan didiskusikan lebih lanjut saat sudah saatnya** sebelum F3/F4. |
| 4 | **`order_management`** | **Konsepan awal: tidak butuh route `/oms` terpisah.** Direncanakan masuk ke ruang kerja operasional toko & pesanan omnichannel (**`/commerce`** atau `/sales`) untuk menangani order masuk dari Web Commerce maupun Shopee/TikTok Shop, approval order, pembatalan, dan retur. | Template TODO `.graphqls` (129 B), tanpa DBML, `enabled: false`. | **Ditunda — akan didiskusikan lebih lanjut saat sudah saatnya** (Fase F3/F4). |
| 5 | **`crm_sales_management`** & **`customer_service_management`** | **Konsepan awal: tidak dipisah menjadi 2 dashboard.** Direncanakan digabung dalam **1 ruang kerja Layanan & Hubungan Pelanggan** (mis. `/customer-care`) atau menjadi sub-menu di bawah operasional **`/commerce`**, agar CS dapat menangani tiket/komplain sambil melihat profil & riwayat pelanggan dalam satu tempat. | Keduanya masih template TODO `.graphqls` (133–140 B), tanpa DBML, `enabled: false`. | **Ditunda — akan didiskusikan lebih lanjut saat sudah saatnya** (Fase F3/F4). |

*(Catatan: Manufacturing Operations berstatus `PAUSED`, tidak termasuk 13 management aktif, dan tidak memiliki direktori di `backend/services/`.)*

#### Aturan Keamanan Mutlak DRM (*Zero IDOR & Zero Privilege Escalation*)
Karena DRM menampung dokumen sangat sensitif lintas departemen (slip gaji karyawan internal, invoice customer, RFQ/penawaran harga vendor & supplier, faktur pajak, hingga kontrak perusahaan):
1. **Wajib RBAC + ABAC + Verifikasi Kepemilikan Objek di Server:** Setiap permintaan metadata, pratinjau (*preview*), atau tautan unduh (`document_id` / `version_id`) wajib diverifikasi oleh DRM bersama context IAM & pemilik bisnis (`owner_management`, `legal_entity_id`, `party_id`, `principal_id`, dan tingkat klasifikasi dokumen) secara ***fail-closed***.
2. **Anti-IDOR (*Insecure Direct Object Reference*):** Mengubah atau menebak UUID `document_id` / `version_id` di URL atau GraphQL **tidak boleh pernah** mengembalikan dokumen pihak lain (misalnya customer A mengunduh invoice customer B, atau supplier melihat RFQ/penawaran supplier lain).
3. **Anti-*Privilege Escalation*:** Akun customer, vendor CMT, atau supplier **tidak boleh pernah** mengakses dokumen internal perusahaan (seperti slip gaji karyawan HRM, bukti bayar AP internal, atau kontrak pegawai) dalam kondisi apa pun.

---

### D. Arsitektur Customer Payment Gateway (Web Commerce) vs Pembayaran Manual AP

Untuk memenuhi kebutuhan checkout **Web Commerce** secara otomatis sekaligus mempertahankan kontrol keuangan internal atas **Vendor (CMT) & Supplier**, batas arsitektur pembayaran ditetapkan sebagai berikut *(ketentuan detail direkam di [decisions.md](00-governance/decisions.md) & [finance.md](03-managements/finance.md))*:

1. **Pemisahan Tegas AP Manual vs Customer Online Payment:**
   - **Supplier & Vendor CMT (Accounts Payable / Pembayaran Keluar):** Tetap **100% manual oleh Finance** (pembayaran periodik / tutup buku bulanan melalui internet banking luar ERP, alur Maker–Checker–Approver–Verifier di `ap_payments`, + unggah bukti ke DRM). API *payout/disbursement* Payment Gateway **tidak digunakan**.
   - **Customer Web Commerce (Accounts Receivable / Checkout Otomatis):** Diproses otomatis melalui **Payment Gateway pihak ketiga** — tahap awal menggunakan **Xendit** (`github.com/xendit/xendit-go/v7`), dibungkus abstraksi multi-PG internal di backend agar siap menampung **Midtrans** di tahap berikutnya.
2. **Pemilik Backend Resmi (Tanpa Menambah Service ke-14):**
   - **`finance_accounting`** adalah **pemilik penuh modul Customer Payment Gateway** (sub-domain baru `payment_gateway` / `online_collections`, terpisah dari tabel manual `ap_payments` & tabel `ar_receipts` SoD 2-orang). Finance memegang secret API key PG, mengeksekusi permintaan pembayaran, menyimpan **database global riwayat percobaan & status pembayaran** (`PENDING`, `PAID`, `EXPIRED`, `FAILED`, `CANCELLED`, `REFUNDED`, kanal VA/QRIS/eWallet/Kartu, `failure_code`, dan fee/MDR) untuk analisis konversi & audit, memproses webhook masuk secara durable/idempoten, menjurnal ke **Akun Clearing PG (`cash_accounts.account_kind = 'clearing'`) + Beban MDR/PPN Fee**, dan melakukan rekonsiliasi settlement cair ke rekening bank perusahaan.
   - **`ecommerce_commerce`** berperan sebagai **UI Storefront & Inisiator Checkout**: menampung *guest cart* (cookie HMAC), menggabungkan cart ke akun saat login di `/login` (`commerceClaimGuestCart`), mengunci harga (`commerce_checkout_quotes`), membuat `commerce_order_submissions`, meminta sesi pembayaran ke Finance (via Gateway/gRPC), lalu merender link/instruksi bayar serta proyeksi status pesanan.
   - **`order_management` (OMS) & `inventory_warehouse_management` (WMS)**: OMS memegang **Sales Order** tunggal (`AWAITING_PAYMENT` → `CONFIRMED`) dan WMS melakukan **Stock Hold sementara (TTL sinkron dengan masa berlaku tagihan Xendit)** agar tidak terjadi *oversell* selama customer menunggu bayar.
   - **`document_records_management` (DRM)**: menerbitkan/menyimpan arsip invoice customer setelah pembayaran terkonfirmasi `PAID`, hanya dapat diunduh oleh customer pemilik transaksi (anti-IDOR).
3. **Strategi Integrasi SDK Xendit Terbaik (Bertahap):**
   - **Tahap 1 (Go-Live Cepat):** Menggunakan **`InvoiceApi.CreateInvoice`** (*Hosted Payment Page* Xendit) yang langsung menyediakan kanal VA, QRIS, eWallet, Retail Outlet, dan Kartu/3DS.
   - **Tahap 2 (Native Embedded UI):** Menggunakan **`PaymentRequestApi.CreatePaymentRequest`** (wajib memakai `.IdempotencyKey(...)`) bila checkout IntegriOne ingin menampilkan nomor VA/QRIS langsung di UI tanpa redirect.
   - **Refund, Rekonsiliasi & Verifikasi Webhook:** Menggunakan `RefundApi` dan `TransactionApi` (`SettlementStatus` + `FeeResponse` untuk biaya MDR & PPN). Karena `xendit-go/v7` **tidak** menyediakan helper verifikasi webhook, Gateway & Finance wajib memvalidasi header `x-callback-token` secara *constant-time comparison* melalui jalur pengecualian webhook terotorisasi (seperti `/api/webhooks/xendit`) ketika fase PG diaktifkan, sementara seluruh lalu lintas UI/aplikasi browser tetap tunggal di `/api/graphql`.

---

## 4. Aturan Kontrak Redirect (`Location`) & Pengamanan Rute

1. **Penentuan Tujuan Otoritatif dari Server:**
   - Gateway dan IAM menentukan target `Location` / `redirectLocation` setelah autentikasi dan validasi grant selesai melalui kontrak `/api/graphql`, bukan berdasarkan pilihan atau input client.
   - Jika pengguna memiliki lebih dari satu domain akses yang mencakup IAM, prioritas tahap pengembangan saat ini mengarahkan ke `/identity` (atau pemilih ruang kerja terotorisasi bila diaktifkan kemudian).
   - Validasi tujuan redirect wajib menggunakan *allowlist* rute internal aplikasi yang sudah eksis (termasuk `/profile` dan rute operasional yang sudah aktif) untuk mencegah *open redirect* dan 404.
2. **Proteksi Server-Side pada Dashboard (`/identity`, `/profile`, dan rute lainnya):**
   - Setiap rute dashboard maupun `/profile` wajib memvalidasi sesi `__Host-integrione_session` dan otorisasi RBAC–ABAC di server/Gateway secara *fail-closed* sebelum merender data atau memproses mutasi.

