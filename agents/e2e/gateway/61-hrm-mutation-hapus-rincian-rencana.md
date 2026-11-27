# Pengujian Mutation penghapusan rincian rencana tenaga kerja

`TestHRMDeleteWorkforceRecordsThroughGateway` menjalankan GraphQL dengan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Grant delete global dan MFA | true |
| RPC menolak penghapusan karena prasyarat workflow | CONFLICT |
| ID rincian tidak valid | INVALID_ARGUMENT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant create dipakai untuk menghapus | FORBIDDEN |

Verifikasi HTTPS dengan grant IAM resmi masih tertunda.
