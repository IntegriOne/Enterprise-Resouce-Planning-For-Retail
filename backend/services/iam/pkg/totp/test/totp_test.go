package totp_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/totp"
)

func TestRFC6238Provider_TestVectorsAndVerify(t *testing.T) {
	t.Parallel()

	// Vektor uji resmi RFC 6238 Appendix B (SHA-1, secret ASCII 20 byte).
	rfcSecret := []byte("12345678901234567890")
	provider := totp.NewDefaultProvider()
	var _ domainservice.TOTPProvider = provider

	// Uji beberapa timestamp dari tabel vektor RFC 6238 (6 digit terakhir dari 8-digit SHA-1 RFC).
	vectors := []struct {
		unix int64
		code string
		step int64
	}{
		{unix: 59, code: "287082", step: 1},
		{unix: 1111111109, code: "081804", step: 37037036},
		{unix: 1111111111, code: "050471", step: 37037037},
		{unix: 1234567890, code: "005924", step: 41152263},
	}

	for _, tc := range vectors {
		at := time.Unix(tc.unix, 0)
		generated, err := provider.GenerateCodeAt(rfcSecret, at)
		if err != nil {
			t.Fatalf("GenerateCodeAt(%d) error = %v", tc.unix, err)
		}
		if generated != tc.code {
			t.Fatalf("GenerateCodeAt(%d) = %q, want %q", tc.unix, generated, tc.code)
		}

		matched, step, err := provider.VerifyCode(rfcSecret, tc.code, at)
		if err != nil {
			t.Fatalf("VerifyCode(%d) error = %v", tc.unix, err)
		}
		if !matched || step != tc.step {
			t.Fatalf("VerifyCode(%d) = (%v, %d), want (true, %d)", tc.unix, matched, step, tc.step)
		}
	}

	// Uji skew +-1 step (30 detik sebelumnya tetap valid pada step sebelumnya).
	matchedPrev, prevStep, err := provider.VerifyCode(rfcSecret, "287082", time.Unix(65, 0))
	if err != nil || !matchedPrev || prevStep != 1 {
		t.Fatalf("VerifyCode(skew -1) = (%v, %d, %v), want (true, 1, nil)", matchedPrev, prevStep, err)
	}

	// Di luar jendela skew (step 3 untuk kode step 1) harus ditolak.
	matchedFar, _, err := provider.VerifyCode(rfcSecret, "287082", time.Unix(120, 0))
	if err != nil || matchedFar {
		t.Fatalf("VerifyCode(out of window) = (%v, %v), want (false, nil)", matchedFar, err)
	}

	// Kode malformed atau panjang salah ditolak tanpa panic.
	if matched, _, err := provider.VerifyCode(rfcSecret, "12ab56", time.Unix(59, 0)); err != nil || matched {
		t.Fatalf("VerifyCode(non-digit) = (%v, %v), want (false, nil)", matched, err)
	}
}

func TestRFC6238Provider_GenerateSecretURI(t *testing.T) {
	t.Parallel()

	provider := totp.NewDefaultProvider()
	secret, uri, err := provider.GenerateSecret("admin@integrione.local", "IntegriOne ERP")
	if err != nil {
		t.Fatalf("GenerateSecret() error = %v", err)
	}
	if len(secret) != 20 {
		t.Fatalf("len(secret) = %d, want 20", len(secret))
	}
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("unexpected otpauth URI prefix: %q", uri)
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("url.Parse(uri) error = %v", err)
	}
	q := parsed.Query()
	if q.Get("issuer") != "IntegriOne ERP" {
		t.Fatalf("issuer query = %q, want %q", q.Get("issuer"), "IntegriOne ERP")
	}
	if q.Get("digits") != "6" || q.Get("period") != "30" || q.Get("algorithm") != "SHA1" {
		t.Fatalf("unexpected totp params: %v", q)
	}
	if q.Get("secret") == "" {
		t.Fatal("missing secret query parameter")
	}

	// Secret hasil GenerateSecret langsung bisa dipakai memverifikasi kode waktu sekarang.
	now := time.Now()
	code, err := provider.GenerateCodeAt(secret, now)
	if err != nil {
		t.Fatalf("GenerateCodeAt() error = %v", err)
	}
	matched, _, err := provider.VerifyCode(secret, code, now)
	if err != nil || !matched {
		t.Fatalf("VerifyCode(generated) = (%v, %v), want (true, nil)", matched, err)
	}

	if _, _, err := provider.GenerateSecret("", "IntegriOne"); err == nil {
		t.Fatal("expected error when accountName is empty")
	}
}
