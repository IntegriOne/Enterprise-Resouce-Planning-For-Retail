# Verifikasi lokal status dan pengajuan jurnal payroll

Kedua replika Gateway dibangun ulang dan berstatus sehat. Pengujian dilakukan melalui port langsung dan Caddy HTTPS dengan CA development terverifikasi.

| Operasi | Gateway 1 | Gateway 2 | Caddy HTTPS |
| --- | --- | --- | --- |
| `hrmPayrollRun` tanpa sesi | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmPayrollRuns` tanpa sesi | UNAUTHENTICATED | UNAUTHENTICATED | UNAUTHENTICATED |
| `hrmRequestPayrollPosting` tanpa CSRF | FORBIDDEN | FORBIDDEN | FORBIDDEN |

Regresi terpisah `TestPayrollRPCAndApprovalWithPostgreSQL` lulus dengan tag `integration`, race detector, database uji khusus dan renderer PDF berfont. Tes gRPC mencakup penolakan sebelum persetujuan, pencatatan pengajuan dengan status tetap approved, penolakan pengajuan kedua, dan event outbox untuk snapshot yang disetujui. Adapter Finance dan DRM dalam pengujian memakai fixture. Akses positif Gateway melalui IAM resmi masih terbuka.
