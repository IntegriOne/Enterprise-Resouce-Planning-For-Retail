# Working History Gateway #10: Penghapusan `registerCustomer`, Kebijakan `createUser` Publik/Login, dan `resendValidationEmailToken`

**Tanggal:** 2026-11-15 09:40 WIB

## Ringkasan Perubahan
- **Penghapusan `registerCustomer`:** Menghapus `RegisterCustomerInput`, `RegisterCustomerPayload`, `Mutation.registerCustomer` dari [`iam_auth.graphql`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/graphql/iam/iam_auth.graphql), adapter, dan resolver; skema digenerasi ulang dengan `go generate ./...`.
- **Kebijakan Baru `createUser`:** Menambahkan `redirectLocation: String` pada `CreateUserPayload` di [`iam_user.graphql`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/graphql/iam/iam_user.graphql). [`Adapter.CreateUser`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/user_admin_adapter.go) membuka akses publik (`AuthenticateOptional`) dengan rate limit auth, menyembunyikan `user` (`nil`) bila anonim atau gagal, mengembalikan `CONFLICT` `"username has been taken"` saat duplikat username, serta mengembalikan header `Location: /login` dan `redirectLocation: "/login"` (`was_successful: true`, `user: nil`) untuk duplikat email pada non-admin (`!isAdmin`).
- **`ResendValidationEmailToken`:** Memastikan error `"your email has been verified"` terpetakan sebagai `CONFLICT`.

## Verifikasi
- `go vet ./...`, `go vet -tags=e2e ./internal/e2e`, unit test `6` package bertest, dan live HTTPS E2E (`71` test / `141` subtest / `0` fail / `57.545s`) `PASS`.
