# Penegakan Dependensi Fail-Closed `UserRoles` & Verifikasi Auto-Assign Peran `*self` Saat Pembuatan Akun

Tanggal: 2026-11-15 17:03 WIB

## Ringkasan Perubahan

1. **Dependensi Fail-Closed `UserRoles`:** Memperketat `validateUserDeps` dan `ensureReady` di [service.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/service.go) agar konstruktor `user.NewService` selalu gagal tertutup bila `UserRoles == nil`, mencegah `assignBaselineSelfRoleInTx` melewati penugasan peran `*self` secara diam-diam.
2. **Pembaruan Harness & Unit Test:** Memasang `UserRoles` di [runtime_rpc_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/transport/grpc/handler/test/runtime_rpc_test.go) dan [mocks_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/test/mocks_test.go), serta menambah `TestUserService_CreateUserAutoAssignsBaselineSelfRoleForAllUserKinds` di [user_lifecycle_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/user/test/user_lifecycle_test.go) untuk `CUSTOMER`, `EMPLOYEE`, `VENDOR`, dan `SUPPLIER`.
3. **Verifikasi Penuh:** `go fmt`, `go vet`, dan `go test -race -count=1` lulus pada `shared/authz` (`1`) & `services/iam` (`20`); build ulang container `iam` serta E2E HTTPS (`75` test, `146` subtest, `0` fail, `0` skip, `64.223s`) lulus penuh.
