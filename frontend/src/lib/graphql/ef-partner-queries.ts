/**
 * Seluruh 30 Query GraphQL Enterprise Foundation untuk domain Mitra & Procurement (`/procurement`):
 * Party (7Q), PartyRole (8Q), PartyContact & Portal (8Q), PartyAddressLink (7Q), plus Relay `node(id)`.
 */

import { executeGraphQL } from "./iam-browser-auth.ts";
import { nullableOnNotFound } from "./ef-helpers.ts";
import { ADDRESS_FIELDS } from "./ef-structure-queries.ts";
import type {
  AddressUsageKind,
  EfParty,
  EfPartyAddress,
  EfPartyContact,
  EfPartyRole,
  PartyAddressConnection,
  PartyConnection,
  PartyContactConnection,
  PartyKind,
  PartyRoleConnection,
  PartyRoleKind,
  PartyStatus,
  PortalAccessEvaluation,
  PortalAccessStatus,
  ScopeTargetDecision,
} from "./ef-types.ts";

const PAGE_INFO = "endCursor hasNextPage hasPreviousPage startCursor";

export const PARTY_FIELDS = `
  id _id partner_code kind display_name legal_name email phone website_url
  tax_id_number tax_registered_name is_pkp is_tax_exempt tax_exemption_reason
  tax_verified_at preferred_language status notes created_at updated_at deleted_at
`;

export const PARTY_ROLE_FIELDS = `
  id _id party_id role is_active reference_note created_at updated_at
`;

export const PARTY_CONTACT_FIELDS = `
  id _id party_id iam_user_id full_name job_title contact_role email phone
  portal_access_status is_primary is_active created_at updated_at deleted_at
`;

export const PARTY_ADDRESS_FIELDS = `
  id _id party_id address_id usage is_primary created_at updated_at
`;

