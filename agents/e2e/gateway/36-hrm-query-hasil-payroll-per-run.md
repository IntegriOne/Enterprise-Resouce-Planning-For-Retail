# Pengujian Query hasil payroll per run

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca hasil dengan MFA | Nominal neto desimal dipertahankan sebagai string |
| Izin baca hasil tanpa MFA | FORBIDDEN |
| Izin baca status run saja | FORBIDDEN |
| Respons gRPC berisi hasil run lain | UNAVAILABLE |

Tes lulus dengan race detector. Akses positif dengan akun resmi melalui HTTPS masih terbuka.
