# Uji HTTPS grant mandat requisition dengan verifikasi mTLS

IAM, Enterprise Foundation, HRM, dan Gateway berjalan pada stack disposable. Superadmin masuk melalui HTTPS. IAM menyinkronkan definisi izin approval dari katalog resmi, lalu role serta grant dipasang melalui GraphQL IAM. Dua unit dibuat melalui GraphQL Enterprise Foundation.

| Kasus | Hasil klien mTLS HRM |
| --- | --- |
| Peninjau hanya memiliki grant GLOBAL | Mandat ditolak. |
| Assign role pada unit pertama | Mandat unit pertama diterima. |
| Periksa unit kedua | Mandat ditolak. |
| Cabut assignment unit pertama lewat GraphQL | Mandat unit pertama kembali ditolak. |

Sertifikat development dibuat oleh generator khusus. CA memvalidasi sertifikat server dan klien. Overlay Compose lolos pemeriksaan konfigurasi. Uji transaksi approve dan open dengan bukti Finance masih terbuka.
