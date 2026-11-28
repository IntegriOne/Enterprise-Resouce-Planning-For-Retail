# Remediasi kode inti DRM

**Status:** Siap dikerjakan. **Pemilik:** DRM. **Temuan:** KI-09, KI-11, KI-12, KI-13, KI-15.

## Task

Benahi delivery revoke/disposition yang saat ini sebatas log. Terapkan satu keputusan akses owner di jalur daftar dan stream. Pisahkan klaim job storage dari request approval dengan cleanup/recovery dan korelasikan exact version payslip. Kerjakan implementasi milik DRM; perubahan pada service pasangan dikerjakan oleh pemiliknya dengan kontrak yang sama. [Laporan audit kode inti](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md) · [Matriks pemanggil dan penerima](../../../agents/audit/gateway/5-2026-11-28-200600-matriks-hubungan-code-inti.md).

## Request

- file: [services/document/service.go](../../../backend/services/document_record/services/document/service.go#L386) · [cmd/main.go](../../../backend/services/document_record/cmd/main.go#L87) · [services/publication/service.go](../../../backend/services/document_record/services/publication/service.go#L445) · [services/records/service.go](../../../backend/services/document_record/services/records/service.go#L695) · [repository/publication/repo.go](../../../backend/services/document_record/repository/publication/repo.go#L182) · [infrastructure/integration/foundation/client.go](../../../backend/services/document_record/infrastructure/integration/foundation/client.go#L130) · [infrastructure/scanner/engine.go](../../../backend/services/document_record/infrastructure/scanner/engine.go#L66) · [cmd/wiring.go](../../../backend/services/document_record/cmd/wiring.go#L40).
- function: `VerifyDocumentVersion`, `ListMyPublications`, `StreamDocument`, `ApproveDisposition`, `runOutboxWorker`.
- logic: Ikuti urutan per temuan pada bagian Action. Gunakan fakta dari owner pada versi transaksi yang sama dan perlakukan status pending, diterima, diterapkan, dan ditolak sesuai tahapan bisnis.
- error_handling: Ketika jawaban owner belum tersedia, pertahankan status yang dapat dipulihkan. Konflik identitas, versi atau digest memerlukan keputusan eksplisit. Retry dengan identitas yang sama wajib mempertahankan satu efek bisnis.
- security: Verifikasi actor termasuk workload, tujuan, owner dan klasifikasi sebelum menyerahkan byte. Proteksi hold, fence dan digest harus bertahan melewati restart.
- expected_result: Restart setelah fence atau setelah delete, slip payroll dari dua result berbeda, party invitation dicabut, dokumen AP/PO restricted serta satu tujuan event yang belum memberi receipt. Perubahan status dan saldo akhir harus dapat dijelaskan dari record bisnis pada service pemilik.
- testing: Gunakan skenario di atas sebagai kriteria penerimaan setelah perbaikan. Untuk uji HTTPS, setup role/permission melalui GraphQL sesuai AGENTS.md.

## Action

1. **KI-09 · Finance meminta pemilik dokumen yang keliru untuk slip.** Gunakan verifier dengan pemilik HRM, result ID, tipe, entity dan digest. Bedakan kegagalan kontrak dari gangguan jaringan. Finance mengesahkan run; HRM membuat PDF; DRM memastikan versinya; Finance mencatat publikasi item. Pertahankan akses pegawai aktif dan jalur HR untuk mantan pegawai. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
2. **KI-11 · DRM menandai event terkirim ketika hanya menulis log.** Daftarkan tujuan sesuai owner dan tipe event. Simpan tanda terima inbox durable per tujuan, jadwalkan ulang pending dan tangani replay. Rekonsiliasi baris historis dengan tanda terima penerima sebelum backfill. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
3. **KI-12 · Kontrol akses dokumen berbeda antara daftar dan isi.** Definisikan satu keputusan akses per version, owner, entity, tujuan, principal dan klasifikasi. Gunakan keputusan itu untuk metadata, daftar, byte dan reissue. Untuk party, evaluasi PIC aktif di EF; untuk workload, ikat identitas pemanggil pada tujuan dan resource. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
4. **KI-13 · Disposisi DRM masih dapat tersangkut setelah restart.** Jalankan worker storage yang dapat mengambil kembali job menurut lease, recheck hold/clearance, periksa generasi objek persis dan simpan bukti delete sebelum tombstone. Pertahankan blok akses/restore selama hasil belum diketahui. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)
5. **KI-15 · Bukti clean DRM bergantung pada pemeriksaan berkas lokal.** Simpan identitas engine, versi kebijakan dan cakupan format bersama digest. Untuk format atau status scan yang belum dapat dipastikan, tahan seal/publikasi sesuai aturan kelas dokumen. Rekonsiliasi hasil saat engine kembali tersedia. [Bukti dan lokasi kedua sisi](../../../agents/audit/gateway/4-2026-11-28-200600-audit-code-inti-hubungan-management.md)

Pada handoff, cantumkan perubahan fungsi, dampak pada data sebelumnya, status rekonsiliasi dan SHA. Jika ada kontrak RPC/SDL yang berubah, sinkronkan inventaris di `agents/grpc/`. Tandai instruksi selesai setelah hasil pada owner penerima sesuai dengan keputusan pada pengirim.

## Context

[Aturan repositori](../../../AGENTS.md) · [Arsitektur](../../../agents/arsitektur.md) · [Kepemilikan management](../../../agents/boundary.md) · [Identitas dan akses](../../../agents/users.md) · [Panduan gRPC](../../../agents/how_to_code_grpc_managements.md) · [Panduan wiring Gateway](../../../agents/how_to_code_wiring_grpc_managements_to_gateway.md).
