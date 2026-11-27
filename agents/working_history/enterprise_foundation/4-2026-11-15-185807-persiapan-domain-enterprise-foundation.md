# 4. Persiapan Domain Murni Enterprise Foundation

**Tanggal:** 15 November 2026 18:58 WIB

## Ringkasan Pekerjaan

Menyelesaikan lapisan domain murni `backend/services/enterprise_foundation/domain/` sebelum integrasi `repository/pgstore` dan `services/*`:

- **`domain/errors`**: Menambahkan `FieldOf`, `newSentinel`, serta sentinel terpetakan (`address.go`, `legalentity.go`, `location.go`, `organizationunit.go`, `party.go`, `outbox.go`).
- **`domain/entity`**: Mengimplementasikan 9 enum plus `IAMScopeType()` dan 9 entitas (`Address`, `LegalEntity`, `Location`, `OrganizationUnit`, `Party`, `PartyRole`, `PartyContact`, `PartyAddress`, `OutboxEvent`) dengan invarian 1:1 terhadap skema PostgreSQL.
- **`domain/repository`**: Mendefinisikan `Transactor`, paginasi/filter ternormalisasi (`filter.go`), dan 9 antarmuka repository (`Select*`, `Insert`, `Update`, `SoftDeleteByID`/`DeleteByID`).
- **`domain/service`**: Menyusun `Clock`, `IDGenerator`, `HierarchyPolicy`, `ScopePolicy`, `PortalAccessPolicy`, dan `OutboxBuilder`.

## Verifikasi

- `go fmt ./...`, `go vet ./...`, `go build ./cmd/...`, dan `go test -race -count=1 ./domain/... ./config/... ./transport/...` lulus tanpa error (seluruh package `domain/*/test`, `config/test`, `transport/grpc/test` `PASS`).
