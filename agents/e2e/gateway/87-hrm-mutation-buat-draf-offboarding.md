# Pengujian Mutation pembuatan draf offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Grant create, Relay ID, dan tanggal terminasi | draft |
| Nomor rencana duplikat | CONFLICT |
| ID pekerja tidak valid | INVALID_ARGUMENT |
| Grant approve dipakai untuk membuat | FORBIDDEN |

Fixture memakai sesi employee, CSRF, dan gRPC.

## HTTPS lokal

Kedua replika sehat setelah rollout image `sha256:c6a0ea5a49b0b3d6e693d8346d0106bdc26ea0e510e99599d0d31a7efd543e1e`. Permintaan tanpa CSRF untuk pembuatan draf dan tugas mencapai guard Gateway lewat Caddy HTTPS dan menghasilkan `FORBIDDEN`. Akses positif menunggu fixture IAM.
