package db

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestReconciliationTargetsVerifiedBaseline(t *testing.T) {
	digest := sha256.Sum256([]byte(baselineSQL))
	if actual := hex.EncodeToString(digest[:]); actual != runtimeBaselineChecksum {
		t.Fatalf("baseline checksum changed from %s to %s", runtimeBaselineChecksum, actual)
	}

	offset := -1
	for _, marker := range []string{"-- Assignment cancellation\n", "-- Worker principal uniqueness\n", "-- HRM runtime\n"} {
		index := strings.Index(baselineReconciliationSQL, marker)
		if index <= offset || strings.Count(baselineReconciliationSQL, marker) != 1 {
			t.Fatalf("invalid reconciliation section %q", marker)
		}
		offset = index
	}
}
