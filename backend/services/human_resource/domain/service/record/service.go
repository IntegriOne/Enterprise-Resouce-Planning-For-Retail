// Package service menghubungkan validasi snapshot domain dengan port write repository.
package service

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// WritePort membatasi dependency pada operasi persistensi yang dipakai validator.
type WritePort[T any] interface {
	Insert(context.Context, *T) error
	Update(context.Context, *T) error
}

// ModelPointer mengikat validator ke pointer entity yang persis sama dengan port repository.
type ModelPointer[T any] interface {
	*T
	entity.Validatable
}

// Writer memvalidasi snapshot sebelum meneruskannya ke repository.
// Invariant lintas-record, IAM, transaksi dan transisi tetap diperiksa use case.
type Writer[T any, P ModelPointer[T]] struct{ port WritePort[T] }

// NewWriter menolak dependency nil termasuk typed nil tanpa membuka koneksi database.
func NewWriter[T any, P ModelPointer[T]](port WritePort[T]) (*Writer[T, P], error) {
	if port == nil {
		return nil, hrerrors.ErrUnavailable
	}
	value := reflect.ValueOf(port)
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		if value.IsNil() {
			return nil, hrerrors.ErrUnavailable
		}
	}
	return &Writer[T, P]{port: port}, nil
}

func (w *Writer[T, P]) validate(ctx context.Context, value *T) error {
	if w == nil || w.port == nil {
		return hrerrors.ErrUnavailable
	}
	if ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return P(value).Validate()
}

// Insert hanya memanggil port setelah seluruh validasi snapshot berhasil.
func (w *Writer[T, P]) Insert(ctx context.Context, value *T) error {
	if err := w.validate(ctx, value); err != nil {
		return err
	}
	return w.port.Insert(ctx, value)
}

// Create adalah alias Insert untuk pemanggil domain.
func (w *Writer[T, P]) Create(ctx context.Context, value *T) error { return w.Insert(ctx, value) }

// Update meneruskan snapshot lengkap tervalidasi dan tidak mengganti error port.
func (w *Writer[T, P]) Update(ctx context.Context, value *T) error {
	if err := w.validate(ctx, value); err != nil {
		return err
	}
	return w.port.Update(ctx, value)
}
