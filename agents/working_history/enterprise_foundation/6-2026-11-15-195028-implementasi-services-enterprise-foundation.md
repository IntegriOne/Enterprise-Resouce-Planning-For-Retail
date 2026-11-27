# 6. Implementasi Application Service & Wiring Enterprise Foundation

**Tanggal:** 15 November 2026 19:50 WIB

## Ringkasan Pekerjaan

Menyelesaikan lapisan *application service* pada `backend/services/enterprise_foundation/services/` dan menghubungkannya dengan repositori serta runtime server:

- **Fondasi & Utilitas**: Menambahkan `pkg/uuid` (UUIDv4 *fail-closed*) dan `services/internal/servicekit` (`CollectAllPages` berbatas `MaxCollectAllItems` untuk `GetAll`, `EnsureAddressExists`, dan *helper* outbox).
- **5 Layanan Domain**: Mengimplementasikan CRUD penuh, pengecekan duplikasi sebelum `Insert` (`Find`/lookup sebelum pembuatan dan pembaruan), `List`, dan `GetAll` pada `address`, `legalentity`, `location`, `organizationunit`, serta `party` (`Party`, `PartyRole`, `PartyContact` + kebijakan portal IAM, dan `PartyAddress`).
- **Registri & Wiring**: Menambahkan `services.Registry` (`services/registry.go`), mengisi `Services` pada `transport/grpc/handler/register.go`, dan menghubungkan `repository.New` serta `services.NewRegistry` ke `cmd/wiring.go`.

## Verifikasi

- `go fmt ./...`, `go vet ./...`, `go build ./...`, serta `go test -p 1 -race -count=1 ./...` (unit test dan integrasi live PostgreSQL 18 pada `enterprise_foundation_test`) lulus seluruhnya (exit 0).
