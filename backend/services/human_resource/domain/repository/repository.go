// Package repository mendefinisikan port persistensi HRM tanpa ketergantungan transport.
package repository

import "context"

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// RecordRepository adalah kontrak persistensi minimum milik domain yang dipakai service.
// Implementasi dan mock menerima transaksi melalui context, bukan callback operasi per record.
type RecordRepository[T any, K any] interface {
	Insert(context.Context, *T) error
	SelectByID(context.Context, K) (*T, error)
	FindByIDForUpdate(context.Context, K) (*T, error)
	SelectWhere(ctx context.Context, filters []Filter, limit, offset int) ([]T, error)
	CountWhere(ctx context.Context, filters []Filter) (int64, error)
	Update(context.Context, *T) error
	DeleteByID(context.Context, K) error
}

// Repository berisi operasi baris lengkap, bukan keputusan bisnis atau otorisasi.
// Update mengganti seluruh kolom non-primary-key, termasuk nol dan NULL.
// Tabel immutable tetap menolak Update/Delete sesuai constraint PostgreSQL.
type Repository[T any, K any] interface {
	RecordRepository[T, K]
	Create(context.Context, *T) error
	FindByID(context.Context, K) (*T, error)
	GetAll(ctx context.Context, limit, offset int) ([]T, error)
	SelectList(ctx context.Context, limit, offset int) ([]T, error)
	Count(context.Context) (int64, error)
	Delete(context.Context, K) error
}

// Transactor membagikan transaksi ke semua repository HRM melalui context.
type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
