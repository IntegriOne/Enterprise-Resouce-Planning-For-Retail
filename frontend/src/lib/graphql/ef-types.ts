/**
 * Tipe TypeScript terketik untuk seluruh entitas, enum, koneksi Relay,
 * dan input mutasi domain Enterprise Foundation (`/human-resource` dan `/procurement`).
 * Seluruh tipe selaras 1:1 dengan kontrak GraphQL Gateway (`ef_*.graphql`).
 */

import type { PageInfo, RelayConnection } from "./iam-types.ts";

export type LegalEntityStatus = "PENDING" | "ACTIVE" | "SUSPENDED" | "DISSOLVED";

export type LocationKind =
  | "HEADQUARTERS"
  | "WAREHOUSE"
  | "RETAIL_STORE"
  | "BRANCH_OFFICE";

export type OrgUnitKind =
  | "DIVISION"
  | "DEPARTMENT"
  | "TEAM"
  | "BRANCH"
  | "OTHER";

export type PartyKind =
  | "ORGANIZATION"
  | "INDIVIDUAL"
  | "FINANCIAL_ENTITY"
  | "GOVERNMENT_BODY";

export type PartyStatus =
  | "PENDING_VERIFICATION"
  | "ACTIVE"
  | "SUSPENDED"
  | "INACTIVE";

export type PartyRoleKind =
  | "SUPPLIER"
  | "VENDOR"
  | "CUSTOMER_B2B"
  | "CARRIER"
  | "FINANCIAL_INSTITUTION"
  | "OTHER";

export type PortalAccessStatus = "NONE" | "PENDING_LINK" | "ACTIVE" | "REVOKED";

export type AddressUsageKind =
  | "REGISTERED_OFFICE"
  | "HEADQUARTERS"
  | "BILLING"
  | "SHIPPING"
  | "OTHER";

export type ScopeTargetKind =
  | "LEGAL_ENTITY"
  | "ORGANIZATION_UNIT"
  | "FACILITY"
  | "SUPPLIER_PARTY"
  | "VENDOR_PARTY";

export type ScopeTargetDecision = {
  eligible: boolean;
  scope_kind: ScopeTargetKind;
  target_id: string;
  code: string;
  legal_entity_id: string;
};

export type EfAddress = {
  id: string;
  _id: string;
  label?: string | null;
  address_line_1: string;
  address_line_2?: string | null;
  city: string;
  province?: string | null;
  postal_code?: string | null;
  country_code: string;
  formatted_address?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  created_at: string;
  updated_at: string;
};

export type EfLegalEntity = {
  id: string;
  _id: string;
  parent_legal_entity_id?: string | null;
  entity_code: string;
  legal_name: string;
  trading_name?: string | null;
  legal_form?: string | null;
  is_primary_operating_entity: boolean;
  status: LegalEntityStatus;
  jurisdiction_country_code: string;
  nib?: string | null;
  tax_id_number?: string | null;
  is_pkp: boolean;
  registered_address_id?: string | null;
  fiscal_year_start_month: number;
  incorporation_date?: string | null;
  dissolution_date?: string | null;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  parent_legal_entity?: EfLegalEntity | null;
  registered_address?: EfAddress | null;
  children?: RelayConnection<EfLegalEntity>;
};

export type EfLocation = {
  id: string;
  _id: string;
  legal_entity_id: string;
  parent_location_id?: string | null;
  address_id?: string | null;
  location_code: string;
  name: string;
  kind: LocationKind;
  phone?: string | null;
  timezone: string;
  operational_attributes_json: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  legal_entity?: EfLegalEntity | null;
  parent_location?: EfLocation | null;
  address?: EfAddress | null;
  children?: RelayConnection<EfLocation>;
};

export type EfOrganizationUnit = {
  id: string;
  _id: string;
  legal_entity_id: string;
  parent_unit_id?: string | null;
  primary_location_id?: string | null;
  unit_code: string;
  name: string;
  kind: OrgUnitKind;
  cost_center_code?: string | null;
  head_person_id?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  legal_entity?: EfLegalEntity | null;
  parent_unit?: EfOrganizationUnit | null;
  primary_location?: EfLocation | null;
  children?: RelayConnection<EfOrganizationUnit>;
};

export type EfPartyRole = {
  id: string;
  _id: string;
  party_id: string;
  role: PartyRoleKind;
  is_active: boolean;
  reference_note?: string | null;
  created_at: string;
  updated_at: string;
};

export type EfPartyContact = {
  id: string;
  _id: string;
  party_id: string;
  iam_user_id?: string | null;
  full_name: string;
  job_title?: string | null;
  contact_role: string;
  email?: string | null;
  phone?: string | null;
  portal_access_status: PortalAccessStatus;
  is_primary: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
};

export type EfPartyAddress = {
  id: string;
  _id: string;
  party_id: string;
  address_id: string;
  usage: AddressUsageKind;
  is_primary: boolean;
  created_at: string;
  updated_at: string;
  address?: EfAddress | null;
};

export type EfParty = {
  id: string;
  _id: string;
  partner_code: string;
  kind: PartyKind;
  display_name: string;
  legal_name?: string | null;
  email?: string | null;
  phone?: string | null;
  website_url?: string | null;
  tax_id_number?: string | null;
  tax_registered_name?: string | null;
  is_pkp: boolean;
  is_tax_exempt: boolean;
  tax_exemption_reason?: string | null;
  tax_verified_at?: string | null;
  preferred_language: string;
  status: PartyStatus;
  notes?: string | null;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  roles?: RelayConnection<EfPartyRole>;
  contacts?: RelayConnection<EfPartyContact>;
  addresses?: RelayConnection<EfPartyAddress>;
};

