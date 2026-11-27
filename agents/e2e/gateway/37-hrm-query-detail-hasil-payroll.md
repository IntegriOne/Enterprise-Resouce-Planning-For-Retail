# Pengujian Query detail hasil payroll

| Kasus Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca hasil dengan MFA | Nominal neto tetap desimal presisi |
| ID hasil tidak ditemukan | NOT_FOUND |
| Izin baca hasil tanpa MFA | FORBIDDEN |

Kasus detail dan regresi daftar tercakup dalam `TestHRMPayrollResultsThroughGateway`. Uji HTTPS dengan grant resmi masih terbuka.
