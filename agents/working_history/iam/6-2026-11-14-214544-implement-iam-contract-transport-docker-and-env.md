# 6 — Implementasi Kontrak Protobuf, Handler gRPC, Wiring `main`, Docker, dan `.env` IAM

Menyelesaikan lapisan kontrak, transport gRPC, wiring aplikasi utama, containerisasi Docker, dan konfigurasi environment lokal untuk `backend/services/iam`:

- **Kontrak Protobuf (`contract/`)**: `buf.yaml`, `buf.gen.yaml`, `generate.go`, dan 7 file `.proto` di `contract/proto/iam/v1/` (6 service `iam.v1`, 74 RPC), ter-lint (`buf lint`) dan ter-generate ke `contract/gen/iam/v1/`.
- **Transport gRPC (`transport/grpc/handler/`)**: `ErrorMapper` (`errors.go`), mapper OS local time (`.Local()`, nol `.UTC()`), `RegisterAll`, dan 6 handler gRPC dengan test `bufconn` penuh (`handler/test/`) mencakup 74 RPC.
- **Wiring & Healthcheck (`cmd/`)**: `wireGRPCServer` di `cmd/wiring.go` ter-wire ke `cmd/main.go`, `cmd/wiring_test.go`, serta CLI `cmd/healthcheck/`.
- **Docker & Environment**: `Dockerfile` (`ubuntu:26.10` + Go `1.26.8` + `tzdata` `Asia/Jakarta` + user `10001`), `.dockerignore`, `docker-compose.yml` (`iam-postgres` `postgres:latest` PG 18 & `iam` gRPC `:9001`), `.env.example`, `Makefile`, dan `.env` (`0600`) terisi DB DSN/port, gRPC URL/port, `IAM_SYSTEM_*`, dan security keys.
- **Verifikasi**: `docker compose up -d --wait` (`healthy`), live `grpcurl` `SERVING` + E2E auth/authz ke Postgres, `go vet`, `go build`, dan 17 paket `go test -race -count=1 ./...` lulus penuh tanpa UTC.
