# Katalog Roles & Permissions: Enterprise Foundation

Dokumen ini merupakan referensi kanonik resmi untuk seluruh *permission codes*, pemetaan *system template roles*, aturan *scope* ABAC/RBAC, serta penegakan *sensitive step-up reauthentication* domain **Enterprise Foundation** yang didefinisikan pada manifest resmi ([`enterprise_foundation.json`](../../../backend/shared/rbac/catalogs/enterprise_foundation.json)) dan disinkronkan otomatis saat startup IAM ([`rbac_catalog_sync.go`](../../../backend/services/iam/services/rbac_catalog_sync.go), [`bootstrap.go`](../../../backend/services/iam/services/bootstrap.go)), ditegakkan secara *fail-closed* oleh Gateway ([permissions.go](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go) & [internal/adapter/enterprise_foundation/](../../../backend/gateway/internal/adapter/enterprise_foundation/), teruji 1:1 di [`permissions_manifest_test.go`](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions_manifest_test.go)), serta didukung kebijakan evaluasi domain di service Enterprise Foundation ([scope_policy.go](../../../backend/services/enterprise_foundation/domain/service/scope_policy.go) & [portal_policy.go](../../../backend/services/enterprise_foundation/domain/service/portal_policy.go)).

Seluruh *permission* mengikuti konvensi `<management>.<resource>.<action>` dan harus terdaftar serta diberikan melalui mekanisme resmi IAM (tidak pernah ditulis ke JSON `attributes` atau langsung ke tabel IAM oleh layanan lain).

**Sumber Acuan Implementasi:**
- Manifest & Sinkronisasi Katalog RBAC: [enterprise_foundation.json](../../../backend/shared/rbac/catalogs/enterprise_foundation.json), [catalog.go](../../../backend/shared/rbac/catalog.go), [rbac_catalog_sync.go](../../../backend/services/iam/services/rbac_catalog_sync.go), [bootstrap.go](../../../backend/services/iam/services/bootstrap.go)
- Permission & Adapter Gateway: [permissions.go](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go), [permissions_manifest_test.go](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions_manifest_test.go), [adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/adapter.go), [address_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/address_adapter.go), [legal_entity_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/legal_entity_adapter.go), [location_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/location_adapter.go), [organization_unit_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/organization_unit_adapter.go), [party_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/party_adapter.go), [party_sub_adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/party_sub_adapter.go), [gid.go](../../../backend/gateway/internal/adapter/enterprise_foundation/gid.go)
- Skema GraphQL & Resolver Gateway: [ef_common.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_common.graphql), [ef_address.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_address.graphql), [ef_legal_entity.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_legal_entity.graphql), [ef_location.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_location.graphql), [ef_organization_unit.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_organization_unit.graphql), [ef_party.graphql](../../../backend/gateway/graphql/enterprise_foundation/ef_party.graphql), [schema.resolvers.go](../../../backend/gateway/graph/resolvers/schema.resolvers.go), [ef_address.resolvers.go](../../../backend/gateway/graph/resolvers/ef_address.resolvers.go), [ef_legal_entity.resolvers.go](../../../backend/gateway/graph/resolvers/ef_legal_entity.resolvers.go), [ef_location.resolvers.go](../../../backend/gateway/graph/resolvers/ef_location.resolvers.go), [ef_organization_unit.resolvers.go](../../../backend/gateway/graph/resolvers/ef_organization_unit.resolvers.go), [ef_party.resolvers.go](../../../backend/gateway/graph/resolvers/ef_party.resolvers.go)
- Seed IAM, Kebijakan Scope, & Ketahanan Transport: [role.go](../../../backend/services/iam/domain/entity/role.go), [scope_verification.go](../../../backend/services/iam/domain/service/scope_verification.go), [config.go](../../../backend/services/iam/config/config.go), [authorizer.go](../../../backend/gateway/internal/middleware/authorizer.go), [ef_client.go](../../../backend/gateway/internal/source_connections/ef_client.go), [interceptor.go](../../../backend/services/enterprise_foundation/transport/grpc/interceptor/interceptor.go), [users.md](../../users.md)

---

## 1. Daftar Permission Codes Domain Enterprise Foundation (10 Kode Aktif)

Seluruh **10 permission** di bawah ini disinkronkan otomatis saat startup IAM dari manifest [enterprise_foundation.json](../../../backend/shared/rbac/catalogs/enterprise_foundation.json) (domain `"enterprise_foundation"`) dan dipetakan eksplisit di [permissions.go](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go).

