# Uji browser rencana tenaga kerja sesuai badan hukum requisition

Uji Chrome menggunakan HTTPS pada stack terisolasi. Superadmin masuk dengan MFA dan memasang role serta izin melalui GraphQL IAM. Masing-masing dari dua badan hukum memiliki unit di Enterprise Foundation dan rencana yang disetujui di HRM.

| Kasus | Hasil |
| --- | --- |
| Pilih unit pertama | Rencana badan hukum pertama tersedia. Rencana badan hukum kedua tidak ditampilkan. |
| Pindah ke unit kedua | Rencana badan hukum kedua tersedia. Rencana pertama tidak lagi ditawarkan. |
| Kembali ke unit pertama, pilih lini, simpan, dan ajukan | Rencana serta lini yang sesuai tersimpan dan tampil kembali di detail. |
| Periksa aturan sebelumnya | Lokasi, badan hukum hiring manager, jabatan, dan kategori tetap mengikuti aturan form. |

TypeScript, ESLint, dan build frontend lulus.
