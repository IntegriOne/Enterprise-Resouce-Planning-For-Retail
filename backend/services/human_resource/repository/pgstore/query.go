// Package pgstore menyediakan helper teknis GORM tanpa implementasi repository generik.
package pgstore

import (
	"context"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// NormalizePage menetapkan default/maksimum dan menolak pagination negatif.
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

// Query memilih transaksi lokal dan model; operasi CRUD tetap ditulis di repository pemilik.
func Query(ctx context.Context, fallback *gorm.DB, model any) (*gorm.DB, error) {
	db, err := DBFromContext(ctx, fallback)
	if err != nil {
		return nil, err
	}
	return db.Model(model), nil
}

// ValidateKeys menolak primary key kosong sebelum kueri mutasi atau pembacaan tunggal.
func ValidateKeys(values ...any) error {
	if len(values) == 0 {
		return hrerrors.ErrInvalidInput
	}
	for _, value := range values {
		if value == nil || reflect.ValueOf(value).IsZero() {
			return hrerrors.ErrInvalidInput
		}
	}
	return nil
}

// ApplyFilters memvalidasi kolom model dan mengikat nilai filter AND/IN sebagai parameter SQL.
// Helper ini tidak mengeksekusi query, pagination atau keputusan akses.
func ApplyFilters(db *gorm.DB, filters []domainrepo.Filter) (*gorm.DB, error) {
	if db == nil || db.Statement == nil || db.Statement.Model == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if len(filters) > 32 {
		return nil, hrerrors.ErrInvalidInput
	}
	if err := db.Statement.Parse(db.Statement.Model); err != nil {
		return nil, MapError(err)
	}
	for _, filter := range filters {
		if _, ok := db.Statement.Schema.FieldsByDBName[filter.Column]; !ok || len(filter.Values) == 0 || len(filter.Values) > 1000 {
			return nil, hrerrors.ErrInvalidInput
		}
		column := clause.Column{Name: filter.Column}
		if len(filter.Values) == 1 {
			db = db.Where(clause.Eq{Column: column, Value: filter.Values[0]})
		} else {
			db = db.Where(clause.IN{Column: column, Values: filter.Values})
		}
	}
	return db, nil
}

// RequireTransaction memastikan pembacaan FOR UPDATE berada dalam transaksi HRM aktif.
func RequireTransaction(ctx context.Context) error {
	return requireTransaction(ctx)
}
