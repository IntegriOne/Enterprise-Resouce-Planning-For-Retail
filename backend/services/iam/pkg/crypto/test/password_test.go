package crypto_test

import (
	"strings"
	"testing"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
)

func TestArgon2idHasher_HashAndVerify(t *testing.T) {
	t.Parallel()

	// Pastikan implementasi memenuhi kontrak port domainservice.PasswordHasher.
	hasher, err := crypto.NewArgon2idHasher(crypto.Argon2Params{
		MemoryKiB:   8 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		t.Fatalf("NewArgon2idHasher() error = %v", err)
	}
	var _ domainservice.PasswordHasher = hasher

	password := "Strong#Passw0rd!2026"
	hash1, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !strings.HasPrefix(hash1, "$argon2id$v=19$m=8192,t=1,p=1$") {
		t.Fatalf("unexpected PHC hash prefix: %q", hash1)
	}

	hash2, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() second call error = %v", err)
	}
	if hash1 == hash2 {
		t.Fatal("expected distinct random salts to produce distinct hashes")
	}

	ok, err := hasher.Verify(password, hash1)
	if err != nil || !ok {
		t.Fatalf("Verify(correct) = (%v, %v), want (true, nil)", ok, err)
	}

	ok, err = hasher.Verify("Wrong#Passw0rd!2026", hash1)
	if err != nil {
		t.Fatalf("Verify(wrong) unexpected error = %v", err)
	}
	if ok {
		t.Fatal("Verify(wrong) = true, want false")
	}
}

func TestArgon2idHasher_RejectsWeakConfigAndMalformedInput(t *testing.T) {
	t.Parallel()

	if _, err := crypto.NewArgon2idHasher(crypto.Argon2Params{
		MemoryKiB:   1024, // Terlalu kecil di bawah batas aman 8 MiB.
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}); err == nil {
		t.Fatal("expected error for insecure Argon2 memory parameter")
	}

	def := crypto.NewDefaultArgon2idHasher()
	if _, err := def.Hash(""); err == nil {
		t.Fatal("expected error hashing empty password")
	}
	if ok, err := def.Verify("", "$argon2id$v=19$m=8192,t=1,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"); err == nil || ok {
		t.Fatalf("Verify(empty) = (%v, %v), want (false, error)", ok, err)
	}
	if ok, err := def.Verify("SomeValidPass!123", "$bcrypt$invalid"); err == nil || ok {
		t.Fatalf("Verify(malformed) = (%v, %v), want (false, error)", ok, err)
	}
}