| Permission Code | `resource` (Entitas IAM) | `action` | `is_sensitive` | Scope Target Valid | Deskripsi & Cakupan Utama |
| --- | --- | --- | --- | --- | --- |
| `enterprise_foundation.addresses.read` | `enterprise_foundation.addresses` | `read` | `false` | `global` | Membaca master alamat geografis (`address*`, `allAddresses`, `isAddressReferenced`) serta resolver bersarang alamat milik entitas lain. |
| `enterprise_foundation.addresses.write` | `enterprise_foundation.addresses` | `write` | `false` (seed); **`true` di Gateway u/ `deleteAddress`** | `global` | Membuat, mencari-atau-membuat idempoten, memperbarui, dan menghapus master alamat geografis. |
| `enterprise_foundation.legal_entities.read` | `enterprise_foundation.legal_entities` | `read` | `false` | `global`, `legal_entity` | Membaca profil/hierarki badan hukum perusahaan (`legalEntity*`, `allLegalEntities`, `primaryOperatingLegalEntity`) dan memvalidasi target scope `legal_entity`. |
| `enterprise_foundation.legal_entities.write` | `enterprise_foundation.legal_entities` | `write` | `false` (seed); **`true` di Gateway u/ 3 aksi kritikal** | `global`, `legal_entity` | Membuat, memperbarui, mengaktifkan, membekukan, menetapkan entitas operasi utama, membubarkan, serta soft-delete badan hukum. |
| `enterprise_foundation.locations.read` | `enterprise_foundation.locations` | `read` | `false` | `global`, `legal_entity`, `facility` | Membaca lokasi fisik internal (HQ, gudang, toko, cabang) dan memvalidasi target scope `facility` maupun `warehouse`. |
| `enterprise_foundation.locations.write` | `enterprise_foundation.locations` | `write` | `false` (seed); **`true` di Gateway u/ `deleteLocation`** | `global`, `legal_entity`, `facility` | Membuat (`legal_entity`), memperbarui, mengaktifkan, menonaktifkan, dan soft-delete lokasi fisik (`facility`). |
| `enterprise_foundation.organization_units.read` | `enterprise_foundation.organization_units` | `read` | `false` | `global`, `legal_entity`, `organization_unit`, `facility` | Membaca unit organisasi internal (divisi, departemen, tim, cabang) berdasarkan entitas/lokasi/kepala unit dan memvalidasi target scope `organization_unit`. |
| `enterprise_foundation.organization_units.write` | `enterprise_foundation.organization_units` | `write` | `false` (seed); **`true` di Gateway u/ `deleteOrganizationUnit`** | `global`, `legal_entity`, `organization_unit` | Membuat (`legal_entity`), memperbarui, menetapkan `head_person_id`, mengaktifkan/menonaktifkan, serta soft-delete unit organisasi (`organization_unit`). |
| `enterprise_foundation.parties.read` | `enterprise_foundation.parties` | `read` | `false` | `global`, `vendor_party`, `supplier_party` | Membaca mitra eksternal B2B (`Party`), peran bisnis (`PartyRole`), PIC portal IAM (`PartyContact`), relasi alamat (`PartyAddress`), validasi scope peran mitra, dan evaluasi akses portal. |
| `enterprise_foundation.parties.write` | `enterprise_foundation.parties` | `write` | `false` (seed); **`true` di Gateway u/ 6 aksi sensitif** | `global`, `vendor_party`, `supplier_party` | Mengelola profil/pajak/lifecycle `Party`, penugasan `PartyRole`, pengelolaan `PartyContact` & tautan portal IAM (`iam_user_id`), serta relasi `PartyAddress`. |

> [!NOTE]
> - **Perbedaan `IsSensitive` Seed IAM vs Penegakan Adapter Gateway:** Di manifest [`enterprise_foundation.json`](../../../backend/shared/rbac/catalogs/enterprise_foundation.json), seluruh 10 permission `enterprise_foundation.*` disemai dengan `is_sensitive: false`. Namun pada layer adapter Gateway, **12 mutasi berisiko tinggi** (seluruh soft-delete, penetapan badan hukum operasi utama, pembubaran badan hukum, penghapusan `PartyRole`, serta aktivasi/pencabutan/pemutusan portal IAM PIC) secara eksplisit memanggil `requirePerm(..., true)` / `requireAnyScope(..., true, ...)` untuk mewajibkan login terbaru (`<= 10 menit`) beserta verifikasi MFA apabila akun memiliki MFA aktif ([authorizer.go](../../../backend/gateway/internal/middleware/authorizer.go)).
> - **Invarian Kolom `resource` & `action` di IAM:** [role.go](../../../backend/services/iam/domain/entity/role.go) (`NewPermission`) memvalidasi bahwa `resource` harus mengandung titik (`<management>.<resource>`), `action` tidak boleh mengandung titik, dan `code == resource + "." + action`.
> - **Tidak Ada Wildcard atau `*self`:** Enterprise Foundation hanya memakai pola `.read` dan `.write`; akses pembacaan/pengelolaan profil akun manusia maupun sesi tetap menjadi tanggung jawab izin `*self` milik domain IAM.

---

## 2. Daftar System Template Roles & Keterikatan Permission

Saat ini manifest [`enterprise_foundation.json`](../../../backend/shared/rbac/catalogs/enterprise_foundation.json) dan [`iam.json`](../../../backend/shared/rbac/catalogs/iam.json) menyemai **2 *system template role*** (`is_system = true`) yang membawa izin `enterprise_foundation.*`, yakni role human admin perdana (`superadmin`) dan role akun layanan non-interaktif milik modul Enterprise Foundation (`ef_automation`).

| Role Code | Nama Tampilan (`Name`) | Domain (`Domain`) | Default Redirect Path | Priority | `user_kind` Diizinkan (`AccessPolicy`) | Tujuan & Cakupan Principal |
| --- | --- | --- | --- | --- | --- | --- |
| `superadmin` | `Super Administrator` | `iam` | `/identity` | `1` | **Hanya `employee`** (`IsSuperAdminRoleCode`) | Akun human admin perdana (`IAM_SUPERADMIN_*` di `.env`). Memegang seluruh **10 permission `enterprise_foundation.*`** beserta seluruh **15 permission `iam.*`** pada `ScopeGlobal` (ditambah `iam_selfservice` pada `ScopeSelf`). |
| `ef_automation` | `Enterprise Foundation Service Automation` | `enterprise_foundation` | `""` | `999` | **Hanya `system`** (`IsAutomationRoleCode`) | Akun layanan mesin `auto_system_ef` (`IAM_AUTO_SYSTEM_EF_USERNAME` di [config.go](../../../backend/services/iam/config/config.go)). Memegang **10 permission `enterprise_foundation.*`** pada `ScopeGlobal`; dilarang login interaktif maupun masuk dashboard UI. |

> [!IMPORTANT]
> - **Status Role Operasional Non-Sistem (Manusia / Mitra):** Hingga pembaruan dokumen ini, [`enterprise_foundation.json`](../../../backend/shared/rbac/catalogs/enterprise_foundation.json) **belum menyemai role operasional khusus manusia** (`is_system = false`) untuk domain Enterprise Foundation (misalnya `ef_admin`, `ef_operator`, `ef_readonly`, atau role portal mitra terlingkup `vendor_party`/`supplier_party`). Role operasional untuk `employee`, `vendor`, atau `supplier` dibuat dan diatur dinamis melalui API IAM (`createRole` + `assignPermissionsToRole` dengan domain `enterprise_foundation`) tanpa menggabungkan hak admin `iam.*`, atau dapat didaftarkan sebagai template *seed-once* (`is_system: false`) di manifest.
> - **Pemisahan `iam_operator` vs `enterprise_foundation.parties.read`:** Sesuai prinsip isolasi lintas modul, role `iam_operator` (milik IAM) **tidak** memegang izin `enterprise_foundation.*`. Operator IAM yang perlu melihat daftar mitra/`parties` saat provisioning akun mitra harus diberikan role terpisah yang memegang `enterprise_foundation.parties.read`.
> - **Penolakan Scope Entitas di IAM Runtime Saat Ini:** [scope_verification.go](../../../backend/services/iam/domain/service/scope_verification.go) (`RequireScopeOwnerVerification`) saat ini baru meluluskan grant `global` dan `self`, sedangkan penugasan role/group pada scope `legal_entity`, `organization_unit`, `facility`, `supplier_party`, atau `vendor_party` masih ditolak *fail-closed* (`iam.scope.owner_verification_unavailable`) di server IAM sampai kontrak verifikasi kepemilikan ke gRPC Enterprise Foundation dihubungkan pada wiring IAM.

