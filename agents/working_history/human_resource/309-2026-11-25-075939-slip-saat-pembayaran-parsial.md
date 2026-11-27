# Slip saat pembayaran parsial

Penerbitan slip pekerja berstatus paid bisa diproses ketika run masih posted jika referensi pembayaran tersedia. Regresi gRPC Payroll membuktikan penolakan hasil paid tanpa referensi dan kelanjutan penerbitan setelah referensi ada. Tes PostgreSQL sementara dengan renderer PDF, race detector dan go vet lulus.
