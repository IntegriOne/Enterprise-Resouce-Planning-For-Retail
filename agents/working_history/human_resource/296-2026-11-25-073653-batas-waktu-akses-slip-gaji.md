# Batas waktu akses slip gaji

Query daftar dan detail membandingkan timestamp penerbitan dengan waktu HRM. Verifier yang mengembalikan waktu penerbitan di masa depan menyebabkan FailedPrecondition. Pembayaran yang masih terjadwal tetap didukung setelah slip terbit. Regresi gRPC PostgreSQL dengan race detector dan go vet lulus.
