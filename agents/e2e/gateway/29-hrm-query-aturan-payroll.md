# Pengujian Query aturan payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin global baca aturan, limit 1 dan offset 1 | Satu versi aturan, waktu persetujuan dan total 2 |
| Izin baca grup saja | FORBIDDEN |

Kedua kasus lulus dalam `TestHRMPayrollReferenceQueriesThroughGateway` dengan race detector. Uji dengan grant IAM resmi melalui HTTPS masih terbuka.
