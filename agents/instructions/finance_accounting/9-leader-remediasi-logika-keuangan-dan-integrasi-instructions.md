# Remediasi Finance Accounting dan integrasi backend

- **id** — 9
- **owner** — Full Backend Developer Finance Accounting. Reviewer Leader Backend. HRM dan Web Commerce memiliki tugas pendamping pada catatan khusus.
- **priority** — P0 untuk pembatasan laporan, akun pajak dan publikasi payroll. Susul dengan integritas subledger dan treasury, kemudian perluasan produk.
- **status** — SIAP DIKERJAKAN
- **baseline** — Checkout bersama ERP Retail, `main` berbasis `bba5f46cfec19129dce58b8a70d813e19f4ce3a1`, 29 November 2026 WIB. Periksa versi lokal terkini sebelum menambal fungsi karena beberapa owner bekerja bersamaan.

## Task

Kerjakan FA-01 sampai FA-12 sampai hasil operasional dan bukti pengujiannya tersedia. Mulai dari [temuan audit](../../../agents/audit/finance_accounting/10-2026-11-29-010010-audit-logika-keuangan-dan-integrasi.md) dan [daftar 45 root, seluruh 74 RPC dan usulan baru](../../../agents/grpc/finance_accounting/7-matriks-operasi-publik-privat-dan-usulan-keuangan.md). Gunakan [format instruksi](../../../agents/instruction_template.md), [panduan gRPC](../../../agents/how_to_code_grpc_managements.md) dan [panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md). Jalankan tahap internal dan mitra lebih dahulu. Kerjakan kontrak dana Commerce sebelum membuka checkout prabayar. Perlakukan daftar fitur tambahan sebagai desain yang harus memenuhi invariant di bawah.

## Request

### FA-01 · Batasi laporan pada entitas dan lokasi yang diizinkan

- file: `backend/gateway/internal/adapter/finance_accounting/permissions.go:49` · `backend/gateway/internal/adapter/finance_accounting/reporting_adapter.go:15` · `backend/services/finance_accounting/internal/security/policy.go:189` · `backend/services/finance_accounting/internal/service/reporting/service.go:64` · `backend/services/finance_accounting/repository/accounting/journal_repository.go:134` · `backend/services/finance_accounting/repository/trade/tax_repository.go`.
- function: `GetTrialBalance, GetProfitAndLoss, GetTaxSummary, RequireScope`
- logic: Turunkan pasangan izin dan scope dari snapshot IAM. Resolusi legal entity/facility memakai relasi EF yang berlaku. Filter transaksi dan hitungan sebelum SUM atau pagination. Untuk laporan yang belum memiliki dimensi terpercaya, syaratkan izin global eksplisit sampai dimensi tersedia.
- error_handling: Tolak target di luar grant. Jika snapshot atau resolusi relasi tidak tersedia, kembalikan kegagalan sebelum membaca hasil.
- security: Terapkan di service untuk Gateway maupun gRPC. Pertahankan target grant yang berpasangan dengan permission. Audit juga penulis manual journal dan konfigurasi global terhadap objek yang diubah.
- expected_result: Pengguna fasilitas A melihat saldo A pada filter kosong dan tidak bisa membaca B. Pengguna global mendapatkan total yang dapat direkonsiliasi.
- testing: Gunakan dua fasilitas dengan transaksi berbeda. Bandingkan nilai, jumlah baris, total, rentang tanggal dan akses langsung melalui service.

### FA-02 · Pisahkan pajak keluaran, pembulatan dan akun kontrol piutang

