# Pengujian Mutation pencatatan kontrak bertanda tangan

| Kasus fixture | Hasil |
| --- | --- |
| Grant record_signature, MFA, CSRF, dan versi sesuai | signed |
| Verifier DRM tidak tersedia | UNAVAILABLE |
| Versi dalam respons RPC berbeda | UNAVAILABLE |
| RPC menolak kontrak di luar kondisi draf | CONFLICT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant create dipakai untuk pencatatan | FORBIDDEN |

Verifikasi tanda tangan runtime masih memerlukan verifier DRM.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:9138e5045f116bab4c910cd6c7f449be3bbb77d40f4c4329c98bd3c95186d09c`. Permintaan dengan ID fixture tanpa CSRF mencapai guard Caddy–Gateway dan menghasilkan `FORBIDDEN` dengan pesan token CSRF absen. Uji positif membutuhkan sesi berizin dan verifier DRM.
