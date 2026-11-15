package crypto_test

import (
	"bytes"
	"regexp"
	"testing"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
)

func TestOpaqueTokenGenerator_TokensAndDigests(t *testing.T) {
	t.Parallel()

	var gen domainservice.TokenGenerator = crypto.NewOpaqueTokenGenerator()

	plaintext, digest, err := gen.GenerateOpaqueToken(32)
	if err != nil {
		t.Fatalf("GenerateOpaqueToken() error = %v", err)
	}
	if plaintext == "" || len(digest) != 32 {
		t.Fatalf("unexpected token output len(plaintext)=%d len(digest)=%d", len(plaintext), len(digest))
	}
	recomputed := gen.Digest(plaintext)
	if !bytes.Equal(digest, recomputed) {
		t.Fatal("Digest(plaintext) does not match digest returned by GenerateOpaqueToken")
	}
	if emptyDigest := gen.Digest("   "); len(emptyDigest) != 0 {
		t.Fatalf("expected nil/empty digest for blank token, got %d bytes", len(emptyDigest))
	}

	if _, _, err := gen.GenerateOpaqueToken(8); err == nil {
		t.Fatal("expected error when byteLen < 16")
	}
	if _, _, err := gen.GenerateOpaqueToken(512); err == nil {
		t.Fatal("expected error when byteLen exceeds safe upper bound")
	}
}

func TestOpaqueTokenGenerator_RecoveryCodes(t *testing.T) {
	t.Parallel()

	gen := crypto.NewOpaqueTokenGenerator()
	codes, err := gen.GenerateRecoveryCodes(8)
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes() error = %v", err)
	}
	if len(codes) != 8 {
		t.Fatalf("len(codes) = %d, want 8", len(codes))
	}

	pattern := regexp.MustCompile(`^[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}$`)
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		if !pattern.MatchString(code) {
			t.Fatalf("unexpected recovery code format: %q", code)
		}
		if _, dup := seen[code]; dup {
			t.Fatalf("duplicate recovery code generated: %q", code)
		}
		seen[code] = struct{}{}
	}

	if _, err := gen.GenerateRecoveryCodes(0); err == nil {
		t.Fatal("expected error for count <= 0")
	}
	if _, err := gen.GenerateRecoveryCodes(50); err == nil {
		t.Fatal("expected error for count exceeding limit")
	}
}
