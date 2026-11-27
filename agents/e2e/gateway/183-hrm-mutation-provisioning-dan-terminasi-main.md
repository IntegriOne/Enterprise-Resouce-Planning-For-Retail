# Provisioning dan terminasi employee melalui HTTPS

## Lingkungan

Binary IAM, HRM, EF, dan Gateway dibangun dari arsip `origin/main` pada commit `08fdec69e0f0ca54d3d4ebca234c7bd2da355720`. Stack memakai database dan Redis disposable, Caddy HTTPS pada port 18443, serta listener mTLS privat untuk terminasi. Akun superadmin login dan melakukan MFA. Role uji dibuat, permission ditautkan, dan assignment global serta badan hukum diberikan lewat mutation GraphQL IAM. Fixture kontrak dan prasyarat terminasi disiapkan di database HRM disposable.

## Matriks kasus

| Kasus | Hasil |
| --- | --- |
| Antrean IAM tanpa sesi | Akses ditolak |
| Pemberian scope badan hukum yang tidak ada | Ditolak oleh verifikasi EF |
| Pembuatan pekerja dengan tanggal hire sebelum inkorporasi | Ditolak HRM |
| Pengajuan tanpa bukti hire | Ditolak |
| Pengajuan worker aktif dan pengajuan ulang | Berhasil |
| Antrean dan detail IAM untuk pengajuan yang benar | Berhasil |
| GID milik tipe lain atau worker berbeda dari pengajuan | Ditolak |
| `createEmployeeUser` dan pengulangan mutation | Berhasil, satu akun dan satu penerimaan |
| Tautan IAM ke worker dan pembacaan status dari HRM | Principal cocok, pengajuan tertaut keluar dari antrean |
| Aktivasi akun setelah tautan | Berhasil |
| Terminasi worker yang terhubung | Berhasil |
| Clearance akun melalui mTLS | IAM mengonfirmasi akun nonaktif |
| Pemakaian sesi lama dan login ulang sesudah terminasi | Ditolak |
| Cetak ulang slip tanpa MFA atau saat layanan Finance/DRM tidak tersedia | Ditolak sesuai syarat akses dan dependensi |
| Pencabutan assignment dan penghapusan role uji | Berhasil melalui GraphQL |

## Perintah

`python backend/gateway/internal/e2e/scripts/run_hrm_employee_provisioning.py --source-root <arsip-main> --termination` keluar dengan kode 0.

Pengulangan dengan `--guards` juga selesai dengan kode 0. Seluruh **52 Query** menolak request tanpa sesi dan seluruh **61 Mutation** menolak request tanpa CSRF. Pemeriksaan SDL menemukan **24 tipe Node** dan **24 Relay Connection**. Matriks guard ini memeriksa batas transport. Skenario bisnis yang lulus tercatat pada tabel di atas.
