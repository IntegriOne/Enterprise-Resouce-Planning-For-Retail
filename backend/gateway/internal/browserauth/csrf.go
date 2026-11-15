package browserauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/integrione/erp-retail/backend/shared/authsession"
)

const randomTokenByteLength = 32

// GenerateOpaquePair menghasilkan token buram acak URL-safe dan digest SHA-256 hex lowercase-nya.
func GenerateOpaquePair() (plaintext string, digestHex string, err error) {
	buf := make([]byte, randomTokenByteLength)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", "", fmt.Errorf("generate csprng token: %w", err)
	}
	plaintext = base64.RawURLEncoding.EncodeToString(buf)
	digestHex = authsession.HashOpaqueToken(plaintext)
	return plaintext, digestHex, nil
}

// VerifyCSRFTokenDigest membandingkan digest SHA-256 token CSRF plaintext dengan expectedDigestHex secara constant-time.
func VerifyCSRFTokenDigest(plaintext, expectedDigestHex string) bool {
	trimmedToken := strings.TrimSpace(plaintext)
	expectedHex := strings.ToLower(strings.TrimSpace(expectedDigestHex))
	if trimmedToken == "" || len(expectedHex) != 64 {
		return false
	}
	expectedBytes, err := hex.DecodeString(expectedHex)
	if err != nil || len(expectedBytes) != 32 {
		return false
	}
	actualHex := authsession.HashOpaqueToken(trimmedToken)
	actualBytes, err := hex.DecodeString(actualHex)
	if err != nil || len(actualBytes) != 32 {
		return false
	}
	return subtle.ConstantTimeCompare(actualBytes, expectedBytes) == 1
}
