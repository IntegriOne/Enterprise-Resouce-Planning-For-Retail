# Working History: Persiapan Folder & Scaffold Dasar gRPC Server Enterprise Foundation

Menyiapkan struktur direktori penuh dan fondasi gRPC server runnable pada `backend/services/enterprise_foundation` selaras pola IAM:

- **Modul & Workspace:** Mendaftarkan `./services/enterprise_foundation` di `backend/go.work`, membuat `go.mod` (`go 1.27.4`), `Makefile`, `Dockerfile`, `.dockerignore`, dan `.env.example` (port default `9002`).
- **Kontrak Protobuf:** Menambahkan `contract/{buf.yaml,buf.gen.yaml,generate.go}` dan `common.proto` (`enterprise_foundation.v1`) berisi 9 enum DBML dan metadata paginasi, ter-generate ke `common.pb.go`.
- **Transport, Error, Config & Entrypoint:** Mengimplementasikan `config`, `domain/errors` (`eferrors`), `transport/grpc` (`Server`, `RecoveryUnaryInterceptor`, `InternalContextUnaryInterceptor`, `LoggingUnaryInterceptor`, `ErrorMapper`, `RegisterAll`, `HealthServer`), serta `cmd/{main.go,wiring.go,healthcheck/main.go}`.
- **Struktur Paket:** Membuat kerangka `db/`, `domain/`, `infrastructure/`, `pkg/`, `repository/`, dan `services/`.
- **Verifikasi:** `buf lint`, `buf generate`, `go fmt ./...`, `go vet ./...`, `go build ./cmd`, `go build ./cmd/healthcheck`, dan `go test -v -race -count=1 ./...` lulus.
