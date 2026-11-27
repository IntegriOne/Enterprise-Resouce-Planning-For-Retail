# Pengujian Mutation penambahan rincian rencana tenaga kerja

`TestHRMCreateWorkforcePlanLineThroughGateway` menjalankan GraphQL dengan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Grant create global dan data valid | Rincian dibuat. Nilai FTE dan biaya diteruskan sebagai string desimal. |
| RPC menolak perubahan rencana berstatus review | CONFLICT |
| ID rencana tidak valid | INVALID_ARGUMENT |
| Grant hanya untuk membaca | FORBIDDEN |

HRM mengunci rencana induk saat validasi dan hanya mengizinkan penulisan pada draf. Verifikasi HTTPS dengan grant IAM resmi masih tertunda.
