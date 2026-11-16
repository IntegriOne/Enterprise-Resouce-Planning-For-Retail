package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// OutboxEventRepository adalah port persistensi hr_outbox_events.
type OutboxEventRepository interface {
	Repository[entity.OutboxEvent, string]
	InsertOutboxEvents(context.Context, *entity.OutboxEvent) error
	UpdateOutboxEvents(context.Context, *entity.OutboxEvent) error
	FindOutboxEventsByUUID(context.Context, string) (*entity.OutboxEvent, error)
	FindOutboxEventsByUUIDForUpdate(context.Context, string) (*entity.OutboxEvent, error)
	DeleteOutboxEventsByUUID(context.Context, string) error
	GetAllOutboxEvents(context.Context, int, int) ([]entity.OutboxEvent, error)
	CountOutboxEvents(context.Context) (int64, error)
	SelectOutboxEventsWhere(context.Context, []Filter, int, int) ([]entity.OutboxEvent, error)
	CountOutboxEventsWhere(context.Context, []Filter) (int64, error)
	FindOutboxEventsByAggregateType(context.Context, string, int, int) ([]entity.OutboxEvent, error)
	FindOutboxEventsByAggregateID(context.Context, string, int, int) ([]entity.OutboxEvent, error)
	FindOutboxEventsByAggregateVersion(context.Context, int64, int, int) ([]entity.OutboxEvent, error)
	FindOutboxEventsByEventType(context.Context, string, int, int) ([]entity.OutboxEvent, error)
	FindOutboxEventsByLastErrorCode(context.Context, *string, int, int) ([]entity.OutboxEvent, error)
	FindByAggregateType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error)
	FindByAggregateID(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error)
	FindByAggregateVersion(ctx context.Context, value int64, limit, offset int) ([]entity.OutboxEvent, error)
	FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error)
	FindByLastErrorCode(ctx context.Context, value *string, limit, offset int) ([]entity.OutboxEvent, error)
}