### Matriks Keterikatan Permission ke Seluruh System Role Seeded (`iam.json` & `enterprise_foundation.json`)

| Permission Code | `superadmin` | `ef_automation` | `iam_operator` | `iam_readonly` | `customer` | `iam_selfservice` | `iam_automation` |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `enterprise_foundation.addresses.read` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.addresses.write` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.legal_entities.read` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.legal_entities.write` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.locations.read` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.locations.write` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.organization_units.read` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.organization_units.write` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.parties.read` | ✓ | ✓ | — | — | — | — | — |
| `enterprise_foundation.parties.write` | ✓ | ✓ | — | — | — | — | — |

---

## 3. Matriks Operasi GraphQL Enterprise Foundation ke Permission, Sensitivitas & Scope

Seluruh operasi GraphQL Enterprise Foundation tidak memerlukan re-autentikasi sensitif untuk aksi baca (`read`), namun penegakan **scope** bervariasi per operasi sesuai resolver dan adapter di [internal/adapter/enterprise_foundation/](../../../backend/gateway/internal/adapter/enterprise_foundation/).

### 3.1 Query, Relay `Node` & Field Resolver Bersarang

#### Relay Global `Query.node` Dispatch (`IsEFNodeID`)

Ketika `Query.node(id: ID!)` di [schema.resolvers.go](../../../backend/gateway/graph/resolvers/schema.resolvers.go) menerima awalan `gid://integrione/<resource>/<uuid>` yang dikenali `IsEFNodeID` ([gid.go](../../../backend/gateway/internal/adapter/enterprise_foundation/gid.go)), pemanggilan diteruskan ke `Adapter.Node` ([adapter.go](../../../backend/gateway/internal/adapter/enterprise_foundation/adapter.go)) dan mewarisi otorisasi query tunggal entitas terkait:

| Prefix Global Object ID (`gid://integrione/...`) | Delegasi Adapter | Permission Code | Scope Dievaluasi Gateway |
| --- | --- | --- | --- |
| `address/<uuid>` | `Address` | `enterprise_foundation.addresses.read` | `global` |
| `legal-entity/<uuid>` | `LegalEntity` | `enterprise_foundation.legal_entities.read` | `legal_entity:<uuid>` \| `global` |
| `location/<uuid>` | `Location` | `enterprise_foundation.locations.read` | `facility:<uuid>` \| `global` |
| `organization-unit/<uuid>` | `OrganizationUnit` | `enterprise_foundation.organization_units.read` | `organization_unit:<uuid>` \| `global` |
| `party/<uuid>` | `Party` | `enterprise_foundation.parties.read` | `vendor_party:<uuid>` \| `supplier_party:<uuid>` \| `global` |
| `party-role/<uuid>` | `PartyRole` | `enterprise_foundation.parties.read` | `global` |
| `party-contact/<uuid>` | `PartyContact` | `enterprise_foundation.parties.read` | `global` |
| `party-address/<uuid>` | `PartyAddressLink` | `enterprise_foundation.parties.read` | `global` |

#### Domain Master Alamat (`Address` — 5 Query)

| Operasi / Field GraphQL | Permission Code | Scope Dievaluasi Gateway | Guard & Perilaku Khusus |
| --- | --- | --- | --- |
| `Query.address(id)` | `enterprise_foundation.addresses.read` | `global` | Memvalidasi `ResourceAddress` UUID/GID via `ParseTypedID`. |
| `Query.addressesByIds(ids)` | `enterprise_foundation.addresses.read` | `global` | Batch fetch; mengembalikan `[]` bila `ids` kosong setelah validasi permission. |
| `Query.addresses(...)` | `enterprise_foundation.addresses.read` | `global` | Relay pagination filter teks, `country_code`, `city`, `province`, `postal_code`. |
| `Query.allAddresses(...)` | `enterprise_foundation.addresses.read` | `global` | Seluruh master alamat terpaginasi. |
| `Query.isAddressReferenced(id)` | `enterprise_foundation.addresses.read` | `global` | Memeriksa apakah master alamat masih dirujuk entitas lain sebelum penghapusan. |

#### Domain Badan Hukum (`LegalEntity` — 10 Query + 3 Nested Resolvers)

