/**
 * Seluruh 26 Mutasi GraphQL Enterprise Foundation untuk domain Mitra & Procurement (`/procurement`):
 * Party (7M), PartyRole (5M), PartyContact & Portal (10M), PartyAddressLink (4M).
 */

import {
  executeGraphQL,
  unwrapMutationPayload,
  type MutationResult,
} from "./iam-browser-auth.ts";
import {
  PARTY_ADDRESS_FIELDS,
  PARTY_CONTACT_FIELDS,
  PARTY_FIELDS,
  PARTY_ROLE_FIELDS,
} from "./ef-partner-queries.ts";
import type {
  AddPartyContactFormInput,
  CreatePartyFormInput,
  EfParty,
  EfPartyAddress,
  EfPartyContact,
  EfPartyRole,
  LinkPartyAddressFormInput,
  PartyRoleKind,
  UpdatePartyAddressLinkFormInput,
  UpdatePartyContactFormInput,
  UpdatePartyProfileFormInput,
  UpdatePartyTaxProfileFormInput,
} from "./ef-types.ts";

const RESULT_FIELDS = "errors { type message } was_successful";

export const EF_PARTNER_MUTATION_DOCS = {
  // Party Core (7)
  createParty: `mutation EfCreateParty($input: CreatePartyInput!) { createParty(input: $input) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  updatePartyProfile: `mutation EfUpdatePartyProfile($input: UpdatePartyProfileInput!) { updatePartyProfile(input: $input) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  updatePartyTaxProfile: `mutation EfUpdatePartyTax($input: UpdatePartyTaxProfileInput!) { updatePartyTaxProfile(input: $input) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  verifyAndActivateParty: `mutation EfVerifyParty($id: ID!) { verifyAndActivateParty(id: $id) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  suspendParty: `mutation EfSuspendParty($id: ID!) { suspendParty(id: $id) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  deactivateParty: `mutation EfDeactivateParty($id: ID!) { deactivateParty(id: $id) { party { ${PARTY_FIELDS} } ${RESULT_FIELDS} } }`,
  deleteParty: `mutation EfDeleteParty($id: ID!) { deleteParty(id: $id) { deleted ${RESULT_FIELDS} } }`,

  // PartyRole (5)
  assignPartyRole: `mutation EfAssignPartyRole($input: AssignPartyRoleInput!) { assignPartyRole(input: $input) { role { ${PARTY_ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
  updatePartyRoleNote: `mutation EfUpdatePartyRoleNote($input: UpdatePartyRoleNoteInput!) { updatePartyRoleNote(input: $input) { role { ${PARTY_ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
  activatePartyRole: `mutation EfActivatePartyRole($role_id: ID!) { activatePartyRole(role_id: $role_id) { role { ${PARTY_ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
  deactivatePartyRole: `mutation EfDeactivatePartyRole($role_id: ID!) { deactivatePartyRole(role_id: $role_id) { role { ${PARTY_ROLE_FIELDS} } ${RESULT_FIELDS} } }`,
  removePartyRole: `mutation EfRemovePartyRole($role_id: ID!) { removePartyRole(role_id: $role_id) { removed ${RESULT_FIELDS} } }`,

  // PartyContact & Portal Access (10)
  addPartyContact: `mutation EfAddPartyContact($input: AddPartyContactInput!) { addPartyContact(input: $input) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  updatePartyContact: `mutation EfUpdatePartyContact($input: UpdatePartyContactInput!) { updatePartyContact(input: $input) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  setPrimaryPartyContact: `mutation EfSetPrimaryPartyContact($party_id: ID!, $contact_id: ID!) { setPrimaryPartyContact(party_id: $party_id, contact_id: $contact_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  requestPartyPortalLink: `mutation EfRequestPortalLink($contact_id: ID!, $iam_user_id: ID!) { requestPartyPortalLink(contact_id: $contact_id, iam_user_id: $iam_user_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  activatePartyPortalAccess: `mutation EfActivatePortalAccess($contact_id: ID!, $iam_user_id: ID!) { activatePartyPortalAccess(contact_id: $contact_id, iam_user_id: $iam_user_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  revokePartyPortalAccess: `mutation EfRevokePortalAccess($contact_id: ID!) { revokePartyPortalAccess(contact_id: $contact_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  unlinkPartyPortalUser: `mutation EfUnlinkPortalUser($contact_id: ID!) { unlinkPartyPortalUser(contact_id: $contact_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  activatePartyContact: `mutation EfActivatePartyContact($contact_id: ID!) { activatePartyContact(contact_id: $contact_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  deactivatePartyContact: `mutation EfDeactivatePartyContact($contact_id: ID!) { deactivatePartyContact(contact_id: $contact_id) { contact { ${PARTY_CONTACT_FIELDS} } ${RESULT_FIELDS} } }`,
  deletePartyContact: `mutation EfDeletePartyContact($contact_id: ID!) { deletePartyContact(contact_id: $contact_id) { deleted ${RESULT_FIELDS} } }`,

  // PartyAddressLink (4)
  linkPartyAddress: `mutation EfLinkPartyAddress($input: LinkPartyAddressInput!) { linkPartyAddress(input: $input) { link { ${PARTY_ADDRESS_FIELDS} } ${RESULT_FIELDS} } }`,
  updatePartyAddressLink: `mutation EfUpdatePartyAddressLink($input: UpdatePartyAddressLinkInput!) { updatePartyAddressLink(input: $input) { link { ${PARTY_ADDRESS_FIELDS} } ${RESULT_FIELDS} } }`,
  setPrimaryPartyAddressLink: `mutation EfSetPrimaryPartyAddress($link_id: ID!) { setPrimaryPartyAddressLink(link_id: $link_id) { link { ${PARTY_ADDRESS_FIELDS} } ${RESULT_FIELDS} } }`,
  unlinkPartyAddress: `mutation EfUnlinkPartyAddress($link_id: ID!) { unlinkPartyAddress(link_id: $link_id) { unlinked ${RESULT_FIELDS} } }`,
};

async function runPartnerMutation<TPayload extends Partial<MutationResult>>(
  doc: string,
  variables: Record<string, unknown>,
  field: string,
  csrfToken?: string | null,
  fallback = "Mutasi Mitra Enterprise Foundation gagal diproses",
): Promise<TPayload> {
  const { data } = await executeGraphQL<Record<string, TPayload | null | undefined>>(
    doc,
    variables,
    csrfToken,
  );
  return unwrapMutationPayload(data?.[field], 200, fallback);
}

export async function createParty(input: CreatePartyFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.createParty, { input }, "createParty", csrfToken);
  return p.party ?? null;
}
export async function updatePartyProfile(input: UpdatePartyProfileFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.updatePartyProfile, { input }, "updatePartyProfile", csrfToken);
  return p.party ?? null;
}
export async function updatePartyTaxProfile(input: UpdatePartyTaxProfileFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.updatePartyTaxProfile, { input }, "updatePartyTaxProfile", csrfToken);
  return p.party ?? null;
}
export async function verifyAndActivateParty(id: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.verifyAndActivateParty, { id }, "verifyAndActivateParty", csrfToken);
  return p.party ?? null;
}
export async function suspendParty(id: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.suspendParty, { id }, "suspendParty", csrfToken);
  return p.party ?? null;
}
export async function deactivateParty(id: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { party?: EfParty | null }>(EF_PARTNER_MUTATION_DOCS.deactivateParty, { id }, "deactivateParty", csrfToken);
  return p.party ?? null;
}
export async function deleteParty(id: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { deleted?: boolean | null }>(EF_PARTNER_MUTATION_DOCS.deleteParty, { id }, "deleteParty", csrfToken);
  return Boolean(p.deleted);
}

export async function assignPartyRole(partyId: string, role: PartyRoleKind, referenceNote?: string | null, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { role?: EfPartyRole | null }>(EF_PARTNER_MUTATION_DOCS.assignPartyRole, { input: { party_id: partyId, role, reference_note: referenceNote || null } }, "assignPartyRole", csrfToken);
  return p.role ?? null;
}
export async function updatePartyRoleNote(id: string, referenceNote?: string | null, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { role?: EfPartyRole | null }>(EF_PARTNER_MUTATION_DOCS.updatePartyRoleNote, { input: { id, reference_note: referenceNote || null } }, "updatePartyRoleNote", csrfToken);
  return p.role ?? null;
}
export async function activatePartyRole(roleId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { role?: EfPartyRole | null }>(EF_PARTNER_MUTATION_DOCS.activatePartyRole, { role_id: roleId }, "activatePartyRole", csrfToken);
  return p.role ?? null;
}
export async function deactivatePartyRole(roleId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { role?: EfPartyRole | null }>(EF_PARTNER_MUTATION_DOCS.deactivatePartyRole, { role_id: roleId }, "deactivatePartyRole", csrfToken);
  return p.role ?? null;
}
export async function removePartyRole(roleId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { removed?: boolean | null }>(EF_PARTNER_MUTATION_DOCS.removePartyRole, { role_id: roleId }, "removePartyRole", csrfToken);
  return Boolean(p.removed);
}

export async function addPartyContact(input: AddPartyContactFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.addPartyContact, { input }, "addPartyContact", csrfToken);
  return p.contact ?? null;
}
export async function updatePartyContact(input: UpdatePartyContactFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.updatePartyContact, { input }, "updatePartyContact", csrfToken);
  return p.contact ?? null;
}
export async function setPrimaryPartyContact(partyId: string, contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.setPrimaryPartyContact, { party_id: partyId, contact_id: contactId }, "setPrimaryPartyContact", csrfToken);
  return p.contact ?? null;
}
export async function requestPartyPortalLink(contactId: string, iamUserId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.requestPartyPortalLink, { contact_id: contactId, iam_user_id: iamUserId }, "requestPartyPortalLink", csrfToken);
  return p.contact ?? null;
}
export async function activatePartyPortalAccess(contactId: string, iamUserId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.activatePartyPortalAccess, { contact_id: contactId, iam_user_id: iamUserId }, "activatePartyPortalAccess", csrfToken);
  return p.contact ?? null;
}
export async function revokePartyPortalAccess(contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.revokePartyPortalAccess, { contact_id: contactId }, "revokePartyPortalAccess", csrfToken);
  return p.contact ?? null;
}
export async function unlinkPartyPortalUser(contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.unlinkPartyPortalUser, { contact_id: contactId }, "unlinkPartyPortalUser", csrfToken);
  return p.contact ?? null;
}
export async function activatePartyContact(contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.activatePartyContact, { contact_id: contactId }, "activatePartyContact", csrfToken);
  return p.contact ?? null;
}
export async function deactivatePartyContact(contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { contact?: EfPartyContact | null }>(EF_PARTNER_MUTATION_DOCS.deactivatePartyContact, { contact_id: contactId }, "deactivatePartyContact", csrfToken);
  return p.contact ?? null;
}
export async function deletePartyContact(contactId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { deleted?: boolean | null }>(EF_PARTNER_MUTATION_DOCS.deletePartyContact, { contact_id: contactId }, "deletePartyContact", csrfToken);
  return Boolean(p.deleted);
}

export async function linkPartyAddress(input: LinkPartyAddressFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { link?: EfPartyAddress | null }>(EF_PARTNER_MUTATION_DOCS.linkPartyAddress, { input }, "linkPartyAddress", csrfToken);
  return p.link ?? null;
}
export async function updatePartyAddressLink(input: UpdatePartyAddressLinkFormInput, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { link?: EfPartyAddress | null }>(EF_PARTNER_MUTATION_DOCS.updatePartyAddressLink, { input }, "updatePartyAddressLink", csrfToken);
  return p.link ?? null;
}
export async function setPrimaryPartyAddressLink(linkId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { link?: EfPartyAddress | null }>(EF_PARTNER_MUTATION_DOCS.setPrimaryPartyAddressLink, { link_id: linkId }, "setPrimaryPartyAddressLink", csrfToken);
  return p.link ?? null;
}
export async function unlinkPartyAddress(linkId: string, csrfToken?: string | null) {
  const p = await runPartnerMutation<Partial<MutationResult> & { unlinked?: boolean | null }>(EF_PARTNER_MUTATION_DOCS.unlinkPartyAddress, { link_id: linkId }, "unlinkPartyAddress", csrfToken);
  return Boolean(p.unlinked);
}
