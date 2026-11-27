/**
 * Allowlist validator untuk tujuan post-login redirect (Defense-in-depth).
 *
 * Selaras dengan daftar rute portal resmi IntegriOne di Gateway (`defaultAllowedRedirects`):
 * - Rute portal utama & sub-rute relatif bersih tanpa traversal, query, maupun fragment.
 * - Menolak URL absolut (http://, https://), protocol-relative (//),
 *   skema javascript:, data:, encoded traversal (%2F, ..), fragment,
 *   query string tidak dikenal, maupun CRLF injection.
 */

const ALLOWED_PORTAL_PREFIXES = [
  "/",
  "/identity",
  "/human-resource",
  "/procurement",
  "/warehouse",
  "/finance",
  "/accounting",
  "/product",
  "/supplier",
  "/vendor",
  "/commerce",
  "/account",
  "/profile",
] as const;

export function validatePostLoginLocation(
  locationHeader: string | null | undefined,
): string | null {
  if (!locationHeader || typeof locationHeader !== "string") {
    return null;
  }

  const trimmed = locationHeader.trim();
  if (!trimmed || !trimmed.startsWith("/")) {
    return null;
  }

  // Tolak CRLF injection
  if (/[\r\n]/.test(trimmed)) {
    return null;
  }

  // Tolak skema berbahaya atau URL absolut
  if (
    trimmed.startsWith("//") ||
    /^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(trimmed) ||
    trimmed.toLowerCase().includes("javascript:") ||
    trimmed.toLowerCase().includes("data:")
  ) {
    return null;
  }

  // Tolak percent-encoding untuk mencegah bypass traversal (mis. %2e%2e atau %2f)
  if (/%[0-9a-fA-F]{2}/.test(trimmed)) {
    return null;
  }

  // Tolak query string atau hash fragment pada post-login location canonical
  if (trimmed.includes("?") || trimmed.includes("#") || trimmed.includes("\\")) {
    return null;
  }

  // Tolak path traversal per segmen
  const segments = trimmed.split("/");
  if (segments.some((seg) => seg === ".." || seg === ".")) {
    return null;
  }

  // Izinkan kecocokan persis pada portal resmi atau sub-rute valid di bawah prefiks non-root
  for (const allowed of ALLOWED_PORTAL_PREFIXES) {
    if (trimmed === allowed) {
      return trimmed;
    }
    if (allowed !== "/" && trimmed.startsWith(`${allowed}/`)) {
      return trimmed;
    }
  }

  return null;
}
