# Uji browser penutupan requisition

Superadmin masuk dengan MFA lewat HTTPS dan memperoleh role serta permission melalui GraphQL IAM. Chrome memakai binary Gateway, IAM, dan HRM serta frontend pada stack disposable. Requisition berstatus open disiapkan di PostgreSQL HRM disposable.

| Kasus | Hasil |
| --- | --- |
| Buka detail record open dan batalkan konfirmasi | Record tetap Dibuka. |
| Buka ulang konfirmasi dan tutup | Tabel menampilkan status Ditutup. |
| Buka detail yang sudah ditutup | Tombol Tutup requisition hilang. Teks UI tidak memuat `gid://`. |
| Periksa database | Status closed dan satu event close. |

TypeScript, ESLint direktori requisition, dan build frontend lulus.
