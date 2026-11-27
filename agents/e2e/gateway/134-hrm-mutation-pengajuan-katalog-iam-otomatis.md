# Uji HTTPS pengajuan akun dengan katalog IAM otomatis

Basis Git `e28491c` dengan perubahan worktree. Gateway dan IAM dibangun ulang. IAM, HRM, EF dan Redis memakai stack serta database disposable. Setup dan cleanup role dan permission menggunakan GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| `hrmRequestEmployeeAccount` hanya dengan `worker_id` | PASS. IAM menyediakan referensi kebijakan akun dasar setelah memeriksa izin dan badan hukum EF. |
| Periksa referensi di PostgreSQL HRM dan IAM | PASS. Keduanya menyimpan ID katalog yang sama; referensi IAM cocok dengan kebijakan sistem `iam_selfservice`. |
| Pengiriman ulang | PASS, ID penerimaan sama dan hanya satu baris tersimpan. |
| Proses melalui `createEmployeeUser` | PASS, satu akun tertaut dan Query HRM menunjukkan `applied`. |

Aktivasi pekerja menggunakan fixture dalam database HRM `*_test`. Referensi katalog dipakai untuk audit kebijakan akun dasar. Pengajuan dari HRM tidak memasang role pengguna secara langsung.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-catalog-reference.py`.
