# Uji browser pengajuan dan penghapusan draf requisition

Chrome mengakses workspace melalui HTTPS pada stack terisolasi. Superadmin masuk, menuntaskan MFA, dan memperoleh role uji beserta permission melalui GraphQL IAM.

| Kasus | Hasil |
| --- | --- |
| Buka konfirmasi pengajuan, lalu batal | Detail kembali menampilkan aksi Ajukan. |
| Konfirmasi pengajuan | Status daftar berubah menjadi Menunggu persetujuan. Detail tidak menawarkan aksi edit, ajukan, atau hapus. |
| Konfirmasi penghapusan draf lain | Pesan sukses tampil dan pencarian nomor draf kosong. |
| Verifikasi GraphQL | Requisition pertama berstatus submitted dengan jumlah yang sesuai. Pencarian draf kedua mengembalikan total nol. |
| Regresi form | Pembuatan, edit, dan pengosongan referensi opsional tetap lulus. |

Jalur permintaan verifikasi ulang dari browser belum diuji pada sesi yang masa verifikasinya sudah habis. TypeScript, ESLint, dan build image frontend lulus.
