# Pengujian Query detail grup dan aturan payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Detail grup dengan izin baca grup | Kode grup dan status aktif |
| Detail aturan dengan izin baca aturan | Versi dan waktu persetujuan kosong untuk aturan belum disetujui |
| Detail aturan dengan izin baca grup saja | FORBIDDEN |

Tes detail dan regresi daftar lulus dengan race detector. Uji HTTPS dengan grant resmi masih terbuka.
