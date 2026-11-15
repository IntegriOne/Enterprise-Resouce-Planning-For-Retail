# Verifikasi E2E Auto-Assignment Peran & Permission `*self` Saat Pembuatan Akun

Tanggal: 2026-11-15 17:03 WIB

## Ringkasan Perubahan

1. **Penambahan Suite `TestCreateUserVariantsAutoAssignSelfRoleAndPermissions`:** Menambahkan pengujian E2E HTTPS di [iam_operations_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go) untuk memverifikasi kelima jalur pembuatan akun (`createUser` publik anonim, `createUser` admin, `createEmployeeUser`, `createVendorUser`, `createSupplierUser`) tanpa memanggil `assignUserRole`.
2. **Pembuktian Status `PENDING` & Pasca-Login:** Memverifikasi `customer` (`CUSTOMER`) dan `iam_selfservice` (`EMPLOYEE`, `VENDOR`, `SUPPLIER`) langsung membawa `5` permission `*self` (`ScopeSelf(user_id)`), tersedia di nested `user.roles`/`user.effective_roles` segera saat status `PENDING`, serta aktif otomatis di `myAccessSnapshot` dan operasi mandiri (`me`, `mySessions`) setelah aktivasi & login.
3. **Verifikasi:** `go fmt`, `go vet -tags=e2e ./...`, `6` package unit test Gateway, dan seluruh `75` E2E test (`146` subtest, `0` fail, `0` skip, `64.223s`) `PASS`.
