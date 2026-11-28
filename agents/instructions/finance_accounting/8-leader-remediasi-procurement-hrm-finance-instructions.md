# Instruksi Finance & Accounting untuk remediasi Procurement, HRM dan Finance

- **id** — 8
- **owner** — Agent Finance & Accounting, reviewer leader lintas management.
- **priority** — Tinggi. Dahulukan kerahasiaan payroll, integritas pembayaran dan validasi owner.
- **status** — Siap dikerjakan.
- **baseline** — Working tree `main` berbasis `da64877287ad`, audit kode produksi 28 November 2026 WIB.

## Task

Implementasikan kontrol payroll dan payout, lindungi event privat, lalu publikasikan operasi yang siap bersama Gateway. Gunakan [laporan PHF-01 sampai PHF-08](../../../agents/audit/finance_accounting/9-2026-11-28-213747-audit-kode-inti-procurement-hrm-finance.md) untuk bukti dan [matriks operasi](../../../agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md) untuk daftar field yang harus ditelusuri.

## Request

### PHF-08 dan PHF-07 · Kerahasiaan outbox dan asal bukti slip

- file: [backend/services/finance_accounting/cmd/wiring.go](../../../backend/services/finance_accounting/cmd/wiring.go#L308) · [backend/services/finance_accounting/infrastructure/outbox/hrm_payroll_publisher.go](../../../backend/services/finance_accounting/infrastructure/outbox/hrm_payroll_publisher.go#L98) · [backend/services/finance_accounting/infrastructure/outbox/publisher.go](../../../backend/services/finance_accounting/infrastructure/outbox/publisher.go#L100) · [backend/services/finance_accounting/internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L482) · [backend/services/finance_accounting/internal/service/payroll/publisher_principal.go](../../../backend/services/finance_accounting/internal/service/payroll/publisher_principal.go#L14) · [backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go](../../../backend/services/finance_accounting/transport/grpc/handler/payroll_handler.go#L112)
- function: procurementInvoiceRoute.Publish, HRMPayrollPublisher.Publish, PublishPayroll dan resolvePayrollPublisherPrincipal
- logic: Tentukan tujuan dari kelas event sebelum memilih transport. Izinkan hanya event payroll yang mempunyai penerima privat. Untuk event slip item, simpan sebagai audit internal atau kirim ke consumer privat yang mendukung kontraknya. Pisahkan bentuk request run dan item, wajibkan verifikasi DRM terhadap versi, digest, owner dan result. Pakai principal terpercaya untuk manusia maupun workload. Tangani baris lama menurut bukti yang ada dan catat jika atribusi historis tidak dapat ditentukan.
- error_handling: Karantina event privat tanpa tujuan. Timeout tujuan privat tetap pending untuk retry. Tolak request ambigu atau bukti bertentangan sebelum menulis state.
- security: Allowlist tujuan payroll dan AP di setiap konfigurasi. Hanya metadata minimum dalam log. Hanya identitas workload tepercaya boleh mencatat receipt item otomatis.
- expected_result: Semua konfigurasi dengan publisher umum tetap menahan event payroll privat dari fallback. Actor tersimpan dapat dirunut ke principal asli.
- testing: Verifikasi service dan routing untuk kombinasi HRM/Procurement hidup atau mati dengan general publisher, jenis payroll dikenal/baru, bukti milik result lain dan replay.

### PHF-01 · Alokasi payroll dan settlement

- file: [backend/services/finance_accounting/db/schema/finance_accounting.dbml](../../../backend/services/finance_accounting/db/schema/finance_accounting.dbml#L1138) · [backend/services/finance_accounting/repository/payroll/payroll_repository.go](../../../backend/services/finance_accounting/repository/payroll/payroll_repository.go#L306) · [backend/services/finance_accounting/internal/service/treasury/service.go](../../../backend/services/finance_accounting/internal/service/treasury/service.go#L348) · [backend/services/finance_accounting/internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L979) · [backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto#L10) · [backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/payroll.proto#L10) · [backend/gateway/graphql/finance_accounting/finance_payroll.graphql](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L16)
- function: PostPayment, InsertPayrollPaymentAllocation, SelectPostedPaymentProof, CompletePayrollSettlement
- logic: Tambahkan alokasi item payroll pada boundary Treasury. Kunci payment dan item dalam urutan tetap. Pastikan posted/confirmed sesuai tahap, pekerja, legal entity, currency, arah dan kind konsisten. Pertahankan UNIQUE(payment_id): satu pembayaran payroll untuk satu item; satu item dapat menerima beberapa pembayaran parsial. Batasi jumlah terhadap payment dan sisa item. Simpan jurnal, alokasi dan efek outbox dalam transaksi posting. Sediakan rekonsiliasi pembayaran payroll yang sudah posted tanpa membuat jurnal kedua. Tambahkan Query detail batch dan connection item agar petugas memperoleh ID dari UI.
- error_handling: Replay tuple dan nominal sama mengembalikan hasil lama. Tolak nominal negatif, kelebihan alokasi, bukti belum sah, replay berbeda dan settlement yang belum lunas. Pertahankan semantik item nol serta recovery.
- security: Aksi bayar/post/settle hanya Finance, pemisahan pembuat dan penyetuju, MFA dan scope payroll. HRM hanya mengirim snapshot dan menerima fakta pembayaran.
- expected_result: Pembayaran parsial belum menyelesaikan item. Pembayaran yang tepat dapat menyelesaikan batch sekali dan menyampaikan result_paid/settled ke HRM. Pembayaran pemasok tetap kompatibel.
- testing: Uji nominal nol, positif, recovery, parsial, dua posting bersamaan, pembayaran historis dan timeout setelah commit. Gunakan jalur Treasury dari aplikasi dan bukti bank/provider resmi dalam lingkungan demo.

### PHF-02 · Reimbursement

- file: [backend/services/finance_accounting/cmd/wiring.go](../../../backend/services/finance_accounting/cmd/wiring.go#L498) · [backend/services/finance_accounting/internal/service/expenses/service.go](../../../backend/services/finance_accounting/internal/service/expenses/service.go#L93) · [backend/services/finance_accounting/domain/service/ports.go](../../../backend/services/finance_accounting/domain/service/ports.go#L127)
- function: CreateExpenseClaim, ApproveExpenseClaim dan port fakta pekerja HRM
- logic: Pasang verifier workload Finance→HRM untuk worker, entitas dan relasi penerima terhadap principal dari sesi; gunakan tanggal bukti yang sesuai aturan klaim. Bedakan input atas nama sendiri dengan klaim yang dicatat petugas. Pada approval independen periksa pembuat dan penerima walau petugas lain mencatat klaim. Tambahkan detail/connection klaim scoped untuk workflow empat Mutation.
- error_handling: Jika settings/fakta HRM wajib gagal dibaca, hentikan command. Bedakan unavailable, identitas tidak cocok, dokumen tidak ditemukan dan konflik status. Simpan keputusan hanya bila semua syarat terpenuhi.
- security: Service owner memeriksa permission, scope entity, pemisahan tugas serta fakta beneficiary. PII dan rincian bank di luar kontrak lookup.
- expected_result: Petugas sah dapat mengajukan untuk pekerja yang sah. Penerima dengan ID principal berbeda dari ID worker tetap ditolak sebagai approver bila independensi diwajibkan.
- testing: Kasus positif pekerja sah, penerima yang merangkap approver, entitas berlainan, verifier mati, settings gagal dibaca, serta approval oleh petugas independen.

### PHF-03 · Policy payout pemasok

- file: [backend/services/finance_accounting/domain/entity/treasury.go](../../../backend/services/finance_accounting/domain/entity/treasury.go#L147) · [backend/services/finance_accounting/internal/service/treasury/payout_service.go](../../../backend/services/finance_accounting/internal/service/treasury/payout_service.go#L33) · [backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/treasury.proto#L10)
- function: ScheduleSupplierPayout, ApprovePayoutSchedule, dispatchScheduledPayout
- logic: Turunkan jadwal monthly dari versi policy, day_of_month, cutoff lokal, zona, bulan pendek dan kalender hari kerja yang disepakati. Untuk explicit_date validasi input. Definisikan kelompok plafon dari policy dan occurrence. Reservasi secara atomik lintas invoice, lepaskan sesuai status final, serta cek ulang outstanding/kredit, beneficiary, funding dan kelayakan saat dispatch. Sediakan detail/connection schedule dan daftar policy yang dapat dipilih.
- error_handling: Calendar atau policy tidak lengkap menolak penjadwalan untuk occurrence bersangkutan. Respons provider ambigu masuk rekonsiliasi memakai idempotency key yang sama.
- security: Grant terpisah untuk pembuat dan approver. Demo Xendit melalui adapter sandbox. Dispatch milik worker privat.
- expected_result: Dua invoice dalam satu occurrence tidak dapat melewati max_batch_amount. Waktu efektif sesuai kebijakan perusahaan dan dapat diterangkan di UI.
- testing: Periksa hari ke-31, bulan pendek, hari libur, cutoff, dua reservasi simultan, kredit sebelum dispatch, timeout provider dan replay settlement.

### PHF-05/06 · Fakta ke Procurement dan reservasi tenaga kerja

- file: [backend/services/procurement_sourcing/internal/service/integration/input.go](../../../backend/services/procurement_sourcing/internal/service/integration/input.go#L40) · [backend/services/finance_accounting/repository/budget/hr_authorization.go](../../../backend/services/finance_accounting/repository/budget/hr_authorization.go#L178) · [backend/services/finance_accounting/contract/proto/finance_accounting/v1/budget.proto](../../../backend/services/finance_accounting/contract/proto/finance_accounting/v1/budget.proto#L1) · [backend/gateway/internal/adapter/finance_accounting/budget_adapter.go](../../../backend/gateway/internal/adapter/finance_accounting/budget_adapter.go#L164) · [agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md](../../../agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md#L1)
- function: Event posting/alokasi/kredit/reversal; CommitHRBudgetReservation dan ReleaseHRBudgetReservation
- logic: Sediakan tanggal pembukuan asli, effect ID, line, currency, nominal/delta dan versi dalam kontrak dengan Procurement. Beri HRM cara privat memastikan hasil intent commit budget, termasuk jika balasan hilang; gunakan subject, digest, entity dan authorization ID. Sepakati kompatibilitas producer/consumer serta pemrosesan ulang data lama.
- error_handling: Efek serupa yang isinya berbeda menghasilkan conflict; timeout privat dapat diulang. Rekonsiliasi menilai status final owner sebelum pelepasan dana.
- security: Pisahkan audience Procurement dan HRM. Commit/release privat membutuhkan workload masing-masing dan scope; issue/revoke manusia sudah ada.
- expected_result: Laporan periodik bisa direkonsiliasi ke Finance dan status anggaran kedua owner konvergen setelah retry.
- testing: Cek invoice satu periode dan pelunasan periode berikutnya, kredit/koreksi, replay/reorder serta Finance commit sukses sebelum HRM restart.

### Publikasi katalog Finance

- file: [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1) · [backend/gateway/internal/adapter/finance_accounting/treasury_adapter.go](../../../backend/gateway/internal/adapter/finance_accounting/treasury_adapter.go#L1) · [backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L1) · [backend/gateway/graphql/finance_accounting/finance_payroll.graphql](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L16)
- function: 14 calon field yang dirinci per RPC dalam matriks
- logic: Selesaikan reimbursement dan payout dengan transaksi dan Query pendukung dahulu. Untuk AR, inventory, fixed assets dan bank statement, periksa fakta owner dan aturan posting pada setiap RPC, sertakan prasyarat dan izin per field di katalog. Sinkronkan 5 field tambahan yang sudah ada. Serahkan kontrak implementabel ke Gateway dan dampingi sampai terhubung.
- error_handling: Bila dependensi domain belum tersedia, catat nama owner dan prasyarat spesifik per operasi. Kegagalan integrasi menghentikan perubahan finansial.
- security: Gunakan grant per aksi, scope objek dan pembatasan data rekening. Validasi posting period dan dokumen sumber di Finance.
- expected_result: Setiap baris punya field, status, owner, validasi serta bukti atau dependensi konkret. Selisih nama tidak menghasilkan resolver ganda.
- testing: Jalankan service scenario per domain, kemudian uji HTTPS setiap operasi yang diaktifkan beserta penolakan pada state dan scope yang salah.

## Action

1. Cocokkan setiap simbol dan status wiring pada checkout terbaru. Tautkan temuan ke perubahan yang sudah masuk agar pekerjaan agent lain dapat dipakai. Kerjakan sampai seluruh kriteria penerimaan dalam scope terpenuhi.
2. Mulai dari invariant DBML/migrasi dan kontrak event/proto, kemudian repository, transaksi service, handler/policy dan wiring startup. Jalankan generator resmi sebelum wiring SDL → resolver → adapter → client. Untuk perubahan lintas owner, sepakati bentuk payload, scope, versi, replay, serta urutan rollout consumer/producer.
3. Pertahankan kompatibilitas data lama dan pemanggil AP, payroll dan provisioning. Perbarui inventaris di `agents/grpc/{management}/` (untuk HRM gunakan `agents/grpc/hrm/`), beserta penjelasan operasi, Query/Mutation/privat, RBAC, validasi, state, error dan sumber fakta.
4. Untuk perubahan Go, lakukan fmt, vet, build dan pemeriksaan service yang relevan. Unit test server gRPC ditempatkan pada lapisan service sesuai AGENTS.md. Gunakan HTTPS `/api/graphql` untuk verifikasi publik, login superadmin dan atur role, permission, scope serta assignment melalui GraphQL IAM. Gunakan principal pekerja/petugas yang sesuai untuk skenario SELF/SoD. Pengujian ini adalah kriteria penyerahan implementasi berikutnya.
5. Lampirkan per ID temuan lokasi file/fungsi, operasi yang dipublikasikan, bukti hasil positif/penolakan, migrasi data lama dan pemulihan dari retry. Catat hambatan owner spesifik bila ada, perbarui katalog dan serahkan ke leader untuk audit ulang.

## Context

[Aturan repo](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Konteks Finance & Accounting](../../../agents/context/finance_accounting/1-context.md) · [how_to_code gRPC](../../../agents/how_to_code_grpc_managements.md) · [how_to_code Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Template instruksi](../../../agents/instruction_template.md)

[agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1) · [agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md#L30) · [agents/logic/procurement/logic.md](../../../agents/logic/procurement/logic.md#L945) · [agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md](../../../agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md#L1) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1) · [agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md](../../../agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md#L1)

[Instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/6-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi Human Resource](../../../agents/instructions/hrm/10-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi GraphQL Gateway](../../../agents/instructions/gateway/6-leader-remediasi-procurement-hrm-finance-instructions.md)
