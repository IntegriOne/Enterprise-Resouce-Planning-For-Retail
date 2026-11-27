# Uji HTTPS pengajuan dan penerapan assignment primer

Dua akun fixture memakai TOTP, CSRF, dan role IAM sementara. Posisi diaktifkan lewat Gateway. Job, status aktif pekerja, serta verifikasi identitas akun kedua disiapkan lewat SQL fixture pada database disposable.

| Kasus | Hasil |
| --- | --- |
| `hrmCreateWorkerAssignment` pada posisi aktif | ID dan waktu permintaan tercatat, approval masih kosong |
| Pemohon memanggil `hrmApproveWorkerAssignment` | `FORBIDDEN` |
| Employee kedua menyetujui assignment | ID sama dan waktu approval tercatat |
| `hrmApplyPrimaryAssignment` | Berhasil, pekerja tetap aktif |
| `hrmWorkerAssignment` dan daftar per pekerja | Assignment yang disetujui muncul |
| Pemeriksaan PostgreSQL | Assignment primer cocok dan approver berbeda dari pemohon |

Perbandingan waktu respons memakai toleransi satu mikrodetik untuk presisi penyimpanan PostgreSQL. Pengulangan lulus setelah penyesuaian assert. Database dan Redis uji telah dibersihkan setelah layanan development kembali sehat.
