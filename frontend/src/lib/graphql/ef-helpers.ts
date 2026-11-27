/**
 * Helper akses izin Enterprise Foundation, pemformatan error aman,
 * daftar mutasi sensitif EF (wajib step-up / re-auth <= 10m), serta label status & enum UI.
 */

import type { AccessSnapshot } from "./iam-types.ts";
import { AuthRequestError } from "./iam-browser-auth.ts";
import {
  formatIamDashboardError,
  isStepUpRequiredError,
} from "./iam-dashboard-queries.ts";
import type {
  AddressUsageKind,
  LegalEntityStatus,
  LocationKind,
  OrgUnitKind,
  PartyKind,
  PartyRoleKind,
  PartyStatus,
  PortalAccessStatus,
} from "./ef-types.ts";

export const EF_SENSITIVE_MUTATIONS = [
  "deleteAddress",
  "setPrimaryOperatingLegalEntity",
  "dissolveLegalEntity",
  "deleteLegalEntity",
  "deleteLocation",
  "deleteOrganizationUnit",
  "deleteParty",
  "removePartyRole",
  "activatePartyPortalAccess",
  "revokePartyPortalAccess",
  "unlinkPartyPortalUser",
  "deletePartyContact",
] as const;

export type EfSensitiveMutationName = (typeof EF_SENSITIVE_MUTATIONS)[number];

export const EF_STRUCTURE_PERMISSIONS = [
  "enterprise_foundation.addresses.read",
  "enterprise_foundation.addresses.write",
  "enterprise_foundation.legal_entities.read",
  "enterprise_foundation.legal_entities.write",
  "enterprise_foundation.locations.read",
  "enterprise_foundation.locations.write",
  "enterprise_foundation.organization_units.read",
  "enterprise_foundation.organization_units.write",
] as const;

export const EF_PARTNER_PERMISSIONS = [
  "enterprise_foundation.addresses.read",
  "enterprise_foundation.addresses.write",
  "enterprise_foundation.parties.read",
  "enterprise_foundation.parties.write",
] as const;

export function hasSuperAdminRole(snapshot?: AccessSnapshot | null): boolean {
  if (!snapshot) return false;
  return (snapshot.roles ?? []).some((role) => {
    const code = (role.code || "").trim().toLowerCase();
    return code === "superadmin" || code === "super_admin";
  });
}

export function hasEfPermission(
  snapshot: AccessSnapshot | null | undefined,
  permissionCode: string,
): boolean {
  if (!snapshot) return false;
  if (hasSuperAdminRole(snapshot)) return true;
  return (snapshot.permissions ?? []).some(
    (perm) => perm.permission_code === permissionCode,
  );
}

export function hasAnyEfStructureAccess(
  snapshot?: AccessSnapshot | null,
): boolean {
  if (!snapshot) return false;
  if (hasSuperAdminRole(snapshot)) return true;
  return EF_STRUCTURE_PERMISSIONS.some((code) => hasEfPermission(snapshot, code));
}

export function hasAnyEfPartnerAccess(
  snapshot?: AccessSnapshot | null,
): boolean {
  if (!snapshot) return false;
  if (hasSuperAdminRole(snapshot)) return true;
  return EF_PARTNER_PERMISSIONS.some((code) => hasEfPermission(snapshot, code));
}

export function hasAnyEfAccess(snapshot?: AccessSnapshot | null): boolean {
  if (!snapshot) return false;
  if (hasSuperAdminRole(snapshot)) return true;
  return (snapshot.permissions ?? []).some((perm) =>
    perm.permission_code.startsWith("enterprise_foundation."),
  );
}

export function isSensitiveEfMutation(name: string): boolean {
  return (EF_SENSITIVE_MUTATIONS as readonly string[]).includes(name);
}

/**
 * Mengembalikan `null` bila query pencarian/lookup tunggal melaporkan `NOT_FOUND`,
 * sementara tetap melempar error otorisasi/jaringan lain secara fail-closed.
 */
export async function nullableOnNotFound<T>(
  op: () => Promise<T | null>,
): Promise<T | null> {
  try {
    return await op();
  } catch (err) {
    if (
      err instanceof AuthRequestError &&
      (err.code === "NOT_FOUND" || err.errorType === "NOT_FOUND")
    ) {
      return null;
    }
    throw err;
  }
}

