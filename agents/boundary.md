# Batas tanggung jawab antar-management ERP Retail

> **Cakupan aktif diperiksa 13 November 2026.** Keputusan owner adalah **13 management Retail ERP** pada [peta management](../docs/03-managements/README.md) dan [governance](../docs/00-governance/decisions.md). Urutan pengembangan: **F1 internal → F2 vendor/supplier → F3 integrasi marketplace/webhook/analytics → F4 customer/calon pembeli terakhir**. Tidak ada izin otomatis memperluas management atau mengaktifkan API publik selain `/api/graphql`.

**Satu service pemilik mengubah fakta bisnisnya sendiri.** Service lain hanya memakai kontrak owner, ID logis, status dan versi yang sah; IAM tetap pemilik grant dan otorisasi. **Manufacturing Operations PAUSED** dan **bukan** anggota 13 management aktif. Tree `main` yang diperiksa tidak memiliki `backend/services/manufacturing_operations/`; pembahasan 32 management pada [audit historis](../docs/09-history/migration-map.md) tidak boleh dipakai sebagai bukti source saat ini.

**Inventaris berdasarkan tree source (bukan audit runtime):** `backend/services/` berisi 13 direktori management. Sembilan mempunyai DBML dalam `db/schema/`; Pricing, OMS, CRM dan Customer Service baru memiliki scaffold/kontrak tanpa DBML pada snapshot ini. Keberadaan DBML, proto, migration, bootstrap, atau README **bukan bukti** implementasi bisnis/E2E/deployment siap. Status aktual diperiksa menurut [bukti implementasi](../docs/00-governance/status-and-evidence.md).

## Aturan untuk developer dan reviewer

1. **Tentukan siapa yang boleh mengoreksi fakta.** Ajukan koreksi kuantitas yang dipesan ke Procurement, kuantitas yang diterima ke Inventory, dan nilai utang ke Finance. Modul yang menampilkan ketiganya harus meneruskan perubahan ke masing-masing pemilik.
2. **Batasi hak tulis sampai ke database.** Tempatkan model, migrasi, aturan bisnis, dan transaksi pada service pemilik. Dilarang memakai DSN, SQL/join, foreign key, trigger, migrasi, atau package Go `internal` milik management lain. FK internal tetap berlaku. Gateway memetakan permintaan dan respons. Package `shared/` hanya untuk kebutuhan teknis. [Arsitektur antar-service](../docs/02-architecture/service-boundaries.md)
3. **Gunakan kontrak pemilik.** Simpan ID logis beserta jenis/sumbernya bila perlu. Untuk keputusan yang bergantung pada data mutakhir, minta status, versi, dan scope badan hukum/unit yang diperlukan lewat API pemilik. Untuk riwayat transaksi, simpan snapshot minimum dengan sumber dan versi atau waktu pengambilan. Koreksi master kembali ke pemilik. ID dari klien, nama kolom yang sama, dan UUID yang kebetulan cocok tidak membuktikan validitas atau izin.
4. **Pisahkan permintaan dari hasil.** Sebuah command meminta pemilik melakukan transisi. Event pemilik mengumumkan fakta yang sudah ia commit. Untuk integrasi asinkron, tulis event/outbox dan perubahan lokal dalam satu transaksi, tangani duplikasi di konsumen, dan koreksi melalui pemilik asal. Pengiriman pesan, upload bukti, dan selesainya tugas approval masing-masing hanya membuktikan kejadian tersebut. [Aturan outbox](../docs/02-architecture/service-boundaries.md)
5. **Jika kontrak belum tersedia**, pertahankan status menunggu atau kembalikan error yang jelas sesuai kontrak operasi. Catat kebutuhan di `agents/notes/{management_pemilik}/{id}-{management_peminta}.md` dengan operasi, data minimum, aturan scope, dan kondisi gagal. Perubahan pada sisi produsen dikerjakan dalam lingkup management pemilik. Tabel lokal pengganti, respons sukses buatan, atau salinan master akan membuat dua sumber kebenaran. [Pedoman proyek](../AGENTS.md)

