// Sumber model adalah baseline SQL HRM; aturan bisnis tetap milik service.
package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// BenefitPlanRepository adalah port persistensi hr_benefit_plans.
type BenefitPlanRepository interface {
	Repository[entity.BenefitPlan, string]
	InsertBenefitPlans(context.Context, *entity.BenefitPlan) error
	UpdateBenefitPlans(context.Context, *entity.BenefitPlan) error
	DeleteBenefitPlansByUUID(context.Context, string) error
	FindBenefitPlansByUUID(context.Context, string) (*entity.BenefitPlan, error)
	FindBenefitPlansByUUIDForUpdate(context.Context, string) (*entity.BenefitPlan, error)
	GetAllBenefitPlans(context.Context, int, int) ([]entity.BenefitPlan, error)
	CountBenefitPlans(context.Context) (int64, error)
	SelectBenefitPlansWhere(context.Context, []Filter, int, int) ([]entity.BenefitPlan, error)
	CountBenefitPlansWhere(context.Context, []Filter) (int64, error)
	FindBenefitPlansByPlanCode(context.Context, string, int, int) ([]entity.BenefitPlan, error)
	FindBenefitPlansByBenefitType(context.Context, string, int, int) ([]entity.BenefitPlan, error)
	FindBenefitPlansByProviderPartyReferenceID(context.Context, *string, int, int) ([]entity.BenefitPlan, error)
	FindBenefitPlansByJurisdictionCode(context.Context, string, int, int) ([]entity.BenefitPlan, error)
	FindBenefitPlansByEmployeeContributionRuleReferenceID(context.Context, *string, int, int) ([]entity.BenefitPlan, error)
	FindBenefitPlansByEmployerContributionRuleReferenceID(context.Context, *string, int, int) ([]entity.BenefitPlan, error)
	FindByPlanCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error)
	FindByBenefitType(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error)
	FindByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.BenefitPlan, error)
	FindByEmployeeContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error)
	FindByEmployerContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitPlan, error)
}

// BenefitEnrollmentRepository adalah port persistensi hr_benefit_enrollments.
type BenefitEnrollmentRepository interface {
	Repository[entity.BenefitEnrollment, string]
	InsertBenefitEnrollments(context.Context, *entity.BenefitEnrollment) error
	UpdateBenefitEnrollments(context.Context, *entity.BenefitEnrollment) error
	DeleteBenefitEnrollmentsByUUID(context.Context, string) error
	FindBenefitEnrollmentsByUUID(context.Context, string) (*entity.BenefitEnrollment, error)
	FindBenefitEnrollmentsByUUIDForUpdate(context.Context, string) (*entity.BenefitEnrollment, error)
	GetAllBenefitEnrollments(context.Context, int, int) ([]entity.BenefitEnrollment, error)
	CountBenefitEnrollments(context.Context) (int64, error)
	SelectBenefitEnrollmentsWhere(context.Context, []Filter, int, int) ([]entity.BenefitEnrollment, error)
	CountBenefitEnrollmentsWhere(context.Context, []Filter) (int64, error)
	FindBenefitEnrollmentsByWorkerID(context.Context, string, int, int) ([]entity.BenefitEnrollment, error)
	FindBenefitEnrollmentsByBenefitPlanID(context.Context, string, int, int) ([]entity.BenefitEnrollment, error)
	FindBenefitEnrollmentsByEnrollmentStatus(context.Context, string, int, int) ([]entity.BenefitEnrollment, error)
	FindBenefitEnrollmentsByCurrencyCode(context.Context, *string, int, int) ([]entity.BenefitEnrollment, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error)
	FindByBenefitPlanID(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error)
	FindByEnrollmentStatus(ctx context.Context, value string, limit, offset int) ([]entity.BenefitEnrollment, error)
	FindByCurrencyCode(ctx context.Context, value *string, limit, offset int) ([]entity.BenefitEnrollment, error)
}
