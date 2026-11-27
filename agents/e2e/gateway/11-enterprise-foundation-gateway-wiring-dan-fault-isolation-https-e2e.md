# Laporan E2E: Wiring Domain Enterprise Foundation ke Gateway & Isolasi Gangguan

- **Tanggal:** 16 November 2026 (03:30 WIB)
- **Lingkungan:** `https://erp.localhost:8443/api/graphql` (`9` container Compose aktif & healthy: `caddy`, `frontend-1`, `frontend-2`, `gateway-1`, `gateway-2`, `iam`, `enterprise-foundation`, `postgres`, `redis`)
- **Suite Dijalankan:** `GOWORK=off /opt/go/bin/go test -tags=e2e -v -count=1 -timeout 10m ./internal/e2e/...`
- **Ringkasan Hasil Penuh:** **`79`** top-level test, **`156`** subtest (`149` level-1 + `7` level-2), **`0` fail**, **`0` skip**, durasi **`81.787s`**.

## Matriks Kasus Enterprise Foundation (`enterprise_foundation_e2e_test.go`)

1. **`TestEnterpriseFoundation_FullLifecycleAndNestedGraphQL` (`2.00s`):**
   - **Fail-Closed Anonim & Customer:** Request anonim ke `allAddresses` ditolak `UNAUTHENTICATED`, mutasi `createAddress` mengembalikan `MutationResult { was_successful: false, errors: [{ type: AUTHENTICATION }] }`. Pengguna `CUSTOMER` tanpa izin EF ditolak `FORBIDDEN` (`AUTHORIZATION`) oleh RBAC Gateway sebelum menyentuh gRPC EF.
   - **Address (`AddressService`):** `createAddress`, `findOrCreateAddress` (`created = true` lalu duplikat `created = false`), `updateAddress`, paginasi kursor Relay `allAddresses(first: 1, after: $after)`, `isAddressReferenced` (`false` sebelum tertaut -> `true` setelah tertaut ke `LegalEntity`), `addressesByIds`, serta `deleteAddress` pada alamat tak terpakai.
   - **LegalEntity (`LegalEntityService`):** `createLegalEntity` induk dan anak, transisi siklus hidup `activateLegalEntity` -> `suspendLegalEntity` -> `activateLegalEntity`, resolver nested `registered_address`, `children`, `parent_legal_entity`, lookup `legalEntityByCode`, `legalEntitiesByNib`, serta `validateLegalEntityScopeTarget`.
   - **Location (`LocationService`):** `createLocation` (`HEADQUARTERS` dan `WAREHOUSE`), resolver nested `legal_entity`, `parent_location`, `address`, query `locationByCode`, `locationsByLegalEntityAndKind`, `locationsByAddressId`, serta `validateFacilityScope` dan `validateWarehouseScope` (`eligible = true`).
   - **OrganizationUnit (`OrganizationUnitService`):** `createOrganizationUnit` (`DIVISION` dan `DEPARTMENT`), mutasi `assignOrganizationUnitHeadPerson`, resolver nested `legal_entity`, `primary_location`, `children`, `organizationUnitsByHeadPersonId`, serta `validateOrganizationUnitScopeTarget`.
   - **Party & Sub-Resource (`PartyService`):** `createParty`, `verifyAndActivateParty`, `assignPartyRole` (`SUPPLIER`), `addPartyContact` (`contact_role = "account_lead"`), `activatePartyPortalAccess`, `linkPartyAddress` (`REGISTERED_OFFICE`), resolver nested `roles`/`contacts`/`addresses.address`, `evaluateRepresentativePortalAccess` (`allowed = true`), dan `validatePartyRoleScope`.
   - **Relay Polimorfik `Query.node`:** Menguji resolusi seluruh `8` tipe GID Enterprise Foundation (`Address`, `LegalEntity`, `Location`, `OrganizationUnit`, `Party`, `PartyRole`, `PartyContact`, `PartyAddress`) beserta GID `User` IAM dalam satu query.
2. **`TestEnterpriseFoundation_ScopedRBACEnforcement` (`0.93s`):**
   - Pengguna berizin `enterprise_foundation.legal_entities.read` dengan scope `LEGAL_ENTITY` terikat ke entitas `A` berhasil membaca `legalEntity(id: A)` namun ditolak `FORBIDDEN` saat mengakses `legalEntity(id: B)`.
3. **`TestEnterpriseFoundation_OutageFaultIsolationAndAutoRecovery` (`11.29s`):**
   - Container `enterprise-foundation` dihentikan (`docker compose stop enterprise-foundation`); container `gateway-1`, `gateway-2`, dan `iam` tetap `running:healthy`.
   - Selama pemadaman lintas replika Gateway: `health` mengembalikan `"ok"`, query IAM `me` dan `users` sukses 100%, sesi pengguna EF tidak terputus, pengguna `CUSTOMER` tetap ditolak `FORBIDDEN` oleh RBAC Gateway, query EF mengembalikan `UNAVAILABLE` tanpa membocorkan detail upstream (`enterprise-foundation:9002`, `127.0.0.1`, `dial tcp`, `postgres`), dan mutasi EF mengembalikan HTTP `200` `MutationResult { was_successful: false, errors: [{ type: UNAVAILABLE }] }`.
   - Setelah `enterprise-foundation` dinyalakan ulang tanpa me-restart Gateway maupun IAM, query `allAddresses` (4 kali beruntun melintasi dua replika Gateway) dan mutasi `createAddress` pulih otomatis penuh.

## Kebersihan Data & Uji Unit Pendamping

- `seedEFPermissionCatalog` membersihkan 10 izin sementara EF via `t.Cleanup` setelah role uji selesai dibersihkan; verifikasi DB pasca-suite menunjukkan `15` izin baseline di database `iam` dan `0` baris tersisa di tabel-tabel `enterprise_foundation`.
- Seluruh uji unit Gateway pada `./internal/adapter/enterprise_foundation/...`, `./internal/source_connections/...`, `./internal/config/...`, dan paket lain diverifikasi kembali melalui `go vet ./...` dan `go test -race -count=1 ./...`.
