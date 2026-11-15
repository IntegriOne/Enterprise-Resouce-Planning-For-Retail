# Implementasi Lapisan Repository IAM Berbasis GORM PostgreSQL

- Menyelaraskan baseline skema `000001` (`must_change_password`, `key_version`, `sessions.mfa_verified`, `group_roles.expires_at`) dan memperluas port `domain/repository` (`InsertWithIdentities`, `SelectByEmailAndUsername`).
- Mengimplementasikan `repository/pgstore` (`Transactor`, pemetaan SQLSTATE `23505`/`23503`/`23001`/`23514`, helper query & waktu lokal OS) serta unit test di `repository/pgstore/test`.
- Mengimplementasikan seluruh 14 repositori GORM PostgreSQL pada `repository/user`, `repository/session`, `repository/permission`, `repository/role`, `repository/group`, dan `repository/registry.go`, termasuk pemeriksaan pra-insert `user_identities.username` dan `email` pada `InsertWithIdentities` agar tidak terjadi duplikasi pengguna.
- Verifikasi: `go fmt ./...`, `go vet ./...`, `go build ./...`, serta `go test -race -count=1 ./...` lulus untuk `domain/*`, `repository/pgstore/test`, dan unit test `db/`; nol kemunculan `utc` di `backend/`.
