import test from "node:test";
import assert from "node:assert/strict";
import {
  AuthRequestError,
  formatAuthErrorMessage,
  getPasswordStrength,
  submitChangeUsername,
  submitCreateUser,
  submitResendValidationEmailToken,
  submitResetPassword,
  submitUserLogin,
  submitUserResetPassword,
  submitVerifyEmail,
} from "./iam-browser-auth.ts";

test("getPasswordStrength calculates scores correctly", () => {
  assert.deepEqual(getPasswordStrength(""), { score: 0, label: "", color: "#dde3e8" });
  assert.equal(getPasswordStrength("abc").label, "Lemah");
  assert.equal(getPasswordStrength("Abcdefgh1").label, "Cukup");
  assert.equal(getPasswordStrength("Abcdefgh1!").label, "Kuat");
});

test("formatAuthErrorMessage handles AuthRequestError variants", () => {
  const netErr = new AuthRequestError(0, "NETWORK");
  assert.match(formatAuthErrorMessage(netErr, "login"), /Koneksi ke layanan terputus/);

  const unavailErr = new AuthRequestError(503, "UNAVAILABLE");
  assert.match(formatAuthErrorMessage(unavailErr, "login"), /Layanan sedang tidak tersedia/);

  const rateLimitErr = new AuthRequestError(429, "TOO_MANY_REQUESTS");
  assert.match(formatAuthErrorMessage(rateLimitErr, "login"), /Terlalu banyak percobaan/);

  const invalidCreds = new AuthRequestError(401, "INVALID_CREDENTIALS");
  assert.match(formatAuthErrorMessage(invalidCreds, "login"), /pastikan email Anda telah diverifikasi/);

  const conflictErr = new AuthRequestError(409, "CONFLICT", "username already taken", "CONFLICT");
  assert.equal(formatAuthErrorMessage(conflictErr, "register"), "username already taken");

  const argErr = new AuthRequestError(400, "ARGUMENT", "email is required", "ARGUMENT");
  assert.equal(formatAuthErrorMessage(argErr, "reset-request"), "email is required");
});

test("formatAuthErrorMessage handles fallback messages per context", () => {
  assert.match(formatAuthErrorMessage(new Error("random"), "mfa"), /Kode verifikasi tidak sesuai/);
  assert.match(formatAuthErrorMessage(new Error("random"), "register"), /Pendaftaran akun belum dapat diproses/);
  assert.match(formatAuthErrorMessage(new Error("random"), "verify"), /Verifikasi email tidak dapat diproses/);
});

