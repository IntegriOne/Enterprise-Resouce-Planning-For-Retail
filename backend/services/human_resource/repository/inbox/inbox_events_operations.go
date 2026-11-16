package inbox

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek mempertahankan jalur transaksi yang sama dengan method lama.
func (r *InboxEventRepository) InsertInboxEvents(ctx context.Context, value *entity.InboxEvent) error {
	return r.Insert(ctx, value)
}
func (r *InboxEventRepository) UpdateInboxEvents(ctx context.Context, value *entity.InboxEvent) error {
	return r.Update(ctx, value)
}
func (r *InboxEventRepository) FindInboxEventsByUUID(ctx context.Context, uuid string) (*entity.InboxEvent, error) {
	return r.FindByID(ctx, uuid)
}
func (r *InboxEventRepository) FindInboxEventsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.InboxEvent, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *InboxEventRepository) DeleteInboxEventsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *InboxEventRepository) GetAllInboxEvents(ctx context.Context, limit, offset int) ([]entity.InboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *InboxEventRepository) CountInboxEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *InboxEventRepository) SelectInboxEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.InboxEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *InboxEventRepository) CountInboxEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *InboxEventRepository) FindInboxEventsBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindBySourceManagement(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindBySourceEventID(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByStatus(ctx context.Context, value string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *InboxEventRepository) FindInboxEventsByFailureCode(ctx context.Context, value *string, limit, offset int) ([]entity.InboxEvent, error) {
	return r.FindByFailureCode(ctx, value, limit, offset)
}
