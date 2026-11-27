# Pengujian Mutation penutupan hak cuti

| Kasus Leave gRPC fixture | Hasil |
| --- | --- |
| Izin close, MFA, dan CSRF | Waktu penutupan |
| Izin close tanpa MFA | FORBIDDEN |
| Izin baca hak cuti saja | FORBIDDEN |

HRM memeriksa akhir periode, saldo nol dan pembukuan koreksi. Jalur HTTPS dengan akun resmi masih terbuka.
