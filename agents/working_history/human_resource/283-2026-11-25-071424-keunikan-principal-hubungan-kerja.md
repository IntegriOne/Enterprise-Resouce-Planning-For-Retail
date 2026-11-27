# Keunikan principal pada hubungan kerja berjalan

Baseline PostgreSQL menolak satu principal yang terikat pada dua hubungan kerja berjalan. Riwayat hubungan kerja yang telah berakhir tetap dapat merujuk principal yang sama. Uji integritas membuktikan penolakan ikatan ganda dan penerimaan hubungan kerja baru setelah terminasi. Verifikasi skema lengkap lulus. RPC pengaitan dan verifikasi principal masih terbuka.
