# Pengujian Query detail koreksi cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| ID ada dan izin baca koreksi | Referensi hak cuti, delta presisi dan kode alasan |
| ID tidak ditemukan | NOT_FOUND |
| Izin membuat draf saja | FORBIDDEN |

Tes integrasi dengan race detector mencakup detail dan regresi daftar. Akses positif dengan grant resmi melalui HTTPS masih terbuka.
