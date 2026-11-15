// Package browserauth mengimplementasikan batas keamanan HTTP peramban untuk Gateway: cookie __Host-, CSRF, Origin, redirect allowlist, dan response header.
package browserauth

import (
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultSessionCookieName adalah nama cookie sesi buram kanonik yang terikat host.
	DefaultSessionCookieName = "__Host-integrione_session"
	// CSRFHeaderName adalah nama header wajib untuk proteksi mutasi GraphQL.
	CSRFHeaderName = "X-CSRF-Token"
	// LocationHeaderName adalah header navigasi server-side pasca-autentikasi yang lolos allowlist.
	LocationHeaderName = "Location"
)

// BuildSessionCookie membentuk http.Cookie terikat host sesuai atribut keamanan D01 (HttpOnly, Secure, Path=/, SameSite=Lax).
func BuildSessionCookie(name, value string, maxAge time.Duration, secure bool) *http.Cookie {
	if strings.TrimSpace(name) == "" {
		name = DefaultSessionCookieName
	}
	maxSec := int(maxAge.Seconds())
	if maxSec < 0 {
		maxSec = 0
	}
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   "",
		MaxAge:   maxSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// BuildClearSessionCookie membentuk cookie penghapus untuk membersihkan handle sesi/pre-auth dari browser.
func BuildClearSessionCookie(name string, secure bool) *http.Cookie {
	if strings.TrimSpace(name) == "" {
		name = DefaultSessionCookieName
	}
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// ReadCookieToken mengekstrak nilai token buram cookie sesi dari permintaan HTTP bila tersedia dan non-kosong.
func ReadCookieToken(r *http.Request, name string) string {
	if r == nil {
		return ""
	}
	if strings.TrimSpace(name) == "" {
		name = DefaultSessionCookieName
	}
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}
