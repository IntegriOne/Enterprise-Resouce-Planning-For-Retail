/**
 * Seluruh 22 Mutasi GraphQL Enterprise Foundation untuk domain Struktur Internal (`/human-resource`):
 * Address (4M), LegalEntity (7M), Location (5M), OrganizationUnit (6M).
 */

import {
  executeGraphQL,
  unwrapMutationPayload,
  type MutationResult,
} from "./iam-browser-auth.ts";
import {
  ADDRESS_FIELDS,
  LEGAL_ENTITY_FIELDS,
  LOCATION_FIELDS,
  ORG_UNIT_FIELDS,
} from "./ef-structure-queries.ts";
import type {
  AddressFormInput,
  CreateLegalEntityFormInput,
  CreateLocationFormInput,
  CreateOrganizationUnitFormInput,
  EfAddress,
  EfLegalEntity,
  EfLocation,
  EfOrganizationUnit,
  UpdateLegalEntityFormInput,
  UpdateLocationFormInput,
  UpdateOrganizationUnitFormInput,
} from "./ef-types.ts";

const RESULT_FIELDS = "errors { type message } was_successful";

export const EF_STRUCTURE_MUTATION_DOCS = {
  // Address (4)
  createAddress: `mutation EfCreateAddress($input: CreateAddressInput!) { createAddress(input: $input) { address { ${ADDRESS_FIELDS} } ${RESULT_FIELDS} } }`,
  findOrCreateAddress: `mutation EfFindOrCreateAddress($input: FindOrCreateAddressInput!) { findOrCreateAddress(input: $input) { address { ${ADDRESS_FIELDS} } created ${RESULT_FIELDS} } }`,
  updateAddress: `mutation EfUpdateAddress($input: UpdateAddressInput!) { updateAddress(input: $input) { address { ${ADDRESS_FIELDS} } ${RESULT_FIELDS} } }`,
  deleteAddress: `mutation EfDeleteAddress($id: ID!) { deleteAddress(id: $id) { deleted ${RESULT_FIELDS} } }`,

  // LegalEntity (7)
  createLegalEntity: `mutation EfCreateLegalEntity($input: CreateLegalEntityInput!) { createLegalEntity(input: $input) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  updateLegalEntity: `mutation EfUpdateLegalEntity($input: UpdateLegalEntityInput!) { updateLegalEntity(input: $input) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  setPrimaryOperatingLegalEntity: `mutation EfSetPrimaryLegalEntity($id: ID!) { setPrimaryOperatingLegalEntity(id: $id) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  activateLegalEntity: `mutation EfActivateLegalEntity($id: ID!) { activateLegalEntity(id: $id) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  suspendLegalEntity: `mutation EfSuspendLegalEntity($id: ID!) { suspendLegalEntity(id: $id) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  dissolveLegalEntity: `mutation EfDissolveLegalEntity($id: ID!, $dissolution_date: Time) { dissolveLegalEntity(id: $id, dissolution_date: $dissolution_date) { legal_entity { ${LEGAL_ENTITY_FIELDS} } ${RESULT_FIELDS} } }`,
  deleteLegalEntity: `mutation EfDeleteLegalEntity($id: ID!) { deleteLegalEntity(id: $id) { deleted ${RESULT_FIELDS} } }`,

  // Location (5)
  createLocation: `mutation EfCreateLocation($input: CreateLocationInput!) { createLocation(input: $input) { location { ${LOCATION_FIELDS} } ${RESULT_FIELDS} } }`,
  updateLocation: `mutation EfUpdateLocation($input: UpdateLocationInput!) { updateLocation(input: $input) { location { ${LOCATION_FIELDS} } ${RESULT_FIELDS} } }`,
  activateLocation: `mutation EfActivateLocation($id: ID!) { activateLocation(id: $id) { location { ${LOCATION_FIELDS} } ${RESULT_FIELDS} } }`,
  deactivateLocation: `mutation EfDeactivateLocation($id: ID!) { deactivateLocation(id: $id) { location { ${LOCATION_FIELDS} } ${RESULT_FIELDS} } }`,
  deleteLocation: `mutation EfDeleteLocation($id: ID!) { deleteLocation(id: $id) { deleted ${RESULT_FIELDS} } }`,

  // OrganizationUnit (6)
  createOrganizationUnit: `mutation EfCreateOrgUnit($input: CreateOrganizationUnitInput!) { createOrganizationUnit(input: $input) { organization_unit { ${ORG_UNIT_FIELDS} } ${RESULT_FIELDS} } }`,
  updateOrganizationUnit: `mutation EfUpdateOrgUnit($input: UpdateOrganizationUnitInput!) { updateOrganizationUnit(input: $input) { organization_unit { ${ORG_UNIT_FIELDS} } ${RESULT_FIELDS} } }`,
  assignOrganizationUnitHeadPerson: `mutation EfAssignOrgUnitHead($id: ID!, $head_person_id: ID) { assignOrganizationUnitHeadPerson(id: $id, head_person_id: $head_person_id) { organization_unit { ${ORG_UNIT_FIELDS} } ${RESULT_FIELDS} } }`,
  activateOrganizationUnit: `mutation EfActivateOrgUnit($id: ID!) { activateOrganizationUnit(id: $id) { organization_unit { ${ORG_UNIT_FIELDS} } ${RESULT_FIELDS} } }`,
  deactivateOrganizationUnit: `mutation EfDeactivateOrgUnit($id: ID!) { deactivateOrganizationUnit(id: $id) { organization_unit { ${ORG_UNIT_FIELDS} } ${RESULT_FIELDS} } }`,
  deleteOrganizationUnit: `mutation EfDeleteOrgUnit($id: ID!) { deleteOrganizationUnit(id: $id) { deleted ${RESULT_FIELDS} } }`,
};

async function runEfMutation<TPayload extends Partial<MutationResult>>(
  doc: string,
  variables: Record<string, unknown>,
  field: string,
  csrfToken?: string | null,
  fallback = "Mutasi Enterprise Foundation gagal diproses",
): Promise<TPayload> {
  const { data } = await executeGraphQL<Record<string, TPayload | null | undefined>>(
    doc,
    variables,
    csrfToken,
  );
  return unwrapMutationPayload(data?.[field], 200, fallback);
}

export async function createAddress(input: AddressFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { address?: EfAddress | null }>(
    EF_STRUCTURE_MUTATION_DOCS.createAddress,
    { input },
    "createAddress",
    csrfToken,
  );
  return payload.address ?? null;
}

export async function findOrCreateAddress(input: AddressFormInput, csrfToken?: string | null) {
  return runEfMutation<Partial<MutationResult> & { address?: EfAddress | null; created?: boolean | null }>(
    EF_STRUCTURE_MUTATION_DOCS.findOrCreateAddress,
    { input },
    "findOrCreateAddress",
    csrfToken,
  );
}

export async function updateAddress(input: AddressFormInput & { id: string }, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { address?: EfAddress | null }>(
    EF_STRUCTURE_MUTATION_DOCS.updateAddress,
    { input },
    "updateAddress",
    csrfToken,
  );
  return payload.address ?? null;
}

export async function deleteAddress(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deleteAddress,
    { id },
    "deleteAddress",
    csrfToken,
  );
  return Boolean(payload.deleted);
}

export async function createLegalEntity(input: CreateLegalEntityFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.createLegalEntity,
    { input },
    "createLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function updateLegalEntity(input: UpdateLegalEntityFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.updateLegalEntity,
    { input },
    "updateLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function setPrimaryOperatingLegalEntity(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.setPrimaryOperatingLegalEntity,
    { id },
    "setPrimaryOperatingLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function activateLegalEntity(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.activateLegalEntity,
    { id },
    "activateLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function suspendLegalEntity(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.suspendLegalEntity,
    { id },
    "suspendLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function dissolveLegalEntity(id: string, dissolutionDate?: string | null, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { legal_entity?: EfLegalEntity | null }>(
    EF_STRUCTURE_MUTATION_DOCS.dissolveLegalEntity,
    { id, dissolution_date: dissolutionDate || null },
    "dissolveLegalEntity",
    csrfToken,
  );
  return payload.legal_entity ?? null;
}

export async function deleteLegalEntity(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deleteLegalEntity,
    { id },
    "deleteLegalEntity",
    csrfToken,
  );
  return Boolean(payload.deleted);
}

export async function createLocation(input: CreateLocationFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { location?: EfLocation | null }>(
    EF_STRUCTURE_MUTATION_DOCS.createLocation,
    { input },
    "createLocation",
    csrfToken,
  );
  return payload.location ?? null;
}

export async function updateLocation(input: UpdateLocationFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { location?: EfLocation | null }>(
    EF_STRUCTURE_MUTATION_DOCS.updateLocation,
    { input },
    "updateLocation",
    csrfToken,
  );
  return payload.location ?? null;
}

export async function activateLocation(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { location?: EfLocation | null }>(
    EF_STRUCTURE_MUTATION_DOCS.activateLocation,
    { id },
    "activateLocation",
    csrfToken,
  );
  return payload.location ?? null;
}

export async function deactivateLocation(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { location?: EfLocation | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deactivateLocation,
    { id },
    "deactivateLocation",
    csrfToken,
  );
  return payload.location ?? null;
}

export async function deleteLocation(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deleteLocation,
    { id },
    "deleteLocation",
    csrfToken,
  );
  return Boolean(payload.deleted);
}

export async function createOrganizationUnit(input: CreateOrganizationUnitFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { organization_unit?: EfOrganizationUnit | null }>(
    EF_STRUCTURE_MUTATION_DOCS.createOrganizationUnit,
    { input },
    "createOrganizationUnit",
    csrfToken,
  );
  return payload.organization_unit ?? null;
}

export async function updateOrganizationUnit(input: UpdateOrganizationUnitFormInput, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { organization_unit?: EfOrganizationUnit | null }>(
    EF_STRUCTURE_MUTATION_DOCS.updateOrganizationUnit,
    { input },
    "updateOrganizationUnit",
    csrfToken,
  );
  return payload.organization_unit ?? null;
}

export async function assignOrganizationUnitHeadPerson(id: string, headPersonId?: string | null, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { organization_unit?: EfOrganizationUnit | null }>(
    EF_STRUCTURE_MUTATION_DOCS.assignOrganizationUnitHeadPerson,
    { id, head_person_id: headPersonId || null },
    "assignOrganizationUnitHeadPerson",
    csrfToken,
  );
  return payload.organization_unit ?? null;
}

export async function activateOrganizationUnit(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { organization_unit?: EfOrganizationUnit | null }>(
    EF_STRUCTURE_MUTATION_DOCS.activateOrganizationUnit,
    { id },
    "activateOrganizationUnit",
    csrfToken,
  );
  return payload.organization_unit ?? null;
}

export async function deactivateOrganizationUnit(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { organization_unit?: EfOrganizationUnit | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deactivateOrganizationUnit,
    { id },
    "deactivateOrganizationUnit",
    csrfToken,
  );
  return payload.organization_unit ?? null;
}

export async function deleteOrganizationUnit(id: string, csrfToken?: string | null) {
  const payload = await runEfMutation<Partial<MutationResult> & { deleted?: boolean | null }>(
    EF_STRUCTURE_MUTATION_DOCS.deleteOrganizationUnit,
    { id },
    "deleteOrganizationUnit",
    csrfToken,
  );
  return Boolean(payload.deleted);
}
