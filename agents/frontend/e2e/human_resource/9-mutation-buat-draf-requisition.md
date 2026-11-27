# Uji browser pembuatan draf requisition

Chrome headless membuka `/human-resource/requisitions` melalui HTTPS pada stack terisolasi. Superadmin masuk dan menuntaskan MFA. Role uji dibuat, permission katalog dipasang pada role, dan role diberikan kepada superadmin melalui GraphQL IAM. Akun hiring manager dibuat melalui `createEmployeeUser` dan terhubung ke pekerja aktif.

| Kasus | Hasil |
| --- | --- |
| Buka form dengan izin pembuatan | Tombol Buat draf tersedia. Simpan draf nonaktif sebelum pencarian pekerja berhasil. |
| Muat referensi | Jabatan menampilkan kode dan judul. Unit EF menampilkan kode dan nama. |
| Cari nomor pekerja tak dikenal | Pesan pekerja tidak ditemukan tampil. Simpan draf tetap nonaktif. |
| Cari hiring manager aktif | Nama dan nomor pekerja tampil. |
| Simpan kebutuhan dua pekerja | Browser mengirim `createRequisition` dengan kunci idempotensi. Tabel menampilkan satu draf dengan nomor, jabatan, unit, dan jumlah yang sesuai. |
| Periksa hasil yang disimpan | GraphQL mengembalikan satu draf. Database HRM terisolasi mencatat principal yang sesuai dengan akun hiring manager. |
| Periksa teks form dan tabel | Tidak ditemukan `gid://`. |

Pemeriksaan TypeScript, ESLint untuk direktori requisition, dan build image frontend lulus.