export const EF_PARTNER_QUERY_DOCS = {
  // Party Core (7)
  party: `query EfParty($id: ID!) {
    party(id: $id) {
      ${PARTY_FIELDS}
      roles(first: 10) { total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} } }
      contacts(first: 15) { total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} } }
      addresses(first: 10) {
        total_count pageInfo { ${PAGE_INFO} }
        nodes { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
      }
    }
  }`,
  partiesByIds: `query EfPartiesByIds($ids: [ID!]!) { partiesByIds(ids: $ids) { ${PARTY_FIELDS} } }`,
  partyByCode: `query EfPartyByCode($partner_code: String!) { partyByCode(partner_code: $partner_code) { ${PARTY_FIELDS} } }`,
  partyByTaxIdNumber: `query EfPartyByTaxId($tax_id_number: String!) { partyByTaxIdNumber(tax_id_number: $tax_id_number) { ${PARTY_FIELDS} } }`,
  partiesByEmail: `query EfPartiesByEmail($email: String!, $first: Int, $after: String) {
    partiesByEmail(email: $email, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_FIELDS} }
    }
  }`,
  parties: `query EfParties($after: String, $before: String, $first: Int, $last: Int, $query: String, $kind: PartyKind, $status: PartyStatus, $role: PartyRoleKind, $include_deleted: Boolean) {
    parties(after: $after, before: $before, first: $first, last: $last, query: $query, kind: $kind, status: $status, role: $role, include_deleted: $include_deleted) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_FIELDS} }
    }
  }`,
  allParties: `query EfAllParties($after: String, $before: String, $first: Int, $last: Int) {
    allParties(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_FIELDS} }
    }
  }`,

  // PartyRole (8)
  partyRole: `query EfPartyRole($id: ID!) { partyRole(id: $id) { ${PARTY_ROLE_FIELDS} } }`,
  partyRolesByParty: `query EfPartyRolesByParty($party_id: ID!, $first: Int, $after: String) {
    partyRolesByParty(party_id: $party_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} }
    }
  }`,
  partyRoleByPartyAndKind: `query EfPartyRoleByPartyAndKind($party_id: ID!, $kind: PartyRoleKind!) {
    partyRoleByPartyAndKind(party_id: $party_id, kind: $kind) { ${PARTY_ROLE_FIELDS} }
  }`,
  partyRoles: `query EfPartyRoles($after: String, $before: String, $first: Int, $last: Int, $party_id: ID, $role: PartyRoleKind, $active_only: Boolean, $query: String) {
    partyRoles(after: $after, before: $before, first: $first, last: $last, party_id: $party_id, role: $role, active_only: $active_only, query: $query) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} }
    }
  }`,
  allPartyRoles: `query EfAllPartyRoles($after: String, $before: String, $first: Int, $last: Int) {
    allPartyRoles(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} }
    }
  }`,
  partyRolesByKind: `query EfPartyRolesByKind($kind: PartyRoleKind!, $active_only: Boolean, $first: Int, $after: String) {
    partyRolesByKind(kind: $kind, active_only: $active_only, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} }
    }
  }`,
  allPartyRolesByKind: `query EfAllPartyRolesByKind($kind: PartyRoleKind!, $active_only: Boolean, $first: Int, $after: String) {
    allPartyRolesByKind(kind: $kind, active_only: $active_only, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ROLE_FIELDS} }
    }
  }`,
  validatePartyRoleScope: `query EfValidatePartyRoleScope($party_id: ID!, $required_role: PartyRoleKind!) {
    validatePartyRoleScope(party_id: $party_id, required_role: $required_role) {
      eligible scope_kind target_id code legal_entity_id
    }
  }`,

  // PartyContact (8)
  partyContact: `query EfPartyContact($id: ID!) { partyContact(id: $id) { ${PARTY_CONTACT_FIELDS} } }`,
  partyContactsByParty: `query EfPartyContactsByParty($party_id: ID!, $first: Int, $after: String) {
    partyContactsByParty(party_id: $party_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} }
    }
  }`,
  primaryPartyContact: `query EfPrimaryPartyContact($party_id: ID!) {
    primaryPartyContact(party_id: $party_id) { ${PARTY_CONTACT_FIELDS} }
  }`,
  partyContactsByIamUserId: `query EfPartyContactsByIam($iam_user_id: ID!, $first: Int, $after: String) {
    partyContactsByIamUserId(iam_user_id: $iam_user_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} }
    }
  }`,
  partyContactsByEmail: `query EfPartyContactsByEmail($email: String!, $first: Int, $after: String) {
    partyContactsByEmail(email: $email, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} }
    }
  }`,
  partyContacts: `query EfPartyContacts($after: String, $before: String, $first: Int, $last: Int, $party_id: ID, $query: String, $contact_role: String, $portal_access_status: PortalAccessStatus, $active_only: Boolean, $include_deleted: Boolean) {
    partyContacts(after: $after, before: $before, first: $first, last: $last, party_id: $party_id, query: $query, contact_role: $contact_role, portal_access_status: $portal_access_status, active_only: $active_only, include_deleted: $include_deleted) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} }
    }
  }`,
  allPartyContacts: `query EfAllPartyContacts($after: String, $before: String, $first: Int, $last: Int) {
    allPartyContacts(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_CONTACT_FIELDS} }
    }
  }`,
  evaluateRepresentativePortalAccess: `query EfEvaluatePortal($party_id: ID!, $iam_user_id: ID!, $required_role: PartyRoleKind!) {
    evaluateRepresentativePortalAccess(party_id: $party_id, iam_user_id: $iam_user_id, required_role: $required_role) {
      allowed party_id contact_id iam_user_id active_roles
    }
  }`,

  // PartyAddressLink (7)
  partyAddressLink: `query EfPartyAddressLink($id: ID!) {
    partyAddressLink(id: $id) { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
  }`,
  partyAddressLinksByParty: `query EfPartyAddressesByParty($party_id: ID!, $first: Int, $after: String) {
    partyAddressLinksByParty(party_id: $party_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} }
      nodes { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
    }
  }`,
  partyAddressLinksByPartyAndUsage: `query EfPartyAddressesByUsage($party_id: ID!, $usage: AddressUsageKind!, $first: Int, $after: String) {
    partyAddressLinksByPartyAndUsage(party_id: $party_id, usage: $usage, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} }
      nodes { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
    }
  }`,
  primaryPartyAddressLink: `query EfPrimaryPartyAddress($party_id: ID!, $usage: AddressUsageKind!) {
    primaryPartyAddressLink(party_id: $party_id, usage: $usage) {
      ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} }
    }
  }`,
  partyAddressLinksByAddressId: `query EfPartyAddressesByAddr($address_id: ID!, $first: Int, $after: String) {
    partyAddressLinksByAddressId(address_id: $address_id, first: $first, after: $after) {
      total_count pageInfo { ${PAGE_INFO} } nodes { ${PARTY_ADDRESS_FIELDS} }
    }
  }`,
  partyAddressLinks: `query EfPartyAddressLinks($after: String, $before: String, $first: Int, $last: Int, $party_id: ID, $address_id: ID, $usage: AddressUsageKind, $primary_only: Boolean) {
    partyAddressLinks(after: $after, before: $before, first: $first, last: $last, party_id: $party_id, address_id: $address_id, usage: $usage, primary_only: $primary_only) {
      total_count pageInfo { ${PAGE_INFO} }
      nodes { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
    }
  }`,
  allPartyAddressLinks: `query EfAllPartyAddressLinks($after: String, $before: String, $first: Int, $last: Int) {
    allPartyAddressLinks(after: $after, before: $before, first: $first, last: $last) {
      total_count pageInfo { ${PAGE_INFO} }
      nodes { ${PARTY_ADDRESS_FIELDS} address { ${ADDRESS_FIELDS} } }
    }
  }`,

  // Relay Node Inspector
  node: `query EfPartnerNode($id: ID!) {
    node(id: $id) {
      id
      ... on Party { partner_code display_name status kind }
      ... on PartyRole { party_id role is_active }
      ... on PartyContact { full_name contact_role portal_access_status }
      ... on PartyAddress { party_id address_id usage is_primary }
    }
  }`,
};

