package payroll

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
	"gorm.io/gorm"
)

// InputSnapshotRepository menyimpan dan membaca snapshot tanpa operasi ubah atau hapus.
type InputSnapshotRepository struct {
	db *gorm.DB
}

func NewInputSnapshotRepository(db *gorm.DB) (*InputSnapshotRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &InputSnapshotRepository{db: db}, nil
}

func (r *InputSnapshotRepository) Insert(ctx context.Context, snapshot *entity.PayrollInputSnapshot) error {
	if snapshot == nil {
		return hrerrors.ErrInvalidInput
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(snapshot).Error)
}

func (r *InputSnapshotRepository) ReadCalculationSnapshot(ctx context.Context, id string) (entity.PayrollInputSnapshot, error) {
	if err := entity.ValidateUUID("input_snapshot_id", id); err != nil {
		return entity.PayrollInputSnapshot{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return entity.PayrollInputSnapshot{}, err
	}
	var snapshot entity.PayrollInputSnapshot
	if err := db.Where("id = ?", id).Take(&snapshot).Error; err != nil {
		return entity.PayrollInputSnapshot{}, pgstore.MapError(err)
	}
	return snapshot, nil
}
