// Package pgstore menyediakan operasi GORM bersama untuk tabel milik HRM.
package pgstore

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Repository menggunakan metadata model statis; input nilai selalu menjadi parameter SQL.
type Repository[T any, K any] struct {
	db        *gorm.DB
	model     *schema.Schema
	keyValues func(K) []any
}

// New memvalidasi model HRM dan primary key tanpa menjalankan AutoMigrate.
func New[T any, K any](db *gorm.DB, keyValues func(K) []any) (*Repository[T, K], error) {
	if err := CheckDB(db); err != nil {
		return nil, err
	}
	if keyValues == nil {
		return nil, hrerrors.ErrInvalidInput
	}
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, MapError(err)
	}
	if !strings.HasPrefix(stmt.Schema.Table, "hr_") || len(stmt.Schema.PrimaryFields) == 0 {
		return nil, fmt.Errorf("%w: hrm model and primary key required", hrerrors.ErrInvalidInput)
	}
	return &Repository[T, K]{db: db, model: stmt.Schema, keyValues: keyValues}, nil
}

// NormalizePage menetapkan default dan menolak pagination negatif atau tanpa batas.
func NormalizePage(limit, offset int) (int, int, error) {
	if limit < 0 || offset < 0 {
		return 0, 0, hrerrors.ErrInvalidInput
	}
	if limit == 0 {
		limit = domainrepo.DefaultLimit
	}
	if limit > domainrepo.MaxLimit {
		limit = domainrepo.MaxLimit
	}
	return limit, offset, nil
}

func (r *Repository[T, K]) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, hrerrors.ErrUnavailable
	}
	db, err := DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return db.Model(new(T)), nil
}

func (r *Repository[T, K]) keyPredicate(values []any) (clause.Where, error) {
	if len(values) != len(r.model.PrimaryFields) {
		return clause.Where{}, hrerrors.ErrInvalidInput
	}
	exprs := make([]clause.Expression, 0, len(values))
	for i, value := range values {
		if value == nil || reflect.ValueOf(value).IsZero() {
			return clause.Where{}, hrerrors.ErrInvalidInput
		}
		exprs = append(exprs, clause.Eq{Column: clause.Column{Name: r.model.PrimaryFields[i].DBName}, Value: value})
	}
	return clause.Where{Exprs: exprs}, nil
}

// Insert menyimpan satu baris dan mengembalikan ID identity melalui pointer model.
// Kolom default non-identity tetap eksplisit agar false/nol/NULL tidak berubah diam-diam.
func (r *Repository[T, K]) Insert(ctx context.Context, value *T) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return MapError(db.Clauses(clause.Returning{}).Create(value).Error)
}

// Create adalah alias operasi INSERT; tidak memakai upsert.
func (r *Repository[T, K]) Create(ctx context.Context, value *T) error { return r.Insert(ctx, value) }

// FindByID membaca satu baris memakai seluruh komponen primary key.
func (r *Repository[T, K]) FindByID(ctx context.Context, key K) (*T, error) {
	return r.find(ctx, key, false)
}

// SelectByID menyediakan penamaan SELECT sesuai konvensi proyek.
func (r *Repository[T, K]) SelectByID(ctx context.Context, key K) (*T, error) {
	return r.FindByID(ctx, key)
}

// FindByIDForUpdate mengunci baris hanya di dalam transaksi yang masih aktif.
func (r *Repository[T, K]) FindByIDForUpdate(ctx context.Context, key K) (*T, error) {
	if err := requireTransaction(ctx); err != nil {
		return nil, err
	}
	return r.find(ctx, key, true)
}

func (r *Repository[T, K]) find(ctx context.Context, key K, lock bool) (*T, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	predicate, err := r.keyPredicate(r.keyValues(key))
	if err != nil {
		return nil, err
	}
	if lock {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	value := new(T)
	if err := db.Where(predicate).First(value).Error; err != nil {
		return nil, MapError(err)
	}
	return value, nil
}

// GetAll membaca satu halaman dengan urutan primary key yang stabil.
func (r *Repository[T, K]) GetAll(ctx context.Context, limit, offset int) ([]T, error) {
	return r.selectPage(ctx, nil, limit, offset)
}

// SelectList menyediakan alias SELECT berhalaman.
func (r *Repository[T, K]) SelectList(ctx context.Context, limit, offset int) ([]T, error) {
	return r.GetAll(ctx, limit, offset)
}

// SelectBy hanya menerima kolom model yang dikenal; nilai termasuk NULL diparameterisasi.
func (r *Repository[T, K]) SelectBy(ctx context.Context, column string, value any, limit, offset int) ([]T, error) {
	if r == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if _, ok := r.model.FieldsByDBName[column]; !ok {
		return nil, hrerrors.ErrInvalidInput
	}
	return r.selectPage(ctx, clause.Eq{Column: clause.Column{Name: column}, Value: value}, limit, offset)
}

func (r *Repository[T, K]) selectPage(ctx context.Context, filter clause.Expression, limit, offset int) ([]T, error) {
	limit, offset, err := NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	if filter != nil {
		db = db.Where(filter)
	}
	for _, field := range r.model.PrimaryFields {
		db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: field.DBName}})
	}
	items := make([]T, 0)
	if err := db.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, MapError(err)
	}
	return items, nil
}

// Count menghitung seluruh baris; snapshot terhadap halaman perlu isolasi transaksi yang sesuai.
func (r *Repository[T, K]) Count(ctx context.Context) (int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return 0, err
	}
	var total int64
	err = db.Count(&total).Error
	return total, MapError(err)
}

// Update mengubah snapshot lengkap tanpa Save agar record hilang tidak menjadi INSERT.
// Pemanggil memakai read-lock dalam transaksi untuk perubahan read-modify-write.
func (r *Repository[T, K]) Update(ctx context.Context, value *T) error {
	if value == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	fields := reflect.ValueOf(value).Elem()
	keys := make([]any, 0, len(r.model.PrimaryFields))
	for _, field := range r.model.PrimaryFields {
		val, _ := field.ValueOf(ctx, fields)
		keys = append(keys, val)
	}
	predicate, err := r.keyPredicate(keys)
	if err != nil {
		return err
	}
	updates := make(map[string]any, len(r.model.DBNames))
	for _, field := range r.model.Fields {
		if field.PrimaryKey || field.DBName == "" || !field.Updatable {
			continue
		}
		val, _ := field.ValueOf(ctx, fields)
		updates[field.DBName] = val
	}
	result := db.Where(predicate).Updates(updates)
	if result.Error != nil {
		return MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// DeleteByID menghapus satu primary key; FK RESTRICT dan immutable guard tetap berlaku.
func (r *Repository[T, K]) DeleteByID(ctx context.Context, key K) error {
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	predicate, err := r.keyPredicate(r.keyValues(key))
	if err != nil {
		return err
	}
	result := db.Where(predicate).Delete(new(T))
	if result.Error != nil {
		return MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return hrerrors.ErrNotFound
	}
	return nil
}

// Delete adalah alias penghapusan berdasarkan primary key lengkap.
func (r *Repository[T, K]) Delete(ctx context.Context, key K) error { return r.DeleteByID(ctx, key) }
