# Konsolidasi baseline setara

Commit `1325fe4` dipublikasikan dan diverifikasi di `origin/main`. Pemasangan baru memakai baseline `000001`. Riwayat lama yang lengkap dengan tujuh checksum yang cocok tetap diterima untuk baseline setara. Ledger dan data bisnis tidak ditulis ulang.

Dua database PostgreSQL disposable menghasilkan dump skema identik. Rollback menyisakan nol tabel HRM. Uji integrasi dengan race detector membuktikan pemasangan baru, startup ulang dengan riwayat lama, data bisnis yang tetap terbaca, dan penolakan checksum salah. Vet, tes race seluruh modul, dan build command lulus pada kandidat terisolasi. Penambahan skema terminasi dan wiring Gateway masih dalam pekerjaan.
