# Uji HTTPS provisioning dengan katalog resmi HRM

Basis Git `e28491c` dengan perubahan worktree. IAM dibangun dengan `go build ./cmd` dari source dan katalog resmi HRM yang tertanam. Gateway memakai build terbaru. PostgreSQL IAM, EF dan HRM serta Redis memakai lingkungan disposable. Pengaturan role dan permission dilakukan lewat GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Empat permission provisioning HRM tersedia pada IAM baru | PASS, katalog dimuat saat bootstrap dan izin dapat dicari lewat GraphQL. |
| `hrmRequestEmployeeAccount` dan pengiriman ulang | PASS, ID penerimaan sama dan satu baris di IAM. |
| Query detail dan daftar antrean IAM | PASS, `worker_number` sama dengan nomor yang dikembalikan HRM. |
| `createEmployeeUser` tanpa pengajuan, tanpa MFA, atau worker berbeda | PASS, seluruhnya ditolak. |
| `createEmployeeUser` valid dan pengulangan | PASS, satu akun tertaut dan pengulangan mengembalikan akun yang sama. |
| `hrmEmployeeAccountRequest` sesudah pemrosesan IAM | PASS, status `applied`. |

Aktivasi worker memakai fixture HRM di PostgreSQL `*_test`. Container dan database uji dibersihkan. Uji tingkat API ini belum memeriksa interaksi browser dengan form IAM.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-required.py`.
