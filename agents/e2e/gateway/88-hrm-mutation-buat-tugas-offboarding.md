# Pengujian Mutation pembuatan tugas offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Grant create dan is_mandatory false eksplisit | Pending dengan is_mandatory false |
| Rencana tidak lagi draf | CONFLICT |
| ID rencana tidak valid | INVALID_ARGUMENT |
| Grant baca dipakai untuk membuat | FORBIDDEN |

Uji menggunakan sesi employee, CSRF, dan fixture gRPC.

## HTTPS lokal

Kedua replika sehat setelah rollout image `sha256:c6a0ea5a49b0b3d6e693d8346d0106bdc26ea0e510e99599d0d31a7efd543e1e`. Permintaan tanpa CSRF untuk pembuatan draf dan tugas mencapai guard Gateway lewat Caddy HTTPS dan menghasilkan `FORBIDDEN`. Akses positif menunggu fixture IAM.
