# E2E IAM ↔ Enterprise Foundation — `TestEnterpriseFoundation_SuperadminAllRolesPermissionsAndOperations`, `TestQueryGroupRoles`, `TestMutationCreateVendorUser`, `TestMutationCreateSupplierUser`

- **Tanggal:** 27 November 2026 (05:54 WIB)
- **Target:** Stack HTTPS lokal `https://erp.localhost:8443/api/graphql` (`gateway-1`, `gateway-2`, `iam` rebuilt, `enterprise-foundation`, `postgres`, `redis`, `caddy`).
- **Perintah Eksekusi:**
  `go test -C backend/gateway -v -count=1 -tags=e2e -run '^(TestEnterpriseFoundation_SuperadminAllRolesPermissionsAndOperations|TestQueryGroupRoles|TestMutationCreateVendorUser|TestMutationCreateSupplierUser)$' ./internal/e2e`

## Matriks Cases & Hasil Observasi

| Test / Subtest | Cakupan Korelasi EF ↔ IAM | Hasil |
| --- | --- | --- |
| `TestEnterpriseFoundation_SuperadminAllRolesPermissionsAndOperations` | 10 izin `enterprise_foundation.*`, 6 peran EF (`ef_automation`, dll.), 64 query + 48 mutasi + 8 node EF, dan penugasan 5 scope eksternal (`LEGAL_ENTITY`, `ORGANIZATION_UNIT`, `FACILITY`, `SUPPLIER_PARTY`, `VENDOR_PARTY`) via IAM → EF gRPC | `PASS` (1.22s) |
| `TestQueryGroupRoles/ScopedActiveAndExpiredGroupRoles` | Pembuatan fixture EF aktif (`createFixtureEFFacility`), `assignGroupRole` berscope `FACILITY` tervalidasi gRPC IAM → EF, serta filter `include_expired` | `PASS` (0.40s) |
| `TestQueryGroupRoles/RBACCustomerForbidden` | Penolakan otorisasi non-privileged pada query `groupRoles` | `PASS` (0.36s) |
| `TestMutationCreateVendorUser` | Provisioning user `VENDOR` + penautan FK logis `party_contacts.iam_user_id` (`PENDING_LINK`), login akun vendor, dan penolakan customer | `PASS` (0.69s) |
| `TestMutationCreateSupplierUser` | Provisioning user `SUPPLIER` + penautan FK logis `party_contacts.iam_user_id` (`PENDING_LINK`), login akun supplier, dan penolakan customer | `PASS` (0.69s) |

## Catatan Cakupan & Isolasi

- Pengujian E2E lain yang memanggil `createFixtureUser(..., "EMPLOYEE", ...)` (`TestQueryUserRoles`, `TestMutationAssignUserRole`, `TestMutationAssignGroupRole`, `TestRBACCreateBusinessUsersStrictEnforcement`, `TestIAMToEF_ScopedAssignmentAndVerifierFullMatrix`, `TestEnterpriseFoundation_ScopedRBACEnforcement`, `TestEnterpriseFoundation_FullLifecycleAndNestedGraphQL`, `TestEnterpriseFoundation_OutageFaultIsolationAndAutoRecovery`, `TestEnterpriseFoundation_CrossServicePartnerUserProvisioningAndFKE2E`) gagal pada langkah provisioning employee karena perubahan kontrak HRM aktif (`CreateEmployeeUserInput.worker_id: ID!` wajib + MFA), bukan pada jalur korelasi EF ↔ IAM.
