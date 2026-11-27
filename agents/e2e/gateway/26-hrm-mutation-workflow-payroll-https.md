# Verifikasi CSRF workflow payroll di Gateway lokal

Kedua replika Gateway dibangun ulang dan sehat. Jalur HTTPS Caddy diuji dengan CA development terverifikasi. Setiap permintaan memakai dokumen GraphQL valid dan Origin yang diizinkan tanpa header CSRF.

| Mutasi | Gateway 1 | Gateway 2 | Caddy HTTPS |
| --- | --- | --- | --- |
| `hrmCreatePayrollRun` | FORBIDDEN | FORBIDDEN | FORBIDDEN |
| `hrmCalculatePayrollRun` | FORBIDDEN | FORBIDDEN | FORBIDDEN |
| `hrmSubmitPayrollRunForReview` | FORBIDDEN | FORBIDDEN | FORBIDDEN |
| `hrmApprovePayrollRun` | FORBIDDEN | FORBIDDEN | FORBIDDEN |
| `hrmCancelPayrollRun` | FORBIDDEN | FORBIDDEN | FORBIDDEN |
| `hrmRequestPayrollPosting` | FORBIDDEN | FORBIDDEN | FORBIDDEN |

Seluruh 18 respons menyebut header CSRF yang hilang. Katalog IAM lokal masih berisi nol izin HRM. Uji positif dengan grant resmi, kalkulator runtime dan publikasi Finance masih terbuka.
