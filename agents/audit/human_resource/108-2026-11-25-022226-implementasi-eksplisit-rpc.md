# Implementasi eksplisit RPC HRM

Server untuk 22 layanan privat kini memakai interface opt-out yang dihasilkan protoc-gen-go-grpc. Kompilasi gagal bila ada RPC kontrak yang belum memiliki method konkret. Seluruh handler berhasil dikompilasi dan tes registrasi terhadap descriptor tetap lulus.

Verifikasi mencakup `go test -race ./...`, `go vet ./...`, `git diff --check`, serta tes integrasi HR Service dengan PostgreSQL sementara. Kelengkapan method di tingkat kontrak tidak membuktikan seluruh jalur bisnis atau integrasi produksi.
