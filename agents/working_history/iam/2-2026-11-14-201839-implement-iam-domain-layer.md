# Working History 2 — Implementasi Lapisan Domain IAM

- **Tanggal (WIB):** 2026-11-14 20:18:39
- **Ruang Lingkup:** [domain/](../../backend/services/iam/domain) (`errors`, `entity`, `repository`, `service`)

## Ringkasan Pekerjaan

Membangun seluruh lapisan domain Clean Architecture IAM menggunakan stdlib Go `1.26.8` secara TDD (*test-first* di subpaket `test/`) sesuai skema DBML/SQL 14 tabel dan kebijakan D01–D12:

1. **`domain/errors` (`package iamerrors`)**: kategori `Kind`, `Code`, `DomainError` (`Unwrap`, `Is`, `errors.AsType`), dan sentinel error domain pengguna, sesi, serta RBAC.
2. **`domain/entity`**: 6 enum PostgreSQL (`UserKind`, `UserStatus`, `IdentityKind`, `TokenPurpose`, `SessionEndReason`, `ScopeType`), validasi UUID/email/username/kode/path internal aman, 14 entitas beraturan skema, redaksi `String()`/`slog.LogValuer`, salinan defensif `[]byte`, anti-replay TOTP, serta `Touch()` clamping sesi `30m/8h`.
3. **`domain/repository`**: paginasi `PageRequest`/`PageResult[T]`, filter ternormalisasi, `Transactor`, dan 14 interface repository berprefiks operasi DB (`Select`/`Insert`/`Update`/`Delete`/`SoftDelete`/`Count`/`Check`).
4. **`domain/service`**: port kriptografi/waktu serta kebijakan `PasswordPolicy`, `RedirectResolver`, `SessionPolicy`, dan `AccessEvaluator` fail-closed.

Seluruh `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test -race -count=1 ./domain/...` lulus penuh tanpa error atau skip.
