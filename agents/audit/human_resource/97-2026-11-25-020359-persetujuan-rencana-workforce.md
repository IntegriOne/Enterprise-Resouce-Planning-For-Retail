# Persetujuan rencana workforce

ApproveWorkforcePlan memakai izin step-up dan memisahkan pembuat sekaligus pengusul dari penyetuju. Snapshot kebutuhan dikirim melalui port verifikasi Finance dengan rencana terkunci. Setelah bukti anggaran diterima, validator memeriksa ulang referensi dan status. Status approved, referensi Finance, actor, waktu server, dan outbox dicatat atomik.

Tes integrasi gRPC ke PostgreSQL membuktikan self approval dan bukti yang ditolak tidak menghasilkan perubahan. Persetujuan dengan verifier tiruan berhasil dan mencatat tepat satu event. Bootstrap belum terhubung ke kontrak Finance otoritatif dan RPC mengembalikan Unavailable sampai port terhubung. Supersede masih terbuka.
