# Remediasi kode inti HRM

**Status:** Siap dikerjakan. **Pemilik:** HRM. **Temuan:** KI-09.

## Task

Pertahankan human_resource sebagai owner PDF dan payroll_result sebagai sumber. Berikan Finance identity result, entity, exact version dan digest untuk bukti publikasi item. Pastikan error saat Finance/DRM tidak tersedia dapat dipulihkan per result. Kerjakan implementasi milik HRM; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [infrastructure/integration/document_records/drm_client.go](../../../backend/services/human_resource/infrastructure/integration/document_records/drm_client.go#L170) · [internal/service/payroll/payslip_reissue.go](../../../backend/services/human_resource/internal/service/payroll/payslip_reissue.go#L27).
- function: `StorePrivatePayslip`, `ReadPrivatePayslip`, `RestorePrivatePayslip`, `ReissueEmployeePayslip`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Pekerja aktif hanya membaca result miliknya. HR berwenang mencetak ulang untuk mantan pegawai pada worker/result yang sah. Simpan audit tanpa byte PDF atau nominal.
- expected_result: Run Finance terbit, dua pegawai dua dokumen, percobaan ulang pada result yang sama; pegawai aktif mencetak sendiri dan permintaan mantan pegawai diproses oleh HR. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-09 · Finance meminta pemilik dokumen yang keliru untuk slip.** Gunakan verifier dengan pemilik HRM, result ID, tipe, entity dan digest. Bedakan kegagalan kontrak dari gangguan jaringan. Finance mengesahkan run; HRM membuat PDF; DRM memastikan versinya; Finance mencatat publikasi item. Pertahankan akses pegawai aktif dan jalur HR untuk mantan pegawai. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
