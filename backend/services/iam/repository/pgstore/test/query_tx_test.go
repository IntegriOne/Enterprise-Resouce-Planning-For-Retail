package pgstore_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func TestLikeAndIPHelpers(t *testing.T) {
	t.Parallel()

	escaped := pgstore.EscapeLike(`a%b_c\d`)
	if escaped != `a\%b\_c\\d` {
		t.Fatalf("unexpected escaped string: %q", escaped)
	}
	if got := pgstore.LikePattern("  john_doe% "); got != `%john\_doe\%%` {
		t.Fatalf("unexpected like pattern: %q", got)
	}

	ipTests := []struct {
		raw     string
		want    string
		wantNil bool
	}{
		{raw: "192.168.1.50", want: "192.168.1.50"},
		{raw: "192.168.1.50/32", want: "192.168.1.50"},
		{raw: "::ffff:10.0.0.5", want: "10.0.0.5"},
		{raw: "2001:db8::1/128", want: "2001:db8::1"},
		{raw: "", wantNil: true},
		{raw: "not-an-ip", wantNil: true},
	}

	for _, tc := range ipTests {
		addr := pgstore.ParseIPAddress(tc.raw)
		if tc.wantNil {
			if addr != nil {
				t.Fatalf("expected nil for %q, got %v", tc.raw, addr)
			}
			continue
		}
		if addr == nil || addr.String() != tc.want {
			t.Fatalf("expected %q for %q, got %v", tc.want, tc.raw, addr)
		}
	}

	parsed := netip.MustParseAddr("::ffff:127.0.0.1")
	formatted := pgstore.FormatIPAddressPtr(&parsed)
	if formatted == nil || *formatted != "127.0.0.1" {
		t.Fatalf("expected unmapped 127.0.0.1, got %v", formatted)
	}
	if pgstore.FormatIPAddressPtr(nil) != nil {
		t.Fatal("expected nil formatted IP for nil input")
	}
}

func TestLocalTimeAndTransactorFailClosed(t *testing.T) {
	t.Parallel()

	otherZone := time.FixedZone("TEST", -4*3600)
	src := time.Date(2026, 11, 14, 10, 30, 0, 0, otherZone)

	norm := pgstore.NormalizeLocalTime(src)
	if norm.Location() != time.Local {
		t.Fatalf("expected time.Local, got %v", norm.Location())
	}
	if !norm.Equal(src) {
		t.Fatal("normalized time instant changed")
	}
	if !pgstore.NormalizeLocalTime(time.Time{}).IsZero() {
		t.Fatal("expected zero time to stay zero")
	}
	if pgstore.NormalizeLocalTimePtr(nil) != nil {
		t.Fatal("expected nil ptr for nil time")
	}
	if pgstore.MustLocalNow(time.Time{}).Location() != time.Local {
		t.Fatal("expected local timezone for MustLocalNow")
	}

	if err := pgstore.CheckDB(nil); !errors.Is(err, iamerrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable for nil DB, got %v", err)
	}
	if _, err := pgstore.DBFromContext(context.Background(), nil); !errors.Is(err, iamerrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable from DBFromContext(nil), got %v", err)
	}
	if _, err := pgstore.NewTransactor(nil); !errors.Is(err, iamerrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable from NewTransactor(nil), got %v", err)
	}

	var nilTx *pgstore.Transactor
	if err := nilTx.WithinTransaction(context.Background(), func(context.Context) error { return nil }); !errors.Is(err, iamerrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable for nil Transactor, got %v", err)
	}

	id, err := pgstore.ValidateUUIDParam("user_id", "  550E8400-E29B-41D4-A716-446655440000 ")
	if err != nil || id != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected uuid normalization: id=%q err=%v", id, err)
	}
	if _, err := pgstore.ValidateUUIDParam("user_id", "invalid"); !errors.Is(err, iamerrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for bad uuid, got %v", err)
	}

	globalScope, err := pgstore.ToScope("global", nil)
	if err != nil || globalScope.Type != entity.ScopeTypeGlobal || globalScope.ID != nil {
		t.Fatalf("unexpected global scope: %+v err=%v", globalScope, err)
	}
	facilityID := "550e8400-e29b-41d4-a716-446655440000"
	facilityScope, err := pgstore.ToScope("facility", &facilityID)
	if err != nil || facilityScope.Type != entity.ScopeTypeFacility || facilityScope.ID == nil || *facilityScope.ID != facilityID {
		t.Fatalf("unexpected facility scope: %+v err=%v", facilityScope, err)
	}
	if _, err := pgstore.ToScope("invalid_scope", nil); !errors.Is(err, iamerrors.ErrInvalidScope) {
		t.Fatalf("expected ErrInvalidScope, got %v", err)
	}
	if pgstore.NullableStringPtr("   ") != nil || pgstore.DerefString(nil) != "" {
		t.Fatal("unexpected nullable string helper behavior")
	}
}
