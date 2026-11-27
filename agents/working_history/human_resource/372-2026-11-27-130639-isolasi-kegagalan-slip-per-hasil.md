# Isolasi kegagalan penerbitan slip

Baris hasil yang tidak cocok dengan run Finance ditandai gagal secara individual. HRM tetap memproses hasil lain dan menggabungkan error dengan referensi hasil untuk diagnosis internal. Pembatalan konteks menghentikan pemrosesan dan mempertahankan error yang sudah tercatat. Tes race, go vet, dan gRPC PostgreSQL lulus.
