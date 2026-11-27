# 32 — Verifikasi & Hardening Korelasi IAM ↔ Enterprise Foundation

- **Tanggal:** 27 November 2026 (05:54 WIB)
- **Cakupan:** [backend/services/iam/](file:///home/muriarfad/Documents/projects/erp-retail/backend/services/iam/) (`ef_verifier.go`, `ef_verifier_test.go`, `role/test/service_test.go`, `group/test/service_test.go`, `pgstore/tx.go`, `.env.example`, `docker-compose.yml`) dan [iam_operations_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_operations_e2e_test.go).
- **Ringkasan:**
  - Mengaudit penuh korelasi IAM ↔ Enterprise Foundation (`EFVerifier` 5 scope eksternal, HMAC `authcontext`, provisioning `createVendorUser`/`createSupplierUser` pada `party_contacts.iam_user_id`, dan katalog RBAC).
  - Memperbaiki `withBoundedTimeout` dengan headroom `500ms` serta keepalive `40s/3s` di `EFVerifier` agar timeout EF terpetakan fail-closed tanpa memicu pemutus sirkuit IAM, dan memperbaiki selector GORM di `tx.go`.
  - Menambahkan `createFixtureEFFacility` pada 5 kasus `FACILITY` di `iam_operations_e2e_test.go`, unit test `bufconn`/headroom/service `ScopeOwnerVerifier`, menyinkronkan `.env.example`/`docker-compose.yml`, membangun ulang image `iam`, dan memverifikasi tes `-race`, lint IAM (`0 issues`), E2E EF/vendor/supplier/group, serta re-index `codebase-memory-mcp`.
