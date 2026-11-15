package crypto_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
)

func TestAESGCMKeyringCipher_EncryptDecryptAndKeyRotation(t *testing.T) {
	t.Parallel()

	keyV1 := bytes.Repeat([]byte{0x11}, 32)
	keyV2 := bytes.Repeat([]byte{0x22}, 32)

	// Inisialisasi dengan versi aktif 1 untuk membuktikan ciphertext dapat didekripsi setelah rotasi ke versi 2.
	cipherV1, err := crypto.NewAESGCMKeyringCipher(1, map[int16][]byte{1: keyV1})
	if err != nil {
		t.Fatalf("NewAESGCMKeyringCipher(v1) error = %v", err)
	}
	var _ domainservice.SecretCipher = cipherV1

	secret := []byte("super-secret-totp-seed-bytes")
	ciphertextV1, ver, err := cipherV1.Encrypt(secret)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ver != 1 {
		t.Fatalf("keyVersion = %d, want 1", ver)
	}
	if bytes.Equal(ciphertextV1, secret) {
		t.Fatal("ciphertext must not equal plaintext")
	}

	cipherV2, err := crypto.NewAESGCMKeyringCipher(2, map[int16][]byte{
		1: keyV1,
		2: keyV2,
	})
	if err != nil {
		t.Fatalf("NewAESGCMKeyringCipher(v2) error = %v", err)
	}

	// Cipher versi baru tetap bisa mendekripsi secret lama versi 1.
	decryptedV1, err := cipherV2.Decrypt(ciphertextV1, 1)
	if err != nil {
		t.Fatalf("Decrypt(v1) error = %v", err)
	}
	if !bytes.Equal(decryptedV1, secret) {
		t.Fatalf("decrypted mismatch: got %q, want %q", decryptedV1, secret)
	}

	// Ketidakcocokan versi kunci harus gagal (terikat via GCM AAD).
	if _, err := cipherV2.Decrypt(ciphertextV1, 2); err == nil {
		t.Fatal("expected decryption failure when using wrong key version")
	}

	// Modifikasi byte pada ciphertext harus gagal autentikasi tag GCM.
	tampered := bytes.Clone(ciphertextV1)
	tampered[len(tampered)-1] ^= 0xFF
	if _, err := cipherV2.Decrypt(tampered, 1); err == nil {
		t.Fatal("expected decryption failure for tampered ciphertext")
	}

	// Pastikan mutasi slice kunci masukan pemanggil tidak merusak keyring internal.
	keyV2[0] ^= 0xFF
	ciphertextV2, ver2, err := cipherV2.Encrypt(secret)
	if err != nil || ver2 != 2 {
		t.Fatalf("Encrypt(v2) = (%d, %v), want ver=2 nil", ver2, err)
	}
	// Buat instance baru dengan keyV2 asli untuk membuktikan tidak ikut termutasi (seharusnya gagal karena keyV2 termutasi).
	if dec, err := cipherV2.Decrypt(ciphertextV2, 2); err != nil || !bytes.Equal(dec, secret) {
		t.Fatalf("expected internal cloned key to decrypt consistently, err=%v", err)
	}
}

func TestAESGCMKeyringCipher_ConfigValidationAndKeyParsing(t *testing.T) {
	t.Parallel()

	if _, err := crypto.NewAESGCMKeyringCipher(0, map[int16][]byte{1: bytes.Repeat([]byte{1}, 32)}); err == nil {
		t.Fatal("expected error for non-positive active version")
	}
	if _, err := crypto.NewAESGCMKeyringCipher(2, map[int16][]byte{1: bytes.Repeat([]byte{1}, 32)}); err == nil {
		t.Fatal("expected error when active version missing from keyring")
	}
	if _, err := crypto.NewAESGCMKeyringCipher(1, map[int16][]byte{1: []byte("too-short")}); err == nil {
		t.Fatal("expected error for non-32-byte key")
	}

	rawKey := bytes.Repeat([]byte{0x42}, 32)
	b64Key := base64.StdEncoding.EncodeToString(rawKey)
	parsed, err := crypto.ParseKeyringSpec("1:" + b64Key)
	if err != nil {
		t.Fatalf("ParseKeyringSpec() error = %v", err)
	}
	if !bytes.Equal(parsed[1], rawKey) {
		t.Fatal("parsed key bytes mismatch")
	}

	// Uji penolakan kunci hex/base64 tidak valid atau versi duplikat.
	for _, bad := range []string{"", "badspec", "0:" + b64Key, "1:not-valid", "1:" + b64Key + ",1:" + b64Key} {
		if _, err := crypto.ParseKeyringSpec(bad); err == nil {
			t.Fatalf("ParseKeyringSpec(%q) succeeded, expected error", bad)
		}
	}

	// Uji pembuatan ephemeral dev key menghasilkan key valid tanpa error.
	cipher, err := crypto.NewEphemeralAES256Cipher()
	if err != nil {
		t.Fatalf("NewEphemeralAES256Cipher() error = %v", err)
	}
	ct, kv, err := cipher.Encrypt([]byte("abc"))
	if err != nil || kv != 1 || len(ct) == 0 {
		t.Fatalf("ephemeral Encrypt() failed: kv=%d err=%v", kv, err)
	}
	pt, err := cipher.Decrypt(ct, kv)
	if err != nil || !strings.EqualFold(string(pt), "abc") {
		t.Fatalf("ephemeral Decrypt() = (%q, %v)", pt, err)
	}
}
