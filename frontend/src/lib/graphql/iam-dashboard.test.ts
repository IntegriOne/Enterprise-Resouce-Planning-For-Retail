import test from "node:test";
import assert from "node:assert/strict";
import { AuthRequestError } from "./iam-browser-auth.ts";
import {
  checkUsernameAvailabilityViaGraphQL,
  formatIamDashboardError,
  isForbiddenError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "./iam-dashboard-queries.ts";
import {
  submitAdminCreateUser,
  submitAssignGroupRole,
  submitAssignUserRole,
  submitCreateRole,
  submitUserLogout,
} from "./iam-dashboard-mutations.ts";
import { lookupPartyContactsByIamUserId } from "./ef-scope-lookup.ts";

test("dashboard error helpers classify unauthenticated, forbidden, and step-up MFA accurately", () => {
  const unauth = new AuthRequestError(401, "UNAUTHENTICATED", "session expired");
  assert.equal(isUnauthenticatedError(unauth), true);
  assert.equal(isForbiddenError(unauth), false);

  const stepUp = new AuthRequestError(
    403,
    "FORBIDDEN",
    "mfa step-up authentication required for sensitive action",
    "AUTHORIZATION",
  );
  assert.equal(isForbiddenError(stepUp), true);
  assert.equal(isStepUpRequiredError(stepUp), true);
  assert.match(formatIamDashboardError(stepUp), /verifikasi ulang MFA/);

  const denied = new AuthRequestError(
    403,
    "PERMISSION_DENIED",
    "missing iam.roles.write",
    "AUTHORIZATION",
  );
  assert.equal(isForbiddenError(denied), true);
  assert.equal(isStepUpRequiredError(denied), false);
});

test("assignUserRole and assignGroupRole enforce scope invariants before issuing network requests", async () => {
  await assert.rejects(
    submitAssignUserRole(
      { user_id: "u1", role_id: "r1", scope_type: "LEGAL_ENTITY", scope_id: "" },
      "csrf-token-1",
    ),
    /Scope ID wajib diisi/,
  );

  await assert.rejects(
    submitAssignGroupRole(
      {
        group_id: "g1",
        role_id: "r1",
        scope_type: "SELF" as unknown as "GLOBAL",
        scope_id: "u1",
      },
      "csrf-token-1",
    ),
    /Scope SELF tidak diizinkan/,
  );
});

test("IAM dashboard mutations bind X-CSRF-Token and handle MutationResult payloads", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ csrf: string | null; query: string }> = [];

  globalThis.fetch = (async (_url: URL | RequestInfo, init?: RequestInit) => {
    const headers = (init?.headers ?? {}) as Record<string, string>;
    const body = JSON.parse(String(init?.body || "{}")) as { query: string; variables?: { input?: Record<string, unknown> } };
    requests.push({
      csrf: headers["X-CSRF-Token"] ?? null,
      query: body.query,
    });

    if (body.query.includes("createEmployeeUser(")) {
      assert.equal(body.variables?.input?.provisioning_request_id, "66666666-6666-4666-8666-666666666666");
      assert.equal(body.variables?.input?.worker_id, "55555555-5555-4555-8555-555555555555");
      return new Response(
        JSON.stringify({
          data: {
            createEmployeeUser: {
              user: { id: "gid://integrione/user/1", _id: "1", kind: "EMPLOYEE", status: "PENDING", display_name: "Bima Sakti" },
              errors: [],
              was_successful: true,
            },
          },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }

    if (body.query.includes("createRole(")) {
      return new Response(
        JSON.stringify({
          data: {
            createRole: {
              role: {
                id: "gid://integrione/role/2",
                _id: "2",
                code: "warehouse_lead",
                name: "Kepala Gudang",
                description: "",
                management_domain: "warehouse",
                default_redirect_path: "/warehouse",
                redirect_priority: 20,
                is_system: false,
                is_active: true,
                created_at: "2026-11-25T00:00:00Z",
                updated_at: "2026-11-25T00:00:00Z",
              },
              errors: [],
              was_successful: true,
            },
          },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }

    return new Response(
      JSON.stringify({
        data: {
          userLogout: { logged_out: true, errors: [], was_successful: true },
        },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    const created = await submitAdminCreateUser(
      {
        kind: "EMPLOYEE",
        worker_id: "55555555-5555-4555-8555-555555555555",
        provisioning_request_id: "66666666-6666-4666-8666-666666666666",
        full_name: "Bima Sakti",
        email: "bima@integrione.id",
        username: "bima.sakti",
        password: "RitelAman#2026!Z",
      },
      "csrf-xyz",
    );
    assert.equal(created?.kind, "EMPLOYEE");

    const role = await submitCreateRole(
      {
        code: "warehouse_lead",
        name: "Kepala Gudang",
        management_domain: "warehouse",
        default_redirect_path: "/warehouse",
        redirect_priority: 20,
      },
      "csrf-xyz",
    );
    assert.equal(role?.code, "warehouse_lead");

    const loggedOut = await submitUserLogout("csrf-xyz");
    assert.equal(loggedOut, true);
    assert.equal(requests.length, 3);
    assert.ok(requests.every((r) => r.csrf === "csrf-xyz"));
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("checkUsernameAvailabilityViaGraphQL returns conflicting user when exact USERNAME exists", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = (async () => {
    return new Response(
      JSON.stringify({
        data: {
          users: {
            nodes: [
              {
                id: "gid://integrione/user/1",
                _id: "00000000-0000-4000-8000-000000000001",
                kind: "SYSTEM",
                status: "ACTIVE",
                display_name: "SystemIAM",
                identities: {
                  nodes: [
                    { kind: "EMAIL", value: "systemiam@integrione.local", is_verified: true },
                    { kind: "USERNAME", value: "systemiam", is_verified: true },
                  ],
                },
              },
            ],
          },
        },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    const taken = await checkUsernameAvailabilityViaGraphQL(" SystemIAM ");
    assert.equal(taken.available, false);
    assert.equal(taken.normalizedUsername, "systemiam");
    assert.equal(taken.existingUser?.display_name, "SystemIAM");
    assert.equal(taken.existingUser?.kind, "SYSTEM");

    const free = await checkUsernameAvailabilityViaGraphQL("bima.baru");
    assert.equal(free.available, true);
    assert.equal(free.normalizedUsername, "bima.baru");
    assert.equal(free.existingUser, undefined);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("submitAdminCreateUser dispatches createVendorUser, createSupplierUser, and createUser by kind", async () => {
  const originalFetch = globalThis.fetch;
  const queries: string[] = [];

  globalThis.fetch = (async (_url: URL | RequestInfo, init?: RequestInit) => {
    const body = JSON.parse(String(init?.body || "{}")) as { query: string };
    queries.push(body.query);

    const kind = body.query.includes("createVendorUser(")
      ? "VENDOR"
      : body.query.includes("createSupplierUser(")
        ? "SUPPLIER"
        : "CUSTOMER";
    const field = body.query.includes("createVendorUser(")
      ? "createVendorUser"
      : body.query.includes("createSupplierUser(")
        ? "createSupplierUser"
        : "createUser";

    return new Response(
      JSON.stringify({
        data: {
          [field]: {
            user: {
              id: `gid://integrione/user/${kind.toLowerCase()}`,
              _id: kind.toLowerCase(),
              kind,
              status: "PENDING",
              display_name: `${kind} User`,
            },
            errors: [],
            was_successful: true,
          },
        },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    const vendor = await submitAdminCreateUser(
      {
        kind: "VENDOR",
        party_id: "gid://integrione/party/11111111-1111-4111-8111-111111111111",
        contact_role: "billing_pic",
        full_name: "Mitra CMT Satu",
        email: "vendor@integrione.id",
        username: "vendor.cmt",
        password: "RitelAman#2026!Z",
      },
      "csrf-v",
    );
    assert.equal(vendor?.kind, "VENDOR");
    assert.match(queries[0], /createVendorUser\(/);
    assert.match(queries[0], /party_contact/);

    const supplier = await submitAdminCreateUser(
      {
        kind: "SUPPLIER",
        party_id: "gid://integrione/party/22222222-2222-4222-8222-222222222222",
        contact_id: "gid://integrione/party-contact/33333333-3333-4333-8333-333333333333",
        full_name: "Pemasok Tekstil",
        email: "supplier@integrione.id",
        username: "supplier.tekstil",
        password: "RitelAman#2026!Z",
      },
      "csrf-v",
    );
    assert.equal(supplier?.kind, "SUPPLIER");
    assert.match(queries[1], /createSupplierUser\(/);
    assert.match(queries[1], /party_contact/);

    await assert.rejects(
      () =>
        submitAdminCreateUser(
          {
            kind: "VENDOR",
            full_name: "Tanpa Mitra",
            email: "noparty@integrione.id",
            username: "no.party",
            password: "RitelAman#2026!Z",
          },
          "csrf-v",
        ),
      /Target mitra/,
    );

    const customer = await submitAdminCreateUser(
      {
        kind: "CUSTOMER",
        full_name: "Pembeli Ritel",
        email: "buyer@integrione.id",
        username: "buyer.ritel",
        password: "RitelAman#2026!Z",
      },
      "csrf-v",
    );
    assert.equal(customer?.kind, "CUSTOMER");
    assert.match(queries[2], /createUser\(/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("lookupPartyContactsByIamUserId resolves Enterprise Foundation party_contacts link and degrades safely", async () => {
  const originalFetch = globalThis.fetch;
  try {
    const queries: string[] = [];
    globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
      const parsed = JSON.parse(String(init?.body ?? "{}")) as { query: string };
      queries.push(parsed.query);
      if (queries.length === 1) {
        return new Response(
          JSON.stringify({
            data: {
              partyContactsByIamUserId: {
                nodes: [
                  {
                    id: "gid://integrione/party-contact/33333333-3333-4333-8333-333333333333",
                    _id: "33333333-3333-4333-8333-333333333333",
                    party_id: "gid://integrione/party/11111111-1111-4111-8111-111111111111",
                    iam_user_id: "gid://integrione/user/99999999-9999-4999-8999-999999999999",
                    full_name: "PIC Vendor Utama",
                    job_title: "Finance Lead",
                    contact_role: "billing_pic",
                    email: "pic@vendor.id",
                    phone: "+6281200011122",
                    portal_access_status: "PENDING_LINK",
                    is_primary: true,
                    is_active: true,
                  },
                ],
              },
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      return new Response(
        JSON.stringify({
          data: null,
          errors: [
            {
              message: "enterprise foundation service unavailable",
              extensions: { code: "UNAVAILABLE" },
            },
          ],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }) as typeof fetch;

    const okRes = await lookupPartyContactsByIamUserId("99999999-9999-4999-8999-999999999999");
    assert.match(queries[0], /partyContactsByIamUserId/);
    assert.equal(okRes.contacts.length, 1);
    assert.equal(okRes.contacts[0]?.iamUserId, "gid://integrione/user/99999999-9999-4999-8999-999999999999");
    assert.equal(okRes.contacts[0]?.portalAccessStatus, "PENDING_LINK");

    const degradedRes = await lookupPartyContactsByIamUserId("99999999-9999-4999-8999-999999999999");
    assert.equal(degradedRes.contacts.length, 0);
    assert.ok(degradedRes.unavailableReason);
  } finally {
    globalThis.fetch = originalFetch;
  }
});
