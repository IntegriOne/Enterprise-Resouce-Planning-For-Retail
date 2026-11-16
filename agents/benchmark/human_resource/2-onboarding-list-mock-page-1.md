# Benchmark daftar onboarding draf

- Waktu: 16 November 2026 WIB.
- Perintah: `go test ./internal/service/onboarding/test -run '^$' -bench . -benchmem -count=1`.
- Lingkungan: Linux amd64, AMD Ryzen 9 9900X 12-Core Processor.
- `BenchmarkServiceList/scoped-page-1-24`: 237334 iterasi, 4863 ns/op, 2612 B/op, 35 allocs/op.

Benchmark memakai mock repository dan halaman kecil. Angka ini hanya baseline biaya service/list dalam proses; tidak mengukur query PostgreSQL, gRPC, jaringan, atau skala data produksi.
