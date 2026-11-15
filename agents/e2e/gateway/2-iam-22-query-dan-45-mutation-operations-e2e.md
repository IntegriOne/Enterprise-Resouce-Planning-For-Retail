# Laporan Pengujian E2E HTTPS Lengkap 22 Query dan 45 Mutation IAM–Gateway (`/api/graphql`)

## Ringkasan Eksekusi

- **Berkas Pengujian Tunggal:** [`iam_operations_e2e_test.go`](../../../backend/gateway/internal/e2e/iam_operations_e2e_test.go) (`//go:build e2e`, pemetaan ketat **1 fungsi = 1 operasi** `TestQuery*` / `TestMutation*`).
- **Perintah Verifikasi & Simpan Log Penuh:** `cd backend/gateway && go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/... > ../../agents/e2e/testing/iam/1-alloperations.md 2>&1` (atau relatif ke root repo: `agents/e2e/testing/iam/1-alloperations.md`).
- **Transparansi Request/Response & Validasi Skema Objek Penuh:** Seluruh pemanggilan `doGQLOK` / `doGQLErrorLabeled` mencetak `Query/Mutation`, `Input Variables`, `HTTP <status>`, dan `Result JSON` secara penuh. Seluruh seleksi GraphQL (termasuk mutasi setup dan query tindak lanjut) meminta field objek lengkap (`gql*Fields`) dan memvalidasi tiap field serta nullability terhadap `iam.graphql` melalui `assert*Object` / `assert*Payload`.
- **Verifikasi Eksekusi Mandiri (Sample 1:1):** `cd backend/gateway && go test -tags=e2e -v -count=1 -run '^(TestQueryRole|TestMutationVerifyCustomerEmail)$' ./internal/e2e/...` (`PASS`, `1.077s`).
- **Verifikasi Statis & Unit Gateway:** `gofmt -w ./internal/e2e/iam_operations_e2e_test.go`, `go vet -tags=e2e ./...` (lulus `exit 0` pasca pembaruan validator objek penuh), serta `go test -race -count=1 ./internal/adapter/... ./internal/middleware/... ./internal/server/... ./internal/config/...` — seluruh paket `PASS` pada baseline.
- **Hasil Keseluruhan E2E Terbaru (Full-Object + Full-Log):** **68/68 LULUS (`PASS`, `52.932s`, 0 gagal, 0 skip)** mencakup **22 Query (`TestQuery*`)**, **45 Mutation (`TestMutation*`)**, dan `TestLiveHTTPSIAMGatewayE2E`. Seluruh log tersimpan pada [`1-alloperations.md`](../testing/iam/1-alloperations.md) (`13174` baris, `0` secret `.env` terekspos). Matriks durasi per operasi pada tabel di bawah mencatat snapshot run awal (`48.601s`), sedangkan seluruh asersi objek lengkap terverifikasi pada run `52.932s`.
- **Kebersihan Database Pasca-Uji:** `users=1` (`SystemIAM`), `roles=4`, `groups=0` (seluruh data fixture dibersihkan via `t.Cleanup`).

## Matriks 22 Query IAM (`TestQuery*`)

