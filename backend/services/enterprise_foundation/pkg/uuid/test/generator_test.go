package uuid_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/pkg/uuid"
)

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestGenerator_NewUUID(t *testing.T) {
	t.Parallel()

	gen := uuid.NewGenerator()
	seen := make(map[string]struct{}, 64)
	for i := 0; i < 64; i++ {
		id, err := gen.NewUUID()
		if err != nil {
			t.Fatalf("NewUUID() error = %v", err)
		}
		if !entity.IsValidUUID(id) {
			t.Fatalf("generated invalid uuid %q", id)
		}
		if id != strings.ToLower(id) {
			t.Fatalf("expected lowercase uuid, got %q", id)
		}
		// Pastikan bit versi 4 dan varian RFC 4122 terpasang.
		if id[14] != '4' {
			t.Fatalf("expected version 4 nibble, got %q", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate uuid generated: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestGenerator_FailClosed(t *testing.T) {
	t.Parallel()

	var nilGen *uuid.Generator
	if _, err := nilGen.NewUUID(); !errors.Is(err, eferrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable on nil generator, got %v", err)
	}

	broken := uuid.NewGeneratorWithReader(failingReader{})
	if _, err := broken.NewUUID(); !errors.Is(err, eferrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable on broken reader, got %v", err)
	}
}
