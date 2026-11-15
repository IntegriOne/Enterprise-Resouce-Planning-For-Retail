# Laporan E2E: Fleksibilitas Input GID dan Raw UUID pada Assign/Revoke Role Permission

- **Tanggal:** 15 November 2026 (17:49 WIB)
- **Lingkungan:** `https://erp.localhost:8443/api/graphql` (8 container Compose aktif & healthy)
- **Suite Dijalankan:** `GOWORK=off go test -C backend/gateway -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...`
- **Ringkasan Hasil Penuh:** **`76`** top-level test, **`156`** subtest, **`0` fail**, **`0` skip**, durasi **`69.101s`**.

## Matriks Kasus Yang Diverifikasi (`TestMutationAssignRolePermission/FlexibleGIDAndRawUUID`)

1. **Kombinasi Input Fleksibel (`assignRolePermission`, `revokeRolePermission`, `rolePermissions`):**
   - `role_id` berupa GID (`gid://integrione/role/<uuid>`) + `permission_id` berupa UUID mentah lowercase -> mutasi sukses (`was_successful = true`), response tetap mengembalikan `role_permission_edge.role_id` berformat GID kanonik, `node.id` berformat GID `permission`, dan `node._id` berformat UUID lowercase; dicabut kembali dengan `role_id` UUID mentah + `permission_id` GID lalu `rolePermissions(role_id: <rawUpper>)` kembali ke `total_count = 0`.
   - `role_id` UUID mentah lowercase + `permission_id` GID `permission` -> sukses dengan output GID kanonik.
   - `role_id` UUID mentah uppercase + `permission_id` UUID mentah uppercase -> sukses dan tetap ternormalisasi ke output lowercase kanonik (`role_id`, `node.id`, `node._id`).
   - Query `rolePermissions(role_id: <rawLower>)` menampilkan `total_count = 2` dengan seluruh `edge.role_id` berbentuk GID kanonik; pencabutan kembali menggunakan kombinasi UUID mentah uppercase dan lowercase menurunkan `total_count` menjadi `0`.
2. **Kasus Negatif (`ARGUMENT` / `VALIDATION`):**
   - `assignRolePermission` dengan GID `permission` pada `role_id`, GID `role` pada `permission_id`, GID `user` pada `permission_id`, atau string bukan UUID (`"not-a-valid-uuid"`) ditolak bersih pada payload (`was_successful = false`, `errors` terisi).
   - `revokeRolePermission` dengan GID resource tertukar (`permission` pada `role_id`, `role` pada `permission_id`) ditolak bersih pada payload (`was_successful = false`, `errors` terisi).

## Catatan Uji Unit Pendamping (`./internal/adapter/iam/...`)

- `go test -C backend/gateway -race -count=1 ./internal/...` lulus seluruh package (`6` package dengan test, `2` package tanpa file test).
- `rbac_adapter_test.go` memverifikasi `ParseTypedID` (`15` subtest), `AssignRolePermission` / `RevokeRolePermission` (`28` subtest termasuk penolakan sebelum panggilan gRPC), serta `Role` / `RolePermissions` (`9` subtest).
