package pgstore_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func TestQueryHelpers(t *testing.T) {
	t.Parallel()

	t.Run("EscapeLike and LikePattern escape wildcards safely", func(t *testing.T) {
		t.Parallel()
		if got := pgstore.EscapeLike(`100%_test\path`); got != `100\%\_test\\path` {
			t.Fatalf("unexpected EscapeLike output: %q", got)
		}
		if got := pgstore.LikePattern("  PT_Maju%  "); got != `%PT\_Maju\%%` {
			t.Fatalf("unexpected LikePattern output: %q", got)
		}
	})

	t.Run("UUID normalization and deduplication", func(t *testing.T) {
		t.Parallel()
		uid, err := pgstore.ValidateUUIDParam("id", "  AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uid != "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee" {
			t.Fatalf("unexpected normalized uuid: %q", uid)
		}

		if _, err := pgstore.ValidateUUIDParam("id", "not-a-uuid"); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Fatalf("expected KindInvalidInput, got %v", err)
		}

		ids, err := pgstore.NormalizeUUIDSlice("ids", []string{
			"AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE",
			"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
			"11111111-2222-4333-8444-555555555555",
		})
		if err != nil {
			t.Fatalf("unexpected slice error: %v", err)
		}
		if len(ids) != 2 || ids[0] != "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee" || ids[1] != "11111111-2222-4333-8444-555555555555" {
			t.Fatalf("unexpected deduplicated slice: %v", ids)
		}

		if _, err := pgstore.NormalizeUUIDSlice("ids", []string{"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", "bad"}); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Fatalf("expected invalid slice element error, got %v", err)
		}
	})

	t.Run("time and pointer cloning helpers", func(t *testing.T) {
		t.Parallel()
		if !pgstore.NormalizeLocalTime(time.Time{}).IsZero() {
			t.Fatal("expected zero time to stay zero")
		}
		utc := time.Date(2026, 11, 15, 10, 30, 0, 0, time.UTC)
		if pgstore.NormalizeLocalTime(utc).Location() != time.Local {
			t.Fatalf("expected time.Local location")
		}
		if d := pgstore.NormalizeDatePtr(&utc); d == nil || d.Hour() != 0 || d.Minute() != 0 || d.Location() != time.UTC {
			t.Fatalf("unexpected normalized calendar date: %v", d)
		}
		if pgstore.MustLocalNow(time.Time{}).IsZero() {
			t.Fatal("expected non-zero current time")
		}

		if ptr := pgstore.NullableStringPtr("   "); ptr != nil {
			t.Fatalf("expected nil for blank string, got %v", *ptr)
		}
		orig := "hello"
		cp := pgstore.CloneStringPtr(&orig)
		orig = "mutated"
		if cp == nil || *cp != "hello" {
			t.Fatalf("expected cloned string to be isolated, got %v", cp)
		}

		f := 1.25
		fp := pgstore.CloneFloat64Ptr(&f)
		f = 9.99
		if fp == nil || *fp != 1.25 {
			t.Fatalf("expected cloned float to be isolated, got %v", fp)
		}

		raw := []byte(`{"a":1}`)
		cloned := pgstore.CloneBytes(raw)
		raw[0] = 'X'
		if !bytes.Equal(cloned, []byte(`{"a":1}`)) {
			t.Fatalf("expected cloned bytes isolated, got %s", string(cloned))
		}
	})
}

func TestTransactorGuardsAndHooks(t *testing.T) {
	t.Parallel()

	t.Run("nil db fails closed", func(t *testing.T) {
		t.Parallel()
		if err := pgstore.CheckDB(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Fatalf("expected KindUnavailable, got %v", err)
		}
		if _, err := pgstore.NewTransactor(nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Fatalf("expected KindUnavailable from NewTransactor(nil), got %v", err)
		}
		if _, err := pgstore.DBFromContext(context.Background(), nil); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Fatalf("expected KindUnavailable from DBFromContext(nil), got %v", err)
		}
		var nilTx *pgstore.Transactor
		if err := nilTx.WithinTransaction(context.Background(), func(context.Context) error { return nil }); eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Fatalf("expected KindUnavailable on nil transactor, got %v", err)
		}
		zeroTx := &pgstore.Transactor{}
		if err := zeroTx.WithinTransaction(context.Background(), nil); eferrors.KindOf(err) != eferrors.KindInvalidInput {
			t.Fatalf("expected KindInvalidInput on nil callback, got %v", err)
		}
	})

	t.Run("RegisterAfterCommit executes immediately outside transaction", func(t *testing.T) {
		t.Parallel()
		called := false
		pgstore.RegisterAfterCommit(context.Background(), func(ctx context.Context) {
			if ctx == nil {
				t.Fatal("expected non-nil ctx")
			}
			called = true
		})
		if !called {
			t.Fatal("expected hook to execute immediately outside transaction")
		}
		// Memastikan input nil aman tanpa panic.
		pgstore.RegisterAfterCommit(nil, nil)
		pgstore.RegisterAfterCommit(nil, func(ctx context.Context) {
			if ctx == nil {
				t.Fatal("expected non-nil background context")
			}
		})
	})
}
