# Remediasi kode inti Procurement

**Status:** Siap dikerjakan. **Pemilik:** Procurement. **Temuan:** KI-03, KI-04, KI-06, KI-10, KI-16.

## Task

Sediakan lookup komitmen SKU dan mapping receipt kanonik untuk WMS. Validasi substitution sesuai keputusan teknis PLM. Pastikan AP sebenar-benarnya diterima melalui jalur Finance yang juga menjalankan payroll. Kerjakan implementasi milik Procurement; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [internal/service/fulfillment/service.go](../../../backend/services/procurement_sourcing/internal/service/fulfillment/service.go#L145) · [internal/service/purchaseorder/service.go](../../../backend/services/procurement_sourcing/internal/service/purchaseorder/service.go#L672) · [repository/purchase_order/purchase_order_repository.go](../../../backend/services/procurement_sourcing/repository/purchase_order/purchase_order_repository.go#L218) · [repository/external_production/external_production_repository.go](../../../backend/services/procurement_sourcing/repository/external_production/external_production_repository.go#L448) · [internal/service/integration/service.go](../../../backend/services/procurement_sourcing/internal/service/integration/service.go#L500) · [internal/service/vendorcontrol/service.go](../../../backend/services/procurement_sourcing/internal/service/vendorcontrol/service.go#L457) · [internal/integration/plm/client.go](../../../backend/services/procurement_sourcing/internal/integration/plm/client.go#L95).
- function: `ApproveMaterialSubstitution`, `ApplyAcceptedReceipt`, `ApplyReceiptReversal`, `ApplyReturnDispatched`, `SelectPurchaseOrders`, `SelectVendorWorkOrders`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Ikat keputusan teknis ke WO, material dan digest. Pertahankan pemisahan pemohon/approver dan kewenangan vendor pada WO yang ditugaskan.
- expected_result: PO dengan SKU pada baris kedua dan lebih dari satu halaman header; WO/material yang berbeda untuk approval yang sama; satu receipt reversal per original link. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-03 · Reversal dan dispatch putus di batas WMS–Procurement.** Simpan pemetaan durable dari effect WMS ke receipt link Procurement atau lakukan lookup melalui source effect. Daftarkan tipe dispatch yang relevan, bedakan target wajib dan efek lokal, pastikan retry tidak menggandakan receipt, reversal, outstanding maupun valuasi. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
2. **KI-04 · Rencana inbound PO memakai otorisasi dari input.** Untuk inbound berbasis PO, dapatkan otorisasi baris dari Procurement sebelum release/accept. Cocokkan entity, SKU, UOM, revisi, status dan saldo yang diizinkan. Bedakan inbound non-PO; bila PO direvisi, rekonsiliasi kuantitas yang sudah diterima. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
3. **KI-06 · Arsip SKU dan ketersediaan memakai hasil daftar yang tidak lengkap.** Sediakan query fakta per SKU pada owner, termasuk scope channel/entity, komitmen aktif dan dimensi custody/transit. Bila memakai pagination, konsumsi seluruh halaman dan detail baris dengan snapshot konsisten. Keadaan tidak diketahui harus menghalangi arsip dan tidak boleh menjadi janji availability. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
4. **KI-10 · Relay privat Finance dapat melewatkan seluruh event Procurement.** Pilih himpunan event yang dapat diklaim sesudah seluruh tujuan diketahui. Pada konfigurasi privat HRM + Procurement, claim harus meliputi posted invoice, payment allocation/reversal, credit dan payroll yang didukung. Pastikan event payroll privat tetap terisolasi dari webhook umum. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
5. **KI-16 · Persetujuan substitusi vendor tidak meminta keputusan teknis PLM.** Bentuk snapshot dari data WO/material di Procurement. Panggil verifier PLM sebelum keputusan final dan pastikan snapshot tetap sama saat commit. Approval tidak sah atau kedaluwarsa menghentikan transisi. Simpan bukti keputusan PLM dan pertahankan SoD komersial. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