| Operasi / Field GraphQL | Permission Code | Scope Dievaluasi Gateway | Guard & Perilaku Khusus |
| --- | --- | --- | --- |
| `Query.legalEntity(id)` | `enterprise_foundation.legal_entities.read` | `legal_entity:<id>` \| `global` | Bila parse ID gagal, tetap jalankan `requirePerm(global)` lebih dulu agar tidak membocorkan error input kepada pemanggil ilegal. |
| `Query.legalEntitiesByIds(ids)` | `enterprise_foundation.legal_entities.read` | Per ID: `legal_entity:<id>` \| `global` | Wajib lulus `global` jika `ids` kosong atau gagal parse tanpa izin `global`; jika tidak `global`, setiap `id` dalam `ids` diperiksa satu per satu. |
| `Query.legalEntityByCode(entity_code)` | `enterprise_foundation.legal_entities.read` | `global` | Pencarian lookup berdasarkan `entity_code` non-kosong. |
| `Query.legalEntityByTaxIdNumber(tax_id_number)` | `enterprise_foundation.legal_entities.read` | `global` | Pencarian lookup berdasarkan NPWP/nomor pajak non-kosong. |
| `Query.legalEntitiesByNib(nib, ...)` | `enterprise_foundation.legal_entities.read` | `global` | Daftar badan hukum terpaginasi berdasarkan NIB non-kosong. |
| `Query.primaryOperatingLegalEntity` | `enterprise_foundation.legal_entities.read` | `global` | Mengambil badan hukum operasi utama perusahaan. |
| `Query.legalEntities(...)` | `enterprise_foundation.legal_entities.read` | `legal_entity:<parent_legal_entity_id>` \| `global` (bila `parent_legal_entity_id` valid); selain itu `global` | Mendukung filter `query`, `parent_legal_entity_id`, `status`, `primary_only`, `include_deleted`. |
| `Query.allLegalEntities(...)` | `enterprise_foundation.legal_entities.read` | `global` | Seluruh badan hukum terpaginasi. |
| `Query.legalEntityChildren(parent_legal_entity_id, ...)` | `enterprise_foundation.legal_entities.read` | `legal_entity:<parent_legal_entity_id>` \| `global` | Anak badan hukum langsung dari induk tertentu. |
| `Query.validateLegalEntityScopeTarget(id)` | `enterprise_foundation.legal_entities.read` | `legal_entity:<id>` \| `global` | `ScopePolicy.EvaluateLegalEntity` memverifikasi badan hukum aktif/eligible sebelum dipakai sebagai target scope `legal_entity` IAM/modul lain. |
| `LegalEntity.parent_legal_entity` | `enterprise_foundation.legal_entities.read` | `legal_entity:<parent_legal_entity_id>` \| `global` | Mengembalikan `null` jika `parent_legal_entity_id` kosong; mendelegasikan ke `Adapter.LegalEntity`. |
| `LegalEntity.registered_address` | **`enterprise_foundation.addresses.read`** | `global` | Mengembalikan `null` jika `registered_address_id` kosong; mendelegasikan ke `Adapter.Address`. Pemanggil yang hanya punya izin `legal_entities.read` tetap membutuhkan `addresses.read` untuk field ini. |
| `LegalEntity.children(...)` | `enterprise_foundation.legal_entities.read` | `legal_entity:<obj.RawID>` \| `global` | Mengembalikan `null` jika parent nil; mendelegasikan ke `Adapter.LegalEntityChildren`. |

#### Domain Lokasi Fisik Operasional (`Location` — 10 Query + 4 Nested Resolvers)

| Operasi / Field GraphQL | Permission Code | Scope Dievaluasi Gateway | Guard & Perilaku Khusus |
| --- | --- | --- | --- |
| `Query.location(id)` | `enterprise_foundation.locations.read` | `facility:<id>` \| `global` | Jika parse ID gagal, tetap cek `requirePerm(global)` terlebih dahulu. |
| `Query.locationsByIds(ids)` | `enterprise_foundation.locations.read` | Per ID: `facility:<id>` \| `global` | Sama seperti batch lookup lainnya: wajib `global` untuk `ids` kosong/parse error, atau verifikasi scope `facility` untuk tiap ID. |
| `Query.locationByCode(location_code)` | `enterprise_foundation.locations.read` | `global` | Lookup tunggal berdasarkan `location_code` non-kosong. |
| `Query.locationsByLegalEntityAndKind(legal_entity_id, kind, ...)` | `enterprise_foundation.locations.read` | `legal_entity:<legal_entity_id>` \| `global` | Pemegang scope `legal_entity` diizinkan membaca lokasi di dalam badan hukumnya. |
| `Query.locationsByAddressId(address_id, ...)` | `enterprise_foundation.locations.read` | `global` | Mencari daftar lokasi yang memakai `address_id` tertentu. |
| `Query.locationChildren(parent_location_id, ...)` | `enterprise_foundation.locations.read` | `facility:<parent_location_id>` \| `global` | Sub-lokasi langsung di bawah lokasi induk. |
| `Query.locations(...)` | `enterprise_foundation.locations.read` | `legal_entity:<legal_entity_id>` \| `global` (bila `legal_entity_id` valid); selain itu `global` | Mendukung filter `query`, `legal_entity_id`, `parent_location_id`, `address_id`, `kind`, `active_only`, `include_deleted`. |
| `Query.allLocations(...)` | `enterprise_foundation.locations.read` | `global` | Seluruh lokasi fisik terpaginasi. |
| `Query.validateFacilityScope(id)` | `enterprise_foundation.locations.read` | `facility:<id>` \| `global` | `ScopePolicy.EvaluateFacility` mengevaluasi lokasi beroperasi (`CanOperate`) dan badan hukum pemilik aktif -> mengembalikan `scope_kind: FACILITY`. |
| `Query.validateWarehouseScope(id)` | `enterprise_foundation.locations.read` | `facility:<id>` \| `global` | `ScopePolicy.EvaluateWarehouse` menambah guard wajib `Location.Kind == WAREHOUSE` -> gagal *fail-closed* jika lokasi bukan gudang. |
| `Location.legal_entity` | **`enterprise_foundation.legal_entities.read`** | `legal_entity:<legal_entity_id>` \| `global` | Mendelegasikan ke `Adapter.LegalEntity`; membutuhkan izin `legal_entities.read`. |
| `Location.parent_location` | `enterprise_foundation.locations.read` | `facility:<parent_location_id>` \| `global` | Mengembalikan `null` jika `parent_location_id` kosong. |
| `Location.address` | **`enterprise_foundation.addresses.read`** | `global` | Mengembalikan `null` jika `address_id` kosong; membutuhkan izin `addresses.read`. |
| `Location.children(...)` | `enterprise_foundation.locations.read` | `facility:<obj.RawID>` \| `global` | Mendelegasikan ke `Adapter.LocationChildren`. |

#### Domain Unit Organisasi (`OrganizationUnit` — 9 Query + 4 Nested Resolvers)

