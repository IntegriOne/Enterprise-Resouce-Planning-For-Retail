import test from "node:test";
import assert from "node:assert/strict";
import {
  EF_SENSITIVE_MUTATIONS,
  hasAnyEfAccess,
  hasAnyEfPartnerAccess,
  hasAnyEfStructureAccess,
  hasEfPermission,
  isSensitiveEfMutation,
  nullableOnNotFound,
} from "./ef-helpers.ts";
import { AuthRequestError } from "./iam-browser-auth.ts";
import { EF_STRUCTURE_QUERY_DOCS } from "./ef-structure-queries.ts";
import { EF_STRUCTURE_MUTATION_DOCS } from "./ef-structure-mutations.ts";
import { EF_PARTNER_QUERY_DOCS } from "./ef-partner-queries.ts";
import { EF_PARTNER_MUTATION_DOCS } from "./ef-partner-mutations.ts";
import type { AccessSnapshot } from "./iam-types.ts";

function makeSnapshot(roles: string[], perms: string[]): AccessSnapshot {
  return {
    principal_id: "00000000-0000-4000-8000-000000000001",
    user_kind: "EMPLOYEE",
    status: "ACTIVE",
    mfa_enabled: false,
    auth_epoch: 1,
    perm_revision: 1,
    recovery_epoch: 1,
    can_access_identity: roles.includes("superadmin"),
    default_redirect_path: "/human-resource",
    allowed_redirect_paths: ["/identity", "/human-resource", "/procurement"],
    roles: roles.map((code, idx) => ({
      id: `gid://integrione/role/${idx}`,
      code,
      name: code,
      priority: idx + 1,
      default_redirect_path: "/human-resource",
    })),
    permissions: perms.map((code) => ({
      permission_code: code,
      is_sensitive: false,
      scope: { type: "GLOBAL", id: null },
    })),
    synced_at_unix: 1,
    fresh_until_unix: 9999999999,
  };
}

test("Enterprise Foundation frontend GraphQL docs cover all 64 queries + 48 mutations and Relay nodes", () => {
  const structureQueryNames = Object.keys(EF_STRUCTURE_QUERY_DOCS).filter((k) => k !== "node");
  const partnerQueryNames = Object.keys(EF_PARTNER_QUERY_DOCS).filter((k) => k !== "node");
  const structureMutationNames = Object.keys(EF_STRUCTURE_MUTATION_DOCS);
  const partnerMutationNames = Object.keys(EF_PARTNER_MUTATION_DOCS);

  assert.equal(structureQueryNames.length, 34);
  assert.equal(partnerQueryNames.length, 30);
  assert.equal(structureQueryNames.length + partnerQueryNames.length, 64);

  assert.equal(structureMutationNames.length, 22);
  assert.equal(partnerMutationNames.length, 26);
  assert.equal(structureMutationNames.length + partnerMutationNames.length, 48);

  // Guard schema field names (OrganizationUnit & Location expose `name`, never `unit_name` / `location_name`).
  assert.doesNotMatch(EF_STRUCTURE_QUERY_DOCS.organizationUnits, /\bunit_name\b/);
  assert.doesNotMatch(EF_STRUCTURE_QUERY_DOCS.locations, /\blocation_name\b/);
});

test("EF sensitive mutations registry matches all 12 step-up guarded operations", () => {
  assert.equal(EF_SENSITIVE_MUTATIONS.length, 12);
  for (const name of EF_SENSITIVE_MUTATIONS) {
    assert.equal(isSensitiveEfMutation(name), true);
  }
  assert.equal(isSensitiveEfMutation("createLegalEntity"), false);
});

test("EF permission helpers grant full access to superadmin and scoped access per EF role", () => {
  const superSnap = makeSnapshot(["superadmin"], []);
  assert.equal(hasAnyEfAccess(superSnap), true);
  assert.equal(hasAnyEfStructureAccess(superSnap), true);
  assert.equal(hasAnyEfPartnerAccess(superSnap), true);
  assert.equal(hasEfPermission(superSnap, "enterprise_foundation.parties.write"), true);

  const partnerOnlySnap = makeSnapshot(
    ["ef_partner_manager"],
    ["enterprise_foundation.parties.read", "enterprise_foundation.parties.write"],
  );
  assert.equal(hasAnyEfAccess(partnerOnlySnap), true);
  assert.equal(hasAnyEfPartnerAccess(partnerOnlySnap), true);
  assert.equal(hasAnyEfStructureAccess(partnerOnlySnap), false);

  const customerSnap = makeSnapshot(["customer"], ["iam.profile.read"]);
  assert.equal(hasAnyEfAccess(customerSnap), false);
});

test("nullableOnNotFound converts NOT_FOUND to null while rethrowing FORBIDDEN and UNAVAILABLE", async () => {
  const missing = await nullableOnNotFound(async () => {
    throw new AuthRequestError(200, "NOT_FOUND", "record not found", "NOT_FOUND");
  });
  assert.equal(missing, null);

  await assert.rejects(
    () =>
      nullableOnNotFound(async () => {
        throw new AuthRequestError(200, "FORBIDDEN", "forbidden", "AUTHORIZATION");
      }),
    (err: unknown) => err instanceof AuthRequestError && err.code === "FORBIDDEN",
  );
});