export function formatEfError(err: unknown): string {
  if (isStepUpRequiredError(err)) {
    return "Operasi sensitif Enterprise Foundation ini mewajibkan verifikasi MFA / autentikasi sesi 10 menit terakhir.";
  }
  return formatIamDashboardError(err);
}

export const LEGAL_ENTITY_STATUS_OPTIONS: Array<{
  value: LegalEntityStatus;
  label: string;
}> = [
  { value: "PENDING", label: "PENDING" },
  { value: "ACTIVE", label: "ACTIVE" },
  { value: "SUSPENDED", label: "SUSPENDED" },
  { value: "DISSOLVED", label: "DISSOLVED" },
];

export const LOCATION_KIND_OPTIONS: Array<{
  value: LocationKind;
  label: string;
}> = [
  { value: "HEADQUARTERS", label: "Kantor Pusat (HQ)" },
  { value: "WAREHOUSE", label: "Gudang / DC" },
  { value: "RETAIL_STORE", label: "Toko Ritel" },
  { value: "BRANCH_OFFICE", label: "Kantor Cabang" },
];

export const ORG_UNIT_KIND_OPTIONS: Array<{
  value: OrgUnitKind;
  label: string;
}> = [
  { value: "DIVISION", label: "Divisi" },
  { value: "DEPARTMENT", label: "Departemen" },
  { value: "TEAM", label: "Tim" },
  { value: "BRANCH", label: "Cabang" },
  { value: "OTHER", label: "Lainnya" },
];

export const PARTY_KIND_OPTIONS: Array<{
  value: PartyKind;
  label: string;
}> = [
  { value: "ORGANIZATION", label: "Badan Usaha" },
  { value: "INDIVIDUAL", label: "Perorangan" },
  { value: "FINANCIAL_ENTITY", label: "Lembaga Keuangan" },
  { value: "GOVERNMENT_BODY", label: "Instansi Pemerintah" },
];

export const PARTY_STATUS_OPTIONS: Array<{
  value: PartyStatus;
  label: string;
}> = [
  { value: "PENDING_VERIFICATION", label: "PENDING_VERIFICATION" },
  { value: "ACTIVE", label: "ACTIVE" },
  { value: "SUSPENDED", label: "SUSPENDED" },
  { value: "INACTIVE", label: "INACTIVE" },
];

export const PARTY_ROLE_KIND_OPTIONS: Array<{
  value: PartyRoleKind;
  label: string;
}> = [
  { value: "SUPPLIER", label: "Supplier Bahan" },
  { value: "VENDOR", label: "Vendor Jasa / CMT" },
  { value: "CUSTOMER_B2B", label: "Customer B2B" },
  { value: "CARRIER", label: "Ekspedisi / Carrier" },
  { value: "FINANCIAL_INSTITUTION", label: "Institusi Keuangan" },
  { value: "OTHER", label: "Lainnya" },
];

export const PORTAL_ACCESS_STATUS_OPTIONS: Array<{
  value: PortalAccessStatus;
  label: string;
}> = [
  { value: "NONE", label: "NONE" },
  { value: "PENDING_LINK", label: "PENDING_LINK" },
  { value: "ACTIVE", label: "ACTIVE" },
  { value: "REVOKED", label: "REVOKED" },
];

export const ADDRESS_USAGE_OPTIONS: Array<{
  value: AddressUsageKind;
  label: string;
}> = [
  { value: "REGISTERED_OFFICE", label: "Kantor Terdaftar" },
  { value: "HEADQUARTERS", label: "Kantor Pusat" },
  { value: "BILLING", label: "Penagihan (Billing)" },
  { value: "SHIPPING", label: "Pengiriman (Shipping)" },
  { value: "OTHER", label: "Lainnya" },
];

export function statusBadgeTone(status: string): string {
  const s = (status || "").toUpperCase();
  if (s === "ACTIVE") {
    return "border-[#abefc6] bg-[#ecfdf3] text-[#067647]";
  }
  if (s === "PENDING" || s === "PENDING_VERIFICATION" || s === "PENDING_LINK") {
    return "border-[#fedf89] bg-[#fffaeb] text-[#b54708]";
  }
  if (s === "SUSPENDED" || s === "REVOKED" || s === "DISSOLVED") {
    return "border-[#fecdca] bg-[#fef3f2] text-[#b42318]";
  }
  return "border-[#dde3e8] bg-[#f4f7f8] text-[#5a6672]";
}
