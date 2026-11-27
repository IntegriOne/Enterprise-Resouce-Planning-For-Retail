# Penguncian posisi pada penerapan penempatan primer

ApplyPrimaryAssignment mengunci posisi tujuan hingga transaksi worker selesai. Tes gRPC membuka transaksi perubahan periode posisi, mengirim RPC sebelum commit, lalu memverifikasi RPC menunggu dan menolak periode yang sudah tidak cocok. Tes security dan gRPC PostgreSQL sementara serta Redis TLS lulus dengan race detector dan go vet.
