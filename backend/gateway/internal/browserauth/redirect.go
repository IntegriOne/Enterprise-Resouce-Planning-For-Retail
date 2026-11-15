package browserauth

import (
	"slices"
	"strings"
)

// RedirectValidator memvalidasi rute tujuan navigasi server-side agar bebas open redirect maupun CRLF injection.
type RedirectValidator struct {
	allowedPrefixes []string
}

// NewRedirectValidator membentuk validator rute berdasarkan allowlist awalan path internal yang aman.
func NewRedirectValidator(allowedPaths []string) *RedirectValidator {
	clean := make([]string, 0, len(allowedPaths))
	for _, p := range allowedPaths {
		trimmed := strings.TrimSpace(p)
		if isSafeInternalPath(trimmed) {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		clean = []string{"/", "/identity"}
	}
	return &RedirectValidator{allowedPrefixes: slices.Clone(clean)}
}

// ResolveSafeLocation mengembalikan jalur aman tervalidasi atau fallback "/" jika kandidat tidak memenuhi allowlist.
func (v *RedirectValidator) ResolveSafeLocation(candidate string) string {
	trimmed := strings.TrimSpace(candidate)
	if !v.IsAllowed(trimmed) {
		return "/"
	}
	return trimmed
}

// IsAllowed memeriksa apakah jalur merupakan path relatif internal bersih dan masuk dalam allowlist.
func (v *RedirectValidator) IsAllowed(path string) bool {
	if v == nil || !isSafeInternalPath(path) {
		return false
	}
	for _, allowed := range v.allowedPrefixes {
		if path == allowed {
			return true
		}
		if allowed != "/" && strings.HasPrefix(path, allowed+"/") {
			return true
		}
	}
	return false
}

func isSafeInternalPath(path string) bool {
	if path == "" || !strings.HasPrefix(path, "/") {
		return false
	}
	if strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\\r\n\t\x00?#@") {
		return false
	}
	if strings.Contains(path, "://") {
		return false
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == ".." || seg == "." {
			return false
		}
	}
	return true
}
