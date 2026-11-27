# Uji browser detail requisition yang diajukan

Chrome mengakses stack terisolasi melalui HTTPS. Superadmin memasang role dan permission lewat GraphQL IAM serta masuk dengan MFA.

| Kasus | Hasil |
| --- | --- |
| Buat dan edit draf berisi referensi | Posisi, lokasi EF, serta lini perekrutan tetap tersimpan ketika patch hanya mengubah jumlah kebutuhan. |
| Ajukan dan buka panel detail | Panel menampilkan status Menunggu persetujuan, kode dan judul posisi, kode dan nama lokasi, kode serta versi rencana, dan nomor lini. |
| Periksa label lain | Jenis kebutuhan dan kategori pekerja tampil dalam Bahasa Indonesia. Teks panel tidak memuat `gid://`. |
| Periksa tindakan draft | Aksi edit, ajukan, dan hapus tidak tampil pada requisition yang diajukan. |

TypeScript, ESLint pada direktori requisition, dan build image frontend lulus.
