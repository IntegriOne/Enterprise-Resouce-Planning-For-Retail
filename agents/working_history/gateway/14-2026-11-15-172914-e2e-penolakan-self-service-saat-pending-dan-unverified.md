# 14. Verifikasi E2E dan Guard Middleware Self-Service Saat PENDING & Unverified

**Tanggal:** 15 November 2026 (17:29 WIB)

Menambahkan uji guard pada layer middleware Gateway serta suite E2E HTTPS langsung terhadap `https://erp.localhost:8443/api/graphql` untuk membuktikan role `*self` yang terpasang otomatis saat pembuatan akun tidak dapat dieksekusi sebelum user terverifikasi dan `ACTIVE`:

- [pending_self_guard_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/middleware/pending_self_guard_test.go): memastikan `Authorizer.RequireSelfPermission` dan `RequireSelfSensitive` menolak status `pending`, `suspended`, dan `disabled` (`CodeForbidden`) meski seluruh grant `*self` (`ScopeSelf`) sudah ada di snapshot.
- [iam_pending_self_guard_e2e_test.go](file:///home/muriarfad/Documents/projects/erp-retail/backend/gateway/internal/e2e/iam_pending_self_guard_e2e_test.go): menguji ke-5 varian pembuatan akun (`createUser` publik/admin, `createEmployeeUser`, `createVendorUser`, `createSupplierUser`), skenario aktivasi tanpa verifikasi identitas (`is_verified=false`), hingga pembukaan login (`email` & `username`) serta `*self` setelah `verifyEmail`.
- `go fmt`, `go vet -tags=e2e`, unit test `-race` (`6` package Gateway), dan E2E penuh (`76` test / `155` subtest / `0` fail / `0` skip) lulus seluruhnya.
