package employeerelations

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// ID event berupa int64, sehingga nama pencarian memakai ByID, bukan ByUUID.
func (r *EmployeeRelationsEventRepository) InsertEmployeeRelationsEvents(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
	return r.Insert(ctx, value)
}
func (r *EmployeeRelationsEventRepository) UpdateEmployeeRelationsEvents(ctx context.Context, value *entity.EmployeeRelationsEvent) error {
	return r.Update(ctx, value)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByID(ctx context.Context, id int64) (*entity.EmployeeRelationsEvent, error) {
	return r.FindByID(ctx, id)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByIDForUpdate(ctx context.Context, id int64) (*entity.EmployeeRelationsEvent, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *EmployeeRelationsEventRepository) DeleteEmployeeRelationsEventsByID(ctx context.Context, id int64) error {
	return r.DeleteByID(ctx, id)
}
func (r *EmployeeRelationsEventRepository) GetAllEmployeeRelationsEvents(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmployeeRelationsEventRepository) CountEmployeeRelationsEvents(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EmployeeRelationsEventRepository) SelectEmployeeRelationsEventsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EmployeeRelationsEventRepository) CountEmployeeRelationsEventsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByCaseID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByCaseID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByEventType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByEventType(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByActorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByActorPrincipalID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsEventRepository) FindEmployeeRelationsEventsByEventReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsEvent, error) {
	return r.FindByEventReferenceID(ctx, value, limit, offset)
}
