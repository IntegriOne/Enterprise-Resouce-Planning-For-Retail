# Penempatan asal saat menerapkan promosi

Penerapan promosi kini mencocokkan penempatan asal rekomendasi dengan penempatan primer pada baris pekerja yang dikunci. Penempatan primer kosong atau berbeda menghasilkan FailedPrecondition sebelum transaksi mengubah penempatan atau menulis event.

Tes integrasi gRPC dengan PostgreSQL sementara mengosongkan referensi primer setelah persetujuan dan memverifikasi penolakan. Alur sukses teruji setelah fixture dikembalikan. Tes service dengan race detector dan go vet bertag integration lulus.
