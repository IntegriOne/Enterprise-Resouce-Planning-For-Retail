# Pengujian Mutation aktivasi posisi

`TestHRMActivatePositionThroughGateway` memakai sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Posisi planned dengan grant dan MFA | active |
| RPC menolak prasyarat aktivasi | CONFLICT |
| ID posisi tidak valid | INVALID_ARGUMENT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant hanya untuk membaca | FORBIDDEN |

Uji HTTPS dengan grant IAM resmi masih tertunda.
