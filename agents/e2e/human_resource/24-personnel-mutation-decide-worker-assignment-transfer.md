# Keputusan transfer penempatan primer

Tes memakai gRPC privat dan PostgreSQL sementara dengan policy serta verifier Foundation tiruan.

| Kasus | Hasil |
| --- | --- |
| Pengaju menyetujui permintaan sendiri | PermissionDenied |
| Penyetuju mencoba sebelum tanggal efektif | FailedPrecondition |
| Kapasitas tujuan tidak cukup | FailedPrecondition, penempatan sumber tetap terbuka |
| Persetujuan setelah kapasitas dipulihkan | Penempatan sumber berakhir H−1, pengganti menjadi primer, satu lifecycle dan satu outbox |
| Retry identik oleh penyetuju yang sama | Penempatan pengganti yang sama dikembalikan tanpa event tambahan |
| Permintaan lama ditolak oleh principal independen | Status rejected tanpa perubahan penempatan primer |

Jika penempatan utama sudah berubah sejak usulan transfer diajukan, approval lama ditolak dengan `FailedPrecondition`. Penyetuju independen tetap dapat menolak usulan lama tanpa mengubah penempatan saat ini. Kasus ini lulus dengan PostgreSQL dan race detector.

Pengajuan dan approval transfer memakai aturan yang sama untuk posisi tujuan. Saat approval, posisi juga harus masih berlaku pada hari penerapan. Tes menolak posisi yang dinonaktifkan setelah pengajuan.

Tes juga menolak approval ketika posisi tujuan berlaku pada tanggal efektif yang diajukan tetapi sudah berakhir pada hari penerapan. Pengujian dilakukan melalui gRPC, PostgreSQL sementara, dan race detector.
