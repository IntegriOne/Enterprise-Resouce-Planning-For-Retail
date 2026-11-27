# Uji HTTPS pemilihan peristiwa aktivasi saat pengajuan akun

Basis `e28491c` dengan perubahan worktree. HRM dan Gateway dibangun ulang. IAM, HRM, EF dan Redis memakai lingkungan disposable. Role dan permission dipasang dan dicabut melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Worker berstatus aktif tanpa event `hire` | PASS. Mutation mengembalikan `CONFLICT` dan IAM belum menerima pengajuan. |
| Worker yang sama sesudah fixture event `hire` dengan transisi dari preboarding ke aktif | PASS. Pengajuan diterima tanpa argumen ID event dari client. |
| Pengiriman ulang | PASS. Satu penerimaan IAM dengan ID yang sama. |
| Pemrosesan di IAM dengan MFA | PASS. Akun tertaut dan status yang dibaca HRM menjadi `applied`. |

Aktivasi dan event memakai fixture pada database HRM `*_test`. Uji ini memverifikasi aturan pengajuan sesudah aktivasi. Command aktivasi lengkap diuji pada suite personnel tersendiri.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-lifecycle.py`.
