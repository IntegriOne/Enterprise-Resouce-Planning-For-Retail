package personnel

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// WorkerAssignmentTransferRepository menyimpan permintaan dan keputusan transfer personel.
type WorkerAssignmentTransferRepository struct {
	db *gorm.DB
}

var _ domainrepo.WorkerAssignmentTransferRepository = (*WorkerAssignmentTransferRepository)(nil)

func NewWorkerAssignmentTransferRepository(db *gorm.DB) (*WorkerAssignmentTransferRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &WorkerAssignmentTransferRepository{db: db}, nil
}

func (r *WorkerAssignmentTransferRepository) query(ctx context.Context) (*gorm.DB, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	return pgstore.Query(ctx, r.db, &entity.WorkerAssignmentTransfer{})
}

// Insert menyimpan permintaan baru.
func (r *WorkerAssignmentTransferRepository) Insert(ctx context.Context, transfer *entity.WorkerAssignmentTransfer) error {
	if transfer == nil || transfer.Status != "submitted" {
		return hrerrors.ErrInvalidInput
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Clauses(clause.Returning{}).Create(transfer).Error)
}

// FindByIdempotencyKey mengambil permintaan untuk rekonsiliasi retry.
func (r *WorkerAssignmentTransferRepository) FindByIdempotencyKey(ctx context.Context, key string) (*entity.WorkerAssignmentTransfer, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	transfer := new(entity.WorkerAssignmentTransfer)
	if err := db.Where("idempotency_key = ?", key).Take(transfer).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return transfer, nil
}

// FindByID membaca permintaan transfer berdasarkan ID.
func (r *WorkerAssignmentTransferRepository) FindByID(ctx context.Context, id string) (*entity.WorkerAssignmentTransfer, error) {
	if err := pgstore.ValidateKeys(id); err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	transfer := new(entity.WorkerAssignmentTransfer)
	if err := db.Where("id = ?", id).Take(transfer).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return transfer, nil
}

// FindByIDForUpdate mengunci permintaan saat keputusan diambil.
func (r *WorkerAssignmentTransferRepository) FindByIDForUpdate(ctx context.Context, id string) (*entity.WorkerAssignmentTransfer, error) {
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	if err := pgstore.ValidateKeys(id); err != nil {
		return nil, err
	}
	db, err := r.query(ctx)
	if err != nil {
		return nil, err
	}
	transfer := new(entity.WorkerAssignmentTransfer)
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(transfer).Error; err != nil {
		return nil, pgstore.MapError(err)
	}
	return transfer, nil
}

// UpdateDecision hanya menyimpan keputusan terminal pada permintaan yang masih diajukan.
func (r *WorkerAssignmentTransferRepository) UpdateDecision(ctx context.Context, transfer *entity.WorkerAssignmentTransfer) error {
	if transfer == nil || (transfer.Status != "applied" && transfer.Status != "rejected") {
		return hrerrors.ErrInvalidInput
	}
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return err
	}
	db, err := r.query(ctx)
	if err != nil {
		return err
	}
	result := db.Where("id = ? AND status = ?", transfer.ID, "submitted").Updates(map[string]any{
		"status":                  transfer.Status,
		"decided_by_principal_id": transfer.DecidedByPrincipalID,
		"decided_at":              transfer.DecidedAt,
		"applied_assignment_id":   transfer.AppliedAssignmentID,
	})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected != 1 {
		return hrerrors.ErrConflict
	}
	return nil
}

// ListByWorkerID membaca halaman riwayat transfer dengan urutan stabil.
func (r *WorkerAssignmentTransferRepository) ListByWorkerID(ctx context.Context, workerID string, limit, offset int) ([]entity.WorkerAssignmentTransfer, int64, error) {
	db, err := r.query(ctx)
	if err != nil {
		return nil, 0, err
	}
	filter := db.Where("worker_id = ?", workerID)
	var total int64
	if err := filter.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, pgstore.MapError(err)
	}
	var items []entity.WorkerAssignmentTransfer
	err = filter.Order("requested_at DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, pgstore.MapError(err)
}
