# Working History #11: Verifikasi Audit RBAC `createEmployeeUser`, `createVendorUser`, dan `createSupplierUser`

- **Waktu:** 2026-11-15 14:57:16 WIB
- **Domain:** GraphQL Gateway (`backend/gateway`) & IAM (`backend/services/iam`)

## Ringkasan Pekerjaan
1. Memastikan mutasi `createEmployeeUser`, `createVendorUser`, dan `createSupplierUser` melewati [`createUserByKind`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/user_admin_adapter.go#L590-L619) -> [`Authorizer.RequirePermission`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/authorizer.go#L40-L68) dengan syarat `iam.users.write` berlingkup `GLOBAL`, akun `ACTIVE`, serta re-autentikasi sensitif `<= 10` menit + MFA jika diaktifkan.
2. Menambahkan unit test [`create_business_user_rbac_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/create_business_user_rbac_test.go) (`13` subtest) serta memperbarui [`adapter_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/adapter_test.go) agar mock session memblokir token `Revoked` maupun status `Security` non-aktif seperti [`authsession.Store.ReadSessionContext`](file:///home/muriarfad/Documents/projects/erp-retail/backend/shared/authsession/store.go#L290-L340).
3. Menambahkan [`TestRBACCreateBusinessUsersStrictEnforcement`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go) (`5` subtest) dan mereset kunci rate-limit sebelum subtest `RegistrationVerificationPasswordResetAndRateLimit` di [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go).
4. Verifikasi: `go vet` & unit test `-race` (`6/6` package bertest), live HTTPS E2E (`72` test, `146` subtest, `0` fail/skip, `64.286s`), dan re-index `codebase-memory-mcp`.
