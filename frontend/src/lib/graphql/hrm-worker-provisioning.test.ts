import assert from "node:assert/strict";
import test from "node:test";
import { submitWorkerAccountRequest } from "./hrm-worker-provisioning.ts";

test("preserves the saved HRM request when delivery to IAM fails", async (t) => {
  const request = { status: "pending", requested_at: "2026-11-27T08:00:00Z" };
  t.mock.method(globalThis, "fetch", async () =>
    Response.json({
      data: {
        requestEmployeeAccount: {
          request,
          was_successful: false,
          errors: [{ type: "UNAVAILABLE", message: "Identity service unavailable" }],
        },
      },
    }),
  );

  assert.deepEqual(await submitWorkerAccountRequest("fixture", "csrf"), {
    request,
    deliveryFailed: true,
  });
});
