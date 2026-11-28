# Audit logika Finance Accounting dan integrasi retail

Audit kode produksi pada checkout bersama `main` dengan basis `bba5f46cfec19129dce58b8a70d813e19f4ce3a1`, 29 November 2026 WIB. Fokus pemeriksaan adalah kontrak yang dipanggil, kontrol di service, efek per akun dan bukti yang diterima management berikutnya. Nama file berformat kode menunjuk sumber lokal yang belum tercatat pada commit dasar. Temuan merupakan hasil inspeksi statis. Dampak pada data historis dan kondisi deployment membutuhkan verifikasi pelaksana.

**Hasil.** Ditemukan masalah pembatasan laporan per fasilitas, pajak AR masuk akun pembulatan dan bukti payroll tanpa jurnal yang belum konsisten sampai slip. Dari 74 RPC di 11 service gRPC, 43 dipanggil lewat 45 field root Finance di Gateway. [Matriks operasi dan saran publikasi](../../../agents/grpc/finance_accounting/7-matriks-operasi-publik-privat-dan-usulan-keuangan.md) · [Instruksi implementasi berurutan](../../../agents/instructions/finance_accounting/9-leader-remediasi-logika-keuangan-dan-integrasi-instructions.md)

## Keputusan atas usulan domain

| Bagian usulan | Keputusan dan alasan |
| --- | --- |
| 1–7 · fondation, CoA, dimensions, GL | Pertahankan double-entry dan immutability posted. Koreksi scope dan pemetaan efektif dahulu (FA-01/02/05). EF menjadi pemilik entitas/lokasi, IAM pemilik grant. |
| 8–17 · pembelian, persediaan, produksi vendor | Adopsi perhitungan GR/IR dan selisih harga berdasar pemakaian aktual. Pisahkan custody dari kepemilikan persediaan. Aktifkan WIP eksternal ketika ada fakta vendor dan QC (FA-04/06/12). |
| 18–23 dan 28 · penjualan, AR, refund dan pajak | Adopsi AR, penerimaan dana, pendapatan saat pemenuhan dan rekonsiliasi provider. Tanggal pengakuan dan mapping pajak mengikuti kebijakan perusahaan yang disetujui. Tarif contoh desain tidak digunakan sebagai default (FA-02/03/10). |
| 24–27 · payroll dan koreksi | HRM menghitung. Finance memvalidasi, memublikasikan, lalu mengonfirmasi kas dari settlement. Pisahkan publikasi slip dari pelunasan (FA-09). |
| 29–33 · treasury, expense, aset dan budget | Prioritaskan bank dan alur aset (FA-08/11). Untuk expense, gunakan bukti dan approver independen yang sudah diwajibkan instruksi 8. Pertahankan reserve/commit/release budget internal serta scope EF. Buka operasi UI setelah alurnya terbukti. |
| 34–39 · CRM, DRM, tutup buku, laporan, dashboard | Berikan CRM biaya aktual dan pendapatan posted dalam read model terscope saat ada kebutuhan atribusi. DRM menyimpan bukti versi immutable. Utamakan close readiness, neraca dan arus kas sebelum membangun dashboard turunan (FA-07/09). |
| 40–48 · organisasi, event, keamanan, data dan urutan kerja | Gunakan server Finance dan kontrak yang ada. Tabel baru diusulkan hanya untuk fakta yang belum bisa disimpan dengan constraint yang cukup. Prioritas pertama ialah akses, akurasi akun dan bukti payroll. |

## Temuan dan implikasi

### FA-01 · P0 · Batasi laporan pada entitas dan lokasi yang diizinkan

Gateway menerima grant reporting pada salah satu scope. Service meneruskan filter lokasi opsional ke kueri agregat. Ringkasan pajak hanya menerima rentang tanggal. Pemeriksaan izin umum tidak mengikat baris dan agregat ke lokasi milik grant.

