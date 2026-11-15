package crypto_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
)

func TestUUIDGenerator_NewUUID(t *testing.T) {
	t.Parallel()

	// Pastikan implementasi memenuhi kontrak port domainservice.IDGenerator.
	var gen domainservice.IDGenerator = crypto.NewUUIDGenerator()

	seen := make(map[string]struct{}, 64)
	for i := 0; i < 64; i++ {
		id, err := gen.NewUUID()
		if err != nil {
			t.Fatalf("NewUUID() error = %v", err)
		}
		if !entity.IsValidUUID(id) {
			t.Fatalf("NewUUID() produced invalid uuid %q", id)
		}
		// Pastikan bit versi UUIDv4 dan varian RFC 4122 terpasang benar.
		if id[14] != '4' {
			t.Fatalf("expected UUID version 4 at index 14, got %q (%q)", id[14], id)
		}
		if id[19] != '8' && id[19] != '9' && id[19] != 'a' && id[19] != 'b' {
			t.Fatalf("expected RFC 4122 variant at index 19, got %q (%q)", id[19], id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate uuid generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}
