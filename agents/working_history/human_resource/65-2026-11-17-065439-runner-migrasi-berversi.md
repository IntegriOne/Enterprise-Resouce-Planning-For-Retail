# Runner migrasi berurutan HRM

Runner migrasi HRM sekarang memvalidasi urutan versi, celah riwayat, dan checksum seluruh versi yang dikenal sebelum menerapkan versi baru dalam transaksi yang sama. Baseline `000001` tetap utuh. Tes integrasi PostgreSQL pada skema terisolasi membuktikan instalasi awal dan idempotensi serta penolakan versi asing dan checksum rusak. Seluruh pemeriksaan Go serta build/healthcheck Docker HRM lulus. Migrasi `000002`, perubahan generator DBML, dan workflow requisition masih menjadi pekerjaan berikutnya.
