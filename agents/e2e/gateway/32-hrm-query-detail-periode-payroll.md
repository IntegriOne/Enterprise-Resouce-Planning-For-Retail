# Pengujian Query detail periode payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| UUID tersedia dengan izin global baca periode | Kode dan status open |
| UUID yang tidak ditemukan | NOT_FOUND |
| ID bukan UUID | INVALID_ARGUMENT |

Kasus detail dan regresi daftar lulus dengan race detector. Uji akun resmi melalui HTTPS masih terbuka.
