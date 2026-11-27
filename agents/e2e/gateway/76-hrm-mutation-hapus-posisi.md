# Pengujian Mutation penghapusan posisi

`TestHRMDeleteWorkforceRecordsThroughGateway` memakai sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Grant hapus posisi dan MFA | true |
| Posisi dirujuk record lain | CONFLICT |
| Sesi tanpa MFA | FORBIDDEN |
| Grant hapus workforce plan dipakai untuk posisi | FORBIDDEN |

Uji positif HTTPS dengan grant IAM resmi masih tertunda.
