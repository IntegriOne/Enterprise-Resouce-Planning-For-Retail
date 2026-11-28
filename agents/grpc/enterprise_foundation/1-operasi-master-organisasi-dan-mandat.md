# List operasi Enterprise Foundation

Baseline 28 November 2026 pada HEAD `c49b9e1119409aa38d58bf9727e37d8360c7944e` dan kode produksi dalam checkout. Terdapat **64 query, 48 mutation, 112 RPC domain pada lima service**. Setiap baris dicocokkan terhadap field root SDL, method resolver, pemanggilan RPC dalam adapter dan deklarasi proto. `node` bersama, health dan reflection berada di luar hitungan root domain.

| Kelompok | Query | Mutation | RPC |
| --- | ---: | ---: | ---: |
| Alamat | 5 | 4 | 9 |
| Badan hukum | 10 | 7 | 17 |
| Lokasi | 10 | 5 | 15 |
| Unit organisasi | 9 | 6 | 15 |
| Mitra, peran, PIC dan alamat mitra | 30 | 26 | 56 |
| **Total** | **64** | **48** | **112** |

## RBAC dan pembacaan tabel

Awalan permission adalah `enterprise_foundation.`. Gunakan resource yang disebut di setiap kelompok. Kolom scope menjelaskan pemeriksaan di adapter saat ini. “Target” menunjuk target/scope yang dihitung adapter (misalnya party asal atau parent). Beberapa finder memilih jalur global atau target sesuai filter. Kolom sensitif mencatat bila adapter meminta pemeriksaan sensitif pada jalur operasi. Autentikasi sesi, MFA dan kebijakan IAM tetap berlaku sesuai konteks. [Implementasi grant dan pemeriksaan sensitif](../../../backend/gateway/internal/adapter/enterprise_foundation/permissions.go) · [Aturan IAM](../../../agents/users.md).

## Alamat

Alamat master bersama dan hubungan referensinya. Perbaiki kesetaraan atomik pada create/find-or-create/update (EFL-07). [SDL](../../../backend/gateway/graphql/enterprise_foundation/ef_address.graphql) · [Resolver](../../../backend/gateway/graph/resolvers/ef_address.resolvers.go) · [Adapter](../../../backend/gateway/internal/adapter/enterprise_foundation/address_adapter.go) · [AddressService](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/address.proto).

| Operasi GraphQL | Butuh GraphQL | RPC | Permission | Scope di adapter | Sensitif |
| --- | --- | --- | --- | --- | --- |
| `address` | Query | `GetAddressById` | `addresses.read` | Global | Tidak |
| `addressesByIds` | Query | `GetAddressesByIds` | `addresses.read` | Global | Tidak |
| `addresses` | Query | `ListAddresses` | `addresses.read` | Global | Tidak |
| `allAddresses` | Query | `GetAllAddresses` | `addresses.read` | Global | Tidak |
| `isAddressReferenced` | Query | `IsAddressReferenced` | `addresses.read` | Global | Tidak |
| `createAddress` | Mutation | `CreateAddress` | `addresses.write` | Global | Tidak |
| `findOrCreateAddress` | Mutation | `FindOrCreateAddress` | `addresses.write` | Global | Tidak |
| `updateAddress` | Mutation | `UpdateAddress` | `addresses.write` | Global | Tidak |
| `deleteAddress` | Mutation | `DeleteAddress` | `addresses.write` | Global | Ya |

## Badan hukum

Identitas, badan hukum utama, validasi scope dan lifecycle. Update profil, status dan parent berbagi perlindungan transaksi (EFL-04/05). [SDL](../../../backend/gateway/graphql/enterprise_foundation/ef_legal_entity.graphql) · [Resolver](../../../backend/gateway/graph/resolvers/ef_legal_entity.resolvers.go) · [Adapter](../../../backend/gateway/internal/adapter/enterprise_foundation/legal_entity_adapter.go) · [LegalEntityService](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/legal_entity.proto).

