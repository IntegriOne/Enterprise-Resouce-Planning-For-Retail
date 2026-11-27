package outbox

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

var _ domainrepo.OutboxRepository = (*Repository)(nil)

// Repository mengimplementasikan OutboxRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun OutboxRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// Insert menyimpan event outbox baru ke tabel foundation_outbox_events dalam transaksi yang sama dengan perubahan domain.
func (r *Repository) Insert(ctx context.Context, event *entity.OutboxEvent) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := outboxToModel(event)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// SelectByID mengambil satu event outbox berdasarkan UUID untuk inspeksi atau diagnosa operasional.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.OutboxEvent, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m outboxEventModel
	if err := db.Where("id = ?", eid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, eferrors.ErrOutboxEventNotFound)
	}
	return outboxToDomain(&m)
}

// SelectByAggregate mengambil riwayat event outbox milik satu entitas agregat tertentu secara berhalaman.
func (r *Repository) SelectByAggregate(ctx context.Context, aggregateType, aggregateID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.OutboxEvent]{}, pgstore.CheckDB(nil)
	}
	aggType, err := entity.ValidateRequiredText(
		aggregateType,
		entity.MaxOutboxAggregateTypeLength,
		eferrors.WithField(eferrors.ErrInvalidInput, "aggregate_type", "aggregate type must be non-empty and within length bounds"),
	)
	if err != nil {
		return domainrepo.PageResult[*entity.OutboxEvent]{}, err
	}
	aggID, err := pgstore.ValidateUUIDParam("aggregate_id", aggregateID)
	if err != nil {
		return domainrepo.PageResult[*entity.OutboxEvent]{}, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.OutboxEvent]{}, err
	}
	q := db.Model(&outboxEventModel{}).
		Where("aggregate_type = ? AND aggregate_id = ?", aggType, aggID)
	return pgstore.Paginate(q, page, "created_at DESC, id ASC", outboxToDomain)
}

// SelectDuePending mengambil batch event outbox berstatus pending yang sudah jatuh tempo menggunakan SKIP LOCKED agar aman di worker paralel.
func (r *Repository) SelectDuePending(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	if limit <= 0 {
		limit = domainrepo.DefaultPageLimit
	} else if limit > domainrepo.MaxPageLimit {
		limit = domainrepo.MaxPageLimit
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	ts := pgstore.MustLocalNow(now)
	var rows []outboxEventModel
	if err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND next_retry_at <= ?", entity.OutboxStatusPending.String(), ts).
		Order("next_retry_at ASC, created_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return pgstore.MapSlice(rows, outboxToDomain)
}

// UpdateEffective menandai event outbox berhasil dikirim pada processedAt.
func (r *Repository) UpdateEffective(ctx context.Context, id string, processedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(processedAt)
	res := db.Model(&outboxEventModel{}).
		Where("id = ?", eid).
		Updates(map[string]any{
			"status":       entity.OutboxStatusEffective.String(),
			"last_error":   nil,
			"processed_at": ts,
			"updated_at":   ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrOutboxEventNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrOutboxEventNotFound
	}
	return nil
}

// UpdateRetry memperbarui percobaan dan jadwal pengiriman ulang event outbox berstatus pending.
func (r *Repository) UpdateRetry(ctx context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	if attemptCount < 0 {
		return eferrors.WithField(eferrors.ErrInvalidInput, "attempt_count", "attempt count cannot be negative")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	now := time.Now()
	retryAt := pgstore.MustLocalNow(nextRetryAt)
	res := db.Model(&outboxEventModel{}).
		Where("id = ?", eid).
		Updates(map[string]any{
			"status":        entity.OutboxStatusPending.String(),
			"attempt_count": attemptCount,
			"next_retry_at": retryAt,
			"last_error":    pgstore.NullableStringPtr(errMsg),
			"updated_at":    now,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrOutboxEventNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrOutboxEventNotFound
	}
	return nil
}

// UpdateFailed menandai event outbox gagal permanen beserta catatan error terakhir.
func (r *Repository) UpdateFailed(ctx context.Context, id string, errMsg string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	eid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	now := time.Now()
	res := db.Model(&outboxEventModel{}).
		Where("id = ?", eid).
		Updates(map[string]any{
			"status":     entity.OutboxStatusFailed.String(),
			"last_error": pgstore.NullableStringPtr(errMsg),
			"updated_at": now,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, eferrors.ErrOutboxEventNotFound)
	}
	if res.RowsAffected == 0 {
		return eferrors.ErrOutboxEventNotFound
	}
	return nil
}

// CountPending menghitung total event outbox yang masih berstatus pending.
func (r *Repository) CountPending(ctx context.Context) (int64, error) {
	return r.countByStatus(ctx, entity.OutboxStatusPending)
}

// CountFailed menghitung total event outbox yang berstatus failed.
func (r *Repository) CountFailed(ctx context.Context) (int64, error) {
	return r.countByStatus(ctx, entity.OutboxStatusFailed)
}

func (r *Repository) countByStatus(ctx context.Context, status entity.OutboxStatus) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := db.Model(&outboxEventModel{}).
		Where("status = ?", status.String()).
		Count(&total).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return total, nil
}

// DeleteEffectiveBefore membersihkan arsip event outbox berstatus effective yang telah diproses sebelum batas waktu cutoff.
func (r *Repository) DeleteEffectiveBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	if cutoff.IsZero() {
		return 0, eferrors.WithField(eferrors.ErrInvalidInput, "cutoff", "cutoff timestamp must not be zero")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	res := db.Where(
		"status = ? AND processed_at IS NOT NULL AND processed_at < ?",
		entity.OutboxStatusEffective.String(),
		pgstore.NormalizeLocalTime(cutoff),
	).Delete(&outboxEventModel{})
	if res.Error != nil {
		return 0, pgstore.MapError(res.Error, nil)
	}
	return res.RowsAffected, nil
}
