# Pengujian Mutation koreksi draf kontrak

| Kasus fixture | Hasil |
| --- | --- |
| Koreksi nomor dan pengosongan dua field opsional | Field mask hanya memuat ketiganya |
| HRM menolak koreksi kontrak bertanda tangan | CONFLICT |
| Nomor kontrak null | INVALID_ARGUMENT |
| Input kosong | INVALID_ARGUMENT |
| Grant baca dipakai untuk koreksi | FORBIDDEN |

Pengujian memakai sesi employee, CSRF, MFA, dan fixture gRPC.

## HTTPS lokal

Kedua replika sehat pada image `sha256:43b02d4252b40d25c19e07edf9e2d524964521cb1f4019087c07b76a26db1971`. Permintaan koreksi draf dengan ID fixture tanpa CSRF menghasilkan `FORBIDDEN` dari guard Gateway melalui Caddy HTTPS. Suite guard pekerja, kontrak, dan posisi juga lulus. Akses positif masih memerlukan fixture IAM resmi.
