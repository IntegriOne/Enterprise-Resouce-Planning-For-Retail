/**
 * Seluruh 34 Query GraphQL Enterprise Foundation untuk domain Struktur Internal (`/human-resource`):
 * Address (5Q), LegalEntity (10Q), Location (10Q), OrganizationUnit (9Q), plus Relay `node(id)`.
 */

import { executeGraphQL } from "./iam-browser-auth.ts";
import { nullableOnNotFound } from "./ef-helpers.ts";
import type {
  AddressConnection,
  EfAddress,
  EfLegalEntity,
  EfLocation,
  EfOrganizationUnit,
  LegalEntityConnection,
  LegalEntityStatus,
  LocationConnection,
  LocationKind,
  OrgUnitKind,
  OrganizationUnitConnection,
  ScopeTargetDecision,
} from "./ef-types.ts";

const PAGE_INFO = "endCursor hasNextPage hasPreviousPage startCursor";

export const ADDRESS_FIELDS = `
  id _id label address_line_1 address_line_2 city province postal_code
  country_code formatted_address latitude longitude created_at updated_at
`;

export const LEGAL_ENTITY_FIELDS = `
  id _id parent_legal_entity_id entity_code legal_name trading_name legal_form
  is_primary_operating_entity status jurisdiction_country_code nib tax_id_number
  is_pkp registered_address_id fiscal_year_start_month incorporation_date
  dissolution_date created_at updated_at deleted_at
`;

export const LOCATION_FIELDS = `
  id _id legal_entity_id parent_location_id address_id location_code name kind
  phone timezone operational_attributes_json is_active created_at updated_at deleted_at
`;

export const ORG_UNIT_FIELDS = `
  id _id legal_entity_id parent_unit_id primary_location_id unit_code name kind
  cost_center_code head_person_id is_active created_at updated_at deleted_at
`;

const DECISION_FIELDS = "eligible scope_kind target_id code legal_entity_id";

