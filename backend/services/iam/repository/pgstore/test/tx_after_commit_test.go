package pgstore_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAfterCommitDiscardsRolledBackSavepointHooks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	tx, err := pgstore.NewTransactor(db)
	if err != nil {
		t.Fatalf("new transactor: %v", err)
	}

	var called []string
	rollbackErr := errors.New("rollback savepoint")
	err = tx.WithinTransaction(context.Background(), func(outer context.Context) error {
		pgstore.RegisterAfterCommit(outer, func(context.Context) { called = append(called, "outer-before") })
		if err := tx.WithinTransaction(outer, func(inner context.Context) error {
			pgstore.RegisterAfterCommit(inner, func(context.Context) { called = append(called, "rolled-back") })
			return rollbackErr
		}); !errors.Is(err, rollbackErr) {
			t.Fatalf("nested transaction error = %v", err)
		}
		if err := tx.WithinTransaction(outer, func(inner context.Context) error {
			pgstore.RegisterAfterCommit(inner, func(context.Context) { called = append(called, "committed") })
			return nil
		}); err != nil {
			return err
		}
		pgstore.RegisterAfterCommit(outer, func(context.Context) { called = append(called, "outer-after") })
		return nil
	})
	if err != nil {
		t.Fatalf("outer transaction: %v", err)
	}
	if want := []string{"outer-before", "committed", "outer-after"}; !slices.Equal(called, want) {
		t.Fatalf("called hooks = %v, want %v", called, want)
	}
}
