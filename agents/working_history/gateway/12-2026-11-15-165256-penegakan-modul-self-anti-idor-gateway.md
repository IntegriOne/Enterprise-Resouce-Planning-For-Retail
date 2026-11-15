# Penegakan Modul Self-Service (`*self`), Anti-IDOR, Anti-Escalation, & White-Box Test Gateway

Tanggal: 2026-11-15 16:52 WIB

## Ringkasan Perubahan

1. **Otorisasi `*self` & Zero-Oracle Anti-IDOR:** Menegakkan `RequireSelfPermission` dan `RequireSelfSensitive` ([authorizer.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/authorizer.go)) serta `requireSelfOrPerm` ([adapter.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/adapter.go)) untuk operasi mandiri, sambil mewajibkan izin staf IAM pada jalur lintas-user (`node`, `user`, nested resolvers, `sessions`, `revokeSession`) beserta proteksi zero-oracle dan anti-delegation escalation di [rbac_adapter.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/rbac_adapter.go).
2. **Perbaikan Layout & Penambahan Test:** Memindahkan white-box test `package <pkg>` ke direktori package induk di `backend/gateway/internal/{config,browserauth,middleware,server,adapter/iam}`, memperbarui fixture `*self`, serta menambah [self_authorizer_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/self_authorizer_test.go) (`10` subtest) dan [self_rbac_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/self_rbac_test.go).
3. **Verifikasi:** `go fmt`, `go vet -tags=e2e ./...`, unit test `-race -count=1` (`6` package bertest), dan E2E HTTPS penuh `TestSelfServiceModuleRBACAndAntiEscalation` (`74` top-level test, `146` subtest, `0` fail, `0` skip, `62.634s`) seluruhnya `PASS`.
