# Instruksi Procurement & Sourcing untuk remediasi Procurement, HRM dan Finance

- **id** — 6
- **owner** — Agent Procurement & Sourcing, reviewer leader lintas management.
- **priority** — Tinggi. Dahulukan kerahasiaan payroll, integritas pembayaran dan validasi owner.
- **status** — Siap dikerjakan.
- **baseline** — Working tree `main` berbasis `da64877287ad`, audit kode produksi 28 November 2026 WIB.

## Task

Buka transisi evaluasi RFQ, hubungkan command yang hilang dan koreksi laporan lintas periode. Gunakan [laporan PHF-01 sampai PHF-08](../../../agents/audit/finance_accounting/9-2026-11-28-213747-audit-kode-inti-procurement-hrm-finance.md) untuk bukti dan [matriks operasi](../../../agents/grpc/finance_accounting/6-matriks-operasi-procurement-hrm-finance.md) untuk daftar field yang harus ditelusuri.

## Request

### PHF-04 · Evaluasi dan penetapan pemenang RFQ

- file: [agents/logic/procurement/logic.md](../../../agents/logic/procurement/logic.md#L945) · [backend/services/procurement_sourcing/internal/service/sourcing/service.go](../../../backend/services/procurement_sourcing/internal/service/sourcing/service.go#L203) · [backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/sourcing.proto](../../../backend/services/procurement_sourcing/contract/proto/procurement_sourcing/v1/sourcing.proto#L1) · [backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go#L88) · [backend/gateway/internal/adapter/procurement_sourcing/core_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/core_adapter.go#L280)
- function: BeginRFQEvaluation, PublishRFQ, penetapan award dan CloseRFQ
- logic: Tambahkan RPC untuk transisi pada service yang sudah ada. Validasi status, scope entitas, versi serta aturan batas pengajuan yang berlaku. Kunci aggregate yang sama dengan submit quotation agar keputusan konsisten. Sambungkan ke adapter, Mutation dan izin sourcing mengikuti katalog owner.
- error_handling: Tolak state/tanggal tidak memenuhi prasyarat; bedakan replay keputusan sama dengan keputusan konflik. Rollback perubahan jika audit/outbox gagal.
- security: Pemisahan kewenangan evaluator, vendor, dan pemberi award sesuai mandat. Tolak input actor dari request.
- expected_result: RFQ yang sah dapat dibuka, menerima quotation, masuk evaluation lalu diaward atau ditutup melalui aplikasi.
- testing: Verifikasi alur positif, close dari open, dua transisi bersamaan dan quotation saat batas waktu; HTTPS setelah wiring.

### Sepuluh command Procurement lain

- file: [agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [backend/services/procurement_sourcing/transport/grpc/handler/configuration_partner_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/configuration_partner_handler.go#L293) · [backend/services/procurement_sourcing/transport/grpc/handler/requisition_approval_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/requisition_approval_handler.go#L67) · [backend/services/procurement_sourcing/transport/grpc/handler/purchase_order_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/purchase_order_handler.go#L138) · [backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/sourcing_handler.go#L88) · [backend/services/procurement_sourcing/transport/grpc/handler/external_production_handler.go](../../../backend/services/procurement_sourcing/transport/grpc/handler/external_production_handler.go#L222) · [backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/orders_adapter.go#L1) · [backend/gateway/internal/adapter/procurement_sourcing/external_adapter.go](../../../backend/gateway/internal/adapter/procurement_sourcing/external_adapter.go#L250)
- function: RevisePartnerOffer, UpdatePurchaseRequestDraft, UpdateRFQDraft, UpdatePurchaseOrderRevisionDraft, UpdateProductionBatchDraft, UpdateVendorWorkOrderDraft, ReplaceVendorWorkOrder, SubmitVendorMaterialReconciliation, CreateVendorQualityAction, ResolveVendorQualityAction
- logic: Telusuri handler existing sampai transaksi. Pada draft, izinkan field yang boleh diubah dan pertahankan beda absent/null. Pada revisi periksa versi efektif dan cegah overlap. Penggantian WO menautkan work order asal dan pengganti. Rekonsiliasi menautkan material ke line dan owner custody. Quality action memeriksa keputusan, bukti, kuantitas tersisa dan remedy. Serahkan seluruh 10 ke Gateway sesuai fase internal/vendor. Secara terpisah nilai kebutuhan koreksi draft quotation yang RPC-nya tersedia.
- error_handling: Stale revision, scope induk salah, status final, bukti tidak sah dan kuantitas berlebih gagal sebelum efek samping. Replay aman untuk command berulang.
- security: Pisahkan akses petugas dan vendor menurut objek yang ditugaskan. Revisi harga serta tindakan kualitas berjaminan sesuai izin transaksi.
- expected_result: Seluruh 11 gap termasuk evaluasi RFQ ditelusuri ke service dan UI. Nama ekuivalen yang telah ada dipertahankan.
- testing: Kasus positif dan negatif untuk tiap command, lintas induk/entity, versi lama, retry dan pemanggilan dari sesi yang hanya memiliki izin baca.

### PHF-05 · Laporan spending periodik dan status PO

- file: [backend/services/procurement_sourcing/repository/integration/integration_repository.go](../../../backend/services/procurement_sourcing/repository/integration/integration_repository.go#L417) · [backend/services/procurement_sourcing/internal/service/integration/service.go](../../../backend/services/procurement_sourcing/internal/service/integration/service.go#L670) · [backend/services/procurement_sourcing/internal/service/integration/input.go](../../../backend/services/procurement_sourcing/internal/service/integration/input.go#L40) · [backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml](../../../backend/services/procurement_sourcing/db/schema/procurement_sourcing.dbml#L1) · [agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md](../../../agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md#L1)
- function: SelectActualSpendingReport, GetPurchaseOrderFinancialStatus dan consumer Finance
- logic: Pertahankan snapshot untuk saldo terkini. Catat efek bertanggal dan berkunci unik untuk aliran per periode, termasuk invoiced, allocated, credited dan reversed. Ikuti tanggal Finance, pisahkan mata uang atau pakai satu mata uang tervalidasi. Definisikan batas waktu from/to. Tangani data lama dengan replay fakta asli; identifikasi periode yang belum lengkap. Bedakan revisi memang belum terbit dari galat database dan propagasikan galat parse nominal.
- error_handling: Inbox dedup menurut sumber dan effect/version. Efek hilang atau ada selisih masuk rekonsiliasi. Kembalikan error atas kegagalan baca, hindari menampilkan total nol yang tampak sah.
- security: Workload Finance terautentikasi. Verifikasi invoice/PO/line/entity sebelum penerapan. Scope laporan dan total harus identik.
- expected_result: Invoice November tetap ada di November setelah pembayaran Desember. Kredit dan reversal tampak pada periode pembukuannya. Nilai lintas periode dapat ditelusuri ke Finance.
- testing: Periode berurutan, invoice lintas beberapa line, payment parsial, reversal/kredit, replay ganda/tidak urut, revisi belum ada serta galat baca.

## Action

1. Cocokkan setiap simbol dan status wiring pada checkout terbaru. Tautkan temuan ke perubahan yang sudah masuk agar pekerjaan agent lain dapat dipakai. Kerjakan sampai seluruh kriteria penerimaan dalam scope terpenuhi.
2. Mulai dari invariant DBML/migrasi dan kontrak event/proto, kemudian repository, transaksi service, handler/policy dan wiring startup. Jalankan generator resmi sebelum wiring SDL → resolver → adapter → client. Untuk perubahan lintas owner, sepakati bentuk payload, scope, versi, replay, serta urutan rollout consumer/producer.
3. Pertahankan kompatibilitas data lama dan pemanggil AP, payroll dan provisioning. Perbarui inventaris di `agents/grpc/{management}/` (untuk HRM gunakan `agents/grpc/hrm/`), beserta penjelasan operasi, Query/Mutation/privat, RBAC, validasi, state, error dan sumber fakta.
4. Untuk perubahan Go, lakukan fmt, vet, build dan pemeriksaan service yang relevan. Unit test server gRPC ditempatkan pada lapisan service sesuai AGENTS.md. Gunakan HTTPS `/api/graphql` untuk verifikasi publik, login superadmin dan atur role, permission, scope serta assignment melalui GraphQL IAM. Gunakan principal pekerja/petugas yang sesuai untuk skenario SELF/SoD. Pengujian ini adalah kriteria penyerahan implementasi berikutnya.
5. Lampirkan per ID temuan lokasi file/fungsi, operasi yang dipublikasikan, bukti hasil positif/penolakan, migrasi data lama dan pemulihan dari retry. Catat hambatan owner spesifik bila ada, perbarui katalog dan serahkan ke leader untuk audit ulang.

## Context

[Aturan repo](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Boundary](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Konteks Procurement & Sourcing](../../../agents/context/procurement_sourcing/1-context.md) · [how_to_code gRPC](../../../agents/how_to_code_grpc_managements.md) · [how_to_code Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md) · [Template instruksi](../../../agents/instruction_template.md)

[agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md](../../../agents/grpc/procurement_sourcing/1-inventaris-rancangan-operasi.md#L1) · [agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md](../../../agents/grpc/finance_accounting/4-matriks-query-mutation-dan-publikasi-gateway.md#L1) · [agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md](../../../agents/grpc/hrm/1-inventaris-operasi-grpc-hrm.md#L30) · [agents/logic/procurement/logic.md](../../../agents/logic/procurement/logic.md#L945) · [agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md](../../../agents/grpc/procurement_sourcing/2-kontrak-integrasi-finance.md#L1) · [agents/grpc/hrm/41-penerimaan-event-payroll-finance.md](../../../agents/grpc/hrm/41-penerimaan-event-payroll-finance.md#L1) · [agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md](../../../agents/grpc/finance_accounting/3-integrasi-inbox-outbox.md#L1)

[Instruksi Finance & Accounting](../../../agents/instructions/finance_accounting/8-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi Human Resource](../../../agents/instructions/hrm/10-leader-remediasi-procurement-hrm-finance-instructions.md) · [Instruksi GraphQL Gateway](../../../agents/instructions/gateway/6-leader-remediasi-procurement-hrm-finance-instructions.md)