| Operasi GraphQL | Butuh GraphQL | RPC | Permission | Scope di adapter | Sensitif |
| --- | --- | --- | --- | --- | --- |
| `legalEntity` | Query | `GetLegalEntityById` | `legal_entities.read` | Target | Tidak |
| `legalEntitiesByIds` | Query | `GetLegalEntitiesByIds` | `legal_entities.read` | Target | Tidak |
| `legalEntityByCode` | Query | `GetLegalEntityByCode` | `legal_entities.read` | Global | Tidak |
| `legalEntityByTaxIdNumber` | Query | `GetLegalEntityByTaxIdNumber` | `legal_entities.read` | Global | Tidak |
| `legalEntitiesByNib` | Query | `ListLegalEntitiesByNib` | `legal_entities.read` | Global | Tidak |
| `primaryOperatingLegalEntity` | Query | `GetPrimaryOperatingLegalEntity` | `legal_entities.read` | Global | Tidak |
| `legalEntities` | Query | `ListLegalEntities` | `legal_entities.read` | Target / global | Tidak |
| `allLegalEntities` | Query | `GetAllLegalEntities` | `legal_entities.read` | Global | Tidak |
| `legalEntityChildren` | Query | `ListLegalEntityChildren` | `legal_entities.read` | Target | Tidak |
| `validateLegalEntityScopeTarget` | Query | `ValidateLegalEntityScopeTarget` | `legal_entities.read` | Target | Tidak |
| `createLegalEntity` | Mutation | `CreateLegalEntity` | `legal_entities.write` | Target / global | Tidak |
| `updateLegalEntity` | Mutation | `UpdateLegalEntity` | `legal_entities.write` | Target | Tidak |
| `setPrimaryOperatingLegalEntity` | Mutation | `SetPrimaryOperatingLegalEntity` | `legal_entities.write` | Target | Ya |
| `activateLegalEntity` | Mutation | `ActivateLegalEntity` | `legal_entities.write` | Target | Tidak |
| `suspendLegalEntity` | Mutation | `SuspendLegalEntity` | `legal_entities.write` | Target | Tidak |
| `dissolveLegalEntity` | Mutation | `DissolveLegalEntity` | `legal_entities.write` | Target | Ya |
| `deleteLegalEntity` | Mutation | `DeleteLegalEntity` | `legal_entities.write` | Target | Ya |

## Lokasi

Fasilitas dan subtipe warehouse merupakan fakta EF. Mutasi hierarki mempertahankan legal entity, batas kedalaman dan konsistensi parent (EFL-05). [SDL](../../../backend/gateway/graphql/enterprise_foundation/ef_location.graphql) · [Resolver](../../../backend/gateway/graph/resolvers/ef_location.resolvers.go) · [Adapter](../../../backend/gateway/internal/adapter/enterprise_foundation/location_adapter.go) · [LocationService](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/location.proto).

| Operasi GraphQL | Butuh GraphQL | RPC | Permission | Scope di adapter | Sensitif |
| --- | --- | --- | --- | --- | --- |
| `location` | Query | `GetLocationById` | `locations.read` | Target | Tidak |
| `locationsByIds` | Query | `GetLocationsByIds` | `locations.read` | Target | Tidak |
| `locationByCode` | Query | `GetLocationByCode` | `locations.read` | Global | Tidak |
| `locationsByLegalEntityAndKind` | Query | `ListLocationsByLegalEntityAndKind` | `locations.read` | Target | Tidak |
| `locationsByAddressId` | Query | `ListLocationsByAddressId` | `locations.read` | Global | Tidak |
| `locationChildren` | Query | `ListLocationChildren` | `locations.read` | Target | Tidak |
| `locations` | Query | `ListLocations` | `locations.read` | Target / global | Tidak |
| `allLocations` | Query | `GetAllLocations` | `locations.read` | Global | Tidak |
| `validateFacilityScope` | Query | `ValidateFacilityScope` | `locations.read` | Target | Tidak |
| `validateWarehouseScope` | Query | `ValidateWarehouseScope` | `locations.read` | Target | Tidak |
| `createLocation` | Mutation | `CreateLocation` | `locations.write` | Target | Tidak |
| `updateLocation` | Mutation | `UpdateLocation` | `locations.write` | Target | Tidak |
| `activateLocation` | Mutation | `ActivateLocation` | `locations.write` | Target | Tidak |
| `deactivateLocation` | Mutation | `DeactivateLocation` | `locations.write` | Target | Tidak |
| `deleteLocation` | Mutation | `DeleteLocation` | `locations.write` | Target | Ya |

