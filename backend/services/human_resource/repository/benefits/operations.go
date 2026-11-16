package benefits

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi berikut menyebut objek dan kolom secara eksplisit untuk memudahkan penelusuran.
func (r *BenefitPlanRepository) InsertBenefitPlans(ctx context.Context, value *entity.BenefitPlan) error {
	return r.Insert(ctx, value)
}
func (r *BenefitPlanRepository) UpdateBenefitPlans(ctx context.Context, value *entity.BenefitPlan) error {
	return r.Update(ctx, value)
}
func (r *BenefitPlanRepository) DeleteBenefitPlansByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *BenefitPlanRepository) FindBenefitPlansByUUID(ctx context.Context, id string) (*entity.BenefitPlan, error) {
	return r.FindByID(ctx, id)
}
func (r *BenefitPlanRepository) FindBenefitPlansByUUIDForUpdate(ctx context.Context, id string) (*entity.BenefitPlan, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *BenefitPlanRepository) GetAllBenefitPlans(ctx context.Context, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *BenefitPlanRepository) CountBenefitPlans(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *BenefitPlanRepository) SelectBenefitPlansWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *BenefitPlanRepository) CountBenefitPlansWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *BenefitPlanRepository) FindBenefitPlansByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByPlanCode(ctx, value, limit, offset)
}
func (r *BenefitPlanRepository) FindBenefitPlansByBenefitType(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByBenefitType(ctx, value, limit, offset)
}
func (r *BenefitPlanRepository) FindBenefitPlansByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByProviderPartyReferenceID(ctx, value, limit, offset)
}
func (r *BenefitPlanRepository) FindBenefitPlansByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByJurisdictionCode(ctx, value, limit, offset)
}
func (r *BenefitPlanRepository) FindBenefitPlansByEmployeeContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByEmployeeContributionRuleReferenceID(ctx, value, limit, offset)
}
func (r *BenefitPlanRepository) FindBenefitPlansByEmployerContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error) {
	return r.FindByEmployerContributionRuleReferenceID(ctx, value, limit, offset)
}

func (r *BenefitEnrollmentRepository) InsertBenefitEnrollments(ctx context.Context, value *entity.BenefitEnrollment) error {
	return r.Insert(ctx, value)
}
func (r *BenefitEnrollmentRepository) UpdateBenefitEnrollments(ctx context.Context, value *entity.BenefitEnrollment) error {
	return r.Update(ctx, value)
}
func (r *BenefitEnrollmentRepository) DeleteBenefitEnrollmentsByUUID(ctx context.Context, id string) error {
	return r.DeleteByID(ctx, id)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByUUID(ctx context.Context, id string) (*entity.BenefitEnrollment, error) {
	return r.FindByID(ctx, id)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByUUIDForUpdate(ctx context.Context, id string) (*entity.BenefitEnrollment, error) {
	return r.FindByIDForUpdate(ctx, id)
}
func (r *BenefitEnrollmentRepository) GetAllBenefitEnrollments(ctx context.Context, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *BenefitEnrollmentRepository) CountBenefitEnrollments(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *BenefitEnrollmentRepository) SelectBenefitEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *BenefitEnrollmentRepository) CountBenefitEnrollmentsWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByWorkerID(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByBenefitPlanID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByBenefitPlanID(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByEnrollmentStatus(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByEnrollmentStatus(ctx, value, limit, offset)
}
func (r *BenefitEnrollmentRepository) FindBenefitEnrollmentsByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitEnrollment, error) {
	return r.FindByCurrencyCode(ctx, value, limit, offset)
}
