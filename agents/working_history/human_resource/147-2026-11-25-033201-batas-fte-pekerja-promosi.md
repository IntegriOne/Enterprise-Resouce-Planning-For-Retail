# Batas FTE pekerja selama penempatan hasil promosi

Penerapan promosi kini memeriksa total FTE pekerja pada seluruh penempatan yang telah disetujui. Penempatan asal dikeluarkan dari perhitungan karena transaksi menutupnya pada H−1. Batas pekerja adalah 100%, dan batas posisi tujuan tetap mengikuti headcount. Pemeriksaan berlangsung selama periode penempatan baru.

Tes gRPC PostgreSQL sementara membuktikan bahwa penempatan tambahan pada posisi lain dapat memblokir promosi jika total FTE pekerja melebihi batas. Tes service dan integrasi gRPC dengan race detector serta go vet bertag integration lulus.
