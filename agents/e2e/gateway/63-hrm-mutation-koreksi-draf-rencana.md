# Pengujian Mutation koreksi draf rencana tenaga kerja

Fixture gRPC diakses melalui GraphQL dengan sesi employee dan CSRF.

| Kasus | Hasil |
| --- | --- |
| Grant update global dan perubahan tanggal akhir | Mask hanya memuat tanggal akhir |
| Tanggal akhir null | INVALID_ARGUMENT |
| Rencana sedang review | CONFLICT |
| Grant create dipakai untuk update | FORBIDDEN |
| Pembuatan draf dengan kontrak baru | RPC menerima input tanpa status, metadata persetujuan, atau referensi anggaran |

Verifikasi HTTPS dengan grant IAM resmi masih tertunda.
