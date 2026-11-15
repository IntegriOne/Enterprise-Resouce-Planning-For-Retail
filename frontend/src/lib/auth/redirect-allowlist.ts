/**
 * Allowlist validator untuk tujuan post-login redirect (Defense-in-depth).
 *
 * Sesuai instruksi dan panduan keamanan IntegriOne:
 * - Hanya menerima canonical exact match: "/identity" atau "/"
 * - Menolak URL absolut (http://, https://), protocol-relative (//),
 *   skema javascript:, data:, encoded traversal (%2F, ..), fragment,
 *   query string tidak dikenal, maupun CRLF injection.
 */

const ALLOWED_EXACT_PATHS = new Set<string>(["/identity", "/"]);

export function validatePostLoginLocation(
  locationHeader: string | null | undefined,
): string | null {
  if (!locationHeader || typeof locationHeader !== "string") {
    return null;
  }

  const trimmed = locationHeader.trim();
  if (!trimmed) {
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
  if (trimmed.includes("?") || trimmed.includes("#")) {
    return null;
  }

  // Tolak path traversal
  if (trimmed.includes("..")) {
    return null;
  }

  // Exact match terhadap allowlist resmi
  if (ALLOWED_EXACT_PATHS.has(trimmed)) {
    return trimmed;
  }

  return null;
}
