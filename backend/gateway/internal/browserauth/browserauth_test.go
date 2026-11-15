package browserauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestBuildCookiesAndCSRF(t *testing.T) {
	c := BuildSessionCookie(DefaultSessionCookieName, "opaque-token", 30*time.Minute, true)
	if c.Name != "__Host-integrione_session" || !c.HttpOnly || !c.Secure || c.Path != "/" || c.Domain != "" || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected cookie attributes: %+v", c)
	}

	clr := BuildClearSessionCookie(DefaultSessionCookieName, true)
	if clr.MaxAge != -1 || clr.Value != "" {
		t.Fatalf("expected clearing cookie, got %+v", clr)
	}

	plain, digest, err := GenerateOpaquePair()
	if err != nil {
		t.Fatalf("GenerateOpaquePair failed: %v", err)
	}
	if !VerifyCSRFTokenDigest(plain, digest) {
		t.Fatal("expected CSRF digest verification to pass")
	}
	if VerifyCSRFTokenDigest(plain+"tampered", digest) {
		t.Fatal("expected tampered CSRF token to fail")
	}
	if VerifyCSRFTokenDigest("", digest) || VerifyCSRFTokenDigest(plain, "not-hex") {
		t.Fatal("expected empty/malformed CSRF check to fail")
	}
}

func TestOriginValidator(t *testing.T) {
	v, err := NewOriginValidator([]string{"https://erp.localhost:8443"})
	if err != nil {
		t.Fatalf("NewOriginValidator failed: %v", err)
	}

	// Mutasi wajib membawa Origin cocok dan Sec-Fetch-Site aman.
	hOK := http.Header{"Origin": []string{"https://erp.localhost:8443"}, "Sec-Fetch-Site": []string{"same-origin"}}
	if err := v.ValidateRequest(hOK, true); err != nil {
		t.Fatalf("expected valid same-origin header, got: %v", err)
	}
	if err := v.ValidateRequest(http.Header{}, true); err == nil {
		t.Fatal("expected missing Origin on mutation to be rejected")
	}
	if err := v.ValidateRequest(http.Header{"Origin": []string{"https://evil.example"}}, true); err == nil {
		t.Fatal("expected foreign Origin to be rejected")
	}
	if err := v.ValidateRequest(http.Header{"Origin": []string{"null"}}, true); err == nil {
		t.Fatal("expected null Origin to be rejected")
	}
	if err := v.ValidateRequest(http.Header{"Origin": []string{"https://erp.localhost:8443"}, "Sec-Fetch-Site": []string{"cross-site"}}, true); err == nil {
		t.Fatal("expected cross-site Sec-Fetch-Site to be rejected")
	}

	// SSR Query cookie-only tidak mengirim Origin tetapi tetap menolak Origin asing.
	if err := v.ValidateRequest(http.Header{}, false); err != nil {
		t.Fatalf("expected cookie-only SSR query without Origin to pass, got: %v", err)
	}
	if err := v.ValidateRequest(http.Header{"Origin": []string{"https://attacker.tld"}}, false); err == nil {
		t.Fatal("expected SSR query with foreign Origin to be rejected")
	}
}

func TestRedirectValidator(t *testing.T) {
	rv := NewRedirectValidator([]string{"/", "/identity", "/commerce"})

	cases := []struct {
		in   string
		want string
	}{
		{"/identity", "/identity"},
		{"/identity/users", "/identity/users"},
		{"/commerce", "/commerce"},
		{"https://evil.com/phish", "/"},
		{"//evil.com", "/"},
		{"/identity\r\nX-Injected: 1", "/"},
		{"/identity?next=https://evil.com", "/"},
		{"/identity/../etc/passwd", "/"},
		{"/unknown-admin", "/"},
	}
	for _, tc := range cases {
		if got := rv.ResolveSafeLocation(tc.in); got != tc.want {
			t.Errorf("ResolveSafeLocation(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResponseEffectsConcurrentAndCRLF(t *testing.T) {
	eff := NewResponseEffects()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			eff.SetHeader(CSRFHeaderName, "csrf-val\r\nX-Injected: bad")
			eff.AddCookie(BuildSessionCookie(fmt.Sprintf("%s_%d", DefaultSessionCookieName, idx), "tok", time.Hour, true))
		}(i)
	}
	wg.Wait()

	rec := httptest.NewRecorder()
	ew := NewEffectResponseWriter(rec, eff)
	_, _ = ew.Write([]byte(`{"data":{"health":"ok"}}`))

	gotCSRF := rec.Header().Get(CSRFHeaderName)
	if gotCSRF != "csrf-valX-Injected: bad" {
		t.Fatalf("unexpected sanitized header value: %q", gotCSRF)
	}
	if len(rec.Result().Cookies()) != 16 {
		t.Fatalf("expected 16 cookies flushed, got %d", len(rec.Result().Cookies()))
	}

	// Deduplikasi cookie bernama sama harus menyisakan efek terakhir (mis. Set -> Clear).
	dedup := NewResponseEffects()
	dedup.AddCookie(BuildSessionCookie(DefaultSessionCookieName, "initial-tok", time.Hour, true))
	dedup.AddCookie(&http.Cookie{Name: DefaultSessionCookieName, Value: "", MaxAge: -1, Path: "/"})
	dedupRec := httptest.NewRecorder()
	dedupWriter := NewEffectResponseWriter(dedupRec, dedup)
	_, _ = dedupWriter.Write([]byte(`{"data":{"logout":true}}`))

	cookies := dedupRec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 deduplicated cookie, got %d", len(cookies))
	}
	if cookies[0].Name != DefaultSessionCookieName || cookies[0].Value != "" || cookies[0].MaxAge != -1 {
		t.Fatalf("expected cleared cookie to win deduplication, got %+v", cookies[0])
	}
}
