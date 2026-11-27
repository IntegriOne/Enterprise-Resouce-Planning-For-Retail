# Uji HTTPS createEmployeeUser wajib memakai pengajuan IAM

Basis `e28491c` dengan perubahan worktree. Gateway dibangun ulang. IAM, HRM, dan EF memakai PostgreSQL `*_test` dan Redis disposable, dengan permission fixture di katalog IAM uji. Role dan permission diberikan serta dicabut lewat GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Mutation dengan Worker ID valid tetapi tanpa `provisioning_request_id` | PASS. GraphQL menolak input. |
| Mutation dengan ID pengajuan dan worker benar sebelum MFA | PASS. IAM Gateway menolak dengan `AUTHORIZATION`. |
| Mutation dengan ID pengajuan dan worker berbeda setelah MFA | PASS. Gateway menolak dengan `CONFLICT`. |
| Mutation dengan pengajuan, worker, dan MFA yang valid | PASS. Akun dibuat dan tertaut. |
| Ulangi Mutation pengajuan yang sama | PASS. Akun yang sama dikembalikan. |
| Periksa PostgreSQL IAM dan Query status HRM | PASS. Satu akun dan pengajuan, status HRM `applied`. |

Aktivasi worker tetap fixture di HRM disposable. Uji ini menggantikan jalur Worker ID manual yang dicatat pada laporan 129. Pemanggil yang masih memakai kontrak lama perlu menyediakan pengajuan IAM.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-required.py`.