test("IAM email verification, password reset, resend, and changeUsername mutations match MutationResult contract", async () => {
  const originalFetch = globalThis.fetch;
  const captured: Array<{ url: string; body: { query: string; variables?: Record<string, unknown> } }> = [];

  globalThis.fetch = (async (input: URL | RequestInfo, init?: RequestInit) => {
    const rawBody = typeof init?.body === "string" ? init.body : "{}";
    const parsedBody = JSON.parse(rawBody) as { query: string; variables?: Record<string, unknown> };
    captured.push({ url: String(input), body: parsedBody });

    let responseData: Record<string, unknown> = {};
    if (parsedBody.query.includes("verifyEmail(")) {
      responseData = {
        verifyEmail: { status: "VERIFIED", verified: true, errors: [], was_successful: true },
      };
    } else if (parsedBody.query.includes("userResetPassword(")) {
      responseData = {
        userResetPassword: {
          message: "If the email is registered and verified, a password reset link has been sent.",
          errors: [],
          was_successful: true,
        },
      };
    } else if (parsedBody.query.includes("resetPassword(")) {
      responseData = {
        resetPassword: { reset: true, errors: [], was_successful: true },
      };
    } else if (parsedBody.query.includes("resendValidationEmailToken(")) {
      responseData = {
        resendValidationEmailToken: { is_resend: true, errors: [], was_successful: true },
      };
    } else if (parsedBody.query.includes("changeUsername(")) {
      const nextUsername = (parsedBody.variables?.input as { username?: string } | undefined)?.username;
      if (nextUsername === "taken_user") {
        responseData = {
          changeUsername: {
            username: null,
            errors: [{ type: "CONFLICT", message: "username already exists" }],
            was_successful: false,
          },
        };
      } else {
        responseData = {
          changeUsername: { username: nextUsername, errors: [], was_successful: true },
        };
      }
    } else if (parsedBody.query.includes("userLogin(")) {
      responseData = {
        userLogin: {
          status: null,
          mfaChallenge: null,
          mfaExpiresAt: null,
          csrfToken: null,
          redirectLocation: null,
          errors: [{ type: "AUTHENTICATION", message: "invalid credentials" }],
          was_successful: false,
        },
      };
    }

    return new Response(JSON.stringify({ data: responseData }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  }) as typeof fetch;

  try {
    const verifyRes = await submitVerifyEmail("tok-123", "csrf-1");
    assert.deepEqual(verifyRes, {
      status: "VERIFIED",
      verified: true,
      errors: [],
      was_successful: true,
    });

    const resetReq = await submitUserResetPassword("user@integrione.test", "csrf-1");
    assert.equal(resetReq.was_successful, true);
    assert.match(resetReq.message || "", /verified/);

    const resetDone = await submitResetPassword("tok-rst", "NewPass#2026!", "csrf-1");
    assert.equal(resetDone, true);

    const resendGuest = await submitResendValidationEmailToken("pending@integrione.test", "csrf-1");
    assert.equal(resendGuest.is_resend, true);

    const resendAuth = await submitResendValidationEmailToken(undefined, "csrf-1");
    assert.equal(resendAuth.is_resend, true);

    const changed = await submitChangeUsername("kibo", "csrf-1");
    assert.equal(changed.username, "kibo");

    await assert.rejects(
      () => submitChangeUsername("taken_user", "csrf-1"),
      (err: unknown) =>
        err instanceof AuthRequestError &&
        err.errorType === "CONFLICT" &&
        err.status === 409 &&
        err.message === "username already exists",
    );

    await assert.rejects(
      () => submitUserLogin("user@integrione.test", "wrong", "csrf-1"),
      (err: unknown) =>
        err instanceof AuthRequestError &&
        err.errorType === "AUTHENTICATION" &&
        err.status === 401 &&
        err.message === "invalid credentials",
    );

    assert.match(captured[0].body.query, /verifyEmail\(input: \$input\) \{ .*errors \{ type message \} was_successful/);
    assert.deepEqual(captured[0].body.variables, { input: { token: "tok-123" } });

    assert.match(captured[1].body.query, /userResetPassword\(input: \$input\) \{ message errors \{ type message \} was_successful/);
    assert.deepEqual(captured[1].body.variables, { input: { email: "user@integrione.test" } });

    assert.match(captured[2].body.query, /resetPassword\(input: \$input\) \{ reset errors \{ type message \} was_successful/);

    assert.match(captured[3].body.query, /resendValidationEmailToken/);
    assert.deepEqual(captured[3].body.variables, { input: { email: "pending@integrione.test" } });

    assert.deepEqual(captured[4].body.variables, {});

    assert.match(captured[5].body.query, /changeUsername\(input: \$input\) \{ username errors \{ type message \} was_successful/);
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("submitCreateUser handles public success, email duplicate redirect /login, and username taken conflict", async () => {
  const originalFetch = globalThis.fetch;
  const captured: Array<{ query: string; variables?: Record<string, unknown> }> = [];

  globalThis.fetch = (async (_input: URL | RequestInfo, init?: RequestInit) => {
    const rawBody = typeof init?.body === "string" ? init.body : "{}";
    const parsedBody = JSON.parse(rawBody) as {
      query: string;
      variables?: { input?: { full_name?: string; email?: string; username?: string; password?: string } };
    };
    captured.push(parsedBody);

    const input = parsedBody.variables?.input;
    if (input?.username === "taken.user") {
      return new Response(
        JSON.stringify({
          data: {
            createUser: {
              user: null,
              redirectLocation: null,
              errors: [{ type: "CONFLICT", message: "username has been taken" }],
              was_successful: false,
            },
          },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      );
    }

    if (input?.email === "existing@integrione.test") {
      return new Response(
        JSON.stringify({
          data: {
            createUser: {
              user: null,
              redirectLocation: "/login",
              errors: [],
              was_successful: true,
            },
          },
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
            Location: "/login",
          },
        },
      );
    }

    return new Response(
      JSON.stringify({
        data: {
          createUser: {
            user: null,
            redirectLocation: null,
            errors: [],
            was_successful: true,
          },
        },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    const okRes = await submitCreateUser(
      "Budi Santoso",
      "budi@integrione.test",
      "budi.santoso",
      "StrongPass#2026!",
      "csrf-1",
    );
    assert.equal(okRes.payload.was_successful, true);
    assert.equal(okRes.payload.user, null);
    assert.equal(okRes.payload.redirectLocation, null);
    assert.equal(okRes.locationHeader, null);

    const dupEmailRes = await submitCreateUser(
      "Budi Duplikat",
      "existing@integrione.test",
      "budi.baru",
      "StrongPass#2026!",
      "csrf-1",
    );
    assert.equal(dupEmailRes.payload.was_successful, true);
    assert.equal(dupEmailRes.payload.user, null);
    assert.equal(dupEmailRes.payload.redirectLocation, "/login");
    assert.equal(dupEmailRes.locationHeader, "/login");

    await assert.rejects(
      () =>
        submitCreateUser(
          "Budi Taken",
          "existing@integrione.test",
          "taken.user",
          "StrongPass#2026!",
          "csrf-1",
        ),
      (err: unknown) =>
        err instanceof AuthRequestError &&
        err.errorType === "CONFLICT" &&
        err.status === 409 &&
        err.message === "username has been taken",
    );

    assert.match(
      captured[0].query,
      /createUser\(input: \$input\) \{ user \{ id _id kind status display_name \} redirectLocation errors \{ type message \} was_successful \}/,
    );
    assert.deepEqual(captured[0].variables, {
      input: {
        full_name: "Budi Santoso",
        email: "budi@integrione.test",
        username: "budi.santoso",
        password: "StrongPass#2026!",
      },
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});
