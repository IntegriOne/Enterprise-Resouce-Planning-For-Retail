# Pengujian Query daftar grup payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca grup, limit 1, offset 1 | Satu grup aktif, total 2 dan offset 1 |
| Izin baca periode saja | FORBIDDEN |

Kedua kasus lulus dalam `TestHRMPayrollReferenceQueriesThroughGateway` dengan race detector. Pengujian HTTPS dengan grant resmi masih terbuka.
