/**
 * Helper pencarian target scope opsional dari domain Enterprise Foundation.
 * Bersifat best-effort agar operator IAM tanpa izin baca EF tetap dapat mengisi
 * Scope ID (UUID) secara manual saat menetapkan peran berscope.
 */

import { executeGraphQL } from "./iam-browser-auth.ts";
import type { ScopeType } from "./iam-types.ts";

export type ScopeTargetOption = {
  id: string;
  rawId: string;
  code: string;
  label: string;
  secondary?: string;
};

const LEGAL_ENTITIES_LOOKUP = `
  query ScopeLookupLegalEntities($query: String, $first: Int) {
    legalEntities(query: $query, first: $first, include_deleted: false) {
      nodes { id _id entity_code legal_name status }
    }
  }
`;

const ORG_UNITS_LOOKUP = `
  query ScopeLookupOrgUnits($query: String, $first: Int) {
    organizationUnits(query: $query, first: $first, active_only: true, include_deleted: false) {
      nodes { id _id unit_code name kind }
    }
  }
`;

const LOCATIONS_LOOKUP = `
  query ScopeLookupLocations($query: String, $first: Int) {
    locations(query: $query, first: $first, active_only: true, include_deleted: false) {
      nodes { id _id location_code name kind }
    }
  }
`;

const PARTIES_LOOKUP = `
  query ScopeLookupParties($query: String, $role: PartyRoleKind, $first: Int) {
    parties(query: $query, role: $role, first: $first, include_deleted: false) {
      nodes { id _id partner_code display_name legal_name status }
    }
  }
`;

const PARTY_CONTACTS_LOOKUP = `
  query LookupPartyContactsForProvisioning($party_id: ID!, $first: Int) {
    partyContactsByParty(party_id: $party_id, first: $first) {
      nodes {
        id
        _id
        party_id
        iam_user_id
        full_name
        job_title
        contact_role
        email
        phone
        portal_access_status
        is_primary
        is_active
      }
    }
  }
`;

export type PartyContactProvisionOption = {
  id: string;
  rawId: string;
  partyId: string;
  fullName: string;
  contactRole: string;
  jobTitle?: string | null;
  email?: string | null;
  phone?: string | null;
  portalAccessStatus: string;
  isPrimary: boolean;
};

export async function lookupEnterpriseScopeTargets(
  scopeType: ScopeType,
  search = "",
): Promise<{ options: ScopeTargetOption[]; unavailableReason?: string }> {
  const q = search.trim() || null;

  try {
    switch (scopeType) {
      case "LEGAL_ENTITY": {
        const { data } = await executeGraphQL<{
          legalEntities: {
            nodes: Array<{
              id: string;
              _id: string;
              entity_code: string;
              legal_name: string;
              status: string;
            }>;
          };
        }>(LEGAL_ENTITIES_LOOKUP, { query: q, first: 15 });
        return {
          options: (data.legalEntities?.nodes ?? []).map((n) => ({
            id: n.id,
            rawId: n._id,
            code: n.entity_code,
            label: n.legal_name,
            secondary: n.status,
          })),
        };
      }
      case "ORGANIZATION_UNIT": {
        const { data } = await executeGraphQL<{
          organizationUnits: {
            nodes: Array<{
              id: string;
              _id: string;
              unit_code: string;
              name: string;
              kind: string;
            }>;
          };
        }>(ORG_UNITS_LOOKUP, { query: q, first: 15 });
        return {
          options: (data.organizationUnits?.nodes ?? []).map((n) => ({
            id: n.id,
            rawId: n._id,
            code: n.unit_code,
            label: n.name,
            secondary: n.kind,
          })),
        };
      }
      case "FACILITY": {
        const { data } = await executeGraphQL<{
          locations: {
            nodes: Array<{
              id: string;
              _id: string;
              location_code: string;
              name: string;
              kind: string;
            }>;
          };
        }>(LOCATIONS_LOOKUP, { query: q, first: 15 });
        return {
          options: (data.locations?.nodes ?? []).map((n) => ({
            id: n.id,
            rawId: n._id,
            code: n.location_code,
            label: n.name,
            secondary: n.kind,
          })),
        };
      }
      case "SUPPLIER_PARTY":
      case "VENDOR_PARTY": {
        const role = scopeType === "SUPPLIER_PARTY" ? "SUPPLIER" : "VENDOR";
        const { data } = await executeGraphQL<{
          parties: {
            nodes: Array<{
              id: string;
              _id: string;
              partner_code: string;
              display_name?: string | null;
              legal_name?: string | null;
              status: string;
            }>;
          };
        }>(PARTIES_LOOKUP, { query: q, role, first: 20 });
        return {
          options: (data.parties?.nodes ?? [])
            .filter((n) => n.status !== "SUSPENDED" && n.status !== "INACTIVE")
            .map((n) => ({
              id: n.id,
              rawId: n._id,
              code: n.partner_code,
              label: (n.display_name || n.legal_name || n.partner_code).trim(),
              secondary: n.status,
            })),
        };
      }
      default:
        return { options: [] };
    }
  } catch {
    return {
      options: [],
      unavailableReason:
        "Pencarian Enterprise Foundation tidak tersedia atau izin baca belum diberikan; masukkan UUID target scope langsung.",
    };
  }
}

