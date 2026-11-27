# 16. Wiring Domain Enterprise Foundation ke Gateway dan Isolasi Gangguan

**Tanggal:** 16 November 2026 (03:30 WIB)

Menyelesaikan wiring `5` service gRPC `enterprise_foundation` ke `/api/graphql` Gateway dengan isolasi gangguan penuh tanpa mengubah Zona 3 maupun service IAM/EF:

- [ef_client.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/source_connections/ef_client.go), [registry.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/source_connections/registry.go), & [main.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/cmd/main.go): koneksi EF non-fatal saat startup, *circuit breaker* (5 gagal beruntun, buka 3 detik), dan pemetaan `codes.Unauthenticated` downstream ke `UNAVAILABLE`.
- [internal/adapter/enterprise_foundation/](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/adapter/enterprise_foundation/), [graphql/enterprise_foundation/](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/graphql/enterprise_foundation/), & `graph/resolvers/ef_*.resolvers.go`: RBAC/scope fail-closed untuk 10 izin EF, *Relay pagination*, `Query.node` polimorfik, resolver nested, dan sanitasi error.
- [enterprise_foundation_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/enterprise_foundation_e2e_test.go): seluruh `79` tes E2E HTTPS penuh (`156` subtest, `0` fail, `0` skip) lulus, membuktikan Gateway dan IAM tetap sehat saat `enterprise-foundation` dimatikan dan pulih otomatis saat dinyalakan ulang.
