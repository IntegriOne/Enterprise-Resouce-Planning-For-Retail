# Uji HTTPS guard GraphQL HRM setelah migrasi Relay

Skrip pemeriksaan menemukan operasi langsung dari SDL HRM dan menambahkan selection set berdasarkan tipe objek. Pengujian memakai Caddy dev dan TLS dengan CA lokal.

| Kasus | Hasil |
| --- | --- |
| Query tanpa sesi | 51 dari 51 ditolak dengan `UNAUTHENTICATED` |
| Mutation tanpa header CSRF | 55 dari 55 ditolak dengan `FORBIDDEN` dan pesan header CSRF hilang |
| Batas laju | Operasi yang tertunda dicoba ulang setelah jendela batas laju |
| Penamaan publik | Pemeriksaan SDL tidak menemukan awalan management pada nama operasi atau tipe |

Hasil ini mencakup guard. Bukti jalur sukses per operasi tercatat pada matriks Query dan Mutation.
