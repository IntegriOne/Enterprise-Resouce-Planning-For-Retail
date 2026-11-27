# Pengujian Mutation pembatalan offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Grant cancel, MFA, dan alasan valid | cancelled |
| Rencana sudah berjalan | CONFLICT |
| Alasan kosong | INVALID_ARGUMENT |
| Grant start dipakai untuk pembatalan | FORBIDDEN |
| Sesi tanpa MFA | FORBIDDEN |

Uji memakai sesi employee, CSRF, dan fixture gRPC.

## HTTPS lokal

Kedua replika Gateway sehat pada image `sha256:b7886cd49e25ac2bf6e785b1934a003e3b7f55abb99601af72154dc6bfab5b79`. Permintaan melalui Caddy dengan CA lokal dan Origin yang sesuai mengembalikan `FORBIDDEN` saat token CSRF absen. Uji fixture pembatalan lulus dengan race detector. Uji berhasil menggunakan grant resmi IAM masih tertunda.

Validasi status, metadata, dan persaingan transaksi mulai/batal juga lulus pada suite gRPC HRM dengan PostgreSQL Docker. [Uji integrasi HRM](../human_resource/53-offboarding-mutation-cancel-offboarding-plan.md).
