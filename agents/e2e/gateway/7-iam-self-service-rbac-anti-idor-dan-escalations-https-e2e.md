# Laporan E2E: Pemisahan Modul `*self`, Anti-IDOR, dan Anti-Privilege Escalation IAM Gateway

Tanggal: 2026-11-15 16:52 WIB  
Target: `https://erp.localhost:8443/api/graphql` (8 container Docker Compose sehat: `postgres`, `redis`, `iam`, `gateway-1`, `gateway-2`, `frontend-1`, `frontend-2`, `caddy`)  
Suite: `go test -C backend/gateway -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...`

## Hasil Eksekusi Ringkas

- **Total Top-Level Test:** `74` `PASS` (`0` fail, `0` skip)
- **Total Subtest:** `146` `PASS` (`0` fail, `0` skip)
- **Durasi Package E2E:** `62.634s`
- **Unit & Lint Pendukung:** `go fmt`, `go vet -tags=e2e ./...`, dan `go test -race -count=1` lulus pada `shared/authz` (`1` package), `services/iam` (`20` package), dan `gateway/internal/...` (`6` package, termasuk `self_authorizer_test.go` & `self_rbac_test.go`).

## Matriks Skenario yang Diuji (`TestSelfServiceModuleRBACAndAntiEscalation` & Suite Lengkap)

| Kategori | Kasus Uji | Ekspektasi Keamanan & Kontrak | Hasil |
| :--- | :--- | :--- | :--- |
| Katalog Permission | Query `permissions` untuk `iam.users.readself`, `iam.users.updateself`, `iam.users.deleteself`, `iam.sessions.readself`, `iam.sessions.deleteself` | Kelima izin terdaftar resmi dan terindeks oleh IAM | `PASS` |
| Baseline Peran & Redirect | Login akun `CUSTOMER` (`customer` -> `/`), `EMPLOYEE` (`iam_selfservice` -> `/account`), `VENDOR` (`iam_selfservice` -> `/vendor`), `SUPPLIER` (`iam_selfservice` -> `/supplier`) | Seluruh `5` izin `*self` terpasang dengan `ScopeSelf(id == user_id)` dan `can_access_identity: false` | `PASS` |
| Self-Service Operations | Pemanggil berizin `*self` membaca profil (`me`), daftar sesi pribadi, mengaktifkan/mutasi email/MFA/session pribadi (`ScopeSelf`) | Operasi mandiri berhasil dan penolakan reauth sensitif tetap berjalan sesuai TTL | `PASS` |
| Anti-IDOR & Zero-Oracle | `CUSTOMER` / `EMPLOYEE` dengan `*self` mencoba `node(id: "gid://integrione/user/<other>")`, `user(id: ...)`, nested lookup, atau `revokeSession` milik user lain | Ditolak (`FORBIDDEN: missing required permission: iam.users.read` / zero-oracle tanpa bocoran PII) | `PASS` |
| Anti-Privilege Escalation | Akun `*self` mencoba mutasi internal (`createEmployeeUser`, `assignUserRole`, `assignRolePermission`) | Ditolak `FORBIDDEN` (`iam.users.write` / `iam.roles.write`), tidak ada mutasi RBAC yang lolos | `PASS` |
| Validasi Scope `SELF` | `assignUserRole` dengan `scope_type: SELF` tetapi `scope_id != target user_id` | Ditolak `ARGUMENT` (`self scope_id must match target user_id`) | `PASS` |
| Regresi Suite IAM Gateway | Seluruh 72 operasi Query/Mutation Relay, `TestUserIDAcceptsGIDAndRawUUID`, serta 7 subtest `TestLiveHTTPSIAMGatewayE2E` | Seluruh alur CSRF pre-auth, RBAC staf, anti-oracle `createUser`, dan failover Redis ACL tetap hijau | `PASS` |
