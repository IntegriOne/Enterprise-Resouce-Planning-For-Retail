# Remediasi kode inti Commerce

**Status:** Siap dikerjakan. **Pemilik:** Commerce. **Temuan:** KI-06, KI-07, KI-08, KI-14.

## Task

Jaga status listing pada batas publish. Gabungkan eligibility SKU, inventory dan invalidasi revision/identifier dengan identitas scope yang sama. Pertahankan checkout prepaid gated sampai owner keputusan tersedia. Kerjakan implementasi milik Commerce; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [infrastructure/integration/wms/client.go](../../../backend/services/web_commerce/infrastructure/integration/wms/client.go#L55) · [services/integration/service.go](../../../backend/services/web_commerce/services/integration/service.go#L161) · [infrastructure/outbox/router.go](../../../backend/services/web_commerce/infrastructure/outbox/router.go#L36) · [cmd/wiring.go](../../../backend/services/web_commerce/cmd/wiring.go#L52) · [services/catalog/service.go](../../../backend/services/web_commerce/services/catalog/service.go#L253).
- function: `CreateSalesChannel`, `PublishStorefrontListing`, `GetEligibleAvailability`, `DeliverInboxEvent`, `RelayPendingOutbox`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Ikat produk, badan usaha, channel dan mandat buyer. Kegagalan owner tidak boleh membuka listing atau menghasilkan stok siap jual yang tidak terverifikasi.
- expected_result: Listing draft yang SKU-nya sudah discontinued, sumber PLM unavailable, stok tersedia di gudang yang tidak memenuhi channel, dan perubahan identifier setelah publish. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-06 · Arsip SKU dan ketersediaan memakai hasil daftar yang tidak lengkap.** Sediakan query fakta per SKU pada owner, termasuk scope channel/entity, komitmen aktif dan dimensi custody/transit. Bila memakai pagination, konsumsi seluruh halaman dan detail baris dengan snapshot konsisten. Keadaan tidak diketahui harus menghalangi arsip dan tidak boleh menjadi janji availability. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
2. **KI-07 · Fan-out perubahan produk dapat selesai setelah satu penerima.** Pisahkan status per pasangan event dan penerima. Simpan efek invalidasi SKU/barcode yang sesuai, tandai tipe tak dikenal untuk rekonsiliasi, serta sertakan versi/digest dan identitas item. Event Commerce yang diwajibkan downstream harus memiliki consumer atau tetap tertunda. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
3. **KI-08 · Listing dapat terbit tanpa konfirmasi PLM.** Izinkan draft sesuai aturan authoring. Untuk publish dan republish, wajibkan fakta sellable per SKU/revisi dari PLM; jika sumber tidak tersedia, pertahankan status sebelumnya. Terapkan juga ketika perubahan lifecycle membatalkan listing lama. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

4. **KI-14 · Badan usaha sales channel.** Saat membuat channel aktif, cocokkan entity lokasi EF dengan entity channel dan periksa scope unit bila ada. Pertahankan keadaan sebelum mutasi jika keputusan EF belum dapat diberikan. [CreateSalesChannel](../../../backend/services/web_commerce/services/storefront/service.go#L179).

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
