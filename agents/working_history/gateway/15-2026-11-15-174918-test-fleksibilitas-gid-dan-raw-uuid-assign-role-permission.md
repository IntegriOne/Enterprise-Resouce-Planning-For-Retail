# 15. Penambahan Unit Test dan E2E Fleksibilitas GID/Raw UUID pada AssignRolePermission

**Tanggal:** 15 November 2026 (17:49 WIB)

Menambahkan cakupan pengujian unit adapter IAM Gateway serta pengujian E2E HTTPS langsung untuk memastikan `assignRolePermission`, `revokeRolePermission`, `role`, dan `rolePermissions` fleksibel menerima Global Object ID (`gid://integrione/...`) maupun UUID mentah (lowercase/uppercase) serta menormalisasi output ke GID kanonik Relay:

- [user_admin_adapter_gid_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/user_admin_adapter_gid_test.go): memperluas harness mock (`gidMockRoleClient`, `gidMockPermClient`) untuk mencatat `lastPermID`, `AssignRolePermission`, `RevokeRolePermission`, dan `GetPermissionById`.
- [rbac_adapter_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/iam/rbac_adapter_test.go): menambahkan `3` fungsi uji tabel (`52` subtest) untuk `ParseTypedID`, kombinasi GID/UUID mentah pada `AssignRolePermission`/`RevokeRolePermission`, query `Role`/`RolePermissions`, dan penolakan GID resource salah sebelum RPC.
- [iam_operations_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go): menambahkan subtest E2E `FlexibleGIDAndRawUUID` pada `TestMutationAssignRolePermission` serta memperbaiki sufiks `uniqueTag` (`_z<seq>`) agar angka urut tidak berbenturan dengan `ErrPasswordContainsIdentity`.
- Verifikasi `go fmt`, `go vet -tags=e2e`, `go test -race` (`./internal/...`), dan E2E penuh (`76` test / `156` subtest / `0` fail / `0` skip / `69.101s`) lulus semua.