6. **Periksa cakupan perubahan saat review.** Untuk tugas satu management, cocokkan setiap berkas yang diubah dengan operasi pemilik di bawah. Jalur yang lazim ialah `backend/services/{management}/`, kontraknya di `contracts/proto/{management}/`, dan adapter/resolver Gateway yang mengekspos operasi itu. Jika diff menyentuh aturan bisnis atau migrasi management kedua, pisahkan pekerjaan milik tim tersebut dan sepakati kontraknya. Untuk integrasi, uji bahwa penolakan dari pemilik tidak menghasilkan perubahan status sukses di peminta.

## Peta pemilik fakta bisnis: 13 management

| Management (service) | Hak tulis utama | Bukti source |
|---|---|---|
| IAM (`iam_management`) | Human/workload principal, kredensial, sesi, RBAC–ABAC, grant dan security audit | [Service IAM](../backend/services/iam_management/README.md) |
| Enterprise Foundation (`enterprise_foundation`) | Legal entity, party, unit organisasi, lokasi dan alamat induk | [Service Foundation](../backend/services/enterprise_foundation/README.md) |
| Product Lifecycle Management (`product_lifecycle_management`) | Identitas/revisi item, SKU, spesifikasi dan rilis produk | [Service PLM](../backend/services/product_lifecycle_management/README.md) |
| E-Commerce Commerce (`ecommerce_commerce`) | Listing kanal, cart, checkout, quote dan submission; bukan sales order | [Service Commerce](../backend/services/ecommerce_commerce/README.md) |
| Pricing & Revenue (`pricing_revenue_management`) | Harga komersial, promo, diskon, currency/expiry; masih scaffold | [Service Pricing](../backend/services/pricing_revenue_management/README.md) |
| Order Management (`order_management`) | Sales order, acceptance/rejection, cancel dan keputusan retur; masih scaffold | [Service OMS](../backend/services/order_management/README.md) |
| Inventory & Warehouse (`inventory_warehouse_management`) | Receipt fisik, saldo/lot, hold, picking, packing dan dispatch | [Service WMS](../backend/services/inventory_warehouse_management/README.md) |
| Finance & Accounting (`finance_accounting`) | AP/AR, receipt terverifikasi, GL, pajak dan laporan keuangan | [Service Finance](../backend/services/finance_accounting/README.md) |
| Procurement & Sourcing (`procurement_sourcing`) | Supplier qualification, sourcing, kontrak, PO dan ASN | [Service Procurement](../backend/services/procurement_sourcing/README.md) |
| Customer Service (`customer_service_management`) | Tiket, komplain dan eskalasi; masih scaffold | [Service Customer Service](../backend/services/customer_service_management/README.md) |
| CRM & Sales (`crm_sales_management`) | Lead, prospek, opt-in campaign dan relationship; masih scaffold | [Service CRM](../backend/services/crm_sales_management/README.md) |
| Document & Records (`document_records_management`) | Versi bukti, digest, record, retensi dan legal hold; runtime belum lengkap | [Service DRM](../backend/services/document_records_management/README.md) |
| Human Resource Management (`human_resource_management`) | Worker, kontrak, posisi, cuti, kompensasi dan payroll | [Service HRM](../backend/services/human_resource_management/README.md) |

## Batas detail management dengan service bisnis yang sudah diperiksa

### iam_management

IAM memegang `principals`, `principal_identifiers`, kredensial dan faktor autentikasi, `authentication_sessions`, identitas layanan, grup, role, izin, grant/delegasi, provisioning dan audit keamanan. IAM menerbitkan dan mencabut sesi, memeriksa izin terhadap scope tepercaya dan menentukan siapa yang bertindak. Modul pemilik transaksi tetap memeriksa syarat bisnis seperti kelayakan kontrak atau persetujuan cuti. Untuk scope unit organisasi, ID unit berasal dari Enterprise Foundation. Untuk akun pekerja, HRM menyimpan `hr_workers.iam_principal_reference_id`; perubahan kontrak/PHK masuk HRM, perubahan akses/sesi masuk IAM. [DBML IAM](../backend/services/iam_management/db/schema/iam_management.dbml) · [Siklus sesi](../backend/services/iam_management/internal/service/session/opaque_runtime.go) · [Evaluasi izin](../backend/services/iam_management/internal/authorization/evaluator.go)

### enterprise_foundation

