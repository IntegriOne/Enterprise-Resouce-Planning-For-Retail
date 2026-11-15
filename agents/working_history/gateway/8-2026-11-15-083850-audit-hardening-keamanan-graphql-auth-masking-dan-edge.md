# 8 - Audit & Hardening Keamanan GraphQL Gateway, Auth Masking, Caddy, dan Edge (`/golang-security`)

Tanggal: 2026-11-15 08:38 WIB

## Ringkasan Perubahan
- Mencegah kebocoran *clue* pada [`auth_adapter.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/auth_adapter.go) dengan `normalizeAuthError` (`AUTHENTICATION` seragam) dan batas `5` percobaan `mfa_challenge`, memblokir pembuatan `is_system=true` di `CreateRole`, penghapusan akun sendiri di `SoftDeleteUser`, serta IDOR non-superadmin di `RevokeUserRole`/`RevokeGroupRole`.
- Menambahkan batas kedalaman AST (`10`), alias (`25`), *fragment cycle*, default `mutation_ip`, `maxPaginationOffset` & pencegahan panik slice [`relay.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/relay.go), deduplikasi cookie [`effects.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/browserauth/effects.go), masking error JSON `"json request body could not be decoded"` di [`error_presenter.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/server/error_presenter.go), header keamanan HTTP [`server.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/server/server.go) & [`Caddyfile`](file:///home/muriarfad/Documents/projects/erp-retail/proxy/Caddyfile), dan validasi `.value` cookie Edge [`middleware.ts`](file:///home/muriarfad/Documents/projects/erp-retail/frontend/src/middleware.ts).
- Menambahkan tes regresi [`relay_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/relay_test.go), [`server_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/server/server_test.go), dan [`browserauth_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/browserauth/browserauth_test.go).

## Verifikasi
- `go test -race -count=1 ./...` di `backend/gateway` (`6` package bertest), `next typegen`, `lint`, `tsc`, `next build`, header Caddy/redirect `/identity`, serta E2E HTTPS (`71` test / `140` subtest, `57.660s`) seluruhnya `PASS`.
