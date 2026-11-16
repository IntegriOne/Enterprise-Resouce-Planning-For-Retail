// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime           operation.Dependencies
	BenefitPlan       operation.Repository[entity.BenefitPlan, string]
	BenefitEnrollment operation.Repository[entity.BenefitEnrollment, string]
}

// Service mengoordinasikan operasi domain benefits.
type Service struct {
	benefitPlan       *operation.Resource[entity.BenefitPlan, string, *entity.BenefitPlan]
	benefitEnrollment *operation.Resource[entity.BenefitEnrollment, string, *entity.BenefitEnrollment]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.benefitPlan, err = operation.NewResource[entity.BenefitPlan, string, *entity.BenefitPlan](deps.BenefitPlan, deps.Runtime, specBenefitPlan())
	if err != nil {
		return nil, err
	}
	s.benefitEnrollment, err = operation.NewResource[entity.BenefitEnrollment, string, *entity.BenefitEnrollment](deps.BenefitEnrollment, deps.Runtime, specBenefitEnrollment())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateBenefitPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateBenefitPlan(ctx context.Context, in CreateBenefitPlanInput) (*entity.BenefitPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.Create(ctx, func() (*entity.BenefitPlan, error) { return createBenefitPlan(in) })
}

// UpdateBenefitPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateBenefitPlan(ctx context.Context, key string, in UpdateBenefitPlanInput) (*entity.BenefitPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.Update(ctx, key, func(value *entity.BenefitPlan) error { return patchBenefitPlan(value, in) })
}

// DeleteBenefitPlan menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteBenefitPlan(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.benefitPlan.Delete(ctx, key)
}

// GetBenefitPlanByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetBenefitPlanByID(ctx context.Context, key string) (*entity.BenefitPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.Get(ctx, key)
}

// GetAllBenefitPlans menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllBenefitPlans(ctx context.Context, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, nil, limit, offset)
}

// FindBenefitPlanByPlanCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByPlanCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "plan_code", Values: []any{value}}}, limit, offset)
}

// FindBenefitPlanByBenefitType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByBenefitType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "benefit_type", Values: []any{value}}}, limit, offset)
}

// FindBenefitPlanByProviderPartyReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByProviderPartyReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "provider_party_reference_id", Values: []any{value}}}, limit, offset)
}

// FindBenefitPlanByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// FindBenefitPlanByEmployeeContributionRuleReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByEmployeeContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "employee_contribution_rule_reference_id", Values: []any{value}}}, limit, offset)
}

// FindBenefitPlanByEmployerContributionRuleReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitPlanByEmployerContributionRuleReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.BenefitPlan], error) {
	if s == nil {
		return operation.Page[entity.BenefitPlan]{}, hrerrors.ErrUnavailable
	}
	return s.benefitPlan.List(ctx, []domainrepo.Filter{{Column: "employer_contribution_rule_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateBenefitEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateBenefitEnrollment(ctx context.Context, in CreateBenefitEnrollmentInput) (*entity.BenefitEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.Create(ctx, func() (*entity.BenefitEnrollment, error) { return createBenefitEnrollment(in) })
}

// UpdateBenefitEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateBenefitEnrollment(ctx context.Context, key string, in UpdateBenefitEnrollmentInput) (*entity.BenefitEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.Update(ctx, key, func(value *entity.BenefitEnrollment) error { return patchBenefitEnrollment(value, in) })
}

// DeleteBenefitEnrollment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteBenefitEnrollment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.Delete(ctx, key)
}

// GetBenefitEnrollmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetBenefitEnrollmentByID(ctx context.Context, key string) (*entity.BenefitEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.Get(ctx, key)
}

// GetAllBenefitEnrollments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllBenefitEnrollments(ctx context.Context, limit, offset int) (operation.Page[entity.BenefitEnrollment], error) {
	if s == nil {
		return operation.Page[entity.BenefitEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.List(ctx, nil, limit, offset)
}

// FindBenefitEnrollmentByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitEnrollmentByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitEnrollment], error) {
	if s == nil {
		return operation.Page[entity.BenefitEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindBenefitEnrollmentByBenefitPlanID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitEnrollmentByBenefitPlanID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitEnrollment], error) {
	if s == nil {
		return operation.Page[entity.BenefitEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.List(ctx, []domainrepo.Filter{{Column: "benefit_plan_id", Values: []any{value}}}, limit, offset)
}

// FindBenefitEnrollmentByEnrollmentStatus menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitEnrollmentByEnrollmentStatus(ctx context.Context, value string, limit, offset int) (operation.Page[entity.BenefitEnrollment], error) {
	if s == nil {
		return operation.Page[entity.BenefitEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.List(ctx, []domainrepo.Filter{{Column: "enrollment_status", Values: []any{value}}}, limit, offset)
}

// FindBenefitEnrollmentByCurrencyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindBenefitEnrollmentByCurrencyCode(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.BenefitEnrollment], error) {
	if s == nil {
		return operation.Page[entity.BenefitEnrollment]{}, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.List(ctx, []domainrepo.Filter{{Column: "currency_code", Values: []any{value}}}, limit, offset)
}
