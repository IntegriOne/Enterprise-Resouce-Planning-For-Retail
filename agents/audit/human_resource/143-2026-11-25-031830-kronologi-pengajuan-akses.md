# Kronologi pengajuan akses pekerja

Tes validator mereproduksi pengajuan yang menunjuk event kepegawaian setelah waktu pengajuan. Validator kini menolak referensi tersebut dan menerima waktu yang sama. Pemeriksaan memakai occurred_at sehingga tanggal efektif yang direncanakan tetap dapat berbeda. Tes validator, gRPC Access Request dengan PostgreSQL sementara dan race detector, serta vet bertag integration lulus.
