# Pengujian Mutation penghapusan draf rencana tenaga kerja

`TestHRMDeleteWorkforceRecordsThroughGateway` menguji GraphQL dengan sesi employee, MFA, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Draf kosong dan grant hapus rencana | true |
| RPC menolak penghapusan draf dengan rincian | CONFLICT |
| Grant hapus rincian dipakai untuk menghapus rencana | FORBIDDEN |

Pengujian penghapusan rincian tetap berada pada suite yang sama. Uji positif HTTPS menunggu grant IAM resmi.
