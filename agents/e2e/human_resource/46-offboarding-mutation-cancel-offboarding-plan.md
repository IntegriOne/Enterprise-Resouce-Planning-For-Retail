# CancelOffboardingPlan melalui gRPC dan PostgreSQL

Tes dengan database sementara dan race detector membuktikan rencana draf tidak bisa dibatalkan lewat command ini. Rencana yang disetujui dan belum berjalan dapat dibatalkan dengan alasan, timestamp, dan riwayat persetujuan tetap terlihat. Start terhadap rencana yang sudah dibatalkan ditolak dan rencana pengganti dapat dibuat.

Saat StartOffboardingPlan dan CancelOffboardingPlan dikirim bersamaan untuk rencana pengganti yang sudah disetujui, satu RPC sukses dan satu RPC menerima FailedPrecondition. PostgreSQL menyimpan tepat satu event outbox di antara kedua aksi tersebut.

Verifikasi SQL menolak metadata pembatalan pada draf, pembatalan yang mendahului persetujuan, dan dua rencana terbuka milik pekerja yang sama. Setelah pembatalan valid, PostgreSQL menerima rencana pengganti.