## Unit organisasi

Struktur, lokasi utama, kepala dan validasi unit. Ketiga jalur penetapan kepala memerlukan verifikasi penugasan HRM (EFL-06), perubahan parent mengikuti EFL-05. [SDL](../../../backend/gateway/graphql/enterprise_foundation/ef_organization_unit.graphql) · [Resolver](../../../backend/gateway/graph/resolvers/ef_organization_unit.resolvers.go) · [Adapter](../../../backend/gateway/internal/adapter/enterprise_foundation/organization_unit_adapter.go) · [OrganizationUnitService](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/organization_unit.proto).

| Operasi GraphQL | Butuh GraphQL | RPC | Permission | Scope di adapter | Sensitif |
| --- | --- | --- | --- | --- | --- |
| `organizationUnit` | Query | `GetOrganizationUnitById` | `organization_units.read` | Target | Tidak |
| `organizationUnitsByIds` | Query | `GetOrganizationUnitsByIds` | `organization_units.read` | Target | Tidak |
| `organizationUnitByEntityAndCode` | Query | `GetOrganizationUnitByEntityAndCode` | `organization_units.read` | Target | Tidak |
| `organizationUnitsByHeadPersonId` | Query | `ListOrganizationUnitsByHeadPersonId` | `organization_units.read` | Global | Tidak |
| `organizationUnitsByPrimaryLocationId` | Query | `ListOrganizationUnitsByPrimaryLocationId` | `organization_units.read` | Target | Tidak |
| `organizationUnitChildren` | Query | `ListOrganizationUnitChildren` | `organization_units.read` | Target | Tidak |
| `organizationUnits` | Query | `ListOrganizationUnits` | `organization_units.read` | Target / global | Tidak |
| `allOrganizationUnits` | Query | `GetAllOrganizationUnits` | `organization_units.read` | Global | Tidak |
| `validateOrganizationUnitScopeTarget` | Query | `ValidateOrganizationUnitScopeTarget` | `organization_units.read` | Target | Tidak |
| `createOrganizationUnit` | Mutation | `CreateOrganizationUnit` | `organization_units.write` | Target | Tidak |
| `updateOrganizationUnit` | Mutation | `UpdateOrganizationUnit` | `organization_units.write` | Target | Tidak |
| `assignOrganizationUnitHeadPerson` | Mutation | `AssignOrganizationUnitHeadPerson` | `organization_units.write` | Target | Tidak |
| `activateOrganizationUnit` | Mutation | `ActivateOrganizationUnit` | `organization_units.write` | Target | Tidak |
| `deactivateOrganizationUnit` | Mutation | `DeactivateOrganizationUnit` | `organization_units.write` | Target | Tidak |
| `deleteOrganizationUnit` | Mutation | `DeleteOrganizationUnit` | `organization_units.write` | Target | Ya |

## Mitra, peran, PIC dan alamat mitra

Pisahkan pendaftaran dari verifikasi, profil dari lifecycle, dan pencatatan kontak dari aktivasi mandat. Semua jalur provisioning dan evaluasi memakai role yang sama (EFL-01/02/03/04/08). [SDL](../../../backend/gateway/graphql/enterprise_foundation/ef_party.graphql) · [Resolver](../../../backend/gateway/graph/resolvers/ef_party.resolvers.go) · [Adapter](../../../backend/gateway/internal/adapter/enterprise_foundation/party_adapter.go) · [Adapter role/PIC/alamat](../../../backend/gateway/internal/adapter/enterprise_foundation/party_sub_adapter.go) · [PartyService](../../../backend/services/enterprise_foundation/contract/proto/enterprise_foundation/v1/party.proto).

