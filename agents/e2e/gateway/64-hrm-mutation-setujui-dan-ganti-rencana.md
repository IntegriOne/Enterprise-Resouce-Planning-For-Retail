# Pengujian Mutation persetujuan dan penggantian rencana tenaga kerja

Fixture gRPC diakses melalui GraphQL dengan sesi employee dan CSRF.

| Kasus | Hasil |
| --- | --- |
| Grant approve dan MFA | approved |
| Verifier Finance tidak tersedia | UNAVAILABLE |
| Grant supersede dan MFA | ID versi pengganti diteruskan dan dikembalikan |
| Rencana menggantikan dirinya sendiri | INVALID_ARGUMENT |
| Grant submit dipakai untuk approve | FORBIDDEN |
| Supersede tanpa MFA | FORBIDDEN |

Persetujuan runtime menunggu verifier Finance. Uji positif HTTPS menunggu grant IAM resmi.
