package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strconv"
	"strings"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

var _ domainservice.CSRFDeriver = (*HMACCSRFDeriver)(nil)

// HMACCSRFDeriver menurunkan token CSRF deterministik per sesi agar aman dipakai serentak di banyak tab browser.
type HMACCSRFDeriver struct {
	activeVersion int
	keyring       map[int][]byte
}

// NewHMACCSRFDeriver membentuk pembangkit CSRF berbasis HMAC-SHA256 terversi.
func NewHMACCSRFDeriver(activeVersion int, keyring map[int][]byte, allowEphemeralDev bool) (*HMACCSRFDeriver, error) {
	if activeVersion <= 0 {
		activeVersion = 1
	}
	copied := make(map[int][]byte, len(keyring))
	for ver, key := range keyring {
		if ver > 0 && len(key) >= 16 {
			k := make([]byte, len(key))
			copy(k, key)
			copied[ver] = k
		}
	}
	if _, ok := copied[activeVersion]; !ok {
		if !allowEphemeralDev {
			return nil, iamerrors.NewValidation("csrf_keyring", iamerrors.Code("iam.crypto.missing_csrf_key"), "active CSRF HMAC key is required")
		}
		ephemeral := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, ephemeral); err != nil {
			return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.csrf_entropy_failed"), "failed to generate ephemeral CSRF key")
		}
		copied[activeVersion] = ephemeral
	}
	return &HMACCSRFDeriver{
		activeVersion: activeVersion,
		keyring:       copied,
	}, nil
}

// ActiveKeyVersion mengembalikan versi kunci HMAC CSRF aktif untuk penerbitan sesi baru.
func (d *HMACCSRFDeriver) ActiveKeyVersion() int {
	if d == nil || d.activeVersion <= 0 {
		return 1
	}
	return d.activeVersion
}

// DeriveToken menghitung pasangan token CSRF plaintext URL-safe dan digest SHA-256 yang terikat pada sesi dan generasinya.
func (d *HMACCSRFDeriver) DeriveToken(keyVersion int, sessionID string, tokenDigest []byte, generation int64) (string, []byte, error) {
	if d == nil {
		return "", nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_csrf_deriver"), "csrf deriver is nil")
	}
	if keyVersion <= 0 {
		keyVersion = d.activeVersion
	}
	key, ok := d.keyring[keyVersion]
	if !ok || len(key) == 0 {
		return "", nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.unknown_csrf_key_version"), "unknown CSRF key version")
	}
	if strings.TrimSpace(sessionID) == "" || len(tokenDigest) == 0 {
		return "", nil, iamerrors.NewValidation("session", iamerrors.Code("iam.crypto.invalid_csrf_binding"), "session ID and token digest are required")
	}
	if generation <= 0 {
		generation = 1
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("ig-csrf-v1|"))
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(sessionID))))
	mac.Write([]byte("|"))
	mac.Write([]byte(hex.EncodeToString(tokenDigest)))
	mac.Write([]byte("|"))
	mac.Write([]byte(strconv.FormatInt(generation, 10)))

	plaintext := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	sum := sha256.Sum256([]byte(plaintext))
	digest := make([]byte, len(sum))
	copy(digest, sum[:])
	return plaintext, digest, nil
}
