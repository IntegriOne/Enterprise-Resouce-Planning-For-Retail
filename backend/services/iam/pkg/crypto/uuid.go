package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// UUIDGenerator membangkitkan pengenal unik UUIDv4 acak secara kriptografis.
type UUIDGenerator struct {
	randSource io.Reader
}

// NewUUIDGenerator membentuk UUIDGenerator menggunakan crypto/rand bawaan sistem.
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{randSource: rand.Reader}
}

// NewUUID menghasilkan string UUIDv4 terformat sesuai standar RFC 4122/9562.
func (g *UUIDGenerator) NewUUID() (string, error) {
	if g == nil {
		return "", iamerrors.Wrap(nil, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_uuid_generator"), "uuid generator is nil")
	}
	src := g.randSource
	if src == nil {
		src = rand.Reader
	}

	var buf [16]byte
	if _, err := io.ReadFull(src, buf[:]); err != nil {
		return "", iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.uuid_entropy_failed"), "failed to read random bytes for uuid")
	}

	// Tetapkan bit versi 4 (0100) dan varian RFC 4122 (10xx).
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	var out [36]byte
	hex.Encode(out[0:8], buf[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], buf[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], buf[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], buf[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], buf[10:16])

	return fmt.Sprintf("%s", out[:]), nil
}
