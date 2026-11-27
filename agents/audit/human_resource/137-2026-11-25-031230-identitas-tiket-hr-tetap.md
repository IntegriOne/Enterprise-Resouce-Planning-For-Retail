# Identitas tiket HR

Tes gRPC mereproduksi perubahan nomor dan idempotency key tiket terbuka yang masing-masing menulis outbox. Validator kini mempertahankan kedua field setelah pembuatan. Integrasi memastikan mutasi itu ditolak, perubahan tenggat yang sah tetap berjalan, dan jumlah event sesuai workflow yang berhasil. Tes HR Service dengan PostgreSQL sementara dan race detector serta vet modul bertag integration lulus.
