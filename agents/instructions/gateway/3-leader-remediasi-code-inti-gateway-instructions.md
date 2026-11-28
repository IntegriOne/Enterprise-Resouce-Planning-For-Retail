# Remediasi kode inti Gateway

**Status:** Siap dikerjakan. **Pemilik:** Gateway. **Temuan:** KI-12.

## Task

Kirim konteks principal/scope terpercaya ke owner dan tegakkan hasil keputusan pada daftar, node, nested dan pembacaan byte. Pertahankan satu endpoint HTTPS GraphQL serta provisioning employee melalui IAM/HRM. Kerjakan implementasi milik Gateway; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [internal/adapter/document_record/ingestion_pub.go](../../../backend/gateway/internal/adapter/document_record/ingestion_pub.go#L234).
- function: `DocumentContentChunk`, `AuthorizeDocumentRecipient`, dan `HasScopedPermission` pada snapshot sesi.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Grant global baca tidak cukup untuk menyimpulkan izin atas byte di semua badan usaha. Teruskan identitas actor dan tujuan; gunakan satu keputusan DRM untuk semua entrypoint.
- expected_result: Izin baca global tanpa mandat badan usaha dokumen, PIC party sah pada daftar dan byte, serta reissue HR sesuai worker/result. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-12 · Kontrol akses dokumen berbeda antara daftar dan isi.** Definisikan satu keputusan akses per version, owner, entity, tujuan, principal dan klasifikasi. Gunakan keputusan itu untuk metadata, daftar, byte dan reissue. Untuk party, evaluasi PIC aktif di EF; untuk workload, ikat identitas pemanggil pada tujuan dan resource. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
