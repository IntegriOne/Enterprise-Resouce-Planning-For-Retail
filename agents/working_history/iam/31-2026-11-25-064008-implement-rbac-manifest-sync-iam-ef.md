# 31 — Sinkronisasi Otomatis Manifest RBAC IAM & Enterprise Foundation ke Database IAM

- **Tanggal:** 25 November 2026 (06:40 WIB)
- **Cakupan:** `backend/shared/rbac`, `backend/services/iam`, `backend/gateway`
- **Ringkasan Perubahan:**
  - Membentuk modul `backend/shared/rbac` (`//go:embed catalogs/*.json`, terdaftar di `backend/go.work`) beserta manifest deklaratif `iam.json` (`15` permission, `6` role) dan `enterprise_foundation.json` (`10` permission, `ef_automation`) lengkap dengan validasi namespace lock, larangan pewarisan otomatis `superadmin`, redirect safety, dan strict JSON decoding.
  - Memperbarui `BootstrapPrincipalsAndRBAC` IAM lewat `rbac_catalog_sync.go` dengan PostgreSQL transaction advisory lock (`718291824`), upgrade `is_sensitive` (`false`→`true`) dengan bump revisi keamanan & outbox, penolakan downgrade *fail-closed*, kebijakan *seed-once* role non-sistem, serta `PermissionRepository.Update` penuh.
  - Menambah test anti-drift `permissions_manifest_test.go` pada adapter Gateway IAM & EF serta menyelaraskan E2E catalog.
- **Verifikasi:** `go vet` & `go test -race -count=1` lulus pada `shared/rbac`, `services/iam`, dan `gateway`; rebuild & restart Docker `iam`, `gateway-1`, `gateway-2` sehat; PostgreSQL `iam` terkonfirmasi menyimpan `25` permission, `7` role, dan `74` `role_permissions`; live HTTPS E2E terkait lulus.
