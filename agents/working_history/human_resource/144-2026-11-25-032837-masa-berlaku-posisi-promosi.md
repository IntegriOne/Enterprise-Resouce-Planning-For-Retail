# Masa berlaku posisi hasil promosi

Penerapan promosi menolak posisi tujuan yang sudah berakhir pada hari penerapan. Penempatan primer hasil promosi mewarisi tanggal akhir posisi tujuan, dan pemeriksaan kapasitas berhenti pada tanggal tersebut. Validasi domain menolak bukti penempatan terbuka atau berakhir melewati posisi yang berbatas waktu.

Tes gRPC dengan PostgreSQL sementara memastikan tanggal akhir posisi tersimpan pada penempatan hasil promosi. Tes domain dan service dengan race detector serta go vet lulus.
