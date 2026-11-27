# Pengujian pencarian draf payroll setelah respons terputus

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Kunci idempotensi draf yang ada | ID dan status draft |
| Kunci yang tidak dikenal | Daftar kosong dan total nol |
| Kunci hanya berisi spasi | INVALID_ARGUMENT |

Tes integrasi dengan race detector lulus. Jalur daftar tanpa filter dan penolakan izin dengan cakupan yang tidak didukung juga lulus.
