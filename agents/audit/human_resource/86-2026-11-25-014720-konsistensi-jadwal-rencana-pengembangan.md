# Konsistensi jadwal rencana pengembangan

Pembaruan tanggal pada draf sebelumnya dapat meninggalkan tindakan di luar periode rencana. Validator kini memeriksa seluruh halaman tindakan ketika jendela tanggal berubah dan sebelum persetujuan. Rencana tetap terkunci selama pemeriksaan. Mutasi tindakan mengambil lock rencana yang sama.

Tes gRPC ke PostgreSQL mereproduksi pemendekan yang keliru sebelum perbaikan. Setelah perbaikan, RPC menolak pemendekan dan pembacaan ulang mengonfirmasi tanggal semula. Tes integrasi Career, unit dengan race detector, dan go vet lulus.
