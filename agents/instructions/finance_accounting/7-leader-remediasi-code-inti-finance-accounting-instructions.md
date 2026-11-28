# Remediasi kode inti Finance

**Status:** Siap dikerjakan. **Pemilik:** Finance. **Temuan:** KI-02, KI-05, KI-09, KI-10.

## Task

Perbaiki kontrak nilai dan pemilihan event lebih dahulu. Teruskan expected bindings dokumen dari pemanggil payroll ke client DRM. Beda hasil verified=false dengan dependency unavailable harus tercermin dalam error bisnis. Kerjakan implementasi milik Finance; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [domain/entity/enums.go](../../../backend/services/finance_accounting/domain/entity/enums.go#L263) · [internal/service/inventory_cost/service.go](../../../backend/services/finance_accounting/internal/service/inventory_cost/service.go#L63) · [transport/grpc/handler/inventory_cost_handler.go](../../../backend/services/finance_accounting/transport/grpc/handler/inventory_cost_handler.go#L34) · [domain/entity/decimal.go](../../../backend/services/finance_accounting/domain/entity/decimal.go#L39) · [internal/service/payroll/service.go](../../../backend/services/finance_accounting/internal/service/payroll/service.go#L497) · [internal/integration/drm/client.go](../../../backend/services/finance_accounting/internal/integration/drm/client.go#L73) · [cmd/wiring.go](../../../backend/services/finance_accounting/cmd/wiring.go#L359) · [repository/integration/integration_repository.go](../../../backend/services/finance_accounting/repository/integration/integration_repository.go#L220) · [internal/integration/procurement/client.go](../../../backend/services/finance_accounting/internal/integration/procurement/client.go#L69).
- function: `ApplyInventoryCostEvent`, `PublishPayroll`, `VerifyDocumentVersionBound`, `buildOutboxRelay`, `SelectPendingBatchForUpdate`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Ikat workload pengirim, owner dokumen, entity dan source record. Hanya Finance dapat melakukan pencatatan pembayaran; payload payroll privat harus menuju HRM.
- expected_result: Unit cost dengan delapan desimal, receipt dan issue bertanda berlawanan; relay standby dengan kedua tujuan privat aktif; PDF HRM sah yang terkait payroll result persis. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-02 · Satu event mencampur data beberapa efek stok.** Bentuk event per efek bisnis dengan relasi baris yang sama, key stabil pada replay, delta bertanda, kategori Finance, dan sumber basis valuasi yang disepakati. Catat penerimaan per tujuan. Pertahankan custody tanpa perubahan kepemilikan sebagai efek fisik. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
2. **KI-05 · Valuasi Finance menerima perhitungan yang kehilangan presisi.** Validasi jenis mutasi, kuantitas dan harga satuan sebelum perkalian. Gunakan skala buku atau kebijakan mata uang yang disahkan saat pembulatan, simpan tanda delta dan alasan reversal, jangan komit jumlah tanpa nilai yang semestinya. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
3. **KI-09 · Finance meminta pemilik dokumen yang keliru untuk slip.** Gunakan verifier dengan pemilik HRM, result ID, tipe, entity dan digest. Bedakan kegagalan kontrak dari gangguan jaringan. Finance mengesahkan run; HRM membuat PDF; DRM memastikan versinya; Finance mencatat publikasi item. Pertahankan akses pegawai aktif dan jalur HR untuk mantan pegawai. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
4. **KI-10 · Relay privat Finance dapat melewatkan seluruh event Procurement.** Pilih himpunan event yang dapat diklaim sesudah seluruh tujuan diketahui. Pada konfigurasi privat HRM + Procurement, claim harus meliputi posted invoice, payment allocation/reversal, credit dan payroll yang didukung. Pastikan event payroll privat tetap terisolasi dari webhook umum. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