| Operasi / Field GraphQL | Permission Code | Scope Dievaluasi Gateway | Guard & Perilaku Khusus |
| --- | --- | --- | --- |
| `Query.organizationUnit(id)` | `enterprise_foundation.organization_units.read` | `organization_unit:<id>` \| `global` | Parse ID gagal tetap dicek `requirePerm(global)` terlebih dahulu. |
| `Query.organizationUnitsByIds(ids)` | `enterprise_foundation.organization_units.read` | Per ID: `organization_unit:<id>` \| `global` | Batch fetch; tiap ID diverifikasi scope `organization_unit`-nya. |
| `Query.organizationUnitByEntityAndCode(legal_entity_id, unit_code)` | `enterprise_foundation.organization_units.read` | `legal_entity:<legal_entity_id>` \| `global` | Lookup unit unik per badan hukum dan `unit_code`. |
| `Query.organizationUnitsByHeadPersonId(head_person_id, ...)` | `enterprise_foundation.organization_units.read` | `global` | Pencarian unit berdasarkan `head_person_id` (UUID atau GID umum via `ParseUUIDOrGID`). |
| `Query.organizationUnitsByPrimaryLocationId(primary_location_id, ...)` | `enterprise_foundation.organization_units.read` | `facility:<primary_location_id>` \| `global` | Pemegang scope `facility` diizinkan melihat unit yang berbasis pada lokasi tersebut. |
| `Query.organizationUnitChildren(parent_unit_id, ...)` | `enterprise_foundation.organization_units.read` | `organization_unit:<parent_unit_id>` \| `global` | Sub-unit langsung di bawah unit induk. |
| `Query.organizationUnits(...)` | `enterprise_foundation.organization_units.read` | `legal_entity:<legal_entity_id>` \| `global` (bila `legal_entity_id` valid); selain itu `global` | Mendukung filter `query`, `legal_entity_id`, `parent_unit_id`, `primary_location_id`, `kind`, `active_only`, `include_deleted`. |
| `Query.allOrganizationUnits(...)` | `enterprise_foundation.organization_units.read` | `global` | Seluruh unit organisasi terpaginasi. |
| `Query.validateOrganizationUnitScopeTarget(id)` | `enterprise_foundation.organization_units.read` | `organization_unit:<id>` \| `global` | `ScopePolicy.EvaluateOrganizationUnit` memverifikasi unit beroperasi serta `LegalEntity` pemilik aktif -> `scope_kind: ORGANIZATION_UNIT`. |
| `OrganizationUnit.legal_entity` | **`enterprise_foundation.legal_entities.read`** | `legal_entity:<legal_entity_id>` \| `global` | Membutuhkan izin `legal_entities.read`. |
| `OrganizationUnit.parent_unit` | `enterprise_foundation.organization_units.read` | `organization_unit:<parent_unit_id>` \| `global` | Mengembalikan `null` jika `parent_unit_id` kosong. |
| `OrganizationUnit.primary_location` | **`enterprise_foundation.locations.read`** | `facility:<primary_location_id>` \| `global` | Mengembalikan `null` jika `primary_location_id` kosong; membutuhkan izin `locations.read`. |
| `OrganizationUnit.children(...)` | `enterprise_foundation.organization_units.read` | `organization_unit:<obj.RawID>` \| `global` | Mendelegasikan ke `Adapter.OrganizationUnitChildren`. |

#### Domain Mitra Eksternal, Peran, PIC & Alamat Mitra (`Party` — 30 Query + 4 Nested Resolvers)

> **Catatan `partyScopes(partyID)`:** Untuk operasi bertarget `partyID`, adapter mengevaluasi apakah caller memiliki izin dengan scope `vendor_party:<partyID>` **atau** `supplier_party:<partyID>` (yang masing-masing terpenuhi pula oleh grant `global`).

