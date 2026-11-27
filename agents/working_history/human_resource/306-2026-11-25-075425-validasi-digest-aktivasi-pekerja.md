# Validasi digest kontrak saat aktivasi pekerja

Aktivasi pekerja mensyaratkan digest SHA-256 sepanjang 32 byte pada bukti kontrak bertanda tangan. Tes regresi mencakup digest kosong, digest satu byte, dan bukti dengan panjang yang valid. Handoff DRM memuat bukti dokumen dan tanda tangan yang dibutuhkan RPC. `go vet ./...`, tes keamanan dengan race detector, dan integrasi gRPC personnel dengan PostgreSQL lulus.
