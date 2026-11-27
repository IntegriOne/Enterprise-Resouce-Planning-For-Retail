/**
 * Helper murni untuk memetakan kompatibilitas Peran RBAC, pengelompokan
 * izin atomik berdasarkan domain management, serta agregasi izin efektif akun.
 * Aturan kompatibilitas diselaraskan 1:1 dengan `IsRoleCompatibleWithUserKind`
 * di service IAM backend.
 */

import type {
  IamPermission,
  IamRole,
  ScopeType,
  UserKind,
} from "./iam-types.ts";

export type ManagementDomainPreset = {
  value: string;
  label: string;
  suggestedRedirect: string;
  note: string;
};

export const MANAGEMENT_DOMAIN_PRESETS: ManagementDomainPreset[] = [
  {
    value: "iam",
    label: "IAM (iam.*)",
    suggestedRedirect: "/identity",
    note: "Pengelolaan user, peran, grup, izin, dan sesi",
  },
  {
    value: "enterprise_foundation",
    label: "Enterprise Foundation (enterprise_foundation.*)",
    suggestedRedirect: "/human-resource",
    note: "Badan hukum, unit organisasi, lokasi, dan mitra party (diserap HR / Procurement / Identity)",
  },
  {
    value: "human_resource",
    label: "Human Resource (human_resource.*)",
    suggestedRedirect: "/human-resource",
    note: "Data pekerja, posisi, kontrak, cuti, dan payroll",
  },
  {
    value: "procurement_sourcing",
    label: "Procurement & Sourcing",
    suggestedRedirect: "/procurement",
    note: "Pembelian, RFx, PO, dan ASN pemasok",
  },
  {
    value: "inventory_warehouse_management",
    label: "Warehouse & Inventory",
    suggestedRedirect: "/warehouse",
    note: "Gudang, penerimaan, stok, dan pengiriman",
  },
  {
    value: "finance_accounting",
    label: "Finance & Accounting",
    suggestedRedirect: "/finance",
    note: "AP/AR, pembayaran, jurnal, dan laporan keuangan",
  },
  {
    value: "product_lifecycle_management",
    label: "Product Lifecycle (PLM)",
    suggestedRedirect: "/product",
    note: "Master SKU, varian, dan spesifikasi produk",
  },
  {
    value: "ecommerce_commerce",
    label: "Web Commerce",
    suggestedRedirect: "/commerce",
    note: "Etalase, keranjang, checkout, dan pesanan toko online",
  },
];

export type RoleCompatibility = {
  compatible: boolean;
  reason: string | null;
  isSystemAutomation: boolean;
  isSuperAdmin: boolean;
};

export function isSuperAdminRoleCode(code: string): boolean {
  const normalized = code.trim().toLowerCase();
  return normalized === "superadmin" || normalized === "super_admin";
}

export function isAutomationRoleCode(code: string): boolean {
  return code.trim().toLowerCase().endsWith("_automation");
}

// Menyelaraskan validasi role <-> user_kind sebelum mutasi sensitif dikirim ke server IAM.
export function checkRoleCompatibility(
  roleCode: string,
  userKind: UserKind,
): RoleCompatibility {
  const isSystemAutomation = isAutomationRoleCode(roleCode);
  const isSuperAdmin = isSuperAdminRoleCode(roleCode);

  if (userKind === "SYSTEM") {
    if (!isSystemAutomation) {
      return {
        compatible: false,
        reason: "Akun SYSTEM hanya boleh menerima peran *_automation.",
        isSystemAutomation,
        isSuperAdmin,
      };
    }
    return { compatible: true, reason: null, isSystemAutomation, isSuperAdmin };
  }

  if (isSystemAutomation) {
    return {
      compatible: false,
      reason: "Peran otomasi (*_automation) khusus akun layanan SYSTEM, bukan akun manusia.",
      isSystemAutomation,
      isSuperAdmin,
    };
  }

  if (isSuperAdmin && userKind !== "EMPLOYEE") {
    return {
      compatible: false,
      reason: "Peran superadmin hanya dapat ditetapkan kepada akun EMPLOYEE internal.",
      isSystemAutomation,
      isSuperAdmin,
    };
  }

  return { compatible: true, reason: null, isSystemAutomation, isSuperAdmin };
}

