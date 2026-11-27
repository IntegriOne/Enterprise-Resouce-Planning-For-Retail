# Pengujian Mutation pembuatan draf kontrak

| Kasus fixture | Hasil |
| --- | --- |
| Grant create, MFA, CSRF, dan Relay ID pekerja | draft dengan waktu tanda tangan null |
| Nomor kontrak duplikat | CONFLICT |
| Masa pemberitahuan negatif | INVALID_ARGUMENT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant baca dipakai untuk membuat | FORBIDDEN |

Akses positif HTTPS dengan grant IAM resmi masih tertunda.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:4f6ce2b6febe50bfe7a8219c69b2c857f28d44b55a48eaaa062d2578df4be29e`. Mutation pembuatan draf tanpa CSRF dikenali schema dan mengembalikan `FORBIDDEN` lewat Caddy. Suite HTTPS pekerja, kontrak dan posisi meluluskan 12 skenario guard.
