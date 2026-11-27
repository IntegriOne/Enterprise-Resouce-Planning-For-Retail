# Working History IAM — Kontrak Eksplisit Blokir Login Akun System (`system cannot login`)

- **Tanggal:** 2026-11-25 04:40:00 WIB
- **Cakupan:** Domain & gRPC contract IAM (`domain/errors`, `domain/entity`, `contract/proto`, `services/user/auth`, `services/session`, `services/user/registration_runtime`), Gateway (`graphql/iam/iam_auth.graphql`, `source_connections/iam_client.go`, `adapter/iam/adapter.go`), dan tes unit/gRPC/E2E.

## Ringkasan Perubahan
1. Menetapkan sentinel `ErrSystemCannotLogin` (`KindForbidden`, code `"iam.user.system_cannot_login"`, pesan `"system cannot login"`) serta menegakkannya langsung di invarian domain `User.CanAuthenticate()`.
2. Memperbarui `AuthenticatePassword` agar meresolusi pengenal sebelum validasi password kosong: setiap percobaan login atas `UserKindSystem` (mis. `auto_system_iam`, `auto_system_ef`), dengan password kosong `""` maupun terisi, langsung mengembalikan `ErrSystemCannotLogin` (`"system cannot login"`), sementara akun manusia/tidak dikenal tetap mengembalikan `ErrInvalidInput` / `ErrInvalidCredentials`.
3. Menambahkan kontrak normatif pada `auth.proto`, `session.proto`, `user.proto`, `common.proto`, `buf generate`, dan `iam_auth.graphql`, memetakan `"iam.user.system_cannot_login"` di Gateway `normalizeAuthError`, serta memverifikasi `govulncheck` (`0` CVE aktif), unit/gRPC test, dan live HTTPS E2E.
