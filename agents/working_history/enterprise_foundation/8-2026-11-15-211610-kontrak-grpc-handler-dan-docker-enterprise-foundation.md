# 8. Kontrak gRPC, Handler, dan Runtime Docker Enterprise Foundation

**Tanggal:** 15 November 2026 21:16 WIB

## Ringkasan Pekerjaan

Menyelesaikan kontrak Protobuf `enterprise_foundation.v1` dan menayangkan gRPC server Enterprise Foundation melalui Docker Compose:

- **Kontrak & Handler gRPC:** Menjalankan `buf lint && buf generate` dan mengimplementasikan mapper serta handler untuk seluruh 5 service (`AddressService`, `LegalEntityService`, `LocationService`, `OrganizationUnitService`, `PartyService`), lengkap dengan `RegisterAll` *fail-closed* ke `transport/grpc/server.go`, sanitasi error gRPC, serta pengujian `bufconn` penuh dan *nil-request guards*.
- **PostgreSQL & Docker Hardening:** Memperkuat `Dockerfile` (SHA256 Go 1.27.4, symlink timezone `Asia/Jakarta`, user non-root `10001`), menambah `02-init-enterprise-foundation.sql`, mengalihkan kepemilikan DB/tabel ke `enterprise_foundation_svc`, mengisolasi `.env` lokal via `.gitignore`, serta mendaftarkan `enterprise-foundation` (`127.0.0.1:9002:9002`, `read_only`, `cap_drop: ALL`, `no-new-privileges`) di `backend/docker-compose.yml`.

## Verifikasi

- Seluruh pengujian `go vet ./...` dan `go test -p 1 -race -count=1 ./...` dengan live PostgreSQL lulus 100%, container berstatus `healthy` tanpa error log, dan uji gRPC live ke `127.0.0.1:9002` sukses.
