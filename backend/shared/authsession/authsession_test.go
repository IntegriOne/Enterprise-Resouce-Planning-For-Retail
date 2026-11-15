package authsession

import (
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authz"
)

func TestKeyspaceAndHashing(t *testing.T) {
	t.Parallel()

	ks := NewKeyspace("Development:Local*1")
	if ks.Env() != "development_local_1" {
		t.Fatalf("unexpected sanitized env: %q", ks.Env())
	}

	h := HashOpaqueToken("sample-cookie-secret")
	if len(h) != 64 {
		t.Fatalf("expected 64-char sha256 hex digest, got %d", len(h))
	}
	if got, want := ks.SessionTokenKey(h), "ig:development_local_1:iam:sess:tok:"+h; got != want {
		t.Fatalf("SessionTokenKey() = %q, want %q", got, want)
	}
	if got, want := ks.SessionActivityKey(h), "ig:development_local_1:iam:sess:act:"+h; got != want {
		t.Fatalf("SessionActivityKey() = %q, want %q", got, want)
	}
	if got, want := ks.PreAuthKey(h), "ig:development_local_1:gw:preauth:"+h; got != want {
		t.Fatalf("PreAuthKey() = %q, want %q", got, want)
	}
}

func TestBuildTLSConfigRequiresCAWhenStrict(t *testing.T) {
	t.Parallel()

	if _, err := NewTLSClient(RedisConfig{
		Addr:     "127.0.0.1:6379",
		Username: "gw_runtime",
		Password: "secret",
	}); err == nil {
		t.Fatal("expected error when CACertPath is missing in strict TLS mode")
	}
}

func TestReadSessionContextValidationDefaults(t *testing.T) {
	t.Parallel()

	if authz.DefaultSnapshotMaxAge != 30*time.Second {
		t.Fatalf("expected 30s snapshot ceiling, got %s", authz.DefaultSnapshotMaxAge)
	}
}
