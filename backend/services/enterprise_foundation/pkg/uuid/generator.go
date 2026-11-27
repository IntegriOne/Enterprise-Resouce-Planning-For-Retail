// Package uuid menyediakan generator pengenal unik UUIDv4 berbasis crypto/rand.
package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Generator membangkitkan string UUIDv4 acak secara kriptografis yang memenuhi standar RFC 4122/9562.
type Generator struct {
	randSource io.Reader
}

// NewGenerator membentuk Generator UUIDv4 menggunakan crypto/rand bawaan sistem.
func NewGenerator() *Generator {
	return &Generator{randSource: rand.Reader}
}

// NewGeneratorWithReader membentuk Generator dengan sumber entropi khusus untuk kebutuhan pengujian deterministik.
func NewGeneratorWithReader(src io.Reader) *Generator {
	return &Generator{randSource: src}
}

// NewUUID menghasilkan string UUIDv4 huruf kecil berformat kanonik 36 karakter.
func (g *Generator) NewUUID() (string, error) {
	if g == nil {
		return "", eferrors.Wrap(
			eferrors.ErrUnavailable,
			eferrors.KindUnavailable,
			eferrors.Code("enterprise_foundation.uuid.nil_generator"),
			"uuid generator is nil",
		)
	}
	src := g.randSource
	if src == nil {
		src = rand.Reader
	}

	var buf [16]byte
	if _, err := io.ReadFull(src, buf[:]); err != nil {
		return "", eferrors.Wrap(
			err,
			eferrors.KindUnavailable,
			eferrors.Code("enterprise_foundation.uuid.entropy_failed"),
			"failed to read random bytes for uuid",
		)
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