export async function fetchParty(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ party: EfParty | null }>(EF_PARTNER_QUERY_DOCS.party, { id });
    return data.party;
  });
}
export async function fetchPartiesByIds(ids: string[]) {
  const { data } = await executeGraphQL<{ partiesByIds: EfParty[] }>(EF_PARTNER_QUERY_DOCS.partiesByIds, { ids });
  return data.partiesByIds ?? [];
}
export async function fetchPartyByCode(partnerCode: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyByCode: EfParty | null }>(EF_PARTNER_QUERY_DOCS.partyByCode, { partner_code: partnerCode });
    return data.partyByCode;
  });
}
export async function fetchPartyByTaxId(taxIdNumber: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyByTaxIdNumber: EfParty | null }>(EF_PARTNER_QUERY_DOCS.partyByTaxIdNumber, { tax_id_number: taxIdNumber });
    return data.partyByTaxIdNumber;
  });
}
export async function fetchPartiesByEmail(email: string, first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ partiesByEmail: PartyConnection }>(EF_PARTNER_QUERY_DOCS.partiesByEmail, { email, first, after: after || null });
  return data.partiesByEmail;
}
export async function fetchParties(vars: { query?: string; kind?: PartyKind | null; status?: PartyStatus | null; role?: PartyRoleKind | null; include_deleted?: boolean; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ parties: PartyConnection }>(EF_PARTNER_QUERY_DOCS.parties, { first: 20, include_deleted: false, ...vars });
  return data.parties;
}
export async function fetchAllParties(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allParties: PartyConnection }>(EF_PARTNER_QUERY_DOCS.allParties, { first, after: after || null });
  return data.allParties;
}

