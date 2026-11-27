# Uji browser kecocokan hiring manager dengan badan hukum unit

Superadmin masuk dengan MFA melalui HTTPS dan memasang role serta izin melalui GraphQL IAM pada stack terisolasi. Enterprise Foundation menyediakan dua badan hukum dan unit terkait. HRM menyediakan pekerja aktif di badan hukum pertama.

| Kasus | Hasil |
| --- | --- |
| Cari pekerja dan pilih unit badan hukum lain | Form menampilkan alasan dan menonaktifkan Simpan draf. |
| Kembali ke unit yang sesuai | Peringatan hilang dan Simpan draf aktif kembali. |
| Simpan dan ajukan | Draf tersimpan dengan hiring manager yang dipilih dan label referensi bisnis tetap tampil setelah pengajuan. |

Uji pertama mengungkap perbedaan format Global ID dari EF dan UUID dari HRM. Perbandingan menggunakan UUID setelah normalisasi. TypeScript, ESLint, build image, dan pengulangan uji HTTPS lulus.
