# Lifecycle penggantian penempatan primer

ApplyPrimaryAssignment menyimpan event transfer saat mengganti referensi primer yang sudah ada. Event mencatat kedua ID penempatan, actor, dan tanggal penerapan server dalam transaksi worker serta outbox. Penerapan primer pertama tidak mencatat transfer. Tes service memverifikasi dua perilaku. Regresi gRPC Personnel dengan PostgreSQL sementara, Redis TLS, race detector, dan go vet lulus. Penggantian dengan event transfer belum diuji melalui gRPC.
