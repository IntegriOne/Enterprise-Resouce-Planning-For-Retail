# Pengujian filter grup pada Query periode payroll

| Kasus dengan Payroll gRPC fixture | Hasil |
| --- | --- |
| Grup valid dengan izin baca periode | Periode dari RPC finder dan total dari HRM |
| ID grup tidak valid | INVALID_ARGUMENT |
| RPC mengembalikan periode grup lain | UNAVAILABLE |

Filter dan regresi daftar tanpa filter diuji dengan race detector dalam `TestHRMPayrollReferenceQueriesThroughGateway`.