export type PartnerContactIamLink = PartyContactProvisionOption & {
  iamUserId?: string | null;
};

type RawPartyContactNode = {
  id: string;
  _id: string;
  party_id: string;
  iam_user_id?: string | null;
  full_name: string;
  job_title?: string | null;
  contact_role: string;
  email?: string | null;
  phone?: string | null;
  portal_access_status: string;
  is_primary: boolean;
  is_active: boolean;
};

const PARTY_CONTACTS_BY_IAM_USER_LOOKUP = /* GraphQL */ `
  query LookupPartyContactsByIamUserId($iam_user_id: ID!, $first: Int) {
    partyContactsByIamUserId(iam_user_id: $iam_user_id, first: $first) {
      nodes {
        id
        _id
        party_id
        iam_user_id
        full_name
        job_title
        contact_role
        email
        phone
        portal_access_status
        is_primary
        is_active
      }
    }
  }
`;

// Memetakan node PartyContact GraphQL ke opsi UI bersama untuk modal & drawer.
function mapPartyContactOption(c: RawPartyContactNode): PartnerContactIamLink {
  return {
    id: c.id,
    rawId: c._id,
    partyId: c.party_id,
    iamUserId: c.iam_user_id,
    fullName: c.full_name,
    contactRole: c.contact_role,
    jobTitle: c.job_title,
    email: c.email,
    phone: c.phone,
    portalAccessStatus: c.portal_access_status,
    isPrimary: c.is_primary,
  };
}

export async function lookupUnlinkedPartyContacts(
  partyId: string,
): Promise<{ contacts: PartyContactProvisionOption[]; unavailableReason?: string }> {
  const cleanParty = partyId.trim();
  if (!cleanParty) {
    return { contacts: [] };
  }
  try {
    const { data } = await executeGraphQL<{
      partyContactsByParty: {
        nodes: RawPartyContactNode[];
      };
    }>(PARTY_CONTACTS_LOOKUP, { party_id: cleanParty, first: 50 });

    const eligible = (data.partyContactsByParty?.nodes ?? [])
      .filter(
        (c) =>
          c.is_active &&
          !c.iam_user_id &&
          c.portal_access_status === "NONE",
      )
      .map(mapPartyContactOption);
    return { contacts: eligible };
  } catch {
    return {
      contacts: [],
      unavailableReason:
        "Daftar PIC mitra tidak dapat dimuat otomatis; Anda tetap dapat membuat kontak PIC baru otomatis atau memasukkan UUID kontak.",
    };
  }
}

// Mengambil daftar PartyContact Enterprise Foundation yang tertaut ke iam_user_id pengguna mitra.
export async function lookupPartyContactsByIamUserId(
  iamUserId: string,
): Promise<{ contacts: PartnerContactIamLink[]; unavailableReason?: string }> {
  const cleanUser = iamUserId.trim();
  if (!cleanUser) {
    return { contacts: [] };
  }
  try {
    const { data } = await executeGraphQL<{
      partyContactsByIamUserId: {
        nodes: RawPartyContactNode[];
      };
    }>(PARTY_CONTACTS_BY_IAM_USER_LOOKUP, { iam_user_id: cleanUser, first: 10 });

    return {
      contacts: (data.partyContactsByIamUserId?.nodes ?? []).map(mapPartyContactOption),
    };
  } catch {
    return {
      contacts: [],
      unavailableReason:
        "Informasi tautan mitra Enterprise Foundation saat ini tidak tersedia atau izin baca belum diberikan.",
    };
  }
}
