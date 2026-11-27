# Uji HTTPS perubahan hiring manager

Superadmin memberikan grant melalui GraphQL IAM dan membuat requisition dengan hiring manager awal. Pekerja pengganti ditautkan pada principal IAM yang sudah ada melalui fixture PostgreSQL tes. Unit dan party berasal dari Enterprise Foundation.

| Kasus | Hasil |
| --- | --- |
| Kirim ID kosong, GID badan hukum, atau ID pekerja yang tidak ditemukan | Mutation menolak |
| Kirim GID pekerja aktif pengganti | Draf berhasil diperbarui |
| Pembacaan database HRM tes | Principal yang tersimpan cocok dengan pekerja pengganti |
| Pengajuan setelah perubahan | Draf menjadi `submitted` |

Uji gRPC PostgreSQL membuktikan mask pekerja dan menolak penggantian lewat principal langsung maupun dua jenis referensi sekaligus.