export async function fetchPartyRole(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyRole: EfPartyRole | null }>(EF_PARTNER_QUERY_DOCS.partyRole, { id });
    return data.partyRole;
  });
}
export async function fetchPartyRolesByParty(partyId: string, first = 20) {
  const { data } = await executeGraphQL<{ partyRolesByParty: PartyRoleConnection }>(EF_PARTNER_QUERY_DOCS.partyRolesByParty, { party_id: partyId, first });
  return data.partyRolesByParty;
}
export async function fetchPartyRoleByPartyAndKind(partyId: string, kind: PartyRoleKind) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyRoleByPartyAndKind: EfPartyRole | null }>(EF_PARTNER_QUERY_DOCS.partyRoleByPartyAndKind, { party_id: partyId, kind });
    return data.partyRoleByPartyAndKind;
  });
}
export async function fetchPartyRoles(vars: { party_id?: string | null; role?: PartyRoleKind | null; active_only?: boolean; query?: string; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ partyRoles: PartyRoleConnection }>(EF_PARTNER_QUERY_DOCS.partyRoles, { first: 20, active_only: false, ...vars });
  return data.partyRoles;
}
export async function fetchAllPartyRoles(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allPartyRoles: PartyRoleConnection }>(EF_PARTNER_QUERY_DOCS.allPartyRoles, { first, after: after || null });
  return data.allPartyRoles;
}
export async function fetchPartyRolesByKind(kind: PartyRoleKind, activeOnly = false, first = 20) {
  const { data } = await executeGraphQL<{ partyRolesByKind: PartyRoleConnection }>(EF_PARTNER_QUERY_DOCS.partyRolesByKind, { kind, active_only: activeOnly, first });
  return data.partyRolesByKind;
}
export async function fetchAllPartyRolesByKind(kind: PartyRoleKind, activeOnly = false, first = 20) {
  const { data } = await executeGraphQL<{ allPartyRolesByKind: PartyRoleConnection }>(EF_PARTNER_QUERY_DOCS.allPartyRolesByKind, { kind, active_only: activeOnly, first });
  return data.allPartyRolesByKind;
}
export async function validatePartyRoleScope(partyId: string, requiredRole: PartyRoleKind) {
  const { data } = await executeGraphQL<{ validatePartyRoleScope: ScopeTargetDecision }>(EF_PARTNER_QUERY_DOCS.validatePartyRoleScope, { party_id: partyId, required_role: requiredRole });
  return data.validatePartyRoleScope;
}

export async function fetchPartyContact(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyContact: EfPartyContact | null }>(EF_PARTNER_QUERY_DOCS.partyContact, { id });
    return data.partyContact;
  });
}
export async function fetchPartyContactsByParty(partyId: string, first = 20) {
  const { data } = await executeGraphQL<{ partyContactsByParty: PartyContactConnection }>(EF_PARTNER_QUERY_DOCS.partyContactsByParty, { party_id: partyId, first });
  return data.partyContactsByParty;
}
export async function fetchPrimaryPartyContact(partyId: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ primaryPartyContact: EfPartyContact | null }>(EF_PARTNER_QUERY_DOCS.primaryPartyContact, { party_id: partyId });
    return data.primaryPartyContact;
  });
}
export async function fetchPartyContactsByIamUserId(iamUserId: string, first = 20) {
  const { data } = await executeGraphQL<{ partyContactsByIamUserId: PartyContactConnection }>(EF_PARTNER_QUERY_DOCS.partyContactsByIamUserId, { iam_user_id: iamUserId, first });
  return data.partyContactsByIamUserId;
}
export async function fetchPartyContactsByEmail(email: string, first = 20) {
  const { data } = await executeGraphQL<{ partyContactsByEmail: PartyContactConnection }>(EF_PARTNER_QUERY_DOCS.partyContactsByEmail, { email, first });
  return data.partyContactsByEmail;
}
export async function fetchPartyContacts(vars: { party_id?: string | null; query?: string; contact_role?: string; portal_access_status?: PortalAccessStatus | null; active_only?: boolean; include_deleted?: boolean; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ partyContacts: PartyContactConnection }>(EF_PARTNER_QUERY_DOCS.partyContacts, { first: 20, active_only: false, include_deleted: false, ...vars });
  return data.partyContacts;
}
export async function fetchAllPartyContacts(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allPartyContacts: PartyContactConnection }>(EF_PARTNER_QUERY_DOCS.allPartyContacts, { first, after: after || null });
  return data.allPartyContacts;
}
export async function evaluateRepresentativePortalAccess(partyId: string, iamUserId: string, requiredRole: PartyRoleKind) {
  const { data } = await executeGraphQL<{ evaluateRepresentativePortalAccess: PortalAccessEvaluation }>(EF_PARTNER_QUERY_DOCS.evaluateRepresentativePortalAccess, { party_id: partyId, iam_user_id: iamUserId, required_role: requiredRole });
  return data.evaluateRepresentativePortalAccess;
}

