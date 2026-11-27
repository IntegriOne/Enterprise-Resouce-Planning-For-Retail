# Pengujian Mutation aktivasi dan terminasi pekerja

Suite `TestHRMWorkerLifecycleThroughGateway` menggunakan GraphQL, sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Aktivasi dengan Relay ID | active |
| Prasyarat aktivasi ditolak RPC | CONFLICT |
| Terminasi dengan ID rencana yang cocok | terminated |
| Rencana offboarding tidak cocok | CONFLICT |
| Format ID rencana tidak valid | INVALID_ARGUMENT |
| Grant activate digunakan untuk terminasi | FORBIDDEN |
| Aktivasi tanpa MFA | FORBIDDEN |

Uji positif HTTPS dengan grant IAM resmi masih tertunda.
