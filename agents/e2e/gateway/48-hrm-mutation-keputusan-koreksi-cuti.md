# Pengujian Mutation keputusan koreksi cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Persetujuan dengan izin approve, MFA dan CSRF | Waktu persetujuan |
| Penolakan dengan izin reject, MFA dan CSRF | Kode alasan yang dirapikan |
| Kode alasan kosong | INVALID_ARGUMENT |
| Izin approve dipakai untuk reject | FORBIDDEN |
| Tanpa MFA | FORBIDDEN |
| Tanpa CSRF | FORBIDDEN |

Tes memakai race detector. Policy HRM memeriksa pemisahan pengusul dan pengambil keputusan. Uji dengan grant resmi melalui HTTPS masih terbuka.
