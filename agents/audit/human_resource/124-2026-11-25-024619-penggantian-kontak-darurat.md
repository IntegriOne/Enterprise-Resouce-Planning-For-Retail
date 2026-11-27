# Penggantian kontak darurat

Tes gRPC mereproduksi waktu konfirmasi dari pihak lama terbawa saat field mask hanya mengganti referensi Party. Mapper kini mengosongkan waktu konfirmasi ketika Party berubah dan permintaan tidak menentukan nilai konfirmasi baru. Konfirmasi eksplisit tetap mengikuti validator waktu serta verifikasi Party. Tes Personnel dengan PostgreSQL sementara dan race detector lulus.
