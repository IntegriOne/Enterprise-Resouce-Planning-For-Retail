# Waktu pembukaan requisition saat penjadwalan wawancara

Pembuatan dan perubahan draf wawancara menolak requisition dengan waktu pembukaan di masa depan meski statusnya open. Tes gRPC menyiapkan waktu buka mendatang, memverifikasi FailedPrecondition, lalu berhasil setelah waktu dikembalikan. Tes Recruitment dengan PostgreSQL sementara, race detector, dan go vet lulus.