| No | Fungsi Test & Durasi | Operasi GraphQL | Matriks Kasus & Verifikasi Non-Null / Keamanan | Hasil |
| --- | --- | --- | --- | --- |
| 1 | `TestQueryBrowserSession` (`0.32s`) | `browserSession` | Pre-auth anonim (`authenticated: false`, `csrfToken` non-null), terautentikasi (`authenticated`, `principalId`, `csrfToken` non-null), dan pasca-`userLogout` (`authenticated: false`). | PASS |
| 2 | `TestQueryMyWorkspaceAccess` (`0.29s`) | `myWorkspaceAccess` | SuperAdmin (`/identity`, `true`), Customer (`/`, `false`), anonim (`UNAUTHENTICATED`). | PASS |
| 3 | `TestQueryMyProfile` (`0.23s`) | `myProfile` | Profil `EMPLOYEE` pasca-login (seluruh field termasuk `last_login_at` non-null), anonim (`UNAUTHENTICATED`). | PASS |
| 4 | `TestQueryMyAccessSnapshot` (`0.29s`) | `myAccessSnapshot` | SuperAdmin (`roles` & `permissions` non-empty, `can_access_identity: true`), Customer (`can_access_identity: false`), anonim (`UNAUTHENTICATED`). | PASS |
| 5 | `TestQueryMySessions` (`0.29s`) | `mySessions` | Sesi aktif step-up (`mfa_verified_at` non-null) & riwayat sesi berakhir (`active_only: false`, `ended_at` & `end_reason=LOGOUT`), anonim (`UNAUTHENTICATED`). | PASS |
| 6 | `TestQueryUser` (`0.64s`) | `user(id)` | User valid pasca-login (`last_login_at` non-null), TestAB Customer (`FORBIDDEN`), UUID tak ada (`NOT_FOUND`). | PASS |
| 7 | `TestQueryUsers` (`0.55s`) | `users(...)` | Filter `query`/`kind`/`status`/`include_deleted` + paginasi + `deleted_at` pasca-soft-delete, TestAB Customer (`FORBIDDEN`). | PASS |
| 8 | `TestQueryUserIdentities` (`0.42s`) | `userIdentities` | 2 identitas terverifikasi (`EMAIL` & `USERNAME`, `verified_at` non-null), TestAB Customer (`FORBIDDEN`). | PASS |
| 9 | `TestQueryUserRoles` (`0.58s`) | `userRoles` | Assignment `FACILITY` + `scope.id` + `expires_at` + `granted_by` non-null, filter `include_expired: false/true`, TestAB Customer (`FORBIDDEN`). | PASS |
| 10 | `TestQueryUserEffectiveRoles` (`0.70s`) | `userEffectiveRoles` | Gabungan peran langsung dan warisan keanggotaan grup, TestAB Customer (`FORBIDDEN`). | PASS |
| 11 | `TestQueryUserGroups` (`0.47s`) | `userGroups` | Array kosong non-null `[]` saat belum tergabung & daftar grup non-null setelah masuk grup, TestAB Customer (`FORBIDDEN`). | PASS |
| 12 | `TestQueryUserSessions` (`0.52s`) | `userSessions` | Inspeksi admin atas sesi aktif (`active_only: true`) dan historis (`active_only: false`), TestAB Customer (`FORBIDDEN`). | PASS |
| 13 | `TestQueryRoles` (`0.55s`) | `roles(...)` | Filter `query`, `management_domain`, `active_only`, paginasi, TestAB Customer (`FORBIDDEN`). | PASS |
| 14 | `TestQueryRole` (`0.55s`) | `role(id)` | Detail role kustom non-null lengkap, TestAB Customer (`FORBIDDEN`). | PASS |
| 15 | `TestQueryRoleByCode` (`0.42s`) | `roleByCode` | Role sistem `superadmin` (`is_system: true`) & role kustom, TestAB Customer (`FORBIDDEN`). | PASS |
| 16 | `TestQueryRolePermissions` (`0.57s`) | `rolePermissions` | Array kosong non-null awal & izin terikat (`iam.users.read`), TestAB Customer (`FORBIDDEN`). | PASS |
| 17 | `TestQueryPermissions` (`0.29s`) | `permissions` | Filter `query: "iam."`, `resource: "users"`, `is_sensitive: true`, paginasi, TestAB Customer (`FORBIDDEN`). | PASS |
| 18 | `TestQueryPermission` (`0.29s`) | `permission(id)` | Pengambilan izin tunggal `iam.users.read` non-null, TestAB Customer (`FORBIDDEN`). | PASS |
| 19 | `TestQueryGroups` (`0.37s`) | `groups(...)` | Filter grup aktif & nonaktif (`active_only: true/false`), paginasi, TestAB Customer (`FORBIDDEN`). | PASS |
| 20 | `TestQueryGroup` (`0.37s`) | `group(id)` | Detail grup valid non-null, TestAB Customer (`FORBIDDEN`). | PASS |
| 21 | `TestQueryGroupMembers` (`0.64s`) | `groupMembers` | Paginasi anggota (`limit: 1`, `has_more: true`, `total: 2`), TestAB Customer (`FORBIDDEN`). | PASS |
| 22 | `TestQueryGroupRoles` (`0.53s`) | `groupRoles` | Role grup scoped `FACILITY` + `expires_at` + filter `include_expired: false/true`, TestAB Customer (`FORBIDDEN`). | PASS |

