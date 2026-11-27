# Validasi ulang manajer saat penerapan penempatan

CreateWorkerAssignment, ApproveWorkerAssignment dan ApplyPrimaryAssignment memakai pemeriksaan manajer yang sama. Tes gRPC membuktikan bahwa perubahan tanggal mulai manajer setelah persetujuan menolak penerapan dan bahwa penerapan berhasil setelah fixture dipulihkan. Tes security serta gRPC PostgreSQL sementara dengan Redis TLS, race detector dan go vet lulus.
