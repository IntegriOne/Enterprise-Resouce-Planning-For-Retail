package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// InboxEventRepository adalah port persistensi hr_inbox_events.
type InboxEventRepository interface {
	Repository[entity.InboxEvent, string]
	InsertInboxEvents(context.Context, *entity.InboxEvent) error
	UpdateInboxEvents(context.Context, *entity.InboxEvent) error
	FindInboxEventsByUUID(context.Context, string) (*entity.InboxEvent, error)
	FindInboxEventsByUUIDForUpdate(context.Context, string) (*entity.InboxEvent, error)
	DeleteInboxEventsByUUID(context.Context, string) error
	GetAllInboxEvents(context.Context, int, int) ([]entity.InboxEvent, error)
	CountInboxEvents(context.Context) (int64, error)
	SelectInboxEventsWhere(context.Context, []Filter, int, int) ([]entity.InboxEvent, error)
	CountInboxEventsWhere(context.Context, []Filter) (int64, error)
	FindInboxEventsBySourceManagement(context.Context, string, int, int) ([]entity.InboxEvent, error)
	FindInboxEventsBySourceEventID(context.Context, string, int, int) ([]entity.InboxEvent, error)
	FindInboxEventsByEventType(context.Context, string, int, int) ([]entity.InboxEvent, error)
	FindInboxEventsByStatus(context.Context, string, int, int) ([]entity.InboxEvent, error)
	FindInboxEventsByFailureCode(context.Context, *string, int, int) ([]entity.InboxEvent, error)
	FindBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error)
	FindBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error)
	FindByEventType(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error)
	FindByStatus(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error)
	FindByFailureCode(ctx context.Context, value *string, limit, offset int) ([]entity.InboxEvent, error)
}