- file: `backend/services/finance_accounting/internal/service/accounts_receivable/service.go:233` · `backend/services/finance_accounting/transport/grpc/handler/ar_handler.go:102` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto` · `backend/services/finance_accounting/internal/service/accounting/service.go:545`.
- function: `PostARInvoice, ConfigurePostingRule`
- logic: Gunakan snapshot pajak per baris dan aturan akun Finance efektif pada tanggal posting. Debit akun kontrol AR, kredit pendapatan neto dan kewajiban pajak. Pakai akun pembulatan khusus selisih pembulatan bertanda. Perlakukan akun dari klien lama sebagai assertion. Sediakan laporan identifikasi jurnal historis terdampak untuk koreksi berotorisasi.
- error_handling: Tolak akun kontrol berbeda, pemetaan kosong/tumpang tindih dan snapshot tidak sah sebelum membuat jurnal. Jurnal yang sudah posted dikoreksi dengan reversal/adjustment terotorisasi.
- security: Batasi posting per badan hukum, pisahkan pembuat dan penyetuju serta simpan versi rule dan referensi pajak.
- expected_result: Saldo per akun, outstanding pelanggan dan ringkasan pajak cocok dengan invoice. Pembulatan dan pajak dapat diaudit terpisah.
- testing: Invoice kena pajak dengan pembulatan positif/negatif, tarif nol, pemetaan absen, retry bersamaan dan verifikasi debit/kredit per akun.

### FA-03 · Hormati kontrak pembuatan invoice AR

- file: `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto:39` · `backend/services/finance_accounting/transport/grpc/handler/ar_handler.go:54` · `backend/services/finance_accounting/internal/service/accounts_receivable/service.go:114`.
- function: `CreateARInvoice`
- logic: Validasi kuantitas serta harga dengan skala domain. Hitung neto/pajak dan total di service berdasar currency settings. Cocokkan periode/tanggal, status customer EF dan due date. Simpan key, identitas sumber, mata uang serta hash isi kanonik dalam transaksi. Dukung FX hanya ketika kurs dan jurnal mata uang fungsional dapat diverifikasi.
- error_handling: INVALID_ARGUMENT untuk angka atau mata uang tak didukung, FAILED_PRECONDITION untuk periode tertutup, konflik bila key sama memuat input berbeda. Replay identik mengembalikan ID awal.
- security: Scope entity, izin AR create dan referensi pelanggan authoritative. Redaksi detail komersial dalam log.
- expected_result: Handler, service dan record menggunakan nilai ekonomi yang sama. Satu key menghasilkan satu invoice termasuk sesudah timeout.
- testing: Pecahan di atas dua desimal, mata uang skala nol, rate invalid, periode salah dan dua request paralel.

### FA-04 · Samakan nilai matching AP dan jurnal persediaan

- file: `backend/services/finance_accounting/internal/service/accounts_payable/service.go:195` · `backend/services/finance_accounting/internal/service/accounts_payable/service.go:614` · `backend/services/finance_accounting/internal/service/accounts_payable/service.go:836` · `backend/services/finance_accounting/internal/integration/procurement/client.go` · `backend/services/finance_accounting/internal/integration/warehouse/client.go`.
- function: `PerformThreeWayMatching, PostAPInvoice`
- logic: Gunakan nilai neto prorata untuk setiap receipt, alokasikan residu pada match terakhir. Untuk barang yang sudah diakru, clear GR/IR terhadap accrual aktual. Bagi selisih menurut kuantitas masih dimiliki, telah terjual atau masuk WIP. Untuk jasa gunakan acceptance Procurement dan alur nota kredit yang tersedia. Rekonsiliasi nilai serta kuantitas sebelum approve/post.
- error_handling: Tolak overmatch, retur/reversal sebelum posting, bukti owner berubah dan sisa alokasi tidak seimbang. Lock baris sumber dan invoice.
- security: SoD, expected version, scope pihak dan lokasi. Terima kuantitas dari bukti owner dan gunakan akun kendali Finance.
- expected_result: GR/IR terurai per penerimaan, AP sesuai invoice neto, nilai sisa persediaan dan HPP sesuai konsumsi.
- testing: Barang sebagian terjual, diskon satu invoice dua receipt, retur parsial, nota kredit jasa, replay dan pembulatan terakhir.

### FA-05 · Simpan kedua sisi aturan posting

- file: `backend/gateway/internal/adapter/finance_accounting/accounting_adapter.go:299` · `backend/services/finance_accounting/transport/grpc/handler/accounting_handler.go:245` · `backend/services/finance_accounting/internal/service/accounting/service.go:545` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounting.proto`.
- function: `ConfigurePostingRule, configureEffectivePostingRule`
- logic: Pertahankan pembaca lama, simpan kedua sisi dan status ekonomi secara atomik atau kembalikan error untuk bentuk yang belum dapat direpresentasikan. Beri periode efektivitas dan dukung penonaktifan. Petakan rule lama sebelum mengubah caller.
- error_handling: Konflik rentang efektif ditolak. Kegagalan sisi kedua me-rollback sisi pertama.
- security: Perubahan global memerlukan scope yang sesuai dan otorisasi sensitif saat commit. Simpan aktor dan perubahan akun.
- expected_result: Baca kembali debit, kredit, tanggal dan status dari penyimpanan. Respons sukses sama dengan aturan yang akan dipilih saat posting.
- testing: Rule nonaktif, satu sisi gagal, overlap, replay, dan kompatibilitas varian payroll/finance.

