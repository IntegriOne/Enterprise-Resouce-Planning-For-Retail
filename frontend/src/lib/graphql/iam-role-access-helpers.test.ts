import test from "node:test";
import assert from "node:assert/strict";
import {
  checkRoleCompatibility,
  extractUniqueDomains,
  getPermissionDomain,
  getScopeVerificationHint,
  groupPermissionsByDomain,
  resolveEffectivePermissions,
} from "./iam-role-access-helpers.ts";
import type { IamPermission, IamRole } from "./iam-types.ts";

function makePerm(code: string, isSensitive = false): IamPermission {
  const parts = code.split(".");
  return {
    id: `gid://integrione/permission/${code}`,
    _id: code,
    code,
    resource: parts.slice(0, 2).join("."),
    action: parts[2] ?? "read",
    description: `Permission ${code}`,
    is_sensitive: isSensitive,
    created_at: "2026-11-25T00:00:00Z",
  };
}

function makeRole(code: string, domain: string): IamRole {
  return {
    id: `gid://integrione/role/${code}`,
    _id: code,
    code,
    name: code,
    description: "",
    management_domain: domain,
    default_redirect_path: "/identity",
    redirect_priority: 50,
    is_system: code === "superadmin" || code.endsWith("_automation"),
    is_active: true,
    created_at: "2026-11-25T00:00:00Z",
    updated_at: "2026-11-25T00:00:00Z",
  };
}

test("checkRoleCompatibility matches IAM backend rules for SYSTEM, superadmin, and human users", () => {
  assert.equal(checkRoleCompatibility("ef_automation", "EMPLOYEE").compatible, false);
  assert.equal(checkRoleCompatibility("ef_automation", "SYSTEM").compatible, true);
  assert.equal(checkRoleCompatibility("iam_operator", "SYSTEM").compatible, false);
  assert.equal(checkRoleCompatibility("superadmin", "EMPLOYEE").compatible, true);
  assert.equal(checkRoleCompatibility("superadmin", "VENDOR").compatible, false);
  assert.equal(checkRoleCompatibility("superadmin", "CUSTOMER").compatible, false);
  assert.equal(checkRoleCompatibility("ef_operator", "EMPLOYEE").compatible, true);
  assert.equal(checkRoleCompatibility("ef_operator", "VENDOR").compatible, true);
});

test("groupPermissionsByDomain and resolveEffectivePermissions aggregate and deduplicate permissions", () => {
  const pIamRead = makePerm("iam.users.read");
  const pIamWrite = makePerm("iam.users.write", true);
  const pEfPartiesRead = makePerm("enterprise_foundation.parties.read");
  const pEfLegalWrite = makePerm("enterprise_foundation.legal_entities.write");

  assert.equal(getPermissionDomain(pEfPartiesRead), "enterprise_foundation");

  const grouped = groupPermissionsByDomain([
    pIamWrite,
    pEfLegalWrite,
    pIamRead,
    pEfPartiesRead,
  ]);
  assert.equal(grouped.length, 2);
  assert.equal(grouped[0]?.domain, "enterprise_foundation");
  assert.equal(grouped[0]?.permissions.length, 2);
  assert.equal(grouped[1]?.domain, "iam");
  assert.equal(grouped[1]?.sensitiveCount, 1);

  const domains = extractUniqueDomains(
    [pIamRead, pEfPartiesRead],
    [makeRole("custom_hr", "human_resource")],
  );
  assert.deepEqual(domains, ["enterprise_foundation", "human_resource", "iam"]);

  const r1 = makeRole("superadmin", "iam");
  const r2 = makeRole("ef_operator", "enterprise_foundation");
  const resolved = resolveEffectivePermissions([r1, r2], {
    superadmin: [pIamRead, pEfPartiesRead],
    ef_operator: [pEfPartiesRead, pEfLegalWrite],
  });
  assert.equal(resolved.length, 3);
  const shared = resolved.find(
    (item) => item.permission.code === "enterprise_foundation.parties.read",
  );
  assert.equal(shared?.grantedByRoles.length, 2);
});

test("getScopeVerificationHint accurately flags GLOBAL/SELF vs EF external object scopes", () => {
  assert.equal(getScopeVerificationHint("GLOBAL").supportedNow, true);
  assert.equal(getScopeVerificationHint("SELF").supportedNow, true);
  assert.equal(getScopeVerificationHint("LEGAL_ENTITY").supportedNow, true);
  assert.equal(getScopeVerificationHint("LEGAL_ENTITY").badgeLabel, "TERVERIFIKASI EF");
  assert.equal(getScopeVerificationHint("ORGANIZATION_UNIT").supportedNow, true);
  assert.equal(getScopeVerificationHint("FACILITY").supportedNow, true);
  assert.equal(getScopeVerificationHint("SUPPLIER_PARTY").supportedNow, true);
  assert.equal(getScopeVerificationHint("VENDOR_PARTY").supportedNow, true);
});