**Hasil yang harus terlihat.** Pengguna fasilitas A melihat saldo A pada filter kosong dan tidak bisa membaca B. Pengguna global mendapatkan total yang dapat direkonsiliasi.

**Lokasi bukti.** `backend/gateway/internal/adapter/finance_accounting/permissions.go:49` · `backend/gateway/internal/adapter/finance_accounting/reporting_adapter.go:15` · `backend/services/finance_accounting/internal/security/policy.go:189` · `backend/services/finance_accounting/internal/service/reporting/service.go:64` · `backend/services/finance_accounting/repository/accounting/journal_repository.go:134` · `backend/services/finance_accounting/repository/trade/tax_repository.go`

### FA-02 · P0 · Pisahkan pajak keluaran, pembulatan dan akun kontrol piutang

PostARInvoice mencari tax rate dengan ID dokumen dan mengabaikan hasil serta error. Pajak positif dikreditkan ke RoundingAccountID. Akun kontrol piutang berasal dari request. Dampaknya, jurnal dapat seimbang dengan saldo kewajiban pajak yang salah.

**Hasil yang harus terlihat.** Saldo per akun, outstanding pelanggan dan ringkasan pajak cocok dengan invoice. Pembulatan dan pajak dapat diaudit terpisah.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/accounts_receivable/service.go:233` · `backend/services/finance_accounting/transport/grpc/handler/ar_handler.go:102` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto` · `backend/services/finance_accounting/internal/service/accounting/service.go:545`

### FA-03 · P1 · Hormati kontrak pembuatan invoice AR

Proto menerima period_id, currency_code, exchange_rate dan idempotency_key. Mapper tidak meneruskan empat field ini. Error parsing angka diabaikan dan total dikunci pada dua desimal.

**Hasil yang harus terlihat.** Handler, service dan record menggunakan nilai ekonomi yang sama. Satu key menghasilkan satu invoice termasuk sesudah timeout.

**Lokasi bukti.** `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto:39` · `backend/services/finance_accounting/transport/grpc/handler/ar_handler.go:54` · `backend/services/finance_accounting/internal/service/accounts_receivable/service.go:114`

### FA-04 · P1 · Samakan nilai matching AP dan jurnal persediaan

Pembuatan invoice menghitung nilai neto setelah diskon. Matching menghitung kuantitas × harga bruto, mengisi variance persediaan nol, lalu mengalokasikan seluruh variance ke HPP. Posting mendebit nilai neto ke akun baris tanpa menjabarkan clearing dan variance dari match.

**Hasil yang harus terlihat.** GR/IR terurai per penerimaan, AP sesuai invoice neto, nilai sisa persediaan dan HPP sesuai konsumsi.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/accounts_payable/service.go:195` · `backend/services/finance_accounting/internal/service/accounts_payable/service.go:614` · `backend/services/finance_accounting/internal/service/accounts_payable/service.go:836` · `backend/services/finance_accounting/internal/integration/procurement/client.go` · `backend/services/finance_accounting/internal/integration/warehouse/client.go`

### FA-05 · P1 · Simpan kedua sisi aturan posting

Pada varian umum Gateway mengirim akun debit, kredit, deskripsi dan status aktif. Handler hanya membentuk aturan debit lalu merespons is_active=true. Varian payroll dan finance memiliki pemetaan tersendiri.

**Hasil yang harus terlihat.** Baca kembali debit, kredit, tanggal dan status dari penyimpanan. Respons sukses sama dengan aturan yang akan dipilih saat posting.

**Lokasi bukti.** `backend/gateway/internal/adapter/finance_accounting/accounting_adapter.go:299` · `backend/services/finance_accounting/transport/grpc/handler/accounting_handler.go:245` · `backend/services/finance_accounting/internal/service/accounting/service.go:545` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounting.proto`

### FA-06 · P1 · Cegah saldo awal menimpa valuasi dan tentukan biaya issue

