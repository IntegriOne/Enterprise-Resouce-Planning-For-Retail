# Uji HTTPS provisioning akun kandidat publikasi

Binary HRM, IAM, Enterprise Foundation, dan Gateway dibangun dari salinan kandidat dengan `GOWORK=off`. Runner menjalankannya di kontainer dengan database disposable dan Redis terpisah. Superadmin masuk melalui HTTPS dan memasang role serta izin melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Grant dengan target badan hukum yang tidak ada | IAM menolak. Setelah badan hukum dibuat di EF, grant dapat dipasang dan dicabut melalui GraphQL. |
| Pengajuan sebelum event pekerja dan aktivasi sebelum MFA atau kontrak | Ditolak sesuai prasyarat. Aktivasi dengan fixture kontrak berhasil. |
| Pengajuan akun diulang | Antrean IAM memuat satu receipt. |
| Worker dan receipt tidak cocok | Provisioning ditolak. |
| `createEmployeeUser` dan pengulangan request | IAM mengembalikan akun yang sama. IAM dan HRM menyimpan principal yang cocok. |
| Baca ulang pengajuan | Receipt keluar dari antrean, HRM mengembalikan status applied, dan refetch Relay berhasil. |
| Audit fixture dan pembersihan | Satu akun dan satu receipt tertaut. Role uji dicabut melalui GraphQL, kontainer dan database disposable dibersihkan. |

Tes verifikasi provisioning Go dengan tag `e2e` juga lulus. [Handoff publikasi](../../grpc/hrm/32-handoff-integrasi-main-gateway.md).

Runner di repositori kini mendukung `--source-root`. Eksekusi dengan root kandidat dan eksekusi bawaan dari checkout lokal sama-sama lulus. Root yang tidak memiliki modul wajib ditolak sebelum stack disiapkan.
