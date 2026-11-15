package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	minOpaqueTokenBytes = 16
	maxOpaqueTokenBytes = 256
	maxRecoveryCodes    = 32
	// Alfabet 32 karakter unambiguous membagi rata 256 sehingga bebas modulo bias.
	recoveryCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	recoveryCodeGroupLen = 4
	recoveryCodeGroups   = 3
)

// OpaqueTokenGenerator mengimplementasikan domainservice.TokenGenerator memakai CSPRNG dan SHA-256.
type OpaqueTokenGenerator struct {
	randSource io.Reader
}

// NewOpaqueTokenGenerator membentuk pembangkit token buram dan kode pemulihan 2FA.
func NewOpaqueTokenGenerator() *OpaqueTokenGenerator {
	return &OpaqueTokenGenerator{randSource: rand.Reader}
}

// GenerateOpaqueToken menghasilkan string token acak URL-safe beserta digest SHA-256 sepanjang 32 byte.
func (g *OpaqueTokenGenerator) GenerateOpaqueToken(byteLen int) (string, []byte, error) {
	if g == nil {
		return "", nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_token_generator"), "token generator is nil")
	}
	if byteLen < minOpaqueTokenBytes || byteLen > maxOpaqueTokenBytes {
		return "", nil, iamerrors.NewValidation("byte_len", iamerrors.Code("iam.crypto.invalid_token_length"), "token byte length must be between 16 and 256")
	}

	src := g.randSource
	if src == nil {
		src = rand.Reader
	}
	raw := make([]byte, byteLen)
	if _, err := io.ReadFull(src, raw); err != nil {
		return "", nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.token_entropy_failed"), "failed to generate opaque token bytes")
	}

	plaintext := base64.RawURLEncoding.EncodeToString(raw)
	digest := g.Digest(plaintext)
	return plaintext, digest, nil
}

// Digest menghitung hash SHA-256 atas token polos, atau nil jika token kosong.
func (g *OpaqueTokenGenerator) Digest(plaintext string) []byte {
	if strings.TrimSpace(plaintext) == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(plaintext))
	out := make([]byte, len(sum))
	copy(out, sum[:])
	return out
}

// GenerateRecoveryCodes menghasilkan daftar kode cadangan 2FA unik berformat XXXX-XXXX-XXXX.
func (g *OpaqueTokenGenerator) GenerateRecoveryCodes(count int) ([]string, error) {
	if g == nil {
		return nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_token_generator"), "token generator is nil")
	}
	if count <= 0 || count > maxRecoveryCodes {
		return nil, iamerrors.NewValidation("count", iamerrors.Code("iam.crypto.invalid_recovery_code_count"), "recovery code count must be between 1 and 32")
	}

	src := g.randSource
	if src == nil {
		src = rand.Reader
	}

	codes := make([]string, 0, count)
	seen := make(map[string]struct{}, count)
	const charCount = recoveryCodeGroupLen * recoveryCodeGroups
	buf := make([]byte, charCount)

	for len(codes) < count {
		if _, err := io.ReadFull(src, buf); err != nil {
			return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.recovery_entropy_failed"), "failed to generate recovery code bytes")
		}

		var b strings.Builder
		b.Grow(charCount + recoveryCodeGroups - 1)
		for i, v := range buf {
			if i > 0 && i%recoveryCodeGroupLen == 0 {
				b.WriteByte('-')
			}
			b.WriteByte(recoveryCodeAlphabet[v&0x1F])
		}
		code := b.String()
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}

	return codes, nil
}
