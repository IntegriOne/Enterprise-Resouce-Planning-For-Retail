# Benchmark Gateway Hot-Path Middleware dan Handler `/api/graphql`

## Metadata Eksekusi

- **Modul:** `github.com/integrione/erp-retail/backend/gateway` (`./internal/server/...`)
- **Sumber Benchmark:** [server_bench_test.go](../../backend/gateway/internal/server/server_bench_test.go)
- **Lingkungan Pengukuran:** `linux/amd64`, CPU `AMD Ryzen 9 9900X 12-Core Processor` (`-bench=. -benchmem`, logging dialihkan ke `io.Discard` via `newSilentBenchLogger`)
- **Perintah:** `cd backend/gateway && go test -run=^$ -bench=. -benchmem ./internal/server/...`
- **Status Kode Saat Ukur:** Sesudah perbaikan akhir (`Immutable: true` + `strings.Clone` di `server.go`), seluruh unit test `-race -count=1` paket Gateway lulus (`100% pass`).

## Hasil Benchmark

| Nama Benchmark | Iterasi (`N`) | Waktu / Op | Alokasi Memori / Op | Jumlah Alokasi / Op |
| --- | ---: | ---: | ---: | ---: |
| `BenchmarkVerifyCSRFTokenDigest-24` | `6027196` | `193.1 ns/op` | `240 B/op` | `5 allocs/op` |
| `BenchmarkOriginValidator-24` | `7656009` | `150.8 ns/op` | `176 B/op` | `2 allocs/op` |
| `BenchmarkAuthorizerSnapshotHit-24` | `29419807` | `39.71 ns/op` | `0 B/op` | `0 allocs/op` |
| `BenchmarkHTTPGraphQLHealth-24` | `63864` | `19303 ns/op` | `22292 B/op` | `187 allocs/op` |
| `BenchmarkHTTPGraphQLCookieQuery-24` | `35928` | `34268 ns/op` | `36827 B/op` | `370 allocs/op` |

## Ringkasan Pengamatan

- Fast-path evaluasi izin snapshot ter-cache (`BenchmarkAuthorizerSnapshotHit-24`) berjalan konstan di `39.71 ns/op` dengan **0 B/op dan 0 allocs/op**.
- Validasi HMAC-SHA256 digest CSRF (`193.1 ns/op`) dan validasi origin (`150.8 ns/op`) tetap beroverhead sub-mikrodetik.
- Pemrosesan penuh Fiber v3 + `adaptor.HTTPHandler` + `gqlgen` kompleksitas/guard untuk kueri `{ health }` dan kueri cookie-authenticated masing-masing stabil di `19.3 µs/op` dan `34.3 µs/op`.
