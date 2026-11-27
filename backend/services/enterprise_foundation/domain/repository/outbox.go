package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

// OutboxRepository mendefinisikan operasi antrean event transaksional pada tabel foundation_outbox_events.
type OutboxRepository interface {
	Insert(ctx context.Context, event *entity.OutboxEvent) error
	SelectByID(ctx context.Context, id string) (*entity.OutboxEvent, error)
	SelectByAggregate(ctx context.Context, aggregateType, aggregateID string, page PageRequest) (PageResult[*entity.OutboxEvent], error)
	SelectDuePending(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error)
	UpdateEffective(ctx context.Context, id string, processedAt time.Time) error
	UpdateRetry(ctx context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error
	UpdateFailed(ctx context.Context, id string, errMsg string) error
	CountPending(ctx context.Context) (int64, error)
	CountFailed(ctx context.Context) (int64, error)
	DeleteEffectiveBefore(ctx context.Context, cutoff time.Time) (int64, error)
}
