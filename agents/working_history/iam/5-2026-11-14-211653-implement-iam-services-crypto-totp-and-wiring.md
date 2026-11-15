# Working History IAM #5 — Implementasi Services Layer, Crypto/TOTP Adapters & Repository Wiring

- **Tanggal:** 2026-11-14 21:16:53 WIB
- **Scope:** `backend/services/iam/pkg/{crypto,totp}`, `backend/services/iam/services/{permission,role,group,session,user}`, `services/registry.go`, `config/security.go`, `cmd/{main.go,wiring.go}`.

## Ringkasan Perubahan
- Mengimplementasikan port `domain/service` di `pkg/crypto` (UUIDv4, Argon2id PHC, token buram CSPRNG + digest SHA-256 + recovery codes, keyring AES-256-GCM berversi AAD) dan `pkg/totp` (RFC 6238 HMAC-SHA1).
- Mengimplementasikan 5 layanan domain (`permission`, `role`, `group`, `session`, `user`) memakai `Transactor.WithinTransaction`, kebijakan sesi 30m/8h & snapshot 30s fail-closed, anti-replay TOTP atomik, rotasi CSRF, serta pra-cek `SelectByEmailAndUsername` + `InsertWithIdentities` agar tidak ada duplikasi email/username.
- Merangkai `repository.New(gormDB)` ke `services.NewRegistryFromConfig` di `cmd/wiring.go` dan `cmd/main.go` bersama `config.LoadSecurity` fail-closed.
- Verifikasi `go mod tidy`, `go fmt`, `go vet`, `go build`, bebas UTC, serta 15 paket test `-race` lulus (`db.TestEnsureAndMigrateDevAndTestDatabases` skipped karena env DSN live belum diset).
