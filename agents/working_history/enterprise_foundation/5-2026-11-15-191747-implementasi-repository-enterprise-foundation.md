# 5. Implementasi Repository GORM PostgreSQL Enterprise Foundation

**Tanggal:** 15 November 2026 19:17 WIB

## Ringkasan Pekerjaan

Mengimplementasikan lapisan persistensi GORM PostgreSQL pada `backend/services/enterprise_foundation/repository/` beserta *lookup* bisnis di `domain/repository/`:

- **`domain/repository`**: Menambahkan *lookup* bisnis (`SelectByIDs`, `SelectByTaxIDNumber`, `SelectByNIB`, `SelectByLegalEntityAndKind`, `SelectByAddressID`, `SelectByHeadPersonID`, `SelectByPrimaryLocationID`, `SelectByEmail`, `SelectByRoleKind`, `SelectPrimaryByPartyID`, `SelectByPartyAndUsage`, `SelectPrimaryByPartyAndUsage`, `CountActiveDependents`, serta `Outbox.SelectByID`/`SelectByAggregate`/`DeleteEffectiveBefore`) dan perluasan filter (`Province`/`PostalCode`, `AddressID`).
- **`repository/pgstore`**: Membangun `Transactor` + `RegisterAfterCommit`, pemetaan galat PostgreSQL (`errors.go`), validasi UUID, *escape* `ILIKE`, dan paginasi (`query.go`).
- **`repository/*` & `Registry`**: Mengimplementasikan seluruh 9 repositori tabel (`address`, `legalentity`, `location`, `organizationunit`, `party/*`, `outbox`) dan `repository.New`.

## Verifikasi

- `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -p 1 -race -count=1 ./...` terhadap PostgreSQL 18 (`enterprise_foundation_test`) lulus seluruh paket uji (`PASS`, exit 0).