### FA-06 · Cegah saldo awal menimpa valuasi dan tentukan biaya issue

- file: `backend/services/finance_accounting/internal/service/inventory_cost/service.go:83` · `backend/services/finance_accounting/repository/inventory_cost/inventory_cost_repository.go:107` · `backend/services/finance_accounting/transport/grpc/handler/inventory_cost_handler.go:49` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/inventory_cost.proto`.
- function: `ApplyInventoryCostEvent, SelectBalanceByItemForUpdate, UpsertBalance`
- logic: Inisialisasi saldo dengan insert-on-conflict-do-nothing lalu kunci saldo. Ambil biaya issue dari buku Finance berdasarkan metode efektif. Simpan fakta sumber, urutan dan tanggal valuasi. Bedakan perpindahan custody milik perusahaan dari pelepasan barang. Koreksi melalui event kompensasi.
- error_handling: Tolak sumber tidak sah, konflik digest dan urutan yang melanggar dependensi. Kesalahan baca/parse saldo menghentikan transaksi.
- security: Allowlist workload mTLS per jenis event. Verifikasi movement WMS dan cocokkan item, kuantitas, pemilik serta tanggal.
- expected_result: Dua receipt pertama terserialisasi tanpa kehilangan nilai. Issue/HPP konsisten dengan buku, perpindahan custody menjaga nilai total.
- testing: Dua receipt simultan, retry identik, digest berbeda, cost lama, perubahan lokasi dan issue saat stok tidak cukup.

### FA-07 · Tutup periode dengan bukti rekonsiliasi

- file: `backend/services/finance_accounting/internal/service/accounting/service.go:272` · `backend/services/finance_accounting/internal/service/accounting/service.go:617` · `backend/services/finance_accounting/repository/accounting/period_repository.go:151` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/reporting.proto` · `backend/services/finance_accounting/internal/service/reporting/service.go:96`. Usulan file baru `backend/services/finance_accounting/internal/service/accounting/close_readiness.go`, `backend/services/finance_accounting/internal/service/reporting/balance_sheet.go`, `backend/services/finance_accounting/internal/service/reporting/cash_flow.go`.
- function: `CloseAccountingPeriod, CreateJournalDraft, GetPeriodCloseReadiness, GetBalanceSheet, GetCashFlow`
- logic: Kunci periode yang sama sebelum validasi dan insert semua penulis. Definisikan close readiness per subledger material dan bank: cutoff, outstanding, rekonsiliasi GL, status dan bukti exception dengan approver. Neraca pakai saldo sampai tanggal, arus kas pakai transaksi kas dan klasifikasi arus yang dapat ditelusuri.
- error_handling: FAILED_PRECONDITION dengan kategori blocker yang dapat dibaca operator. Revalidasi snapshot readiness saat commit. Reopen dengan alasan dan otorisasi.
- security: Buka/tutup dan pengecualian memerlukan izin, scope dan rekam audit terpisah. Gunakan filter FA-01 untuk semua laporan.
- expected_result: Setelah close tidak ada draft baru untuk periode itu. Aset = liabilitas + ekuitas dan perubahan kas sesuai laporan arus kas.
- testing: Race draft–close, posting sebelum/sesudah cutoff, subledger belum direkonsiliasi, reopen, transaksi nonkas dan saldo awal.

### FA-08 · Validasi mutasi bank dan selesaikan rekonsiliasi

