package pgstore

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

func (r *Repository[T, K]) filtered(ctx context.Context, filters []domainrepo.Filter) (*gorm.DB, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	if len(filters) > 32 {
		return nil, hrerrors.ErrInvalidInput
	}
	for _, filter := range filters {
		if _, ok := r.model.FieldsByDBName[filter.Column]; !ok || len(filter.Values) == 0 || len(filter.Values) > 1000 {
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

// SelectWhere menerapkan filter scope sebelum limit/offset, dengan parameter SQL terikat.
func (r *Repository[T, K]) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]T, error) {
	limit, offset, err := NormalizePage(limit, offset)
	if err != nil {
		return nil, err
	}
	db, err := r.filtered(ctx, filters)
	if err != nil {
		return nil, err
	}
	for _, field := range r.model.PrimaryFields {
		db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: field.DBName}})
	}
	rows := make([]T, 0)
	if err := db.Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, MapError(err)
	}
	return rows, nil
}

// CountWhere menghitung hanya baris yang memenuhi filter policy yang sama dengan list.
func (r *Repository[T, K]) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	db, err := r.filtered(ctx, filters)
	if err != nil {
		return 0, err
	}
	var count int64
	err = db.Count(&count).Error
	return count, MapError(err)
}
