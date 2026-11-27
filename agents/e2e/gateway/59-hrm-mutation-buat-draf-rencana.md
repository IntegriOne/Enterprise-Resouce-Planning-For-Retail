# Pengujian Mutation pembuatan draf rencana tenaga kerja

Uji `TestHRMCreateWorkforcePlanThroughGateway` menjalankan GraphQL dengan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Grant create global dan data valid | Draf terbentuk. Fixture memverifikasi bahwa status dan metadata persetujuan tidak dikirim. |
| Konflik duplikat dari gRPC | CONFLICT |
| Versi nol | INVALID_ARGUMENT |
| Grant submit tanpa grant create | FORBIDDEN |

Verifikasi HTTPS dengan grant IAM resmi masih terbuka.
