package repository

import "context"

// Transactor mengeksekusi unit kerja atomik di dalam satu transaksi database tanpa membocorkan detail SQL ke domain.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
