# Pengujian filter periode pada daftar run payroll

| Kasus Payroll gRPC fixture | Hasil |
| --- | --- |
| ID periode valid | Run periode yang dipilih dan total dari HRM |
| Respons gRPC berisi run periode berbeda | UNAVAILABLE |
| Filter periode dan kunci idempotensi dikirim bersamaan | INVALID_ARGUMENT |

Kasus baru dan regresi daftar tercakup dalam `TestHRMPayrollRunsThroughGateway` dengan race detector.
