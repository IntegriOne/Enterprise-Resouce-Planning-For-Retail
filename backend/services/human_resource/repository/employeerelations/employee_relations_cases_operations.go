package employeerelations

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek tidak mengubah batas kerahasiaan dan otorisasi di service.
func (r *EmployeeRelationsCaseRepository) InsertEmployeeRelationsCases(ctx context.Context, value *entity.EmployeeRelationsCase) error {
	return r.Insert(ctx, value)
}
func (r *EmployeeRelationsCaseRepository) UpdateEmployeeRelationsCases(ctx context.Context, value *entity.EmployeeRelationsCase) error {
	return r.Update(ctx, value)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByUUID(ctx context.Context, uuid string) (*entity.EmployeeRelationsCase, error) {
	return r.FindByID(ctx, uuid)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.EmployeeRelationsCase, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *EmployeeRelationsCaseRepository) DeleteEmployeeRelationsCasesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *EmployeeRelationsCaseRepository) GetAllEmployeeRelationsCases(ctx context.Context, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) CountEmployeeRelationsCases(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *EmployeeRelationsCaseRepository) SelectEmployeeRelationsCasesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) CountEmployeeRelationsCasesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByCaseNumber(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByCaseNumber(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByCaseType(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByCaseType(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByStatus(ctx context.Context, value entity.CaseStatus, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByStatus(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesBySubjectWorkerID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindBySubjectWorkerID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByReportingPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByReportingPartyReferenceID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByAssignedInvestigatorPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByAssignedInvestigatorPrincipalID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByOutcomeCode(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByOutcomeCode(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByRestrictedSummaryReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByRestrictedSummaryReferenceID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByLegalHoldReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByLegalHoldReferenceID(ctx, value, limit, offset)
}
func (r *EmployeeRelationsCaseRepository) FindEmployeeRelationsCasesByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.EmployeeRelationsCase, error) {
	return r.FindByIdempotencyKey(ctx, value, limit, offset)
}