export async function fetchPartyAddressLink(id: string) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ partyAddressLink: EfPartyAddress | null }>(EF_PARTNER_QUERY_DOCS.partyAddressLink, { id });
    return data.partyAddressLink;
  });
}
export async function fetchPartyAddressLinksByParty(partyId: string, first = 20) {
  const { data } = await executeGraphQL<{ partyAddressLinksByParty: PartyAddressConnection }>(EF_PARTNER_QUERY_DOCS.partyAddressLinksByParty, { party_id: partyId, first });
  return data.partyAddressLinksByParty;
}
export async function fetchPartyAddressLinksByPartyAndUsage(partyId: string, usage: AddressUsageKind, first = 20) {
  const { data } = await executeGraphQL<{ partyAddressLinksByPartyAndUsage: PartyAddressConnection }>(EF_PARTNER_QUERY_DOCS.partyAddressLinksByPartyAndUsage, { party_id: partyId, usage, first });
  return data.partyAddressLinksByPartyAndUsage;
}
export async function fetchPrimaryPartyAddressLink(partyId: string, usage: AddressUsageKind) {
  return nullableOnNotFound(async () => {
    const { data } = await executeGraphQL<{ primaryPartyAddressLink: EfPartyAddress | null }>(EF_PARTNER_QUERY_DOCS.primaryPartyAddressLink, { party_id: partyId, usage });
    return data.primaryPartyAddressLink;
  });
}
export async function fetchPartyAddressLinksByAddressId(addressId: string, first = 20) {
  const { data } = await executeGraphQL<{ partyAddressLinksByAddressId: PartyAddressConnection }>(EF_PARTNER_QUERY_DOCS.partyAddressLinksByAddressId, { address_id: addressId, first });
  return data.partyAddressLinksByAddressId;
}
export async function fetchPartyAddressLinks(vars: { party_id?: string | null; address_id?: string | null; usage?: AddressUsageKind | null; primary_only?: boolean | null; first?: number; after?: string | null; before?: string | null; last?: number }) {
  const { data } = await executeGraphQL<{ partyAddressLinks: PartyAddressConnection }>(EF_PARTNER_QUERY_DOCS.partyAddressLinks, { first: 20, ...vars });
  return data.partyAddressLinks;
}
export async function fetchAllPartyAddressLinks(first = 20, after?: string | null) {
  const { data } = await executeGraphQL<{ allPartyAddressLinks: PartyAddressConnection }>(EF_PARTNER_QUERY_DOCS.allPartyAddressLinks, { first, after: after || null });
  return data.allPartyAddressLinks;
}

export async function fetchEfPartnerNode(id: string) {
  const { data } = await executeGraphQL<{ node: Record<string, unknown> | null }>(EF_PARTNER_QUERY_DOCS.node, { id });
  return data.node;
}
