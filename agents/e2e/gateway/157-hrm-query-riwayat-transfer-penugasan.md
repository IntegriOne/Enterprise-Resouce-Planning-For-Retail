# Uji HTTPS riwayat transfer penugasan

Superadmin memberi izin baca melalui GraphQL IAM. Dua transfer dari pekerja aktif dengan tautan IAM dan posisi pada unit EF disiapkan pada database HRM tes.

| Kasus | Hasil |
| --- | --- |
| Halaman pertama menggunakan GID pekerja | Transfer terbaru muncul dan total dua |
| Halaman kedua menggunakan UUID pekerja dan cursor | Transfer sebelumnya muncul dan tidak ada halaman lanjut |
| Isi edge dibandingkan dengan node | ID dan cursor akhir konsisten |
| GID transfer dipakai sebagai worker ID | Ditolak dengan `INVALID_ARGUMENT` |

Uji gRPC PostgreSQL juga memeriksa total, transfer hasil command submit, dan halaman kosong sesudahnya. Go test dan Go vet Gateway lulus.

## Metadata transfer

Uji HTTPS tambahan membaca `reason_code`, `requested_at`, dan `decided_at` dari transfer berstatus `submitted`. Sesudah fixture PostgreSQL tes mencatat penolakan, alasan dan waktu pengajuan tetap sama serta waktu keputusan muncul. Uji Gateway dengan race detector memeriksa metadata pada hasil command submit dan approve.