export const EF_STRUCTURE_QUERY_DOCS = {
  // Address (5)
  address: `query EfAddress($id: ID!) { address(id: $id) { ${ADDRESS_FIELDS} } }`,
  addressesByIds: `query EfAddressesByIds($ids: [ID!]!) { addressesByIds(ids: $ids) { ${ADDRESS_FIELDS} } }`,
  addresses: `query EfAddresses($after: String, $before: String, $first: Int, $last: Int, $query: String, $country_code: String, $city: String, $province: String, $postal_code: String) {
    addresses(after: $after, before: $before, first: $first, last: $last, query: $query, country_code: $country_code, city: $city, province: $province, postal_code: $postal_code) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ADDRESS_FIELDS} }
    }
  }`,
  allAddresses: `query EfAllAddresses($after: String, $before: String, $first: Int, $last: Int) {
    allAddresses(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ADDRESS_FIELDS} }
    }
  }`,
  isAddressReferenced: `query EfIsAddressReferenced($id: ID!) { isAddressReferenced(id: $id) }`,

  // LegalEntity (10)
  legalEntity: `query EfLegalEntity($id: ID!) {
    legalEntity(id: $id) {
      ${LEGAL_ENTITY_FIELDS}
      parent_legal_entity { ${LEGAL_ENTITY_FIELDS} }
      registered_address { ${ADDRESS_FIELDS} }
      children(first: 10) { total_count pageInfo { ${PAGE_INFO} } nodes { ${LEGAL_ENTITY_FIELDS} } }
    }
  }`,
  legalEntitiesByIds: `query EfLegalEntitiesByIds($ids: [ID!]!) { legalEntitiesByIds(ids: $ids) { ${LEGAL_ENTITY_FIELDS} } }`,
  legalEntityByCode: `query EfLegalEntityByCode($entity_code: String!) { legalEntityByCode(entity_code: $entity_code) { ${LEGAL_ENTITY_FIELDS} } }`,
  legalEntityByTaxIdNumber: `query EfLegalEntityByTaxId($tax_id_number: String!) { legalEntityByTaxIdNumber(tax_id_number: $tax_id_number) { ${LEGAL_ENTITY_FIELDS} } }`,
  legalEntitiesByNib: `query EfLegalEntitiesByNib($nib: String!, $first: Int, $after: String) {
    legalEntitiesByNib(nib: $nib, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LEGAL_ENTITY_FIELDS} }
    }
  }`,
  primaryOperatingLegalEntity: `query EfPrimaryLegalEntity { primaryOperatingLegalEntity { ${LEGAL_ENTITY_FIELDS} } }`,
  legalEntities: `query EfLegalEntities($after: String, $before: String, $first: Int, $last: Int, $query: String, $parent_legal_entity_id: ID, $status: LegalEntityStatus, $primary_only: Boolean, $include_deleted: Boolean) {
    legalEntities(after: $after, before: $before, first: $first, last: $last, query: $query, parent_legal_entity_id: $parent_legal_entity_id, status: $status, primary_only: $primary_only, include_deleted: $include_deleted) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LEGAL_ENTITY_FIELDS} }
    }
  }`,
  allLegalEntities: `query EfAllLegalEntities($after: String, $before: String, $first: Int, $last: Int) {
    allLegalEntities(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LEGAL_ENTITY_FIELDS} }
    }
  }`,
  legalEntityChildren: `query EfLegalEntityChildren($parent_legal_entity_id: ID!, $first: Int, $after: String) {
    legalEntityChildren(parent_legal_entity_id: $parent_legal_entity_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LEGAL_ENTITY_FIELDS} }
    }
  }`,
  validateLegalEntityScopeTarget: `query EfValidateLegalEntityScope($id: ID!) {
    validateLegalEntityScopeTarget(id: $id) { ${DECISION_FIELDS} }
  }`,

  // Location (10)
  location: `query EfLocation($id: ID!) {
    location(id: $id) {
      ${LOCATION_FIELDS}
      legal_entity { id _id entity_code legal_name status }
      parent_location { id _id location_code name kind }
      address { ${ADDRESS_FIELDS} }
      children(first: 10) { total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} } }
    }
  }`,
  locationsByIds: `query EfLocationsByIds($ids: [ID!]!) { locationsByIds(ids: $ids) { ${LOCATION_FIELDS} } }`,
  locationByCode: `query EfLocationByCode($location_code: String!) { locationByCode(location_code: $location_code) { ${LOCATION_FIELDS} } }`,
  locationsByLegalEntityAndKind: `query EfLocationsByEntityAndKind($legal_entity_id: ID!, $kind: LocationKind!, $active_only: Boolean, $first: Int, $after: String) {
    locationsByLegalEntityAndKind(legal_entity_id: $legal_entity_id, kind: $kind, active_only: $active_only, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} }
    }
  }`,
  locationsByAddressId: `query EfLocationsByAddressId($address_id: ID!, $first: Int, $after: String) {
    locationsByAddressId(address_id: $address_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} }
    }
  }`,
  locationChildren: `query EfLocationChildren($parent_location_id: ID!, $first: Int, $after: String) {
    locationChildren(parent_location_id: $parent_location_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} }
    }
  }`,
  locations: `query EfLocations($after: String, $before: String, $first: Int, $last: Int, $query: String, $legal_entity_id: ID, $parent_location_id: ID, $address_id: ID, $kind: LocationKind, $active_only: Boolean, $include_deleted: Boolean) {
    locations(after: $after, before: $before, first: $first, last: $last, query: $query, legal_entity_id: $legal_entity_id, parent_location_id: $parent_location_id, address_id: $address_id, kind: $kind, active_only: $active_only, include_deleted: $include_deleted) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} }
    }
  }`,
  allLocations: `query EfAllLocations($after: String, $before: String, $first: Int, $last: Int) {
    allLocations(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${LOCATION_FIELDS} }
    }
  }`,
  validateFacilityScope: `query EfValidateFacilityScope($id: ID!) { validateFacilityScope(id: $id) { ${DECISION_FIELDS} } }`,
  validateWarehouseScope: `query EfValidateWarehouseScope($id: ID!) { validateWarehouseScope(id: $id) { ${DECISION_FIELDS} } }`,

  // OrganizationUnit (9)
  organizationUnit: `query EfOrgUnit($id: ID!) {
    organizationUnit(id: $id) {
      ${ORG_UNIT_FIELDS}
      legal_entity { id _id entity_code legal_name status }
      parent_unit { id _id unit_code name kind }
      primary_location { id _id location_code name kind }
      children(first: 10) { total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} } }
    }
  }`,
  organizationUnitsByIds: `query EfOrgUnitsByIds($ids: [ID!]!) { organizationUnitsByIds(ids: $ids) { ${ORG_UNIT_FIELDS} } }`,
  organizationUnitByEntityAndCode: `query EfOrgUnitByEntityAndCode($legal_entity_id: ID!, $unit_code: String!) {
    organizationUnitByEntityAndCode(legal_entity_id: $legal_entity_id, unit_code: $unit_code) { ${ORG_UNIT_FIELDS} }
  }`,
  organizationUnitsByHeadPersonId: `query EfOrgUnitsByHead($head_person_id: ID!, $first: Int, $after: String) {
    organizationUnitsByHeadPersonId(head_person_id: $head_person_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} }
    }
  }`,
  organizationUnitsByPrimaryLocationId: `query EfOrgUnitsByLocation($primary_location_id: ID!, $first: Int, $after: String) {
    organizationUnitsByPrimaryLocationId(primary_location_id: $primary_location_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} }
    }
  }`,
  organizationUnitChildren: `query EfOrgUnitChildren($parent_unit_id: ID!, $first: Int, $after: String) {
    organizationUnitChildren(parent_unit_id: $parent_unit_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} }
    }
  }`,
  organizationUnits: `query EfOrgUnits($after: String, $before: String, $first: Int, $last: Int, $query: String, $legal_entity_id: ID, $parent_unit_id: ID, $primary_location_id: ID, $kind: OrgUnitKind, $active_only: Boolean, $include_deleted: Boolean) {
    organizationUnits(after: $after, before: $before, first: $first, last: $last, query: $query, legal_entity_id: $legal_entity_id, parent_unit_id: $parent_unit_id, primary_location_id: $primary_location_id, kind: $kind, active_only: $active_only, include_deleted: $include_deleted) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} }
    }
  }`,
  allOrganizationUnits: `query EfAllOrgUnits($after: String, $before: String, $first: Int, $last: Int) {
    allOrganizationUnits(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${ORG_UNIT_FIELDS} }
    }
  }`,
  validateOrganizationUnitScopeTarget: `query EfValidateOrgUnitScope($id: ID!) {
    validateOrganizationUnitScopeTarget(id: $id) { ${DECISION_FIELDS} }
  }`,

  // Relay Node Inspector
  node: `query EfStructureNode($id: ID!) {
    node(id: $id) {
      id
      ... on Address { label address_line_1 city country_code }
      ... on LegalEntity { entity_code legal_name status is_primary_operating_entity }
      ... on Location { location_code name location_kind: kind is_active }
      ... on OrganizationUnit { unit_code name org_unit_kind: kind is_active }
    }
  }`,
};

