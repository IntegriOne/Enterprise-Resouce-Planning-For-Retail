# Working History: Implementasi Lapisan Database & Migrasi Enterprise Foundation

Mengimplementasikan fondasi database PostgreSQL 18 dan runner migrasi tersemat pada `backend/services/enterprise_foundation/db`:

- **Skema & Migrasi Baseline:** Membuat `migrations/000001_create_enterprise_foundation_schema.{up,down}.sql` dan `schema/schema.sql` identik dengan `enterprise_foundation.dbml` (ekstensi `pgcrypto` + `pg_trgm`, 9 enum, 9 tabel, 12 FK, CHECK constraints, partial unique index soft-delete, GIN trigram `parties.display_name`, dan trigger `set_updated_at()`).
- **Bootstrap & Runner Go:** Menambahkan `database.go` (`EnsureDatabase`, `DeriveTestDatabaseDSN`), `init.go` (`Init` dengan pool + guard `_test`, `InitDevAndTest`), serta `migrate.go` (`RunMigrationsUp`, `RunMigrationsDown`, helper `withMigrationLock` dengan PostgreSQL advisory lock `718291824` dan `schema_migrations`).
- **Wiring Startup:** Menghubungkan `efdb.Init` pada `cmd/wiring.go` (wajib di `production`, opsional di `development`).
- **Verifikasi:** `go fmt`, `go vet`, `go build`, dan `go test -v -race -count=1 ./...` terhadap database `enterprise_foundation` & `enterprise_foundation_test` seluruhnya lulus.
