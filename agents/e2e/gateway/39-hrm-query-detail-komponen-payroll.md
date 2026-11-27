# Pengujian Query detail komponen payroll

| Kasus Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin baca komponen dengan MFA | Kode komponen dan nominal desimal presisi |
| ID komponen tidak ditemukan | NOT_FOUND |
| Izin baca komponen tanpa MFA | FORBIDDEN |

Kasus detail dan regresi daftar tercakup dalam `TestHRMPayrollResultLinesThroughGateway`. Uji akses positif dengan akun IAM resmi masih terbuka.
