# Kelengkapan data sebelum mutasi promosi

Pencatatan keputusan dan pemeriksaan kapasitas saat penerapan promosi kini mencocokkan jumlah baris dari query dengan count pada filter yang sama. Keduanya memakai batas 10.000 baris per himpunan melalui CollectComplete.

Tes service menyimulasikan jumlah penempatan yang tidak cocok dan jumlah di atas batas. Kedua percobaan ditolak, lalu penerapan berhasil ketika data yang dilaporkan kembali lengkap. Tes service dan gRPC PostgreSQL sementara dengan race detector serta go vet bertag integration lulus.