export async function fetchAddress(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ address: EfAddress | null }>(EF_STRUCTURE_QUERY_DOCS.address, { id });
    return data.address;
  });
}
export async function fetchAddressesByIds(ids: string[]) {
  const { data } = await executeGraphQL<{ addressesByIds: EfAddress[] }>(EF_STRUCTURE_QUERY_DOCS.addressesByIds, { ids });
  return data.addressesByIds ?? [];
}
export async function fetchAddresses(vars: { query?: string; country_code?: string; city?: string; province?: string; postal_code?: string; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ addresses: AddressConnection }>(EF_STRUCTURE_QUERY_DOCS.addresses, { first: 20, ...vars });
  return data.addresses;
}
export async function fetchAllAddresses(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allAddresses: AddressConnection }>(EF_STRUCTURE_QUERY_DOCS.allAddresses, { first, after: after || null });
  return data.allAddresses;
}
export async function fetchIsAddressReferenced(id: string) {
  const { data } = await executeGraphQL<{ isAddressReferenced: boolean }>(EF_STRUCTURE_QUERY_DOCS.isAddressReferenced, { id });
  return Boolean(data.isAddressReferenced);
}

export async function fetchLegalEntity(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ legalEntity: EfLegalEntity | null }>(EF_STRUCTURE_QUERY_DOCS.legalEntity, { id });
    return data.legalEntity;
  });
}
export async function fetchLegalEntitiesByIds(ids: string[]) {
  const { data } = await executeGraphQL<{ legalEntitiesByIds: EfLegalEntity[] }>(EF_STRUCTURE_QUERY_DOCS.legalEntitiesByIds, { ids });
  return data.legalEntitiesByIds ?? [];
}
export async function fetchLegalEntityByCode(entityCode: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ legalEntityByCode: EfLegalEntity | null }>(EF_STRUCTURE_QUERY_DOCS.legalEntityByCode, { entity_code: entityCode });
    return data.legalEntityByCode;
  });
}
export async function fetchLegalEntityByTaxId(taxIdNumber: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ legalEntityByTaxIdNumber: EfLegalEntity | null }>(EF_STRUCTURE_QUERY_DOCS.legalEntityByTaxIdNumber, { tax_id_number: taxIdNumber });
    return data.legalEntityByTaxIdNumber;
  });
}
export async function fetchLegalEntitiesByNib(nib: string, first = 15, after?: string | null) {
  const { data } = await executeGraphQL<{ legalEntitiesByNib: LegalEntityConnection }>(EF_STRUCTURE_QUERY_DOCS.legalEntitiesByNib, { nib, first, after: after || null });
  return data.legalEntitiesByNib;
}
export async function fetchPrimaryOperatingLegalEntity() {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ primaryOperatingLegalEntity: EfLegalEntity | null }>(EF_STRUCTURE_QUERY_DOCS.primaryOperatingLegalEntity, {});
    return data.primaryOperatingLegalEntity;
  });
}
export async function fetchLegalEntities(vars: { query?: string; parent_legal_entity_id?: string | null; status?: LegalEntityStatus | null; primary_only?: boolean | null; include_deleted?: boolean; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ legalEntities: LegalEntityConnection }>(EF_STRUCTURE_QUERY_DOCS.legalEntities, { first: 20, include_deleted: false, ...vars });
  return data.legalEntities;
}
export async function fetchAllLegalEntities(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allLegalEntities: LegalEntityConnection }>(EF_STRUCTURE_QUERY_DOCS.allLegalEntities, { first, after: after || null });
  return data.allLegalEntities;
}
export async function fetchLegalEntityChildren(parentId: string, first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ legalEntityChildren: LegalEntityConnection }>(EF_STRUCTURE_QUERY_DOCS.legalEntityChildren, { parent_legal_entity_id: parentId, first, after: after || null });
  return data.legalEntityChildren;
}
export async function validateLegalEntityScopeTarget(id: string) {
  const { data } = await executeGraphQL<{ validateLegalEntityScopeTarget: ScopeTargetDecision }>(EF_STRUCTURE_QUERY_DOCS.validateLegalEntityScopeTarget, { id });
  return data.validateLegalEntityScopeTarget;
}

