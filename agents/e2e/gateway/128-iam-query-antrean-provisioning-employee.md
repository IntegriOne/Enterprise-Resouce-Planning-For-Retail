# Uji HTTPS antrean provisioning akun pekerja

Basis `e28491c` dengan perubahan worktree. Gateway dan IAM dibangun ulang. PostgreSQL IAM, HRM, dan EF serta Redis menggunakan lingkungan disposable. Katalog IAM uji dimuat melalui overlay. Pemasangan dan pencabutan role/permission memakai GraphQL.

| Operasi dan kasus | Hasil |
| --- | --- |
| `employeeProvisioningRequests(first: 1)` tanpa sesi | PASS, `UNAUTHENTICATED`. |
| `employeeProvisioningRequests` dengan sesi superadmin sebelum MFA | PASS, akses ditolak. |
| `employeeProvisioningRequests(first: 1)` setelah MFA | PASS, total satu, tanpa halaman berikutnya, item berstatus pending. |
| `employeeProvisioningRequest(id: ...)` setelah MFA | PASS, ID penerimaan, Global ID pekerja dan badan hukum sama dengan hasil pengajuan HRM. |
| Pengiriman ulang pengajuan yang sama | PASS, receipt yang sama dan satu baris di PostgreSQL IAM. |

Antrean memakai izin global `iam.users.write`. Hasil uji mencakup daftar dan detail API. Pemilihan di UI dan penggunaan pengajuan dalam `createEmployeeUser` masih perlu implementasi.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-queue.py`.
