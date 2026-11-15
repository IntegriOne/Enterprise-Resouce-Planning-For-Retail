# Laporan E2E: Penolakan Layanan Self-Service Selama Status PENDING & Belum Terverifikasi

- **Tanggal:** 15 November 2026 (17:29 WIB)
- **Lingkungan:** `https://erp.localhost:8443/api/graphql` (8 container Compose aktif & healthy, termasuk image terbaru `iam`)
- **Suite Dijalankan:** `GOWORK=off go test -C backend/gateway -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...`
- **Ringkasan Hasil Penuh:** **`76`** top-level test, **`155`** subtest (`148` subtest tingkat-1 + `7` subtest tingkat-2), **`0` fail**, **`0` skip**, durasi **`68.048s`**.

## Matriks Kasus Yang Diverifikasi (`TestPendingAndUnverifiedAutoAssignedSelfServiceGuardE2E`)

1. **`All5CreationVariants_PendingBlockedThenVerifyEmailUnlocks`** (`createUser` publik anonim, `createUser` admin, `createEmployeeUser`, `createVendorUser`, `createSupplierUser`):
   - **Auto-assignment pada `PENDING`:** setiap akun baru berstatus `PENDING` sudah langsung memiliki role baseline (`customer` untuk `CUSTOMER`, `iam_selfservice` untuk `EMPLOYEE`/`VENDOR`/`SUPPLIER`) pada `ScopeSelf(user.id)` (`userRoles` & `userEffectiveRoles` `total_count == 1`), serta `1` token aktif `email_verification`.
   - **Blokir saat `PENDING` (`is_verified = false`):**
     - `userLogin` menggunakan `email` maupun `username` ditolak `AUTHENTICATION`; `browserSession.authenticated == false`.
     - Seluruh query `*self` (`me`, `mySessions`, `myAccessSnapshot`, `myWorkspaceAccess`) ditolak `UNAUTHENTICATED`.
     - Seluruh mutasi `*self` (`updateMyProfile`, `updateUserEmail`, `changeUsername`, `deletePendingUserEmail`, `changeUserPassword`, `enrollMyTotp`, `disableMyTotp`, `regenerateMyRecoveryCodes`) ditolak `AUTHENTICATION`.
   - **Pembukaan pasca-`verifyEmail`:**
     - Pada registrasi publik, `resendValidationEmailToken` tetap mempertahankan tepat `1` token aktif.
     - Setelah `verifyEmail` mengembalikan `status = VERIFIED`, `verified = true`, user otomatis menjadi `ACTIVE` tanpa `activateUser`; `userLogin` (baik via `email` maupun `username`) berhasil `AUTHENTICATED`; `myAccessSnapshot` memuat `status = "active"` dan kelima permission `*self` pada `ScopeSelf(user.id)`; serta `me`, `mySessions`, dan `updateMyProfile` langsung bekerja.
2. **`ActiveWithoutVerifiedIdentityStillBlockedFromLoginAndSelf`** (`CustomerActivatedBeforeVerify` & `EmployeeActivatedBeforeVerify`):
   - Admin memanggil `activateUser` ke status `ACTIVE` ketika `user_identities.is_verified` masih `false`.
   - `userLogin` (`email` & `username`) serta query/mutasi `*self` tetap gagal hingga `verifyEmail` dijalankan, lalu langsung terbuka setelah verifikasi berhasil.
