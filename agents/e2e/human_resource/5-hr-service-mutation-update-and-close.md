# Regresi mutasi tiket HR melalui gRPC

Tes memakai server gRPC, PostgreSQL disposable, dan policy serta verifier DRM tiruan.

| Kasus | Hasil |
| --- | --- |
| Ubah status tiket terbuka menjadi cancelled lewat UpdateServiceRequest | FailedPrecondition |
| Ubah nomor tiket atau idempotency key setelah pembuatan | FailedPrecondition untuk kedua field |
| Tetapkan tenggat sebelum waktu pembukaan | InvalidArgument |
| Ubah tenggat valid pada tiket terbuka | Berhasil |
| Buat event triaged melalui RPC generik | FailedPrecondition |
| Buat komentar dengan lampiran yang belum terbukti di DRM | Unavailable |
| Buat komentar tanpa lampiran saat tiket masih terbuka | Berhasil |
| Ubah tenggat setelah penugasan | FailedPrecondition |
| Selesaikan tiket dengan versi dokumen yang tidak ditemukan | FailedPrecondition |
| Selesaikan tiket dengan dokumen terverifikasi dan tutup oleh pemohon | Berhasil |
| Buat komentar setelah penutupan | FailedPrecondition |
| Hitung outbox tiket | Lima event sesuai mutasi tiket yang berhasil |

Katalog tiket mencantumkan assign, resolve, dan close. Status cancelled tersedia pada skema. Command pembatalan memerlukan aturan bisnis dan izin tersendiri dalam katalog yang disetujui.
