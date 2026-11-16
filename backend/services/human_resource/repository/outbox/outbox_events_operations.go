package outbox

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek menyimpan/membaca outbox lokal tanpa publikasi jaringan.
func (r *OutboxEventRepository) InsertOutboxEvents(ctx context.Context, value *entity.OutboxEvent) error {
	return r.Insert(ctx, value)
}
func (r *OutboxEventRepository) UpdateOutboxEvents(ctx context.Context, value *entity.OutboxEvent) error {
	return r.Update(ctx, value)
}
func (r *OutboxEventRepository) FindOutboxEventsByUUID(ctx context.Context, uuid string) (*entity.OutboxEvent, error) {
	return r.FindByID(ctx, uuid)
}
func (r *OutboxEventRepository) FindOutboxEventsByUUIDForUpdate(ctx context.Context, uuid string) (*entity.OutboxEvent, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *OutboxEventRepository) DeleteOutboxEventsByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *OutboxEventRepository) GetAllOutboxEvents(ctx context.Context, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *OutboxEventRepository) CountOutboxEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *OutboxEventRepository) SelectOutboxEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *OutboxEventRepository) CountOutboxEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateType(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateID(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateID(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByAggregateVersion(ctx context.Context, value int64, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByAggregateVersion(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *OutboxEventRepository) FindOutboxEventsByLastErrorCode(ctx context.Context, value *string, limit, offset int) ([]entity.OutboxEvent, error) {
	return r.FindByLastErrorCode(ctx, value, limit, offset)
}
