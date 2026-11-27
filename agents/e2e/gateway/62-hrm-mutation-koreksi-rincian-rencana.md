# Pengujian Mutation koreksi rincian rencana tenaga kerja

`TestHRMUpdateWorkforcePlanLineThroughGateway` menggunakan sesi employee, CSRF, dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Koreksi FTE dan pengosongan tanggal akhir | Field mask hanya memuat kedua field. FTE tetap desimal dan tanggal akhir menjadi null. |
| Input kosong | INVALID_ARGUMENT |
| RPC menolak perubahan saat review | CONFLICT |
| Grant create dipakai untuk koreksi | FORBIDDEN |

Verifikasi HTTPS dengan grant IAM resmi masih tertunda.
