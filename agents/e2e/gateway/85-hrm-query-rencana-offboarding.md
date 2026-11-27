# Pengujian Query rencana offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Detail rencana dengan grant baca | Status dan tanggal efektif dikembalikan |
| ID rencana tidak ditemukan | NOT_FOUND |
| Daftar dengan Relay ID pekerja | Halaman dikembalikan |
| RPC mengembalikan rencana pekerja lain | UNAVAILABLE |
| Grant terminasi dipakai untuk membaca | FORBIDDEN |

Pengujian menggunakan sesi employee dan fixture gRPC. Query baru belum diuji pada replika HTTPS lokal.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:9c3c8d9a529cd0725f3aad5bc3a9a30bc441989d870199221024b13455f19f0c`. Permintaan tanpa cookie melalui Caddy mengembalikan `UNAUTHENTICATED` untuk detail maupun daftar. Suite guard pekerja dan posisi, race seluruh modul Gateway, serta vet lulus. Akses positif masih menunggu fixture IAM resmi.