Dua transaksi awal dapat sama-sama melihat saldo kosong. Upsert saldo nol memakai ON CONFLICT UPDATE termasuk kuantitas, nilai dan sequence. Unique sequence dapat menggagalkan transaksi kedua. Handler menghitung delta dari unit cost pengirim juga untuk issue.

**Hasil yang harus terlihat.** Dua receipt pertama terserialisasi tanpa kehilangan nilai. Issue/HPP konsisten dengan buku, perpindahan custody menjaga nilai total.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/inventory_cost/service.go:83` · `backend/services/finance_accounting/repository/inventory_cost/inventory_cost_repository.go:107` · `backend/services/finance_accounting/transport/grpc/handler/inventory_cost_handler.go:49` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/inventory_cost.proto`

### FA-07 · P1 · Tutup periode dengan bukti rekonsiliasi

Penutupan periode mengecek jurnal draft/pending/approved. CreateJournalDraft memilih periode tanpa lock yang dipakai CloseAccountingPeriod. Pada interleaving tertentu draft dapat memasuki periode yang baru ditutup. Reporting menyediakan tiga RPC.

**Hasil yang harus terlihat.** Setelah close tidak ada draft baru untuk periode itu. Aset = liabilitas + ekuitas dan perubahan kas sesuai laporan arus kas.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/accounting/service.go:272` · `backend/services/finance_accounting/internal/service/accounting/service.go:617` · `backend/services/finance_accounting/repository/accounting/period_repository.go:151` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/reporting.proto` · `backend/services/finance_accounting/internal/service/reporting/service.go:96`

### FA-08 · P1 · Validasi mutasi bank dan selesaikan rekonsiliasi

Mapper memilih debit dengan membandingkan string kosong/0/0.00. Nilai nol dengan skala lain bisa menutupi kredit, dan dua sisi positif tidak ditolak di mapper. Service memeriksa persamaan saldo. Tabel reconciliation dan matches ada, tetapi proto Treasury berhenti di import untuk rekonsiliasi.

**Hasil yang harus terlihat.** Rekonsiliasi dapat ditelusuri dari statement ke jurnal, fee dan saldo. Satu jadwal yang sudah jatuh tempo menghasilkan satu kewajiban transfer.

