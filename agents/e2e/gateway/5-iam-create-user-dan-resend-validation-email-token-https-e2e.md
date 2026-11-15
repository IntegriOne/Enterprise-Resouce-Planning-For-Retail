# Laporan E2E Live HTTPS: Penghapusan `registerCustomer`, Migrasi `createUser`, dan `resendValidationEmailToken`

- **Tanggal Verifikasi:** 2026-11-15 09:40 WIB
- **Target Endpoint:** `https://erp.localhost:8443/api/graphql` (Caddy TLS -> `gateway-1`/`gateway-2` -> `iam:9001` -> PostgreSQL `18.7` + Redis `8.10.3` TLS ACL)
- **Berkas Uji:** [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go), [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go)
- **Perintah Eksekusi:**
  ```bash
  cd /home/muriarfad/Documents/projects/erp-retail/backend/gateway && GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e
  ```
- **Ringkasan Hasil:** **`PASS`** — **`71`** top-level test, **`141`** subtest, **`0` gagal**, **`0` skip**, durasi paket `57.545s`.

## 1. Status Lingkungan Compose & Database Live

Seluruh 8 container Compose aktif dan sehat setelah rebuild image (`integrione-retail-iam:dev`, `integrione-retail-gateway:dev`, `integrione-retail-frontend:dev`) serta restart `caddy`:
- `caddy` (`Up`), `frontend-1`, `frontend-2`, `gateway-1`, `gateway-2`, `iam`, `postgres`, `redis` (`healthy`).
- Indeks unik parsial `user_identities` aktif pada database `iam`: `uq_user_identities_kind_value`, `uq_user_identities_one_pending_email_per_user`, `uq_user_identities_one_verified_email_per_user`, `uq_user_identities_one_username_per_user`.
- Kondisi database `iam` pasca-pengujian: `users=1`, `roles=4`, `groups=0`.

## 2. Rute Frontend & GraphQL Proxy Smoke Check

- `GET https://erp.localhost:8443/login` -> `HTTP 200`
- `GET https://erp.localhost:8443/register` -> `HTTP 200` (`id="reg-email"`, `id="reg-username"`, `id="reg-password"` ter-render; header `X-Powered-By` absen)
- `GET https://erp.localhost:8443/verify-email` -> `HTTP 200`
- `GET https://erp.localhost:8443/identity` (anonim) -> `HTTP 307`
- `GET https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D` -> `HTTP 200` (`{"data":{"health":"ok"}}`)

## 3. Sorotan Skenario Terverifikasi (`createUser` & `resendValidationEmailToken`)

| Test / Subtest | Skenario Terverifikasi | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `TestMutationCreateUser/PublicCreateUserAndAntiOracleBehavior` | Pendaftaran publik anonim mengembalikan `was_successful: true`, `user: nil`, `redirectLocation: nil`, dan menerbitkan 1 token `email_verification` aktif; duplikat username mengembalikan `CONFLICT` `"username has been taken"`; duplikat email (dengan username baru) mengembalikan `was_successful: true`, `user: nil`, `redirectLocation: "/login"`, HTTP header `Location: /login`, tanpa membuat user baru; duplikat email + username mengutamakan `"username has been taken"` (anti-oracle email) | `0.48s` | `PASS` |
| `TestMutationCreateUser/AuthenticatedCustomerAndAdminCreateUser` | Pelanggan terautentikasi maupun Super Admin menerima `user` non-null saat membuat user baru; duplikat email bagi pelanggan non-admin (`!isAdmin`) dialihkan diam-diam ke `/login` sedangkan admin berizin `iam.users.write` tetap menerima `CONFLICT` | `0.53s` | `PASS` |
| `TestMutationResendValidationEmailTokenAlreadyVerifiedAndNeutral/LoggedInVerifiedWithoutPendingReturnsConflictAndCreatesNoToken` | Pengguna terautentikasi tanpa pending email yang email-nya sudah terverifikasi menerima `CONFLICT` `"your email has been verified"` dan tidak membuat token verifikasi baru | `0.41s` | `PASS` |
| `TestMutationResendValidationEmailTokenAlreadyVerifiedAndNeutral/UnauthenticatedAlreadyVerifiedEmailRemainsNeutral` | Pemanggil tidak login pada email terverifikasi tetap netral (`is_resend: true`, `was_successful: true`, anti-enumerasi) | `0.24s` | `PASS` |
| `TestMutationVerifyEmail/VerifiedAlreadyUsedExpiredAndInvalid` | Registrasi publik melalui `createUser` -> token valid (`VERIFIED`), ulangi (`ALREADY_USED`), kedaluwarsa (`EXPIRED`), dan token tidak dikenal (`INVALID`) | `0.48s` | `PASS` |
| `TestMutationUserResetPassword/VerifiedEmailIssuesSingleActiveTokenAndUnverifiedOrGhostReturnsNeutralMessage` | Registrasi pending melalui `createUser` tetap netral pada `userResetPassword` tanpa menerbitkan token reset | `0.56s` | `PASS` |
| `TestLiveHTTPSIAMGatewayE2E` (`7` subtest) | Endpoint tunggal `/api/graphql`, bootstrap CSRF/Origin, login & RBAC Test A/B, MFA TOTP, registrasi via `createUser` (`user: nil`) + verifikasi + reset + rate limit, serta *fail-closed* Redis ACL (`18.64s`) | `20.23s` | `PASS` |
