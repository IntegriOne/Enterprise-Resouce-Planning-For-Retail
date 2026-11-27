# Pengujian Mutation penghapusan draf kontrak

| Kasus fixture | Hasil |
| --- | --- |
| Kontrak draf serta grant hapus dengan MFA | true |
| RPC menolak penghapusan kontrak bertanda tangan | CONFLICT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant hapus assignment dipakai untuk kontrak | FORBIDDEN |

Uji memakai sesi employee, CSRF, dan fixture gRPC. Verifikasi positif lewat HTTPS masih menunggu fixture IAM.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:d0f64e9cc9d5a3bc6fcd2f5e2d49a598ef6ce6cb9221e3d596ce00b52671646d`. Mutation hapus draf dan catat tanda tangan ditolak tanpa CSRF. Seluruh 14 skenario guard HTTPS lulus.