export type PortalAccessEvaluation = {
  allowed: boolean;
  party_id: string;
  contact_id: string;
  iam_user_id: string;
  active_roles: PartyRoleKind[];
};

export type AddressConnection = RelayConnection<EfAddress> & { pageInfo: PageInfo };
export type LegalEntityConnection = RelayConnection<EfLegalEntity>;
export type LocationConnection = RelayConnection<EfLocation>;
export type OrganizationUnitConnection = RelayConnection<EfOrganizationUnit>;
export type PartyConnection = RelayConnection<EfParty>;
export type PartyRoleConnection = RelayConnection<EfPartyRole>;
export type PartyContactConnection = RelayConnection<EfPartyContact>;
export type PartyAddressConnection = RelayConnection<EfPartyAddress>;

export type AddressFormInput = {
  id?: string;
  label?: string | null;
  address_line_1: string;
  address_line_2?: string | null;
  city: string;
  province?: string | null;
  postal_code?: string | null;
  country_code: string;
  formatted_address?: string | null;
  latitude?: number | null;
  longitude?: number | null;
};

export type CreateLegalEntityFormInput = {
  parent_legal_entity_id?: string | null;
  entity_code: string;
  legal_name: string;
  trading_name?: string | null;
  legal_form?: string | null;
  is_primary_operating_entity?: boolean;
  status?: LegalEntityStatus;
  jurisdiction_country_code: string;
  nib?: string | null;
  tax_id_number?: string | null;
  is_pkp?: boolean;
  registered_address_id?: string | null;
  fiscal_year_start_month?: number;
  incorporation_date?: string | null;
  dissolution_date?: string | null;
};

export type UpdateLegalEntityFormInput = {
  id: string;
  parent_legal_entity_id?: string | null;
  legal_name: string;
  trading_name?: string | null;
  legal_form?: string | null;
  jurisdiction_country_code: string;
  nib?: string | null;
  tax_id_number?: string | null;
  is_pkp?: boolean;
  registered_address_id?: string | null;
  fiscal_year_start_month?: number;
  incorporation_date?: string | null;
  dissolution_date?: string | null;
};

export type CreateLocationFormInput = {
  legal_entity_id: string;
  parent_location_id?: string | null;
  address_id?: string | null;
  location_code: string;
  name: string;
  kind: LocationKind;
  phone?: string | null;
  timezone: string;
  operational_attributes_json?: string | null;
};

export type UpdateLocationFormInput = {
  id: string;
  parent_location_id?: string | null;
  address_id?: string | null;
  name: string;
  kind: LocationKind;
  phone?: string | null;
  timezone: string;
  operational_attributes_json?: string | null;
};

export type CreateOrganizationUnitFormInput = {
  legal_entity_id: string;
  parent_unit_id?: string | null;
  primary_location_id?: string | null;
  unit_code: string;
  name: string;
  kind: OrgUnitKind;
  cost_center_code?: string | null;
  head_person_id?: string | null;
};

export type UpdateOrganizationUnitFormInput = {
  id: string;
  parent_unit_id?: string | null;
  primary_location_id?: string | null;
  name: string;
  kind: OrgUnitKind;
  cost_center_code?: string | null;
  head_person_id?: string | null;
  preserve_head_person?: boolean;
};

export type CreatePartyFormInput = {
  partner_code: string;
  kind: PartyKind;
  display_name: string;
  legal_name?: string | null;
  email?: string | null;
  phone?: string | null;
  website_url?: string | null;
  tax_id_number?: string | null;
  tax_registered_name?: string | null;
  is_pkp?: boolean;
  is_tax_exempt?: boolean;
  tax_exemption_reason?: string | null;
  tax_verified_at?: string | null;
  preferred_language?: string;
  status?: PartyStatus;
  notes?: string | null;
};

export type UpdatePartyProfileFormInput = {
  id: string;
  kind: PartyKind;
  display_name: string;
  legal_name?: string | null;
  email?: string | null;
  phone?: string | null;
  website_url?: string | null;
  preferred_language: string;
  notes?: string | null;
};

export type UpdatePartyTaxProfileFormInput = {
  id: string;
  tax_id_number?: string | null;
  tax_registered_name?: string | null;
  is_pkp?: boolean;
  is_tax_exempt?: boolean;
  tax_exemption_reason?: string | null;
  tax_verified_at?: string | null;
};

export type AddPartyContactFormInput = {
  party_id: string;
  iam_user_id?: string | null;
  full_name: string;
  job_title?: string | null;
  contact_role: string;
  email?: string | null;
  phone?: string | null;
  portal_access_status?: PortalAccessStatus;
  is_primary?: boolean;
};

export type UpdatePartyContactFormInput = {
  id: string;
  full_name: string;
  job_title?: string | null;
  contact_role: string;
  email?: string | null;
  phone?: string | null;
  is_primary?: boolean;
};

export type LinkPartyAddressFormInput = {
  party_id: string;
  address_id: string;
  usage: AddressUsageKind;
  is_primary?: boolean;
  replace_primary?: boolean;
};

export type UpdatePartyAddressLinkFormInput = {
  id: string;
  usage: AddressUsageKind;
  is_primary?: boolean;
  replace_primary?: boolean;
};
