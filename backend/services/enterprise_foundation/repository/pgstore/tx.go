// Package pgstore menyediakan utilitas persistensi PostgreSQL berbasis GORM dan pemetaan error domain untuk repositori Enterprise Foundation.
package pgstore

import (
	"context"
	"errors"

	"gorm.io/gorm"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
)

type txContextKey struct{}
type afterCommitKey struct{}

type afterCommitCollector struct {
	hooks []func(context.Context)
}

// RegisterAfterCommit mendaftarkan callback pasca-commit transaksi PG atau mengeksekusi langsung bila tidak sedang di dalam transaksi.
func RegisterAfterCommit(ctx context.Context, fn func(context.Context)) {
	if fn == nil {
		return
	}
	if ctx == nil {
		fn(context.Background())
		return
	}
	if collector, ok := ctx.Value(afterCommitKey{}).(*afterCommitCollector); ok && collector != nil {
		collector.hooks = append(collector.hooks, fn)
		return
	}
	fn(ctx)
}

// Memastikan Transactor memenuhi kontrak port domain.
var _ domainrepo.Transactor = (*Transactor)(nil)

// Transactor mengelola eksekusi transaksi atomik PostgreSQL melalui context propagation.
type Transactor struct {
	db *gorm.DB
}

// NewTransactor membangun Transactor baru dengan validasi fail-closed terhadap koneksi database nil.
func NewTransactor(db *gorm.DB) (*Transactor, error) {
	if err := CheckDB(db); err != nil {
		return nil, err
	}
	return &Transactor{db: db}, nil
}

// WithinTransaction menjalankan callback dalam satu transaksi PostgreSQL atau savepoint bila sudah berada di dalam transaksi.
func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if t == nil {
		return eferrors.Wrap(errors.New("transactor is nil"), eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database transactor is unavailable")
	}
	if fn == nil {
		return eferrors.NewValidation("fn", eferrors.Code("enterprise_foundation.invalid_transaction_callback"), "transaction callback must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	baseDB, err := DBFromContext(ctx, t.db)
	if err != nil {
		return err
	}

	isOutermost := ctx.Value(afterCommitKey{}) == nil
	var collector *afterCommitCollector
	baseCtx := ctx
	if isOutermost {
		collector = &afterCommitCollector{}
		baseCtx = context.WithValue(ctx, afterCommitKey{}, collector)
	}

	txErr := baseDB.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(baseCtx, txContextKey{}, tx)
		return fn(txCtx)
	})
	if txErr != nil {
		return MapError(txErr, nil)
	}

	if isOutermost && collector != nil && len(collector.hooks) > 0 {
		// Gunakan context induk non-transaksi tanpa txContextKey agar publikasi outbox atau cache pasca-commit berjalan di luar transaksi.
		postCtx := context.WithoutCancel(ctx)
		for _, hook := range collector.hooks {
			if hook != nil {
				hook(postCtx)
			}
		}
	}
	return nil
}

// CheckDB memastikan pointer GORM DB tersedia sebelum operasi persistensi dijalankan.
func CheckDB(db *gorm.DB) error {
	if db == nil {
		return eferrors.Wrap(errors.New("gorm db is nil"), eferrors.KindUnavailable, eferrors.ErrUnavailable.Code, "database connection is unavailable")
	}
	return nil
}

// DBFromContext mengambil handle transaksi aktif dari context atau memakai koneksi fallback bila tidak ada transaksi berjalan.
func DBFromContext(ctx context.Context, fallback *gorm.DB) (*gorm.DB, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx), nil
	}
	if err := CheckDB(fallback); err != nil {
		return nil, err
	}
	return fallback.WithContext(ctx), nil
}