export async function fetchLocation(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ location: EfLocation | null }>(EF_STRUCTURE_QUERY_DOCS.location, { id });
    return data.location;
  });
}
export async function fetchLocationsByIds(ids: string[]) {
  const { data } = await executeGraphQL<{ locationsByIds: EfLocation[] }>(EF_STRUCTURE_QUERY_DOCS.locationsByIds, { ids });
  return data.locationsByIds ?? [];
}
export async function fetchLocationByCode(locationCode: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ locationByCode: EfLocation | null }>(EF_STRUCTURE_QUERY_DOCS.locationByCode, { location_code: locationCode });
    return data.locationByCode;
  });
}
export async function fetchLocationsByLegalEntityAndKind(legalEntityId: string, kind: LocationKind, activeOnly = false, first = 20) {
  const { data } = await executeGraphQL<{ locationsByLegalEntityAndKind: LocationConnection }>(EF_STRUCTURE_QUERY_DOCS.locationsByLegalEntityAndKind, { legal_entity_id: legalEntityId, kind, active_only: activeOnly, first });
  return data.locationsByLegalEntityAndKind;
}
export async function fetchLocationsByAddressId(addressId: string, first = 20) {
  const { data } = await executeGraphQL<{ locationsByAddressId: LocationConnection }>(EF_STRUCTURE_QUERY_DOCS.locationsByAddressId, { address_id: addressId, first });
  return data.locationsByAddressId;
}
export async function fetchLocationChildren(parentLocationId: string, first = 20) {
  const { data } = await executeGraphQL<{ locationChildren: LocationConnection }>(EF_STRUCTURE_QUERY_DOCS.locationChildren, { parent_location_id: parentLocationId, first });
  return data.locationChildren;
}
export async function fetchLocations(vars: { query?: string; legal_entity_id?: string | null; parent_location_id?: string | null; address_id?: string | null; kind?: LocationKind | null; active_only?: boolean; include_deleted?: boolean; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ locations: LocationConnection }>(EF_STRUCTURE_QUERY_DOCS.locations, { first: 20, active_only: false, include_deleted: false, ...vars });
  return data.locations;
}
export async function fetchAllLocations(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allLocations: LocationConnection }>(EF_STRUCTURE_QUERY_DOCS.allLocations, { first, after: after || null });
  return data.allLocations;
}
export async function validateFacilityScope(id: string) {
  const { data } = await executeGraphQL<{ validateFacilityScope: ScopeTargetDecision }>(EF_STRUCTURE_QUERY_DOCS.validateFacilityScope, { id });
  return data.validateFacilityScope;
}
export async function validateWarehouseScope(id: string) {
  const { data } = await executeGraphQL<{ validateWarehouseScope: ScopeTargetDecision }>(EF_STRUCTURE_QUERY_DOCS.validateWarehouseScope, { id });
  return data.validateWarehouseScope;
}

