# Pengujian Mutation draf koreksi cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Izin create dan CSRF | Draf dengan delta -0.50 dan waktu keputusan kosong |
| Duplikat dari HRM | CONFLICT |
| Kunci idempotensi kosong | INVALID_ARGUMENT |

Kasus menggunakan race detector. Akses positif melalui HTTPS dengan grant resmi masih terbuka.