Foundation menulis badan hukum, registrasi, hierarki `organization_units`, `locations`, alamat, `parties`, peran rekanan, identitas pajak rekanan dan pengenal eksternal. Jabatan dan penempatan pekerja dikelola HRM. `organization_units.head_person_id` adalah referensi pekerja HRM; service unit saat ini menyimpan UUID tersebut. Peran vendor/customer pada `party_role_assignments` menjelaskan hubungan rekanan dengan perusahaan. Kualifikasi pemasok dan PO masuk Procurement; peran akses masuk IAM; fasilitas dan alamat induk masuk Foundation, sedangkan bin dan saldo fisik masuk Inventory. Kode pusat biaya pada unit organisasi tetap atribut organisasi, sedangkan dimensi dan posting jurnal dikelola Finance. [DBML Foundation](../backend/services/enterprise_foundation/db/schema/enterprise_foundation.dbml) · [Hierarki unit](../backend/services/enterprise_foundation/internal/service/organization_unit/service.go) · [Peran rekanan](../backend/services/enterprise_foundation/internal/service/party/party_role_assignment.go)

### human_resource_management

HRM menulis `hr_workers`, profil pribadi, posisi/penempatan, kontrak, rekrutmen dan onboarding, kompensasi/tunjangan, penggajian termasuk BPJS, hak/saldo cuti, kinerja, pembelajaran, suksesi serta hubungan industrial. Untuk posisi, adapter HRM membaca badan hukum, unit, dan lokasi dari Foundation lalu memeriksa status dan keterkaitannya. Untuk pelaku mutasi, interceptor HRM membaca `GetAuthenticationContext` dari IAM. Pemilik kebijakan dan saldo cuti adalah HRM (`ApproveWithBalance`); permohonan absen operasional, shift dan jam yang disetujui masuk rancangan Workforce. HRM menghitung hasil payroll, Finance mengendalikan jurnal dan pembayaran. `hr_worker_documents` mengaitkan pekerja dengan bukti dan status verifikasi HR; pengelolaan berkas berversi/retensi resmi adalah kebutuhan ke Document & Records. [DBML HRM](../backend/services/human_resource_management/db/schema/human_resource_management.dbml) · [Validasi Foundation](../backend/services/human_resource_management/internal/integration/enterprise_foundation/position_references.go) · [Identitas pelaku](../backend/services/human_resource_management/internal/api/authentication.go) · [Koreksi saldo cuti](../backend/services/human_resource_management/internal/service/leave/adjustment_service.go) · [Kalkulasi payroll](../backend/services/human_resource_management/internal/service/payroll/calculation_service.go)

### finance_accounting

- **Hak tulis.** Finance memiliki buku dan periode akuntansi, bagan akun, jurnal, utang (AP), piutang (AR), kas/bank, aset tetap dan depresiasi, pajak pembukuan, pengakuan pendapatan, anggaran, konsolidasi, serta laporan keuangan. Contoh tabelnya `journal_entries`, `ap_invoices`, `ar_receipts`, `fixed_asset_books`, `tax_returns` dan `revenue_recognition_schedules`. Modul asal menyerahkan fakta bisnis yang disetujui. Finance menentukan posting jurnal dan koreksinya. Transfer uang dikerjakan manusia di luar ERP. Finance mencatat dan memverifikasi transaksi internal dengan rujukan ke versi bukti yang dikelola Document Records. [DBML Finance](../backend/services/finance_accounting/db/schema/finance_accounting.dbml) · [Implementasi per domain](../backend/services/finance_accounting/internal/service)
- **Faktur pemasok.** Procurement memiliki PO beserta barisnya, Inventory memiliki kuantitas dan status penerimaan fisik, Foundation memiliki identitas legal pemasok. Finance memiliki faktur, hasil pencocokan, jadwal utang dan pencatatan pembayaran. Saat ini `PerformThreeWayMatching` menolak pencocokan dengan `ErrMatchingSourcesUnavailable` ketika ada rujukan PO/penerimaan. Developer Finance harus meminta data baris, kuantitas, status dan versi dari pemilik sebelum menandai cocok. [Implementasi AP](../backend/services/finance_accounting/internal/service/accounts_payable/service.go) · [Batas Procurement](../backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml)
- **Kasbon pekerja.** Finance membuat dan menyetujui kasbon setelah adapter `GetWorker` membaca pekerja dari HRM dan mencocokkan badan hukum dengan buku Finance. HRM menguasai status kerja. `DisburseCashAdvance` dan `SettleCashAdvance` masih mengembalikan error dependensi. Modul lain tidak boleh menandai kasbon cair/lunas. [Adapter HRM](../backend/services/finance_accounting/internal/integration/hrm/worker.go) · [Alur kasbon](../backend/services/finance_accounting/internal/service/cash_bank/service_cash_advance.go)
- **Nama domain yang beririsan.** Penyusutan aset, pengakuan pendapatan akuntansi, faktur dan pelaporan pajak yang sudah ada tetap di Finance. Folder rencana `enterprise_asset_management`, `pricing_revenue_management`, `tax_localization_compliance` dan `travel_expense_management` tidak memberi izin memindahkan pencatatan tersebut atau membuat buku kedua. [Depresiasi](../backend/services/finance_accounting/internal/service/fixed_asset/service_depreciation.go) · [Pengakuan pendapatan](../backend/services/finance_accounting/internal/service/revenue_recognition/service_schedule.go) · [Pelaporan pajak](../backend/services/finance_accounting/internal/service/taxation/service_tax_return.go)

