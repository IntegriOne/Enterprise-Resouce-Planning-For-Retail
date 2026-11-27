# Pengujian Query assignment pekerja

`TestHRMWorkerAssignmentQueriesThroughGateway` memakai sesi employee dan fixture gRPC.

| Kasus | Hasil |
| --- | --- |
| Detail tanpa persetujuan | FTE tetap desimal dan waktu persetujuan null |
| Daftar memakai Relay ID pekerja | Halaman dikembalikan |
| RPC mengembalikan pekerja lain | UNAVAILABLE |
| Daftar lewat filter posisi | Halaman memuat posisi yang diminta |
| RPC mengembalikan posisi lain | UNAVAILABLE |
| Filter pekerja dan posisi dikirim bersamaan | INVALID_ARGUMENT |
| Grant approve dipakai untuk membaca | FORBIDDEN |

Uji positif HTTPS dengan grant IAM resmi masih tertunda.