## Matriks 45 Mutation IAM (`TestMutation*`)

| No | Fungsi Test & Durasi | Operasi GraphQL | Matriks Kasus & Verifikasi Non-Null / Keamanan | Hasil |
| --- | --- | --- | --- | --- |
| 1 | `TestMutationUserLogin` (`0.61s`) | `userLogin` | Login via `email` & `username`, status `MFA_REQUIRED` (`mfaChallenge`/`mfaExpiresAt` non-null), penolakan `INVALID_CREDENTIALS`. | PASS |
| 2 | `TestMutationValidateUserMfa` (`0.38s`) | `validateUserMfa` | Verifikasi TOTP 6 digit & `recoveryCode` (`csrfToken`/`redirectLocation` non-null), proteksi replay `recoveryCode`. | PASS |
| 3 | `TestMutationRegisterCustomer` (`0.22s`) | `registerCustomer` | Pendaftaran sukses (`verificationPendingEmail`, `requiresEmailVerification: true`), konflik duplikasi email (`CONFLICT`). | PASS |
| 4 | `TestMutationVerifyCustomerEmail` (`0.49s`) | `verifyCustomerEmail` | Token nyata (`VERIFIED`/`verified: true` + login berhasil), `ALREADY_USED`, `EXPIRED`, `INVALID`. | PASS |
| 5 | `TestMutationRequestPasswordReset` (`0.23s`) | `requestPasswordReset` | Pengembalian `true` non-null untuk akun valid & non-existent (anti-enumerasi). | PASS |
| 6 | `TestMutationResetPassword` (`0.33s`) | `resetPassword` | Token reset nyata (`true` + login password baru `AUTHENTICATED`), penolakan token invalid. | PASS |
| 7 | `TestMutationUserLogout` (`0.24s`) | `userLogout` | Pengembalian `true`, invalidasi sesi (`UNAUTHENTICATED`), penolakan anonim (`UNAUTHENTICATED`). | PASS |
| 8 | `TestMutationUpdateMyProfile` (`0.24s`) | `updateMyProfile` | Pembaruan `display_name`, `avatar_url`, `locale`, `timezone` non-null, penolakan anonim (`UNAUTHENTICATED`). | PASS |
| 9 | `TestMutationChangeMyPassword` (`0.48s`) | `changeMyPassword` | Ubah password berhasil (`true`), wrong current (`INVALID_CREDENTIALS`), sesi basi `>10m` (`FORBIDDEN`). | PASS |
| 10 | `TestMutationVerifyMySessionMfa` (`0.24s`) | `verifyMySessionMfa` | Step-up TOTP (`true`), rotasi `X-CSRF-Token`, set `mfa_verified_at` non-null, penolakan kode salah. | PASS |
| 11 | `TestMutationEnrollMyTotp` (`0.46s`) | `enrollMyTotp` | Penerbitan `factor_id` & `otp_auth_uri` (`otpauth://totp/...`), penolakan sesi basi `>10m` (`FORBIDDEN`). | PASS |
| 12 | `TestMutationConfirmMyTotp` (`0.26s`) | `confirmMyTotp` | Konfirmasi `enabled: true` + `recovery_codes` non-empty, penolakan kode TOTP salah. | PASS |
| 13 | `TestMutationDisableMyTotp` (`0.26s`) | `disableMyTotp` | Penolakan tanpa step-up MFA (`FORBIDDEN`), berhasil setelah `verifyMySessionMfa` (`true` & `mfa_enabled: false`). | PASS |
| 14 | `TestMutationRegenerateMyRecoveryCodes` (`0.33s`) | `regenerateMyRecoveryCodes` | Penerbitan kode pemulihan baru non-null & invalidasi recovery code lama pada login berikutnya. | PASS |
| 15 | `TestMutationCreateUser` (`0.42s`) | `createUser` | Pembuatan `CUSTOMER` lengkap field `snake_case`, TestAB Customer (`FORBIDDEN`). | PASS |
| 16 | `TestMutationCreateEmployeeUser` (`0.47s`) | `createEmployeeUser` | Pembuatan `EMPLOYEE` & verifikasi login `AUTHENTICATED`, TestAB Customer (`FORBIDDEN`). | PASS |
| 17 | `TestMutationCreateVendorUser` (`0.47s`) | `createVendorUser` | Pembuatan `VENDOR` & verifikasi login `AUTHENTICATED`, TestAB Customer (`FORBIDDEN`). | PASS |
| 18 | `TestMutationCreateSupplierUser` (`0.46s`) | `createSupplierUser` | Pembuatan `SUPPLIER` & verifikasi login `AUTHENTICATED`, TestAB Customer (`FORBIDDEN`). | PASS |
| 19 | `TestMutationActivateUser` (`0.47s`) | `activateUser` | Aktivasi kembali user `SUSPENDED -> ACTIVE` + login berhasil, TestAB Customer (`FORBIDDEN`). | PASS |
| 20 | `TestMutationSuspendUser` (`0.47s`) | `suspendUser` | Perubahan status `SUSPENDED` & invalidasi sesi aktif (`UNAUTHENTICATED`), TestAB Customer (`FORBIDDEN`). | PASS |
| 21 | `TestMutationDisableUser` (`0.47s`) | `disableUser` | Perubahan status `DISABLED` & invalidasi sesi aktif (`UNAUTHENTICATED`), TestAB Customer (`FORBIDDEN`). | PASS |
| 22 | `TestMutationSoftDeleteUser` (`0.43s`) | `softDeleteUser` | Penghapusan lunak (`true` + `deleted_at` non-null pada `users(include_deleted: true)`), TestAB Customer (`FORBIDDEN`). | PASS |
| 23 | `TestMutationAddUserIdentity` (`0.44s`) | `addUserIdentity` | Penambahan identitas `USERNAME` sekunder non-null & konflik duplikasi (`CONFLICT`), TestAB Customer (`FORBIDDEN`). | PASS |
| 24 | `TestMutationSetPrimaryUserIdentity` (`0.43s`) | `setPrimaryUserIdentity` | Pergantian `is_primary: true` antar identitas terverifikasi, penolakan identitas unverified, TestAB Customer (`FORBIDDEN`). | PASS |
| 25 | `TestMutationRemoveUserIdentity` (`0.41s`) | `removeUserIdentity` | Penghapusan identitas sekunder (`true`), proteksi identitas primary, TestAB Customer (`FORBIDDEN`). | PASS |
| 26 | `TestMutationCreateRole` (`0.41s`) | `createRole` | Pembuatan role kustom non-null, penolakan redirect eksternal & duplikasi (`CONFLICT`), TestAB Customer (`FORBIDDEN`). | PASS |
| 27 | `TestMutationUpdateRoleDetails` (`0.53s`) | `updateRoleDetails` | Pembaruan `name`, `description`, `management_domain`, `default_redirect_path`, `redirect_priority`, TestAB Customer (`FORBIDDEN`). | PASS |
| 28 | `TestMutationActivateRole` (`0.55s`) | `activateRole` | Pengaktifan role nonaktif (`is_active: true`), TestAB Customer (`FORBIDDEN`). | PASS |
| 29 | `TestMutationDeactivateRole` (`0.53s`) | `deactivateRole` | Penonaktifan role kustom (`is_active: false`), proteksi role sistem `superadmin`, TestAB Customer (`FORBIDDEN`). | PASS |
| 30 | `TestMutationDeleteRole` (`0.55s`) | `deleteRole` | Penghapusan role kustom (`true` -> `NOT_FOUND`), proteksi role sistem `superadmin`, TestAB Customer (`FORBIDDEN`). | PASS |
| 31 | `TestMutationAssignRolePermission` (`0.55s`) | `assignRolePermission` | Pengikatan `iam.users.read` non-null (`role_id`, `permission_id`, `created_at`), TestAB Customer (`FORBIDDEN`). | PASS |
| 32 | `TestMutationRevokeRolePermission` (`0.54s`) | `revokeRolePermission` | Pencabutan izin peran (`true` & `rolePermissions` kosong), TestAB Customer (`FORBIDDEN`). | PASS |
| 33 | `TestMutationAssignUserRole` (`0.79s`) | `assignUserRole` | Pengikatan role user scope `GLOBAL` dan `FACILITY` + `scope_id` + `expires_at` + `granted_by` non-null, TestAB Customer (`FORBIDDEN`). | PASS |
| 34 | `TestMutationRevokeUserRole` (`0.56s`) | `revokeUserRole` | Pencabutan role user (`true` & `userRoles` kosong), TestAB Customer (`FORBIDDEN`). | PASS |
| 35 | `TestMutationCreateGroup` (`0.34s`) | `createGroup` | Pembuatan grup non-null & konflik duplikasi (`CONFLICT`), TestAB Customer (`FORBIDDEN`). | PASS |
| 36 | `TestMutationUpdateGroupDetails` (`0.38s`) | `updateGroupDetails` | Pembaruan `name` & `description` grup non-null, TestAB Customer (`FORBIDDEN`). | PASS |
| 37 | `TestMutationActivateGroup` (`0.38s`) | `activateGroup` | Pengaktifan grup nonaktif (`is_active: true`), TestAB Customer (`FORBIDDEN`). | PASS |
| 38 | `TestMutationDeactivateGroup` (`0.38s`) | `deactivateGroup` | Penonaktifan grup (`is_active: false`), TestAB Customer (`FORBIDDEN`). | PASS |
| 39 | `TestMutationDeleteGroup` (`0.37s`) | `deleteGroup` | Penghapusan grup (`true` -> `NOT_FOUND`), TestAB Customer (`FORBIDDEN`). | PASS |
| 40 | `TestMutationAddGroupMember` (`0.49s`) | `addGroupMember` | Penambahan anggota grup (`group_id`, `user_id`, `created_at` non-null), TestAB Customer (`FORBIDDEN`). | PASS |
| 41 | `TestMutationRemoveGroupMember` (`0.52s`) | `removeGroupMember` | Penghapusan anggota grup (`true` & `groupMembers` kosong), TestAB Customer (`FORBIDDEN`). | PASS |
| 42 | `TestMutationAssignGroupRole` (`0.73s`) | `assignGroupRole` | Pengikatan role grup `FACILITY` + `scope.id` + `expires_at` non-null & pewarisan ke `userEffectiveRoles`, TestAB Customer (`FORBIDDEN`). | PASS |
| 43 | `TestMutationRevokeGroupRole` (`0.46s`) | `revokeGroupRole` | Pencabutan role grup (`true` & `groupRoles` kosong), TestAB Customer (`FORBIDDEN`). | PASS |
| 44 | `TestMutationRevokeSession` (`0.48s`) | `revokeSession` | Pencabutan sesi target (`true` -> `UNAUTHENTICATED`, `ended_at` & `end_reason=ADMIN_REVOKED`), TestAB Customer (`FORBIDDEN`). | PASS |
| 45 | `TestMutationRevokeAllUserSessions` (`0.53s`) | `revokeAllUserSessions` | Pencabutan serentak multi-sesi (`count >= 2` non-null -> semua klien `UNAUTHENTICATED`), TestAB Customer (`FORBIDDEN`). | PASS |
