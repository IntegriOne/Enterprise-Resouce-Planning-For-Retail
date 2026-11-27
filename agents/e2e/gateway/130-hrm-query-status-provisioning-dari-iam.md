# Uji HTTPS status pengajuan akun dari IAM ke HRM

Basis `e28491c` dengan perubahan worktree. IAM dan Gateway dibangun ulang. HRM, IAM, dan Enterprise Foundation memakai PostgreSQL `*_test`, dengan Redis disposable serta katalog IAM fixture. Superadmin memasang dan mencabut role/permission lewat GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| `hrmRequestEmployeeAccount` dan pengiriman ulang | PASS. Keduanya menghasilkan ID penerimaan IAM yang sama dan status `submitted`. |
| `hrmEmployeeAccountRequest` sebelum pemrosesan IAM | PASS. ID HRM dan IAM sama dengan payload pengajuan, dengan status `submitted`. |
| `createEmployeeUser` dari antrean IAM dengan MFA | PASS. Akun dan worker tertaut, penerimaan IAM `linked`. |
| `hrmEmployeeAccountRequest` sesudah penautan | PASS. Pengajuan yang sama tampil dengan status `applied`. |
| Pemeriksaan PostgreSQL | PASS. Satu akun IAM, satu penerimaan tertaut dan principal yang sama di record worker. |

Status operasional dibaca langsung dari IAM. HRM tetap menjadi pemilik record pengajuan dan pekerja. Aktivasi worker memakai fixture disposable. Database dan container uji dibersihkan setelah eksekusi.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-status.py`.
