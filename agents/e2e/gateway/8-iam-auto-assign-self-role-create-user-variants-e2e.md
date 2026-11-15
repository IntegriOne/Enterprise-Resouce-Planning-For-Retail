# Laporan E2E: Auto-Assignment Peran & Permission `*self` pada `createUser`, `createEmployeeUser`, `createVendorUser`, dan `createSupplierUser`

Tanggal: 2026-11-15 17:03 WIB  
Target: `https://erp.localhost:8443/api/graphql` (8 container Docker Compose sehat: `postgres`, `redis`, `iam`, `gateway-1`, `gateway-2`, `frontend-1`, `frontend-2`, `caddy`)  
Suite: `go test -C backend/gateway -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...` (`TestCreateUserVariantsAutoAssignSelfRoleAndPermissions` + regresi penuh)

## Hasil Eksekusi Ringkas

- **Total Top-Level Test:** `75` `PASS` (`0` fail, `0` skip)
- **Total Subtest:** `146` `PASS` (`0` fail, `0` skip)
- **Durasi Package E2E:** `64.223s`
- **Unit Test Pendukung:** `go test -race -count=1` lulus di `shared/authz` (`1` package), `services/iam` (`20` package, termasuk `TestUserService_CreateUserAutoAssignsBaselineSelfRoleForAllUserKinds`), dan `gateway/internal/...` (`6` package).

## Matriks Kasus Uji (`TestCreateUserVariantsAutoAssignSelfRoleAndPermissions`)

| Jalur Pembuatan Akun | Verifikasi Saat Status `PENDING` (Tanpa `assignUserRole`) | Verifikasi Pasca Aktivasi & Login (`myAccessSnapshot`) | Hasil |
| :--- | :--- | :--- | :--- |
| Template `customer` & `iam_selfservice` | Query `rolePermissions` memuat tepat `5` izin `*self` (`iam.users.{readself,updateself,deleteself}`, `iam.sessions.{readself,deleteself}`) tanpa izin lintas-user | Berlaku untuk seluruh akun baru tanpa intervensi admin | `PASS` |
| `createUser` (Anonim Publik -> `CUSTOMER`) | Response payload masked `user: nil, was_successful: true`; admin `users`, `userRoles`, `userEffectiveRoles` menunjuk `customer` `ScopeSelf(id == user_id)` | Snapshot otomatis berisi `customer` + `5` izin `*self` (`SELF`), redirect `/`, `can_access_identity: false`, `me` & `mySessions` sukses | `PASS` |
| `createUser` (Admin -> `CUSTOMER`) | Nested `user.roles` & `user.effective_roles` langsung pada payload mutasi `createUser` (`PENDING`) sudah memuat `customer` `ScopeSelf(user_id)` | Snapshot otomatis berisi `customer` + `5` izin `*self` (`SELF`), redirect `/`, `users` tetap `FORBIDDEN` | `PASS` |
| `createEmployeeUser` (`EMPLOYEE`) | Nested `user.roles` & `user.effective_roles` langsung pada payload mutasi `createEmployeeUser` (`PENDING`) sudah memuat `iam_selfservice` `ScopeSelf(user_id)` | Snapshot otomatis berisi `iam_selfservice` + `5` izin `*self` (`SELF`), redirect `/account`, `can_access_identity: false`, `me` & `mySessions` sukses | `PASS` |
| `createVendorUser` (`VENDOR`) | Nested `user.roles` & `user.effective_roles` langsung pada payload mutasi `createVendorUser` (`PENDING`) sudah memuat `iam_selfservice` `ScopeSelf(user_id)` | Snapshot otomatis berisi `iam_selfservice` + `5` izin `*self` (`SELF`), redirect `/vendor`, `can_access_identity: false`, `me` & `mySessions` sukses | `PASS` |
| `createSupplierUser` (`SUPPLIER`) | Nested `user.roles` & `user.effective_roles` langsung pada payload mutasi `createSupplierUser` (`PENDING`) sudah memuat `iam_selfservice` `ScopeSelf(user_id)` | Snapshot otomatis berisi `iam_selfservice` + `5` izin `*self` (`SELF`), redirect `/supplier`, `can_access_identity: false`, `me` & `mySessions` sukses | `PASS` |
