# Remediasi kode inti PLM

**Status:** Siap dikerjakan. **Pemilik:** PLM. **Temuan:** KI-06, KI-07, KI-08, KI-14, KI-16.

## Task

Koordinasikan fakta komitmen dengan Procurement dan stok dengan WMS. Hanya arsipkan SKU bila fakta lengkap pada scope keputusan. Catat status tiap consumer, termasuk barcode/identifier. Ikat bukti tech pack ke source revisi sesuai kontrak DRM. Kerjakan implementasi milik PLM; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [services/changecycle/service.go](../../../backend/services/product_lifecycle/services/changecycle/service.go#L387) · [infrastructure/integration/procurement/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/procurement/client.go#L63) · [infrastructure/integration/inventory/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/inventory/client.go#L60) · [infrastructure/outbox/router.go](../../../backend/services/product_lifecycle/infrastructure/outbox/router.go#L42) · [cmd/wiring.go](../../../backend/services/product_lifecycle/cmd/wiring.go#L85) · [infrastructure/integration/commerce/client.go](../../../backend/services/product_lifecycle/infrastructure/integration/commerce/client.go#L91) · [services/revision/service.go](../../../backend/services/product_lifecycle/services/revision/service.go#L672) · [services/masterdata/service.go](../../../backend/services/product_lifecycle/services/masterdata/service.go#L85) · [transport/grpc/handler/revision_structure_change.go](../../../backend/services/product_lifecycle/transport/grpc/handler/revision_structure_change.go#L934).
- function: `ArchiveProduct`, `CheckItemOpenCommitments`, `CheckItemResidualStock`, `ValidateSubstitution`, `Router.Publish`, `LinkControlledDocument`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Verifikasi mandat vendor, WO dan revisi untuk akses tech pack. Tahan lifecycle ketika fakta Procurement/WMS belum pasti.
- expected_result: SKU di luar 100 item pertama, PO header tanpa nested lines, vendor WO dengan beberapa bahan, dan satu penerima revisi offline. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-06 · Arsip SKU dan ketersediaan memakai hasil daftar yang tidak lengkap.** Sediakan query fakta per SKU pada owner, termasuk scope channel/entity, komitmen aktif dan dimensi custody/transit. Bila memakai pagination, konsumsi seluruh halaman dan detail baris dengan snapshot konsisten. Keadaan tidak diketahui harus menghalangi arsip dan tidak boleh menjadi janji availability. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
2. **KI-07 · Fan-out perubahan produk dapat selesai setelah satu penerima.** Pisahkan status per pasangan event dan penerima. Simpan efek invalidasi SKU/barcode yang sesuai, tandai tipe tak dikenal untuk rekonsiliasi, serta sertakan versi/digest dan identitas item. Event Commerce yang diwajibkan downstream harus memiliki consumer atau tetap tertunda. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
3. **KI-08 · Listing dapat terbit tanpa konfirmasi PLM.** Izinkan draft sesuai aturan authoring. Untuk publish dan republish, wajibkan fakta sellable per SKU/revisi dari PLM; jika sumber tidak tersedia, pertahankan status sebelumnya. Terapkan juga ketika perubahan lifecycle membatalkan listing lama. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
4. **KI-14 · Validasi referensi EF bergantung pada wiring opsional.** Bedakan draft lokal dari transisi yang memerlukan fakta EF. Pada transisi itu wajibkan owner benar, entity serta status saat efektif. Kembalikan unavailable saat keputusan belum tersedia; role party global perlu dilengkapi scope dari pemilik transaksi. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
5. **KI-16 · Persetujuan substitusi vendor tidak meminta keputusan teknis PLM.** Bentuk snapshot dari data WO/material di Procurement. Panggil verifier PLM sebelum keputusan final dan pastikan snapshot tetap sama saat commit. Approval tidak sah atau kedaluwarsa menghentikan transisi. Simpan bukti keputusan PLM dan pertahankan SoD komersial. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
