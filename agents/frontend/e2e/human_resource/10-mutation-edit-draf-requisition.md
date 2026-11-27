# Uji browser edit draf requisition

Chrome membuka workspace Human Resource melalui HTTPS pada stack terisolasi. Superadmin masuk dan menuntaskan MFA. Role uji dan pemasangan permission katalog dilakukan melalui GraphQL IAM, termasuk pemberian role kepada superadmin.

| Kasus | Hasil |
| --- | --- |
| Buat draf dengan jumlah dua pekerja | Data tampil dengan status Draf dan label bisnis dari HRM serta Enterprise Foundation. |
| Buka form edit | Jumlah, pilihan jabatan, dan pilihan unit terisi. Nomor requisition hanya bisa dibaca. |
| Simpan tanpa perubahan | Pesan tampil dan browser tidak mengirim mutasi. |
| Ubah jumlah menjadi tiga | Browser mengirim `updateRequisition` dengan patch `headcount` saja. Daftar kembali menampilkan draf yang sama dengan jumlah tiga. |
| Periksa hasil tersimpan | GraphQL mengembalikan satu draf. HRM masih mencatat akun hiring manager yang dipilih saat pembuatan. |
| Periksa teks tampilan | Tidak ditemukan `gid://`. |

TypeScript, ESLint direktori requisition, dan build image frontend lulus.
