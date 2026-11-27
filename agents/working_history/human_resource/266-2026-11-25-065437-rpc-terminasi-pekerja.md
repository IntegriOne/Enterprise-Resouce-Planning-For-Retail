# RPC terminasi pekerja

Personnel menerima keputusan terminasi melalui RPC privat, memvalidasi rencana, menjalankan mutasi dalam transaksi, dan mencatat lifecycle serta outbox. Kontrak dan penempatan yang sedang berjalan dipotong pada hari kerja terakhir. Jadwal mendatang yang memerlukan pembatalan memblokir transaksi. Handoff Gateway memuat Mutation dan izin MFA. Tes gRPC–PostgreSQL dengan race detector membuktikan terminasi hingga penutupan offboarding pada pekerja tanpa penempatan atau kontrak aktif. [Laporan](../../e2e/human_resource/41-personnel-mutation-terminate-worker.md)
