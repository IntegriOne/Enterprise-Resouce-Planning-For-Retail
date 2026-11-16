package outbox

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

// Store menyediakan operasi pengiriman outbox di luar CRUD bisnis.
type Store struct{ db *gorm.DB }

// NewStore mengikat pengiriman event pada koneksi dan transaksi HRM yang sama.
func NewStore(db *gorm.DB) (*Store, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) query(ctx context.Context) (*gorm.DB, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := pgstore.RequireTransaction(ctx); err != nil {
		return nil, err
	}
	return pgstore.DBFromContext(ctx, s.db)
}

// SelectDue mengunci event jatuh tempo tanpa menunggu worker lain.
func (s *Store) SelectDue(ctx context.Context, now time.Time, limit, maxAttempts int) ([]entity.OutboxEvent, error) {
	if limit < 1 || limit > 100 || maxAttempts < 1 || now.IsZero() {
		return nil, hrerrors.ErrInvalidInput
	}
	db, err := s.query(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]entity.OutboxEvent, 0, limit)
	err = db.Model(&entity.OutboxEvent{}).
		Where("published_at IS NULL AND available_at <= ? AND attempt_count < ?", now, maxAttempts).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Order("occurred_at ASC").Order("id ASC").Limit(limit).Find(&items).Error
	return items, pgstore.MapError(err)
}

// MarkPublished menandai satu event terkirim setelah publisher menerima respons sukses.
func (s *Store) MarkPublished(ctx context.Context, id string, when time.Time, attempts int32) error {
	if err := pgstore.ValidateKeys(id); err != nil {
		return err
	}
	if when.IsZero() || attempts < 1 {
		return hrerrors.ErrInvalidInput
	}
	db, err := s.query(ctx)
	if err != nil {
		return err
	}
	result := db.Model(&entity.OutboxEvent{}).Where("id = ? AND published_at IS NULL", id).
		Updates(map[string]any{"published_at": when, "attempt_count": attempts, "last_error_code": nil})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected != 1 {
		return hrerrors.ErrConflict
	}
	return nil
}

// MarkRetry menyimpan percobaan dan waktu berikutnya; attempts maksimum menjadi dead letter.
func (s *Store) MarkRetry(ctx context.Context, id string, when time.Time, attempts int32, code string) error {
	if err := pgstore.ValidateKeys(id); err != nil {
		return err
	}
	if when.IsZero() || attempts < 1 || code == "" || len(code) > 80 {
		return hrerrors.ErrInvalidInput
	}
	db, err := s.query(ctx)
	if err != nil {
		return err
	}
	result := db.Model(&entity.OutboxEvent{}).Where("id = ? AND published_at IS NULL", id).
		Updates(map[string]any{"available_at": when, "attempt_count": attempts, "last_error_code": code})
	if result.Error != nil {
		return pgstore.MapError(result.Error)
	}
	if result.RowsAffected != 1 {
		return hrerrors.ErrConflict
	}
	return nil
}