- file: `backend/services/finance_accounting/transport/grpc/handler/treasury_handler.go:153` · `backend/services/finance_accounting/internal/service/treasury/service.go:687` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto` · `backend/services/finance_accounting/db/migrations/000001_create_finance_accounting_schema.up.sql:696` · `backend/services/finance_accounting/internal/service/treasury/payout_service.go:33` · `backend/services/finance_accounting/infrastructure/third_party/xendit/payout.go`. Usulan file baru `backend/services/finance_accounting/internal/service/treasury/bank_reconciliation.go`.
- function: `ImportBankStatement, ScheduleSupplierPayout, ApprovePayoutSchedule, MatchBankStatementLine, ApproveBankReconciliation`
- logic: Parse debit/kredit desimal, tepat satu sisi positif, lalu validasi saldo. Gunakan match terhadap journal line bank yang posted, batas jumlah unmatched, fee/selisih dengan jurnal sah, maker-checker dan riwayat supersede. Buka penjadwalan dan approval payout pemasok di Gateway setelah gate. Worker Xendit sandbox tetap privat dan memulihkan status berdasar reference/idempotency key.
- error_handling: Konflik duplikat dengan isi berbeda, debit/kredit ambigu, satu baris dipakai melebihi saldo dan status provider tak pasti harus terlihat di antrean. Setelah timeout rekonsiliasi attempt sebelum retry pengiriman.
- security: Scope rekening/badan hukum, MFA untuk approve, redaksi instrumen dan larangan pemohon menyetujui sendiri. Finance satu-satunya pemilik transfer.
- expected_result: Rekonsiliasi dapat ditelusuri dari statement ke jurnal, fee dan saldo. Satu jadwal yang sudah jatuh tempo menghasilkan satu kewajiban transfer.
- testing: Kredit dengan debit 0.000000, kedua sisi positif, split match, attempt timeout, replay provider, jam dan zona waktu perusahaan di sandbox.

### FA-09 · Teruskan bukti publikasi payroll sampai slip tersimpan

- file: `backend/services/finance_accounting/internal/service/payroll/service.go:607` · `backend/services/finance_accounting/internal/service/payroll/service.go:891` · `backend/services/finance_accounting/internal/service/payroll/service.go:1062` · `backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go:311` · `backend/services/finance_accounting/internal/service/treasury/payroll_allocation.go` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto` · [backend/services/human_resource/internal/service/payroll/payslip_publication.go:149](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L149) · [backend/services/human_resource/internal/service/payroll/payslip_reissue.go:79](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L79) · `backend/services/human_resource/infrastructure/integration/finance/payroll_verifier.go`.
- function: `PublishPayroll, VerifyPayrollPublication, FindPayrollPublicationByRun, PublishPayslipForResult, ReissueEmployeePayslip`
- logic: Gunakan satu dari journal_id atau no_journal_posting_id di publication, lookup dan event. Cocokkan bukti, entity, run dan digest terhadap posting Finance. Untuk settlement, bedakan no_transfer_required, partially_paid, paid, reversed, termasuk outstanding per worker. Pertahankan edisi slip immutable dan reconciliasi DRM.
- error_handling: Tolak pasangan kosong/ganda, mismatch dan event versi lama tanpa bukti terverifikasi. Retry DRM membuat satu versi per publikasi/hasil. Publikasi Finance dapat selesai sebelum cash settlement.
- security: HR mengajukan rincian. Finance memublikasikan dan mencairkan. Employee aktif dapat mencetak sendiri, mantan karyawan melalui HR berwenang dengan bukti audit. Batasi data sensitif per recipient.
- expected_result: Slip dibuat setelah publikasi, termasuk run tanpa jurnal. Kehilangan PDF dapat dipulihkan dari edisi terverifikasi. Status nol tidak menampilkan transfer uang.
- testing: Run dengan dan tanpa jurnal, hasil nol, dua pembayaran parsial, pembalikan, DRM unavailable lalu pulih, dan permintaan cetak mantan karyawan.

### FA-10 · Hubungkan dana pelanggan dan refund per fulfillment

