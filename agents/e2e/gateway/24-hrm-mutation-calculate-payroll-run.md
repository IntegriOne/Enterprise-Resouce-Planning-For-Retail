# Pengujian integrasi kalkulasi payroll

| Kasus GraphQL dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Izin calculate, MFA dan CSRF dengan adapter tersedia | Status calculated dan waktu kalkulasi |
| Adapter gRPC mengembalikan unavailable | UNAVAILABLE |
| Izin approve tanpa izin calculate | FORBIDDEN |

Tiga kasus lulus dengan race detector. Runtime HRM memerlukan kalkulator yang memverifikasi snapshot input dan artefak aturan privat sebelum operasi dapat berhasil.
