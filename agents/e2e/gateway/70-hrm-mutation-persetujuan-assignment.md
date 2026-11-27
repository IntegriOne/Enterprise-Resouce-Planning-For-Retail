# Pengujian Mutation persetujuan assignment

`TestHRMApproveWorkerAssignmentThroughGateway` memakai GraphQL, sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Grant approve dan MFA | FTE tetap desimal dan waktu persetujuan dikembalikan |
| HRM menolak karena kapasitas | CONFLICT |
| ID tidak valid | INVALID_ARGUMENT |
| Tanpa MFA | FORBIDDEN |
| Grant baca dipakai untuk menyetujui | FORBIDDEN |

Uji HTTPS positif dengan grant IAM resmi masih tertunda.