### procurement_sourcing

Pemilik komitmen pembelian. Delapan domain Go menangani kategori/kebijakan, kualifikasi pemasok, permintaan pembelian, RFx, penawaran dan award, kontrak, PO, serta kolaborasi pemasok. `purchase_orders` merekam pesanan komersial. `inbound_shipment_notices` merekam pemberitahuan kirim pemasok (ASN). Status PO dan ASN menunjukkan perkembangan komitmen serta pengiriman pemasok. Bukti jumlah yang diterima berasal dari penerimaan di gudang. [DBML](../backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml) · [Penerbitan PO](../backend/services/procurement_sourcing/internal/service/purchase_order/service.go) · [Pencatatan ASN](../backend/services/procurement_sourcing/internal/service/supplier_collaboration/service.go)

Serahkan pemeriksaan jumlah fisik dan posting stok ke Inventory, invoice/AP, pencocokan tagihan, pembayaran serta jurnal ke Finance. Profil kualifikasi dan katalog pemasok adalah data pembelian. Identitas pihak dan badan hukum mengacu ke Foundation, identitas produk ke PLM, pelaku persetujuan ke IAM. Dari 55 deklarasi `Table`, delapan ditandai stub: `legal_entities`, `parties`, `organization_units`, `locations` (Foundation), `principals` (IAM), `projects` (Project, rencana), `currencies` dan `reference_units` (Reference Data, rencana). Perlakukan `Ref` ke stub sebagai referensi logis. Consumer event Procurement masih TODO. [Stub dan relasi](../backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml#L1342) · [Consumer](../backend/services/procurement_sourcing/internal/events/consumer.go)

### inventory_warehouse_management

Pemilik stok fisik per gudang/lokasi/lot/serial, unit muat, saldo dan mutasi, penerimaan–putaway, reservasi dan alokasi stok nyata, picking–packing–dispatch, transfer, serta hasil hitung ulang. `CompleteReceivingSession` mencatat transaksi gudang berdasarkan kuantitas yang diterima. `CreateAndPostTransaction` dan fungsi posting internal mengelola mutasi `inventory_transactions` dan `inventory_balances`. Hanya Inventory yang mengoreksi saldo fisik. [DBML](../backend/services/inventory_warehouse_management/db/schema/inventory_warehouse_management.dbml) · [Penerimaan](../backend/services/inventory_warehouse_management/internal/service/inbound/service.go) · [Posting stok](../backend/services/inventory_warehouse_management/internal/service/inventory_control/transaction.go)

Gunakan tipe/ID dokumen asal untuk order, PO dan perintah produksi. Alokasi gudang tidak mengubah janji penjualan atau proyeksi ATP. WMS saat ini dapat membuat shift dan menugaskan pekerja untuk eksekusi lokal di gudang. HRM memegang hubungan kerja. Roster lintas perusahaan dirancang di Workforce dan belum tersinkron ke WMS. `DispatchShipment` mencatat barang berangkat dari gudang. Penugasan angkutan dan perjalanan berada di ranah Logistics yang masih berupa kerangka. Status karantina lot di gudang membatasi pemakaian stok, sedangkan keputusan inspeksi dan CAPA berada dalam rancangan QMS. [Penugasan gudang](../backend/services/inventory_warehouse_management/internal/service/workforce/service.go) · [Dispatch](../backend/services/inventory_warehouse_management/internal/service/shipping_execution/service.go)

### product_lifecycle_management

Pemilik identitas dan revisi produk, konfigurasi/varian, struktur desain (EBOM), spesifikasi dan verifikasi desain, risiko desain, perubahan rekayasa dan waiver, persetujuan rilis, serta bukti regulasi produk/TKDN. Misalnya, `product_structure_lines` memuat komponen struktur desain dan `ReleaseProductStructure` merilis struktur tersebut. Developer Manufacturing menggunakan identitas/revisi itu untuk menyusun MBOM, resep dan routing eksekusi di management-nya. [DBML](../backend/services/product_lifecycle_management/db/schema/product_lifecycle_management.dbml) · [EBOM](../backend/services/product_lifecycle_management/internal/service/bom_product_structure/service.go) · [Revisi item](../backend/services/product_lifecycle_management/internal/service/product_catalog_definition/service_items.go)

Spesifikasi desain diberikan sebagai acuan inspeksi. Hasil inspeksi operasional masuk rancangan QMS, lot/serial dan isi kemasan fisik di Inventory, katalog pemasok di Procurement, harga jual dan listing kanal disisihkan untuk rancangan Pricing dan Commerce. `LinkControlledDocument` mengikat revisi produk ke referensi dokumen. DRM memegang versi dan siklus arsip bila kontraknya tersedia. [Verifikasi desain](../backend/services/product_lifecycle_management/internal/service/specification_verification_risk/service_verifications_risks.go) · [Tautan dokumen produk](../backend/services/product_lifecycle_management/internal/service/release_regulatory_tkdn/service_artifacts_documents.go)

## Management yang masih bertahap atau scaffold

### ecommerce_commerce

Commerce memiliki skema DBML dan implementasi bertahap untuk channel Stage A; API dan E2E lintas owner harus dibuktikan sebelum aktif. Commerce memiliki channel, listing, cart, quote dan submission, **bukan** pesanan penjualan resmi, harga akhir owner atau status dana. PLM menyediakan identitas produk; Pricing harga; OMS menerima/menolak order; WMS memegang stok dan fulfillment; Finance memverifikasi receipt. [Commerce](../backend/services/ecommerce_commerce/README.md) · [Kontrak lintas domain](../docs/02-architecture/service-boundaries.md).

### pricing_revenue_management

Pricing pemilik keputusan harga jual, promo, diskon dan batas waktu quote. Pada snapshot ini belum terdapat DBML maupun service bisnis lengkap di checkout; pengguna kontrak Pricing harus fail-closed saat harga tidak tersedia, bukan menetapkan fallback permanen. [Scaffold Pricing](../backend/services/pricing_revenue_management/README.md).

### order_management

OMS satu-satunya pemilik sales order dan keputusan ACCEPTED/REJECTED serta cancel/return. Commerce hanya mengajukan pesanan idempoten; pengiriman fisik WMS dan penerimaan uang Finance tidak boleh diganti event transport. Snapshot saat ini masih scaffold. [Scaffold OMS](../backend/services/order_management/README.md).

### crm_sales_management

CRM memiliki lead, pipeline, hubungan pelanggan dan consent pemasaran sesuai kontrak. Akun buyer Commerce, ticket Customer Service dan invoice Finance bukan master alternatif CRM. Snapshot saat ini masih scaffold. [Scaffold CRM](../backend/services/crm_sales_management/README.md).

### customer_service_management

Customer Service mengelola tiket, komplain, prioritas dan eskalasi masalah pelanggan. Pemutakhiran order, stok, pembayaran atau dokumen resmi tetap harus diminta kepada service pemilik. Snapshot saat ini masih scaffold. [Scaffold Customer Service](../backend/services/customer_service_management/README.md).

### document_records_management

DRM mengelola metadata dokumen, versi immutable, integritas/digest, klasifikasi, retensi dan legal hold; bukti yang disimpan **tidak** mengesahkan status pembayaran atau PO. DBML dan baseline migrasi ada, tetapi service bisnis/E2E tidak boleh dianggap lengkap. [Skema dan status DRM](../backend/services/document_records_management/README.md).

## Alur lintas-owner yang perlu dibuktikan

| Permintaan | Pembagian hak tulis |
|---|---|
| PO pemasok, ASN, menerima barang, lalu mencocokkan invoice | Procurement memiliki PO/ASN; WMS memiliki jumlah penerimaan dan stok nyata; Finance hanya melakukan matching dan pencatatan menurut sumber versi/status yang sah. Tidak boleh menganggap kuantitas order sama dengan kuantitas terima. |
| Perpindahan pegawai dan pembaruan akses | Foundation menulis unit/lokasi; HRM menulis penempatan; IAM menilai pengajuan akses, grant, revoke dan sesi. Perubahan HRM tidak otomatis memberi IAM role. |
| Checkout website atau order marketplace | Commerce mengelola listing/cart/quote/submission; Pricing menentukan harga; OMS memiliki sales order; WMS memiliki stock hold/dispatch; Finance memiliki AR/receipt/GL. Gunakan idempotency, outbox/inbox dan rekonsiliasi, bukan shared SQL. |
| Dokumen bukti pembayaran selesai diunggah | DRM mengelola versi dan integritas dokumen, tetapi hanya Finance yang menilai kebenaran pembayaran dan memperbarui transaksi. |

**Dependency non-scope:** kebutuhan seperti Manufacturing, QMS, workforce, transportasi/logistik tersendiri, Planning atau EAM pernah muncul dalam audit lama, tetapi **tidak** merupakan service management aktif di tree sekarang. Jangan mengarahkan perubahan ke folder/document yang tidak ada atau menganggapnya tersedia. Catat permintaan ke management owner yang relevan dalam daftar 13 dan mintakan keputusan owner bila memerlukan management baru.

## Gerbang integrasi, aktor dan error (panduan review lintas management)

Sebuah management yang sudah memiliki DBML, health endpoint, skema GraphQL, atau kontrak proto **belum otomatis menjadi service bisnis yang siap dipanggil**. Bedakan secara eksplisit: keputusan owner (DECIDED), schema/kontrak (DESIGNED), logika yang telah tersambung (IMPLEMENTED), hasil E2E nyata (VERIFIED), dependency tidak tersedia (BLOCKED), dan deploy (DEPLOYED). Jangan memasukkan service design-only ke daftar integrasi aktif hanya karena foldernya ada.

Setiap interface yang menyeberang ownership harus memuat pemilik sumber kebenaran; identity audience/realm dan scope; ID dan versi objek; masa berlaku hasil; deadline/timeout; idempotency key; status pending versus final; kode error bisnis versus dependency; serta strategi retry dan rekonsiliasi. Gateway boleh menerapkan penolakan perimeter, tetapi **service pemilik** yang memeriksa permission, scope objek aktual, atribut ABAC, SoD dan invariant pada query/mutation maupun RPC. Rinciannya ada di [inter-service boundary](../docs/02-architecture/service-boundaries.md) dan [arsitektur IAM](arsitektur.md).

Kasus penolakan **tidak boleh menambah data domain atau outbox**. Percobaan akses oleh principal sah yang tidak berizin harus dibedakan dari sesi kedaluwarsa atau IAM/Redis unavailable; penolakan harus berlaku juga pada `node(id:)`, nested field, list, count, export, dan cursor. Katalog izin per management disepakati dengan IAM melalui [proses role dan permission](../agents/principals.md) dan bukan diambil dari `principals.attributes`.

**Kontrak yang belum tersedia:** catat dependency pada pemilik yang termasuk scope aktif dan jangan menciptakan stub lintas database, FK lintas service, atau respons sukses palsu demi meloloskan diagram.

## Memperbarui batas

Saat menambah service bisnis atau mengubah pemilik data, perbarui baris dan bagian terkait pada perubahan yang sama. Cantumkan tabel/agregat, operasi yang menulisnya, service yang mengonsumsi, dan tautan ke implementasi. Jika sebuah integrasi dinyatakan tersedia, sertakan kontrak dan bukti pemanggilan ke pemilik. Selama hanya ada rancangan atau stub, pertahankan statusnya di inventaris dan catat dependensi yang masih dibutuhkan.