export async function fetchOrganizationUnit(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ organizationUnit: EfOrganizationUnit | null }>(EF_STRUCTURE_QUERY_DOCS.organizationUnit, { id });
    return data.organizationUnit;
  });
}
export async function fetchOrganizationUnitsByIds(ids: string[]) {
  const { data } = await executeGraphQL<{ organizationUnitsByIds: EfOrganizationUnit[] }>(EF_STRUCTURE_QUERY_DOCS.organizationUnitsByIds, { ids });
  return data.organizationUnitsByIds ?? [];
}
export async function fetchOrganizationUnitByEntityAndCode(legalEntityId: string, unitCode: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ organizationUnitByEntityAndCode: EfOrganizationUnit | null }>(EF_STRUCTURE_QUERY_DOCS.organizationUnitByEntityAndCode, { legal_entity_id: legalEntityId, unit_code: unitCode });
    return data.organizationUnitByEntityAndCode;
  });
}
export async function fetchOrganizationUnitsByHeadPersonId(headPersonId: string, first = 20) {
  const { data } = await executeGraphQL<{ organizationUnitsByHeadPersonId: OrganizationUnitConnection }>(EF_STRUCTURE_QUERY_DOCS.organizationUnitsByHeadPersonId, { head_person_id: headPersonId, first });
  return data.organizationUnitsByHeadPersonId;
}
export async function fetchOrganizationUnitsByPrimaryLocationId(primaryLocationId: string, first = 20) {
  const { data } = await executeGraphQL<{ organizationUnitsByPrimaryLocationId: OrganizationUnitConnection }>(EF_STRUCTURE_QUERY_DOCS.organizationUnitsByPrimaryLocationId, { primary_location_id: primaryLocationId, first });
  return data.organizationUnitsByPrimaryLocationId;
}
export async function fetchOrganizationUnitChildren(parentUnitId: string, first = 20) {
  const { data } = await executeGraphQL<{ organizationUnitChildren: OrganizationUnitConnection }>(EF_STRUCTURE_QUERY_DOCS.organizationUnitChildren, { parent_unit_id: parentUnitId, first });
  return data.organizationUnitChildren;
}
export async function fetchOrganizationUnits(vars: { query?: string; legal_entity_id?: string | null; parent_unit_id?: string | null; primary_location_id?: string | null; kind?: OrgUnitKind | null; active_only?: boolean; include_deleted?: boolean; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ organizationUnits: OrganizationUnitConnection }>(EF_STRUCTURE_QUERY_DOCS.organizationUnits, { first: 20, active_only: false, include_deleted: false, ...vars });
  return data.organizationUnits;
}
export async function fetchAllOrganizationUnits(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allOrganizationUnits: OrganizationUnitConnection }>(EF_STRUCTURE_QUERY_DOCS.allOrganizationUnits, { first, after: after || null });
  return data.allOrganizationUnits;
}
export async function validateOrganizationUnitScopeTarget(id: string) {
  const { data } = await executeGraphQL<{ validateOrganizationUnitScopeTarget: ScopeTargetDecision }>(EF_STRUCTURE_QUERY_DOCS.validateOrganizationUnitScopeTarget, { id });
  return data.validateOrganizationUnitScopeTarget;
}

export async function fetchEfStructureNode(id: string) {
  const { data } = await executeGraphQL<{ node: Record<string, unknown> | null }>(EF_STRUCTURE_QUERY_DOCS.node, { id });
  return data.node;
}
