# Riwayat penugasan dibatalkan

Commit `12fb870` di `origin/main` menambah bukti pembatalan penugasan dan mengeluarkan penugasan yang dibatalkan dari batas periode aktif. Baseline tetap `000001`. Runner merekonsiliasi riwayat tujuh migrasi lama atau baseline setara sebelumnya dalam satu transaksi. Setelah pembaruan ledger, startup memerlukan binary baru.

Uji PostgreSQL disposable lulus untuk pemasangan baru, kedua riwayat lama, retensi data, startup ulang, dan penolakan riwayat yang tidak dikenal. Tes race, vet, dan build kandidat lulus. Database aktif tidak dimodifikasi. Integrasi operasi bisnis terminasi dan Gateway masih dalam pekerjaan.
