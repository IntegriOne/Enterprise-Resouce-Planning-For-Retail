# Verifikasi Foundation pada penerapan penempatan primer

ApplyPrimaryAssignment kini meminta Enterprise Foundation memverifikasi unit organisasi, badan hukum dan lokasi posisi untuk periode penempatan. Referensi yang ditolak pemilik menghasilkan FailedPrecondition. Kegagalan akses verifier menghasilkan Unavailable dan pembaruan worker dibatalkan.

Tes unit membuktikan bahwa fakta posisi diteruskan dan penolakan pemilik menghentikan mutasi. Tes security dan cmd dengan race detector, integrasi gRPC dengan PostgreSQL sementara dan Redis TLS, serta go vet lulus. Integrasi gRPC memakai verifier tiruan, sehingga pertukaran produksi dengan Foundation belum diverifikasi pada pengujian ini.
