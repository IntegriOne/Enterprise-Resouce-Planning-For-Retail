# Laporan E2E Live HTTPS: Kontrak `MutationResult`, `changeUsername`, dan `userResetPassword`

- **Tanggal Verifikasi:** 2026-11-15 07:42 WIB
- **Target Endpoint:** `https://erp.localhost:8443/api/graphql` (Caddy TLS -> `gateway-1`/`gateway-2` -> `iam:9001` -> PostgreSQL `18.7` + Redis `8.10.3` TLS ACL)
- **Berkas Uji:** [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go), [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go)
- **Perintah Eksekusi:**
  ```bash
  cd /home/muriarfad/Documents/projects/erp-retail/backend/gateway && GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e
  ```
- **Ringkasan Hasil:** **`PASS`** — **`71`** top-level test, **`140`** subtest, **`0` gagal**, **`0` skip**, durasi paket `57.428s`.

## 1. Status Lingkungan Compose & Database Live

Seluruh 8 container aktif dan berstatus sehat setelah rebuild image (`integrione-retail-iam:dev`, `integrione-retail-gateway:dev`, `integrione-retail-frontend:dev`):
- `caddy` (`Up`), `frontend-1`, `frontend-2`, `gateway-1`, `gateway-2`, `iam`, `postgres`, `redis` (`healthy`).
- Indeks unik parsial `user_identities` aktif pada database `iam` dan `iam_test`: `uq_user_identities_kind_value`, `uq_user_identities_one_pending_email_per_user`, `uq_user_identities_one_verified_email_per_user`, `uq_user_identities_one_username_per_user`.
- Kondisi database `iam` pasca-pengujian: `users = 1`, `roles = 4`, `groups = 0`, `active_sessions = 2`.

## 2. Rute Frontend & GraphQL Proxy Smoke Check

- `GET https://erp.localhost:8443/login` -> `HTTP 200`
- `GET https://erp.localhost:8443/register` -> `HTTP 200`
- `GET https://erp.localhost:8443/verify-email` -> `HTTP 200`
- `GET https://erp.localhost:8443/identity` (anonim) -> `HTTP 307`
- `GET https://erp.localhost:8443/api/graphql?query=%7Bhealth%7D` -> `HTTP 200` (`{"data":{"health":"ok"}}`)

## 3. Sorotan Skenario Baru & Diperbarui (`MutationResult` / `MutationError`)

| Test / Subtest | Skenario Terverifikasi | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `TestMutationChangeUsername/RenameReleasesOldUsernameAndAllowsReuseAndRejectsConflicts` | Ubah username `user1` ke `kibo`, verifikasi username sama -> `CONFLICT`, login username baru & gagal username lama, `user2` gagal pakai `kibo` (`CONFLICT`) namun berhasil mengambil username lama yang telah dilepas, serta customer tanpa username menambahkan username pertamanya | `0.91s` | `PASS` |
| `TestMutationChangeUsername/AnonymousRejected` | Pemanggil anonim ditolak (`UNAUTHENTICATED`) | `0.00s` | `PASS` |
| `TestMutationUserResetPassword/VerifiedEmailIssuesSingleActiveTokenAndUnverifiedOrGhostReturnsNeutralMessage` | Mutasi `userResetPassword` mengembalikan `was_successful: true` & `message` netral baik untuk email terverifikasi, pending, maupun non-existent tanpa membocorkan eksistensi akun | `0.59s` | `PASS` |
| `TestMutationUpdateUserEmailDeletePendingAndResend` | Kasus 1 (pending email tanpa memutus email aktif lama + resend + delete + verify) & Kasus 2 (`PENDING` resend) + anonim -> `ARGUMENT`/`AUTHENTICATION` via `errors { type message }` | `1.09s` | `PASS` |
| `TestLiveHTTPSIAMGatewayE2E` (`7` subtest) | Endpoint tunggal `/api/graphql`, CSRF/Origin, login & RBAC Test A/B (`zero DB mutation`), MFA TOTP/recovery code, rate limit, dan *fail-closed* + pemulihan Redis ACL / IAM | `20.22s` | `PASS` |
