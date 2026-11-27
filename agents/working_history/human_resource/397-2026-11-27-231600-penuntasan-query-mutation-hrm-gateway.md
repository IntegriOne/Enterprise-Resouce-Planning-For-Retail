# 397 — Penuntasan Query & Mutation HRM di GraphQL Gateway

- **Tanggal:** 27 November 2026
- **Cakupan:** RBAC Manifest (`backend/shared/rbac`), GraphQL SDL, Adapter & Resolver Gateway (`backend/gateway`), dan Dokumentasi (`agents/`).

Menutup seluruh selisih Query dan Mutation antara 22 service gRPC HRM (`1.016` RPC bisnis) dengan GraphQL Gateway (`/api/graphql`):
1. Menyinkronkan `human_resource.json` ke **409** izin (`434` total RBAC) dan menginisialisasi ke-22 client gRPC di `human_resource.Adapter`.
2. Menambahkan **59 entitas baru** (termasuk `LearningCourseCompetency` composite key dan `int64` IDs), **57 mutasi workflow command**, **30 mutasi CRUD tambahan**, serta filter finder konsolidasi pada entitas existing dengan proteksi IDOR/MFA berbasis session cookie.
3. `golangci-lint` (`0 issues`), `go test -race` (`human_resource`, `server`, `rbac`), `check_hrm_schema_guards.py` (**86 Node**, **86 Connection**, **176/176 Query**, **326/326 Mutation**), serta introspection Caddy `8443` lulus.
