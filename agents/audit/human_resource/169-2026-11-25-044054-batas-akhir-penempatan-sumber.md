# Batas akhir penempatan sumber

Promosi sebelumnya menolak penempatan sumber dengan akhir terjadwal tepat sehari sebelum tanggal efektif pengganti. Helper kini menerima H−1 dan masih menolak sumber yang sudah berakhir lebih awal. Uji workflow mereproduksi penolakan H−1 sebelum perbaikan dan memastikan celah H−2 tetap ditolak.

Transfer umum memerlukan penyimpanan permintaan terpisah dari penempatan aktif. Constraint `ex_hr_primary_assignment_period` memeriksa semua baris primer, termasuk draf. Menyimpan draf pengganti yang mulai saat penempatan lama masih terbuka akan ditolak PostgreSQL. Tahap pengajuan dan persetujuan transfer perlu membawa identitas pengaju, penyetuju, sumber, target, dan tanggal efektif. Penerapan harus menutup penempatan lama, memasang pengganti, dan mencatat lifecycle dalam satu transaksi.
