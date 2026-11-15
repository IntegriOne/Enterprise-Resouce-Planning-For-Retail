# Laporan E2E Live HTTPS: Verifikasi RBAC & Sensitif Re-Auth pada `createEmployeeUser`, `createVendorUser`, dan `createSupplierUser`

- **Tanggal Verifikasi:** 2026-11-15 14:56 WIB
- **Target Endpoint:** `https://erp.localhost:8443/api/graphql` (Caddy TLS -> `gateway-1`/`gateway-2` -> `iam:9001` -> PostgreSQL `18.7` + Redis `8.10.3` TLS ACL)
- **Berkas Uji:** [`iam_operations_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go), [`live_https_e2e_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/live_https_e2e_test.go), [`create_business_user_rbac_test.go`](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/create_business_user_rbac_test.go)
- **Perintah Eksekusi:**
  ```bash
  cd /home/muriarfad/Documents/projects/erp-retail/backend/gateway && GOWORK=off go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e
  ```
- **Ringkasan Hasil:** **`PASS`** — **`72`** top-level test, **`146`** subtest, **`0` gagal**, **`0` skip**, durasi paket `64.286s`.

## 1. Status Lingkungan Compose & Database Live

Seluruh 8 container Compose berjalan dan sehat: `caddy`, `frontend-1`, `frontend-2`, `gateway-1`, `gateway-2`, `iam`, `postgres`, `redis`. Kondisi database `iam` pasca-pengujian bersih (`users=1`, `roles=4`, `groups=0`).

## 2. Matriks Skenario RBAC (`TestRBACCreateBusinessUsersStrictEnforcement`)

Seluruh 5 subtest dijalankan terhadap mutasi `createEmployeeUser`, `createVendorUser`, dan `createSupplierUser`:

| Subtest | Skenario & Pembuktian | Durasi | Status |
| :--- | :--- | :--- | :--- |
| `AnonymousRejectedAuthenticationAndZeroAntiOracle` | Pemanggil anonim dengan cookie/CSRF pre-auth ditolak `AUTHENTICATION` (`was_successful: false`, `user: null`), termasuk saat mencoba email terdaftar (tanpa oracle `CONFLICT`), dan `0` baris `user_identities` baru tersisip | `0.36s` | `PASS` |
| `CustomerEmployeeWithoutRoleAndReadOnlyRoleForbidden` | Pelanggan (`CUSTOMER`), `EMPLOYEE` tanpa role, dan `EMPLOYEE` dengan role bawaan `iam_readonly` (`GLOBAL`, tanpa `iam.users.write`) semuanya ditolak `AUTHORIZATION` dengan `0` baris baru di database | `1.12s` | `PASS` |
| `NonGlobalScopedUsersWriteForbidden` | Pemegang role kustom dengan `iam.users.write` pada scope `FACILITY` (bukan `GLOBAL`) ditolak `AUTHORIZATION` dengan `0` baris baru | `0.54s` | `PASS` |
| `DirectCustomRoleGlobalAllowedThenRevokedAndStaleBlocked` | Pemegang role kustom dengan `iam.users.write` scope `GLOBAL` dan sesi segar berhasil membuat user `EMPLOYEE`, `VENDOR`, `SUPPLIER` (`status: PENDING`); begitu `revokeRolePermission` dijalankan tanpa logout atau sesi dibuat kedaluwarsa (`>10` menit via `backdateSessionForStaleSensitiveTest`), permintaan berikutnya langsung ditolak `AUTHORIZATION` dengan `0` baris baru | `1.09s` | `PASS` |
| `GroupInheritedGlobalPermissionAllowedThenMemberRemovedBlocks` | Pewarisan `iam.users.write` scope `GLOBAL` melalui `group` + `assignGroupRole` berhasil membuat user `EMPLOYEE`, `VENDOR`, `SUPPLIER`; setelah `removeGroupMember` tanpa logout, permintaan berikutnya langsung ditolak `AUTHORIZATION` dengan `0` baris baru | `1.00s` | `PASS` |

## 3. Verifikasi Unit Test Gateway

`GOWORK=off go test -race -count=1 ./...` pada `backend/gateway` lulus di seluruh `6` package bertest, termasuk `TestCreateBusinessUser_RBACAndSensitiveReauth` (`13` subtest) yang menguji anonim, sesi revoked, customer, tanpa role, `iam_readonly`, scope non-global (`facility`), sesi sensitif basi (`>10` menit), MFA aktif belum terverifikasi, akun terblokir/tersuspensi, SuperAdmin MFA segar, role kustom `iam.users.write` `GLOBAL`, dan pencabutan izin real-time pasca-kenaikan `PermRevision`.
