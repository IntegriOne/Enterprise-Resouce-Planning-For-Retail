# Audit 16 — Korelasi IAM ↔ Enterprise Foundation Backend

- **Tanggal:** 27 November 2026 (05:54 WIB)
- **Cakupan:** Katalog RBAC lintas modul (`iam`, `enterprise_foundation`), `EFVerifier` gRPC (`services/iam/infrastructure/scopeverifier`), layanan `role` & `group` IAM, handler validasi scope & portal Enterprise Foundation, Gateway provisioning & adapter EF/IAM, konfigurasi runtime, serta E2E HTTPS.

## Ringkasan Temuan & Kondisi Aktual

1. **Verifikasi Scope Eksternal IAM → EF (`ScopeOwnerVerifier` / `EFVerifier`)**
   - Pemetaan 5 target eksternal (`legal_entity`, `organization_unit`, `facility`, `supplier_party`, `vendor_party`) serta bypass lokal `global`/`self` (menjaga isolasi startup IAM) terhubung dengan proof HMAC `authcontext`.
   - **Celah yang diperbaiki:**
     - `withBoundedTimeout` di `backend/services/iam/infrastructure/scopeverifier/ef_verifier.go` sebelumnya mengembalikan `ctx` pemanggil apa adanya saat sisa deadline `<= timeout` (`5s` Gateway == `5s` IAM), berisiko memicu `codes.Canceled` sebelum `ErrScopeOwnerVerificationUnavailable` terpetakan (melanggar isolasi circuit breaker Gateway D12). Kini dicadangkan headroom `500ms` dan keepalive diperketat menjadi `Time: 40s, Timeout: 3s`.
     - `.env.example` dan `docker-compose.yml` lokal IAM belum mendokumentasikan variabel `IAM_EF_GRPC_*`. Kini telah diselaraskan dengan `backend/docker-compose.yml`.
     - Unit test `role/test/service_test.go`, `group/test/service_test.go`, dan `ef_verifier_test.go` kini mencakup skenario `ScopeOwnerVerifier` penuh (sukses, `NotFound`, `NotEligible`, `Unavailable` tanpa penulisan DB/outbox), deadline headroom, serta roundtrip `bufconn` ber-HMAC.
   - **Temuan lint IAM:** `repository/pgstore/tx.go` (`tx.Dialector.Name()` → `tx.Name()`) diperbaiki sehingga `golangci-lint run ./...` pada `backend/services/iam` lolos `0 issues`.

2. **Provisioning Mitra & Akses Portal EF (`party_contacts.iam_user_id`)**
   - Alur `createVendorUser` / `createSupplierUser` di Gateway menegakkan `iam.users.write` GLOBAL + `enterprise_foundation.parties.write` (GLOBAL atau scoped `vendor_party`/`supplier_party`), pra-validasi `PreparePortalUserProvisioning`, penautan `CompletePortalUserProvisioning` (`pending_link`), serta kompensasi `SoftDeleteUser` jika penautan gagal.

3. **Katalog RBAC & Prinsipal Otomasi**
   - Katalog tersemat `iam.json` (`15` izin) + `enterprise_foundation.json` (`10` izin) sinkron; `superadmin` memegang seluruh `25` izin dan `auto_system_ef` memegang `ef_automation` pada scope `global`.

4. **Perbaikan Fixture E2E & Blocker Lintas Modul**
   - `backend/gateway/internal/e2e/iam_operations_e2e_test.go` sebelumnya masih memakai UUID tiruan (`1111...`–`4444...`) pada 5 kasus `scope_type: FACILITY` sehingga gagal `NOT_FOUND` terhadap `EFVerifier`. Kini memakai `createFixtureEFFacility`.
   - **Blocker eksternal (di luar EF↔IAM):** E2E yang memanggil `createFixtureUser(..., "EMPLOYEE", ...)` maupun `createEmployeeUser` tanpa `worker_id` + sesi MFA saat ini tertahan validasi GraphQL `CreateEmployeeUserInput.worker_id: ID!` / guard MFA HRM ketika `HRM_GRPC_ENABLED=true`. Kontrak HRM tidak diubah.
