# Perluasan katalog tindakan HRM

Katalog HRM diperjelas menjadi izin `read`, `create`, `update`, `delete` yang terpisah, tanpa alias `write`. Matriks rancangan workflow dilengkapi untuk publikasi lowongan, asesmen, tugas onboarding/offboarding, tahap payroll, verifikasi kualifikasi, suksesi, layanan karyawan, serta siklus dan pool talent. Kode rancangan dibedakan dari 323 kode runtime yang sudah memiliki mapping; lima command runtime tetap berlaku. Tidak ada permission atau grant IAM yang dibuat.

Pemeriksaan DBML membatalkan rancangan waiver onboarding yang tidak memiliki status/bukti di skema. Permintaan posting payroll tetap diajukan ke Finance; bukti jurnal dan pembayaran diterima dari pemiliknya. `git diff --check` lulus. Perubahan ini dokumentasi katalog, sehingga tes Go/Docker tidak dijalankan ulang.
