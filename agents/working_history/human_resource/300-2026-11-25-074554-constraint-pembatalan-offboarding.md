# Constraint pembatalan dan rencana offboarding terbuka

PostgreSQL memeriksa bahwa metadata pembatalan hanya ada saat status cancelled, persetujuan sudah tercatat, dan waktu pembatalan tidak mendahului persetujuan. Indeks parsial mencegah dua rencana terbuka untuk pekerja yang sama. Verifikasi SQL lengkap dan regresi gRPC PostgreSQL dengan race detector lulus.
