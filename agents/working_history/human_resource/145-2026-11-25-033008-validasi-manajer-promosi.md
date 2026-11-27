# Validasi manajer pada penempatan hasil promosi

Penerapan promosi memvalidasi manajer yang diwarisi dari penempatan asal. Manajer harus aktif, berada pada badan hukum pekerja yang sama, dan masa kerjanya mencakup periode penempatan baru jika tanggal mulai atau terminasi tercatat. Pemeriksaan berjalan sebelum mutasi penempatan.

Tes gRPC dengan PostgreSQL sementara mengatur tanggal terminasi manajer di fixture, memverifikasi FailedPrecondition, lalu membuktikan promosi berhasil setelah tanggal dipulihkan. Tes service dengan race detector dan go vet bertag integration lulus.
