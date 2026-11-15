# Working History — Setup Schema & Migrasi Database IAM (`iam` & `iam_test`)

- **Tanggal/Waktu:** 2026-11-14 19:57:30 WIB
- **Scope:** `backend/services/iam/db`, `backend/services/iam/config`, `backend/services/iam/cmd`

## Ringkasan Pekerjaan

Menyelesaikan penyelarasan DDL PostgreSQL 18+ di `backend/services/iam/db/schema/schema.sql` dan migrasi baseline `000001` (`up.sql` & `down.sql`) dari `identity_access_management.dbml` (6 enum, 14 tabel IAM, dan trigger `set_updated_at()`). Mengimplementasikan `EnsureDatabase`, `DeriveTestDatabaseDSN`, `RunMigrationsUp`, `RunMigrationsDown`, `Init`, serta `InitDevAndTest` menggunakan `pgx/v5`, `gorm.io/driver/postgres`, dan `embed.FS`, serta mengintegrasikan `db.Init` pada startup `cmd/main.go`.

## Verifikasi

- `go mod tidy && go fmt ./... && go vet ./... && go build ./...` lulus tanpa error.
- `go test -race -v ./...` tanpa DSN melewati `TestEnsureAndMigrateDevAndTestDatabases` secara aman dan meluluskan unit test DSN/fail-fast.
- `go test -race -v ./...` dengan `IAM_DATABASE_DSN` aktif ke container PostgreSQL 18.6 (`127.0.0.1:6543`) berhasil membuat database `iam` dan `iam_test`, menguji rollback + re-apply `000001`, serta memverifikasi 15 tabel (14 IAM + `schema_migrations`) dan 6 enum identik di kedua database.
- Eksekusi langsung `schema.sql` pada `iam_test` (`ON_ERROR_STOP=1`) dan `diff -u` antara `schema.sql` dengan `000001_create_iam_schema.up.sql` bersih.