| Operasi GraphQL | Butuh GraphQL | RPC | Permission | Scope di adapter | Sensitif |
| --- | --- | --- | --- | --- | --- |
| `party` | Query | `GetPartyById` | `parties.read` | Target | Tidak |
| `partiesByIds` | Query | `GetPartiesByIds` | `parties.read` | Target | Tidak |
| `partyByCode` | Query | `GetPartyByCode` | `parties.read` | Global | Tidak |
| `partyByTaxIdNumber` | Query | `GetPartyByTaxIdNumber` | `parties.read` | Global | Tidak |
| `partiesByEmail` | Query | `ListPartiesByEmail` | `parties.read` | Global | Tidak |
| `parties` | Query | `ListParties` | `parties.read` | Global | Tidak |
| `allParties` | Query | `GetAllParties` | `parties.read` | Global | Tidak |
| `partyRole` | Query | `GetPartyRoleById` | `parties.read` | Global | Tidak |
| `partyRolesByParty` | Query | `ListPartyRolesByParty` | `parties.read` | Target | Tidak |
| `partyRoleByPartyAndKind` | Query | `GetPartyRoleByPartyAndKind` | `parties.read` | Target | Tidak |
| `partyRoles` | Query | `ListPartyRoles` | `parties.read` | Target / global | Tidak |
| `allPartyRoles` | Query | `GetAllPartyRoles` | `parties.read` | Global | Tidak |
| `partyRolesByKind` | Query | `ListPartyRolesByKind` | `parties.read` | Global | Tidak |
| `allPartyRolesByKind` | Query | `GetAllPartyRolesByKind` | `parties.read` | Global | Tidak |
| `validatePartyRoleScope` | Query | `ValidatePartyRoleScope` | `parties.read` | Target | Tidak |
| `partyContact` | Query | `GetPartyContactById` | `parties.read` | Global | Tidak |
| `partyContactsByParty` | Query | `ListPartyContactsByParty` | `parties.read` | Target | Tidak |
| `primaryPartyContact` | Query | `GetPrimaryPartyContact` | `parties.read` | Target | Tidak |
| `partyContactsByIamUserId` | Query | `ListPartyContactsByIamUserId` | `parties.read` | Global | Tidak |
| `partyContactsByEmail` | Query | `ListPartyContactsByEmail` | `parties.read` | Global | Tidak |
| `partyContacts` | Query | `ListPartyContacts` | `parties.read` | Target / global | Tidak |
| `allPartyContacts` | Query | `GetAllPartyContacts` | `parties.read` | Global | Tidak |
| `evaluateRepresentativePortalAccess` | Query | `EvaluateRepresentativePortalAccess` | `parties.read` | Target | Tidak |
| `partyAddressLink` | Query | `GetPartyAddressLinkById` | `parties.read` | Global | Tidak |
| `partyAddressLinksByParty` | Query | `ListPartyAddressLinksByParty` | `parties.read` | Target | Tidak |
| `partyAddressLinksByPartyAndUsage` | Query | `ListPartyAddressLinksByPartyAndUsage` | `parties.read` | Target | Tidak |
| `primaryPartyAddressLink` | Query | `GetPrimaryPartyAddressLink` | `parties.read` | Target | Tidak |
| `partyAddressLinksByAddressId` | Query | `ListPartyAddressLinksByAddressId` | `parties.read` | Global | Tidak |
| `partyAddressLinks` | Query | `ListPartyAddressLinks` | `parties.read` | Target / global | Tidak |
| `allPartyAddressLinks` | Query | `GetAllPartyAddressLinks` | `parties.read` | Global | Tidak |
| `createParty` | Mutation | `CreateParty` | `parties.write` | Global | Tidak |
| `updatePartyProfile` | Mutation | `UpdatePartyProfile` | `parties.write` | Target | Tidak |
| `updatePartyTaxProfile` | Mutation | `UpdatePartyTaxProfile` | `parties.write` | Target | Tidak |
| `verifyAndActivateParty` | Mutation | `VerifyAndActivateParty` | `parties.write` | Target | Tidak |
| `suspendParty` | Mutation | `SuspendParty` | `parties.write` | Target | Tidak |
| `deactivateParty` | Mutation | `DeactivateParty` | `parties.write` | Target | Tidak |
| `deleteParty` | Mutation | `DeleteParty` | `parties.write` | Target | Ya |
| `assignPartyRole` | Mutation | `AssignPartyRole` | `parties.write` | Target | Tidak |
| `updatePartyRoleNote` | Mutation | `UpdatePartyRoleNote` | `parties.write` | Global | Tidak |
| `activatePartyRole` | Mutation | `ActivatePartyRole` | `parties.write` | Global | Tidak |
| `deactivatePartyRole` | Mutation | `DeactivatePartyRole` | `parties.write` | Global | Tidak |
| `removePartyRole` | Mutation | `RemovePartyRole` | `parties.write` | Global | Ya |
| `addPartyContact` | Mutation | `AddPartyContact` | `parties.write` | Target | Tidak |
| `updatePartyContact` | Mutation | `UpdatePartyContact` | `parties.write` | Global | Tidak |
| `setPrimaryPartyContact` | Mutation | `SetPrimaryPartyContact` | `parties.write` | Target | Tidak |
| `requestPartyPortalLink` | Mutation | `RequestPartyPortalLink` | `parties.write` | Global | Tidak |
| `activatePartyPortalAccess` | Mutation | `ActivatePartyPortalAccess` | `parties.write` | Global | Ya |
| `revokePartyPortalAccess` | Mutation | `RevokePartyPortalAccess` | `parties.write` | Global | Ya |
| `unlinkPartyPortalUser` | Mutation | `UnlinkPartyPortalUser` | `parties.write` | Global | Ya |
| `activatePartyContact` | Mutation | `ActivatePartyContact` | `parties.write` | Global | Tidak |
| `deactivatePartyContact` | Mutation | `DeactivatePartyContact` | `parties.write` | Global | Tidak |
| `deletePartyContact` | Mutation | `DeletePartyContact` | `parties.write` | Global | Ya |
| `linkPartyAddress` | Mutation | `LinkPartyAddress` | `parties.write` | Target | Tidak |
| `updatePartyAddressLink` | Mutation | `UpdatePartyAddressLink` | `parties.write` | Global | Tidak |
| `setPrimaryPartyAddressLink` | Mutation | `SetPrimaryPartyAddressLink` | `parties.write` | Global | Tidak |
| `unlinkPartyAddress` | Mutation | `UnlinkPartyAddress` | `parties.write` | Global | Tidak |

