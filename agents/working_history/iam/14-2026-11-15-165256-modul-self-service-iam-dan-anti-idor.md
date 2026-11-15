# Pemisahan Modul Self-Service (`*self`), Peran Baseline `iam_selfservice`, & Perbaikan Redirect `UserKind`

Tanggal: 2026-11-15 16:52 WIB

## Ringkasan Perubahan

1. **Katalog & Peran Baseline `*self`:** Memisahkan 5 permission personal (`iam.users.readself`, `iam.users.updateself`, `iam.users.deleteself`, `iam.sessions.readself`, `iam.sessions.deleteself`) dari permission staf lintas-user (`iam.users.read`, `iam.sessions.read`, `iam.sessions.revoke`). Peran `customer` (`/` prio `100`) dan `iam_selfservice` (`""` prio `999`) otomatis memberi seluruh `*self` ber-`ScopeSelf{ID: userID}` untuk `CUSTOMER`, `EMPLOYEE`, `VENDOR`, dan `SUPPLIER` tanpa akses `/identity` (`can_access_identity: false`).
2. **Perbaikan Mapper & Rekonsiliasi Bootstrap:** Memperbaiki `roleToModel` di [mappers.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/repository/role/mappers.go) agar `DefaultRedirectPath` kosong tidak ditimpa menjadi `"/"`, serta menambah rekonsiliasi metadata redirect peran sistem pada `ensureBootstrapRoles` ([bootstrap.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/services/bootstrap.go)) sehingga login `EMPLOYEE`/`VENDOR`/`SUPPLIER` terarah ke `/account`, `/vendor`, dan `/supplier`.
3. **Verifikasi:** `go fmt`, `go vet`, dan `go test -race -count=1` lulus pada `shared/authz` (`1` package) dan `services/iam` (`20` package); rebuild container `iam` serta E2E HTTPS penuh (`74` test, `146` subtest, `0` fail, `0` skip, `62.634s`) lulus.
