package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"io"
	"strconv"
	"strings"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const aes256KeyBytes = 32

// AESGCMKeyringCipher mengimplementasikan domainservice.SecretCipher dengan rotasi key_version AES-256-GCM.
type AESGCMKeyringCipher struct {
	activeVersion int16
	keys          map[int16][]byte
	randSource    io.Reader
}

// NewAESGCMKeyringCipher membentuk enkriptor keyring AES-256-GCM dengan validasi versi aktif dan panjang kunci 32 byte.
func NewAESGCMKeyringCipher(activeVersion int16, keys map[int16][]byte) (*AESGCMKeyringCipher, error) {
	if activeVersion <= 0 {
		return nil, iamerrors.NewValidation("active_key_version", iamerrors.Code("iam.crypto.invalid_key_version"), "active key version must be positive")
	}
	if len(keys) == 0 {
		return nil, iamerrors.NewValidation("keys", iamerrors.Code("iam.crypto.empty_keyring"), "keyring must contain at least one aes-256 key")
	}

	clonedKeys := make(map[int16][]byte, len(keys))
	for ver, raw := range keys {
		if ver <= 0 {
			return nil, iamerrors.NewValidation("key_version", iamerrors.Code("iam.crypto.invalid_key_version"), "all key versions must be positive")
		}
		if len(raw) != aes256KeyBytes {
			return nil, iamerrors.NewValidation("key", iamerrors.Code("iam.crypto.invalid_key_length"), "aes-256 key must be exactly 32 bytes")
		}
		clonedKeys[ver] = bytes.Clone(raw)
	}
	if _, ok := clonedKeys[activeVersion]; !ok {
		return nil, iamerrors.NewValidation("active_key_version", iamerrors.Code("iam.crypto.missing_active_key"), "active key version is missing from keyring")
	}

	return &AESGCMKeyringCipher{
		activeVersion: activeVersion,
		keys:          clonedKeys,
		randSource:    rand.Reader,
	}, nil
}

// NewEphemeralAES256Cipher menghasilkan kunci sementara sekali jalan khusus mode pengembangan lokal.
func NewEphemeralAES256Cipher() (*AESGCMKeyringCipher, error) {
	key := make([]byte, aes256KeyBytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.ephemeral_key_failed"), "failed to generate ephemeral aes-256 key")
	}
	return NewAESGCMKeyringCipher(1, map[int16][]byte{1: key})
}

// ParseKeyringSpec mem-parsing spesifikasi env berbentuk "<ver>:<base64-atau-hex>,...".
func ParseKeyringSpec(spec string) (map[int16][]byte, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.empty_keyring_spec"), "keyring specification must not be empty")
	}

	entries := strings.Split(trimmed, ",")
	keys := make(map[int16][]byte, len(entries))
	for _, entry := range entries {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
		if len(parts) != 2 {
			return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.invalid_keyring_entry"), "keyring entry must use <version>:<key> format")
		}

		verInt, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 16)
		if err != nil || verInt <= 0 {
			return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.invalid_key_version"), "keyring version must be a positive int16")
		}
		ver := int16(verInt)
		if _, dup := keys[ver]; dup {
			return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.duplicate_key_version"), "duplicate key version in keyring")
		}

		rawKey, err := decode32ByteKey(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		keys[ver] = rawKey
	}
	return keys, nil
}

func decode32ByteKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.empty_key_material"), "key material must not be empty")
	}
	if len(raw) == aes256KeyBytes*2 {
		if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == aes256KeyBytes {
			return decoded, nil
		}
	}
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, err := enc.DecodeString(raw); err == nil && len(decoded) == aes256KeyBytes {
			return decoded, nil
		}
	}
	return nil, iamerrors.NewValidation("totp_keyring", iamerrors.Code("iam.crypto.invalid_key_material"), "key material must decode to 32 bytes in base64 or hex")
}

// Encrypt mengenkripsi secret mentah dengan kunci aktif dan mengikat key_version pada AAD GCM.
func (c *AESGCMKeyringCipher) Encrypt(plaintext []byte) ([]byte, int16, error) {
	if c == nil {
		return nil, 0, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_cipher"), "secret cipher is nil")
	}
	if len(plaintext) == 0 {
		return nil, 0, iamerrors.NewValidation("plaintext", iamerrors.Code("iam.crypto.empty_plaintext"), "plaintext secret must not be empty")
	}

	key, ok := c.keys[c.activeVersion]
	if !ok {
		return nil, 0, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.missing_active_key"), "active encryption key not available")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, 0, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.aes_init_failed"), "failed to initialize aes cipher")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, 0, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.gcm_init_failed"), "failed to initialize gcm")
	}

	src := c.randSource
	if src == nil {
		src = rand.Reader
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(src, nonce); err != nil {
		return nil, 0, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nonce_entropy_failed"), "failed to generate gcm nonce")
	}

	var aad [2]byte
	binary.BigEndian.PutUint16(aad[:], uint16(c.activeVersion))
	sealed := gcm.Seal(nonce, nonce, plaintext, aad[:])
	return sealed, c.activeVersion, nil
}

// Decrypt mendekripsi ciphertext berdasarkan key_version dan memverifikasi integritas tag serta AAD.
func (c *AESGCMKeyringCipher) Decrypt(ciphertext []byte, keyVersion int16) ([]byte, error) {
	if c == nil {
		return nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_cipher"), "secret cipher is nil")
	}
	key, ok := c.keys[keyVersion]
	if !ok {
		return nil, iamerrors.New(iamerrors.KindUnauthorized, iamerrors.Code("iam.crypto.unknown_key_version"), "unknown totp encryption key version")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.aes_init_failed"), "failed to initialize aes cipher")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.gcm_init_failed"), "failed to initialize gcm")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize+gcm.Overhead() {
		return nil, iamerrors.NewValidation("ciphertext", iamerrors.Code("iam.crypto.malformed_ciphertext"), "ciphertext shorter than nonce and tag overhead")
	}

	nonce := ciphertext[:nonceSize]
	payload := ciphertext[nonceSize:]
	var aad [2]byte
	binary.BigEndian.PutUint16(aad[:], uint16(keyVersion))

	plaintext, err := gcm.Open(nil, nonce, payload, aad[:])
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnauthorized, iamerrors.Code("iam.crypto.decryption_failed"), "failed to decrypt protected secret")
	}
	return bytes.Clone(plaintext), nil
}
