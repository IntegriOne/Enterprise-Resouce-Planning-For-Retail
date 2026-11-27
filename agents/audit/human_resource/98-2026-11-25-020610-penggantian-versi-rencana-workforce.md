# Penggantian versi rencana workforce

SupersedeWorkforcePlan memerlukan izin step-up, actor manusia, alasan dan rencana pengganti yang sudah disetujui. Pengganti harus memakai versi lebih tinggi dengan kode, badan hukum, skenario dan periode yang sama. Snapshot pengganti diperiksa ulang dengan lock. Baseline 000001 menyimpan relasi pengganti dan alasan serta memeriksa konsistensi status. Status lama dan event outbox dicatat atomik sambil menjaga bukti approval dan anggaran lama.

Tes gRPC ke PostgreSQL lulus untuk penolakan versi dibatalkan, penggantian berhasil, bukti anggaran lama tersimpan, pengulangan ditolak, dan tepat satu event. Pengganti disiapkan dengan fixture. Tes migrasi baseline lulus. Verifier Finance produksi tetap belum tersedia.