| Operasi / Field GraphQL | Permission Code | Scope Dievaluasi Gateway | Guard & Perilaku Khusus |
| --- | --- | --- | --- |
| `Query.party(id)` | `enterprise_foundation.parties.read` | `vendor_party:<id>` \| `supplier_party:<id>` \| `global` | Profil mitra eksternal B2B tunggal. |
| `Query.partiesByIds(ids)` | `enterprise_foundation.parties.read` | Per ID: `vendor_party:<id>` \| `supplier_party:<id>` \| `global` | Batch fetch; tiap ID diperiksa via `partyScopes(id)`. |
| `Query.partyByCode(partner_code)` | `enterprise_foundation.parties.read` | `global` | Lookup tunggal berdasarkan `partner_code`. |
| `Query.partyByTaxIdNumber(tax_id_number)` | `enterprise_foundation.parties.read` | `global` | Lookup tunggal berdasarkan NPWP mitra. |
| `Query.partiesByEmail(email, ...)` | `enterprise_foundation.parties.read` | `global` | Daftar mitra terpaginasi berdasarkan email. |
| `Query.parties(...)` | `enterprise_foundation.parties.read` | `global` | Daftar mitra terpaginasi dengan filter `query`, `kind`, `status`, `role`, `include_deleted`. |
| `Query.allParties(...)` | `enterprise_foundation.parties.read` | `global` | Seluruh daftar mitra terpaginasi. |
| `Party.roles(...)` | `enterprise_foundation.parties.read` | `vendor_party:<obj.RawID>` \| `supplier_party:<obj.RawID>` \| `global` | Mendelegasikan ke `Adapter.PartyRolesByParty`. |
| `Party.contacts(...)` | `enterprise_foundation.parties.read` | `vendor_party:<obj.RawID>` \| `supplier_party:<obj.RawID>` \| `global` | Mendelegasikan ke `Adapter.PartyContactsByParty`. |
| `Party.addresses(...)` | `enterprise_foundation.parties.read` | `vendor_party:<obj.RawID>` \| `supplier_party:<obj.RawID>` \| `global` | Mendelegasikan ke `Adapter.PartyAddressLinksByParty`. |
| `Query.partyRole(id)` | `enterprise_foundation.parties.read` | `global` | Lookup sub-entitas `PartyRole` berdasarkan ID/GID role. |
| `Query.partyRolesByParty(party_id, ...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Daftar peran bisnis milik satu mitra. |
| `Query.partyRoleByPartyAndKind(party_id, kind)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Lookup peran spesifik milik satu mitra. |
| `Query.partyRoles(...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` (bila `party_id` valid); selain itu `global` | Filter `party_id`, `role`, `active_only`, `query`. |
| `Query.allPartyRoles(...)` | `enterprise_foundation.parties.read` | `global` | Seluruh penugasan peran bisnis mitra. |
| `Query.partyRolesByKind(kind, active_only, ...)` | `enterprise_foundation.parties.read` | `global` | Daftar peran berdasarkan jenis (`SUPPLIER`, `VENDOR`, `CUSTOMER_B2B`, `CARRIER`, `FINANCIAL_INSTITUTION`, `OTHER`). |
| `Query.allPartyRolesByKind(kind, active_only, ...)` | `enterprise_foundation.parties.read` | `global` | Sama di atas tanpa filter lanjutan. |
| `Query.validatePartyRoleScope(party_id, required_role)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | `ScopePolicy.EvaluatePartyRole` memastikan mitra aktif & peran aktif; memetakan `SUPPLIER -> SUPPLIER_PARTY` dan `VENDOR -> VENDOR_PARTY`. |
| `Query.partyContact(id)` | `enterprise_foundation.parties.read` | `global` | Lookup sub-entitas PIC `PartyContact` berdasarkan ID/GID kontak. |
| `Query.partyContactsByParty(party_id, ...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Daftar PIC milik satu mitra. |
| `Query.primaryPartyContact(party_id)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | PIC utama milik satu mitra. |
| `Query.partyContactsByIamUserId(iam_user_id, ...)` | `enterprise_foundation.parties.read` | `global` | Daftar PIC yang tertaut ke `iam_user_id` tertentu (`ResourceIAMUser` / `gid://integrione/user/<uuid>` atau UUID mentah). |
| `Query.partyContactsByEmail(email, ...)` | `enterprise_foundation.parties.read` | `global` | Daftar PIC berdasarkan email. |
| `Query.partyContacts(...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` (bila `party_id` valid); selain itu `global` | Filter `party_id`, `query`, `contact_role`, `portal_access_status`, `active_only`, `include_deleted`. |
| `Query.allPartyContacts(...)` | `enterprise_foundation.parties.read` | `global` | Seluruh PIC mitra terpaginasi. |
| `Query.evaluateRepresentativePortalAccess(party_id, iam_user_id, required_role)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | `PortalAccessPolicy.EvaluateRepresentativeAccess` memverifikasi `Party` layak transaksi, `PartyContact` aktif + `portal_access_status == active` + `iam_user_id` cocok, serta `required_role` aktif pada `PartyRole`. |
| `Query.partyAddressLink(id)` | `enterprise_foundation.parties.read` | `global` | Lookup relasi alamat mitra berdasarkan ID/GID link. |
| `Query.partyAddressLinksByParty(party_id, ...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Daftar tautan alamat milik satu mitra. |
| `Query.partyAddressLinksByPartyAndUsage(party_id, usage, ...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Daftar tautan alamat milik mitra menurut peruntukan (`REGISTERED_OFFICE`, `HEADQUARTERS`, `BILLING`, `SHIPPING`, `OTHER`). |
| `Query.primaryPartyAddressLink(party_id, usage)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` | Tautan alamat utama mitra untuk peruntukan tertentu. |
| `Query.partyAddressLinksByAddressId(address_id, ...)` | `enterprise_foundation.parties.read` | `global` | Daftar tautan mitra yang merujuk pada suatu `address_id`. |
| `Query.partyAddressLinks(...)` | `enterprise_foundation.parties.read` | `partyScopes(party_id)` \| `global` (bila `party_id` valid); selain itu `global` | Filter `party_id`, `address_id`, `usage`, `primary_only`. |
| `Query.allPartyAddressLinks(...)` | `enterprise_foundation.parties.read` | `global` | Seluruh relasi alamat mitra terpaginasi. |
| `PartyAddress.address` | **`enterprise_foundation.addresses.read`** | `global` | Mendelegasikan ke `Adapter.Address`; pemanggil tetap membutuhkan izin `addresses.read` untuk membaca detail `Address`. |

---

### 3.2 Mutation (48 Mutasi Terdaftar: 12 Sensitif, 36 Standar)

Seluruh mutasi mengembalikan `MutationResult` (`errors`, `was_successful`). Penolakan autentikasi/otorisasi/step-up dari Gateway atau downstream gRPC dipetakan oleh `ToMutationErrors` ([mutation_result.go](../../../backend/gateway/internal/adapter/enterprise_foundation/mutation_result.go)) menjadi `AUTHENTICATION`, `AUTHORIZATION`, atau `UNAVAILABLE` tanpa menyentuh downstream EF bila gagal di Gateway.

#### Mutasi Address (4 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `createAddress` | `enterprise_foundation.addresses.write` | Tidak | `global` | Membuat master alamat geografis baru. |
| `findOrCreateAddress` | `enterprise_foundation.addresses.write` | Tidak | `global` | Mencari alamat identik secara idempoten atau membuat baru bila belum ada. |
| `updateAddress` | `enterprise_foundation.addresses.write` | Tidak | `global` | Memperbarui atribut master alamat (`ResourceAddress`). |
| `deleteAddress` | `enterprise_foundation.addresses.write` | **Ya** | `global` | Penghapusan alamat wajib login terbaru `<= 10 menit` + `MFA` aktif. |

#### Mutasi LegalEntity (7 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `createLegalEntity` | `enterprise_foundation.legal_entities.write` | Tidak | `legal_entity:<parent_legal_entity_id>` \| `global` (bila `parent_legal_entity_id` valid); selain itu `global` | Opsional memvalidasi `registered_address_id` sebagai `ResourceAddress`. |
| `updateLegalEntity` | `enterprise_foundation.legal_entities.write` | Tidak | `legal_entity:<id>` \| `global` | Memperbarui profil, induk, alamat terdaftar, maupun status PKP/NPWP. |
| `setPrimaryOperatingLegalEntity` | `enterprise_foundation.legal_entities.write` | **Ya** | `legal_entity:<id>` \| `global` | Menetapkan badan hukum operasi utama perusahaan; wajib step-up `10m + MFA`. |
| `activateLegalEntity` | `enterprise_foundation.legal_entities.write` | Tidak | `legal_entity:<id>` \| `global` | Mengubah status badan hukum ke `ACTIVE`. |
| `suspendLegalEntity` | `enterprise_foundation.legal_entities.write` | Tidak | `legal_entity:<id>` \| `global` | Membekukan sementara badan hukum (`SUSPENDED`). |
| `dissolveLegalEntity` | `enterprise_foundation.legal_entities.write` | **Ya** | `legal_entity:<id>` \| `global` | Membubarkan badan hukum secara permanen (`DISSOLVED`); wajib step-up `10m + MFA`. |
| `deleteLegalEntity` | `enterprise_foundation.legal_entities.write` | **Ya** | `legal_entity:<id>` \| `global` | Soft-delete badan hukum (ditolak jika masih ada dependen); wajib step-up `10m + MFA`. |

#### Mutasi Location (5 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `createLocation` | `enterprise_foundation.locations.write` | Tidak | `legal_entity:<legal_entity_id>` \| `global` | Pemilik scope `legal_entity` dapat membuat lokasi di bawah badan hukumnya. |
| `updateLocation` | `enterprise_foundation.locations.write` | Tidak | `facility:<id>` \| `global` | Memperbarui atribut lokasi, induk, alamat, zona waktu, dan atribut operasional. |
| `activateLocation` | `enterprise_foundation.locations.write` | Tidak | `facility:<id>` \| `global` | Mengaktifkan kembali lokasi fisik. |
| `deactivateLocation` | `enterprise_foundation.locations.write` | Tidak | `facility:<id>` \| `global` | Menonaktifkan lokasi fisik tanpa menghapus riwayat. |
| `deleteLocation` | `enterprise_foundation.locations.write` | **Ya** | `facility:<id>` \| `global` | Soft-delete lokasi fisik; wajib step-up `10m + MFA`. |

#### Mutasi OrganizationUnit (6 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `createOrganizationUnit` | `enterprise_foundation.organization_units.write` | Tidak | `legal_entity:<legal_entity_id>` \| `global` | Pemilik scope `legal_entity` dapat membentuk unit kerja di bawah badan hukumnya. |
| `updateOrganizationUnit` | `enterprise_foundation.organization_units.write` | Tidak | `organization_unit:<id>` \| `global` | Memperbarui struktur, lokasi utama, cost center, dan `head_person_id`. |
| `assignOrganizationUnitHeadPerson` | `enterprise_foundation.organization_units.write` | Tidak | `organization_unit:<id>` \| `global` | Menetapkan atau mengosongkan `head_person_id` (UUID/GID umum). |
| `activateOrganizationUnit` | `enterprise_foundation.organization_units.write` | Tidak | `organization_unit:<id>` \| `global` | Mengaktifkan unit organisasi. |
| `deactivateOrganizationUnit` | `enterprise_foundation.organization_units.write` | Tidak | `organization_unit:<id>` \| `global` | Menonaktifkan unit organisasi. |
| `deleteOrganizationUnit` | `enterprise_foundation.organization_units.write` | **Ya** | `organization_unit:<id>` \| `global` | Soft-delete unit organisasi; wajib step-up `10m + MFA`. |

#### Mutasi Party Core (7 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `createParty` | `enterprise_foundation.parties.write` | Tidak | `global` | Mendaftarkan mitra eksternal B2B baru (default status `PENDING_VERIFICATION`). |
| `updatePartyProfile` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(id)` \| `global` | Memperbarui profil umum, email, telepon, bahasa komunikasi. |
| `updatePartyTaxProfile` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(id)` \| `global` | Memperbarui NPWP, nama pajak, status PKP, dan pengecualian pajak. |
| `verifyAndActivateParty` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(id)` \| `global` | Memverifikasi dan mengaktifkan mitra agar layak bertransaksi (`ACTIVE`). |
| `suspendParty` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(id)` \| `global` | Menangguhkan mitra sementara (`SUSPENDED`). |
| `deactivateParty` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(id)` \| `global` | Menonaktifkan mitra (`INACTIVE`). |
| `deleteParty` | `enterprise_foundation.parties.write` | **Ya** | `partyScopes(id)` \| `global` | Soft-delete entitas mitra; wajib step-up `10m + MFA`. |

#### Mutasi PartyRole (5 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `assignPartyRole` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(party_id)` \| `global` | Menugaskan peran bisnis (`SUPPLIER`, `VENDOR`, `CUSTOMER_B2B`, `CARRIER`, `FINANCIAL_INSTITUTION`, `OTHER`). |
| `updatePartyRoleNote` | `enterprise_foundation.parties.write` | Tidak | `global` | Input hanya membawa `role_id` (`ResourcePartyRole`), sehingga Gateway memeriksa scope `global`. |
| `activatePartyRole` | `enterprise_foundation.parties.write` | Tidak | `global` | Mengaktifkan peran berdasarkan `role_id` (butuh `global`). |
| `deactivatePartyRole` | `enterprise_foundation.parties.write` | Tidak | `global` | Menonaktifkan peran berdasarkan `role_id` (butuh `global`). |
| `removePartyRole` | `enterprise_foundation.parties.write` | **Ya** | `global` | Menghapus penugasan peran bisnis mitra; wajib `global` dan step-up `10m + MFA`. |

#### Mutasi PartyContact & Akses Portal IAM (10 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `addPartyContact` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(party_id)` \| `global` | Menambahkan PIC; bila `portal_access_status = ACTIVE`, service EF wajib memvalidasi `Party` layak transaksi serta belum ada duplikasi `iam_user_id` di party sama. |
| `updatePartyContact` | `enterprise_foundation.parties.write` | Tidak | `global` | Input hanya membawa `contact_id` (`ResourcePartyContact`); mengevaluasi scope `global`. |
| `setPrimaryPartyContact` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(party_id)` \| `global` | Menetapkan satu PIC aktif sebagai kontak utama mitra secara atomik. |
| `requestPartyPortalLink` | `enterprise_foundation.parties.write` | Tidak | `global` | Mengajukan penautan `contact_id` dengan `iam_user_id` (`PENDING_LINK`) dan memicu event outbox portal. |
| `activatePartyPortalAccess` | `enterprise_foundation.parties.write` | **Ya** | `global` | Mengaktifkan akses portal (`ACTIVE`) jika mitra layak transaksi & transisi `PortalAccessPolicy` sah; wajib `global` + step-up `10m + MFA`. |
| `revokePartyPortalAccess` | `enterprise_foundation.parties.write` | **Ya** | `global` | Mencabut akses portal (`REVOKED`) sambil menyimpan jejak kontak; wajib `global` + step-up `10m + MFA`. |
| `unlinkPartyPortalUser` | `enterprise_foundation.parties.write` | **Ya** | `global` | Memutus `iam_user_id` dari PIC (`NONE`); wajib `global` + step-up `10m + MFA`. |
| `activatePartyContact` | `enterprise_foundation.parties.write` | Tidak | `global` | Mengaktifkan kembali PIC berdasarkan `contact_id`. |
| `deactivatePartyContact` | `enterprise_foundation.parties.write` | Tidak | `global` | Menonaktifkan sementara PIC beserta status portal efektifnya. |
| `deletePartyContact` | `enterprise_foundation.parties.write` | **Ya** | `global` | Soft-delete PIC sekaligus mencabut portal aktif bila tertaut IAM; wajib `global` + step-up `10m + MFA`. |

#### Mutasi PartyAddress (4 Mutasi)

| Mutation GraphQL | Permission Code | Sensitif (Gateway) | Scope Dievaluasi Gateway | Keterangan & Guard Utama |
| --- | --- | :---: | --- | --- |
| `linkPartyAddress` | `enterprise_foundation.parties.write` | Tidak | `partyScopes(party_id)` \| `global` | Menautkan `address_id` (`ResourceAddress`) ke mitra dengan `usage` tertentu. |
| `updatePartyAddressLink` | `enterprise_foundation.parties.write` | Tidak | `global` | Input hanya membawa `link_id` (`ResourcePartyAddress`); mengevaluasi scope `global`. |
| `setPrimaryPartyAddressLink` | `enterprise_foundation.parties.write` | Tidak | `global` | Menetapkan satu tautan alamat mitra sebagai utama untuk peruntukannya. |
| `unlinkPartyAddress` | `enterprise_foundation.parties.write` | Tidak | `global` | Menghapus relasi alamat dari mitra berdasarkan `link_id`. |

---

## 4. Prinsip Trust Boundary, Isolasi & Ketahanan

1. **Verifikasi IAM Fail-Closed sebelum RPC & Proteksi Oracle ID:** Gateway selalu memeriksa autentikasi sesi (`RequireAuthenticated` dengan kesegaran `AccessSnapshot <= 30s`), kepemilikan permission & scope (`RequirePermission` / `requireAnyScope`), serta jendela step-up sensitif (`<= 10 menit` + `MFA` aktif) sebelum mengirimkan panggilan gRPC ke Enterprise Foundation ([permissions.go](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go), [authorizer.go](../../../backend/gateway/internal/middleware/authorizer.go)). Apabila parsing ID/GID gagal pada operasi berscope entitas, adapter tetap mengeksekusi `requirePerm(..., global, requireSensitive)` terlebih dulu agar pemanggil tanpa sesi atau tanpa izin tidak memperoleh bocoran error validasi input; permintaan ditolak `UNAUTHENTICATED` / `AUTHORIZATION` dengan `0` panggilan gRPC ke service EF.
2. **Pewarisan Scope ABAC & Konsistensi Step-Up Fallback:** Menurut `ScopeGrant.Matches` ([snapshot.go](../../../backend/shared/authz/snapshot.go)), pemegang grant `global` otomatis memenuhi pemeriksaan `legal_entity`, `organization_unit`, `facility`, `vendor_party`, dan `supplier_party`, sedangkan pemegang grant berscope entitas ditolak pada target entitas berbeda maupun pada query/mutasi yang menuntut `global`. Di `requireAnyScope` (seperti `partyScopes` yang mencoba `vendor_party` lalu `supplier_party`), penolakan step-up re-autentikasi (`!strings.HasPrefix(gwErr.Message, "missing required permission:")`) menghentikan loop secara instan tanpa mencoba scope fallback.
3. **Isolasi Trust Boundary Gateway ↔ gRPC EF & Keamanan Sesi Browser:** Service Enterprise Foundation tidak membaca database IAM dan hanya memverifikasi `CallerContext` bertanda tangan HMAC-SHA256 dari Gateway melalui `InternalContextUnaryInterceptor` ([interceptor.go](../../../backend/services/enterprise_foundation/transport/grpc/interceptor/interceptor.go)) selain metode healthcheck (`/grpc.health.v1.*`) dan reflection (`/grpc.reflection.*`). Penolakan `codes.Unauthenticated` dari downstream EF diterjemahkan oleh `MapEFGRPCError` ([ef_client.go](../../../backend/gateway/internal/source_connections/ef_client.go)) menjadi `CodeUnavailable` (`UNAVAILABLE`), sehingga kegagalan secret internal tidak pernah dianggap sesi browser kedaluwarsa, tidak menghapus cookie login pengguna, dan tidak menjatuhkan layanan IAM.
4. **Ketahanan Startup, Sanitasi Error & Circuit Breaker:** Konstruktor `NewAdapter` mengizinkan client EF bernilai `nil` saat modul EF tidak diaktifkan sehingga Gateway tetap menyala dan operasi non-EF (seperti IAM) tetap berjalan normal, sedangkan operasi EF mengembalikan `UNAVAILABLE`. Pemanggilan gRPC EF dilindungi `circuitBreaker` (`5` kegagalan beruntun membuka sirkuit selama `3s` sebelum pulih otomatis tanpa restart Gateway) serta `sanitizeEFMessage` agar detail IP, host, skema database PostgreSQL, atau HMAC internal tidak bocor ke klien GraphQL.
5. **Keterhubungan Logis IAM–Mitra (`party_contacts.iam_user_id`) Tanpa Auto-Grant:** Kolom `party_contacts.iam_user_id` hanya menjadi referensi logis ke `iam.users.id` (tanpa FK SQL lintas database). Penautan atau aktivasi portal PIC di EF (`requestPartyPortalLink`, `activatePartyPortalAccess`) menerbitkan event outbox domain (`foundation_outbox_events`, default `EF_OUTBOX_MODE=standby` selama belum ada konsumen eksternal aktif) dan dipakai untuk validasi `evaluateRepresentativePortalAccess`, **namun tidak pernah** membuat user IAM maupun menambahkan role/scope RBAC secara otomatis di database IAM ([users.md](../../users.md)).
6. **Isolasi Akun Mesin `auto_system_ef`:** Akun non-interaktif `auto_system_ef` (`user_kind = system`) terikat pada role `ef_automation` untuk otorisasi konteks layanan domain EF dan secara kontrak IAM dilarang keras melakukan login interaktif (`ErrSystemCannotLogin` & `ErrSystemUserInteractiveForbidden`).
