# Pengujian Query daftar tugas offboarding

| Kasus fixture | Hasil |
| --- | --- |
| Grant baca tugas dan ID rencana sesuai | Halaman tugas dan total dikembalikan |
| RPC mengembalikan tugas rencana lain | UNAVAILABLE |
| ID rencana tidak valid | INVALID_ARGUMENT |
| Grant baca rencana dipakai untuk tugas | FORBIDDEN |

Query memakai sesi employee dan fixture gRPC.

## HTTPS lokal

Kedua replika sehat dengan image `sha256:e6d6bdb54b5ae7335f03907a9fccb418e515651558f7a0efabe73a153c16faa3`. Query daftar tugas dikenal dan mengembalikan `UNAUTHENTICATED` tanpa sesi. [Matriks guard seluruh schema](84-hrm-query-mutation-guard-seluruh-schema.md).
