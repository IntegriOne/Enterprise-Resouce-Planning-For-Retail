# Uji browser lokasi requisition per badan hukum

Chrome memakai HTTPS pada stack terisolasi. Superadmin masuk dengan MFA dan memasang role serta permission melalui GraphQL IAM. Fixture Enterprise Foundation berisi dua badan hukum, masing-masing dengan satu unit dan lokasi aktif.

| Kasus | Hasil |
| --- | --- |
| Pilih unit badan hukum pertama | Hanya lokasi badan hukum pertama yang ditawarkan. |
| Pindah ke unit badan hukum kedua | Pilihan lama hilang dan lokasi badan hukum kedua tersedia. |
| Kembali ke unit awal, simpan, lalu ajukan | Lokasi yang dipilih tersimpan dan tampil sebagai kode serta nama pada detail. |
| Ulangi alur edit dan penghapusan draf terpisah | Lulus. |

TypeScript, ESLint direktori requisition, dan build image frontend lulus.