## Kontrak yang terdampak dan usulan privat

- **EF ↔ IAM/Procurement.** Perluas kontrak aktivasi, pencabutan dan evaluasi agar menunjuk mandat party–contact–principal–role yang tepat. Semantik per-role harus eksplisit untuk pemanggil baru dan permintaan lama yang ambigu. `evaluateRepresentativePortalAccess` sudah tersedia. Konsumen workload memakai gRPC dengan allowlist. Bukti EF dipakai bersama keputusan IAM dan kualifikasi Procurement.
- **EF → HRM.** Gunakan verifier privat untuk eligibility kepala unit. Nama usulan `VerifyUnitHeadEligibility` berlaku jika kontrak yang ada belum mengembalikan fakta yang dibutuhkan. Tidak ada query/mutation publik baru untuk validasi ini. EF tetap melayani tiga operasi penulisan kepala di katalog.
- **EF → IAM via outbox.** Ganti payload pelepasan agar mandat lama dapat diidentifikasi dan dideduplikasi. Pembuktian aplikasi hak di IAM memerlukan hasil konsumsi. Event internal tidak menambah field root GraphQL.
- **Gateway.** Sesuaikan SDL dan regenerasi untuk perubahan kontrak yang benar-benar diperlukan. Verifikasi publikasi, typed ID dan hasil melalui `/api/graphql`. Alur pendaftaran vendor/supplier tetap memakai IAM dan verifikasi domain yang sesuai. Seluruh usulan di bagian ini berada di luar hitungan baseline. [Audit kode](../../../agents/audit/enterprise_foundation/2-2026-11-28-233707-audit-logika-master-dan-mandat.md) · [Instruksi perbaikan](../../../agents/instructions/enterprise_foundation/4-leader-remediasi-logika-master-dan-mandat-instructions.md).