- file: `backend/services/web_commerce/infrastructure/integration/finance/client.go:49` · `backend/services/web_commerce/services/checkout/service.go` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto` · `backend/services/finance_accounting/infrastructure/inbox/processor.go` · `backend/services/finance_accounting/infrastructure/outbox/publisher.go`. Usulan file baru `backend/services/finance_accounting/contract/proto/finance_accounting/v1/commerce_funding.proto`, `backend/services/finance_accounting/internal/service/accounts_receivable/customer_advance.go`.
- function: `CalculateSalesTax, RequestPaymentSession, RequestCustomerRefund, ReserveFulfillmentFunding, ReleaseFulfillmentFunding`
- logic: Bangun kontrak privat Finance untuk quote pajak, payment session, status dana, reserve/release/consume per shipment, refund dan PSP settlement. Setelah bukti penerimaan verified, catat uang muka pelanggan. Gunakan bukti fulfillment sesuai kebijakan untuk revenue/COGS. Jurnal biaya provider dan net bank direkonsiliasi terpisah. Hormati prioritas internal → mitra → pelanggan.
- error_handling: Retry memakai sumber dan key yang sama. Satu reservasi tidak boleh sekaligus refund dan dipakai dispatch. Timeout menunggu status terverifikasi. Snapshot cap refund setelah alokasi, refund, return dan reversal.
- security: Workload Commerce/OMS/WMS sesuai peran, validasi owner order dan legal entity. Bedakan jalur collection/refund dari adapter payout pemasok. Status browser tidak menjadi bukti penerimaan.
- expected_result: Split shipment hanya mengonsumsi bagian didanai, refund dibatasi saldo aktual dan pendapatan dapat ditelusuri ke pemenuhan. Finance memberi status terverifikasi ke Commerce.
- testing: Split shipment, funding kurang, refund bersamaan dengan dispatch, callback/outbox duplikat, settlement net fee serta refund sebelum dan sesudah sale.

### FA-11 · Lengkapi lifecycle aset sebelum membuka depresiasi

- file: `backend/services/finance_accounting/internal/service/fixed_assets/service.go:103` · `backend/services/finance_accounting/internal/service/fixed_assets/service.go:135` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/fixed_assets.proto`.
- function: `RegisterFixedAsset, CapitalizeFixedAsset, PlaceFixedAssetInService, PostDepreciation`
- logic: Gunakan bukti acquisition/AP yang posted. Catat capitalization, cost basis, residual, metode, useful life dan tanggal mulai pakai. Aktivasi dan jurnal perolehan/reklasifikasi harus kompatibel dengan jurnal AP supaya harga perolehan dihitung sekali.
- error_handling: Tolak capitalization ganda, umur atau residual invalid, tanggal sebelum aktivasi, periode tutup dan depresiasi duplikat pada aset/periode.
- security: Cek kustodian lewat HRM, lokasi EF, otorisasi akuisisi dan posting Finance. Bukti DRM berdasarkan versi.
- expected_result: Aset baru dapat melewati draft → in service → depreciation. Jadwal dan GL dapat direkonsiliasi per aset.
- testing: Aset siap pakai di tengah bulan sesuai policy, nilai residu, dua retry per bulan, periode tutup dan transfer lokasi.

### FA-12 · Hitung biaya produksi yang dititipkan ke vendor

