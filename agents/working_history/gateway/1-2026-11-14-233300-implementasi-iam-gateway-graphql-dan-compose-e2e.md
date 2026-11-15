# 1 — Implementasi Penuh IAM–Gateway GraphQL, Stack Compose, Benchmark, dan E2E HTTPS

Menyelesaikan penuh layer IAM–Gateway pada endpoint tunggal `/api/graphql`:

- Menegakkan kontrak [iam.graphql](../../backend/gateway/graphql/iam.graphql) bertipe `snake_case`, varian `createUser` (`CUSTOMER`), `createEmployeeUser`, `createVendorUser`, `createSupplierUser`, serta mempertahankan 8 kontrak autentikasi browser frontend.
- Menyelesaikan stabilitas runtime HTTP/2 gRPC pada [server.go](../../backend/gateway/internal/server/server.go) dengan `Immutable: true` dan `strings.Clone` atas metadata request untuk menghilangkan panic `hpack` `fasthttp` zero-copy.
- Menyusun [backend/gateway/Dockerfile](../../backend/gateway/Dockerfile) dan [backend/docker-compose.yml](../../backend/docker-compose.yml) terpadu (8 kontainer sehat dengan `RestartCount=0` di balik Caddy HTTPS `https://erp.localhost:8443`).
- Memverifikasi `go test -race -count=1 ./...`, benchmark hot-path `internal/server`, dan seluruh 7 subtest E2E HTTPS [live_https_e2e_test.go](../../backend/gateway/internal/e2e/live_https_e2e_test.go) (`21.158s`), lalu memperbarui indeks `codebase-memory-mcp`.
