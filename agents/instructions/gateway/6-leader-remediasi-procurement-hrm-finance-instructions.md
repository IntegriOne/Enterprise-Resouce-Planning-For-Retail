# Instruksi GraphQL Gateway untuk remediasi Procurement, HRM dan Finance

- **id** — 6
- **owner** — Agent GraphQL Gateway, reviewer leader lintas management.
- **priority** — Tinggi. Dahulukan kerahasiaan payroll, integritas pembayaran dan validasi owner.
- **status** — Siap dikerjakan.
- **baseline** — Working tree `main` berbasis `da64877287ad`, audit kode produksi 28 November 2026 WIB.

## Task

Tuntaskan wiring operasi manusia dari service yang lolos kriteria penerimaan sampai /api/graphql. Gunakan [laporan PHF-01 sampai PHF-08](../../../agents/audit/finance_accounting/9-2026-11-28-213747-audit-kode-inti-procurement-hrm-finance.md) untuk bukti dan [matriks operasi](../../../agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md) untuk daftar field yang harus ditelusuri.

## Request

### Publikasi Procurement dan penelusuran ke owner

- file: [backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql](../../../backend/gateway/graphql/procurement_sourcing/procurement_sourcing_core.graphql#L366) · [backend/gateway/internal/adapter/procurement_sourcing/core_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/core_adapter.go#L280) · [backend/gateway/internal/adapter/procurement_sourcing/external_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/external_adapter.go#L250) · [backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go#L1) · [agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go#L88)
- function: Field Mutation untuk 11 gap Procurement pada matriks
- logic: Tambahkan field internal dulu, hubungkan ke RPC aktual. Untuk evaluasi RFQ tunggu proto yang telah terimplementasi. Pertahankan ejaan field dan alias yang sudah digunakan. Draft dan field opsional harus membedakan absent, clear dan zero. Validasi tipe ID parent/line dan decode GID internal. Ikuti pola SDL Procurement, IAM serta EF yang ada.
- error_handling: Kembalikan validation atau workflow conflict dari owner sesuai MutationResult. Jaga konsistensi respons retry.
- security: Sesi opaque, CSRF, scope entity/partner dan MFA saat diperlukan; cek sebelum adapter dan ulang di service owner.
- expected_result: Seluruh gap terpetakan ke owner, resolver dan adapter. UI menggunakan kode/nama dan status serta dapat menuntaskan siklus RFQ.
- testing: HTTPS per command dari scope benar dan salah, parent tidak cocok, stale state, termasuk award setelah evaluasi.

### Publikasi Finance dan layar kerja payroll

- file: [backend/gateway/internal/adapter/finance_accounting/treasury_adapter.go](../../../backend/gateway/internal/adapter/finance_accounting/treasury_adapter.go#L1) · [backend/gateway/graphql/finance_accounting/finance_payroll.graphql](../../../backend/gateway/graphql/finance_accounting/finance_payroll.graphql#L16) · [backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql](../../../backend/gateway/graphql/finance_accounting/finance_accounts_payable.graphql#L1) · [backend/gateway/internal/adapter/finance_accounting/budget_adapter.go](../../../backend/gateway/internal/adapter/finance_accounting/budget_adapter.go#L164) · [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1)
- function: postPayment, completePayrollSettlement, 14 gap Finance pada matriks, Query batch/item/claim/payout
- logic: Koordinasikan readiness per RPC. Prioritaskan alokasi payroll, expense dan payout; pertahankan kompatibilitas AP. Beri UI daftar/detail dengan Relay connection, edges/nodes, total terotorisasi dan detail berisi nomor bisnis, nama, status, jumlah serta tanggal. Katalog menyebut semua 14 calon operasi dan 5 tambahan di SDL. Tampilkan state pending/reconcile dari owner saat berlaku.
- error_handling: Jangan konfirmasi pembayaran saat provider baru menerima instruksi. Ubah kesalahan owner menjadi error stabil, hindari mengembalikan detail rekening dan raw gRPC.
- security: Permission terpisah untuk create/approve/post, MFA, akses per badan hukum dan redaksi detail pekerja. Gateway hanya antarmuka manusia; alur worker privat tetap mTLS.
- expected_result: Operator menemukan ID batch/item dari UI dan dapat menelusuri lunas berdasarkan fakta Finance. Dependency yang belum siap tercatat per field.
- testing: HTTPS atas transaksi benar, over-allocation, receiver/self approval, jadwal melanggar policy, node/list di entity lain dan replay.

### HRM, IAM, anggaran dan slip

- file: [backend/gateway/graphql/human_resource/hrm_payslip.graphql](../../../backend/gateway/graphql/human_resource/hrm_payslip.graphql#L31) · [backend/gateway/graph/resolvers/finance_budget.resolvers.go](../../../backend/gateway/graph/resolvers/finance_budget.resolvers.go#L45) · [backend/gateway/internal/adapter/finance_accounting/budget_adapter.go](../../../backend/gateway/internal/adapter/finance_accounting/budget_adapter.go#L164) · [agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md#L30) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1)
- function: issueBudgetAuthorization, revokeBudgetAuthorization, myPayslip, myPayslipPdf, myPayslips, reissueEmployeePayslip dan createEmployeeUser
- logic: Gunakan operasi budget yang sudah ada dan snapshot HRM otoritatif. Pertahankan perbedaan jam terbit dan bayar pada tampilan. Sinkronkan status intent dari HRM lewat resource existing. Penautan worker ke principal tetap pada provisioning IAM. Rekonsiliasi internal memakai gRPC privat. Perbarui matriks per management dengan lokasi SDL–resolver–adapter–RPC dan status tiap fase.
- error_handling: Tangani owner unavailable dan status pending secara eksplisit. Pemeriksaan akses tetap berlaku untuk pagination, node dan unduhan PDF.
- security: Mantan pegawai memakai bantuan petugas HR berizin. Jangan menampilkan GID sebagai label UI atau memasukkan konten PDF, ID sensitif dan data rekening dalam log.
- expected_result: Jumlah field terhitung dari SDL dan setiap operasi terkait mempunyai alur serta pemilik jelas. Jalur HR tidak memberi akses transfer.
- testing: HTTPS self aktif, mantan pegawai, HR lintas entity ditolak, grant unit budget, dan worker provisioning IAM.

## Action

1. Cocokkan setiap simbol dan status wiring pada checkout terbaru. Tautkan temuan ke perubahan yang sudah masuk agar pekerjaan agent lain dapat dipakai. Kerjakan sampai seluruh kriteria penerimaan dalam scope terpenuhi.
2. Mulai dari invariant DBML/migrasi dan kontrak event/proto, kemudian repository, transaksi service, handler/policy dan wiring startup. Jalankan generator resmi sebelum wiring SDL → resolver → adapter → client. Untuk perubahan lintas owner, sepakati bentuk payload, scope, versi, replay, serta urutan rollout consumer/producer.
3. Pertahankan kompatibilitas data lama dan pemanggil AP, payroll dan provisioning. Perbarui inventaris di `agents/grpc/{management}/` (untuk HRM gunakan `agents/grpc/hrm/`), beserta penjelasan operasi, Query/Mutation/privat, RBAC, validasi, state, error dan sumber fakta.
4. Untuk perubahan Go, lakukan fmt, vet, build dan pemeriksaan service yang relevan. Unit test server gRPC ditempatkan pada lapisan service sesuai AGENTS.md. Gunakan HTTPS `/api/graphql` untuk verifikasi publik, login superadmin dan atur role, permission, scope serta assignment melalui GraphQL IAM. Gunakan principal pekerja/petugas yang sesuai untuk skenario SELF/SoD. Pengujian ini adalah kriteria penyerahan implementasi berikutnya.
5. Lampirkan per ID temuan lokasi file/fungsi, operasi yang dipublikasikan, bukti hasil positif/penolakan, migrasi data lama dan pemulihan dari retry. Catat hambatan owner spesifik bila ada, perbarui katalog dan serahkan ke leader untuk audit ulang.

## Context

[Aturan repo](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Konteks GraphQL Gateway](../../../agents/context/gateway/1-context.md) · [how_to_code gRPC](../../../agents/how_to_code_grpc_managements.md) · [how_to_code Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Template instruksi](../../../agents/instruction_template.md)

[agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1) · [agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md#L30) · [agents/logic/procurement/logic.md](../../../agents/logic/procurement/logic.md#L945) · [agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md](../../../agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md#L1) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1) · [agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md](../../../agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md#L1)

[Instruksi Finance & Accounting](../../../agents/instructions/finance_accounting/8-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi Procurement & Sourcing](../../../agents/instructions/procurement_sourcing/6-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi Human Resource](../../../agents/instructions/hrm/10-leader-remediasi-procurement-hrm-finance-instructions.md)

Baca SDL acuan sebelum menulis field. [Schema bersama](../../../backend/gateway/graphql/schema.graphql) · [IAM](../../../backend/gateway/graphql/iam/iam_user.graphql) · [Enterprise Foundation](../../../backend/gateway/graphql/enterprise_foundation/ef_legal_entity.graphql)