- file: `backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/external_production_vendor_control.proto` · `backend/services/inventory_warehouse/contract/proto/inventory_warehouse/v1/production_reporting_integration.proto` · `backend/services/product_lifecycle/contract/proto/product_lifecycle/v1/variant.proto` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/inventory_cost.proto` · `backend/services/finance_accounting/internal/service/inventory_cost/service.go:63`. Usulan file baru `backend/services/finance_accounting/internal/service/inventory_cost/production_cost.go`.
- function: `GetProductionBatchCost, PostProductionCostAllocation`
- logic: Hubungkan order dan jasa vendor, custody perusahaan, material issued/returned, scrap, hasil QC dan versi item. Alokasi WIP menurut kebijakan biaya terdokumentasi, abnormal loss terpisah, hanya hasil diterima masuk barang jadi. Koreksi biaya vendor terlambat dengan delta dan bukti asal.
- error_handling: Tunda kapitalisasi jika proof QC/material kurang, blok alokasi ganda dan selisih di luar policy. Pertahankan identitas sumber.
- security: Procurement pemilik dokumen vendor, WMS pemilik stok dan QC yang diotorisasi, PLM pemilik versi produk, Finance pemilik valuasi. Manufacturing Operations tetap berstatus paused.
- expected_result: Saldo WIP awal + tambahan − hasil − kerugian = saldo akhir per batch, tanpa menduplikasi stok atau master produk.
- testing: Dua output satu batch, reject sebagian, material milik vendor, invoice terlambat, retur sisa dan replay.

## Action

1. **DB dan kontrak.** Baca DBML, SQL baseline dan constraint terkait. Pertahankan data posted, siapkan kompatibilitas record lama dan rencana migrasi untuk deployment yang telah memiliki data. Ikuti aturan migrasi repo. Susun perintah eksplisit, tipe decimal dan evidence pada proto. Regenerasi kontrak dengan target resmi, cek kedua caller.
2. **Service dan integrasi.** Terapkan perubahan di repository dan service owner. Untuk setiap efek, simpan perubahan domain, jurnal yang diperlukan dan outbox dalam transaksi yang sama. Inbox mendeteksi duplikat dari event ID, producer, versi dan digest. Error dependensi tetap retryable, mismatch permanen masuk antrean rekonsiliasi. Selesaikan dua sisi wiring bersama owner. [Tugas HRM](../../../agents/notes/human_resource/1-finance_accounting.md) · [Tugas Web Commerce](../../../agents/notes/web_commerce/2-finance_accounting.md)
3. **Gateway dan UI.** Ikuti pemetaan satu per satu pada matriks. Perbarui katalog izin, resolver domain, adapter gRPC, operasi SDL, error mapping serta entry point yang sudah dipakai. Gunakan GraphQL via HTTPS pada `/api/graphql` dan gRPC internal dengan mTLS. Pertahankan API lama dengan pemetaan yang benar. Dokumentasikan perubahan aktif dan operasi privat di `agents/grpc/finance_accounting/`.
4. **Verifikasi.** Unit test server gRPC ditempatkan pada layer service. Jalankan `go vet ./...` dan `go test -race -count=1 ./internal/service/...` dari modul Finance. Untuk perubahan Gateway jalankan pengujian adapter/resolver yang terpengaruh. Uji E2E melalui Gateway dengan login superadmin, lalu kelola role, permission, scope dan assignment melalui GraphQL IAM. Lanjutkan verifikasi memakai dua aktor dengan scope berbeda dan approver independen. Verifikasi HTTPS/mTLS, audit trail dan saldo akhir. Simpan laporan E2E aktual, status dependency serta SHA per milestone. Tandai instruksi done sesudah seluruh acceptance terbukti.

### Kepemilikan event dan wiring

| Produsen → penerima | Fakta dan tanggung jawab | Waktu dan penanganan ulang |
| --- | --- | --- |
| Procurement/WMS → Finance | PO/receipt/acceptance, return dan reversal menyertakan baris, unit, owner, legal entity, tanggal, kuantitas, versi serta bukti aktif. | Verifikasi ulang pada approve/post/payout. Tahan kewajiban terkait bila receipt dibalik. |
| Finance → Procurement | Kirim status invoice/jurnal dan kas posted, termasuk reversal atau credit note, jumlah teralokasi dan sisa per invoice/PO. | Kirim setelah commit, pisahkan scheduled/processing dari paid. Consumer dedup per efek dan versi. |
| HRM → Finance → HRM/DRM | HRM memberi snapshot approved dan digest, Finance mencatat akrual, publikasi dan kas dengan bukti bertipe, HRM memicu PDF ke DRM. | Idempotensi per run/result/publication, hanya penerima berwenang mendapatkan data worker. Perbaiki secara terkoordinasi lewat FA-09. |
| Commerce/OMS/WMS → Finance → Commerce | Order, alokasi fulfillment, bukti pengiriman, pembatalan atau return menentukan reservation, release dan konsumsi dana. | Korelasikan versi event, order, shipment dan refund. Satu bagian dana hanya dapat digunakan sekali. |
| Bank/PSP → Finance | Verifikasi status provider dan alokasi per rekening. Clearing, fee dan unmatched diproyeksikan untuk rekonsiliasi. | Timeout tetap pending sampai ada bukti provider/bank. Worker otomatis menggunakan kalender/jadwal perusahaan, satu identitas payout per attempt ekonomi. |
| EF, IAM, DRM dan PLM → Finance | Verifikasi scope dan master EF, role/sesi IAM, dokumen DRM serta versi item PLM dari owner. | Simpan referensi versi dan digest yang diperlukan saja. Untuk referensi wajib yang gagal diverifikasi, hentikan tindakan finansial. |

### Acceptance Criteria dan laporan akhir

| Kasus | Bukti penerimaan |
| --- | --- |
| TC-01 · akses | Dua scope dan grant global, pencabutan akses, MFA serta penolakan self-approval pada operasi sensitif. Bandingkan hasil agregat dan jumlah row. |
| TC-02 · pembukuan | Debit/kredit **per akun**, pajak, outstanding, GR/IR, persediaan dan saldo kas sesuai peristiwa ekonomi, termasuk replay dan reversal. |
| TC-03 · konkurensi | Dua transaksi awal saldo stok, close versus draft dan refund versus reserve. Buktikan state persisten dan identitas transaksi yang menang. |
| TC-04 · integrasi | HRM–Finance–DRM dengan jurnal dan no-journal, bukti WMS/Procurement berubah, retry inbox/outbox dan pemulihan setelah recipient offline. |
| TC-05 · akses operator | Query/mutation berjalan via HTTPS dengan response terscope, picker menampilkan nama/nomor dokumen. Xendit hanya sandbox dan statusnya direkonsiliasi. |

Laporan akhir per FA memuat lokasi file/fungsi, kontrak yang berubah dan owner, migrasi/kompatibilitas, hasil pengujian aktual, SHA, bukti status Gateway, dan kasus yang masih menunggu dependensi. Untuk demo payout gunakan dokumentasi SDK yang dipin modul melalui `go doc github.com/xendit/xendit-go/v7/payout PayoutApi` dan `go doc github.com/xendit/xendit-go/v7/payout CreatePayoutRequest` dari modul Finance.

## Context

**Urutan baca file.** Awali dengan [AGENTS.md](../../../AGENTS.md) · [agents/arsitektur.md](../../../agents/arsitektur.md) · [agents/boundary.md](../../../agents/boundary.md) · [agents/users.md](../../../agents/users.md) · [docs/README.md](../../../docs/README.md) · [docs/03-managements/finance.md](../../../docs/03-managements/finance.md) · `agents/logic/finance_accounting/logic.md`. Baca konteks checkout di `agents/context/finance_accounting/1-context.md`, lalu [agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md](../../../agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md) · [agents/instructions/finance_accounting/7-leader-remediasi-code-inti-finance-accounting-instructions.md](../../../agents/instructions/finance_accounting/7-leader-remediasi-code-inti-finance-accounting-instructions.md) · [agents/instructions/finance_accounting/8-leader-remediasi-procurement-hrm-finance-instructions.md](../../../agents/instructions/finance_accounting/8-leader-remediasi-procurement-hrm-finance-instructions.md). Gunakan matriks untuk nama SDL dan resolver aktual. Path kode berformat monospace adalah file checkout lokal, path berlabel usulan merupakan file yang dapat dibuat saat implementasi.

| Tahap | File atau direktori aktual |
| --- | --- |
| Model dan migrasi | `backend/services/finance_accounting/db/schema/finance_accounting.dbml` · `backend/services/finance_accounting/db/migrations/000001_create_finance_accounting_schema.up.sql` · `backend/services/finance_accounting/db/migrations/000001_create_finance_accounting_schema.down.sql` |
| Repository, service dan composition root | `backend/services/finance_accounting/repository/registry.go` · `backend/services/finance_accounting/internal/service/registry.go` · `backend/services/finance_accounting/cmd/wiring.go` |
| IAM dan Gateway | [backend/gateway/graphql/schema.graphql](../../../backend/gateway/graphql/schema.graphql) · `backend/gateway/graphql/finance_accounting/finance_common.graphql` · [backend/gateway/internal/source_connections/registry.go](../../../backend/gateway/internal/source_connections/registry.go) · `backend/shared/rbac/catalogs/finance_accounting.json` |
| Inbox / outbox dan recipient | `backend/services/finance_accounting/infrastructure/outbox/hrm_payroll_publisher.go` · `backend/services/finance_accounting/infrastructure/inbox/payroll_worker.go` · `backend/services/finance_accounting/infrastructure/inbox/service_acceptance_worker.go` · `backend/services/finance_accounting/internal/integration/ef/client.go` |
| Referensi master dan bukti | `backend/services/finance_accounting/internal/integration/ef/client.go` · `backend/services/finance_accounting/internal/integration/drm/client.go` · `backend/services/finance_accounting/internal/integration/hrm/client.go` · `backend/services/finance_accounting/internal/service/expenses/service.go` · `backend/services/finance_accounting/internal/service/budget/service.go` |
| Procurement dan dokumen | `backend/services/procurement_sourcing/internal/integration/finance/client.go` · `backend/services/document_record/contract/proto/document_record/v1/ingestion_publication.proto` |

File untuk setiap operasi dan fungsi yang dikerjakan tercantum pada blok FA-01 sampai FA-12 di atas.
