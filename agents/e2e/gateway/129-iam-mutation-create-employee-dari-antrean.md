# Uji HTTPS createEmployeeUser dari antrean provisioning

Basis Git `e28491c` dengan perubahan worktree. IAM dan Gateway dibangun ulang. PostgreSQL IAM, HRM, EF, dan Redis menggunakan lingkungan disposable. IAM uji memuat katalog permission fixture. Penyiapan dan pencabutan akses dilakukan melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Ajukan akun dari HRM dan buka detail antrean setelah MFA | PASS, pengajuan mengacu ke worker yang sama. |
| Kirim `provisioning_request_id` yang valid dengan worker berbeda | PASS, `CONFLICT`. |
| Kirim `provisioning_request_id` dengan worker sesuai pengajuan | PASS, akun IAM `PENDING` dan tautan HRM terkonfirmasi. |
| Ulangi Mutation dari operator yang sama | PASS, akun yang sama dikembalikan. |
| Baca detail IAM dan daftar antrean terbuka | PASS, pengajuan berstatus `linked` dan keluar dari antrean. |
| Periksa PostgreSQL IAM dan HRM | PASS, satu akun fixture dan satu penerimaan tertaut, dengan principal yang sama pada worker. |

Aktivasi pekerja memakai fixture disposable. Argumen pengajuan saat ini opsional untuk kompatibilitas pemanggil lama. Proyeksi hasil pengajuan ke HRM dan integrasi form IAM belum diverifikasi dalam uji ini.

Runner lokal: `python3 /tmp/erp-cross-service.UUJx1n/run-claim.py`.
