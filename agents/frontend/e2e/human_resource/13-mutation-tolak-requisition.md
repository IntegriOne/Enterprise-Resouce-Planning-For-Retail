# Uji browser penolakan requisition

Pada stack terisolasi, superadmin masuk melalui HTTPS, menuntaskan MFA, dan mengatur role serta permission lewat GraphQL IAM. Superadmin membuat dan mengajukan draf. Akun peninjau terpisah menerima role lewat GraphQL IAM dan masuk dengan MFA untuk menguji pemisahan aktor.

| Kasus | Hasil |
| --- | --- |
| Peninjau membuka requisition yang diajukan | Aksi Tolak tampil. |
| Masukkan kode alasan berisi spasi | Form meminta kode yang valid. |
| Tolak dengan kode `NOT_APPROVED` | Daftar menampilkan status Ditolak. |
| Buka detail hasil | Kode alasan tampil, aksi Tolak hilang, dan teks UI tidak memuat `gid://`. |
| Verifikasi readback | GraphQL mengembalikan status dan kode alasan. HRM mencatat akun peninjau yang berbeda dari pembuat. |

Regresi pembuatan, edit, pengajuan, dan penghapusan draf juga lulus. TypeScript, ESLint, dan build image frontend lulus.
