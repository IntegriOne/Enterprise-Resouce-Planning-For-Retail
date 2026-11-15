# 7 - Kontrak `MutationResult`, `changeUsername`, dan `userResetPassword` IAM Gateway

Tanggal: 2026-11-15 07:29 WIB

## Ringkasan Perubahan
- Menyeragamkan seluruh payload mutasi IAM agar mengimplementasikan `MutationResult` (`errors: [MutationError!]!`, `was_successful: Boolean!`) dengan `MutationError { type: ErrorTypeEnum! message: String! }`.
- Mengganti mutasi reset kata sandi menjadi `userResetPassword(input: UserResetPasswordInput!): UserResetPasswordPayload!` dan menambahkan `changeUsername(input: UsernameInput!): UserNamePayload!` lengkap pada skema GraphQL, `ErrorPresenter`, pemetaan error gRPC `iam_client.go` (termasuk `iam.role.system_immutable`), adapter, dan resolver.
- Memperbarui seluruh unit test Gateway (`adapter/iam`, termasuk `TestToMutationErrorsReasonMappings`) serta seluruh `TestQuery*` dan `TestMutation*` di [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go) dan [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go), termasuk `TestMutationChangeUsername` dan `TestMutationUserResetPassword`.

## Verifikasi
- `go fmt ./... && go vet ./... && go test -race -count=1 ./... && go vet -tags=e2e ./... && go test -tags=e2e -c -o /dev/null ./internal/e2e` lulus dengan kode keluar `0`.
- Rebuild `iam`, `gateway-1`/`gateway-2`, `frontend-1`/`frontend-2` sukses (`healthy`) dan `GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e` lulus (`71` test top-level, `140` subtest, `0` gagal, `57.428s`).
