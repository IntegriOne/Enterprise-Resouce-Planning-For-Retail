# Working History IAM — Audit & Perbaikan Error Handling, Panic/Recover, dan RBAC Fail-Closed

- **Tanggal:** 25 November 2026 (`2026-11-25 05:13:49` WIB)
- **Cakupan:** `backend/services/iam` (domain, crypto/totp, config, db, repository, services, reconciler/projector/mailer, gRPC transport/interceptor, cmd) serta batas integrasi IAM pada `backend/gateway/internal/{source_connections,middleware,adapter/iam}`.

## Ringkasan Perubahan
1. Memperkuat *panic recovery* & siklus hidup: menambah `RecoveryStreamInterceptor`, fallback `slog.Default()`, level log `Error` vs `Warn` pada `LoggingUnaryInterceptor`, guard nil + `defer close(done)` + `recover()` di `Server`, refactor `cmd/main.go` ke `run()`, dan `errors.Join` cleanup pada `wiring.go`/`db`.
2. Menutup celah *fail-open* outbox/RBAC & *circuit breaker*: melindungi Gateway breaker dari `iam.scope.owner_verification_unavailable`, memperbaiki `iam.session.not_found`, menahan proyeksi Redis saat `GetRecoveryEpoch`/JSON/hex/`SessionRevoked`/baris sesi aktif rusak, serta memperbaiki reload user pascacommit dan `bootstrap.go`.
3. Menambahkan redaksi rahasia/PII (`String`/`GoString`/`LogValue`, DSN parse, email error body), batas `MaxRedirectPriority = 32767` + SQLSTATE `22003`, *single-handling* mailer, nil-guards Gateway, tes regresi baru, dan verifikasi penuh (`golangci-lint`, `-race`, `govulncheck`, Docker, live HTTPS E2E).
