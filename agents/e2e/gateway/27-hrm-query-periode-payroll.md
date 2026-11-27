# Pengujian Query daftar periode payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca periode, limit 1, offset 1 | Satu periode open, total 2 dan offset 1 |
| Izin baca run saja | FORBIDDEN |

Kedua kasus lulus dengan race detector. Pengujian HTTPS dengan grant resmi masih terbuka.
