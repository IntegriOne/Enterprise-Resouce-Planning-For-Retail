import test from "node:test";
import assert from "node:assert/strict";
import { validatePostLoginLocation } from "./redirect-allowlist.ts";

test("validatePostLoginLocation allows exact allowed canonical paths", () => {
  assert.equal(validatePostLoginLocation("/identity"), "/identity");
  assert.equal(validatePostLoginLocation("/"), "/");
  assert.equal(validatePostLoginLocation("  /identity  "), "/identity");
  assert.equal(validatePostLoginLocation("  /  "), "/");
});

test("validatePostLoginLocation rejects missing or empty values", () => {
  assert.equal(validatePostLoginLocation(null), null);
  assert.equal(validatePostLoginLocation(undefined), null);
  assert.equal(validatePostLoginLocation(""), null);
  assert.equal(validatePostLoginLocation("   "), null);
});

test("validatePostLoginLocation rejects absolute URLs and open redirects", () => {
  assert.equal(validatePostLoginLocation("https://evil.com"), null);
  assert.equal(validatePostLoginLocation("http://evil.com/identity"), null);
  assert.equal(validatePostLoginLocation("//evil.com/identity"), null);
  assert.equal(validatePostLoginLocation("//identity"), null);
});

test("validatePostLoginLocation rejects javascript:, data:, and custom schemes", () => {
  assert.equal(validatePostLoginLocation("javascript:alert(1)"), null);
  assert.equal(validatePostLoginLocation("data:text/html,<script>alert(1)</script>"), null);
  assert.equal(validatePostLoginLocation("vbscript:msgbox(1)"), null);
});

test("validatePostLoginLocation rejects percent-encoded and traversal attempts", () => {
  assert.equal(validatePostLoginLocation("%2Fidentity"), null);
  assert.equal(validatePostLoginLocation("/identity%2Fevil"), null);
  assert.equal(validatePostLoginLocation("/identity/.."), null);
  assert.equal(validatePostLoginLocation("/../identity"), null);
});

test("validatePostLoginLocation rejects CRLF injection", () => {
  assert.equal(validatePostLoginLocation("/identity\r\nSet-Cookie: evil=1"), null);
  assert.equal(validatePostLoginLocation("/\nLocation: https://evil.com"), null);
});

test("validatePostLoginLocation rejects queries, fragments, and unlisted paths", () => {
  assert.equal(validatePostLoginLocation("/identity?next=admin"), null);
  assert.equal(validatePostLoginLocation("/#hash"), null);
  assert.equal(validatePostLoginLocation("/dashboard"), null);
  assert.equal(validatePostLoginLocation("/admin"), null);
  assert.equal(validatePostLoginLocation("/login"), null);
});