**Lokasi bukti.** `backend/services/finance_accounting/transport/grpc/handler/treasury_handler.go:153` · `backend/services/finance_accounting/internal/service/treasury/service.go:687` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto` · `backend/services/finance_accounting/db/migrations/000001_create_finance_accounting_schema.up.sql:696` · `backend/services/finance_accounting/internal/service/treasury/payout_service.go:33` · `backend/services/finance_accounting/infrastructure/third_party/xendit/payout.go`

### FA-09 · P0 · Teruskan bukti publikasi payroll sampai slip tersimpan

HRM menerima run dengan no-journal posting, lalu pada salinan yang ditinjau membaca pointer journal tanpa syarat saat verifikasi publikasi. Finance menerbitkan payroll.published dengan journal_entry_id saja. Verifier publikasi dan lookup tidak mengikat bukti no-journal seperti verifier posting/payment. Risiko mencakup kegagalan penerbitan slip payroll yang sah.

**Hasil yang harus terlihat.** Slip dibuat setelah publikasi, termasuk run tanpa jurnal. Kehilangan PDF dapat dipulihkan dari edisi terverifikasi. Status nol tidak menampilkan transfer uang.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/payroll/service.go:607` · `backend/services/finance_accounting/internal/service/payroll/service.go:891` · `backend/services/finance_accounting/internal/service/payroll/service.go:1062` · `backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go:311` · `backend/services/finance_accounting/internal/service/treasury/payroll_allocation.go` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto` · [backend/services/human_resource/internal/service/payroll/payslip_publication.go:149](../../../backend/services/human_resource/internal/service/payroll/payslip_publication.go#L149) · [backend/services/human_resource/internal/service/payroll/payslip_reissue.go:79](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L79) · `backend/services/human_resource/infrastructure/integration/finance/payroll_verifier.go`

### FA-10 · P1 · Hubungkan dana pelanggan dan refund per fulfillment

Client Commerce selalu mengembalikan false pada IsPrepaidContractReady. Kalkulasi pajak, pembuatan sesi pembayaran dan refund mengembalikan dependency unavailable. Kontrak Finance yang diperiksa belum memuat lifecycle dana pelanggan ini.

**Hasil yang harus terlihat.** Split shipment hanya mengonsumsi bagian didanai, refund dibatasi saldo aktual dan pendapatan dapat ditelusuri ke pemenuhan. Finance memberi status terverifikasi ke Commerce.

**Lokasi bukti.** `backend/services/web_commerce/infrastructure/integration/finance/client.go:49` · `backend/services/web_commerce/services/checkout/service.go` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/accounts_receivable.proto` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto` · `backend/services/finance_accounting/infrastructure/inbox/processor.go` · `backend/services/finance_accounting/infrastructure/outbox/publisher.go`

### FA-11 · P2 · Lengkapi lifecycle aset sebelum membuka depresiasi

RegisterFixedAsset menyimpan status draft. PostDepreciation mensyaratkan in_service dengan cost basis dari transaksi. Kontrak hanya memuat register dan depreciation sehingga jalur normal tidak dapat mengaktifkan aset baru.

**Hasil yang harus terlihat.** Aset baru dapat melewati draft → in service → depreciation. Jadwal dan GL dapat direkonsiliasi per aset.

**Lokasi bukti.** `backend/services/finance_accounting/internal/service/fixed_assets/service.go:103` · `backend/services/finance_accounting/internal/service/fixed_assets/service.go:135` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/fixed_assets.proto`

### FA-12 · P2 · Hitung biaya produksi yang dititipkan ke vendor

Procurement dan WMS mempunyai kontrak produksi vendor. Dua RPC inventory cost Finance belum memaparkan biaya per batch, tahap dan hasil lulus QC. Ini kebutuhan desain untuk produksi eksternal.

**Hasil yang harus terlihat.** Saldo WIP awal + tambahan − hasil − kerugian = saldo akhir per batch, tanpa menduplikasi stok atau master produk.

**Lokasi bukti.** `backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/external_production_vendor_control.proto` · `backend/services/inventory_warehouse/contract/proto/inventory_warehouse/v1/production_reporting_integration.proto` · `backend/services/product_lifecycle/contract/proto/product_lifecycle/v1/variant.proto` · `backend/services/finance_accounting/contract/proto/finance_accounting/v1/inventory_cost.proto` · `backend/services/finance_accounting/internal/service/inventory_cost/service.go:63`

## Bukti yang masih perlu dikumpulkan

- Rekonsiliasi sampel jurnal AR posted dengan invoice dan mapping efektif perusahaan. Laporkan periode dan akun terdampak sebelum mengajukan jurnal koreksi.
- Bandingkan saldo receipt yang masih ada, telah terjual dan berada di vendor dengan GR/IR serta alokasi selisih per invoice. Kuantifikasi dampak setelah memperoleh data sah.
- Verifikasi konkurensi close dan inisialisasi inventory pada transaksi DB aktual melalui layer service. Periksa efek yang benar-benar commit dan kegagalan yang di-rollback.
- Uji HTTPS, IAM, certificate workload, status sandbox provider dan pembuatan PDF DRM saat implementasi. Audit ini tidak menjalankan pembayaran, E2E atau mengulas source test.

Bacaan domain awal ada di `agents/logic/finance_accounting/logic.md`. Koordinasi khusus ada pada [handoff untuk HRM](../../../agents/notes/human_resource/1-finance_accounting.md) dan [handoff untuk Commerce](../../../agents/notes/web_commerce/2-finance_accounting.md).
