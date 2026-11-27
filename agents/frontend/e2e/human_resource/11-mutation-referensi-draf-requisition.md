# Uji browser referensi opsional draf requisition

Chrome mengakses workspace melalui HTTPS pada stack terisolasi. Superadmin masuk dan menuntaskan MFA. Role uji, pemasangan permission katalog, dan pemberian role dijalankan melalui GraphQL IAM. Enterprise Foundation menyediakan lokasi aktif. Fixture HRM menyediakan posisi aktif dan rencana tenaga kerja disetujui dengan lini perekrutan.

| Kasus | Hasil |
| --- | --- |
| Pilih ketiga referensi | Posisi dan lokasi tampil dengan kode serta nama. Rencana tampil dengan kode dan versi, lalu lini tampil dengan nomor. |
| Simpan draf baru | Browser mengirim tiga UUID referensi. Daftar menampilkan draf baru. |
| Buka form edit | Posisi, lokasi, rencana, dan lini tersimpan terpilih kembali. |
| Simpan tanpa perubahan | Browser tidak mengirim mutasi. |
| Kosongkan referensi dan ubah jumlah | Browser mengirim tiga nilai null dan jumlah baru. Daftar menampilkan jumlah baru. |
| Periksa hasil dan teks UI | GraphQL menampilkan satu draf. HRM masih mencatat hiring manager yang sama. Teks UI tidak memuat `gid://`. |
