# 11 - Audit & Hardening Keamanan Auth, RBAC, dan Reconciler IAM (`/golang-security`)

Tanggal: 2026-11-15 08:38 WIB

## Ringkasan Perubahan
- Menghilangkan *auth/timing oracle* pada [`AuthenticatePassword`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/auth.go), `VerifyMFAChallenge`, `ConfirmPasswordReset`, dan `ConfirmEmailVerificationByToken` lewat verifikasi Argon2id *dummy* konstan serta pemetaan error ambigu (`ErrInvalidCredentials`, `ErrTokenNotFound`, `VerifyEmailStatusInvalid`).
- Memperketat *fail-closed* keamanan pada [`services/registry.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/registry.go) (menghapus bypass keyring CSRF kosong), validasi HMAC `firstClientIP` & interceptor gRPC [`authcontext`](file:///home/muriarfad/Documents/projects/erp-retail/backend/shared/authcontext/context.go), imutabilitas peran sistem `ErrSystemRoleImmutable` pada `role/service.go`, serta mutex/panik/outbox `CountFailed` & `sync.WaitGroup` pada `reconciler.go`, `dispatcher.go`, `grpc/server.go`, dan `cmd/main.go`.
- Menambahkan uji regresi keamanan di [`role/test/service_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/role/test/service_test.go), [`user/test/auth_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/test/auth_test.go), dan [`user/test/mfa_identity_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/test/mfa_identity_test.go), serta mendokumentasikan audit DREAD di [`1-2026-11-15-083850-golang-security-iam-gateway-audit.md`](file:///home/muriarfad/Documents/projects/erp-retail/agents/audit/iam/1-2026-11-15-083850-golang-security-iam-gateway-audit.md).

## Verifikasi
- `go fmt ./... && go vet ./... && GOWORK=off go test -race -count=1 ./...` lulus untuk `shared/{authcontext,authsession,authz}` dan seluruh **20** package bertest `services/iam`.
- Rebuild container `iam` (`healthy`) serta live HTTPS E2E (`71` test top-level, `140` subtest, `57.660s`) seluruhnya `PASS`.
