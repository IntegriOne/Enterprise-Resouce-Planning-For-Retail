package pgstore

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type txContextKey struct{}
type transactionHandle struct{ db, owner *gorm.DB }

// Transactor mengelola transaksi lokal HRM; nested transaction memakai savepoint GORM.
type Transactor struct{ db *gorm.DB }

var _ domainrepo.Transactor = (*Transactor)(nil)

// NewTransactor menolak koneksi nil dan dialect selain PostgreSQL.
func NewTransactor(db *gorm.DB) (*Transactor, error) {
	if err := CheckDB(db); err != nil {
		return nil, err
	}
	return &Transactor{db: db}, nil
}

// CheckDB memeriksa handle tanpa membuat koneksi palsu atau mengubah skema.
func CheckDB(db *gorm.DB) error {
	if db == nil || db.Config == nil || db.Dialector == nil {
		return hrerrors.ErrUnavailable
	}
	if db.Name() != "postgres" {
		return fmt.Errorf("%w: postgres dialect required", hrerrors.ErrInvalidInput)
	}
	return nil
}

// DBFromContext memilih transaksi yang dimiliki handle database yang sama.
func DBFromContext(ctx context.Context, fallback *gorm.DB) (*gorm.DB, error) {
	if ctx == nil {
		return nil, hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := CheckDB(fallback); err != nil {
		return nil, err
	}
	if tx, ok := ctx.Value(txContextKey{}).(transactionHandle); ok {
		if tx.owner != fallback {
			return nil, fmt.Errorf("%w: transaction database mismatch", hrerrors.ErrInvalidInput)
		}
		return tx.db.WithContext(ctx), nil
	}
	return fallback.WithContext(ctx), nil
}

// WithinTransaction mengembalikan error callback dan membatalkan seluruh perubahan saat gagal.
func (t *Transactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if t == nil {
		return hrerrors.ErrUnavailable
	}
	if fn == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := DBFromContext(ctx, t.db)
	if err != nil {
		return err
	}
	var callbackErr error
	txErr := db.Transaction(func(tx *gorm.DB) error {
		callbackErr = fn(context.WithValue(ctx, txContextKey{}, transactionHandle{db: tx, owner: t.db}))
		return callbackErr
	})
	if callbackErr != nil {
		return callbackErr
	}
	return MapError(txErr)
}

func requireTransaction(ctx context.Context) error {
	if ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if _, ok := ctx.Value(txContextKey{}).(transactionHandle); !ok {
		return hrerrors.ErrTransactionRequired
	}
	return nil
}
