package payroll

import (
	"context"
	_ "embed"
	"time"

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

//go:embed input_snapshot_capture.sql
var capturePayrollInputSQL string

// Capture menyimpan input HRM dalam satu pernyataan agar seluruh pembacaan memakai snapshot PostgreSQL yang sama.
func (r *InputSnapshotRepository) Capture(ctx context.Context, id, periodID, principalID string, at time.Time) (entity.PayrollInputSnapshot, error) {
	for _, field := range []struct{ name, value string }{
		{"input_snapshot_id", id}, {"payroll_period_id", periodID}, {"principal_id", principalID},
	} {
		if err := entity.ValidateUUID(field.name, field.value); err != nil {
			return entity.PayrollInputSnapshot{}, err
		}
	}
	if at.IsZero() {
		return entity.PayrollInputSnapshot{}, hrerrors.ErrInvalidInput
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return entity.PayrollInputSnapshot{}, err
	}
	var snapshot entity.PayrollInputSnapshot
	result := db.Raw(capturePayrollInputSQL, periodID, at, at, id, principalID, at).Scan(&snapshot)
	if result.Error != nil {
		return entity.PayrollInputSnapshot{}, pgstore.MapError(result.Error)
	}
	if result.RowsAffected != 1 {
		return entity.PayrollInputSnapshot{}, hrerrors.ErrWorkflowRequired
	}
	return snapshot, nil
}