export function getPermissionDomain(
  permission: Pick<IamPermission, "code" | "resource">,
): string {
  const source = (permission.code || permission.resource || "").trim();
  const dotIdx = source.indexOf(".");
  if (dotIdx > 0) {
    return source.slice(0, dotIdx);
  }
  return source || "other";
}

export type PermissionDomainGroup = {
  domain: string;
  permissions: IamPermission[];
  sensitiveCount: number;
};

export function groupPermissionsByDomain(
  permissions: IamPermission[],
): PermissionDomainGroup[] {
  const buckets = new Map<string, IamPermission[]>();
  for (const perm of permissions) {
    const domain = getPermissionDomain(perm);
    const list = buckets.get(domain);
    if (list) {
      list.push(perm);
    } else {
      buckets.set(domain, [perm]);
    }
  }

  return Array.from(buckets.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([domain, list]) => {
      const sorted = [...list].sort((x, y) => x.code.localeCompare(y.code));
      return {
        domain,
        permissions: sorted,
        sensitiveCount: sorted.filter((item) => item.is_sensitive).length,
      };
    });
}

export function extractUniqueDomains(
  permissions: IamPermission[],
  roles: IamRole[] = [],
): string[] {
  const domains = new Set<string>();
  for (const perm of permissions) {
    const d = getPermissionDomain(perm);
    if (d) domains.add(d);
  }
  for (const role of roles) {
    const d = role.management_domain?.trim();
    if (d) domains.add(d);
  }
  return Array.from(domains).sort((a, b) => a.localeCompare(b));
}

export type EffectivePermissionItem = {
  permission: IamPermission;
  domain: string;
  grantedByRoles: Array<{ id: string; _id: string; code: string; name: string }>;
};

export function resolveEffectivePermissions(
  effectiveRoles: IamRole[],
  rolePermissionsMap: Record<string, IamPermission[]>,
): EffectivePermissionItem[] {
  const byCode = new Map<string, EffectivePermissionItem>();

  for (const role of effectiveRoles) {
    const perms = rolePermissionsMap[role._id] ?? [];
    for (const perm of perms) {
      const existing = byCode.get(perm.code);
      const roleRef = {
        id: role.id,
        _id: role._id,
        code: role.code,
        name: role.name,
      };
      if (existing) {
        if (!existing.grantedByRoles.some((r) => r._id === role._id)) {
          existing.grantedByRoles.push(roleRef);
        }
      } else {
        byCode.set(perm.code, {
          permission: perm,
          domain: getPermissionDomain(perm),
          grantedByRoles: [roleRef],
        });
      }
    }
  }

  return Array.from(byCode.values()).sort((a, b) =>
    a.permission.code.localeCompare(b.permission.code),
  );
}

export type ScopeVerificationHint = {
  supportedNow: boolean;
  badgeLabel: string;
  message: string;
};

export function getScopeVerificationHint(
  scopeType: ScopeType,
): ScopeVerificationHint {
  if (scopeType === "GLOBAL" || scopeType === "SELF") {
    return {
      supportedNow: true,
      badgeLabel: "TERVERIFIKASI",
      message:
        scopeType === "GLOBAL"
          ? "Scope GLOBAL langsung aktif dan dapat ditetapkan ke peran staf/admin."
          : "Scope SELF mengikat hak akses ke akun pengguna ini sendiri.",
    };
  }

  return {
    supportedNow: true,
    badgeLabel: "TERVERIFIKASI EF",
    message:
      "Server IAM memverifikasi kelayakan target scope ini secara real-time ke layanan Enterprise Foundation (target harus ada dan berstatus ACTIVE).",
  };
}
